package handlers

import (
	"net/http"

	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	wsService  *services.WebSocketService
	authClient *services.AuthClient
}

func NewWebSocketHandler(wsService *services.WebSocketService, authClient *services.AuthClient) *WebSocketHandler {
	return &WebSocketHandler{
		wsService:  wsService,
		authClient: authClient,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Extract user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	services.ServeWebSocket(h.wsService.GetHub(), c.Writer, c.Request, userIDStr)
}

// HandlePublicWebSocket handles public WebSocket connections (no auth required)
func (h *WebSocketHandler) HandlePublicWebSocket(c *gin.Context) {
	// For public connections, use empty user ID
	services.ServeWebSocket(h.wsService.GetHub(), c.Writer, c.Request, "")
}

// GetWebSocketStats returns WebSocket connection statistics
func (h *WebSocketHandler) GetWebSocketStats(c *gin.Context) {
	hub := h.wsService.GetHub()

	hub.Mutex.RLock()
	totalClients := len(hub.Clients)
	totalUsers := len(hub.UserClients)
	hub.Mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"total_clients": totalClients,
		"total_users":   totalUsers,
		"status":        "active",
	})
}
