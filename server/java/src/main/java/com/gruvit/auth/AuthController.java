package com.gruvit.auth;

import com.gruvit.auth.dto.*;
import com.gruvit.auth.entity.User;
import com.gruvit.auth.repository.UserRepository;
import com.gruvit.auth.service.JwtService;
import com.gruvit.auth.service.TwoFactorService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.web.bind.annotation.*;
import jakarta.validation.Valid;
import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/auth")
public class AuthController {
    
    @Autowired
    private UserRepository userRepository;
    
    @Autowired
    private JwtService jwtService;
    
    @Autowired
    private TwoFactorService twoFactorService;
    
    private final BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();

    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody @Valid LoginRequest request) {
        try {
            Optional<User> userOpt = userRepository.findByUsername(request.getUsername());
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Invalid credentials"));
            }
            
            User user = userOpt.get();
            if (!encoder.matches(request.getPassword(), user.getPasswordHash())) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Invalid credentials"));
            }
            
            if (!user.isEnabled()) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Account is disabled"));
            }
            
            // Update last login
            user.setLastLoginAt(LocalDateTime.now());
            userRepository.save(user);
            
            // Generate tokens
            String accessToken = jwtService.generateToken(user.getUsername(), user.getId(), 
                user.getRoles().iterator().next());
            String refreshToken = jwtService.generateRefreshToken(user.getUsername(), user.getId());
            
            Map<String, Object> response = new HashMap<>();
            response.put("accessToken", accessToken);
            response.put("refreshToken", refreshToken);
            response.put("user", Map.of(
                "id", user.getId(),
                "username", user.getUsername(),
                "email", user.getEmail(),
                "roles", user.getRoles()
            ));
            
            // Check if 2FA is required
            if (user.isTwoFactorEnabled()) {
                response.put("requires2FA", true);
            }
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Login failed"));
        }
    }
    
    @PostMapping("/signup")
    public ResponseEntity<?> signup(@RequestBody @Valid SignupRequest request) {
        try {
            // Check if user already exists
            if (userRepository.existsByUsername(request.getUsername())) {
                return ResponseEntity.status(HttpStatus.CONFLICT)
                    .body(Map.of("error", "Username already exists"));
            }
            
            if (userRepository.existsByEmail(request.getEmail())) {
                return ResponseEntity.status(HttpStatus.CONFLICT)
                    .body(Map.of("error", "Email already exists"));
            }
            
            // Create new user
            User user = new User();
            user.setUsername(request.getUsername());
            user.setEmail(request.getEmail());
            user.setPasswordHash(encoder.encode(request.getPassword()));
            user.setFirstName(request.getFirstName());
            user.setLastName(request.getLastName());
            
            user = userRepository.save(user);
            
            // Generate tokens
            String accessToken = jwtService.generateToken(user.getUsername(), user.getId(), "USER");
            String refreshToken = jwtService.generateRefreshToken(user.getUsername(), user.getId());
            
            Map<String, Object> response = new HashMap<>();
            response.put("accessToken", accessToken);
            response.put("refreshToken", refreshToken);
            response.put("user", Map.of(
                "id", user.getId(),
                "username", user.getUsername(),
                "email", user.getEmail()
            ));
            
            return ResponseEntity.status(HttpStatus.CREATED).body(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Registration failed"));
        }
    }
    
    @PostMapping("/refresh")
    public ResponseEntity<?> refresh(@RequestBody RefreshRequest request) {
        try {
            if (!jwtService.isRefreshToken(request.getRefreshToken())) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Invalid refresh token"));
            }
            
            String username = jwtService.extractUsername(request.getRefreshToken());
            Optional<User> userOpt = userRepository.findByUsername(username);
            
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            String newAccessToken = jwtService.generateToken(user.getUsername(), user.getId(), 
                user.getRoles().iterator().next());
            String newRefreshToken = jwtService.generateRefreshToken(user.getUsername(), user.getId());
            
            Map<String, Object> response = new HashMap<>();
            response.put("accessToken", newAccessToken);
            response.put("refreshToken", newRefreshToken);
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("error", "Token refresh failed"));
        }
    }
    
    @PostMapping("/validate")
    public ResponseEntity<?> validate(@RequestHeader("Authorization") String authHeader) {
        try {
            if (authHeader == null || !authHeader.startsWith("Bearer ")) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Invalid authorization header"));
            }
            
            String token = authHeader.substring(7);
            if (!jwtService.validateToken(token)) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Invalid token"));
            }
            
            String userId = jwtService.extractUserId(token);
            String username = jwtService.extractUsername(token);
            String role = jwtService.extractRole(token);
            
            Map<String, Object> response = new HashMap<>();
            response.put("valid", true);
            response.put("userId", userId);
            response.put("username", username);
            response.put("role", role);
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("error", "Token validation failed"));
        }
    }
    
    @PostMapping("/2fa/setup")
    public ResponseEntity<?> setup2FA(@RequestHeader("Authorization") String authHeader) {
        try {
            String token = authHeader.substring(7);
            String userId = jwtService.extractUserId(token);
            
            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            String secret = twoFactorService.generateSecret();
            user.setTwoFactorSecret(secret);
            userRepository.save(user);
            
            String qrCodeUrl = twoFactorService.generateQRCodeUrl(user.getUsername(), secret);
            
            Map<String, Object> response = new HashMap<>();
            response.put("secret", secret);
            response.put("qrCodeUrl", qrCodeUrl);
            
            return ResponseEntity.ok(response);
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "2FA setup failed"));
        }
    }
    
    @PostMapping("/2fa/verify")
    public ResponseEntity<?> verify2FA(@RequestBody TwoFactorRequest request) {
        try {
            String token = request.getToken();
            String code = request.getCode();
            
            String userId = jwtService.extractUserId(token);
            Optional<User> userOpt = userRepository.findById(userId);
            
            if (userOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "User not found"));
            }
            
            User user = userOpt.get();
            if (twoFactorService.verifyCode(user.getTwoFactorSecret(), code)) {
                user.setTwoFactorEnabled(true);
                userRepository.save(user);
                
                return ResponseEntity.ok(Map.of("success", true));
            } else {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "Invalid 2FA code"));
            }
            
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "2FA verification failed"));
        }
    }
}


