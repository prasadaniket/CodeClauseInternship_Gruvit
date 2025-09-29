package services

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type WebSocketClient struct {
	ID       string
	UserID   string
	Conn     *websocket.Conn
	Send     chan WebSocketMessage
	Hub      *WebSocketHub
	LastPing time.Time
}

type WebSocketHub struct {
	Clients     map[*WebSocketClient]bool
	register    chan *WebSocketClient
	unregister  chan *WebSocketClient
	broadcast   chan WebSocketMessage
	UserClients map[string][]*WebSocketClient // Map user ID to their clients
	Mutex       sync.RWMutex
	redisClient *redis.Client
}

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should check the origin
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewWebSocketHub(redisClient *redis.Client) *WebSocketHub {
	return &WebSocketHub{
		Clients:     make(map[*WebSocketClient]bool),
		register:    make(chan *WebSocketClient),
		unregister:  make(chan *WebSocketClient),
		broadcast:   make(chan WebSocketMessage),
		UserClients: make(map[string][]*WebSocketClient),
		redisClient: redisClient,
	}
}

func (h *WebSocketHub) Run() {
	// Start Redis subscriber for real-time updates
	go h.subscribeToRedis()

	for {
		select {
		case client := <-h.register:
			h.Mutex.Lock()
			h.Clients[client] = true
			if client.UserID != "" {
				h.UserClients[client.UserID] = append(h.UserClients[client.UserID], client)
			}
			h.Mutex.Unlock()
			log.Printf("WebSocket client connected: %s (User: %s)", client.ID, client.UserID)

		case client := <-h.unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)

				// Remove from user clients
				if client.UserID != "" {
					clients := h.UserClients[client.UserID]
					for i, c := range clients {
						if c == client {
							h.UserClients[client.UserID] = append(clients[:i], clients[i+1:]...)
							break
						}
					}
					if len(h.UserClients[client.UserID]) == 0 {
						delete(h.UserClients, client.UserID)
					}
				}
			}
			h.Mutex.Unlock()
			log.Printf("WebSocket client disconnected: %s", client.ID)

		case message := <-h.broadcast:
			h.Mutex.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.Mutex.RUnlock()
		}
	}
}

func (h *WebSocketHub) BroadcastToUser(userID string, message WebSocketMessage) {
	h.Mutex.RLock()
	clients, exists := h.UserClients[userID]
	h.Mutex.RUnlock()

	if !exists {
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// Client is not ready to receive, skip
		}
	}
}

func (h *WebSocketHub) BroadcastToAll(message WebSocketMessage) {
	h.broadcast <- message
}

func (h *WebSocketHub) subscribeToRedis() {
	if h.redisClient == nil {
		return
	}

	pubsub := h.redisClient.Subscribe(context.Background(), "notifications:global", "live_updates")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			log.Printf("Redis subscription error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var message WebSocketMessage
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("Failed to unmarshal Redis message: %v", err)
			continue
		}

		// Broadcast to all connected clients
		h.BroadcastToAll(message)
	}
}

// WebSocket client methods
func (c *WebSocketClient) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg WebSocketMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages from client
		switch msg.Type {
		case "ping":
			c.Conn.WriteJSON(WebSocketMessage{
				Type:      "pong",
				Timestamp: time.Now(),
			})
		case "join_room":
			// Handle room joining logic
			log.Printf("User %s joining room: %v", c.UserID, msg.Data)
		case "leave_room":
			// Handle room leaving logic
			log.Printf("User %s leaving room: %v", c.UserID, msg.Data)
		default:
			log.Printf("Unknown message type from client: %s", msg.Type)
		}
	}
}

func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWebSocket handles WebSocket upgrade requests
func ServeWebSocket(hub *WebSocketHub, w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WebSocketClient{
		ID:       generateClientID(),
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan WebSocketMessage, 256),
		Hub:      hub,
		LastPing: time.Now(),
	}

	client.Hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// WebSocketService handles WebSocket connections and real-time features
type WebSocketService struct {
	hub         *WebSocketHub
	redisClient *redis.Client
}

func NewWebSocketService(redisClient *redis.Client) *WebSocketService {
	hub := NewWebSocketHub(redisClient)
	go hub.Run()

	return &WebSocketService{
		hub:         hub,
		redisClient: redisClient,
	}
}

func (s *WebSocketService) GetHub() *WebSocketHub {
	return s.hub
}

func (s *WebSocketService) SendToUser(userID string, messageType string, data map[string]interface{}) {
	message := WebSocketMessage{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}
	s.hub.BroadcastToUser(userID, message)
}

func (s *WebSocketService) SendToAll(messageType string, data map[string]interface{}) {
	message := WebSocketMessage{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}
	s.hub.BroadcastToAll(message)
}

func (s *WebSocketService) PublishToRedis(channel string, message WebSocketMessage) {
	if s.redisClient == nil {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}

	s.redisClient.Publish(context.Background(), channel, data)
}

// Real-time features

// NotifyPlaylistUpdate sends playlist update notifications to users
func (s *WebSocketService) NotifyPlaylistUpdate(playlistID string, ownerID string, action string, track interface{}) {
	message := WebSocketMessage{
		Type: "playlist_update",
		Data: map[string]interface{}{
			"playlist_id": playlistID,
			"action":      action, // "added", "removed", "reordered"
			"track":       track,
			"timestamp":   time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Send to playlist owner
	s.SendToUser(ownerID, "playlist_update", message.Data)

	// Publish to Redis for scaling across multiple instances
	s.PublishToRedis("playlist_updates", message)
}

// NotifyUserPresence broadcasts user online/offline status
func (s *WebSocketService) NotifyUserPresence(userID string, status string) {
	message := WebSocketMessage{
		Type: "user_presence",
		Data: map[string]interface{}{
			"user_id":   userID,
			"status":    status, // "online", "offline", "away"
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Broadcast to all connected users
	s.SendToAll("user_presence", message.Data)

	// Publish to Redis for scaling
	s.PublishToRedis("user_presence", message)
}

// NotifyNowPlaying broadcasts currently playing track
func (s *WebSocketService) NotifyNowPlaying(userID string, track interface{}) {
	message := WebSocketMessage{
		Type: "now_playing",
		Data: map[string]interface{}{
			"user_id":   userID,
			"track":     track,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Send to user's followers or friends
	s.SendToUser(userID, "now_playing", message.Data)

	// Publish to Redis for scaling
	s.PublishToRedis("now_playing", message)
}

// NotifyLivePlaylist handles live collaborative playlist events
func (s *WebSocketService) NotifyLivePlaylist(playlistID string, userID string, action string, data interface{}) {
	message := WebSocketMessage{
		Type: "live_playlist",
		Data: map[string]interface{}{
			"playlist_id": playlistID,
			"user_id":     userID,
			"action":      action, // "user_joined", "user_left", "track_added", "track_removed", "track_voted"
			"data":        data,
			"timestamp":   time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Broadcast to all users in the playlist room
	s.BroadcastToPlaylistRoom(playlistID, message)

	// Publish to Redis for scaling
	s.PublishToRedis("live_playlists", message)
}

// NotifySystemMessage sends system-wide notifications
func (s *WebSocketService) NotifySystemMessage(messageType string, title string, body string) {
	message := WebSocketMessage{
		Type: "system_notification",
		Data: map[string]interface{}{
			"type":      messageType, // "info", "warning", "error", "success"
			"title":     title,
			"body":      body,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Broadcast to all connected users
	s.SendToAll("system_notification", message.Data)

	// Publish to Redis for scaling
	s.PublishToRedis("system_notifications", message)
}

// Room management for collaborative features

// BroadcastToPlaylistRoom sends messages to all users in a playlist room
func (s *WebSocketService) BroadcastToPlaylistRoom(playlistID string, message WebSocketMessage) {
	s.hub.Mutex.RLock()
	defer s.hub.Mutex.RUnlock()

	for client := range s.hub.Clients {
		// Check if client is in the playlist room (this would need to be tracked)
		// For now, broadcast to all clients
		select {
		case client.Send <- message:
		default:
			// Client is not ready to receive, skip
		}
	}
}

// GetConnectedUsers returns list of currently connected users
func (s *WebSocketService) GetConnectedUsers() []string {
	s.hub.Mutex.RLock()
	defer s.hub.Mutex.RUnlock()

	var users []string
	for userID := range s.hub.UserClients {
		if userID != "" {
			users = append(users, userID)
		}
	}
	return users
}

// GetUserConnectionCount returns the number of connections for a specific user
func (s *WebSocketService) GetUserConnectionCount(userID string) int {
	s.hub.Mutex.RLock()
	defer s.hub.Mutex.RUnlock()

	if clients, exists := s.hub.UserClients[userID]; exists {
		return len(clients)
	}
	return 0
}
