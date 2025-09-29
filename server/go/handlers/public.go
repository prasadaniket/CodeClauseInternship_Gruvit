package handlers

import (
	"net/http"
	"strconv"

	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

type PublicHandler struct {
	playlistService *services.PlaylistService
	cacheService    *services.CacheService
}

func NewPublicHandler(playlistService *services.PlaylistService, cacheService *services.CacheService) *PublicHandler {
	return &PublicHandler{
		playlistService: playlistService,
		cacheService:    cacheService,
	}
}

// GetPublicPlaylists returns public/featured playlists
func (h *PublicHandler) GetPublicPlaylists(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Mock data for public playlists
	// In a real implementation, you would query public playlists from database
	playlists := []gin.H{
		{
			"id":          "playlist1",
			"name":        "Today's Top Hits",
			"description": "The most played songs right now",
			"image":       "https://via.placeholder.com/300x300",
			"trackCount":  50,
			"isPublic":    true,
			"createdBy":   "Spotify",
			"createdAt":   "2024-01-01T00:00:00Z",
		},
		{
			"id":          "playlist2",
			"name":        "Rock Classics",
			"description": "The greatest rock songs of all time",
			"image":       "https://via.placeholder.com/300x300",
			"trackCount":  75,
			"isPublic":    true,
			"createdBy":   "Music Lover",
			"createdAt":   "2024-01-15T00:00:00Z",
		},
		{
			"id":          "playlist3",
			"name":        "Chill Vibes",
			"description": "Relaxing music for any time",
			"image":       "https://via.placeholder.com/300x300",
			"trackCount":  40,
			"isPublic":    true,
			"createdBy":   "Chill Master",
			"createdAt":   "2024-02-01T00:00:00Z",
		},
		{
			"id":          "playlist4",
			"name":        "Workout Mix",
			"description": "High energy songs to keep you motivated",
			"image":       "https://via.placeholder.com/300x300",
			"trackCount":  60,
			"isPublic":    true,
			"createdBy":   "Fitness Fan",
			"createdAt":   "2024-02-10T00:00:00Z",
		},
	}

	// Calculate pagination
	offset := (page - 1) * limit
	total := len(playlists)

	// Apply pagination
	var paginatedPlaylists []gin.H
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedPlaylists = playlists[offset:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"playlists": paginatedPlaylists,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}
