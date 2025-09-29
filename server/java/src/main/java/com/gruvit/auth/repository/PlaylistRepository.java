package com.gruvit.auth.repository;

import com.gruvit.auth.entity.Playlist;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;
import java.util.List;

@Repository
public interface PlaylistRepository extends MongoRepository<Playlist, String> {
    List<Playlist> findByUserId(String userId);
    List<Playlist> findByUserIdAndIsPublic(String userId, boolean isPublic);
    List<Playlist> findByIsPublic(boolean isPublic);
    List<Playlist> findByIsPublicTrue();
}
