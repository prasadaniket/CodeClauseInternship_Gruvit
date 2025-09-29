package com.gruvit.auth.entity;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import org.springframework.data.mongodb.core.mapping.Field;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import java.time.LocalDateTime;
import java.util.List;

@Document(collection = "playlists")
public final class Playlist {
    @Id
    private String id;
    
    @NotBlank
    @Size(min = 1, max = 100)
    private String name;
    
    @Field("user_id")
    private String userId;
    
    private String description;
    private boolean isPublic = false;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
    
    private List<Track> tracks;
    
    // Constructors
    public Playlist() {
        this.createdAt = LocalDateTime.now();
        this.updatedAt = LocalDateTime.now();
    }
    
    public Playlist(final String name, final String userId) {
        this();
        this.name = name;
        this.userId = userId;
    }
    
    // Inner Track class
    public static class Track {
        private String trackId;
        private String apiSource; // "jamendo", "musicbrainz", etc.
        private String title;
        private String artist;
        private String album;
        private String duration;
        private String previewUrl;
        private String imageUrl;
        
        // Constructors
        public Track() {}
        
        public Track(final String trackId, final String apiSource) {
            this.trackId = trackId;
            this.apiSource = apiSource;
        }
        
        // Getters and Setters
        public String getTrackId() {
            return trackId;
        }
        
        public void setTrackId(final String trackId) {
            this.trackId = trackId;
        }
        
        public String getApiSource() {
            return apiSource;
        }
        
        public void setApiSource(final String apiSource) {
            this.apiSource = apiSource;
        }
        
        public String getTitle() {
            return title;
        }
        
        public void setTitle(final String title) {
            this.title = title;
        }
        
        public String getArtist() {
            return artist;
        }
        
        public void setArtist(final String artist) {
            this.artist = artist;
        }
        
        public String getAlbum() {
            return album;
        }
        
        public void setAlbum(final String album) {
            this.album = album;
        }
        
        public String getDuration() {
            return duration;
        }
        
        public void setDuration(final String duration) {
            this.duration = duration;
        }
        
        public String getPreviewUrl() {
            return previewUrl;
        }
        
        public void setPreviewUrl(final String previewUrl) {
            this.previewUrl = previewUrl;
        }
        
        public String getImageUrl() {
            return imageUrl;
        }
        
        public void setImageUrl(final String imageUrl) {
            this.imageUrl = imageUrl;
        }
    }
    
    // Getters and Setters
    public String getId() {
        return id;
    }
    
    public void setId(final String id) {
        this.id = id;
    }
    
    public String getName() {
        return name;
    }
    
    public void setName(final String name) {
        this.name = name;
    }
    
    public String getUserId() {
        return userId;
    }
    
    public void setUserId(final String userId) {
        this.userId = userId;
    }
    
    public String getDescription() {
        return description;
    }
    
    public void setDescription(final String description) {
        this.description = description;
    }
    
    public boolean isPublic() {
        return isPublic;
    }
    
    public void setPublic(final boolean isPublic) {
        this.isPublic = isPublic;
    }
    
    public LocalDateTime getCreatedAt() {
        return createdAt;
    }
    
    public void setCreatedAt(final LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }
    
    public LocalDateTime getUpdatedAt() {
        return updatedAt;
    }
    
    public void setUpdatedAt(final LocalDateTime updatedAt) {
        this.updatedAt = updatedAt;
    }
    
    public List<Track> getTracks() {
        return tracks;
    }
    
    public void setTracks(final List<Track> tracks) {
        this.tracks = tracks;
    }
}
