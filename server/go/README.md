# Gruvit Music API Service

A Go-based microservice for handling music search, streaming, and playlist management.

## Features

- **Music Search**: Search tracks from multiple external APIs (Jamendo, MusicBrainz)
- **Streaming**: Get streaming URLs for tracks
- **Playlist Management**: Full CRUD operations for user playlists
- **Caching**: Redis-based caching for improved performance
- **Authentication**: JWT-based authentication integration
- **Docker Support**: Containerized for easy deployment

## API Endpoints

### Public Endpoints
- `GET /health` - Health check
- `GET /search?q={query}` - Search for tracks (cached)

### Protected Endpoints (require JWT)
- `GET /api/stream/{trackId}` - Get streaming URL for a track
- `POST /api/playlists` - Create a new playlist
- `GET /api/playlists` - Get user's playlists
- `GET /api/playlists/{id}` - Get specific playlist
- `PUT /api/playlists/{id}` - Update playlist
- `DELETE /api/playlists/{id}` - Delete playlist
- `POST /api/playlists/{id}/tracks` - Add track to playlist
- `DELETE /api/playlists/{id}/tracks?track_id={id}` - Remove track from playlist

### Public Playlist Endpoints
- `GET /public/playlists/{id}` - Get public playlist (optional auth)

## Environment Variables

Create a `.env` file with the following variables:

```env
# Database Configuration
MONGO_URI=mongodb://localhost:27017

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-jwt-secret-key-change-this-in-production

# External API Keys
JAMENDO_API_KEY=your-jamendo-api-key

# Server Configuration
PORT=8080
```

## Running the Service

### Using Docker Compose (Recommended)

```bash
docker-compose up -d
```

This will start the music API service along with MongoDB and Redis.

### Running Locally

1. Install dependencies:
```bash
go mod tidy
```

2. Start MongoDB and Redis services

3. Set up environment variables

4. Run the service:
```bash
go run main.go
```

## External API Integration

### Jamendo API
- Register at https://developer.jamendo.com/
- Get your API key and set `JAMENDO_API_KEY` in environment variables

### MusicBrainz API
- No API key required
- Rate limited to 1 request per second

## Project Structure

```
server/go/
├── handlers/          # HTTP request handlers
│   ├── playlist.go   # Playlist management endpoints
│   ├── search.go     # Search functionality
│   └── stream.go     # Streaming endpoints
├── middleware/        # HTTP middleware
│   └── auth.go       # JWT authentication
├── models/           # Data models
│   └── track.go      # Track and playlist models
├── services/         # Business logic services
│   ├── external_api.go    # External API integration
│   ├── playlist_service.go # Playlist operations
│   └── redis_service.go   # Redis caching
├── main.go           # Application entry point
├── Dockerfile        # Docker configuration
└── docker-compose.yml # Development setup
```

## Development

### Adding New External APIs

1. Add the API client to `services/external_api.go`
2. Update the `SearchTracks` method to include the new API
3. Add appropriate data mapping in the response parsing

### Adding New Endpoints

1. Create handler functions in the appropriate handler file
2. Register routes in `main.go`
3. Add authentication middleware if needed

## Performance Considerations

- Search results are cached in Redis for 1 hour
- Stream URLs are cached for 1 hour
- Concurrent requests are handled efficiently with Go's goroutines
- Database connections are pooled for optimal performance

## Security

- JWT tokens are validated for protected endpoints
- CORS is configured for frontend integration
- Input validation is performed on all endpoints
- Non-root user is used in Docker container
