package handlers

import (
	"net/http"
	"strconv"

	"gruvit/server/go/models"
	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	playlistService *services.PlaylistService
	cacheService    *services.CacheService
	userService     *services.UserService
}

func NewUserHandler(playlistService *services.PlaylistService, cacheService *services.CacheService, userService *services.UserService) *UserHandler {
	return &UserHandler{
		playlistService: playlistService,
		cacheService:    cacheService,
		userService:     userService,
	}
}

// GetUserFavorites returns user's favorite tracks
func (h *UserHandler) GetUserFavorites(c *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	// Get user's favorite tracks from database
	favorites, err := h.userService.GetUserFavorites(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user favorites"})
		return
	}

	// Convert to the expected format
	var tracks []gin.H
	for _, favorite := range favorites {
		tracks = append(tracks, gin.H{
			"id":         favorite.Track.ID,
			"title":      favorite.Track.Title,
			"artist":     favorite.Track.Artist,
			"album":      favorite.Track.Album,
			"stream_url": favorite.Track.StreamURL,
			"source":     favorite.Track.Source,
			"added_at":   favorite.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"favorites": tracks,
		"total":     len(tracks),
	})
}

// GetUserListeningHistory returns user's listening history
func (h *UserHandler) GetUserListeningHistory(c *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	// Get user's listening history from database
	history, err := h.userService.GetUserListeningHistory(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get listening history"})
		return
	}

	// Convert to the expected format
	var tracks []gin.H
	for _, entry := range history {
		tracks = append(tracks, gin.H{
			"id":         entry.Track.ID,
			"title":      entry.Track.Title,
			"artist":     entry.Track.Artist,
			"album":      entry.Track.Album,
			"duration":   entry.Track.Duration,
			"stream_url": entry.Track.StreamURL,
			"image_url":  entry.Track.ImageURL,
			"genre":      entry.Track.Genre,
			"source":     entry.Track.Source,
			"played_at":  entry.PlayedAt,
			"play_count": 1, // Default to 1 if PlayCount field doesn't exist
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"history": tracks,
		"total":   len(tracks),
	})
}

// AddToFavorites adds a track to user's favorites
func (h *UserHandler) AddToFavorites(c *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	trackID := c.Param("trackId")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track ID is required"})
		return
	}

	// Add track to user's favorites
	// Create a basic track object with the trackID
	track := models.Track{ID: trackID}
	err := h.userService.AddToFavorites(userID.(string), track)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Track added to favorites",
		"track_id": trackID,
	})
}

// RemoveFromFavorites removes a track from user's favorites
func (h *UserHandler) RemoveFromFavorites(c *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	trackID := c.Param("trackId")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track ID is required"})
		return
	}

	// Remove track from user's favorites
	err := h.userService.RemoveFromFavorites(userID.(string), trackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Track removed from favorites",
		"track_id": trackID,
	})
}

// RecordPlay records a track play in user's listening history
func (h *UserHandler) RecordPlay(c *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		TrackID  string `json:"track_id" binding:"required"`
		Duration int    `json:"duration,omitempty"` // How long the track was played
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Record the play in user's listening history
	err := h.userService.RecordPlay(userID.(string), request.TrackID, request.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record play"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Play recorded",
		"track_id": request.TrackID,
	})
}

// GetUserTopArtists returns user's top artists
func (h *UserHandler) GetUserTopArtists(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	// Get user's top artists from listening history
	artists, err := h.userService.GetUserTopArtists(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user top artists"})
		return
	}

	// Convert to the expected format
	var artistList []gin.H
	for _, artist := range artists {
		artistList = append(artistList, gin.H{
			"id":         artist.ID,
			"name":       artist.Name,
			"image":      artist.Image,
			"genre":      artist.Genre,
			"followers":  artist.Followers,
			"isVerified": artist.IsVerified,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"artists": artistList,
		"user_id": userID,
	})
}

// GetUserTopTracks returns user's top tracks
func (h *UserHandler) GetUserTopTracks(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	// Get user's top tracks from listening history
	tracks, err := h.userService.GetUserTopTracks(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user top tracks"})
		return
	}

	// Convert to the expected format
	var trackList []gin.H
	for _, track := range tracks {
		trackList = append(trackList, gin.H{
			"id":         track.ID,
			"title":      track.Title,
			"artist":     track.Artist,
			"album":      track.Album,
			"image":      track.ImageURL,
			"duration":   track.Duration,
			"stream_url": track.StreamURL,
			"source":     track.Source,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tracks":  trackList,
		"user_id": userID,
	})
}

// GetUserFollowings returns artists the user follows
func (h *UserHandler) GetUserFollowings(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Get user's followings from database
	followings, err := h.userService.GetUserFollowings(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user followings"})
		return
	}

	// Convert to the expected format
	var followingList []gin.H
	for _, follow := range followings {
		followingList = append(followingList, gin.H{
			"id":          follow.FollowedID,
			"type":        follow.Type,
			"followed_at": follow.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"followings": followingList,
		"user_id":    userID,
	})
}

// FollowArtist allows user to follow an artist
func (h *UserHandler) FollowArtist(c *gin.Context) {
	artistID := c.Param("artistId")
	if artistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Artist ID is required"})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Add follow relationship to database
	err := h.userService.FollowArtist(userID.(string), artistID)
	if err != nil {
		if err.Error() == "already following this artist" {
			c.JSON(http.StatusConflict, gin.H{"error": "Already following this artist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow artist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully followed artist",
		"artist_id": artistID,
		"user_id":   userID,
	})
}

// UnfollowArtist allows user to unfollow an artist
func (h *UserHandler) UnfollowArtist(c *gin.Context) {
	artistID := c.Param("artistId")
	if artistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Artist ID is required"})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Remove follow relationship from database
	err := h.userService.UnfollowArtist(userID.(string), artistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfollow artist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully unfollowed artist",
		"artist_id": artistID,
		"user_id":   userID,
	})
}

// GetUserFollowers returns user's followers
func (h *UserHandler) GetUserFollowers(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Get user's followers from database
	followers, err := h.userService.GetUserFollowers(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user followers"})
		return
	}

	// Convert to the expected format
	var followerList []gin.H
	for _, follow := range followers {
		followerList = append(followerList, gin.H{
			"id":          follow.UserID,
			"followed_at": follow.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"followers": followerList,
		"user_id":   userID,
	})
}
