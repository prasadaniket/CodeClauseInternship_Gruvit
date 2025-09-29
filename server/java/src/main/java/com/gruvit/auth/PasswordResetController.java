package com.gruvit.auth;

import com.gruvit.auth.dto.PasswordResetConfirmRequest;
import com.gruvit.auth.dto.PasswordResetRequest;
import com.gruvit.auth.entity.PasswordReset;
import com.gruvit.auth.entity.User;
import com.gruvit.auth.repository.PasswordResetRepository;
import com.gruvit.auth.repository.UserRepository;
import com.gruvit.auth.service.EmailService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.web.bind.annotation.*;
import jakarta.validation.Valid;
import java.time.LocalDateTime;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/auth")
public class PasswordResetController {
    
    @Autowired
    private PasswordResetRepository passwordResetRepository;
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private EmailService emailService;
    
    private final BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();
    
    // Request password reset
    @PostMapping("/forgot-password")
    public ResponseEntity<?> forgotPassword(@RequestBody @Valid PasswordResetRequest request) {
        try {
            Optional<User> userOpt = userRepository.findByEmail(request.getEmail());
            if (userOpt.isEmpty()) {
                // Don't reveal if email exists or not for security
                return ResponseEntity.ok(Map.of("success", true, "message", "If the email exists, a reset link has been sent"));
            }
            
            User user = userOpt.get();
            
            // Delete any existing reset requests for this user
            passwordResetRepository.deleteByUserId(user.getId());
            
            // Create new reset request
            String token = emailService.generateToken();
            PasswordReset passwordReset = new PasswordReset(user.getId(), user.getEmail(), token);
            passwordResetRepository.save(passwordReset);
            
            // Send reset email
            emailService.sendPasswordResetEmail(user.getEmail(), user.getUsername(), token);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "If the email exists, a reset link has been sent"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to process password reset request"));
        }
    }
    
    // Confirm password reset
    @PostMapping("/reset-password")
    public ResponseEntity<?> resetPassword(@RequestBody @Valid PasswordResetConfirmRequest request) {
        try {
            Optional<PasswordReset> resetOpt = passwordResetRepository.findByToken(request.getToken());
            if (resetOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Invalid reset token"));
            }
            
            PasswordReset passwordReset = resetOpt.get();
            
            if (passwordReset.isExpired()) {
                passwordResetRepository.delete(passwordReset);
                return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Reset token has expired"));
            }
            
            if (passwordReset.isUsed()) {
                return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Reset token has already been used"));
            }
            
            // Update user password
            Optional<User> userOpt = userRepository.findById(passwordReset.getUserId());
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            user.setPasswordHash(encoder.encode(request.getNewPassword()));
            userRepository.save(user);
            
            // Mark reset as used
            passwordReset.setUsed(true);
            passwordReset.setUsedAt(LocalDateTime.now());
            passwordResetRepository.save(passwordReset);
            
            return ResponseEntity.ok(Map.of("success", true, "message", "Password reset successfully"));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to reset password"));
        }
    }
    
    // Validate reset token
    @GetMapping("/validate-reset-token")
    public ResponseEntity<?> validateResetToken(@RequestParam String token) {
        try {
            Optional<PasswordReset> resetOpt = passwordResetRepository.findByToken(token);
            if (resetOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Invalid reset token"));
            }
            
            PasswordReset passwordReset = resetOpt.get();
            
            if (passwordReset.isExpired()) {
                return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Reset token has expired"));
            }
            
            if (passwordReset.isUsed()) {
                return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Reset token has already been used"));
            }
            
            return ResponseEntity.ok(Map.of("valid", true, "email", passwordReset.getEmail()));
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to validate reset token"));
        }
    }
}
