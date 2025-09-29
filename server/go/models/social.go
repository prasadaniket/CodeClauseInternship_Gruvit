package models

import "time"

// SocialActivity represents a social activity in the feed
type SocialActivity struct {
	ID           string                 `json:"id" bson:"_id,omitempty"`
	UserID       string                 `json:"user_id" bson:"user_id"`
	ActivityType string                 `json:"activity_type" bson:"activity_type"` // "track_played", "playlist_created", "playlist_shared", "user_followed", "track_liked"
	TargetID     string                 `json:"target_id" bson:"target_id"`         // ID of the target (track, playlist, user)
	TargetType   string                 `json:"target_type" bson:"target_type"`     // "track", "playlist", "user"
	Metadata     map[string]interface{} `json:"metadata" bson:"metadata"`           // Additional data about the activity
	CreatedAt    time.Time              `json:"created_at" bson:"created_at"`
	IsPublic     bool                   `json:"is_public" bson:"is_public"` // Whether this activity is visible to followers
}

// SocialFeed represents a user's social feed
type SocialFeed struct {
	ID         string           `json:"id" bson:"_id,omitempty"`
	UserID     string           `json:"user_id" bson:"user_id"`
	Activities []SocialActivity `json:"activities" bson:"activities"`
	UpdatedAt  time.Time        `json:"updated_at" bson:"updated_at"`
}

// UserInteraction represents interactions between users
type UserInteraction struct {
	ID              string    `json:"id" bson:"_id,omitempty"`
	UserID          string    `json:"user_id" bson:"user_id"`
	TargetUserID    string    `json:"target_user_id" bson:"target_user_id"`
	InteractionType string    `json:"interaction_type" bson:"interaction_type"` // "follow", "unfollow", "block", "unblock"
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
}

// SocialStats represents social statistics for a user
type SocialStats struct {
	UserID          string    `json:"user_id" bson:"user_id"`
	Followers       int       `json:"followers" bson:"followers"`
	Following       int       `json:"following" bson:"following"`
	Playlists       int       `json:"playlists" bson:"playlists"`
	PublicPlaylists int       `json:"public_playlists" bson:"public_playlists"`
	TotalPlays      int       `json:"total_plays" bson:"total_plays"`
	LastActive      time.Time `json:"last_active" bson:"last_active"`
	UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`
}

// ActivityNotification represents a notification for social activities
type ActivityNotification struct {
	ID               string    `json:"id" bson:"_id,omitempty"`
	UserID           string    `json:"user_id" bson:"user_id"`
	ActivityID       string    `json:"activity_id" bson:"activity_id"`
	NotificationType string    `json:"notification_type" bson:"notification_type"` // "new_follower", "playlist_shared", "track_liked"
	Message          string    `json:"message" bson:"message"`
	IsRead           bool      `json:"is_read" bson:"is_read"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
}
