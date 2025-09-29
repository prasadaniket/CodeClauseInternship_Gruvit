package com.gruvit.auth;

import com.gruvit.auth.entity.Playlist;
import com.gruvit.auth.entity.User;
import com.gruvit.auth.entity.UserStats;
import com.gruvit.auth.repository.PlaylistRepository;
import com.gruvit.auth.repository.UserRepository;
import com.gruvit.auth.repository.UserStatsRepository;
import com.gruvit.auth.service.JwtService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.time.LocalDateTime;
import java.util.*;

@RestController
@RequestMapping("/stats")
public class UserStatsController {
    
    @Autowired
    private UserStatsRepository userStatsRepository;
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private PlaylistRepository playlistRepository;
    
    @Autowired
    private JwtService jwtService;
    
    // Get user statistics
    @GetMapping("/user")
    public ResponseEntity<?> getUserStats(@RequestHeader("Authorization") String authHeader) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<UserStats> statsOpt = userStatsRepository.findByUserId(userId);
            UserStats stats;
            
            if (statsOpt.isEmpty()) {
                // Create new stats if they don't exist
                stats = new UserStats(userId);
                stats = userStatsRepository.save(stats);
            } else {
                stats = statsOpt.get();
            }
            
            // Update stats with current data
            updateUserStats(userId, stats);
            
            Map<String, Object> response = new HashMap<>();
            response.put("userId", stats.getUserId());
            response.put("totalPlaylists", stats.getTotalPlaylists());
            response.put("totalTracks", stats.getTotalTracks());
            response.put("totalPlayTime", stats.getTotalPlayTime());
            response.put("totalLogins", stats.getTotalLogins());
            response.put("lastActiveAt", stats.getLastActiveAt());
            response.put("favoriteGenre", stats.getFavoriteGenre());
            response.put("mostPlayedArtist", stats.getMostPlayedArtist());
            response.put("mostPlayedTrack", stats.getMostPlayedTrack());
            response.put("followersCount", stats.getFollowersCount());
            response.put("followingCount", stats.getFollowingCount());
            response.put("playlistsShared", stats.getPlaylistsShared());
            response.put("createdAt", stats.getCreatedAt());
            response.put("updatedAt", stats.getUpdatedAt());
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get user statistics"));
        }
    }
    
    // Update user activity
    @PostMapping("/activity")
    public ResponseEntity<?> updateActivity(@RequestHeader("Authorization") String authHeader,
                                          @RequestBody Map<String, Object> activityData) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<UserStats> statsOpt = userStatsRepository.findByUserId(userId);
            UserStats stats;
            
            if (statsOpt.isEmpty()) {
                stats = new UserStats(userId);
            } else {
                stats = statsOpt.get();
            }
            
            // Update activity
            stats.setLastActiveAt(LocalDateTime.now());
            stats.setUpdatedAt(LocalDateTime.now());
            
            // Update specific activity data
            if (activityData.containsKey("playTime")) {
                int playTime = (Integer) activityData.get("playTime");
                stats.setTotalPlayTime(stats.getTotalPlayTime() + playTime);
            }
            
            if (activityData.containsKey("trackPlayed")) {
                String trackTitle = (String) activityData.get("trackPlayed");
                stats.setMostPlayedTrack(trackTitle);
            }
            
            if (activityData.containsKey("artistPlayed")) {
                String artistName = (String) activityData.get("artistPlayed");
                stats.setMostPlayedArtist(artistName);
            }
            
            if (activityData.containsKey("genrePlayed")) {
                String genre = (String) activityData.get("genrePlayed");
                stats.setFavoriteGenre(genre);
            }
            
            stats = userStatsRepository.save(stats);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Activity updated"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to update activity"));
        }
    }
    
    // Get leaderboard
    @GetMapping("/leaderboard")
    public ResponseEntity<?> getLeaderboard(@RequestParam(defaultValue = "playlists") String type,
                                          @RequestParam(defaultValue = "10") int limit) {
        try {
            List<UserStats> leaderboard;
            
            switch (type.toLowerCase()) {
                case "playlists":
                    leaderboard = userStatsRepository.findAll()
                        .stream()
                        .sorted((a, b) -> Integer.compare(b.getTotalPlaylists(), a.getTotalPlaylists()))
                        .limit(limit)
                        .toList();
                    break;
                case "tracks":
                    leaderboard = userStatsRepository.findAll()
                        .stream()
                        .sorted((a, b) -> Integer.compare(b.getTotalTracks(), a.getTotalTracks()))
                        .limit(limit)
                        .toList();
                    break;
                case "playtime":
                    leaderboard = userStatsRepository.findAll()
                        .stream()
                        .sorted((a, b) -> Integer.compare(b.getTotalPlayTime(), a.getTotalPlayTime()))
                        .limit(limit)
                        .toList();
                    break;
                default:
                    return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                        .body(Map.of("error", "Invalid leaderboard type"));
            }
            
            List<Map<String, Object>> response = new ArrayList<>();
            for (int i = 0; i < leaderboard.size(); i++) {
                UserStats stats = leaderboard.get(i);
                Optional<User> userOpt = userRepository.findById(stats.getUserId());
                
                Map<String, Object> entry = new HashMap<>();
                entry.put("rank", i + 1);
                entry.put("userId", stats.getUserId());
                entry.put("username", userOpt.map(User::getUsername).orElse("Unknown"));
                entry.put("value", getLeaderboardValue(stats, type));
                entry.put("totalPlaylists", stats.getTotalPlaylists());
                entry.put("totalTracks", stats.getTotalTracks());
                entry.put("totalPlayTime", stats.getTotalPlayTime());
                
                response.add(entry);
            }
            
            return ResponseEntity.ok(Map.of("leaderboard", response, "type", type));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get leaderboard"));
        }
    }
    
    // Get user's music preferences
    @GetMapping("/preferences")
    public ResponseEntity<?> getUserPreferences(@RequestHeader("Authorization") String authHeader) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<UserStats> statsOpt = userStatsRepository.findByUserId(userId);
            if (statsOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User statistics not found"));
            }
            
            UserStats stats = statsOpt.get();
            
            Map<String, Object> response = new HashMap<>();
            response.put("favoriteGenre", stats.getFavoriteGenre());
            response.put("mostPlayedArtist", stats.getMostPlayedArtist());
            response.put("mostPlayedTrack", stats.getMostPlayedTrack());
            response.put("totalPlayTime", stats.getTotalPlayTime());
            response.put("totalTracks", stats.getTotalTracks());
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get user preferences"));
        }
    }
    
    // Helper method to update user stats
    private void updateUserStats(String userId, UserStats stats) {
        // Update playlist count
        List<Playlist> playlists = playlistRepository.findByUserId(userId);
        stats.setTotalPlaylists(playlists.size());
        
        // Update track count
        int totalTracks = playlists.stream()
            .mapToInt(playlist -> playlist.getTracks() != null ? playlist.getTracks().size() : 0)
            .sum();
        stats.setTotalTracks(totalTracks);
        
        // Update public playlists count
        long publicPlaylists = playlists.stream()
            .mapToLong(playlist -> playlist.isPublic() ? 1 : 0)
            .sum();
        stats.setPlaylistsShared((int) publicPlaylists);
        
        stats.setUpdatedAt(LocalDateTime.now());
        userStatsRepository.save(stats);
    }
    
    // Helper method to get leaderboard value
    private int getLeaderboardValue(UserStats stats, String type) {
        switch (type.toLowerCase()) {
            case "playlists":
                return stats.getTotalPlaylists();
            case "tracks":
                return stats.getTotalTracks();
            case "playtime":
                return stats.getTotalPlayTime();
            default:
                return 0;
        }
    }
}
