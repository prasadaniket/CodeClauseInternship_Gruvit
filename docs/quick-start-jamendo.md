# Quick Start Guide - Jamendo API Integration

## 🚀 Get Started in 5 Minutes

Your Jamendo API is now configured with:
- **Client ID**: `be6cb53f`
- **Client Secret**: `94b8586b8053ee3e2bb1ff3606e0e7d5`

## Step 1: Test the API Directly

```bash
# Navigate to the Go service directory
cd server/go

# Test Jamendo API directly
go run test_jamendo.go indie

# Test with different queries
go run test_jamendo.go electronic
go run test_jamendo.go rock
go run test_jamendo.go jazz
```

## Step 2: Start the Music Service

```bash
# Make sure MongoDB and Redis are running
# Then start the Go service
cd server/go
go run main.go
```

## Step 3: Test the Endpoints

### Basic Search
```bash
curl "http://localhost:8080/search?q=indie"
```

### Enhanced Music Search
```bash
curl "http://localhost:8080/music/search?q=electronic&limit=5"
```

### Search by Artist
```bash
curl "http://localhost:8080/music/artist?artist=Radiohead&limit=5"
```

### Health Check
```bash
curl "http://localhost:8080/health"
```

### Cache Statistics
```bash
curl "http://localhost:8080/music/stats"
```

## Expected Results

### Successful Search Response
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

### Health Check Response
```json
{
  "status": "ok",
  "service": "music-api"
}
```

## Troubleshooting

### Issue: "No results found"
**Solution**: Try different search terms:
```bash
curl "http://localhost:8080/music/search?q=music"
curl "http://localhost:8080/music/search?q=instrumental"
curl "http://localhost:8080/music/search?q=ambient"
```

### Issue: "Connection refused"
**Solution**: Make sure the service is running:
```bash
# Check if the service is running
curl http://localhost:8080/health

# If not running, start it
cd server/go
go run main.go
```

### Issue: "MongoDB connection failed"
**Solution**: Start MongoDB:
```bash
# macOS
brew services start mongodb-community

# Linux
sudo systemctl start mongod

# Windows
net start MongoDB
```

### Issue: "Redis connection failed"
**Solution**: Start Redis:
```bash
# macOS
brew services start redis

# Linux
sudo systemctl start redis

# Windows
redis-server
```

## What's Working Now

✅ **Jamendo API Integration**
- Creative Commons music streaming
- 600K+ tracks available
- Full MP3 streaming URLs
- Global indie music coverage

✅ **Search Functionality**
- Basic search endpoint
- Enhanced music search
- Artist-specific search
- Genre-based search

✅ **Caching System**
- Redis for fast responses
- MongoDB for persistent storage
- Automatic cache management

✅ **Rate Limiting**
- 2 requests/second for Jamendo
- IP-based rate limiting
- Graceful error handling

## Next Steps

1. **Add Spotify API** for mainstream music previews
2. **Add Deezer API** for regional hits
3. **Test with Frontend** integration
4. **Deploy to Production** using Kubernetes

## API Documentation

- **Jamendo API**: [developer.jamendo.com](https://developer.jamendo.com)
- **Rate Limits**: 2 requests/second
- **Content**: Creative Commons music
- **Streaming**: Full MP3 tracks

Your Gruvit music platform now has access to thousands of Creative Commons tracks from Jamendo! 🎵
