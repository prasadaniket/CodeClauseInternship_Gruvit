package models

import "time"

// Track represents a music track
type Track struct {
	ID        string    `json:"id" bson:"id"`
	Title     string    `json:"title" bson:"title"`
	Artist    string    `json:"artist" bson:"artist"`
	Album     string    `json:"album,omitempty" bson:"album,omitempty"`
	Duration  int       `json:"duration,omitempty" bson:"duration,omitempty"`
	StreamURL string    `json:"stream_url,omitempty" bson:"stream_url,omitempty"`
	ImageURL  string    `json:"image_url,omitempty" bson:"image_url,omitempty"`
	Genre     string    `json:"genre,omitempty" bson:"genre,omitempty"`
	Source    string    `json:"source" bson:"source"`
	UpdatedAt time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

// Playlist represents a user playlist
type Playlist struct {
	ID              string    `json:"id" bson:"_id,omitempty"`
	Name            string    `json:"name" bson:"name"`
	Description     string    `json:"description,omitempty" bson:"description,omitempty"`
	Owner           string    `json:"owner" bson:"owner"`
	Tracks          []Track   `json:"tracks" bson:"tracks"`
	IsPublic        bool      `json:"is_public" bson:"is_public"`
	IsCollaborative bool      `json:"is_collaborative" bson:"is_collaborative"`
	TrackCount      int       `json:"track_count" bson:"track_count"`
	Duration        int       `json:"duration" bson:"duration"`
	Followers       int       `json:"followers" bson:"followers"`
	Likes           int       `json:"likes" bson:"likes"`
	Shares          int       `json:"shares" bson:"shares"`
	ImageURL        string    `json:"image_url,omitempty" bson:"image_url,omitempty"`
	Tags            []string  `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`
}

// PlaylistTrack represents a track in a playlist
type PlaylistTrack struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	PlaylistID string    `json:"playlist_id" bson:"playlist_id"`
	TrackID    string    `json:"track_id" bson:"track_id"`
	Position   int       `json:"position" bson:"position"`
	AddedBy    string    `json:"added_by" bson:"added_by"`
	AddedAt    time.Time `json:"added_at" bson:"added_at"`
	Track      *Track    `json:"track,omitempty" bson:"track,omitempty"`
}

// PlaylistCollaborator represents a user who can collaborate on a playlist
type PlaylistCollaborator struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	PlaylistID string    `json:"playlist_id" bson:"playlist_id"`
	UserID     string    `json:"user_id" bson:"user_id"`
	Role       string    `json:"role" bson:"role"` // "editor", "viewer"
	AddedBy    string    `json:"added_by" bson:"added_by"`
	AddedAt    time.Time `json:"added_at" bson:"added_at"`
}

// PlaylistFollow represents a user following a playlist
type PlaylistFollow struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	PlaylistID string    `json:"playlist_id" bson:"playlist_id"`
	UserID     string    `json:"user_id" bson:"user_id"`
	FollowedAt time.Time `json:"followed_at" bson:"followed_at"`
}

// PlaylistLike represents a user liking a playlist
type PlaylistLike struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	PlaylistID string    `json:"playlist_id" bson:"playlist_id"`
	UserID     string    `json:"user_id" bson:"user_id"`
	LikedAt    time.Time `json:"liked_at" bson:"liked_at"`
}

// PlaylistShare represents a playlist share
type PlaylistShare struct {
	ID         string     `json:"id" bson:"_id,omitempty"`
	PlaylistID string     `json:"playlist_id" bson:"playlist_id"`
	SharedBy   string     `json:"shared_by" bson:"shared_by"`
	SharedWith string     `json:"shared_with,omitempty" bson:"shared_with,omitempty"` // User ID or email
	ShareType  string     `json:"share_type" bson:"share_type"`                       // "public", "private", "link"
	ShareToken string     `json:"share_token" bson:"share_token"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
}

// PlaylistRecommendation represents a playlist recommendation
type PlaylistRecommendation struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	UserID     string    `json:"user_id" bson:"user_id"`
	PlaylistID string    `json:"playlist_id" bson:"playlist_id"`
	Score      float64   `json:"score" bson:"score"`
	Reason     string    `json:"reason" bson:"reason"` // "similar_taste", "trending", "new_releases"
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
}

// PlaylistActivity represents activity on a playlist
type PlaylistActivity struct {
	ID         string                 `json:"id" bson:"_id,omitempty"`
	PlaylistID string                 `json:"playlist_id" bson:"playlist_id"`
	UserID     string                 `json:"user_id" bson:"user_id"`
	Action     string                 `json:"action" bson:"action"` // "track_added", "track_removed", "track_moved", "playlist_updated"
	Details    map[string]interface{} `json:"details" bson:"details"`
	CreatedAt  time.Time              `json:"created_at" bson:"created_at"`
}

// PlaylistStats represents statistics for a playlist
type PlaylistStats struct {
	PlaylistID      string     `json:"playlist_id" bson:"playlist_id"`
	TotalPlays      int        `json:"total_plays" bson:"total_plays"`
	UniqueListeners int        `json:"unique_listeners" bson:"unique_listeners"`
	AvgPlayTime     float64    `json:"avg_play_time" bson:"avg_play_time"`
	LastPlayed      *time.Time `json:"last_played,omitempty" bson:"last_played,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at" bson:"updated_at"`
}
