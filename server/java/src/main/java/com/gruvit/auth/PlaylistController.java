package com.gruvit.auth;

import com.gruvit.auth.dto.*;
import com.gruvit.auth.entity.Playlist;
import com.gruvit.auth.entity.User;
import com.gruvit.auth.repository.PlaylistRepository;
import com.gruvit.auth.repository.UserRepository;
import com.gruvit.auth.service.JwtService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import jakarta.validation.Valid;
import java.time.LocalDateTime;
import java.util.*;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/playlists")
public class PlaylistController {
    
    @Autowired
    private PlaylistRepository playlistRepository;
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private JwtService jwtService;
    
    // Create playlist
    @PostMapping
    public ResponseEntity<?> createPlaylist(@RequestHeader("Authorization") String authHeader,
                                          @RequestBody @Valid PlaylistCreateRequest request) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            Playlist playlist = new Playlist();
            playlist.setName(request.getName());
            playlist.setDescription(request.getDescription());
            playlist.setUserId(userId);
            playlist.setPublic(request.isPublic());
            playlist.setCreatedAt(LocalDateTime.now());
            playlist.setUpdatedAt(LocalDateTime.now());
            
            playlist = playlistRepository.save(playlist);
            
            // Add playlist ID to user's playlist list
            User user = userOpt.get();
            if (user.getPlaylistIds() == null) {
                user.setPlaylistIds(new ArrayList<>());
            }
            user.getPlaylistIds().add(playlist.getId());
            userRepository.save(user);
            
            Map<String, Object> response = new HashMap<>();
            response.put("id", playlist.getId());
            response.put("name", playlist.getName());
            response.put("description", playlist.getDescription());
            response.put("isPublic", playlist.isPublic());
            response.put("createdAt", playlist.getCreatedAt());
            response.put("trackCount", playlist.getTracks() != null ? playlist.getTracks().size() : 0);
            
            return ResponseEntity.status(HttpStatus.CREATED).body(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to create playlist"));
        }
    }
    
    // Get user's playlists
    @GetMapping
    public ResponseEntity<?> getUserPlaylists(@RequestHeader("Authorization") String authHeader) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            List<Playlist> playlists = playlistRepository.findByUserId(userId);
            
            List<Map<String, Object>> response = playlists.stream()
                .map(playlist -> {
                    Map<String, Object> playlistData = new HashMap<>();
                    playlistData.put("id", playlist.getId());
                    playlistData.put("name", playlist.getName());
                    playlistData.put("description", playlist.getDescription());
                    playlistData.put("isPublic", playlist.isPublic());
                    playlistData.put("createdAt", playlist.getCreatedAt());
                    playlistData.put("updatedAt", playlist.getUpdatedAt());
                    playlistData.put("trackCount", playlist.getTracks() != null ? playlist.getTracks().size() : 0);
                    return playlistData;
                })
                .collect(Collectors.toList());
            
            return ResponseEntity.ok(Map.of("playlists", response));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get playlists"));
        }
    }
    
    // Get specific playlist
    @GetMapping("/{playlistId}")
    public ResponseEntity<?> getPlaylist(@RequestHeader("Authorization") String authHeader,
                                       @PathVariable String playlistId) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<Playlist> playlistOpt = playlistRepository.findById(playlistId);
            if (playlistOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Playlist not found"));
            }
            
            Playlist playlist = playlistOpt.get();
            
            // Check if user has access to this playlist
            if (!playlist.getUserId().equals(userId) && !playlist.isPublic()) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            Map<String, Object> response = new HashMap<>();
            response.put("id", playlist.getId());
            response.put("name", playlist.getName());
            response.put("description", playlist.getDescription());
            response.put("isPublic", playlist.isPublic());
            response.put("createdAt", playlist.getCreatedAt());
            response.put("updatedAt", playlist.getUpdatedAt());
            response.put("tracks", playlist.getTracks() != null ? playlist.getTracks() : new ArrayList<>());
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get playlist"));
        }
    }
    
    // Update playlist
    @PutMapping("/{playlistId}")
    public ResponseEntity<?> updatePlaylist(@RequestHeader("Authorization") String authHeader,
                                          @PathVariable String playlistId,
                                          @RequestBody @Valid PlaylistUpdateRequest request) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<Playlist> playlistOpt = playlistRepository.findById(playlistId);
            if (playlistOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Playlist not found"));
            }
            
            Playlist playlist = playlistOpt.get();
            
            // Check if user owns this playlist
            if (!playlist.getUserId().equals(userId)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            // Update fields
            if (request.getName() != null) {
                playlist.setName(request.getName());
            }
            if (request.getDescription() != null) {
                playlist.setDescription(request.getDescription());
            }
            playlist.setPublic(request.isPublic());
            playlist.setUpdatedAt(LocalDateTime.now());
            
            playlist = playlistRepository.save(playlist);
            
            Map<String, Object> response = new HashMap<>();
            response.put("id", playlist.getId());
            response.put("name", playlist.getName());
            response.put("description", playlist.getDescription());
            response.put("isPublic", playlist.isPublic());
            response.put("updatedAt", playlist.getUpdatedAt());
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to update playlist"));
        }
    }
    
    // Delete playlist
    @DeleteMapping("/{playlistId}")
    public ResponseEntity<?> deletePlaylist(@RequestHeader("Authorization") String authHeader,
                                          @PathVariable String playlistId) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<Playlist> playlistOpt = playlistRepository.findById(playlistId);
            if (playlistOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Playlist not found"));
            }
            
            Playlist playlist = playlistOpt.get();
            
            // Check if user owns this playlist
            if (!playlist.getUserId().equals(userId)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            playlistRepository.deleteById(playlistId);
            
            // Remove playlist ID from user's playlist list
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isPresent()) {
                User user = userOpt.get();
                if (user.getPlaylistIds() != null) {
                    user.getPlaylistIds().remove(playlistId);
                    userRepository.save(user);
                }
            }
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Playlist deleted successfully"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to delete playlist"));
        }
    }
    
    // Add track to playlist
    @PostMapping("/{playlistId}/tracks")
    public ResponseEntity<?> addTrackToPlaylist(@RequestHeader("Authorization") String authHeader,
                                              @PathVariable String playlistId,
                                              @RequestBody @Valid TrackAddRequest request) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<Playlist> playlistOpt = playlistRepository.findById(playlistId);
            if (playlistOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Playlist not found"));
            }
            
            Playlist playlist = playlistOpt.get();
            
            // Check if user owns this playlist
            if (!playlist.getUserId().equals(userId)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            // Create new track
            Playlist.Track track = new Playlist.Track();
            track.setTrackId(request.getTrackId());
            track.setApiSource(request.getApiSource());
            track.setTitle(request.getTitle());
            track.setArtist(request.getArtist());
            track.setAlbum(request.getAlbum());
            track.setDuration(request.getDuration());
            track.setPreviewUrl(request.getPreviewUrl());
            track.setImageUrl(request.getImageUrl());
            
            // Add track to playlist
            if (playlist.getTracks() == null) {
                playlist.setTracks(new ArrayList<>());
            }
            playlist.getTracks().add(track);
            playlist.setUpdatedAt(LocalDateTime.now());
            
            playlist = playlistRepository.save(playlist);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Track added to playlist"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to add track to playlist"));
        }
    }
    
    // Remove track from playlist
    @DeleteMapping("/{playlistId}/tracks/{trackId}")
    public ResponseEntity<?> removeTrackFromPlaylist(@RequestHeader("Authorization") String authHeader,
                                                   @PathVariable String playlistId,
                                                   @PathVariable String trackId) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<Playlist> playlistOpt = playlistRepository.findById(playlistId);
            if (playlistOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Playlist not found"));
            }
            
            Playlist playlist = playlistOpt.get();
            
            // Check if user owns this playlist
            if (!playlist.getUserId().equals(userId)) {
                return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "Access denied"));
            }
            
            // Remove track from playlist
            if (playlist.getTracks() != null) {
                playlist.getTracks().removeIf(track -> track.getTrackId().equals(trackId));
                playlist.setUpdatedAt(LocalDateTime.now());
                playlist = playlistRepository.save(playlist);
            }
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Track removed from playlist"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to remove track from playlist"));
        }
    }
    
    // Get public playlists
    @GetMapping("/public")
    public ResponseEntity<?> getPublicPlaylists(@RequestParam(defaultValue = "0") int page,
                                              @RequestParam(defaultValue = "20") int size) {
        try {
            // This would need a custom repository method to find public playlists with pagination
            List<Playlist> playlists = playlistRepository.findByIsPublicTrue();
            
            List<Map<String, Object>> response = playlists.stream()
                .map(playlist -> {
                    Map<String, Object> playlistData = new HashMap<>();
                    playlistData.put("id", playlist.getId());
                    playlistData.put("name", playlist.getName());
                    playlistData.put("description", playlist.getDescription());
                    playlistData.put("createdAt", playlist.getCreatedAt());
                    playlistData.put("trackCount", playlist.getTracks() != null ? playlist.getTracks().size() : 0);
                    return playlistData;
                })
                .collect(Collectors.toList());
            
            return ResponseEntity.ok(Map.of("playlists", response));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get public playlists"));
        }
    }
}
