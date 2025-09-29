package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"gruvit/server/go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RecommendationService struct {
	db                        *mongo.Database
	playlistsCollection       *mongo.Collection
	tracksCollection          *mongo.Collection
	historyCollection         *mongo.Collection
	recommendationsCollection *mongo.Collection
	userService               *UserService
}

type RecommendationScore struct {
	ItemID    string  `json:"item_id"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
	Algorithm string  `json:"algorithm"`
}

type UserProfile struct {
	UserID            string             `json:"user_id"`
	TopGenres         []string           `json:"top_genres"`
	TopArtists        []string           `json:"top_artists"`
	ListeningPattern  map[string]int     `json:"listening_pattern"`
	SocialConnections []string           `json:"social_connections"`
	Preferences       map[string]float64 `json:"preferences"`
}

func NewRecommendationService(db *mongo.Database, userService *UserService) *RecommendationService {
	return &RecommendationService{
		db:                        db,
		playlistsCollection:       db.Collection("playlists"),
		tracksCollection:          db.Collection("tracks"),
		historyCollection:         db.Collection("user_listening_history"),
		recommendationsCollection: db.Collection("recommendations"),
		userService:               userService,
	}
}

// GetPersonalizedRecommendations returns personalized recommendations for a user
func (s *RecommendationService) GetPersonalizedRecommendations(userID string, limit int) ([]models.Playlist, error) {
	// Get user profile
	profile, err := s.buildUserProfile(userID)
	if err != nil {
		return nil, err
	}

	// Get recommendations using multiple algorithms
	recommendations := make(map[string]RecommendationScore)

	// 1. Collaborative Filtering
	collabRecs, err := s.getCollaborativeRecommendations(profile, limit)
	if err == nil {
		for _, rec := range collabRecs {
			recommendations[rec.ItemID] = rec
		}
	}

	// 2. Content-Based Filtering
	contentRecs, err := s.getContentBasedRecommendations(profile, limit)
	if err == nil {
		for _, rec := range contentRecs {
			if existing, exists := recommendations[rec.ItemID]; exists {
				// Combine scores
				recommendations[rec.ItemID] = RecommendationScore{
					ItemID:    rec.ItemID,
					Score:     (existing.Score + rec.Score) / 2,
					Reason:    fmt.Sprintf("Combined: %s + %s", existing.Reason, rec.Reason),
					Algorithm: "hybrid",
				}
			} else {
				recommendations[rec.ItemID] = rec
			}
		}
	}

	// 3. Social Recommendations
	socialRecs, err := s.getSocialRecommendations(profile, limit)
	if err == nil {
		for _, rec := range socialRecs {
			if existing, exists := recommendations[rec.ItemID]; exists {
				recommendations[rec.ItemID] = RecommendationScore{
					ItemID:    rec.ItemID,
					Score:     existing.Score*0.7 + rec.Score*0.3,
					Reason:    fmt.Sprintf("Social + %s", existing.Reason),
					Algorithm: "social_hybrid",
				}
			} else {
				recommendations[rec.ItemID] = rec
			}
		}
	}

	// 4. Trending Recommendations
	trendingRecs, err := s.getTrendingRecommendations(limit)
	if err == nil {
		for _, rec := range trendingRecs {
			if existing, exists := recommendations[rec.ItemID]; exists {
				recommendations[rec.ItemID] = RecommendationScore{
					ItemID:    rec.ItemID,
					Score:     existing.Score*0.8 + rec.Score*0.2,
					Reason:    fmt.Sprintf("Trending + %s", existing.Reason),
					Algorithm: "trending_hybrid",
				}
			} else {
				recommendations[rec.ItemID] = rec
			}
		}
	}

	// Sort by score and get top recommendations
	var sortedRecs []RecommendationScore
	for _, rec := range recommendations {
		sortedRecs = append(sortedRecs, rec)
	}

	sort.Slice(sortedRecs, func(i, j int) bool {
		return sortedRecs[i].Score > sortedRecs[j].Score
	})

	// Get playlist details for top recommendations
	var playlistIDs []string
	for i, rec := range sortedRecs {
		if i >= limit {
			break
		}
		playlistIDs = append(playlistIDs, rec.ItemID)
	}

	// Fetch playlist details
	filter := bson.M{"_id": bson.M{"$in": playlistIDs}}
	cursor, err := s.playlistsCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var playlists []models.Playlist
	if err := cursor.All(context.Background(), &playlists); err != nil {
		return nil, err
	}

	// Store recommendations for future use
	for _, rec := range sortedRecs[:min(len(sortedRecs), limit)] {
		recommendation := &models.PlaylistRecommendation{
			UserID:     userID,
			PlaylistID: rec.ItemID,
			Score:      rec.Score,
			Reason:     rec.Reason,
			CreatedAt:  time.Now(),
		}
		s.recommendationsCollection.InsertOne(context.Background(), recommendation)
	}

	return playlists, nil
}

// buildUserProfile creates a comprehensive user profile for recommendations
func (s *RecommendationService) buildUserProfile(userID string) (*UserProfile, error) {
	// Get user's listening history
	history, err := s.getUserListeningHistory(userID)
	if err != nil {
		return nil, err
	}

	// Extract top genres
	topGenres := s.extractTopGenres(history)

	// Extract top artists
	topArtists := s.extractTopArtists(history)

	// Build listening pattern
	listeningPattern := s.buildListeningPattern(history)

	// Get social connections
	socialConnections, err := s.getSocialConnections(userID)
	if err != nil {
		socialConnections = []string{}
	}

	// Build preferences
	preferences := s.buildPreferences(history)

	return &UserProfile{
		UserID:            userID,
		TopGenres:         topGenres,
		TopArtists:        topArtists,
		ListeningPattern:  listeningPattern,
		SocialConnections: socialConnections,
		Preferences:       preferences,
	}, nil
}

// getCollaborativeRecommendations uses collaborative filtering
func (s *RecommendationService) getCollaborativeRecommendations(profile *UserProfile, limit int) ([]RecommendationScore, error) {
	// Find users with similar taste
	similarUsers, err := s.findSimilarUsers(profile)
	if err != nil {
		return nil, err
	}

	// Get playlists from similar users
	var recommendations []RecommendationScore
	for _, userID := range similarUsers {
		userPlaylists, err := s.getUserPlaylists(userID)
		if err != nil {
			continue
		}

		for _, playlist := range userPlaylists {
			score := s.calculateCollaborativeScore(profile, playlist)
			recommendations = append(recommendations, RecommendationScore{
				ItemID:    playlist.ID,
				Score:     score,
				Reason:    "Similar users like this",
				Algorithm: "collaborative",
			})
		}
	}

	// Sort by score
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return recommendations[:min(len(recommendations), limit)], nil
}

// getContentBasedRecommendations uses content-based filtering
func (s *RecommendationService) getContentBasedRecommendations(profile *UserProfile, limit int) ([]RecommendationScore, error) {
	// Find playlists with similar content
	filter := bson.M{
		"is_public": true,
		"tags":      bson.M{"$in": profile.TopGenres},
	}

	cursor, err := s.playlistsCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var playlists []models.Playlist
	if err := cursor.All(context.Background(), &playlists); err != nil {
		return nil, err
	}

	var recommendations []RecommendationScore
	for _, playlist := range playlists {
		score := s.calculateContentBasedScore(profile, playlist)
		recommendations = append(recommendations, RecommendationScore{
			ItemID:    playlist.ID,
			Score:     score,
			Reason:    "Matches your music taste",
			Algorithm: "content_based",
		})
	}

	// Sort by score
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return recommendations[:min(len(recommendations), limit)], nil
}

// getSocialRecommendations uses social connections
func (s *RecommendationService) getSocialRecommendations(profile *UserProfile, limit int) ([]RecommendationScore, error) {
	var recommendations []RecommendationScore

	for _, connectionID := range profile.SocialConnections {
		connectionPlaylists, err := s.getUserPlaylists(connectionID)
		if err != nil {
			continue
		}

		for _, playlist := range connectionPlaylists {
			score := s.calculateSocialScore(profile, playlist)
			recommendations = append(recommendations, RecommendationScore{
				ItemID:    playlist.ID,
				Score:     score,
				Reason:    "Your connections like this",
				Algorithm: "social",
			})
		}
	}

	// Sort by score
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return recommendations[:min(len(recommendations), limit)], nil
}

// getTrendingRecommendations gets trending playlists
func (s *RecommendationService) getTrendingRecommendations(limit int) ([]RecommendationScore, error) {
	// Get trending playlists based on recent activity
	filter := bson.M{
		"is_public":  true,
		"followers":  bson.M{"$gte": 5},
		"updated_at": bson.M{"$gte": time.Now().AddDate(0, 0, -7)}, // Last 7 days
	}

	opts := options.Find().
		SetSort(bson.M{"followers": -1, "likes": -1}).
		SetLimit(int64(limit))

	cursor, err := s.playlistsCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var playlists []models.Playlist
	if err := cursor.All(context.Background(), &playlists); err != nil {
		return nil, err
	}

	var recommendations []RecommendationScore
	for _, playlist := range playlists {
		score := float64(playlist.Followers) / 100.0 // Normalize score
		recommendations = append(recommendations, RecommendationScore{
			ItemID:    playlist.ID,
			Score:     score,
			Reason:    "Trending now",
			Algorithm: "trending",
		})
	}

	return recommendations, nil
}

// Helper methods

func (s *RecommendationService) getUserListeningHistory(userID string) ([]models.UserListeningHistory, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "played_at", Value: -1}}).SetLimit(1000)

	cursor, err := s.historyCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var history []models.UserListeningHistory
	if err := cursor.All(context.Background(), &history); err != nil {
		return nil, err
	}

	return history, nil
}

func (s *RecommendationService) extractTopGenres(history []models.UserListeningHistory) []string {
	genreCount := make(map[string]int)
	for _, entry := range history {
		if entry.Track.Genre != "" {
			genreCount[entry.Track.Genre]++
		}
	}

	var genres []string
	for genre := range genreCount {
		genres = append(genres, genre)
	}

	// Sort by count
	sort.Slice(genres, func(i, j int) bool {
		return genreCount[genres[i]] > genreCount[genres[j]]
	})

	return genres[:min(len(genres), 10)]
}

func (s *RecommendationService) extractTopArtists(history []models.UserListeningHistory) []string {
	artistCount := make(map[string]int)
	for _, entry := range history {
		if entry.Track.Artist != "" {
			artistCount[entry.Track.Artist]++
		}
	}

	var artists []string
	for artist := range artistCount {
		artists = append(artists, artist)
	}

	// Sort by count
	sort.Slice(artists, func(i, j int) bool {
		return artistCount[artists[i]] > artistCount[artists[j]]
	})

	return artists[:min(len(artists), 20)]
}

func (s *RecommendationService) buildListeningPattern(history []models.UserListeningHistory) map[string]int {
	pattern := make(map[string]int)
	for _, entry := range history {
		hour := entry.PlayedAt.Hour()
		key := fmt.Sprintf("hour_%d", hour)
		pattern[key]++
	}
	return pattern
}

func (s *RecommendationService) getSocialConnections(userID string) ([]string, error) {
	// This would integrate with the social service
	// For now, return empty slice
	return []string{}, nil
}

func (s *RecommendationService) buildPreferences(history []models.UserListeningHistory) map[string]float64 {
	preferences := make(map[string]float64)

	// Calculate genre preferences
	genreCount := make(map[string]int)
	totalPlays := len(history)

	for _, entry := range history {
		if entry.Track.Genre != "" {
			genreCount[entry.Track.Genre]++
		}
	}

	for genre, count := range genreCount {
		preferences[genre] = float64(count) / float64(totalPlays)
	}

	return preferences
}

func (s *RecommendationService) findSimilarUsers(profile *UserProfile) ([]string, error) {
	// Find users with similar listening patterns
	// This is a simplified implementation
	return []string{}, nil
}

func (s *RecommendationService) getUserPlaylists(userID string) ([]models.Playlist, error) {
	filter := bson.M{"owner": userID, "is_public": true}
	cursor, err := s.playlistsCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var playlists []models.Playlist
	if err := cursor.All(context.Background(), &playlists); err != nil {
		return nil, err
	}

	return playlists, nil
}

func (s *RecommendationService) calculateCollaborativeScore(profile *UserProfile, playlist models.Playlist) float64 {
	// Calculate score based on user similarity
	score := 0.0

	// Genre match
	for _, genre := range profile.TopGenres {
		for _, tag := range playlist.Tags {
			if genre == tag {
				score += 0.3
			}
		}
	}

	// Popularity factor
	score += math.Log(float64(playlist.Followers+1)) * 0.1

	return math.Min(score, 1.0)
}

func (s *RecommendationService) calculateContentBasedScore(profile *UserProfile, playlist models.Playlist) float64 {
	score := 0.0

	// Genre matching
	genreMatches := 0
	for _, genre := range profile.TopGenres {
		for _, tag := range playlist.Tags {
			if genre == tag {
				genreMatches++
			}
		}
	}

	if len(profile.TopGenres) > 0 {
		score += float64(genreMatches) / float64(len(profile.TopGenres)) * 0.7
	}

	// Track count factor
	score += math.Log(float64(playlist.TrackCount+1)) * 0.1

	return math.Min(score, 1.0)
}

func (s *RecommendationService) calculateSocialScore(profile *UserProfile, playlist models.Playlist) float64 {
	// Social connections factor
	score := 0.5 // Base score for social connections

	// Recency factor
	if time.Since(playlist.UpdatedAt) < 7*24*time.Hour {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
