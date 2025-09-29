package com.gruvit.auth.dto;

import jakarta.validation.constraints.NotBlank;

public class TrackAddRequest {
    
    @NotBlank
    private String trackId;
    
    @NotBlank
    private String apiSource;
    
    private String title;
    private String artist;
    private String album;
    private String duration;
    private String previewUrl;
    private String imageUrl;
    
    // Constructors
    public TrackAddRequest() {}
    
    public TrackAddRequest(String trackId, String apiSource, String title, String artist) {
        this.trackId = trackId;
        this.apiSource = apiSource;
        this.title = title;
        this.artist = artist;
    }
    
    // Getters and Setters
    public String getTrackId() {
        return trackId;
    }
    
    public void setTrackId(String trackId) {
        this.trackId = trackId;
    }
    
    public String getApiSource() {
        return apiSource;
    }
    
    public void setApiSource(String apiSource) {
        this.apiSource = apiSource;
    }
    
    public String getTitle() {
        return title;
    }
    
    public void setTitle(String title) {
        this.title = title;
    }
    
    public String getArtist() {
        return artist;
    }
    
    public void setArtist(String artist) {
        this.artist = artist;
    }
    
    public String getAlbum() {
        return album;
    }
    
    public void setAlbum(String album) {
        this.album = album;
    }
    
    public String getDuration() {
        return duration;
    }
    
    public void setDuration(String duration) {
        this.duration = duration;
    }
    
    public String getPreviewUrl() {
        return previewUrl;
    }
    
    public void setPreviewUrl(String previewUrl) {
        this.previewUrl = previewUrl;
    }
    
    public String getImageUrl() {
        return imageUrl;
    }
    
    public void setImageUrl(String imageUrl) {
        this.imageUrl = imageUrl;
    }
}
