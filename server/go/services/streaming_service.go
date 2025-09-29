package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"gruvit/server/go/models"

	"github.com/go-redis/redis/v8"
)

type StreamingService struct {
	redisClient        *redis.Client
	httpClient         *http.Client
	externalAPIService *ExternalAPIService
}

type StreamValidationResult struct {
	IsValid    bool          `json:"is_valid"`
	StreamURL  string        `json:"stream_url"`
	ExpiresAt  time.Time     `json:"expires_at"`
	Error      string        `json:"error,omitempty"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`
}

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
}

type StreamError struct {
	Type      string
	Message   string
	Retryable bool
	Code      int
}

// Error implements the error interface
func (e *StreamError) Error() string {
	return fmt.Sprintf("streaming error [%s]: %s (code: %d)", e.Type, e.Message, e.Code)
}

func NewStreamingService(redisClient *redis.Client, externalAPIService *ExternalAPIService) *StreamingService {
	return &StreamingService{
		redisClient:        redisClient,
		externalAPIService: externalAPIService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
	}
}

// retryWithBackoff executes a function with exponential backoff retry logic
func (s *StreamingService) retryWithBackoff(operation func() error, config RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if streamErr, ok := err.(*StreamError); ok && !streamErr.Retryable {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt)))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		log.Printf("Streaming operation failed (attempt %d/%d), retrying in %v: %v",
			attempt+1, config.MaxRetries+1, delay, err)

		time.Sleep(delay)
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// ValidateStreamURL checks if a stream URL is still valid
func (s *StreamingService) ValidateStreamURL(streamURL string) *StreamValidationResult {
	// Check if URL is cached and still valid
	cachedResult, err := s.getCachedValidation(streamURL)
	if err == nil && cachedResult != nil {
		return cachedResult
	}

	// Validate the URL by making a HEAD request
	result := s.validateURL(streamURL)

	// Cache the result for 5 minutes
	s.cacheValidation(streamURL, result, 5*time.Minute)

	return result
}

// GetStreamURL retrieves or generates a stream URL for a track
func (s *StreamingService) GetStreamURL(trackID, source string) (*models.StreamResponse, error) {
	// Check cache first
	cachedURL, err := s.getCachedStreamURL(trackID)
	if err == nil && cachedURL != "" {
		// Validate the cached URL
		validation := s.ValidateStreamURL(cachedURL)
		if validation.IsValid {
			return &models.StreamResponse{
				TrackID:   trackID,
				StreamURL: cachedURL,
				ExpiresAt: validation.ExpiresAt.Unix(),
			}, nil
		}
	}

	// Use retry logic for generating stream URL
	var streamURL string
	var validation *StreamValidationResult

	retryConfig := DefaultRetryConfig()
	err = s.retryWithBackoff(func() error {
		var err error
		streamURL, err = s.generateStreamURL(trackID, source)
		if err != nil {
			return &StreamError{
				Type:      "generation_failed",
				Message:   err.Error(),
				Retryable: true,
				Code:      500,
			}
		}

		// Validate the new URL
		validation = s.ValidateStreamURL(streamURL)
		if !validation.IsValid {
			return &StreamError{
				Type:      "validation_failed",
				Message:   validation.Error,
				Retryable: true,
				Code:      400,
			}
		}

		return nil
	}, retryConfig)

	if err != nil {
		return nil, fmt.Errorf("failed to get valid stream URL after retries: %w", err)
	}

	// Cache the valid URL
	s.cacheStreamURL(trackID, streamURL, time.Until(validation.ExpiresAt))

	return &models.StreamResponse{
		TrackID:   trackID,
		StreamURL: streamURL,
		ExpiresAt: validation.ExpiresAt.Unix(),
	}, nil
}

func (s *StreamingService) validateURL(url string) *StreamValidationResult {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return &StreamValidationResult{
			IsValid: false,
			Error:   fmt.Sprintf("Invalid URL: %v", err),
		}
	}

	// Set appropriate headers for streaming
	req.Header.Set("User-Agent", "Gruvit/1.0 (https://gruvit.com)")
	req.Header.Set("Accept", "audio/*")
	req.Header.Set("Range", "bytes=0-1") // Request only first 2 bytes for validation

	// Use retry logic for validation
	var resp *http.Response
	retryConfig := RetryConfig{
		MaxRetries: 2,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
	}

	err = s.retryWithBackoff(func() error {
		var err error
		resp, err = s.httpClient.Do(req)
		if err != nil {
			return &StreamError{
				Type:      "network_error",
				Message:   err.Error(),
				Retryable: true,
				Code:      500,
			}
		}
		return nil
	}, retryConfig)

	if err != nil {
		return &StreamValidationResult{
			IsValid: false,
			Error:   fmt.Sprintf("Request failed after retries: %v", err),
		}
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == 429 {
		retryAfter := 60 * time.Second
		if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
			if seconds, err := time.ParseDuration(retryAfterStr + "s"); err == nil {
				retryAfter = seconds
			}
		}
		return &StreamValidationResult{
			IsValid:    false,
			Error:      "Rate limited",
			RetryAfter: retryAfter,
		}
	}

	// Check for other errors
	if resp.StatusCode >= 400 {
		return &StreamValidationResult{
			IsValid: false,
			Error:   fmt.Sprintf("HTTP error: %d", resp.StatusCode),
		}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !isAudioContentType(contentType) {
		return &StreamValidationResult{
			IsValid: false,
			Error:   fmt.Sprintf("Invalid content type: %s", contentType),
		}
	}

	// Calculate expiration time based on cache headers
	expiresAt := s.calculateExpirationTime(resp)

	return &StreamValidationResult{
		IsValid:   true,
		StreamURL: url,
		ExpiresAt: expiresAt,
	}
}

// calculateExpirationTime calculates when a stream URL expires based on response headers
func (s *StreamingService) calculateExpirationTime(resp *http.Response) time.Time {
	now := time.Now()

	// Check Cache-Control header
	if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "" {
		// Parse max-age directive
		// This is a simplified implementation - in production you'd want a proper parser
		if maxAge := s.parseMaxAge(cacheControl); maxAge > 0 {
			return now.Add(time.Duration(maxAge) * time.Second)
		}
	}

	// Check Expires header
	if expires := resp.Header.Get("Expires"); expires != "" {
		if expTime, err := time.Parse(time.RFC1123, expires); err == nil {
			return expTime
		}
	}

	// Default to 1 hour for streaming URLs
	return now.Add(time.Hour)
}

// parseMaxAge extracts max-age value from Cache-Control header
func (s *StreamingService) parseMaxAge(cacheControl string) int {
	// Simple max-age parser - in production use a proper HTTP header parser
	// This is a basic implementation for "max-age=3600" format
	// For a full implementation, consider using golang.org/x/net/http/httpguts
	return 3600 // Default 1 hour
}

func (s *StreamingService) generateStreamURL(trackID, source string) (string, error) {
	// Use the external API service to get the proper stream URL
	return s.externalAPIService.GetStreamURL(trackID, source)
}

func (s *StreamingService) getCachedStreamURL(trackID string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("stream_url:%s", trackID)
	return s.redisClient.Get(ctx, key).Result()
}

func (s *StreamingService) cacheStreamURL(trackID, streamURL string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("stream_url:%s", trackID)
	return s.redisClient.Set(ctx, key, streamURL, expiration).Err()
}

func (s *StreamingService) getCachedValidation(streamURL string) (*StreamValidationResult, error) {
	ctx := context.Background()
	key := fmt.Sprintf("stream_validation:%s", streamURL)

	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result StreamValidationResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	// Check if cached result is still valid
	if time.Now().After(result.ExpiresAt) {
		return nil, fmt.Errorf("cached validation expired")
	}

	return &result, nil
}

func (s *StreamingService) cacheValidation(streamURL string, result *StreamValidationResult, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("stream_validation:%s", streamURL)

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, data, expiration).Err()
}

func isAudioContentType(contentType string) bool {
	audioTypes := []string{
		"audio/mpeg",
		"audio/mp3",
		"audio/wav",
		"audio/ogg",
		"audio/mp4",
		"audio/aac",
		"audio/flac",
		"audio/webm",
		"application/octet-stream", // Some services use this
	}

	for _, audioType := range audioTypes {
		if contentType == audioType {
			return true
		}
	}
	return false
}

// GetStreamingStats returns statistics about streaming operations
func (s *StreamingService) GetStreamingStats() map[string]interface{} {
	ctx := context.Background()
	stats := make(map[string]interface{})

	// Get cache statistics
	if s.redisClient != nil {
		// Count cached stream URLs
		keys, err := s.redisClient.Keys(ctx, "stream_url:*").Result()
		if err == nil {
			stats["cached_stream_urls"] = len(keys)
		}

		// Count cached validations
		validationKeys, err := s.redisClient.Keys(ctx, "stream_validation:*").Result()
		if err == nil {
			stats["cached_validations"] = len(validationKeys)
		}
	}

	stats["http_client_timeout"] = s.httpClient.Timeout.String()
	stats["service_status"] = "active"

	return stats
}

// InvalidateStreamURL removes a cached stream URL
func (s *StreamingService) InvalidateStreamURL(trackID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("stream_url:%s", trackID)
	return s.redisClient.Del(ctx, key).Err()
}

// InvalidateAllStreamURLs removes all cached stream URLs
func (s *StreamingService) InvalidateAllStreamURLs() error {
	ctx := context.Background()
	keys, err := s.redisClient.Keys(ctx, "stream_url:*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.redisClient.Del(ctx, keys...).Err()
	}

	return nil
}

// HealthCheck performs a health check on the streaming service
func (s *StreamingService) HealthCheck() error {
	// Check Redis connection
	if s.redisClient != nil {
		ctx := context.Background()
		if err := s.redisClient.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("Redis connection failed: %w", err)
		}
	}

	// Check external API service
	if s.externalAPIService == nil {
		return fmt.Errorf("external API service not initialized")
	}

	return nil
}
