package services

import (
	"context"
	"time"

	"gruvit/server/go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlaylistService struct {
	collection *mongo.Collection
}

func NewPlaylistService(db *mongo.Database) *PlaylistService {
	return &PlaylistService{
		collection: db.Collection("playlists"),
	}
}

func (s *PlaylistService) CreatePlaylist(userID, name, description string, isPublic bool) (*models.Playlist, error) {
	playlist := &models.Playlist{
		Owner:       userID,
		Name:        name,
		Description: description,
		Tracks:      []models.Track{},
		IsPublic:    isPublic,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := s.collection.InsertOne(context.Background(), playlist)
	if err != nil {
		return nil, err
	}

	playlist.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return playlist, nil
}

func (s *PlaylistService) GetPlaylistsByUser(userID string) ([]models.Playlist, error) {
	filter := bson.M{"owner": userID}
	cursor, err := s.collection.Find(context.Background(), filter)
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

func (s *PlaylistService) GetPlaylistByID(playlistID string) (*models.Playlist, error) {
	objectID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var playlist models.Playlist
	err = s.collection.FindOne(context.Background(), filter).Decode(&playlist)
	if err != nil {
		return nil, err
	}

	return &playlist, nil
}

func (s *PlaylistService) UpdatePlaylist(playlistID, userID, name, description string, isPublic bool) (*models.Playlist, error) {
	objectID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID, "owner": userID}
	update := bson.M{
		"$set": bson.M{
			"name":        name,
			"description": description,
			"is_public":   isPublic,
			"updated_at":  time.Now(),
		},
	}

	_, err = s.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}

	return s.GetPlaylistByID(playlistID)
}

func (s *PlaylistService) DeletePlaylist(playlistID, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID, "owner": userID}
	_, err = s.collection.DeleteOne(context.Background(), filter)
	return err
}

func (s *PlaylistService) AddTrackToPlaylist(playlistID, userID string, track models.Track) error {
	objectID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID, "owner": userID}
	update := bson.M{
		"$push": bson.M{
			"tracks": track,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = s.collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (s *PlaylistService) RemoveTrackFromPlaylist(playlistID, userID, trackExternalID string) error {
	objectID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID, "owner": userID}
	update := bson.M{
		"$pull": bson.M{
			"tracks": bson.M{"id": trackExternalID},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = s.collection.UpdateOne(context.Background(), filter, update)
	return err
}
