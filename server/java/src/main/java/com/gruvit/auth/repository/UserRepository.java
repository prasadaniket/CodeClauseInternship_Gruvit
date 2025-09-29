package com.gruvit.auth.repository;

import com.gruvit.auth.entity.User;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;

@Repository
public interface UserRepository extends MongoRepository<User, String> {
    Optional<User> findByUsername(String username);
    Optional<User> findByEmail(String email);
    Optional<User> findByGoogleId(String googleId);
    Optional<User> findByGithubId(String githubId);
    boolean existsByUsername(String username);
    boolean existsByEmail(String email);
}
