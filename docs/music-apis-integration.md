# Music APIs Integration Guide

## Overview

The Gruvit music platform integrates multiple music APIs to provide comprehensive worldwide music coverage. This document outlines the API integration strategy, endpoints, and usage examples.

## API Strategy

### Primary APIs (Core Foundation)
1. **MusicBrainz** - Metadata and discovery (2M+ artists, 15M+ recordings)
2. **Jamendo** - Creative Commons streaming (600K+ tracks)

### Secondary APIs (Enhanced Coverage)
3. **MusicBrainz** - Comprehensive music metadata (40M+ recordings)

## API Coverage by Region

| Region | MusicBrainz | Jamendo | Total Coverage |
|--------|-------------|---------|---------|--------|----------------|
| North America | ✅ | ✅ | ~95% |
| Europe | ✅ | ✅ | ~98% |
| Asia | ✅ | ✅ | ~90% |
| Latin America | ✅ | ✅ | ~85% |
| Africa | ✅ | ✅ | ~80% |
| Oceania | ✅ | ✅ | ~90% |

## API Endpoints

### Search Endpoints

#### 1. General Search
```http
GET /music/search?q={query}&page={page}&limit={limit}
```

**Parameters:**
- `q` (required): Search query
- `page` (optional): Page number (default: 1)
- `limit` (optional): Results per page (default: 20, max: 100)

**Response:**
```json
{
  "query": "indie rock",
  "results": [
    {
      "id": "track_id",
      "external_id": "mbid_123",
      "title": "Song Title",
      "artist": "Artist Name",
      "album": "Album Name",
      "duration": 240,
      "stream_url": "https://...",
      "image_url": "https://...",
      "genre": "Indie Rock",
      "source": "musicbrainz"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 20
}
```

#### 2. Search by Artist
```http
GET /music/artist?artist={artist_name}&limit={limit}
```

#### 3. Search by Genre
```http
GET /music/genre?genre={genre}&limit={limit}
```

#### 4. Popular Tracks
```http
GET /music/popular?limit={limit}
```

#### 5. Track Details
```http
GET /music/track/{track_id}?source={source}
```

### Cache Management

#### 6. Cache Statistics
```http
GET /music/stats
```

**Response:**
```json
{
  "tracks_by_source": [
    {"_id": "musicbrainz", "count": 50000},
    {"_id": "jamendo", "count": 30000}
  ],
  "total_tracks": 80000,
  "total_searches": 5000
}
```

#### 7. Clean Cache
```http
POST /music/clean-cache
```

## API Integration Details

### MusicBrainz Integration

**Purpose:** Primary metadata source for global music discovery
**Rate Limit:** 1 request/second
**Authentication:** None required
**Coverage:** 2M+ artists, 15M+ recordings worldwide

```go
// Example MusicBrainz search
url := "https://musicbrainz.org/ws/2/recording/?query=recording:" + query + "&fmt=json"
```

**Strengths:**
- Comprehensive metadata
- Global coverage
- Open source
- Links to other databases

**Limitations:**
- No direct streaming
- Rate limited
- Requires pairing with streaming APIs

### Jamendo Integration

**Purpose:** Creative Commons music streaming
**Rate Limit:** 2 requests/second
**Authentication:** Client ID required
**Coverage:** 600K+ tracks, indie/CC focus

```go
// Example Jamendo search
url := fmt.Sprintf("https://api.jamendo.com/v3.0/tracks/?client_id=%s&search=%s&limit=%d", 
    clientID, query, limit)
```

**Strengths:**
- Full streaming URLs
- Creative Commons licensed
- Global indie music
- No authentication for basic usage

**Limitations:**
- Limited to CC music
- Smaller catalog
- Indie focus only

### MusicBrainz Integration

**Purpose:** Comprehensive music metadata and discovery
**Rate Limit:** 1 request/second
**Authentication:** User-Agent header required
**Coverage:** 40M+ recordings, global metadata

```go
// Example MusicBrainz search
url := "https://musicbrainz.org/ws/2/recording?query=" + query + "&fmt=json&limit=" + strconv.Itoa(limit)
```

**Strengths:**
- Comprehensive metadata
- Global coverage
- Free and open
**Limitations:**
- Metadata only (no streaming)
- Rate limited
- Commercial licensing

## Caching Strategy

### Multi-Level Caching

1. **Redis Cache** (L1)
   - Duration: 1 hour
   - Purpose: Fast response for repeated queries
   - Key format: `search:{query}`

2. **MongoDB Cache** (L2)
   - Duration: 24 hours
   - Purpose: Persistent storage and complex queries
   - Collections: `tracks`, `search_cache`, `artists`, `albums`

3. **API Response Caching**
   - Duration: Varies by API
   - Purpose: Reduce external API calls
   - Strategy: Cache successful responses, retry failed ones

### Cache Invalidation

- **Time-based:** Automatic expiration
- **Manual:** Admin endpoint for cache cleaning
- **Size-based:** LRU eviction when cache is full

## Rate Limiting Strategy

### Per-API Rate Limits

| API | Rate Limit | Strategy |
|-----|------------|----------|
| MusicBrainz | 1 req/sec | Sequential processing |
| Jamendo | 2 req/sec | Sequential processing |

### Global Rate Limiting

- **Search endpoints:** 60 requests/minute per IP
- **Stream endpoints:** 30 requests/minute per IP
- **Admin endpoints:** 10 requests/minute per IP

## Error Handling

### API Error Types

1. **Rate Limit Exceeded**
   ```json
   {
     "error": "Rate limit exceeded",
     "retry_after": 60
   }
   ```

2. **API Unavailable**
   ```json
   {
     "error": "External API temporarily unavailable",
     "fallback": "Using cached results"
   }
   ```

3. **Invalid Query**
   ```json
   {
     "error": "Query parameter 'q' is required"
   }
   ```

### Fallback Strategy

1. **Primary:** Try all APIs in parallel
2. **Secondary:** Use cached results if APIs fail
3. **Tertiary:** Return partial results from available APIs
4. **Last Resort:** Return error with retry suggestion

## Performance Optimization

### Concurrent Processing

```go
// Example concurrent API calls
go func() { musicbrainzTracks, _ = searchMusicBrainz(query) }()
go func() { jamendoTracks, _ = searchJamendo(query) }()
```

### Database Indexing

```javascript
// MongoDB indexes for optimal performance
db.tracks.createIndex({ "title": "text", "artist": "text", "album": "text" })
db.tracks.createIndex({ "source": 1, "external_id": 1 })
db.tracks.createIndex({ "created_at": -1 })
db.search_cache.createIndex({ "query": 1, "expires_at": 1 })
```

## Monitoring and Analytics

### Key Metrics

1. **API Usage**
   - Requests per API per hour
   - Success/failure rates
   - Response times

2. **Cache Performance**
   - Hit/miss ratios
   - Cache size and growth
   - Eviction rates

3. **Search Analytics**
   - Popular queries
   - Results per query
   - User engagement

### Health Checks

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "service": "music-api",
  "apis": {
    "musicbrainz": "healthy",
    "jamendo": "healthy"
  },
  "cache": {
    "redis": "healthy",
    "mongodb": "healthy"
  }
}
```

## Security Considerations

### API Key Management

- Store API keys in environment variables
- Use Kubernetes secrets in production
- Rotate keys regularly
- Monitor key usage

### Rate Limiting

- Implement per-IP rate limiting
- Use Redis for distributed rate limiting
- Add exponential backoff for retries

### Data Privacy

- Don't log sensitive user queries
- Implement query sanitization
- Use HTTPS for all API calls

## Deployment Configuration

### Environment Variables

```bash
# MusicBrainz (no auth required)
# Rate limit: 1 req/sec

# Jamendo
JAMENDO_API_KEY=your_jamendo_client_id


# Database
MONGO_URI=mongodb://localhost:27017
REDIS_ADDR=localhost:6379
```

### Docker Configuration

```dockerfile
# Add to Dockerfile
ENV JAMENDO_API_KEY=""
```

## Testing

### Unit Tests

```go
func TestMusicSearch(t *testing.T) {
    // Test individual API integrations
    // Test caching behavior
    // Test error handling
}
```

### Integration Tests

```go
func TestSearchEndpoints(t *testing.T) {
    // Test complete search flow
    // Test rate limiting
    // Test cache behavior
}
```

### Load Testing

```bash
# Example load test
hey -n 1000 -c 10 "http://localhost:8080/music/search?q=indie"
```

## Troubleshooting

### Common Issues

1. **Rate Limit Exceeded**
   - Check API key configuration
   - Implement proper rate limiting
   - Use caching to reduce API calls

2. **API Timeouts**
   - Increase timeout values
   - Implement retry logic
   - Use circuit breakers

3. **Cache Misses**
   - Check cache configuration
   - Verify MongoDB connection
   - Monitor cache statistics

### Debug Commands

```bash
# Check API status
curl http://localhost:8080/health

# Check cache stats
curl http://localhost:8080/music/stats

# Test search
curl "http://localhost:8080/music/search?q=test"

# Clean cache
curl -X POST http://localhost:8080/music/clean-cache
```

## Future Enhancements

### Planned Features

1. **Machine Learning Recommendations**
   - User preference learning
   - Collaborative filtering
   - Content-based recommendations

2. **Advanced Search**
   - Fuzzy search
   - Semantic search
   - Multi-language support

3. **Real-time Features**
   - Live music events
   - Social features
   - Real-time recommendations

4. **Additional APIs**
   - Last.fm integration
   - SoundCloud integration
   - YouTube Music integration

### Performance Improvements

1. **CDN Integration**
   - Cache static assets
   - Global content delivery
   - Reduced latency

2. **Microservices Architecture**
   - Separate API services
   - Independent scaling
   - Better fault isolation

3. **Advanced Caching**
   - Predictive caching
   - Smart cache warming
   - Distributed caching
