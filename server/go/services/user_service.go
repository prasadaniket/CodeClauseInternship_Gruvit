package services

import (
	"context"
	"fmt"
	"time"

	"gruvit/server/go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService struct {
	usersCollection            *mongo.Collection
	profilesCollection         *mongo.Collection
	favoritesCollection        *mongo.Collection
	listeningHistoryCollection *mongo.Collection
	followsCollection          *mongo.Collection
	artistsCollection          *mongo.Collection
	statsCollection            *mongo.Collection
}

func NewUserService(db *mongo.Database) *UserService {
	return &UserService{
		usersCollection:            db.Collection("users"),
		profilesCollection:         db.Collection("user_profiles"),
		favoritesCollection:        db.Collection("user_favorites"),
		listeningHistoryCollection: db.Collection("listening_history"),
		followsCollection:          db.Collection("user_follows"),
		artistsCollection:          db.Collection("artists"),
		statsCollection:            db.Collection("user_stats"),
	}
}

// User Profile Operations
func (s *UserService) CreateUserProfile(userID, displayName string) (*models.UserProfile, error) {
	profile := &models.UserProfile{
		UserID:      userID,
		DisplayName: displayName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := s.profilesCollection.InsertOne(context.Background(), profile)
	if err != nil {
		return nil, err
	}

	profile.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return profile, nil
}

func (s *UserService) GetUserProfile(userID string) (*models.UserProfile, error) {
	filter := bson.M{"user_id": userID}
	var profile models.UserProfile
	err := s.profilesCollection.FindOne(context.Background(), filter).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Profile not found
		}
		return nil, err
	}
	return &profile, nil
}

func (s *UserService) UpdateUserProfile(userID string, updates bson.M) (*models.UserProfile, error) {
	updates["updated_at"] = time.Now()

	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": updates}

	_, err := s.profilesCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}

	return s.GetUserProfile(userID)
}

// User Favorites Operations
func (s *UserService) AddToFavorites(userID string, track models.Track) error {
	favorite := &models.UserFavorite{
		UserID:    userID,
		TrackID:   track.ID,
		Track:     track,
		CreatedAt: time.Now(),
	}

	// Check if already exists
	filter := bson.M{
		"user_id":  userID,
		"track_id": track.ID,
	}

	var existing models.UserFavorite
	err := s.favoritesCollection.FindOne(context.Background(), filter).Decode(&existing)
	if err == nil {
		return fmt.Errorf("track already in favorites")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	_, err = s.favoritesCollection.InsertOne(context.Background(), favorite)
	return err
}

func (s *UserService) RemoveFromFavorites(userID, trackID string) error {
	filter := bson.M{
		"user_id":  userID,
		"track_id": trackID,
	}

	_, err := s.favoritesCollection.DeleteOne(context.Background(), filter)
	return err
}

func (s *UserService) GetUserFavorites(userID string, limit int) ([]models.UserFavorite, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := s.favoritesCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var favorites []models.UserFavorite
	if err := cursor.All(context.Background(), &favorites); err != nil {
		return nil, err
	}

	return favorites, nil
}

// Listening History Operations
func (s *UserService) AddListeningHistory(userID string, track models.Track, duration int) error {
	history := &models.UserListeningHistory{
		UserID:   userID,
		TrackID:  track.ID,
		Track:    track,
		PlayedAt: time.Now(),
		Duration: duration,
	}

	_, err := s.listeningHistoryCollection.InsertOne(context.Background(), history)
	return err
}

func (s *UserService) GetUserListeningHistory(userID string, limit int) ([]models.UserListeningHistory, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "played_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := s.listeningHistoryCollection.Find(context.Background(), filter, opts)
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

// Follow Operations
func (s *UserService) FollowUser(userID, followedUserID string) error {
	follow := &models.UserFollow{
		UserID:     userID,
		FollowedID: followedUserID,
		Type:       "user",
		CreatedAt:  time.Now(),
	}

	// Check if already following
	filter := bson.M{
		"user_id":     userID,
		"followed_id": followedUserID,
		"type":        "user",
	}

	var existing models.UserFollow
	err := s.followsCollection.FindOne(context.Background(), filter).Decode(&existing)
	if err == nil {
		return fmt.Errorf("already following this user")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	_, err = s.followsCollection.InsertOne(context.Background(), follow)
	return err
}

func (s *UserService) UnfollowUser(userID, followedUserID string) error {
	filter := bson.M{
		"user_id":     userID,
		"followed_id": followedUserID,
		"type":        "user",
	}

	_, err := s.followsCollection.DeleteOne(context.Background(), filter)
	return err
}

func (s *UserService) FollowArtist(userID, artistID string) error {
	follow := &models.UserFollow{
		UserID:     userID,
		FollowedID: artistID,
		Type:       "artist",
		CreatedAt:  time.Now(),
	}

	// Check if already following
	filter := bson.M{
		"user_id":     userID,
		"followed_id": artistID,
		"type":        "artist",
	}

	var existing models.UserFollow
	err := s.followsCollection.FindOne(context.Background(), filter).Decode(&existing)
	if err == nil {
		return fmt.Errorf("already following this artist")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	_, err = s.followsCollection.InsertOne(context.Background(), follow)
	return err
}

func (s *UserService) UnfollowArtist(userID, artistID string) error {
	filter := bson.M{
		"user_id":     userID,
		"followed_id": artistID,
		"type":        "artist",
	}

	_, err := s.followsCollection.DeleteOne(context.Background(), filter)
	return err
}

func (s *UserService) GetUserFollowings(userID string, limit int) ([]models.UserFollow, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := s.followsCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var followings []models.UserFollow
	if err := cursor.All(context.Background(), &followings); err != nil {
		return nil, err
	}

	return followings, nil
}

func (s *UserService) GetUserFollowers(userID string, limit int) ([]models.UserFollow, error) {
	filter := bson.M{
		"followed_id": userID,
		"type":        "user",
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := s.followsCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var followers []models.UserFollow
	if err := cursor.All(context.Background(), &followers); err != nil {
		return nil, err
	}

	return followers, nil
}

// Statistics Operations
func (s *UserService) GetUserStats(userID string) (*models.UserStats, error) {
	filter := bson.M{"user_id": userID}
	var stats models.UserStats
	err := s.statsCollection.FindOne(context.Background(), filter).Decode(&stats)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create default stats
			stats = models.UserStats{
				UserID:         userID,
				TotalPlays:     0,
				TotalPlaylists: 0,
				TotalFavorites: 0,
				TotalFollowing: 0,
				TotalFollowers: 0,
				LastActive:     time.Now(),
			}
			_, err = s.statsCollection.InsertOne(context.Background(), stats)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &stats, nil
}

func (s *UserService) UpdateUserStats(userID string, updates bson.M) error {
	updates["last_active"] = time.Now()

	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": updates}

	_, err := s.statsCollection.UpdateOne(context.Background(), filter, update)
	return err
}

// Get top artists based on listening history
func (s *UserService) GetUserTopArtists(userID string, limit int) ([]models.Artist, error) {
	// Aggregate listening history to get top artists
	pipeline := []bson.M{
		{
			"$match": bson.M{"user_id": userID},
		},
		{
			"$group": bson.M{
				"_id":        "$track.artist",
				"play_count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"play_count": -1},
		},
		{
			"$limit": int64(limit),
		},
	}

	cursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	var artists []models.Artist
	for _, result := range results {
		artist := models.Artist{
			Name: result["_id"].(string),
			// You might want to fetch more details from the artists collection
		}
		artists = append(artists, artist)
	}

	return artists, nil
}

// Get top tracks based on listening history
func (s *UserService) GetUserTopTracks(userID string, limit int) ([]models.Track, error) {
	// Aggregate listening history to get top tracks
	pipeline := []bson.M{
		{
			"$match": bson.M{"user_id": userID},
		},
		{
			"$group": bson.M{
				"_id":        "$track",
				"play_count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"play_count": -1},
		},
		{
			"$limit": int64(limit),
		},
	}

	cursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	var tracks []models.Track
	for _, result := range results {
		if trackData, ok := result["_id"].(bson.M); ok {
			var track models.Track
			bsonBytes, _ := bson.Marshal(trackData)
			bson.Unmarshal(bsonBytes, &track)
			tracks = append(tracks, track)
		}
	}

	return tracks, nil
}

// Get detailed listening statistics with enhanced analytics
func (s *UserService) GetUserListeningStats(userID string) (map[string]interface{}, error) {
	// Get total plays
	totalPlays, err := s.listeningHistoryCollection.CountDocuments(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	// Get unique artists count
	uniqueArtists, err := s.listeningHistoryCollection.Distinct(context.Background(), "track.artist", bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	// Get unique tracks count
	uniqueTracks, err := s.listeningHistoryCollection.Distinct(context.Background(), "track.id", bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	// Get total listening time
	totalDurationPipeline := []bson.M{
		{
			"$match": bson.M{"user_id": userID},
		},
		{
			"$group": bson.M{
				"_id":            nil,
				"total_duration": bson.M{"$sum": "$duration"},
			},
		},
	}

	durationCursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), totalDurationPipeline)
	if err != nil {
		return nil, err
	}
	defer durationCursor.Close(context.Background())

	var totalDurationResult []bson.M
	if err := durationCursor.All(context.Background(), &totalDurationResult); err != nil {
		return nil, err
	}

	totalDuration := 0
	if len(totalDurationResult) > 0 {
		if duration, ok := totalDurationResult[0]["total_duration"].(int32); ok {
			totalDuration = int(duration)
		}
	}

	// Get listening time by day (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"played_at": bson.M{"$gte": thirtyDaysAgo},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"year":  bson.M{"$year": "$played_at"},
					"month": bson.M{"$month": "$played_at"},
					"day":   bson.M{"$dayOfMonth": "$played_at"},
				},
				"total_duration": bson.M{"$sum": "$duration"},
				"play_count":     bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var dailyStats []bson.M
	if err := cursor.All(context.Background(), &dailyStats); err != nil {
		return nil, err
	}

	// Get top genres with enhanced data
	genrePipeline := []bson.M{
		{
			"$match": bson.M{"user_id": userID},
		},
		{
			"$group": bson.M{
				"_id":            "$track.genre",
				"play_count":     bson.M{"$sum": 1},
				"total_duration": bson.M{"$sum": "$duration"},
				"unique_tracks":  bson.M{"$addToSet": "$track.id"},
				"unique_artists": bson.M{"$addToSet": "$track.artist"},
			},
		},
		{
			"$addFields": bson.M{
				"unique_tracks_count":  bson.M{"$size": "$unique_tracks"},
				"unique_artists_count": bson.M{"$size": "$unique_artists"},
			},
		},
		{
			"$sort": bson.M{"play_count": -1},
		},
		{
			"$limit": 10,
		},
	}

	genreCursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), genrePipeline)
	if err != nil {
		return nil, err
	}
	defer genreCursor.Close(context.Background())

	var topGenres []bson.M
	if err := genreCursor.All(context.Background(), &topGenres); err != nil {
		return nil, err
	}

	// Get listening patterns by hour of day
	hourlyPatternPipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"played_at": bson.M{"$gte": thirtyDaysAgo},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"hour": bson.M{"$hour": "$played_at"},
				},
				"play_count":     bson.M{"$sum": 1},
				"total_duration": bson.M{"$sum": "$duration"},
			},
		},
		{
			"$sort": bson.M{"_id.hour": 1},
		},
	}

	hourlyCursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), hourlyPatternPipeline)
	if err != nil {
		return nil, err
	}
	defer hourlyCursor.Close(context.Background())

	var hourlyPatterns []bson.M
	if err := hourlyCursor.All(context.Background(), &hourlyPatterns); err != nil {
		return nil, err
	}

	// Get listening patterns by day of week
	weeklyPatternPipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"played_at": bson.M{"$gte": thirtyDaysAgo},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"day_of_week": bson.M{"$dayOfWeek": "$played_at"},
				},
				"play_count":     bson.M{"$sum": 1},
				"total_duration": bson.M{"$sum": "$duration"},
			},
		},
		{
			"$sort": bson.M{"_id.day_of_week": 1},
		},
	}

	weeklyCursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), weeklyPatternPipeline)
	if err != nil {
		return nil, err
	}
	defer weeklyCursor.Close(context.Background())

	var weeklyPatterns []bson.M
	if err := weeklyCursor.All(context.Background(), &weeklyPatterns); err != nil {
		return nil, err
	}

	// Get most active listening periods
	activePeriodsPipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"played_at": bson.M{"$gte": thirtyDaysAgo},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"year":  bson.M{"$year": "$played_at"},
					"month": bson.M{"$month": "$played_at"},
					"week":  bson.M{"$week": "$played_at"},
				},
				"play_count":     bson.M{"$sum": 1},
				"total_duration": bson.M{"$sum": "$duration"},
			},
		},
		{
			"$sort": bson.M{"total_duration": -1},
		},
		{
			"$limit": 5,
		},
	}

	activePeriodsCursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), activePeriodsPipeline)
	if err != nil {
		return nil, err
	}
	defer activePeriodsCursor.Close(context.Background())

	var activePeriods []bson.M
	if err := activePeriodsCursor.All(context.Background(), &activePeriods); err != nil {
		return nil, err
	}

	// Calculate average session duration
	avgSessionDuration := 0
	if totalPlays > 0 {
		avgSessionDuration = totalDuration / int(totalPlays)
	}

	// Calculate listening streak (consecutive days with plays)
	listeningStreak := s.calculateListeningStreak(userID, thirtyDaysAgo)

	return map[string]interface{}{
		"total_plays":          totalPlays,
		"total_duration":       totalDuration,
		"avg_session_duration": avgSessionDuration,
		"unique_artists":       len(uniqueArtists),
		"unique_tracks":        len(uniqueTracks),
		"daily_stats":          dailyStats,
		"top_genres":           topGenres,
		"hourly_patterns":      hourlyPatterns,
		"weekly_patterns":      weeklyPatterns,
		"active_periods":       activePeriods,
		"listening_streak":     listeningStreak,
		"period":               "last_30_days",
		"insights": map[string]interface{}{
			"most_active_hour":      s.getMostActiveHour(hourlyPatterns),
			"most_active_day":       s.getMostActiveDay(weeklyPatterns),
			"favorite_genre":        s.getFavoriteGenre(topGenres),
			"listening_consistency": s.calculateListeningConsistency(dailyStats),
		},
	}, nil
}

// Get listening history with pagination
func (s *UserService) GetUserListeningHistoryPaginated(userID string, page, limit int) ([]models.UserListeningHistory, int64, error) {
	// Get total count
	total, err := s.listeningHistoryCollection.CountDocuments(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get paginated results
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "played_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := s.listeningHistoryCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	var history []models.UserListeningHistory
	if err := cursor.All(context.Background(), &history); err != nil {
		return nil, 0, err
	}

	return history, total, nil
}

// RecordPlay records a track play in user's listening history
func (s *UserService) RecordPlay(userID, trackID string, duration int) error {
	// Create listening history entry
	historyEntry := &models.UserListeningHistory{
		UserID:   userID,
		TrackID:  trackID,
		Duration: duration,
		PlayedAt: time.Now(),
	}

	// Insert into database
	_, err := s.listeningHistoryCollection.InsertOne(context.Background(), historyEntry)
	return err
}

// Helper methods for enhanced analytics

// calculateListeningStreak calculates consecutive days with plays
func (s *UserService) calculateListeningStreak(userID string, since time.Time) int {
	// Get all days with plays in the last 30 days
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":   userID,
				"played_at": bson.M{"$gte": since},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"year":  bson.M{"$year": "$played_at"},
					"month": bson.M{"$month": "$played_at"},
					"day":   bson.M{"$dayOfMonth": "$played_at"},
				},
			},
		},
		{
			"$sort": bson.M{"_id": -1},
		},
	}

	cursor, err := s.listeningHistoryCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return 0
	}
	defer cursor.Close(context.Background())

	var daysWithPlays []bson.M
	if err := cursor.All(context.Background(), &daysWithPlays); err != nil {
		return 0
	}

	// Calculate consecutive days from today backwards
	today := time.Now()
	streak := 0

	for _, day := range daysWithPlays {
		dayData := day["_id"].(bson.M)
		year := int(dayData["year"].(int32))
		month := int(dayData["month"].(int32))
		dayOfMonth := int(dayData["day"].(int32))

		playDate := time.Date(year, time.Month(month), dayOfMonth, 0, 0, 0, 0, time.UTC)
		daysDiff := int(today.Sub(playDate).Hours() / 24)

		if daysDiff == streak {
			streak++
		} else {
			break
		}
	}

	return streak
}

// getMostActiveHour finds the hour with most plays
func (s *UserService) getMostActiveHour(hourlyPatterns []bson.M) map[string]interface{} {
	if len(hourlyPatterns) == 0 {
		return map[string]interface{}{"hour": 0, "play_count": 0}
	}

	maxPlays := 0
	mostActiveHour := 0

	for _, pattern := range hourlyPatterns {
		hourData := pattern["_id"].(bson.M)
		hour := int(hourData["hour"].(int32))
		playCount := int(pattern["play_count"].(int32))

		if playCount > maxPlays {
			maxPlays = playCount
			mostActiveHour = hour
		}
	}

	return map[string]interface{}{
		"hour":       mostActiveHour,
		"play_count": maxPlays,
	}
}

// getMostActiveDay finds the day of week with most plays
func (s *UserService) getMostActiveDay(weeklyPatterns []bson.M) map[string]interface{} {
	if len(weeklyPatterns) == 0 {
		return map[string]interface{}{"day": "Unknown", "play_count": 0}
	}

	maxPlays := 0
	mostActiveDay := 1

	dayNames := []string{"", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	for _, pattern := range weeklyPatterns {
		dayData := pattern["_id"].(bson.M)
		dayOfWeek := int(dayData["day_of_week"].(int32))
		playCount := int(pattern["play_count"].(int32))

		if playCount > maxPlays {
			maxPlays = playCount
			mostActiveDay = dayOfWeek
		}
	}

	return map[string]interface{}{
		"day":        dayNames[mostActiveDay],
		"play_count": maxPlays,
	}
}

// getFavoriteGenre finds the most played genre
func (s *UserService) getFavoriteGenre(topGenres []bson.M) map[string]interface{} {
	if len(topGenres) == 0 {
		return map[string]interface{}{"genre": "Unknown", "play_count": 0}
	}

	topGenre := topGenres[0]
	genreName := topGenre["_id"]
	playCount := int(topGenre["play_count"].(int32))

	return map[string]interface{}{
		"genre":      genreName,
		"play_count": playCount,
	}
}

// calculateListeningConsistency calculates how consistent the user's listening is
func (s *UserService) calculateListeningConsistency(dailyStats []bson.M) map[string]interface{} {
	if len(dailyStats) == 0 {
		return map[string]interface{}{"consistency_score": 0, "active_days": 0}
	}

	activeDays := len(dailyStats)
	totalDays := 30 // Last 30 days
	consistencyScore := float64(activeDays) / float64(totalDays) * 100

	return map[string]interface{}{
		"consistency_score": int(consistencyScore),
		"active_days":       activeDays,
		"total_days":        totalDays,
	}
}
