# 🔐 Authentication Integration Guide

This guide explains how to integrate the Go music service with the Java authentication service for real user authentication.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Go Service    │    │  Java Service   │
│   (Next.js)     │◄──►│  (Music API)    │◄──►│  (Auth API)     │
│   Port: 3000    │    │   Port: 3001    │    │   Port: 8081    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### 1. Start Java Authentication Service
```bash
cd server/java
./mvnw spring-boot:run
```

### 2. Switch to Integrated Authentication
```bash
cd server/go
./switch-to-integrated.sh
```

### 3. Start Go Music Service
```bash
go run main.go
```

## 🔄 Switching Between Auth Modes

### Switch to Integrated Auth (Real Users)
```bash
./switch-to-integrated.sh
```

### Switch to Mock Auth (Development)
```bash
./switch-to-mock.sh
```

## 📋 Authentication Flow

### 1. User Registration
```bash
curl -X POST http://localhost:3001/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "firstName": "Test",
    "lastName": "User"
  }'
```

### 2. User Login
```bash
curl -X POST http://localhost:3001/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### 3. Access Protected Endpoints
```bash
curl -X GET http://localhost:3001/api/playlists \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 🔧 Configuration

### Environment Variables

#### Go Service (`config.dev.env`)
```env
# Auth Service Configuration
AUTH_SERVICE_URL=http://localhost:8081

# JWT Configuration (must match Java service)
JWT_SECRET=gruvit-super-secret-jwt-key-change-this-in-production-2024

# Other configurations...
MONGO_URI=mongodb+srv://...
REDIS_ADDR=localhost:6379
JAMENDO_API_KEY=be6cb53f
```

#### Java Service (`application.properties`)
```properties
# JWT Configuration (must match Go service)
jwt.secret=gruvit-super-secret-jwt-key-change-this-in-production-2024

# Server Configuration
server.port=8081
spring.data.mongodb.uri=mongodb+srv://...
```

## 🏗️ Implementation Details

### Auth Client (`services/auth_client.go`)
- HTTP client for communicating with Java auth service
- Handles token validation, login, registration, and refresh
- Includes health checks and error handling

### Auth Middleware (`middleware/auth.go`)
- Validates JWT tokens with Java service
- Sets user context for protected routes
- Supports optional authentication for public endpoints

### Integrated Auth Handlers (`handlers/integrated_auth.go`)
- Delegates authentication to Java service
- Provides consistent API interface
- Handles token refresh and user profile management

## 🔒 Security Features

### JWT Token Validation
- Tokens are validated against Java service
- Proper issuer and audience validation
- Token blacklisting support (logout)

### User Context
- User ID, username, and role available in request context
- Automatic user identification for protected endpoints
- Role-based access control support

### Error Handling
- Graceful fallback when auth service is unavailable
- Detailed error messages for debugging
- Proper HTTP status codes

## 🧪 Testing

### Test Authentication Flow
```bash
# 1. Register a user
curl -X POST http://localhost:3001/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "email": "test@example.com", "password": "password123"}'

# 2. Login
TOKEN=$(curl -s -X POST http://localhost:3001/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}' | jq -r '.access_token')

# 3. Access protected endpoint
curl -X GET http://localhost:3001/api/profile \
  -H "Authorization: Bearer $TOKEN"
```

### Health Check
```bash
curl http://localhost:3001/health
```

Response:
```json
{
  "status": "ok",
  "service": "music-api-integrated",
  "version": "1.0.0",
  "auth_service": true
}
```

## 🐳 Docker Configuration

### Docker Compose
The `docker-compose.yml` includes:
- Java auth service on port 8080
- Go music service on port 8081
- Shared JWT secret
- Auth service URL configuration

### Kubernetes
The K8s configurations include:
- Service discovery between Java and Go services
- Shared secrets and config maps
- Health checks and readiness probes

## 🚨 Troubleshooting

### Common Issues

#### 1. Auth Service Connection Failed
```
Warning: Auth service health check failed: connection refused
```
**Solution**: Ensure Java auth service is running on port 8081

#### 2. JWT Secret Mismatch
```
Token validation failed: invalid token issuer
```
**Solution**: Ensure both services use the same JWT_SECRET

#### 3. CORS Issues
```
Access to fetch at 'http://localhost:3001/auth/login' from origin 'http://localhost:3000' has been blocked by CORS policy
```
**Solution**: Check CORS configuration in both services

### Debug Mode
Enable debug logging:
```bash
export GIN_MODE=debug
go run main.go
```

## 📚 API Endpoints

### Authentication Endpoints
- `POST /auth/login` - User login
- `POST /auth/register` - User registration
- `POST /auth/refresh` - Token refresh
- `POST /auth/validate` - Token validation
- `GET /auth/profile` - Get user profile
- `PUT /auth/profile` - Update user profile
- `POST /auth/logout` - User logout

### Protected Endpoints
- `GET /api/playlists` - Get user playlists
- `POST /api/playlists` - Create playlist
- `GET /api/stream/:trackId` - Stream music
- `GET /api/profile` - Get user profile

### Public Endpoints
- `GET /search` - Search music
- `GET /health` - Health check

## 🔄 Migration from Mock Auth

1. **Backup current implementation**:
   ```bash
   cp main.go main_mock_auth.go.backup
   ```

2. **Switch to integrated auth**:
   ```bash
   ./switch-to-integrated.sh
   ```

3. **Test authentication flow**:
   ```bash
   # Test with real user registration/login
   curl -X POST http://localhost:3001/auth/register ...
   ```

4. **Update frontend**:
   - Update API endpoints to use new auth flow
   - Handle new response formats
   - Implement token refresh logic

## 🎯 Next Steps

1. **Implement token blacklisting** in Java service
2. **Add rate limiting** for auth endpoints
3. **Implement 2FA support** in Go service
4. **Add user profile management** endpoints
5. **Implement password reset** functionality
6. **Add audit logging** for auth events

## 📞 Support

For issues or questions:
1. Check the troubleshooting section
2. Review the logs for both services
3. Verify configuration matches between services
4. Test with curl commands provided above
