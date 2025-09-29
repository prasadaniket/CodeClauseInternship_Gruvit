package handlers

import (
	"net/http"
	"strconv"
	"time"

	"gruvit/server/go/models"
	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

type MusicSearchHandler struct {
	externalAPIService *services.ExternalAPIService
	redisService       *services.RedisService
	cacheService       *services.CacheService
}

func NewMusicSearchHandler(externalAPIService *services.ExternalAPIService, redisService *services.RedisService, cacheService *services.CacheService) *MusicSearchHandler {
	return &MusicSearchHandler{
		externalAPIService: externalAPIService,
		redisService:       redisService,
		cacheService:       cacheService,
	}
}

// SearchTracks performs comprehensive music search across all APIs
func (h *MusicSearchHandler) SearchTracks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Check Redis cache first
	cachedResponse, err := h.redisService.GetSearchResults(query)
	if err == nil && cachedResponse != nil {
		c.JSON(http.StatusOK, cachedResponse)
		return
	}

	// Check MongoDB cache
	cachedTracks, err := h.cacheService.GetCachedSearchResults(query)
	if err == nil && len(cachedTracks) > 0 {
		response := &models.SearchResponse{
			Query:   query,
			Results: cachedTracks,
			Total:   len(cachedTracks),
			Page:    page,
			Limit:   limit,
		}

		// Cache in Redis for faster access
		h.redisService.SetSearchResults(query, response, time.Hour)

		c.JSON(http.StatusOK, response)
		return
	}

	// Search external APIs
	tracks, err := h.externalAPIService.SearchTracks(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search tracks"})
		return
	}

	// Cache results in MongoDB
	h.cacheService.CacheSearchResults(query, tracks, 24*time.Hour)

	// Create response
	response := &models.SearchResponse{
		Query:   query,
		Results: tracks,
		Total:   len(tracks),
		Page:    page,
		Limit:   limit,
	}

	// Cache in Redis for 1 hour
	h.redisService.SetSearchResults(query, response, time.Hour)

	c.JSON(http.StatusOK, response)
}

// SearchByArtist searches tracks by artist
func (h *MusicSearchHandler) SearchByArtist(c *gin.Context) {
	artist := c.Query("artist")
	if artist == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Artist parameter is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Search in cache first
	tracks, err := h.cacheService.GetTracksByArtist(artist, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search tracks by artist"})
		return
	}

	// If not enough results in cache, search external APIs
	if len(tracks) < limit {
		externalTracks, err := h.externalAPIService.SearchTracks(artist, limit)
		if err == nil {
			// Filter tracks by artist
			for _, track := range externalTracks {
				if len(tracks) >= limit {
					break
				}
				// Simple artist name matching
				if containsIgnoreCase(track.Artist, artist) {
					tracks = append(tracks, track)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"artist":  artist,
		"results": tracks,
		"total":   len(tracks),
	})
}

// SearchByGenre searches tracks by genre
func (h *MusicSearchHandler) SearchByGenre(c *gin.Context) {
	genre := c.Query("genre")
	if genre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Genre parameter is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Search in cache first
	tracks, err := h.cacheService.GetTracksByGenre(genre, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search tracks by genre"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"genre":   genre,
		"results": tracks,
		"total":   len(tracks),
	})
}

// GetTrackDetails returns detailed information about a specific track
func (h *MusicSearchHandler) GetTrackDetails(c *gin.Context) {
	trackID := c.Param("id")
	source := c.Query("source")

	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track ID is required"})
		return
	}

	// Get track from cache
	track, err := h.cacheService.GetTrack(trackID, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get track details"})
		return
	}

	if track == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track not found"})
		return
	}

	c.JSON(http.StatusOK, track)
}

// GetCacheStats returns cache statistics
func (h *MusicSearchHandler) GetCacheStats(c *gin.Context) {
	stats, err := h.cacheService.GetCacheStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cache stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CleanCache cleans expired cache entries
func (h *MusicSearchHandler) CleanCache(c *gin.Context) {
	err := h.cacheService.CleanExpiredCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clean cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cache cleaned successfully"})
}

// GetAlbumDetails returns detailed information about an album
func (h *MusicSearchHandler) GetAlbumDetails(c *gin.Context) {
	albumID := c.Param("albumId")
	if albumID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Album ID is required"})
		return
	}

	// Try to get from cache first
	cacheKey := "album:" + albumID
	cachedAlbum, err := h.cacheService.Get(cacheKey)
	if err == nil && cachedAlbum != "" {
		c.JSON(http.StatusOK, gin.H{"album": cachedAlbum})
		return
	}

	// Search for album across all APIs
	album, err := h.externalAPIService.SearchAlbum(albumID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// Cache the result
	h.cacheService.Set(cacheKey, album, 1*time.Hour)

	c.JSON(http.StatusOK, gin.H{"album": album})
}

// GetArtistDetails returns detailed information about an artist
func (h *MusicSearchHandler) GetArtistDetails(c *gin.Context) {
	artistID := c.Param("artistId")
	if artistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Artist ID is required"})
		return
	}

	// Try to get from cache first
	cacheKey := "artist:" + artistID
	cachedArtist, err := h.cacheService.Get(cacheKey)
	if err == nil && cachedArtist != "" {
		c.JSON(http.StatusOK, gin.H{"artist": cachedArtist})
		return
	}

	// Search for artist across all APIs
	artist, err := h.externalAPIService.SearchArtist(artistID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artist not found"})
		return
	}

	// Cache the result
	h.cacheService.Set(cacheKey, artist, 1*time.Hour)

	c.JSON(http.StatusOK, gin.H{"artist": artist})
}

// GetGenres returns a list of available music genres
func (h *MusicSearchHandler) GetGenres(c *gin.Context) {
	genres := []string{
		"Pop", "Rock", "Hip-Hop", "R&B", "Country", "Electronic", "Jazz", "Classical",
		"Blues", "Folk", "Reggae", "Funk", "Soul", "Alternative", "Indie", "Metal",
		"Punk", "Gospel", "Latin", "World", "Ambient", "Techno", "House", "Trance",
		"Disco", "Funk", "Ska", "Grunge", "New Wave", "Synthpop",
	}

	c.JSON(http.StatusOK, gin.H{"genres": genres})
}

// GetTracksByGenre returns tracks filtered by genre
func (h *MusicSearchHandler) GetTracksByGenre(c *gin.Context) {
	genre := c.Param("genre")
	if genre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Genre is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Try to get from cache first
	cacheKey := "genre:" + genre + ":limit:" + limitStr
	cachedTracks, err := h.cacheService.Get(cacheKey)
	if err == nil && cachedTracks != "" {
		c.JSON(http.StatusOK, gin.H{"tracks": cachedTracks})
		return
	}

	// Search for tracks by genre
	tracks, err := h.externalAPIService.SearchTracksByGenre(genre, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tracks by genre"})
		return
	}

	// Cache the result
	h.cacheService.Set(cacheKey, tracks, 30*time.Minute)

	c.JSON(http.StatusOK, gin.H{"tracks": tracks, "genre": genre})
}

// AdvancedSearch performs advanced search with multiple filters
func (h *MusicSearchHandler) AdvancedSearch(c *gin.Context) {
	query := c.Query("q")
	genre := c.Query("genre")
	year := c.Query("year")
	duration := c.Query("duration")
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Build search filters
	filters := make(map[string]string)
	if genre != "" {
		filters["genre"] = genre
	}
	if year != "" {
		filters["year"] = year
	}
	if duration != "" {
		filters["duration"] = duration
	}

	// Try to get from cache first
	cacheKey := "advanced:" + query + ":" + genre + ":" + year + ":" + duration + ":" + limitStr
	cachedResults, err := h.cacheService.Get(cacheKey)
	if err == nil && cachedResults != "" {
		c.JSON(http.StatusOK, gin.H{"tracks": cachedResults})
		return
	}

	// Perform advanced search
	tracks, err := h.externalAPIService.AdvancedSearch(query, filters, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Advanced search failed"})
		return
	}

	// Cache the result
	h.cacheService.Set(cacheKey, tracks, 15*time.Minute)

	c.JSON(http.StatusOK, gin.H{"tracks": tracks, "filters": filters})
}

// GetTrendingTracks returns currently trending tracks
func (h *MusicSearchHandler) GetTrendingTracks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Try to get from cache first
	cacheKey := "trending:limit:" + limitStr
	cachedTracks, err := h.cacheService.Get(cacheKey)
	if err == nil && cachedTracks != "" {
		c.JSON(http.StatusOK, gin.H{"tracks": cachedTracks})
		return
	}

	// Get trending tracks
	tracks, err := h.externalAPIService.GetTrendingTracks(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending tracks"})
		return
	}

	// Cache the result for 1 hour
	h.cacheService.Set(cacheKey, tracks, 1*time.Hour)

	c.JSON(http.StatusOK, gin.H{"tracks": tracks})
}

// GetPopularTracks returns popular tracks
func (h *MusicSearchHandler) GetPopularTracks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Try to get from cache first
	cacheKey := "popular:limit:" + limitStr
	cachedTracks, err := h.cacheService.Get(cacheKey)
	if err == nil && cachedTracks != "" {
		c.JSON(http.StatusOK, gin.H{"tracks": cachedTracks})
		return
	}

	// Get popular tracks
	tracks, err := h.externalAPIService.GetPopularTracks(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular tracks"})
		return
	}

	// Cache the result for 2 hours
	h.cacheService.Set(cacheKey, tracks, 2*time.Hour)

	c.JSON(http.StatusOK, gin.H{"tracks": tracks})
}

// GetDiscoverTracks returns personalized discovery tracks
func (h *MusicSearchHandler) GetDiscoverTracks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Get discovery tracks
	tracks, err := h.externalAPIService.GetDiscoverTracks(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get discovery tracks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tracks": tracks})
}

// GetSimilarArtists returns artists similar to the given artist
func (h *MusicSearchHandler) GetSimilarArtists(c *gin.Context) {
	artistID := c.Param("artistId")
	if artistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Artist ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Try to get from cache first
	cacheKey := "similar_artists:" + artistID + ":limit:" + limitStr
	cachedArtists, err := h.cacheService.Get(cacheKey)
	if err == nil && cachedArtists != "" {
		c.JSON(http.StatusOK, gin.H{"artists": cachedArtists})
		return
	}

	// Get similar artists
	artists, err := h.externalAPIService.GetSimilarArtists(artistID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get similar artists"})
		return
	}

	// Cache the result for 1 hour
	h.cacheService.Set(cacheKey, artists, 1*time.Hour)

	c.JSON(http.StatusOK, gin.H{"artists": artists})
}

// GetSimilarTracks returns tracks similar to the given track
func (h *MusicSearchHandler) GetSimilarTracks(c *gin.Context) {
	trackID := c.Param("trackId")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// Try to get from cache first
	cacheKey := "similar_tracks:" + trackID + ":limit:" + limitStr
	cachedTracks, err := h.cacheService.Get(cacheKey)
	if err == nil && cachedTracks != "" {
		c.JSON(http.StatusOK, gin.H{"tracks": cachedTracks})
		return
	}

	// Get similar tracks
	tracks, err := h.externalAPIService.GetSimilarTracks(trackID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get similar tracks"})
		return
	}

	// Cache the result for 1 hour
	h.cacheService.Set(cacheKey, tracks, 1*time.Hour)

	c.JSON(http.StatusOK, gin.H{"tracks": tracks})
}

// Helper function for case-insensitive string matching
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr))
}
