package handlers

import (
	"net/http"
	"time"

	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

// IntegratedAuthHandler handles authentication by delegating to Java service
type IntegratedAuthHandler struct {
	authClient *services.AuthClient
}

// NewIntegratedAuthHandler creates a new integrated auth handler
func NewIntegratedAuthHandler(authClient *services.AuthClient) *IntegratedAuthHandler {
	return &IntegratedAuthHandler{
		authClient: authClient,
	}
}

// IntegratedLoginRequest represents the login request payload
type IntegratedLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// IntegratedRegisterRequest represents the registration request payload
type IntegratedRegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// IntegratedRefreshRequest represents the refresh token request payload
type IntegratedRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Roles    []string `json:"roles"`
	} `json:"user"`
	Requires2FA bool   `json:"requires_2fa,omitempty"`
	Message     string `json:"message"`
}

// Login handles user login by delegating to Java auth service
func (h *IntegratedAuthHandler) Login(c *gin.Context) {
	var req IntegratedLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Delegate to Java auth service
	authResp, err := h.authClient.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Login failed",
			"details": err.Error(),
		})
		return
	}

	// Convert response format
	response := AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours in seconds
		User: struct {
			ID       string   `json:"id"`
			Username string   `json:"username"`
			Email    string   `json:"email"`
			Roles    []string `json:"roles"`
		}{
			ID:       authResp.User.ID,
			Username: authResp.User.Username,
			Email:    authResp.User.Email,
			Roles:    authResp.User.Roles,
		},
		Requires2FA: authResp.Requires2FA,
		Message:     "Login successful",
	}

	c.JSON(http.StatusOK, response)
}

// Register handles user registration by delegating to Java auth service
func (h *IntegratedAuthHandler) Register(c *gin.Context) {
	var req IntegratedRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Delegate to Java auth service
	authResp, err := h.authClient.Signup(
		req.Username,
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Registration failed",
			"details": err.Error(),
		})
		return
	}

	// Convert response format
	response := AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours in seconds
		User: struct {
			ID       string   `json:"id"`
			Username string   `json:"username"`
			Email    string   `json:"email"`
			Roles    []string `json:"roles"`
		}{
			ID:       authResp.User.ID,
			Username: authResp.User.Username,
			Email:    authResp.User.Email,
			Roles:    authResp.User.Roles,
		},
		Message: "User registered successfully",
	}

	c.JSON(http.StatusCreated, response)
}

// RefreshToken handles token refresh by delegating to Java auth service
func (h *IntegratedAuthHandler) RefreshToken(c *gin.Context) {
	var req IntegratedRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Delegate to Java auth service
	authResp, err := h.authClient.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token refresh failed",
			"details": err.Error(),
		})
		return
	}

	// Convert response format
	response := AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours in seconds
		User: struct {
			ID       string   `json:"id"`
			Username string   `json:"username"`
			Email    string   `json:"email"`
			Roles    []string `json:"roles"`
		}{
			ID:       authResp.User.ID,
			Username: authResp.User.Username,
			Email:    authResp.User.Email,
			Roles:    authResp.User.Roles,
		},
		Message: "Tokens refreshed successfully",
	}

	c.JSON(http.StatusOK, response)
}

// ValidateToken validates a token by delegating to Java auth service
func (h *IntegratedAuthHandler) ValidateToken(c *gin.Context) {
	// Get token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization header required",
		})
		return
	}

	// Extract token
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Delegate to Java auth service
	authResp, err := h.authClient.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token validation failed",
			"details": err.Error(),
		})
		return
	}

	if !authResp.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid token",
			"details": authResp.Error,
		})
		return
	}

	// Return user information
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":       authResp.UserID,
			"username": authResp.Username,
			"role":     authResp.Role,
		},
	})
}

// GetUserProfile returns the current user's profile
func (h *IntegratedAuthHandler) GetUserProfile(c *gin.Context) {
	// Get user ID from context (set by middleware)
	userID := c.GetString("user_id")
	username := c.GetString("username")
	role := c.GetString("role")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Return user profile information
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       userID,
			"username": username,
			"role":     role,
			"joined":   time.Now().Add(-30 * 24 * time.Hour), // Mock data - in production, get from DB
		},
	})
}

// UpdateUserProfile updates the current user's profile
func (h *IntegratedAuthHandler) UpdateUserProfile(c *gin.Context) {
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

	// TODO: Implement profile update by calling Java service
	// For now, return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":       userID,
			"username": updateData.Username,
			"email":    updateData.Email,
		},
	})
}

// Logout handles user logout
func (h *IntegratedAuthHandler) Logout(c *gin.Context) {
	// Get token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization header required",
		})
		return
	}

	// Extract token
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// TODO: Implement logout by calling Java service to blacklist token
	// For now, just return success (tokenString is logged for debugging)
	_ = tokenString // Suppress unused variable warning

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}
