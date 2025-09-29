package handlers

import (
	"net/http"

	"gruvit/server/go/models"
	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

type PlaylistHandler struct {
	playlistService *services.PlaylistService
}

func NewPlaylistHandler(playlistService *services.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{
		playlistService: playlistService,
	}
}

func (h *PlaylistHandler) CreatePlaylist(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlist, err := h.playlistService.CreatePlaylist(userID, req.Name, req.Description, req.IsPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
		return
	}

	c.JSON(http.StatusCreated, playlist)
}

func (h *PlaylistHandler) GetUserPlaylists(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	playlists, err := h.playlistService.GetPlaylistsByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get playlists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	playlistID := c.Param("id")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	playlist, err := h.playlistService.GetPlaylistByID(playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check if user owns the playlist or if it's public
	userID := c.GetString("user_id")
	if userID != "" && playlist.Owner == userID {
		// User owns the playlist
		c.JSON(http.StatusOK, playlist)
		return
	}

	if playlist.IsPublic {
		// Public playlist
		c.JSON(http.StatusOK, playlist)
		return
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
}

func (h *PlaylistHandler) UpdatePlaylist(c *gin.Context) {
	playlistID := c.Param("id")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlist, err := h.playlistService.UpdatePlaylist(playlistID, userID, req.Name, req.Description, req.IsPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update playlist"})
		return
	}

	c.JSON(http.StatusOK, playlist)
}

func (h *PlaylistHandler) DeletePlaylist(c *gin.Context) {
	playlistID := c.Param("id")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := h.playlistService.DeletePlaylist(playlistID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist deleted successfully"})
}

func (h *PlaylistHandler) AddTrackToPlaylist(c *gin.Context) {
	playlistID := c.Param("id")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var track models.Track
	if err := c.ShouldBindJSON(&track); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.playlistService.AddTrackToPlaylist(playlistID, userID, track)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add track to playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track added to playlist successfully"})
}

func (h *PlaylistHandler) RemoveTrackFromPlaylist(c *gin.Context) {
	playlistID := c.Param("id")
	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Playlist ID is required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	trackExternalID := c.Query("track_id")
	if trackExternalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track ID is required"})
		return
	}

	err := h.playlistService.RemoveTrackFromPlaylist(playlistID, userID, trackExternalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove track from playlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track removed from playlist successfully"})
}
