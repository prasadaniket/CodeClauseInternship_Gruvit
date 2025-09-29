package com.gruvit.auth;

import com.gruvit.auth.entity.Playlist;
import com.gruvit.auth.entity.User;
import com.gruvit.auth.entity.UserStats;
import com.gruvit.auth.repository.PlaylistRepository;
import com.gruvit.auth.repository.UserRepository;
import com.gruvit.auth.repository.UserStatsRepository;
import com.gruvit.auth.service.JwtService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.time.LocalDateTime;
import java.util.*;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/admin")
public class AdminController {
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private PlaylistRepository playlistRepository;
    
    @Autowired
    private UserStatsRepository userStatsRepository;
    
    @Autowired
    private JwtService jwtService;
    
    // Check if user is admin
    private boolean isAdmin(String token) {
        try {
            String role = jwtService.extractRole(token);
            return "ADMIN".equals(role);
        } catch (Exception e) {
            return false;
        }
    }
    
    // Get all users with pagination
    @GetMapping("/users")
    public ResponseEntity<?> getAllUsers(@RequestHeader("Authorization") String authHeader,
                                       @RequestParam(defaultValue = "0") int page,
                                       @RequestParam(defaultValue = "20") int size) {
        try {
            String token = authHeader.substring(7);
            if (!isAdmin(token)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            List<User> users = userRepository.findAll();
            
            // Simple pagination
            int start = page * size;
            int end = Math.min(start + size, users.size());
            List<User> paginatedUsers = users.subList(start, end);
            
            List<Map<String, Object>> response = paginatedUsers.stream()
                .map(user -> {
                    Map<String, Object> userData = new HashMap<>();
                    userData.put("id", user.getId());
                    userData.put("username", user.getUsername());
                    userData.put("email", user.getEmail());
                    userData.put("firstName", user.getFirstName());
                    userData.put("lastName", user.getLastName());
                    userData.put("enabled", user.isEnabled());
                    userData.put("emailVerified", user.isEmailVerified());
                    userData.put("twoFactorEnabled", user.isTwoFactorEnabled());
                    userData.put("roles", user.getRoles());
                    userData.put("createdAt", user.getCreatedAt());
                    userData.put("lastLoginAt", user.getLastLoginAt());
                    return userData;
                })
                .collect(Collectors.toList());
            
            Map<String, Object> result = new HashMap<>();
            result.put("users", response);
            result.put("total", users.size());
            result.put("page", page);
            result.put("size", size);
            
            return ResponseEntity.ok(result);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get users"));
        }
    }
    
    // Get user by ID
    @GetMapping("/users/{userId}")
    public ResponseEntity<?> getUserById(@RequestHeader("Authorization") String authHeader,
                                      @PathVariable String userId) {
        try {
            String token = authHeader.substring(7);
            if (!isAdmin(token)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            Map<String, Object> response = new HashMap<>();
            response.put("id", user.getId());
            response.put("username", user.getUsername());
            response.put("email", user.getEmail());
            response.put("firstName", user.getFirstName());
            response.put("lastName", user.getLastName());
            response.put("enabled", user.isEnabled());
            response.put("emailVerified", user.isEmailVerified());
            response.put("twoFactorEnabled", user.isTwoFactorEnabled());
            response.put("roles", user.getRoles());
            response.put("createdAt", user.getCreatedAt());
            response.put("lastLoginAt", user.getLastLoginAt());
            response.put("playlistIds", user.getPlaylistIds());
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get user"));
        }
    }
    
    // Update user status
    @PutMapping("/users/{userId}/status")
    public ResponseEntity<?> updateUserStatus(@RequestHeader("Authorization") String authHeader,
                                            @PathVariable String userId,
                                            @RequestBody Map<String, Object> statusData) {
        try {
            String token = authHeader.substring(7);
            if (!isAdmin(token)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            
            if (statusData.containsKey("enabled")) {
                user.setEnabled((Boolean) statusData.get("enabled"));
            }
            
            if (statusData.containsKey("emailVerified")) {
                user.setEmailVerified((Boolean) statusData.get("emailVerified"));
            }
            
            if (statusData.containsKey("roles")) {
                @SuppressWarnings("unchecked")
                Set<String> roles = new HashSet<>((List<String>) statusData.get("roles"));
                user.setRoles(roles);
            }
            
            user = userRepository.save(user);
            
            Map<String, Object> response = new HashMap<>();
            response.put("id", user.getId());
            response.put("username", user.getUsername());
            response.put("enabled", user.isEnabled());
            response.put("emailVerified", user.isEmailVerified());
            response.put("roles", user.getRoles());
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to update user status"));
        }
    }
    
    // Delete user
    @DeleteMapping("/users/{userId}")
    public ResponseEntity<?> deleteUser(@RequestHeader("Authorization") String authHeader,
                                      @PathVariable String userId) {
        try {
            String token = authHeader.substring(7);
            if (!isAdmin(token)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            // Delete user's playlists
            List<Playlist> userPlaylists = playlistRepository.findByUserId(userId);
            playlistRepository.deleteAll(userPlaylists);
            
            // Delete user's stats
            userStatsRepository.deleteByUserId(userId);
            
            // Delete user
            userRepository.deleteById(userId);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "User deleted successfully"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to delete user"));
        }
    }
    
    // Get system statistics
    @GetMapping("/stats")
    public ResponseEntity<?> getSystemStats(@RequestHeader("Authorization") String authHeader) {
        try {
            String token = authHeader.substring(7);
            if (!isAdmin(token)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            List<User> allUsers = userRepository.findAll();
            List<Playlist> allPlaylists = playlistRepository.findAll();
            List<UserStats> allStats = userStatsRepository.findAll();
            
            Map<String, Object> response = new HashMap<>();
            response.put("totalUsers", allUsers.size());
            response.put("activeUsers", allUsers.stream().mapToLong(user -> user.isEnabled() ? 1 : 0).sum());
            response.put("verifiedUsers", allUsers.stream().mapToLong(user -> user.isEmailVerified() ? 1 : 0).sum());
            response.put("twoFactorUsers", allUsers.stream().mapToLong(user -> user.isTwoFactorEnabled() ? 1 : 0).sum());
            response.put("totalPlaylists", allPlaylists.size());
            response.put("publicPlaylists", allPlaylists.stream().mapToLong(playlist -> playlist.isPublic() ? 1 : 0).sum());
            response.put("totalTracks", allPlaylists.stream()
                .mapToInt(playlist -> playlist.getTracks() != null ? playlist.getTracks().size() : 0)
                .sum());
            response.put("totalPlayTime", allStats.stream().mapToInt(UserStats::getTotalPlayTime).sum());
            
            // Recent activity
            LocalDateTime last24Hours = LocalDateTime.now().minusHours(24);
            long recentLogins = allUsers.stream()
                .filter(user -> user.getLastLoginAt() != null && user.getLastLoginAt().isAfter(last24Hours))
                .count();
            response.put("recentLogins", recentLogins);
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get system statistics"));
        }
    }
    
    // Get all playlists with admin view
    @GetMapping("/playlists")
    public ResponseEntity<?> getAllPlaylists(@RequestHeader("Authorization") String authHeader,
                                           @RequestParam(defaultValue = "0") int page,
                                           @RequestParam(defaultValue = "20") int size) {
        try {
            String token = authHeader.substring(7);
            if (!isAdmin(token)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            List<Playlist> playlists = playlistRepository.findAll();
            
            // Simple pagination
            int start = page * size;
            int end = Math.min(start + size, playlists.size());
            List<Playlist> paginatedPlaylists = playlists.subList(start, end);
            
            List<Map<String, Object>> response = paginatedPlaylists.stream()
                .map(playlist -> {
                    Map<String, Object> playlistData = new HashMap<>();
                    playlistData.put("id", playlist.getId());
                    playlistData.put("name", playlist.getName());
                    playlistData.put("description", playlist.getDescription());
                    playlistData.put("userId", playlist.getUserId());
                    playlistData.put("isPublic", playlist.isPublic());
                    playlistData.put("createdAt", playlist.getCreatedAt());
                    playlistData.put("updatedAt", playlist.getUpdatedAt());
                    playlistData.put("trackCount", playlist.getTracks() != null ? playlist.getTracks().size() : 0);
                    return playlistData;
                })
                .collect(Collectors.toList());
            
            Map<String, Object> result = new HashMap<>();
            result.put("playlists", response);
            result.put("total", playlists.size());
            result.put("page", page);
            result.put("size", size);
            
            return ResponseEntity.ok(result);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get playlists"));
        }
    }
    
    // Delete playlist (admin)
    @DeleteMapping("/playlists/{playlistId}")
    public ResponseEntity<?> deletePlaylist(@RequestHeader("Authorization") String authHeader,
                                         @PathVariable String playlistId) {
        try {
            String token = authHeader.substring(7);
            if (!isAdmin(token)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            Optional<Playlist> playlistOpt = playlistRepository.findById(playlistId);
            if (playlistOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Playlist not found"));
            }
            
            playlistRepository.deleteById(playlistId);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Playlist deleted successfully"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to delete playlist"));
        }
    }
}
