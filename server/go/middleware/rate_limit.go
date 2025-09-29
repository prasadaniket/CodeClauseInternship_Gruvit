package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimitMiddleware struct {
	redisClient *redis.Client
}

func NewRateLimitMiddleware(redisClient *redis.Client) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: redisClient,
	}
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Requests int                       // Number of requests allowed
	Window   time.Duration             // Time window
	KeyFunc  func(*gin.Context) string // Function to generate rate limit key
}

// RateLimit creates a rate limiting middleware
func (r *RateLimitMiddleware) RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := config.KeyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		// Use sliding window rate limiting
		allowed, err := r.slidingWindowRateLimit(key, config.Requests, config.Window)
		if err != nil {
			// If Redis is down, allow the request but log the error
			fmt.Printf("Rate limit check failed: %v\n", err)
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": config.Window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// slidingWindowRateLimit implements sliding window rate limiting using Redis
func (r *RateLimitMiddleware) slidingWindowRateLimit(key string, requests int, window time.Duration) (bool, error) {
	ctx := context.Background()
	now := time.Now()
	windowStart := now.Add(-window)

	// Use Redis pipeline for atomic operations
	pipe := r.redisClient.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.Unix()))

	// Count current requests in window
	pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.Unix()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set expiration
	pipe.Expire(ctx, key, window)

	// Execute pipeline
	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// Get count result (second command in pipeline)
	countResult := results[1].(*redis.IntCmd)
	currentCount, err := countResult.Result()
	if err != nil {
		return false, err
	}

	return currentCount <= int64(requests), nil
}

// Default rate limit configurations
func (r *RateLimitMiddleware) DefaultConfigs() map[string]RateLimitConfig {
	return map[string]RateLimitConfig{
		"search": {
			Requests: 60,          // 60 requests
			Window:   time.Minute, // per minute
			KeyFunc: func(c *gin.Context) string {
				clientIP := c.ClientIP()
				return fmt.Sprintf("rate_limit:search:%s", clientIP)
			},
		},
		"stream": {
			Requests: 30,          // 30 requests
			Window:   time.Minute, // per minute
			KeyFunc: func(c *gin.Context) string {
				clientIP := c.ClientIP()
				return fmt.Sprintf("rate_limit:stream:%s", clientIP)
			},
		},
		"playlist": {
			Requests: 20,          // 20 requests
			Window:   time.Minute, // per minute
			KeyFunc: func(c *gin.Context) string {
				userID := c.GetString("user_id")
				if userID == "" {
					return ""
				}
				return fmt.Sprintf("rate_limit:playlist:%s", userID)
			},
		},
		"auth": {
			Requests: 10,          // 10 requests
			Window:   time.Minute, // per minute
			KeyFunc: func(c *gin.Context) string {
				clientIP := c.ClientIP()
				return fmt.Sprintf("rate_limit:auth:%s", clientIP)
			},
		},
	}
}

// External API rate limiting for external service calls
func (r *RateLimitMiddleware) ExternalAPIRateLimit(service string, requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("external_api:%s", service)

		allowed, err := r.slidingWindowRateLimit(key, requests, window)
		if err != nil {
			fmt.Printf("External API rate limit check failed: %v\n", err)
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":       "External API rate limit exceeded, please try again later",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
