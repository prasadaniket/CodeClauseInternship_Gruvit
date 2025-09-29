package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"gruvit/server/go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlaylistEnhancedService struct {
	db                        *mongo.Database
	playlistsCollection       *mongo.Collection
	tracksCollection          *mongo.Collection
	collaboratorsCollection   *mongo.Collection
	followsCollection         *mongo.Collection
	likesCollection           *mongo.Collection
	sharesCollection          *mongo.Collection
	recommendationsCollection *mongo.Collection
	activityCollection        *mongo.Collection
	statsCollection           *mongo.Collection
	userService               *UserService
}

func NewPlaylistEnhancedService(db *mongo.Database, userService *UserService) *PlaylistEnhancedService {
	return &PlaylistEnhancedService{
		db:                        db,
		playlistsCollection:       db.Collection("playlists"),
		tracksCollection:          db.Collection("playlist_tracks"),
		collaboratorsCollection:   db.Collection("playlist_collaborators"),
		followsCollection:         db.Collection("playlist_follows"),
		likesCollection:           db.Collection("playlist_likes"),
		sharesCollection:          db.Collection("playlist_shares"),
		recommendationsCollection: db.Collection("playlist_recommendations"),
		activityCollection:        db.Collection("playlist_activities"),
		statsCollection:           db.Collection("playlist_stats"),
		userService:               userService,
	}
}

// CreateCollaborativePlaylist creates a new collaborative playlist
func (s *PlaylistEnhancedService) CreateCollaborativePlaylist(userID, title, description string, isPublic bool) (*models.Playlist, error) {
	playlist := &models.Playlist{
		Name:            title,
		Description:     description,
		Owner:           userID,
		IsPublic:        isPublic,
		IsCollaborative: true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		TrackCount:      0,
		Duration:        0,
		Followers:       0,
		Likes:           0,
		Shares:          0,
	}

	result, err := s.playlistsCollection.InsertOne(context.Background(), playlist)
	if err != nil {
		return nil, err
	}

	playlist.ID = result.InsertedID.(string)

	// Add owner as collaborator with editor role
	collaborator := &models.PlaylistCollaborator{
		PlaylistID: playlist.ID,
		UserID:     userID,
		Role:       "editor",
		AddedBy:    userID,
		AddedAt:    time.Now(),
	}

	_, err = s.collaboratorsCollection.InsertOne(context.Background(), collaborator)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

// AddCollaborator adds a user as a collaborator to a playlist
func (s *PlaylistEnhancedService) AddCollaborator(playlistID string, userID, addedBy, role string) error {
	// Check if the user adding the collaborator has permission
	hasPermission, err := s.hasCollaboratorPermission(playlistID, addedBy, "editor")
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("user does not have permission to add collaborators")
	}

	collaborator := &models.PlaylistCollaborator{
		PlaylistID: playlistID,
		UserID:     userID,
		Role:       role,
		AddedBy:    addedBy,
		AddedAt:    time.Now(),
	}

	_, err = s.collaboratorsCollection.InsertOne(context.Background(), collaborator)
	return err
}

// RemoveCollaborator removes a user from playlist collaborators
func (s *PlaylistEnhancedService) RemoveCollaborator(playlistID string, userID, removedBy string) error {
	// Check if the user removing the collaborator has permission
	hasPermission, err := s.hasCollaboratorPermission(playlistID, removedBy, "editor")
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("user does not have permission to remove collaborators")
	}

	// Don't allow removing the owner
	playlist, err := s.GetPlaylist(playlistID)
	if err != nil {
		return err
	}
	if playlist.Owner == userID {
		return fmt.Errorf("cannot remove playlist owner")
	}

	_, err = s.collaboratorsCollection.DeleteOne(context.Background(), bson.M{
		"playlist_id": playlistID,
		"user_id":     userID,
	})
	return err
}

// GetCollaborators returns all collaborators for a playlist
func (s *PlaylistEnhancedService) GetCollaborators(playlistID string) ([]models.PlaylistCollaborator, error) {
	cursor, err := s.collaboratorsCollection.Find(context.Background(), bson.M{"playlist_id": playlistID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var collaborators []models.PlaylistCollaborator
	if err = cursor.All(context.Background(), &collaborators); err != nil {
		return nil, err
	}

	return collaborators, nil
}

// AddTrackToPlaylist adds a track to a playlist
func (s *PlaylistEnhancedService) AddTrackToPlaylist(playlistID string, trackID, userID string) error {
	// Check if user has permission to add tracks
	hasPermission, err := s.hasCollaboratorPermission(playlistID, userID, "editor")
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("user does not have permission to add tracks")
	}

	// Get current track count for position
	count, err := s.tracksCollection.CountDocuments(context.Background(), bson.M{"playlist_id": playlistID})
	if err != nil {
		return err
	}

	track := &models.PlaylistTrack{
		PlaylistID: playlistID,
		TrackID:    trackID,
		Position:   int(count),
		AddedBy:    userID,
		AddedAt:    time.Now(),
	}

	_, err = s.tracksCollection.InsertOne(context.Background(), track)
	if err != nil {
		return err
	}

	// Update playlist track count
	_, err = s.playlistsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": playlistID},
		bson.M{"$inc": bson.M{"track_count": 1}},
	)
	if err != nil {
		return err
	}

	// Record activity
	activity := &models.PlaylistActivity{
		PlaylistID: playlistID,
		UserID:     userID,
		Action:     "track_added",
		Details:    map[string]interface{}{"track_id": trackID},
		CreatedAt:  time.Now(),
	}
	s.activityCollection.InsertOne(context.Background(), activity)

	return nil
}

// RemoveTrackFromPlaylist removes a track from a playlist
func (s *PlaylistEnhancedService) RemoveTrackFromPlaylist(playlistID string, trackID, userID string) error {
	// Check if user has permission to remove tracks
	hasPermission, err := s.hasCollaboratorPermission(playlistID, userID, "editor")
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("user does not have permission to remove tracks")
	}

	_, err = s.tracksCollection.DeleteOne(context.Background(), bson.M{
		"playlist_id": playlistID,
		"track_id":    trackID,
	})
	if err != nil {
		return err
	}

	// Update playlist track count
	_, err = s.playlistsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": playlistID},
		bson.M{"$inc": bson.M{"track_count": -1}},
	)
	if err != nil {
		return err
	}

	// Record activity
	activity := &models.PlaylistActivity{
		PlaylistID: playlistID,
		UserID:     userID,
		Action:     "track_removed",
		Details:    map[string]interface{}{"track_id": trackID},
		CreatedAt:  time.Now(),
	}
	s.activityCollection.InsertOne(context.Background(), activity)

	return nil
}

// FollowPlaylist makes a user follow a playlist
func (s *PlaylistEnhancedService) FollowPlaylist(playlistID string, userID string) error {
	// Check if already following
	existing := s.followsCollection.FindOne(context.Background(), bson.M{
		"playlist_id": playlistID,
		"user_id":     userID,
	})
	if existing.Err() == nil {
		return fmt.Errorf("user already follows this playlist")
	}

	follow := &models.PlaylistFollow{
		PlaylistID: playlistID,
		UserID:     userID,
		FollowedAt: time.Now(),
	}

	_, err := s.followsCollection.InsertOne(context.Background(), follow)
	if err != nil {
		return err
	}

	// Update follower count
	_, err = s.playlistsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": playlistID},
		bson.M{"$inc": bson.M{"followers": 1}},
	)

	return err
}

// UnfollowPlaylist makes a user unfollow a playlist
func (s *PlaylistEnhancedService) UnfollowPlaylist(playlistID string, userID string) error {
	_, err := s.followsCollection.DeleteOne(context.Background(), bson.M{
		"playlist_id": playlistID,
		"user_id":     userID,
	})
	if err != nil {
		return err
	}

	// Update follower count
	_, err = s.playlistsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": playlistID},
		bson.M{"$inc": bson.M{"followers": -1}},
	)

	return err
}

// LikePlaylist makes a user like a playlist
func (s *PlaylistEnhancedService) LikePlaylist(playlistID string, userID string) error {
	// Check if already liked
	existing := s.likesCollection.FindOne(context.Background(), bson.M{
		"playlist_id": playlistID,
		"user_id":     userID,
	})
	if existing.Err() == nil {
		return fmt.Errorf("user already likes this playlist")
	}

	like := &models.PlaylistLike{
		PlaylistID: playlistID,
		UserID:     userID,
		LikedAt:    time.Now(),
	}

	_, err := s.likesCollection.InsertOne(context.Background(), like)
	if err != nil {
		return err
	}

	// Update like count
	_, err = s.playlistsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": playlistID},
		bson.M{"$inc": bson.M{"likes": 1}},
	)

	return err
}

// UnlikePlaylist makes a user unlike a playlist
func (s *PlaylistEnhancedService) UnlikePlaylist(playlistID string, userID string) error {
	_, err := s.likesCollection.DeleteOne(context.Background(), bson.M{
		"playlist_id": playlistID,
		"user_id":     userID,
	})
	if err != nil {
		return err
	}

	// Update like count
	_, err = s.playlistsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": playlistID},
		bson.M{"$inc": bson.M{"likes": -1}},
	)

	return err
}

// SharePlaylist creates a share link for a playlist
func (s *PlaylistEnhancedService) SharePlaylist(playlistID string, userID, shareType string, expiresAt *time.Time) (*models.PlaylistShare, error) {
	shareToken := s.generateShareToken()

	share := &models.PlaylistShare{
		PlaylistID: playlistID,
		SharedBy:   userID,
		ShareType:  shareType,
		ShareToken: shareToken,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	result, err := s.sharesCollection.InsertOne(context.Background(), share)
	if err != nil {
		return nil, err
	}

	share.ID = result.InsertedID.(string)

	// Update share count
	_, err = s.playlistsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": playlistID},
		bson.M{"$inc": bson.M{"shares": 1}},
	)

	return share, err
}

// GetPlaylistByShareToken retrieves a playlist by share token
func (s *PlaylistEnhancedService) GetPlaylistByShareToken(shareToken string) (*models.Playlist, error) {
	var share models.PlaylistShare
	err := s.sharesCollection.FindOne(context.Background(), bson.M{"share_token": shareToken}).Decode(&share)
	if err != nil {
		return nil, err
	}

	// Check if share has expired
	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		return nil, fmt.Errorf("share link has expired")
	}

	return s.GetPlaylist(share.PlaylistID)
}

// GetPlaylistRecommendations returns playlist recommendations for a user
func (s *PlaylistEnhancedService) GetPlaylistRecommendations(userID string, limit int) ([]models.Playlist, error) {
	// Get user's listening history and preferences
	_, err := s.userService.GetUserStats(userID)
	if err != nil {
		return nil, err
	}

	// Get user's top artists and genres
	topArtists, err := s.userService.GetUserTopArtists(userID, 10)
	if err != nil {
		return nil, err
	}

	// Build recommendation query
	filter := bson.M{
		"is_public": true,
		"followers": bson.M{"$gte": 5}, // Only recommend playlists with some followers
	}

	// Add genre-based filtering if we have user preferences
	if len(topArtists) > 0 {
		// This is a simplified approach - in production, you'd use more sophisticated algorithms
		filter["tags"] = bson.M{"$in": s.extractGenresFromArtists(topArtists)}
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
	if err = cursor.All(context.Background(), &playlists); err != nil {
		return nil, err
	}

	// Store recommendations for future use
	for _, playlist := range playlists {
		recommendation := &models.PlaylistRecommendation{
			UserID:     userID,
			PlaylistID: playlist.ID,
			Score:      rand.Float64(), // Simplified scoring
			Reason:     "similar_taste",
			CreatedAt:  time.Now(),
		}
		s.recommendationsCollection.InsertOne(context.Background(), recommendation)
	}

	return playlists, nil
}

// GetPlaylist returns a playlist by ID
func (s *PlaylistEnhancedService) GetPlaylist(playlistID string) (*models.Playlist, error) {
	var playlist models.Playlist
	err := s.playlistsCollection.FindOne(context.Background(), bson.M{"_id": playlistID}).Decode(&playlist)
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

// GetUserPlaylists returns playlists for a user
func (s *PlaylistEnhancedService) GetUserPlaylists(userID string, limit int) ([]models.Playlist, error) {
	opts := options.Find().
		SetSort(bson.M{"updated_at": -1}).
		SetLimit(int64(limit))

	cursor, err := s.playlistsCollection.Find(context.Background(), bson.M{"owner_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var playlists []models.Playlist
	if err = cursor.All(context.Background(), &playlists); err != nil {
		return nil, err
	}

	return playlists, nil
}

// GetFollowedPlaylists returns playlists followed by a user
func (s *PlaylistEnhancedService) GetFollowedPlaylists(userID string, limit int) ([]models.Playlist, error) {
	// Get followed playlist IDs
	cursor, err := s.followsCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var follows []models.PlaylistFollow
	if err = cursor.All(context.Background(), &follows); err != nil {
		return nil, err
	}

	if len(follows) == 0 {
		return []models.Playlist{}, nil
	}

	playlistIDs := make([]string, len(follows))
	for i, follow := range follows {
		playlistIDs[i] = follow.PlaylistID
	}

	opts := options.Find().
		SetSort(bson.M{"updated_at": -1}).
		SetLimit(int64(limit))

	cursor, err = s.playlistsCollection.Find(context.Background(), bson.M{"_id": bson.M{"$in": playlistIDs}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var playlists []models.Playlist
	if err = cursor.All(context.Background(), &playlists); err != nil {
		return nil, err
	}

	return playlists, nil
}

// Helper methods

func (s *PlaylistEnhancedService) hasCollaboratorPermission(playlistID string, userID, requiredRole string) (bool, error) {
	// Check if user is the owner
	playlist, err := s.GetPlaylist(playlistID)
	if err != nil {
		return false, err
	}
	if playlist.Owner == userID {
		return true, nil
	}

	// Check if user is a collaborator with required role
	var collaborator models.PlaylistCollaborator
	err = s.collaboratorsCollection.FindOne(context.Background(), bson.M{
		"playlist_id": playlistID,
		"user_id":     userID,
		"role":        requiredRole,
	}).Decode(&collaborator)

	return err == nil, nil
}

func (s *PlaylistEnhancedService) generateShareToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *PlaylistEnhancedService) extractGenresFromArtists(artists []models.Artist) []string {
	genres := make(map[string]bool)
	for _, artist := range artists {
		if artist.Genre != "" {
			genres[artist.Genre] = true
		}
	}

	var result []string
	for genre := range genres {
		result = append(result, genre)
	}
	return result
}
