package com.gruvit.auth.dto;

import jakarta.validation.constraints.Size;

public class PlaylistUpdateRequest {
    
    @Size(min = 1, max = 100, message = "Playlist name must be between 1 and 100 characters")
    private String name;
    
    @Size(max = 500, message = "Description must not exceed 500 characters")
    private String description;
    
    private Boolean isPublic;
    
    // Constructors
    public PlaylistUpdateRequest() {}
    
    public PlaylistUpdateRequest(String name, String description, Boolean isPublic) {
        this.name = name;
        this.description = description;
        this.isPublic = isPublic;
    }
    
    // Getters and Setters
    public String getName() {
        return name;
    }
    
    public void setName(String name) {
        this.name = name;
    }
    
    public String getDescription() {
        return description;
    }
    
    public void setDescription(String description) {
        this.description = description;
    }
    
    public Boolean isPublic() {
        return isPublic;
    }
    
    public void setPublic(Boolean isPublic) {
        this.isPublic = isPublic;
    }
}
