package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gruvit/server/go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CacheService struct {
	tracksCollection  *mongo.Collection
	searchCollection  *mongo.Collection
	artistsCollection *mongo.Collection
	albumsCollection  *mongo.Collection
}

func NewCacheService(db *mongo.Database) *CacheService {
	return &CacheService{
		tracksCollection:  db.Collection("tracks"),
		searchCollection:  db.Collection("search_cache"),
		artistsCollection: db.Collection("artists"),
		albumsCollection:  db.Collection("albums"),
	}
}

// CacheSearchResults stores search results in MongoDB
func (c *CacheService) CacheSearchResults(query string, results []models.Track, expiration time.Duration) error {
	ctx := context.Background()

	// Create search cache document
	cacheDoc := bson.M{
		"query":      query,
		"results":    results,
		"created_at": time.Now(),
		"expires_at": time.Now().Add(expiration),
		"total":      len(results),
	}

	// Upsert search cache
	filter := bson.M{"query": query}
	update := bson.M{"$set": cacheDoc}
	opts := options.Update().SetUpsert(true)

	_, err := c.searchCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to cache search results: %v", err)
	}

	// Cache individual tracks
	for _, track := range results {
		if err := c.CacheTrack(track); err != nil {
			fmt.Printf("Failed to cache track %s: %v\n", track.ID, err)
		}
	}

	return nil
}

// GetCachedSearchResults retrieves cached search results
func (c *CacheService) GetCachedSearchResults(query string) ([]models.Track, error) {
	ctx := context.Background()

	filter := bson.M{
		"query":      query,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	var cacheDoc struct {
		Results []models.Track `bson:"results"`
	}

	err := c.searchCollection.FindOne(ctx, filter).Decode(&cacheDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached search results: %v", err)
	}

	return cacheDoc.Results, nil
}

// CacheTrack stores a track in the tracks collection
func (c *CacheService) CacheTrack(track models.Track) error {
	ctx := context.Background()

	filter := bson.M{
		"id":     track.ID,
		"source": track.Source,
	}

	update := bson.M{
		"$set": bson.M{
			"id":         track.ID,
			"title":      track.Title,
			"artist":     track.Artist,
			"album":      track.Album,
			"duration":   track.Duration,
			"stream_url": track.StreamURL,
			"image_url":  track.ImageURL,
			"genre":      track.Genre,
			"source":     track.Source,
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := c.tracksCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to cache track: %v", err)
	}

	return nil
}

// GetTrack retrieves a track from cache
func (c *CacheService) GetTrack(id, source string) (*models.Track, error) {
	ctx := context.Background()

	filter := bson.M{
		"id":     id,
		"source": source,
	}

	var track models.Track
	err := c.tracksCollection.FindOne(ctx, filter).Decode(&track)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Track not found
		}
		return nil, fmt.Errorf("failed to get track: %v", err)
	}

	return &track, nil
}

// SearchTracks searches tracks in the cache
func (c *CacheService) SearchTracks(query string, limit int) ([]models.Track, error) {
	ctx := context.Background()

	// Create text search index if it doesn't exist
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "artist", Value: "text"},
			{Key: "album", Value: "text"},
		},
	}
	c.tracksCollection.Indexes().CreateOne(ctx, indexModel)

	// Search using text index
	filter := bson.M{
		"$text": bson.M{"$search": query},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := c.tracksCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search tracks: %v", err)
	}
	defer cursor.Close(ctx)

	var tracks []models.Track
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, fmt.Errorf("failed to decode tracks: %v", err)
	}

	return tracks, nil
}

// GetTracksByArtist retrieves tracks by artist from cache
func (c *CacheService) GetTracksByArtist(artist string, limit int) ([]models.Track, error) {
	ctx := context.Background()

	filter := bson.M{
		"artist": bson.M{"$regex": artist, "$options": "i"},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := c.tracksCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracks by artist: %v", err)
	}
	defer cursor.Close(ctx)

	var tracks []models.Track
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, fmt.Errorf("failed to decode tracks: %v", err)
	}

	return tracks, nil
}

// GetTracksByGenre retrieves tracks by genre from cache
func (c *CacheService) GetTracksByGenre(genre string, limit int) ([]models.Track, error) {
	ctx := context.Background()

	filter := bson.M{
		"genre": bson.M{"$regex": genre, "$options": "i"},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := c.tracksCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracks by genre: %v", err)
	}
	defer cursor.Close(ctx)

	var tracks []models.Track
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, fmt.Errorf("failed to decode tracks: %v", err)
	}

	return tracks, nil
}

// GetPopularTracks retrieves popular tracks from cache
func (c *CacheService) GetPopularTracks(limit int) ([]models.Track, error) {
	ctx := context.Background()

	// Sort by creation date (newest first) as a simple popularity metric
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := c.tracksCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular tracks: %v", err)
	}
	defer cursor.Close(ctx)

	var tracks []models.Track
	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, fmt.Errorf("failed to decode tracks: %v", err)
	}

	return tracks, nil
}

// CleanExpiredCache removes expired cache entries
func (c *CacheService) CleanExpiredCache() error {
	ctx := context.Background()

	// Clean expired search cache
	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	}

	_, err := c.searchCollection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to clean expired search cache: %v", err)
	}

	return nil
}

// GetCacheStats returns cache statistics
func (c *CacheService) GetCacheStats() (map[string]interface{}, error) {
	ctx := context.Background()

	stats := make(map[string]interface{})

	// Count tracks by source
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$source",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := c.tracksCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get track counts: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode track counts: %v", err)
	}

	stats["tracks_by_source"] = results

	// Total track count
	totalTracks, err := c.tracksCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total tracks: %v", err)
	}
	stats["total_tracks"] = totalTracks

	// Total search cache entries
	totalSearches, err := c.searchCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count search cache: %v", err)
	}
	stats["total_searches"] = totalSearches

	return stats, nil
}

// Get retrieves a value from cache by key
func (c *CacheService) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result bson.M
	err := c.searchCollection.FindOne(ctx, bson.M{"key": key}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("key not found")
		}
		return "", err
	}

	// Check if expired
	if expiresAt, ok := result["expires_at"].(time.Time); ok {
		if time.Now().After(expiresAt) {
			// Delete expired entry
			c.searchCollection.DeleteOne(ctx, bson.M{"key": key})
			return "", fmt.Errorf("key expired")
		}
	}

	if value, ok := result["value"].(string); ok {
		return value, nil
	}

	return "", fmt.Errorf("invalid value type")
}

// Set stores a value in cache with expiration
func (c *CacheService) Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Serialize value to JSON
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	expiresAt := time.Now().Add(expiration)

	_, err = c.searchCollection.ReplaceOne(
		ctx,
		bson.M{"key": key},
		bson.M{
			"key":        key,
			"value":      string(jsonValue),
			"expires_at": expiresAt,
			"created_at": time.Now(),
		},
		options.Replace().SetUpsert(true),
	)

	return err
}
