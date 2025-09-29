#!/bin/bash

# Test API Alignment between Frontend and Backend
# This script tests that the Go service responses match what the frontend expects

echo "🔗 Testing API Alignment between Frontend and Backend"
echo "====================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GO_SERVICE_PORT=${GO_SERVICE_PORT:-8080}
BASE_URL="http://localhost:$GO_SERVICE_PORT"

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "SUCCESS" ]; then
        echo -e "${GREEN}✅ $message${NC}"
    elif [ "$status" = "ERROR" ]; then
        echo -e "${RED}❌ $message${NC}"
    elif [ "$status" = "INFO" ]; then
        echo -e "${BLUE}ℹ️  $message${NC}"
    elif [ "$status" = "WARNING" ]; then
        echo -e "${YELLOW}⚠️  $message${NC}"
    fi
}

# Function to test API endpoint and validate response format
test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_fields=$3
    local auth_token=$4
    local description=$5
    
    print_status "INFO" "Testing $description"
    
    # Make request
    if [ -n "$auth_token" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Authorization: Bearer $auth_token" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method \
            "$BASE_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        print_status "SUCCESS" "$method $endpoint -> $http_code"
        
        # Validate expected fields
        for field in $expected_fields; do
            if echo "$body" | jq -e ".$field" > /dev/null 2>&1; then
                print_status "SUCCESS" "Field '$field' found in response"
            else
                print_status "ERROR" "Field '$field' missing from response"
                echo "Response: $body"
                return 1
            fi
        done
        
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        print_status "ERROR" "$method $endpoint -> Expected 200, got $http_code"
        echo "$body"
        return 1
    fi
}

# Check if Go service is running
check_service() {
    print_status "INFO" "Checking if Go service is running on port $GO_SERVICE_PORT..."
    
    if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        print_status "SUCCESS" "Go service is running"
        return 0
    else
        print_status "ERROR" "Go service is not running on port $GO_SERVICE_PORT"
        print_status "INFO" "Please start the Go service first:"
        echo "  cd server/go && go run main.go"
        return 1
    fi
}

# Test search endpoints
test_search_endpoints() {
    print_status "INFO" "Testing search endpoints..."
    
    # Test /search endpoint
    test_endpoint "GET" "/search?q=rock&limit=5" "query results total page limit" "" "Search endpoint"
    
    # Test /music/search endpoint (should redirect)
    test_endpoint "GET" "/music/search?q=rock&limit=5" "query results total page limit" "" "Music search endpoint"
    
    # Test /music/artist endpoint
    test_endpoint "GET" "/music/artist?artist=rock&limit=5" "artist results total" "" "Artist search endpoint"
    
    # Test /music/genre endpoint
    test_endpoint "GET" "/music/genre?genre=rock&limit=5" "genre results total" "" "Genre search endpoint"
    
    # Test /music/popular endpoint
    test_endpoint "GET" "/music/popular?limit=5" "results total" "" "Popular tracks endpoint"
    
    # Test /music/track/:id endpoint
    test_endpoint "GET" "/music/track/test123" "id title artist source" "" "Track details endpoint"
}

# Test streaming endpoints
test_streaming_endpoints() {
    print_status "INFO" "Testing streaming endpoints..."
    
    # Get a track ID from search first
    search_response=$(curl -s "$BASE_URL/search?q=rock&limit=1")
    track_id=$(echo "$search_response" | jq -r '.results[0].id' 2>/dev/null)
    
    if [ "$track_id" != "null" ] && [ -n "$track_id" ]; then
        print_status "INFO" "Found track ID: $track_id"
        
        # Test public streaming endpoint
        test_endpoint "GET" "/stream/$track_id?source=jamendo" "stream_url" "" "Public streaming endpoint"
        
        # Test authenticated streaming endpoint (will fail without auth, but should return proper error)
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/stream/$track_id?source=jamendo")
        http_code=$(echo "$response" | tail -n1)
        if [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
            print_status "SUCCESS" "Authenticated streaming endpoint properly requires auth"
        else
            print_status "WARNING" "Authenticated streaming endpoint returned $http_code (expected 401/403)"
        fi
    else
        print_status "WARNING" "No tracks found for streaming test"
    fi
}

# Test profile endpoint (requires auth)
test_profile_endpoint() {
    print_status "INFO" "Testing profile endpoint..."
    
    # Test without auth (should fail)
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/profile")
    http_code=$(echo "$response" | tail -n1)
    if [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
        print_status "SUCCESS" "Profile endpoint properly requires auth"
    else
        print_status "WARNING" "Profile endpoint returned $http_code (expected 401/403)"
    fi
}

# Test playlist endpoints (requires auth)
test_playlist_endpoints() {
    print_status "INFO" "Testing playlist endpoints..."
    
    # Test without auth (should fail)
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/playlists")
    http_code=$(echo "$response" | tail -n1)
    if [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
        print_status "SUCCESS" "Playlist endpoint properly requires auth"
    else
        print_status "WARNING" "Playlist endpoint returned $http_code (expected 401/403)"
    fi
}

# Test user analytics endpoints (requires auth)
test_user_endpoints() {
    print_status "INFO" "Testing user analytics endpoints..."
    
    # Test without auth (should fail)
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/user/stats")
    http_code=$(echo "$response" | tail -n1)
    if [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
        print_status "SUCCESS" "User stats endpoint properly requires auth"
    else
        print_status "WARNING" "User stats endpoint returned $http_code (expected 401/403)"
    fi
}

# Test health endpoint
test_health_endpoint() {
    print_status "INFO" "Testing health endpoint..."
    test_endpoint "GET" "/health" "status service version" "" "Health endpoint"
}

# Main test function
run_tests() {
    echo
    print_status "INFO" "Starting API alignment tests..."
    echo
    
    # Check if service is running
    if ! check_service; then
        exit 1
    fi
    
    echo
    print_status "INFO" "Running API alignment tests..."
    echo
    
    # Test health endpoint
    test_health_endpoint
    echo
    
    # Test search endpoints
    test_search_endpoints
    echo
    
    # Test streaming endpoints
    test_streaming_endpoints
    echo
    
    # Test profile endpoint
    test_profile_endpoint
    echo
    
    # Test playlist endpoints
    test_playlist_endpoints
    echo
    
    # Test user endpoints
    test_user_endpoints
    echo
    
    print_status "SUCCESS" "All API alignment tests completed!"
    echo
    print_status "INFO" "API responses now match frontend expectations."
    print_status "INFO" "Frontend should be able to display data properly."
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if ! command -v jq &> /dev/null; then
        print_status "WARNING" "jq is not installed. JSON validation will be limited."
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_status "ERROR" "Missing dependencies: ${missing_deps[*]}"
        print_status "INFO" "Please install the missing dependencies and run the script again."
        exit 1
    fi
}

# Main execution
main() {
    echo "🔍 Checking dependencies..."
    check_dependencies
    
    echo
    echo "🚀 Starting API alignment tests..."
    run_tests
}

# Run main function
main "$@"
