package services

import (
	"encoding/json"
	"fmt"
	"gruvit/server/go/models"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type JamendoTrack struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist_name"`
	Album    string `json:"album_name"`
	Duration int    `json:"duration"`
	Audio    string `json:"audio"`
	Image    string `json:"album_image"`
	Genre    string `json:"musicinfo_genres"`
}

type JamendoResponse struct {
	Headers struct {
		Status    string `json:"status"`
		Code      int    `json:"code"`
		ErrorType string `json:"error_type"`
		ErrorMsg  string `json:"error_message"`
	} `json:"headers"`
	Results []JamendoTrack `json:"results"`
}

type MusicBrainzTrack struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ArtistCredit []struct {
		Name string `json:"name"`
	} `json:"artist-credit"`
	Duration int `json:"length"`
	Releases []struct {
		Title string `json:"title"`
	} `json:"releases"`
}

type MusicBrainzResponse struct {
	Recordings []MusicBrainzTrack `json:"recordings"`
}

type ExternalAPIService struct {
	jamendoClient       *http.Client
	musicbrainzClient   *http.Client
	jamendoAPIKey       string
	jamendoClientSecret string
	rateLimiter         *RateLimiter
	retryConfig         RetryConfig
}

type RateLimiter struct {
	jamendoLimiter     *time.Ticker
	musicbrainzLimiter *time.Ticker
}

type APIError struct {
	Service    string
	Message    string
	Code       int
	RetryAfter time.Duration
}

func NewExternalAPIService(jamendoAPIKey, jamendoClientSecret string) *ExternalAPIService {
	return &ExternalAPIService{
		jamendoClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		musicbrainzClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		jamendoAPIKey:       jamendoAPIKey,
		jamendoClientSecret: jamendoClientSecret,
		rateLimiter: &RateLimiter{
			jamendoLimiter:     time.NewTicker(time.Second / 2), // 2 requests per second
			musicbrainzLimiter: time.NewTicker(time.Second),     // 1 request per second
		},
		retryConfig: RetryConfig{
			MaxRetries: 3,
			BaseDelay:  time.Second,
			MaxDelay:   30 * time.Second,
			Multiplier: 2.0,
		},
	}
}

// retryWithBackoff executes a function with exponential backoff retry logic
func (s *ExternalAPIService) retryWithBackoff(operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= s.retryConfig.MaxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if apiErr, ok := err.(*APIError); ok && !apiErr.IsRetryable() {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == s.retryConfig.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(s.retryConfig.BaseDelay) * math.Pow(s.retryConfig.Multiplier, float64(attempt)))
		if delay > s.retryConfig.MaxDelay {
			delay = s.retryConfig.MaxDelay
		}

		// If it's a rate limit error, use the retry-after header
		if apiErr, ok := err.(*APIError); ok && apiErr.Code == 429 {
			if retryAfter := apiErr.GetRetryAfter(); retryAfter > 0 {
				delay = retryAfter
			}
		}

		log.Printf("External API operation failed (attempt %d/%d), retrying in %v: %v",
			attempt+1, s.retryConfig.MaxRetries+1, delay, err)

		time.Sleep(delay)
	}

	return fmt.Errorf("operation failed after %d attempts: %w", s.retryConfig.MaxRetries+1, lastErr)
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s API error: %s (code: %d)", e.Service, e.Message, e.Code)
}

// IsRetryable checks if the error is retryable
func (e *APIError) IsRetryable() bool {
	// Retryable status codes: 429 (rate limit), 5xx (server errors), 408 (timeout)
	return e.Code == 429 || e.Code >= 500 || e.Code == 408
}

// GetRetryAfter returns the retry after duration
func (e *APIError) GetRetryAfter() time.Duration {
	if e.RetryAfter > 0 {
		return e.RetryAfter
	}
	// Default retry after for rate limiting
	if e.Code == 429 {
		return 60 * time.Second
	}
	// Default retry after for server errors
	if e.Code >= 500 {
		return 5 * time.Second
	}
	return 0
}

func (s *ExternalAPIService) SearchTracks(query string, limit int) ([]models.Track, error) {
	var allTracks []models.Track
	var errors []error

	// Search MusicBrainz for metadata (primary source) with retry
	<-s.rateLimiter.musicbrainzLimiter.C // Wait for rate limit
	var musicbrainzTracks []models.Track
	err := s.retryWithBackoff(func() error {
		var err error
		musicbrainzTracks, err = s.searchMusicBrainz(query, limit)
		return err
	})
	if err != nil {
		log.Printf("MusicBrainz search error: %v", err)
		errors = append(errors, err)
	} else {
		allTracks = append(allTracks, musicbrainzTracks...)
	}

	// Search Jamendo for Creative Commons streaming with retry
	<-s.rateLimiter.jamendoLimiter.C // Wait for rate limit
	var jamendoTracks []models.Track
	err = s.retryWithBackoff(func() error {
		var err error
		jamendoTracks, err = s.searchJamendo(query, limit)
		return err
	})
	if err != nil {
		log.Printf("Jamendo search error: %v", err)
		errors = append(errors, err)
	} else {
		allTracks = append(allTracks, jamendoTracks...)
	}

	// If all APIs failed, return an error
	if len(allTracks) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all external APIs failed: %v", errors)
	}

	return allTracks, nil
}

func (s *ExternalAPIService) searchJamendo(query string, limit int) ([]models.Track, error) {
	if s.jamendoAPIKey == "" {
		return nil, &APIError{
			Service: "Jamendo",
			Message: "API key not configured",
			Code:    500,
		}
	}

	url := fmt.Sprintf("https://api.jamendo.com/v3.0/tracks/?client_id=%s&search=%s&limit=%d&format=json&include=musicinfo",
		s.jamendoAPIKey, query, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &APIError{
			Service: "Jamendo",
			Message: fmt.Sprintf("Failed to create request: %v", err),
			Code:    500,
		}
	}

	// Set appropriate headers
	req.Header.Set("User-Agent", "Gruvit/1.0 (https://gruvit.com)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.jamendoClient.Do(req)
	if err != nil {
		return nil, &APIError{
			Service: "Jamendo",
			Message: fmt.Sprintf("Request failed: %v", err),
			Code:    500,
		}
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == 429 {
		retryAfter := 60 * time.Second // Default retry after 1 minute
		if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return nil, &APIError{
			Service:    "Jamendo",
			Message:    "Rate limit exceeded",
			Code:       429,
			RetryAfter: retryAfter,
		}
	}

	// Check for other HTTP errors
	if resp.StatusCode >= 400 {
		return nil, &APIError{
			Service: "Jamendo",
			Message: fmt.Sprintf("HTTP error: %d", resp.StatusCode),
			Code:    resp.StatusCode,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			Service: "Jamendo",
			Message: fmt.Sprintf("Failed to read response: %v", err),
			Code:    500,
		}
	}

	var jamendoResp JamendoResponse
	if err := json.Unmarshal(body, &jamendoResp); err != nil {
		return nil, &APIError{
			Service: "Jamendo",
			Message: fmt.Sprintf("Failed to parse JSON: %v", err),
			Code:    500,
		}
	}

	if jamendoResp.Headers.Code != 0 {
		return nil, &APIError{
			Service: "Jamendo",
			Message: jamendoResp.Headers.ErrorMsg,
			Code:    jamendoResp.Headers.Code,
		}
	}

	var tracks []models.Track
	for _, jamendoTrack := range jamendoResp.Results {
		track := models.Track{
			ID:        jamendoTrack.ID,
			Title:     jamendoTrack.Name,
			Artist:    jamendoTrack.Artist,
			Album:     jamendoTrack.Album,
			Duration:  jamendoTrack.Duration,
			StreamURL: jamendoTrack.Audio,
			ImageURL:  jamendoTrack.Image,
			Genre:     jamendoTrack.Genre,
			Source:    "jamendo",
			UpdatedAt: time.Now(),
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *ExternalAPIService) searchMusicBrainz(query string, limit int) ([]models.Track, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/recording?query=%s&fmt=json&limit=%d", query, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &APIError{
			Service: "MusicBrainz",
			Message: fmt.Sprintf("Failed to create request: %v", err),
			Code:    500,
		}
	}

	req.Header.Set("User-Agent", "Gruvit/1.0 (https://gruvit.com)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.musicbrainzClient.Do(req)
	if err != nil {
		return nil, &APIError{
			Service: "MusicBrainz",
			Message: fmt.Sprintf("Request failed: %v", err),
			Code:    500,
		}
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == 429 {
		retryAfter := 60 * time.Second
		if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return nil, &APIError{
			Service:    "MusicBrainz",
			Message:    "Rate limit exceeded",
			Code:       429,
			RetryAfter: retryAfter,
		}
	}

	// Check for other HTTP errors
	if resp.StatusCode >= 400 {
		return nil, &APIError{
			Service: "MusicBrainz",
			Message: fmt.Sprintf("HTTP error: %d", resp.StatusCode),
			Code:    resp.StatusCode,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			Service: "MusicBrainz",
			Message: fmt.Sprintf("Failed to read response: %v", err),
			Code:    500,
		}
	}

	var musicbrainzResp MusicBrainzResponse
	if err := json.Unmarshal(body, &musicbrainzResp); err != nil {
		return nil, &APIError{
			Service: "MusicBrainz",
			Message: fmt.Sprintf("Failed to parse JSON: %v", err),
			Code:    500,
		}
	}

	var tracks []models.Track
	for _, mbTrack := range musicbrainzResp.Recordings {
		album := ""
		if len(mbTrack.Releases) > 0 {
			album = mbTrack.Releases[0].Title
		}

		artist := ""
		if len(mbTrack.ArtistCredit) > 0 {
			artist = mbTrack.ArtistCredit[0].Name
		}

		track := models.Track{
			ID:        mbTrack.ID,
			Title:     mbTrack.Title,
			Artist:    artist,
			Album:     album,
			Duration:  mbTrack.Duration / 1000, // Convert from milliseconds to seconds
			Source:    "musicbrainz",
			UpdatedAt: time.Now(),
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *ExternalAPIService) GetStreamURL(trackID, source string) (string, error) {
	switch source {
	case "jamendo":
		return s.getJamendoStreamURL(trackID)
	case "musicbrainz":
		return s.getMusicBrainzStreamURL(trackID)
	default:
		return "", fmt.Errorf("unsupported source: %s", source)
	}
}

func (s *ExternalAPIService) getJamendoStreamURL(trackID string) (string, error) {
	// Jamendo provides direct streaming URLs
	// The track ID from search results should already be the correct ID
	apiKey := os.Getenv("JAMENDO_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("JAMENDO_API_KEY not configured")
	}

	// Jamendo streaming URL format
	streamURL := fmt.Sprintf("https://api.jamendo.com/v3.0/tracks/stream?client_id=%s&id=%s", apiKey, trackID)

	// Validate the URL by making a HEAD request
	req, err := http.NewRequest("HEAD", streamURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", os.Getenv("USER_AGENT"))
	req.Header.Set("Accept", "audio/*")

	resp, err := s.jamendoClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to validate stream URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("stream URL validation failed with status: %d", resp.StatusCode)
	}

	return streamURL, nil
}

func (s *ExternalAPIService) getMusicBrainzStreamURL(trackID string) (string, error) {
	// MusicBrainz doesn't provide direct streaming
	// We need to resolve through external services like Spotify, YouTube, etc.
	// For now, we'll return an error indicating this needs external resolution
	return "", fmt.Errorf("MusicBrainz tracks require external streaming service resolution (Spotify, YouTube, etc.)")
}

// SearchAlbum searches for album details by ID
func (s *ExternalAPIService) SearchAlbum(albumID string) (interface{}, error) {
	// For now, return a mock album structure
	// In a real implementation, this would query external APIs
	album := map[string]interface{}{
		"id":           albumID,
		"title":        "Sample Album",
		"artist":       "Sample Artist",
		"release_date": "2023-01-01",
		"genre":        "Pop",
		"tracks":       []map[string]interface{}{},
		"image_url":    "https://via.placeholder.com/300x300",
		"description":  "A sample album for demonstration",
	}
	return album, nil
}

// SearchArtist searches for artist details by ID
func (s *ExternalAPIService) SearchArtist(artistID string) (interface{}, error) {
	// For now, return a mock artist structure
	// In a real implementation, this would query external APIs
	artist := map[string]interface{}{
		"id":          artistID,
		"name":        "Sample Artist",
		"genre":       "Pop",
		"followers":   1000000,
		"image_url":   "https://via.placeholder.com/300x300",
		"description": "A sample artist for demonstration",
		"albums":      []map[string]interface{}{},
		"tracks":      []map[string]interface{}{},
	}
	return artist, nil
}

// SearchTracksByGenre searches for tracks by genre
func (s *ExternalAPIService) SearchTracksByGenre(genre string, limit int) ([]models.Track, error) {
	// For now, return mock tracks
	// In a real implementation, this would query external APIs
	tracks := []models.Track{
		{
			ID:        "genre_track_1",
			Title:     "Sample " + genre + " Track 1",
			Artist:    "Sample Artist",
			Album:     "Sample Album",
			Duration:  180,
			StreamURL: "https://example.com/stream1",
			ImageURL:  "https://via.placeholder.com/300x300",
			Genre:     genre,
			Source:    "external",
		},
		{
			ID:        "genre_track_2",
			Title:     "Sample " + genre + " Track 2",
			Artist:    "Another Artist",
			Album:     "Another Album",
			Duration:  200,
			StreamURL: "https://example.com/stream2",
			ImageURL:  "https://via.placeholder.com/300x300",
			Genre:     genre,
			Source:    "external",
		},
	}
	return tracks, nil
}

// AdvancedSearch performs advanced search with filters
func (s *ExternalAPIService) AdvancedSearch(query string, filters map[string]string, limit int) ([]models.Track, error) {
	// For now, return mock tracks based on filters
	// In a real implementation, this would query external APIs with filters
	tracks := []models.Track{
		{
			ID:        "advanced_track_1",
			Title:     "Advanced Search Result 1",
			Artist:    "Sample Artist",
			Album:     "Sample Album",
			Duration:  180,
			StreamURL: "https://example.com/stream1",
			ImageURL:  "https://via.placeholder.com/300x300",
			Genre:     filters["genre"],
			Source:    "external",
		},
	}
	return tracks, nil
}

// GetTrendingTracks returns trending tracks
func (s *ExternalAPIService) GetTrendingTracks(limit int) ([]models.Track, error) {
	// For now, return mock trending tracks
	// In a real implementation, this would query external APIs for trending data
	tracks := []models.Track{
		{
			ID:        "trending_1",
			Title:     "Trending Track 1",
			Artist:    "Popular Artist",
			Album:     "Hit Album",
			Duration:  180,
			StreamURL: "https://example.com/trending1",
			ImageURL:  "https://via.placeholder.com/300x300",
			Genre:     "Pop",
			Source:    "external",
		},
		{
			ID:        "trending_2",
			Title:     "Trending Track 2",
			Artist:    "Another Popular Artist",
			Album:     "Another Hit Album",
			Duration:  200,
			StreamURL: "https://example.com/trending2",
			ImageURL:  "https://via.placeholder.com/300x300",
			Genre:     "Rock",
			Source:    "external",
		},
	}
	return tracks, nil
}

// GetPopularTracks returns popular tracks
func (s *ExternalAPIService) GetPopularTracks(limit int) ([]models.Track, error) {
	// For now, return mock popular tracks
	// In a real implementation, this would query external APIs for popular data
	tracks := []models.Track{
		{
			ID:        "popular_1",
			Title:     "Popular Track 1",
			Artist:    "Famous Artist",
			Album:     "Best Album",
			Duration:  180,
			StreamURL: "https://example.com/popular1",
			ImageURL:  "https://via.placeholder.com/300x300",
			Genre:     "Pop",
			Source:    "external",
		},
	}
	return tracks, nil
}

// GetDiscoverTracks returns discovery tracks
func (s *ExternalAPIService) GetDiscoverTracks(limit int) ([]models.Track, error) {
	// For now, return mock discovery tracks
	// In a real implementation, this would use recommendation algorithms
	tracks := []models.Track{
		{
			ID:        "discover_1",
			Title:     "Discovery Track 1",
			Artist:    "New Artist",
			Album:     "Fresh Album",
			Duration:  180,
			StreamURL: "https://example.com/discover1",
			ImageURL:  "https://via.placeholder.com/300x300",
			Genre:     "Indie",
			Source:    "external",
		},
	}
	return tracks, nil
}

// GetSimilarArtists returns similar artists
func (s *ExternalAPIService) GetSimilarArtists(artistID string, limit int) ([]models.Artist, error) {
	// For now, return mock similar artists
	// In a real implementation, this would use recommendation algorithms
	artists := []models.Artist{
		{
			ID:         "similar_1",
			Name:       "Similar Artist 1",
			Genre:      "Pop",
			Followers:  500000,
			Image:      "https://via.placeholder.com/300x300",
			IsVerified: true,
		},
		{
			ID:         "similar_2",
			Name:       "Similar Artist 2",
			Genre:      "Pop",
			Followers:  300000,
			Image:      "https://via.placeholder.com/300x300",
			IsVerified: false,
		},
	}
	return artists, nil
}

// GetSimilarTracks returns similar tracks
func (s *ExternalAPIService) GetSimilarTracks(trackID string, limit int) ([]models.Track, error) {
	// For now, return mock similar tracks
	// In a real implementation, this would use recommendation algorithms
	tracks := []models.Track{
		{
			ID:        "similar_track_1",
			Title:     "Similar Track 1",
			Artist:    "Similar Artist",
			Album:     "Similar Album",
			Duration:  180,
			StreamURL: "https://example.com/similar1",
			ImageURL:  "https://via.placeholder.com/300x300",
			Genre:     "Pop",
			Source:    "external",
		},
	}
	return tracks, nil
}

// HealthCheck performs a health check on external APIs
func (s *ExternalAPIService) HealthCheck() error {
	// Test Jamendo API
	if s.jamendoAPIKey != "" {
		testURL := fmt.Sprintf("https://api.jamendo.com/v3.0/tracks/?client_id=%s&limit=1&format=json", s.jamendoAPIKey)
		req, err := http.NewRequest("GET", testURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create Jamendo test request: %w", err)
		}

		resp, err := s.jamendoClient.Do(req)
		if err != nil {
			return fmt.Errorf("Jamendo API health check failed: %w", err)
		}
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("Jamendo API returned error status: %d", resp.StatusCode)
		}
	}

	// Test MusicBrainz API
	testURL := "https://musicbrainz.org/ws/2/recording?query=test&fmt=json&limit=1"
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create MusicBrainz test request: %w", err)
	}
	req.Header.Set("User-Agent", "Gruvit/1.0 (https://gruvit.com)")

	resp, err := s.musicbrainzClient.Do(req)
	if err != nil {
		return fmt.Errorf("MusicBrainz API health check failed: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("MusicBrainz API returned error status: %d", resp.StatusCode)
	}

	return nil
}

// GetServiceStats returns statistics about external API usage
func (s *ExternalAPIService) GetServiceStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["jamendo_api_key_configured"] = s.jamendoAPIKey != ""
	stats["jamendo_client_secret_configured"] = s.jamendoClientSecret != ""
	stats["jamendo_rate_limit"] = "2 requests per second"
	stats["musicbrainz_rate_limit"] = "1 request per second"
	stats["retry_config"] = s.retryConfig
	stats["http_client_timeout"] = s.jamendoClient.Timeout.String()

	return stats
}
