package models

import (
	"time"
)

type SearchResponse struct {
	Query   string  `json:"query"`
	Results []Track `json:"results"`
	Total   int     `json:"total"`
	Page    int     `json:"page"`
	Limit   int     `json:"limit"`
}

type StreamResponse struct {
	TrackID   string `json:"track_id"`
	StreamURL string `json:"stream_url"`
	ExpiresAt int64  `json:"expires_at"`
}

type User struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Username  string    `json:"username" bson:"username"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"-" bson:"password"` // Hidden from JSON
	Role      string    `json:"role" bson:"role"`
	IsActive  bool      `json:"is_active" bson:"is_active"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// UserProfile represents user profile information
type UserProfile struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	UserID      string    `json:"user_id" bson:"user_id"`
	DisplayName string    `json:"display_name" bson:"display_name"`
	Bio         string    `json:"bio" bson:"bio"`
	Avatar      string    `json:"avatar" bson:"avatar"`
	Location    string    `json:"location" bson:"location"`
	Website     string    `json:"website" bson:"website"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

// UserFavorite represents a user's favorite track
type UserFavorite struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	UserID    string    `json:"user_id" bson:"user_id"`
	TrackID   string    `json:"track_id" bson:"track_id"`
	Track     Track     `json:"track" bson:"track"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// UserListeningHistory represents user's listening history
type UserListeningHistory struct {
	ID       string    `json:"id" bson:"_id,omitempty"`
	UserID   string    `json:"user_id" bson:"user_id"`
	TrackID  string    `json:"track_id" bson:"track_id"`
	Track    Track     `json:"track" bson:"track"`
	PlayedAt time.Time `json:"played_at" bson:"played_at"`
	Duration int       `json:"duration" bson:"duration"` // How long the user listened (in seconds)
}

// UserFollow represents user following another user or artist
type UserFollow struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	UserID     string    `json:"user_id" bson:"user_id"`
	FollowedID string    `json:"followed_id" bson:"followed_id"` // Can be user ID or artist ID
	Type       string    `json:"type" bson:"type"`               // "user" or "artist"
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
}

// Artist represents an artist
type Artist struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	Name       string    `json:"name" bson:"name"`
	Bio        string    `json:"bio" bson:"bio"`
	Image      string    `json:"image" bson:"image"`
	Genre      string    `json:"genre" bson:"genre"`
	Followers  int       `json:"followers" bson:"followers"`
	IsVerified bool      `json:"is_verified" bson:"is_verified"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

// UserStats represents user statistics
type UserStats struct {
	UserID         string    `json:"user_id" bson:"user_id"`
	TotalPlays     int       `json:"total_plays" bson:"total_plays"`
	TotalPlaylists int       `json:"total_playlists" bson:"total_playlists"`
	TotalFavorites int       `json:"total_favorites" bson:"total_favorites"`
	TotalFollowing int       `json:"total_following" bson:"total_following"`
	TotalFollowers int       `json:"total_followers" bson:"total_followers"`
	LastActive     time.Time `json:"last_active" bson:"last_active"`
}
