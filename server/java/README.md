# Gruvit Authentication Service

A robust Spring Boot microservice for authentication and user management with JWT, 2FA, and MongoDB integration.

## Features

### 🔐 Authentication
- **JWT-based authentication** with access and refresh tokens
- **Password hashing** using BCrypt
- **Role-based access control** (USER, ADMIN)
- **Two-Factor Authentication (2FA)** with Google Authenticator
- **OAuth integration** ready (Google, GitHub)

### 🛡️ Security
- **CORS configuration** for frontend integration
- **JWT token validation** for protected endpoints
- **Password strength validation**
- **Account lockout protection**
- **Email verification** support

### 📊 Database
- **MongoDB Atlas** integration
- **User profiles** with playlist management
- **Audit trails** (created, last login)
- **Flexible schema** for future expansions

## API Endpoints

### Authentication
- `POST /auth/login` - User login
- `POST /auth/signup` - User registration
- `POST /auth/refresh` - Refresh access token
- `POST /auth/validate` - Validate JWT token

### User Management
- `GET /users/profile` - Get user profile
- `PUT /users/profile` - Update user profile
- `POST /users/change-password` - Change password

### Two-Factor Authentication
- `POST /auth/2fa/setup` - Setup 2FA
- `POST /auth/2fa/verify` - Verify 2FA code

## Configuration

### Environment Variables
```bash
# MongoDB
SPRING_DATA_MONGODB_URI=mongodb+srv://username:password@cluster.mongodb.net/gruvit

# JWT
JWT_SECRET=your-secret-key-your-secret-key-32bytes

# Email (for notifications)
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=your-email@gmail.com
MAIL_PASSWORD=your-app-password

# Admin
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
```

### Application Properties
```properties
# JWT Configuration
jwt.secret=${JWT_SECRET:your-secret-key-your-secret-key-32bytes}
jwt.expiration=86400000
jwt.refresh.expiration=604800000

# CORS
spring.web.cors.allowed-origins=http://localhost:3000
```

## Database Schema

### Users Collection
```json
{
  "_id": "ObjectId",
  "username": "string",
  "email": "string",
  "passwordHash": "string",
  "firstName": "string",
  "lastName": "string",
  "profileImageUrl": "string",
  "createdAt": "DateTime",
  "lastLoginAt": "DateTime",
  "enabled": "boolean",
  "emailVerified": "boolean",
  "twoFactorSecret": "string",
  "twoFactorEnabled": "boolean",
  "roles": ["USER"],
  "playlistIds": ["ObjectId"],
  "googleId": "string",
  "githubId": "string"
}
```

### Playlists Collection
```json
{
  "_id": "ObjectId",
  "userId": "ObjectId",
  "name": "string",
  "description": "string",
  "isPublic": "boolean",
  "createdAt": "DateTime",
  "updatedAt": "DateTime",
  "tracks": [
    {
      "trackId": "string",
      "apiSource": "string",
      "title": "string",
      "artist": "string",
      "album": "string",
      "duration": "string",
      "previewUrl": "string",
      "imageUrl": "string"
    }
  ]
}
```

## Usage Examples

### Login
```bash
curl -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### Signup
```bash
curl -X POST http://localhost:8081/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "user@example.com",
    "password": "password123",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

### Get Profile
```bash
curl -X GET http://localhost:8081/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Setup 2FA
```bash
curl -X POST http://localhost:8081/auth/2fa/setup \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Running the Service

### Prerequisites
- Java 25 (or Java 21 for compatibility)
- Maven 3.6+
- MongoDB Atlas account

### Development
```bash
cd server/java
./mvnw spring-boot:run
```

### Production
```bash
cd server/java
./mvnw clean package
java -jar target/auth-service-0.0.1-SNAPSHOT.jar
```

## Security Best Practices

1. **Environment Variables**: Store sensitive data in environment variables
2. **JWT Secrets**: Use strong, random JWT secrets (32+ bytes)
3. **Password Policy**: Implement password strength requirements
4. **Rate Limiting**: Add rate limiting for login attempts
5. **HTTPS**: Always use HTTPS in production
6. **Database Security**: Use MongoDB Atlas security features

## Integration with Frontend

The service is configured for CORS with `http://localhost:3000`. For production, update the CORS configuration:

```java
@Bean
public CorsConfigurationSource corsConfigurationSource() {
    CorsConfiguration configuration = new CorsConfiguration();
    configuration.setAllowedOriginPatterns(Arrays.asList("https://yourdomain.com"));
    // ... rest of configuration
}
```

## Monitoring and Logging

- **Health Check**: `/actuator/health`
- **Metrics**: `/actuator/metrics`
- **Logs**: Configured for structured logging

## Next Steps

1. **Email Verification**: Implement email verification flow
2. **Password Reset**: Add password reset functionality
3. **OAuth Integration**: Add Google/GitHub OAuth
4. **Rate Limiting**: Implement rate limiting
5. **Audit Logging**: Add comprehensive audit trails
6. **API Documentation**: Add Swagger/OpenAPI documentation
