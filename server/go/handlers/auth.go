package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gruvit/server/go/models"
	"gruvit/server/go/services"
)

type AuthHandler struct {
	jwtService *services.JWTService
}

func NewAuthHandler(jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RefreshRequest represents the refresh token request payload
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	TokenPair *services.TokenPair `json:"token_pair"`
	User      *models.User        `json:"user"`
	Message   string              `json:"message"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// TODO: Implement user registration logic
	// For now, we'll create a mock user
	userID := "user_" + req.Username
	user := &models.User{
		ID:        userID,
		Username:  req.Username,
		Email:     req.Email,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate token pair
	tokenPair, err := h.jwtService.GenerateTokenPair(
		user.ID,
		user.Username,
		user.Email,
		user.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate tokens",
		})
		return
	}

	c.JSON(http.StatusCreated, LoginResponse{
		TokenPair: tokenPair,
		User:      user,
		Message:   "User registered successfully",
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// TODO: Implement actual authentication logic
	// For now, we'll create a mock user for any valid credentials
	userID := "user_" + req.Username
	user := &models.User{
		ID:        userID,
		Username:  req.Username,
		Email:     req.Username + "@example.com",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Generate token pair
	tokenPair, err := h.jwtService.GenerateTokenPair(
		user.ID,
		user.Username,
		user.Email,
		user.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate tokens",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		TokenPair: tokenPair,
		User:      user,
		Message:   "Login successful",
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate refresh token and get user ID
	userID, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
		})
		return
	}

	// TODO: Get user details from database
	// For now, we'll use mock data
	user := &models.User{
		ID:        userID,
		Username:  "user_" + userID,
		Email:     userID + "@example.com",
		Role:      "user",
		UpdatedAt: time.Now(),
	}

	// Generate new token pair
	tokenPair, err := h.jwtService.GenerateTokenPair(
		user.ID,
		user.Username,
		user.Email,
		user.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate new tokens",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		TokenPair: tokenPair,
		User:      user,
		Message:   "Tokens refreshed successfully",
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from header
	authHeader := c.GetHeader("Authorization")
	tokenString, err := h.jwtService.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid authorization header",
		})
		return
	}

	// Revoke token (add to blacklist)
	err = h.jwtService.RevokeToken(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// ValidateToken validates a token and returns user information
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// Get token from header
	authHeader := c.GetHeader("Authorization")
	tokenString, err := h.jwtService.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid authorization header",
		})
		return
	}

	// Validate token
	claims, err := h.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
		return
	}

	// Check if token is blacklisted
	if h.jwtService.BlacklistToken(tokenString) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token has been revoked",
		})
		return
	}

	// Return user information
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":       claims.UserID,
			"username": claims.Username,
			"email":    claims.Email,
			"role":     claims.Role,
		},
		"expires_at": claims.ExpiresAt.Time,
	})
}

// GetUserProfile returns the current user's profile
func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	// Get user ID from context (set by middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// TODO: Get user details from database
	// For now, we'll return mock data
	user := &models.User{
		ID:        userID,
		Username:  c.GetString("username"),
		Email:     userID + "@example.com",
		Role:      "user",
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateUserProfile updates the current user's profile
func (h *AuthHandler) UpdateUserProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var updateData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// TODO: Update user in database
	// For now, we'll return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":       userID,
			"username": updateData.Username,
			"email":    updateData.Email,
		},
	})
}
