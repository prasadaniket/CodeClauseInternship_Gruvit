package com.gruvit.auth.entity;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import org.springframework.data.mongodb.core.index.Indexed;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Set;

@Document(collection = "users")
public final class User {
    @Id
    private String id;
    
    @NotBlank
    @Size(min = 3, max = 50)
    @Indexed(unique = true)
    private String username;
    
    @NotBlank
    @Email
    @Indexed(unique = true)
    private String email;
    
    @NotBlank
    private String passwordHash;
    
    private String firstName;
    private String lastName;
    private String profileImageUrl;
    private LocalDateTime createdAt;
    private LocalDateTime lastLoginAt;
    private boolean enabled = true;
    private boolean emailVerified = false;
    
    // 2FA fields
    private String twoFactorSecret;
    private boolean twoFactorEnabled = false;
    
    // Role-based access
    private Set<String> roles = Set.of("USER");
    
    // Playlists
    private List<String> playlistIds;
    
    // OAuth fields
    private String googleId;
    private String githubId;
    
    // Constructors
    public User() {
        this.createdAt = LocalDateTime.now();
    }
    
    public User(String username, String email, String passwordHash) {
        this();
        this.username = username;
        this.email = email;
        this.passwordHash = passwordHash;
    }
    
    // Getters and Setters
    public String getId() {
        return id;
    }
    
    public void setId(final String id) {
        this.id = id;
    }
    
    public String getUsername() {
        return username;
    }
    
    public void setUsername(final String username) {
        this.username = username;
    }
    
    public String getEmail() {
        return email;
    }
    
    public void setEmail(final String email) {
        this.email = email;
    }
    
    public String getPasswordHash() {
        return passwordHash;
    }
    
    public void setPasswordHash(final String passwordHash) {
        this.passwordHash = passwordHash;
    }
    
    public String getFirstName() {
        return firstName;
    }
    
    public void setFirstName(final String firstName) {
        this.firstName = firstName;
    }
    
    public String getLastName() {
        return lastName;
    }
    
    public void setLastName(final String lastName) {
        this.lastName = lastName;
    }
    
    public String getProfileImageUrl() {
        return profileImageUrl;
    }
    
    public void setProfileImageUrl(final String profileImageUrl) {
        this.profileImageUrl = profileImageUrl;
    }
    
    public LocalDateTime getCreatedAt() {
        return createdAt;
    }
    
    public void setCreatedAt(final LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }
    
    public LocalDateTime getLastLoginAt() {
        return lastLoginAt;
    }
    
    public void setLastLoginAt(final LocalDateTime lastLoginAt) {
        this.lastLoginAt = lastLoginAt;
    }
    
    public boolean isEnabled() {
        return enabled;
    }
    
    public void setEnabled(final boolean enabled) {
        this.enabled = enabled;
    }
    
    public boolean isEmailVerified() {
        return emailVerified;
    }
    
    public void setEmailVerified(final boolean emailVerified) {
        this.emailVerified = emailVerified;
    }
    
    public String getTwoFactorSecret() {
        return twoFactorSecret;
    }
    
    public void setTwoFactorSecret(final String twoFactorSecret) {
        this.twoFactorSecret = twoFactorSecret;
    }
    
    public boolean isTwoFactorEnabled() {
        return twoFactorEnabled;
    }
    
    public void setTwoFactorEnabled(final boolean twoFactorEnabled) {
        this.twoFactorEnabled = twoFactorEnabled;
    }
    
    public Set<String> getRoles() {
        return roles;
    }
    
    public void setRoles(final Set<String> roles) {
        this.roles = roles;
    }
    
    public List<String> getPlaylistIds() {
        return playlistIds;
    }
    
    public void setPlaylistIds(final List<String> playlistIds) {
        this.playlistIds = playlistIds;
    }
    
    public String getGoogleId() {
        return googleId;
    }
    
    public void setGoogleId(final String googleId) {
        this.googleId = googleId;
    }
    
    public String getGithubId() {
        return githubId;
    }
    
    public void setGithubId(final String githubId) {
        this.githubId = githubId;
    }
}
