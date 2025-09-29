# Gruvit Go Service Tests

This directory contains integration tests for the Gruvit Go service.

## Test Files

### 1. `test_auth_integration.sh`
**Purpose**: Tests the complete authentication flow between Go and Java services
**What it tests**:
- User registration
- User login
- Token validation
- Protected endpoint access
- Token refresh
- Music search (public endpoint)

**Usage**:
```bash
# Make sure both Go and Java services are running
cd server/go
./tests/test_auth_integration.sh
```

### 2. `test_mongodb_integration.sh` (Bash)
**Purpose**: Tests MongoDB integration and database operations
**What it tests**:
- Health endpoint
- Music search
- User profile (with real data)
- User favorites
- User top artists/tracks
- User followings/followers
- Playlist operations

**Usage**:
```bash
# Make sure Go service and MongoDB are running
cd server/go
./tests/test_mongodb_integration.sh
```

### 3. `test_mongodb_integration.ps1` (PowerShell)
**Purpose**: Same as above, but for Windows PowerShell
**Usage**:
```powershell
# Make sure Go service and MongoDB are running
cd server/go
.\tests\test_mongodb_integration.ps1
```

## Prerequisites

### For Auth Integration Tests:
- Go service running on port 3001
- Java auth service running on port 8081
- Both services properly configured

### For MongoDB Integration Tests:
- Go service running on port 3001
- MongoDB accessible (local or Atlas)
- Redis running (optional, for caching)

## Running All Tests

### On Linux/Mac:
```bash
cd server/go
./tests/test_auth_integration.sh
./tests/test_mongodb_integration.sh
```

### On Windows:
```powershell
cd server/go
.\tests\test_auth_integration.sh  # If you have Git Bash
.\tests\test_mongodb_integration.ps1
```

## Test Results

Each test script provides:
- ✅ **SUCCESS** - Test passed
- ❌ **ERROR** - Test failed
- ⚠️ **WARNING** - Non-critical issues
- ℹ️ **INFO** - Informational messages

## Troubleshooting

### Common Issues:

1. **Service not running**
   - Start Go service: `go run main.go`
   - Start Java service: `./mvnw spring-boot:run`

2. **MongoDB connection failed**
   - Check MONGO_URI in config.dev.env
   - Verify MongoDB is accessible

3. **Authentication failed**
   - Check AUTH_SERVICE_URL configuration
   - Verify Java service is running

4. **Port conflicts**
   - Change GO_SERVICE_PORT environment variable
   - Update test scripts with new port

## CI/CD Integration

These tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Test Auth Integration
  run: ./server/go/tests/test_auth_integration.sh

- name: Test MongoDB Integration
  run: ./server/go/tests/test_mongodb_integration.sh
```

## Development Workflow

1. **Before committing**: Run relevant tests
2. **After changes**: Verify integration still works
3. **Before deployment**: Run all tests
4. **Debugging**: Use individual test functions

## Test Data

- Tests use temporary data that doesn't affect production
- User data is created with timestamps to avoid conflicts
- Database operations are isolated per test run
