package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationType string

const (
	NotificationTypeTrackAdded     NotificationType = "track_added"
	NotificationTypePlaylistShared NotificationType = "playlist_shared"
	NotificationTypeUserFollowed   NotificationType = "user_followed"
	NotificationTypeNewRelease     NotificationType = "new_release"
	NotificationTypeSystemUpdate   NotificationType = "system_update"
)

type Notification struct {
	ID        string                 `json:"id" bson:"_id,omitempty"`
	UserID    string                 `json:"user_id" bson:"user_id"`
	Type      NotificationType       `json:"type" bson:"type"`
	Title     string                 `json:"title" bson:"title"`
	Message   string                 `json:"message" bson:"message"`
	Data      map[string]interface{} `json:"data" bson:"data"`
	IsRead    bool                   `json:"is_read" bson:"is_read"`
	CreatedAt time.Time              `json:"created_at" bson:"created_at"`
	ReadAt    *time.Time             `json:"read_at" bson:"read_at,omitempty"`
}

type NotificationService struct {
	notificationsCollection *mongo.Collection
	redisClient             *redis.Client
}

func NewNotificationService(db *mongo.Database, redisClient *redis.Client) *NotificationService {
	return &NotificationService{
		notificationsCollection: db.Collection("notifications"),
		redisClient:             redisClient,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(userID string, notificationType NotificationType, title, message string, data map[string]interface{}) (*Notification, error) {
	notification := &Notification{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Data:      data,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	result, err := s.notificationsCollection.InsertOne(context.Background(), notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %v", err)
	}

	notification.ID = result.InsertedID.(primitive.ObjectID).Hex()

	// Publish to Redis for real-time updates
	s.publishNotification(notification)

	return notification, nil
}

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(userID string, limit int, offset int) ([]Notification, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := s.notificationsCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %v", err)
	}
	defer cursor.Close(context.Background())

	var notifications []Notification
	if err := cursor.All(context.Background(), &notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %v", err)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID, userID string) error {
	filter := bson.M{"_id": notificationID, "user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"is_read": true,
			"read_at": time.Now(),
		},
	}

	_, err := s.notificationsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %v", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(userID string) error {
	filter := bson.M{"user_id": userID, "is_read": false}
	update := bson.M{
		"$set": bson.M{
			"is_read": true,
			"read_at": time.Now(),
		},
	}

	_, err := s.notificationsCollection.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %v", err)
	}

	return nil
}

// GetUnreadCount gets the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID string) (int64, error) {
	filter := bson.M{"user_id": userID, "is_read": false}
	count, err := s.notificationsCollection.CountDocuments(context.Background(), filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %v", err)
	}

	return count, nil
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID, userID string) error {
	filter := bson.M{"_id": notificationID, "user_id": userID}
	_, err := s.notificationsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %v", err)
	}

	return nil
}

// Publish notification to Redis for real-time updates
func (s *NotificationService) publishNotification(notification *Notification) {
	if s.redisClient == nil {
		return
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return
	}

	// Publish to user-specific channel
	channel := fmt.Sprintf("notifications:%s", notification.UserID)
	s.redisClient.Publish(context.Background(), channel, data)

	// Also publish to global notifications channel
	s.redisClient.Publish(context.Background(), "notifications:global", data)
}

// CreateSystemNotification creates a system-wide notification
func (s *NotificationService) CreateSystemNotification(title, message string, data map[string]interface{}) error {
	// Get all users (in a real implementation, you'd have a users collection)
	// For now, we'll create a notification for a specific user or use a system user ID
	systemUserID := "system"

	_, err := s.CreateNotification(systemUserID, NotificationTypeSystemUpdate, title, message, data)
	return err
}

// CreateTrackAddedNotification creates a notification when a track is added to a playlist
func (s *NotificationService) CreateTrackAddedNotification(userID, trackTitle, playlistName string) error {
	title := "Track Added to Playlist"
	message := fmt.Sprintf("'%s' was added to playlist '%s'", trackTitle, playlistName)
	data := map[string]interface{}{
		"track_title":   trackTitle,
		"playlist_name": playlistName,
	}

	_, err := s.CreateNotification(userID, NotificationTypeTrackAdded, title, message, data)
	return err
}

// CreatePlaylistSharedNotification creates a notification when a playlist is shared
func (s *NotificationService) CreatePlaylistSharedNotification(userID, playlistName, sharedBy string) error {
	title := "Playlist Shared"
	message := fmt.Sprintf("'%s' shared playlist '%s' with you", sharedBy, playlistName)
	data := map[string]interface{}{
		"playlist_name": playlistName,
		"shared_by":     sharedBy,
	}

	_, err := s.CreateNotification(userID, NotificationTypePlaylistShared, title, message, data)
	return err
}

// CreateUserFollowedNotification creates a notification when a user is followed
func (s *NotificationService) CreateUserFollowedNotification(userID, followerName string) error {
	title := "New Follower"
	message := fmt.Sprintf("'%s' started following you", followerName)
	data := map[string]interface{}{
		"follower_name": followerName,
	}

	_, err := s.CreateNotification(userID, NotificationTypeUserFollowed, title, message, data)
	return err
}

// CreateNewReleaseNotification creates a notification for new releases from followed artists
func (s *NotificationService) CreateNewReleaseNotification(userID, artistName, releaseTitle string) error {
	title := "New Release"
	message := fmt.Sprintf("'%s' released '%s'", artistName, releaseTitle)
	data := map[string]interface{}{
		"artist_name":   artistName,
		"release_title": releaseTitle,
	}

	_, err := s.CreateNotification(userID, NotificationTypeNewRelease, title, message, data)
	return err
}
