package services

import (
	"context"
	"fmt"
	"time"

	"gruvit/server/go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SocialService struct {
	db                      *mongo.Database
	activitiesCollection    *mongo.Collection
	feedsCollection         *mongo.Collection
	interactionsCollection  *mongo.Collection
	statsCollection         *mongo.Collection
	notificationsCollection *mongo.Collection
}

func NewSocialService(db *mongo.Database) *SocialService {
	return &SocialService{
		db:                      db,
		activitiesCollection:    db.Collection("social_activities"),
		feedsCollection:         db.Collection("social_feeds"),
		interactionsCollection:  db.Collection("user_interactions"),
		statsCollection:         db.Collection("social_stats"),
		notificationsCollection: db.Collection("activity_notifications"),
	}
}

// RecordActivity records a social activity
func (s *SocialService) RecordActivity(userID, activityType, targetID, targetType string, metadata map[string]interface{}, isPublic bool) error {
	activity := &models.SocialActivity{
		UserID:       userID,
		ActivityType: activityType,
		TargetID:     targetID,
		TargetType:   targetType,
		Metadata:     metadata,
		CreatedAt:    time.Now(),
		IsPublic:     isPublic,
	}

	_, err := s.activitiesCollection.InsertOne(context.Background(), activity)
	if err != nil {
		return err
	}

	// Update user's social feed
	err = s.updateUserFeed(userID, activity)
	if err != nil {
		return err
	}

	// Notify followers if activity is public
	if isPublic {
		err = s.notifyFollowers(userID, activity)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetUserFeed returns the social feed for a user
func (s *SocialService) GetUserFeed(userID string, limit int) ([]models.SocialActivity, error) {
	// Get user's following list
	following, err := s.getUserFollowing(userID)
	if err != nil {
		return nil, err
	}

	// Add user's own activities
	following = append(following, userID)

	// Get activities from followed users
	filter := bson.M{
		"user_id":   bson.M{"$in": following},
		"is_public": true,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := s.activitiesCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var activities []models.SocialActivity
	if err := cursor.All(context.Background(), &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// GetUserActivities returns activities for a specific user
func (s *SocialService) GetUserActivities(userID string, limit int) ([]models.SocialActivity, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := s.activitiesCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var activities []models.SocialActivity
	if err := cursor.All(context.Background(), &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// FollowUser makes a user follow another user
func (s *SocialService) FollowUser(userID, targetUserID string) error {
	// Check if already following
	existing := s.interactionsCollection.FindOne(context.Background(), bson.M{
		"user_id":          userID,
		"target_user_id":   targetUserID,
		"interaction_type": "follow",
	})
	if existing.Err() == nil {
		return fmt.Errorf("already following this user")
	}

	// Create follow interaction
	interaction := &models.UserInteraction{
		UserID:          userID,
		TargetUserID:    targetUserID,
		InteractionType: "follow",
		CreatedAt:       time.Now(),
	}

	_, err := s.interactionsCollection.InsertOne(context.Background(), interaction)
	if err != nil {
		return err
	}

	// Update social stats
	err = s.updateSocialStats(userID, "following", 1)
	if err != nil {
		return err
	}

	err = s.updateSocialStats(targetUserID, "followers", 1)
	if err != nil {
		return err
	}

	// Record activity
	metadata := map[string]interface{}{
		"target_user_id": targetUserID,
	}
	err = s.RecordActivity(userID, "user_followed", targetUserID, "user", metadata, true)
	if err != nil {
		return err
	}

	// Create notification for target user
	notification := &models.ActivityNotification{
		UserID:           targetUserID,
		ActivityID:       interaction.ID,
		NotificationType: "new_follower",
		Message:          fmt.Sprintf("Someone started following you"),
		IsRead:           false,
		CreatedAt:        time.Now(),
	}
	s.notificationsCollection.InsertOne(context.Background(), notification)

	return nil
}

// UnfollowUser makes a user unfollow another user
func (s *SocialService) UnfollowUser(userID, targetUserID string) error {
	_, err := s.interactionsCollection.DeleteOne(context.Background(), bson.M{
		"user_id":          userID,
		"target_user_id":   targetUserID,
		"interaction_type": "follow",
	})
	if err != nil {
		return err
	}

	// Update social stats
	err = s.updateSocialStats(userID, "following", -1)
	if err != nil {
		return err
	}

	err = s.updateSocialStats(targetUserID, "followers", -1)
	if err != nil {
		return err
	}

	return nil
}

// GetUserStats returns social statistics for a user
func (s *SocialService) GetUserStats(userID string) (*models.SocialStats, error) {
	var stats models.SocialStats
	err := s.statsCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&stats)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create new stats if none exist
			stats = models.SocialStats{
				UserID:          userID,
				Followers:       0,
				Following:       0,
				Playlists:       0,
				PublicPlaylists: 0,
				TotalPlays:      0,
				LastActive:      time.Now(),
				UpdatedAt:       time.Now(),
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

// GetNotifications returns notifications for a user
func (s *SocialService) GetNotifications(userID string, limit int) ([]models.ActivityNotification, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := s.notificationsCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var notifications []models.ActivityNotification
	if err := cursor.All(context.Background(), &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func (s *SocialService) MarkNotificationAsRead(notificationID string) error {
	_, err := s.notificationsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": notificationID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	return err
}

// Helper methods

func (s *SocialService) updateUserFeed(userID string, activity *models.SocialActivity) error {
	// Update or create user's feed
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$push": bson.M{
			"activities": bson.M{
				"$each":  []models.SocialActivity{*activity},
				"$slice": -100, // Keep only last 100 activities
			},
		},
		"$set": bson.M{"updated_at": time.Now()},
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.feedsCollection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

func (s *SocialService) getUserFollowing(userID string) ([]string, error) {
	cursor, err := s.interactionsCollection.Find(context.Background(), bson.M{
		"user_id":          userID,
		"interaction_type": "follow",
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var interactions []models.UserInteraction
	if err := cursor.All(context.Background(), &interactions); err != nil {
		return nil, err
	}

	var following []string
	for _, interaction := range interactions {
		following = append(following, interaction.TargetUserID)
	}

	return following, nil
}

func (s *SocialService) notifyFollowers(userID string, activity *models.SocialActivity) error {
	// Get followers of the user
	followers, err := s.getUserFollowers(userID)
	if err != nil {
		return err
	}

	// Create notifications for followers
	for _, followerID := range followers {
		notification := &models.ActivityNotification{
			UserID:           followerID,
			ActivityID:       activity.ID,
			NotificationType: "activity_update",
			Message:          fmt.Sprintf("New activity from someone you follow"),
			IsRead:           false,
			CreatedAt:        time.Now(),
		}
		s.notificationsCollection.InsertOne(context.Background(), notification)
	}

	return nil
}

func (s *SocialService) getUserFollowers(userID string) ([]string, error) {
	cursor, err := s.interactionsCollection.Find(context.Background(), bson.M{
		"target_user_id":   userID,
		"interaction_type": "follow",
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var interactions []models.UserInteraction
	if err := cursor.All(context.Background(), &interactions); err != nil {
		return nil, err
	}

	var followers []string
	for _, interaction := range interactions {
		followers = append(followers, interaction.UserID)
	}

	return followers, nil
}

func (s *SocialService) updateSocialStats(userID, field string, increment int) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$inc": bson.M{field: increment},
		"$set": bson.M{"updated_at": time.Now()},
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.statsCollection.UpdateOne(context.Background(), filter, update, opts)
	return err
}
