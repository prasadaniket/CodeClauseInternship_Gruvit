# Music Streaming Implementation - Gruvit Go Service

## Overview

The Gruvit Go service now has proper music streaming functionality implemented. This document outlines the streaming architecture, API endpoints, and how to use the streaming features.

## What Was Implemented

### 1. **Proper Jamendo Streaming**
- Real Jamendo API streaming URL resolution
- URL validation and caching
- Proper error handling

### 2. **MusicBrainz Streaming Support**
- Framework for external streaming service resolution
- Graceful error handling for non-streamable sources

### 3. **Streaming Service Architecture**
- Centralized streaming logic
- URL validation and caching
- Rate limiting and error handling

### 4. **Multiple Streaming Endpoints**
- Public streaming (Jamendo only)
- Authenticated streaming (all sources)
- Proper error responses

## API Endpoints

### Public Streaming (No Authentication Required)

**Endpoint**: `GET /stream/:trackId`

**Parameters**:
- `trackId` (path): The track ID
- `source` (query, optional): Source type (defaults to "jamendo")

**Example**:
```bash
curl "http://localhost:3001/stream/12345?source=jamendo"
```

**Response**:
```json
{
  "track_id": "12345",
  "stream_url": "https://api.jamendo.com/v3.0/tracks/stream?client_id=YOUR_KEY&id=12345",
  "expires_at": 1640995200
}
```

### Authenticated Streaming (Authentication Required)

**Endpoint**: `GET /api/stream/:trackId`

**Headers**:
- `Authorization: Bearer <token>`

**Parameters**:
- `trackId` (path): The track ID
- `source` (query, required): Source type ("jamendo" or "musicbrainz")

**Example**:
```bash
curl -H "Authorization: Bearer your_token" \
     "http://localhost:3001/api/stream/12345?source=jamendo"
```

**Response**:
```json
{
  "track_id": "12345",
  "stream_url": "https://api.jamendo.com/v3.0/tracks/stream?client_id=YOUR_KEY&id=12345",
  "expires_at": 1640995200
}
```

## Streaming Sources

### 1. **Jamendo** ✅ Fully Supported
- **Direct streaming**: Yes
- **URL format**: `https://api.jamendo.com/v3.0/tracks/stream?client_id=KEY&id=TRACK_ID`
- **Authentication**: API key required
- **Caching**: Yes (1 hour)
- **Validation**: Yes

### 2. **MusicBrainz** ⚠️ Limited Support
- **Direct streaming**: No (requires external services)
- **Status**: Returns error indicating external resolution needed
- **Future**: Can be extended to integrate with Spotify, YouTube, etc.

## Architecture

### Services

#### 1. **StreamingService**
- **Purpose**: Centralized streaming logic
- **Features**:
  - URL generation and validation
  - Caching with Redis
  - Error handling and retry logic
  - Rate limiting

#### 2. **ExternalAPIService**
- **Purpose**: External API integration
- **Features**:
  - Jamendo API integration
  - MusicBrainz API integration
  - URL validation
  - Error handling

#### 3. **StreamHandler**
- **Purpose**: HTTP request handling
- **Features**:
  - Request validation
  - Response formatting
  - Error handling

### Data Flow

```
1. Client Request → StreamHandler
2. StreamHandler → StreamingService
3. StreamingService → ExternalAPIService (if needed)
4. ExternalAPIService → External API (Jamendo)
5. Response ← StreamingService ← ExternalAPIService
6. Response ← StreamHandler ← StreamingService
7. Client ← StreamHandler
```

## Configuration

### Environment Variables

```env
# Jamendo API Configuration
JAMENDO_API_KEY=your_jamendo_api_key
JAMENDO_CLIENT_SECRET=your_jamendo_client_secret

# Redis Configuration (for caching)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# User Agent (for API requests)
USER_AGENT=Gruvit/1.0 (contact@gruvit.com)
```

### Required Setup

1. **Jamendo API Key**: Get from [Jamendo Developer](https://developer.jamendo.com/)
2. **Redis Server**: For caching stream URLs
3. **MongoDB**: For track metadata

## Usage Examples

### 1. **Search and Stream Jamendo Track**

```bash
# 1. Search for tracks
curl "http://localhost:3001/search?q=rock&limit=5"

# 2. Get track ID from response
# 3. Stream the track
curl "http://localhost:3001/stream/TRACK_ID?source=jamendo"
```

### 2. **Authenticated Streaming**

```bash
# 1. Login to get token
curl -X POST "http://localhost:3001/auth/login" \
     -H "Content-Type: application/json" \
     -d '{"username":"user","password":"pass"}'

# 2. Use token for streaming
curl -H "Authorization: Bearer TOKEN" \
     "http://localhost:3001/api/stream/TRACK_ID?source=jamendo"
```

### 3. **Frontend Integration**

```javascript
// Get stream URL
async function getStreamURL(trackId, source = 'jamendo') {
  const response = await fetch(`/stream/${trackId}?source=${source}`);
  const data = await response.json();
  return data.stream_url;
}

// Play audio
async function playTrack(trackId) {
  const streamURL = await getStreamURL(trackId);
  const audio = new Audio(streamURL);
  audio.play();
}
```

## Error Handling

### Common Errors

1. **400 Bad Request**
   - Missing track ID
   - Missing source parameter

2. **403 Forbidden**
   - Public endpoint with non-Jamendo source

3. **500 Internal Server Error**
   - Jamendo API key not configured
   - External API failure
   - MusicBrainz source (not supported)

### Error Response Format

```json
{
  "error": "Error message describing what went wrong"
}
```

## Caching Strategy

### Redis Caching

- **Stream URLs**: Cached for 1 hour
- **URL Validation**: Cached for 5 minutes
- **Keys**:
  - `stream_url:TRACK_ID` → Stream URL
  - `stream_validation:URL` → Validation result

### Cache Invalidation

- Automatic expiration
- Manual invalidation on errors
- Retry logic for failed validations

## Testing

### Test Scripts

1. **Bash**: `server/go/tests/test_streaming_integration.sh`
2. **PowerShell**: `server/go/tests/test_streaming_integration.ps1`

### Running Tests

```bash
# Bash (Linux/Mac)
cd server/go
./tests/test_streaming_integration.sh

# PowerShell (Windows)
cd server/go
.\tests\test_streaming_integration.ps1
```

### Test Coverage

- ✅ Public streaming (Jamendo)
- ✅ Authenticated streaming
- ✅ Error handling
- ✅ URL validation
- ✅ Caching
- ✅ Invalid sources
- ✅ Missing parameters

## Performance Considerations

### 1. **Caching**
- Stream URLs cached for 1 hour
- Validation results cached for 5 minutes
- Reduces external API calls

### 2. **Rate Limiting**
- Built-in rate limiting for external APIs
- Prevents API quota exhaustion
- Graceful degradation

### 3. **Error Handling**
- Fast failure for invalid sources
- Retry logic for transient errors
- Proper HTTP status codes

## Security

### 1. **Authentication**
- Public endpoint limited to Jamendo only
- Authenticated endpoint supports all sources
- JWT token validation

### 2. **API Keys**
- Jamendo API key stored in environment variables
- Not exposed in client responses
- Proper key validation

### 3. **URL Validation**
- All stream URLs validated before caching
- HEAD requests to verify accessibility
- Content type validation

## Future Enhancements

### 1. **Additional Streaming Sources**
- Spotify integration
- YouTube Music integration
- Apple Music integration

### 2. **Advanced Features**
- Adaptive bitrate streaming
- Preloading and buffering
- Analytics and usage tracking

### 3. **Performance Improvements**
- CDN integration
- Edge caching
- Load balancing

## Troubleshooting

### Common Issues

1. **"JAMENDO_API_KEY not configured"**
   - Set the environment variable
   - Restart the service

2. **"Stream URL validation failed"**
   - Check Jamendo API key validity
   - Verify network connectivity

3. **"MusicBrainz tracks require external streaming service resolution"**
   - This is expected behavior
   - MusicBrainz doesn't provide direct streaming

### Debug Mode

Enable debug logging:
```env
GIN_MODE=debug
```

## Conclusion

The music streaming implementation is now complete and functional. Users can:

- ✅ Stream Jamendo tracks directly
- ✅ Get proper streaming URLs
- ✅ Use both public and authenticated endpoints
- ✅ Benefit from caching and validation
- ✅ Handle errors gracefully

The system is ready for production use with proper configuration and monitoring.
