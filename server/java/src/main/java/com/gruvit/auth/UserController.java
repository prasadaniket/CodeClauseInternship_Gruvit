package com.gruvit.auth;

import com.gruvit.auth.dto.*;
import com.gruvit.auth.entity.User;
import com.gruvit.auth.repository.UserRepository;
import com.gruvit.auth.service.JwtService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import jakarta.validation.Valid;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/users")
public class UserController {
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private JwtService jwtService;
    
    @GetMapping("/profile")
    public ResponseEntity<?> getProfile(@RequestHeader("Authorization") String authHeader) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            Map<String, Object> profile = new HashMap<>();
            profile.put("id", user.getId());
            profile.put("username", user.getUsername());
            profile.put("email", user.getEmail());
            profile.put("firstName", user.getFirstName());
            profile.put("lastName", user.getLastName());
            profile.put("profileImageUrl", user.getProfileImageUrl());
            profile.put("createdAt", user.getCreatedAt());
            profile.put("lastLoginAt", user.getLastLoginAt());
            profile.put("emailVerified", user.isEmailVerified());
            profile.put("twoFactorEnabled", user.isTwoFactorEnabled());
            profile.put("roles", user.getRoles());
            profile.put("playlistIds", user.getPlaylistIds());
            
            return ResponseEntity.ok(profile);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get profile"));
        }
    }
    
    @PutMapping("/profile")
    public ResponseEntity<?> updateProfile(@RequestHeader("Authorization") String authHeader,
                                         @RequestBody @Valid ProfileUpdateRequest request) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            
            // Update allowed fields
            if (request.getFirstName() != null) {
                user.setFirstName(request.getFirstName());
            }
            if (request.getLastName() != null) {
                user.setLastName(request.getLastName());
            }
            if (request.getProfileImageUrl() != null) {
                user.setProfileImageUrl(request.getProfileImageUrl());
            }
            
            user = userRepository.save(user);
            
            Map<String, Object> response = new HashMap<>();
            response.put("id", user.getId());
            response.put("username", user.getUsername());
            response.put("email", user.getEmail());
            response.put("firstName", user.getFirstName());
            response.put("lastName", user.getLastName());
            response.put("profileImageUrl", user.getProfileImageUrl());
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to update profile"));
        }
    }
    
    @PostMapping("/change-password")
    public ResponseEntity<?> changePassword(@RequestHeader("Authorization") String authHeader,
                                         @RequestBody @Valid PasswordChangeRequest request) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            
            // Verify current password
            if (!user.getPasswordHash().equals(request.getCurrentPassword())) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Current password is incorrect"));
            }
            
            // Update password
            user.setPasswordHash(request.getNewPassword());
            userRepository.save(user);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Password updated successfully"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to change password"));
        }
    }
}
