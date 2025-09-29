package handlers

import (
	"net/http"
	"strconv"
	"time"

	"gruvit/server/go/models"
	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

type PlaylistEnhancedHandler struct {
	playlistService *services.PlaylistEnhancedService
	userService     *services.UserService
}

func NewPlaylistEnhancedHandler(playlistService *services.PlaylistEnhancedService, userService *services.UserService) *PlaylistEnhancedHandler {
	return &PlaylistEnhancedHandler{
		playlistService: playlistService,
		userService:     userService,
	}
}

// CreateCollaborativePlaylist creates a new collaborative playlist
func (h *PlaylistEnhancedHandler) CreateCollaborativePlaylist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	playlist, err := h.playlistService.CreateCollaborativePlaylist(
		userID.(string),
		request.Title,
		request.Description,
		request.IsPublic,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"playlist": playlist})
}

// AddCollaborator adds a user as a collaborator to a playlist
func (h *PlaylistEnhancedHandler) AddCollaborator(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	var request struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.playlistService.AddCollaborator(playlistID, request.UserID, userID.(string), request.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collaborator added successfully"})
}

// RemoveCollaborator removes a user from playlist collaborators
func (h *PlaylistEnhancedHandler) RemoveCollaborator(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	collaboratorID := c.Param("userId")
	if collaboratorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	err := h.playlistService.RemoveCollaborator(playlistID, collaboratorID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collaborator removed successfully"})
}

// GetCollaborators returns all collaborators for a playlist
func (h *PlaylistEnhancedHandler) GetCollaborators(c *gin.Context) {
	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	collaborators, err := h.playlistService.GetCollaborators(playlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get collaborators"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"collaborators": collaborators})
}

// AddTrackToPlaylist adds a track to a playlist
func (h *PlaylistEnhancedHandler) AddTrackToPlaylist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	var request struct {
		TrackID string `json:"track_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.playlistService.AddTrackToPlaylist(playlistID, request.TrackID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track added to playlist successfully"})
}

// RemoveTrackFromPlaylist removes a track from a playlist
func (h *PlaylistEnhancedHandler) RemoveTrackFromPlaylist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	trackID := c.Param("trackId")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track ID is required"})
		return
	}

	err := h.playlistService.RemoveTrackFromPlaylist(playlistID, trackID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track removed from playlist successfully"})
}

// FollowPlaylist makes a user follow a playlist
func (h *PlaylistEnhancedHandler) FollowPlaylist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	err := h.playlistService.FollowPlaylist(playlistID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist followed successfully"})
}

// UnfollowPlaylist makes a user unfollow a playlist
func (h *PlaylistEnhancedHandler) UnfollowPlaylist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	err := h.playlistService.UnfollowPlaylist(playlistID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist unfollowed successfully"})
}

// LikePlaylist makes a user like a playlist
func (h *PlaylistEnhancedHandler) LikePlaylist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	err := h.playlistService.LikePlaylist(playlistID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist liked successfully"})
}

// UnlikePlaylist makes a user unlike a playlist
func (h *PlaylistEnhancedHandler) UnlikePlaylist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	err := h.playlistService.UnlikePlaylist(playlistID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist unliked successfully"})
}

// SharePlaylist creates a share link for a playlist
func (h *PlaylistEnhancedHandler) SharePlaylist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	var request struct {
		ShareType string `json:"share_type" binding:"required"`
		ExpiresIn int    `json:"expires_in"` // Hours until expiration
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var expiresAt *time.Time
	if request.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(request.ExpiresIn) * time.Hour)
		expiresAt = &exp
	}

	share, err := h.playlistService.SharePlaylist(playlistID, userID.(string), request.ShareType, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"share":     share,
		"share_url": "/playlist/shared/" + share.ShareToken,
	})
}

// GetPlaylistByShareToken retrieves a playlist by share token
func (h *PlaylistEnhancedHandler) GetPlaylistByShareToken(c *gin.Context) {
	shareToken := c.Param("token")
	if shareToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share token is required"})
		return
	}

	playlist, err := h.playlistService.GetPlaylistByShareToken(shareToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found or share link expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlist": playlist})
}

// GetPlaylistRecommendations returns playlist recommendations for a user
func (h *PlaylistEnhancedHandler) GetPlaylistRecommendations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	playlists, err := h.playlistService.GetPlaylistRecommendations(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

// GetUserPlaylists returns playlists for a user
func (h *PlaylistEnhancedHandler) GetUserPlaylists(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	playlists, err := h.playlistService.GetUserPlaylists(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get playlists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

// GetFollowedPlaylists returns playlists followed by a user
func (h *PlaylistEnhancedHandler) GetFollowedPlaylists(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	playlists, err := h.playlistService.GetFollowedPlaylists(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get followed playlists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

// GetPublicPlaylists returns public playlists
func (h *PlaylistEnhancedHandler) GetPublicPlaylists(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// This would need to be implemented in the service
	// For now, return empty array
	c.JSON(http.StatusOK, gin.H{"playlists": []models.Playlist{}})
}
