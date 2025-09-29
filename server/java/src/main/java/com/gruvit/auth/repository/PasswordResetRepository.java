package com.gruvit.auth.repository;

import com.gruvit.auth.entity.PasswordReset;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;

@Repository
public interface PasswordResetRepository extends MongoRepository<PasswordReset, String> {
    Optional<PasswordReset> findByToken(String token);
    Optional<PasswordReset> findByUserId(String userId);
    Optional<PasswordReset> findByEmail(String email);
    void deleteByUserId(String userId);
    void deleteByEmail(String email);
}
