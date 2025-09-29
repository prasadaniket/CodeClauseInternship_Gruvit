package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"` // "refresh"
	jwt.RegisteredClaims
}

func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey:     []byte(secretKey),
		accessExpiry:  24 * time.Hour,     // 24 hours
		refreshExpiry: 7 * 24 * time.Hour, // 7 days
	}
}

// GenerateTokenPair creates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(userID, username, email, role string) (*TokenPair, error) {
	// Generate access token
	accessToken, err := j.generateAccessToken(userID, username, email, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	// Generate refresh token
	refreshToken, err := j.generateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.accessExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// generateAccessToken creates an access token with user information
func (j *JWTService) generateAccessToken(userID, username, email, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "gruvit-auth-service",
			Subject:   userID,
			Audience:  []string{"gruvit-music-service"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// generateRefreshToken creates a refresh token
func (j *JWTService) generateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := &RefreshTokenClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "gruvit-auth-service",
			Subject:   userID,
			Audience:  []string{"gruvit-music-service"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateAccessToken validates an access token and returns claims
func (j *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Additional validation
		if claims.Issuer != "gruvit-auth-service" {
			return nil, errors.New("invalid token issuer")
		}
		if !contains(claims.Audience, "gruvit-music-service") {
			return nil, errors.New("invalid token audience")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken validates a refresh token and returns user ID
func (j *JWTService) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse refresh token: %v", err)
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		if claims.Type != "refresh" {
			return "", errors.New("invalid token type")
		}
		if claims.Issuer != "gruvit-auth-service" {
			return "", errors.New("invalid token issuer")
		}
		return claims.UserID, nil
	}

	return "", errors.New("invalid refresh token")
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (j *JWTService) RefreshAccessToken(refreshTokenString string, userID, username, email, role string) (*TokenPair, error) {
	// Validate refresh token
	validatedUserID, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %v", err)
	}

	// Ensure the user ID matches
	if validatedUserID != userID {
		return nil, errors.New("user ID mismatch in refresh token")
	}

	// Generate new token pair
	return j.GenerateTokenPair(userID, username, email, role)
}

// ExtractTokenFromHeader extracts token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Check if it starts with "Bearer "
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	return authHeader[7:], nil
}

// GetTokenExpiry returns the expiry time of a token
func (j *JWTService) GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt.Time, nil
}

// IsTokenExpired checks if a token is expired
func (j *JWTService) IsTokenExpired(tokenString string) bool {
	expiry, err := j.GetTokenExpiry(tokenString)
	if err != nil {
		return true
	}
	return time.Now().After(expiry)
}

// RevokeToken adds a token to a blacklist (in production, use Redis)
func (j *JWTService) RevokeToken(tokenString string) error {
	// In a production environment, you would:
	// 1. Add the token to a Redis blacklist
	// 2. Set expiry to match the token's expiry
	// 3. Check blacklist during token validation

	// For now, we'll just log the revocation
	fmt.Printf("Token revoked: %s\n", tokenString[:20]+"...")
	return nil
}

// BlacklistToken checks if a token is blacklisted
func (j *JWTService) BlacklistToken(tokenString string) bool {
	// In production, check Redis blacklist
	// For now, return false (not blacklisted)
	return false
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetDefaultClaims returns default claims for testing
func (j *JWTService) GetDefaultClaims() *Claims {
	return &Claims{
		UserID:   "test-user-id",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "gruvit-auth-service",
			Subject:   "test-user-id",
			Audience:  []string{"gruvit-music-service"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpiry)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}
