package com.gruvit.auth.repository;

import com.gruvit.auth.entity.EmailVerification;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;

@Repository
public interface EmailVerificationRepository extends MongoRepository<EmailVerification, String> {
    Optional<EmailVerification> findByToken(String token);
    Optional<EmailVerification> findByUserId(String userId);
    Optional<EmailVerification> findByEmail(String email);
    void deleteByUserId(String userId);
    void deleteByEmail(String email);
}
