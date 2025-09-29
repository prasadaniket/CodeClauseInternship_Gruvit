package com.gruvit.auth;

import com.gruvit.auth.entity.EmailVerification;
import com.gruvit.auth.entity.User;
import com.gruvit.auth.repository.EmailVerificationRepository;
import com.gruvit.auth.repository.UserRepository;
import com.gruvit.auth.service.EmailService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.time.LocalDateTime;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/auth")
public class EmailVerificationController {
    
    @Autowired
    private EmailVerificationRepository emailVerificationRepository;
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private EmailService emailService;
    
    // Send verification email
    @PostMapping("/send-verification")
    public ResponseEntity<?> sendVerificationEmail(@RequestHeader("Authorization") String authHeader) {
        try {
            // Extract user ID from token (you'll need to implement this)
            String userId = extractUserIdFromToken(authHeader);
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            
            if (user.isEmailVerified()) {
                return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Email already verified"));
            }
            
            // Delete any existing verification for this user
            emailVerificationRepository.deleteByUserId(userId);
            
            // Create new verification
            String token = emailService.generateToken();
            EmailVerification verification = new EmailVerification(userId, user.getEmail(), token);
            emailVerificationRepository.save(verification);
            
            // Send email
            emailService.sendVerificationEmail(user.getEmail(), user.getUsername(), token);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Verification email sent"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to send verification email"));
        }
    }
    
    // Verify email
    @PostMapping("/verify-email")
    public ResponseEntity<?> verifyEmail(@RequestParam String token) {
        try {
            Optional<EmailVerification> verificationOpt = emailVerificationRepository.findByToken(token);
            if (verificationOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Invalid verification token"));
            }
            
            EmailVerification verification = verificationOpt.get();
            
            if (verification.isExpired()) {
                emailVerificationRepository.delete(verification);
                return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Verification token has expired"));
            }
            
            if (verification.isVerified()) {
                return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Email already verified"));
            }
            
            // Update user
            Optional<User> userOpt = userRepository.findById(verification.getUserId());
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            user.setEmailVerified(true);
            userRepository.save(user);
            
            // Update verification
            verification.setVerified(true);
            verification.setVerifiedAt(LocalDateTime.now());
            emailVerificationRepository.save(verification);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Email verified successfully"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to verify email"));
        }
    }
    
    // Check verification status
    @GetMapping("/verification-status")
    public ResponseEntity<?> getVerificationStatus(@RequestHeader("Authorization") String authHeader) {
        try {
            String userId = extractUserIdFromToken(authHeader);
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            
            Map<String, Object> response = Map.of(
                "emailVerified", user.isEmailVerified(),
                "email", user.getEmail()
            );
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to get verification status"));
        }
    }
    
    // Helper method to extract user ID from token
    private String extractUserIdFromToken(String authHeader) {
        // This is a simplified implementation
        // In a real application, you would use JWT service to extract user ID
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            throw new IllegalArgumentException("Invalid authorization header");
        }
        
        String token = authHeader.substring(7);
        // You would use your JWT service here to extract user ID
        // For now, return a placeholder
        return "user-id-placeholder";
    }
}
