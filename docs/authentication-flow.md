# Authentication Flow Documentation

## Overview

The Gruvit music application uses a JWT-based authentication system with the following flow:

1. **User Registration/Login** → Java Auth Service
2. **JWT Token Generation** → Java Auth Service  
3. **Token Validation** → Go Music Service
4. **API Requests** → Frontend with JWT in headers

## Architecture

```
Frontend (Next.js)
    ↓ JWT in Authorization header
Nginx API Gateway
    ↓ Routes to appropriate service
Java Auth Service (Port 8080)
    ↓ Validates JWT
Go Music Service (Port 8081)
```

## Authentication Flow

### 1. User Registration

```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "securepassword",
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "user_id",
    "username": "user@example.com",
    "email": "user@example.com"
  }
}
```

### 2. User Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "refresh_token_here",
  "user": {
    "id": "user_id",
    "username": "user@example.com",
    "email": "user@example.com"
  },
  "expiresIn": 86400
}
```

### 3. Making Authenticated Requests

All requests to protected endpoints must include the JWT token in the Authorization header:

```http
GET /api/music/playlists
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 4. Token Refresh

```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refreshToken": "refresh_token_here"
}
```

**Response:**
```json
{
  "token": "new_jwt_token_here",
  "expiresIn": 86400
}
```

## JWT Token Structure

The JWT token contains the following claims:

```json
{
  "sub": "user_id",
  "username": "user@example.com",
  "email": "user@example.com",
  "iat": 1640995200,
  "exp": 1641081600,
  "iss": "gruvit-auth-service"
}
```

## Security Features

### 1. Token Validation

The Go service validates JWT tokens using the same secret key as the Java service:

```go
func (a *AuthMiddleware) ValidateJWT() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from Authorization header
        // Validate token signature and expiration
        // Set user context
    }
}
```

### 2. Rate Limiting

Different endpoints have different rate limits:

- **Search**: 60 requests/minute per IP
- **Stream**: 30 requests/minute per IP  
- **Playlist**: 20 requests/minute per user
- **Auth**: 10 requests/minute per IP

### 3. CORS Configuration

The Nginx gateway handles CORS for cross-origin requests:

```nginx
add_header Access-Control-Allow-Origin "http://localhost:3000" always;
add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
add_header Access-Control-Allow-Headers "Origin, Content-Type, Accept, Authorization" always;
add_header Access-Control-Allow-Credentials true always;
```

## Error Handling

### 1. Invalid Token

```json
{
  "error": "Invalid token"
}
```
**Status Code:** 401 Unauthorized

### 2. Expired Token

```json
{
  "error": "Token is not valid"
}
```
**Status Code:** 401 Unauthorized

### 3. Missing Token

```json
{
  "error": "Authorization header required"
}
```
**Status Code:** 401 Unauthorized

### 4. Rate Limit Exceeded

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```
**Status Code:** 429 Too Many Requests

## Frontend Integration

### 1. Storing Tokens

```javascript
// Store JWT token in localStorage or secure storage
localStorage.setItem('authToken', response.data.token);
localStorage.setItem('refreshToken', response.data.refreshToken);
```

### 2. Adding to Requests

```javascript
// Add JWT to all API requests
const token = localStorage.getItem('authToken');
const config = {
  headers: {
    'Authorization': `Bearer ${token}`
  }
};

axios.get('/api/music/playlists', config);
```

### 3. Token Refresh

```javascript
// Automatically refresh expired tokens
const refreshToken = async () => {
  const refreshTokenValue = localStorage.getItem('refreshToken');
  const response = await axios.post('/api/auth/refresh', {
    refreshToken: refreshTokenValue
  });
  
  localStorage.setItem('authToken', response.data.token);
  return response.data.token;
};
```

## Production Considerations

### 1. Secret Management

- Use Kubernetes secrets for JWT secrets
- Rotate secrets regularly
- Use different secrets for different environments

### 2. Token Expiration

- Set appropriate expiration times (e.g., 24 hours for access tokens)
- Implement refresh token rotation
- Handle token expiration gracefully

### 3. HTTPS

- Always use HTTPS in production
- Configure SSL certificates in Nginx
- Use secure cookie settings

### 4. Monitoring

- Log authentication attempts
- Monitor for suspicious activity
- Set up alerts for failed authentication

## Testing

### 1. Unit Tests

Test JWT validation logic:

```go
func TestJWTValidation(t *testing.T) {
    // Test valid token
    // Test expired token
    // Test invalid signature
    // Test missing token
}
```

### 2. Integration Tests

Test complete authentication flow:

```javascript
describe('Authentication Flow', () => {
  it('should register user and return JWT', async () => {
    // Test registration
  });
  
  it('should login user and return JWT', async () => {
    // Test login
  });
  
  it('should validate JWT for protected routes', async () => {
    // Test protected route access
  });
});
```

## Troubleshooting

### Common Issues

1. **CORS Errors**: Check Nginx CORS configuration
2. **Token Validation Fails**: Ensure JWT secrets match between services
3. **Rate Limiting**: Check Redis connection and rate limit configuration
4. **Token Expiration**: Implement proper token refresh logic

### Debug Commands

```bash
# Check service logs
docker-compose logs java-service
docker-compose logs go-service

# Check Redis cache
redis-cli keys "rate_limit:*"

# Validate JWT token
echo "your-jwt-token" | base64 -d
```
