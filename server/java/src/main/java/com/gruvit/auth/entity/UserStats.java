package com.gruvit.auth.entity;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import org.springframework.data.mongodb.core.mapping.Field;
import java.time.LocalDateTime;

@Document(collection = "user_stats")
public final class UserStats {
    @Id
    private String id;
    
    @Field("user_id")
    private String userId;
    
    private int totalPlaylists = 0;
    private int totalTracks = 0;
    private int totalPlayTime = 0; // in seconds
    private int totalLogins = 0;
    private LocalDateTime lastActiveAt;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
    
    // Music preferences
    private String favoriteGenre;
    private String mostPlayedArtist;
    private String mostPlayedTrack;
    
    // Social stats
    private int followersCount = 0;
    private int followingCount = 0;
    private int playlistsShared = 0;
    
    // Constructors
    public UserStats() {
        this.createdAt = LocalDateTime.now();
        this.updatedAt = LocalDateTime.now();
    }
    
    public UserStats(final String userId) {
        this();
        this.userId = userId;
    }
    
    // Getters and Setters
    public String getId() {
        return id;
    }
    
    public void setId(final String id) {
        this.id = id;
    }
    
    public String getUserId() {
        return userId;
    }
    
    public void setUserId(final String userId) {
        this.userId = userId;
    }
    
    public int getTotalPlaylists() {
        return totalPlaylists;
    }
    
    public void setTotalPlaylists(final int totalPlaylists) {
        this.totalPlaylists = totalPlaylists;
    }
    
    public int getTotalTracks() {
        return totalTracks;
    }
    
    public void setTotalTracks(final int totalTracks) {
        this.totalTracks = totalTracks;
    }
    
    public int getTotalPlayTime() {
        return totalPlayTime;
    }
    
    public void setTotalPlayTime(final int totalPlayTime) {
        this.totalPlayTime = totalPlayTime;
    }
    
    public int getTotalLogins() {
        return totalLogins;
    }
    
    public void setTotalLogins(final int totalLogins) {
        this.totalLogins = totalLogins;
    }
    
    public LocalDateTime getLastActiveAt() {
        return lastActiveAt;
    }
    
    public void setLastActiveAt(final LocalDateTime lastActiveAt) {
        this.lastActiveAt = lastActiveAt;
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
    
    public String getFavoriteGenre() {
        return favoriteGenre;
    }
    
    public void setFavoriteGenre(final String favoriteGenre) {
        this.favoriteGenre = favoriteGenre;
    }
    
    public String getMostPlayedArtist() {
        return mostPlayedArtist;
    }
    
    public void setMostPlayedArtist(final String mostPlayedArtist) {
        this.mostPlayedArtist = mostPlayedArtist;
    }
    
    public String getMostPlayedTrack() {
        return mostPlayedTrack;
    }
    
    public void setMostPlayedTrack(final String mostPlayedTrack) {
        this.mostPlayedTrack = mostPlayedTrack;
    }
    
    public int getFollowersCount() {
        return followersCount;
    }
    
    public void setFollowersCount(final int followersCount) {
        this.followersCount = followersCount;
    }
    
    public int getFollowingCount() {
        return followingCount;
    }
    
    public void setFollowingCount(final int followingCount) {
        this.followingCount = followingCount;
    }
    
    public int getPlaylistsShared() {
        return playlistsShared;
    }
    
    public void setPlaylistsShared(final int playlistsShared) {
        this.playlistsShared = playlistsShared;
    }
}
