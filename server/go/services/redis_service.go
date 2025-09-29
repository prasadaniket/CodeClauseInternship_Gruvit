package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gruvit/server/go/models"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(addr, password string, db int) *RedisService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisService{
		client: rdb,
	}
}

func NewNilRedisService() *RedisService {
	return &RedisService{
		client: nil,
	}
}

func (r *RedisService) GetSearchResults(query string) (*models.SearchResponse, error) {
	if r.client == nil {
		return nil, nil // Redis not available, treat as cache miss
	}

	ctx := context.Background()
	key := fmt.Sprintf("search:%s", query)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var response models.SearchResponse
	if err := json.Unmarshal([]byte(val), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (r *RedisService) SetSearchResults(query string, response *models.SearchResponse, expiration time.Duration) error {
	if r.client == nil {
		return nil // Redis not available, silently skip caching
	}

	ctx := context.Background()
	key := fmt.Sprintf("search:%s", query)

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *RedisService) GetStreamURL(trackID string) (string, error) {
	if r.client == nil {
		return "", nil // Redis not available, treat as cache miss
	}

	ctx := context.Background()
	key := fmt.Sprintf("stream:%s", trackID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Cache miss
		}
		return "", err
	}

	return val, nil
}

func (r *RedisService) SetStreamURL(trackID, streamURL string, expiration time.Duration) error {
	if r.client == nil {
		return nil // Redis not available, silently skip caching
	}

	ctx := context.Background()
	key := fmt.Sprintf("stream:%s", trackID)

	return r.client.Set(ctx, key, streamURL, expiration).Err()
}

func (r *RedisService) Close() error {
	if r.client == nil {
		return nil // Redis not available
	}
	return r.client.Close()
}
