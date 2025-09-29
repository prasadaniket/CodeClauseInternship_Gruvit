package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AuthClient handles communication with the Java authentication service
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

// AuthValidationResponse represents the response from Java auth service
type AuthValidationResponse struct {
	Valid    bool   `json:"valid"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Error    string `json:"error,omitempty"`
}

// AuthLoginRequest represents login request to Java service
type AuthLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthLoginResponse represents login response from Java service
type AuthLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Roles    []string `json:"roles"`
	} `json:"user"`
	Requires2FA bool   `json:"requires2FA,omitempty"`
	Error       string `json:"error,omitempty"`
}

// AuthSignupRequest represents signup request to Java service
type AuthSignupRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// NewAuthClient creates a new authentication client
func NewAuthClient() *AuthClient {
	baseURL := os.Getenv("AUTH_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081" // Default to Java service port
	}

	return &AuthClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateToken validates a JWT token with the Java authentication service
func (ac *AuthClient) ValidateToken(tokenString string) (*AuthValidationResponse, error) {
	url := fmt.Sprintf("%s/auth/validate", ac.baseURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var authResp AuthValidationResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &authResp, fmt.Errorf("auth service error: %s", authResp.Error)
	}

	return &authResp, nil
}

// Login authenticates a user with the Java authentication service
func (ac *AuthClient) Login(username, password string) (*AuthLoginResponse, error) {
	url := fmt.Sprintf("%s/auth/login", ac.baseURL)

	loginReq := AuthLoginRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var loginResp AuthLoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &loginResp, fmt.Errorf("login failed: %s", loginResp.Error)
	}

	return &loginResp, nil
}

// Signup registers a new user with the Java authentication service
func (ac *AuthClient) Signup(username, email, password, firstName, lastName string) (*AuthLoginResponse, error) {
	url := fmt.Sprintf("%s/auth/signup", ac.baseURL)

	signupReq := AuthSignupRequest{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	jsonData, err := json.Marshal(signupReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signup request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to signup: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var signupResp AuthLoginResponse
	if err := json.Unmarshal(body, &signupResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return &signupResp, fmt.Errorf("signup failed: %s", signupResp.Error)
	}

	return &signupResp, nil
}

// RefreshToken refreshes an access token using a refresh token
func (ac *AuthClient) RefreshToken(refreshToken string) (*AuthLoginResponse, error) {
	url := fmt.Sprintf("%s/auth/refresh", ac.baseURL)

	refreshReq := map[string]string{
		"refreshToken": refreshToken,
	}

	jsonData, err := json.Marshal(refreshReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refresh request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var refreshResp AuthLoginResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &refreshResp, fmt.Errorf("token refresh failed: %s", refreshResp.Error)
	}

	return &refreshResp, nil
}

// HealthCheck checks if the authentication service is healthy
func (ac *AuthClient) HealthCheck() error {
	url := fmt.Sprintf("%s/actuator/health", ac.baseURL)

	resp, err := ac.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("auth service health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth service is unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
