# Jamendo API Setup Guide

## Your Jamendo API Credentials

✅ **Client ID**: `be6cb53f`  
✅ **Client Secret**: `94b8586b8053ee3e2bb1ff3606e0e7d5`

## Configuration

### 1. Local Development

Create a `.env` file in `server/go/` directory:

```bash
# Copy the development config
cp server/go/config.dev.env server/go/.env
```

The file should contain:
```env
# Jamendo API Configuration
JAMENDO_API_KEY=be6cb53f

# Other required variables
MONGO_URI=mongodb://localhost:27017
REDIS_ADDR=localhost:6379
JWT_SECRET=your-jwt-secret-key
PORT=8080
```

### 2. Docker Compose

Set the environment variable before running Docker Compose:

```bash
# Set the Jamendo API key
export JAMENDO_API_KEY=be6cb53f

# Start services
docker-compose up -d
```

### 3. Kubernetes Production

The Kubernetes secrets are already configured with your Jamendo API key (base64 encoded).

## Testing the Jamendo Integration

### 1. Start the Go Service

```bash
cd server/go
go run main.go
```

### 2. Test Jamendo Search

```bash
# Test basic search
curl "http://localhost:8080/search?q=indie"

# Test enhanced music search
curl "http://localhost:8080/music/search?q=electronic"

# Test with pagination
curl "http://localhost:8080/music/search?q=rock&page=1&limit=10"
```

### 3. Expected Response

```json
{
  "query": "indie",
  "results": [
    {
      "id": "jamendo_123456",
      "external_id": "123456",
      "title": "Indie Song Title",
      "artist": "Indie Artist",
      "album": "Indie Album",
      "duration": 240,
      "stream_url": "https://mp3d.jamendo.com/download/track/123456/mp32/",
      "image_url": "https://imgjam.com/artists/artist_cover.jpg",
      "genre": "Indie Rock",
      "source": "jamendo",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

## Jamendo API Features

### Available Endpoints

1. **Track Search**: Search for tracks by query
2. **Artist Search**: Search for artists
3. **Album Search**: Search for albums
4. **Genre Search**: Search by genre
5. **Direct Streaming**: Full MP3 streaming URLs

### Rate Limits

- **Free Tier**: Unlimited requests
- **Rate Limit**: 2 requests per second (implemented in code)
- **No Authentication**: Required for basic usage

### Content Types

- **Creative Commons**: All tracks are CC licensed
- **Full Streaming**: Complete tracks, not previews
- **High Quality**: MP3 format, various bitrates
- **Global Coverage**: Worldwide indie music

## Troubleshooting

### Common Issues

1. **API Key Not Working**
   ```bash
   # Check if the API key is set correctly
   echo $JAMENDO_API_KEY
   ```

2. **No Results Returned**
   ```bash
   # Test with a simple query
   curl "http://localhost:8080/music/search?q=music"
   ```

3. **Rate Limit Exceeded**
   - The service automatically handles rate limiting
   - Wait a few seconds and try again

### Debug Commands

```bash
# Check service health
curl http://localhost:8080/health

# Check cache stats
curl http://localhost:8080/music/stats

# View service logs
docker-compose logs go-service
```

## Next Steps

1. **Test the Integration**: Run the test commands above
2. **Add Spotify API**: Get credentials from [developer.spotify.com](https://developer.spotify.com)
3. **Add Deezer API**: Get credentials from [developers.deezer.com](https://developers.deezer.com)
4. **Deploy to Production**: Use the Kubernetes configurations

## API Documentation

- **Jamendo API Docs**: [developer.jamendo.com](https://developer.jamendo.com)
- **Rate Limits**: 2 requests/second
- **Response Format**: JSON
- **Streaming**: Direct MP3 URLs available

Your Jamendo integration is now ready to provide Creative Commons music streaming for your Gruvit platform! 🎵
