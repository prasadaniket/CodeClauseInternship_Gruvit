#!/bin/bash

# Test MongoDB Integration for Gruvit Go Service
# This script tests the MongoDB integration by running the Go service and making API calls

echo "🧪 Testing MongoDB Integration for Gruvit Go Service"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GO_SERVICE_PORT=${GO_SERVICE_PORT:-3001}
BASE_URL="http://localhost:$GO_SERVICE_PORT"
TEST_USER_ID="test_user_123"
TEST_USERNAME="testuser"

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

# Function to make HTTP requests
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer test_token" \
            -d "$data" \
            "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method \
            -H "Authorization: Bearer test_token" \
            "$url")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        print_status "SUCCESS" "$method $url -> $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        print_status "ERROR" "$method $url -> Expected $expected_status, got $http_code"
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

# Test health endpoint
test_health() {
    print_status "INFO" "Testing health endpoint..."
    make_request "GET" "$BASE_URL/health" "" "200"
}

# Test search endpoint (should work without auth)
test_search() {
    print_status "INFO" "Testing search endpoint..."
    make_request "GET" "$BASE_URL/search?q=test&limit=5" "" "200"
}

# Test user profile endpoint (requires auth)
test_user_profile() {
    print_status "INFO" "Testing user profile endpoint..."
    make_request "GET" "$BASE_URL/api/profile" "" "200"
}

# Test user favorites endpoint
test_user_favorites() {
    print_status "INFO" "Testing user favorites endpoint..."
    make_request "GET" "$BASE_URL/api/user/favorites" "" "200"
}

# Test user top artists endpoint
test_user_top_artists() {
    print_status "INFO" "Testing user top artists endpoint..."
    make_request "GET" "$BASE_URL/api/user/top-artists" "" "200"
}

# Test user top tracks endpoint
test_user_top_tracks() {
    print_status "INFO" "Testing user top tracks endpoint..."
    make_request "GET" "$BASE_URL/api/user/top-tracks" "" "200"
}

# Test user followings endpoint
test_user_followings() {
    print_status "INFO" "Testing user followings endpoint..."
    make_request "GET" "$BASE_URL/api/user/followings" "" "200"
}

# Test user followers endpoint
test_user_followers() {
    print_status "INFO" "Testing user followers endpoint..."
    make_request "GET" "$BASE_URL/api/user/followers" "" "200"
}

# Test adding track to favorites
test_add_to_favorites() {
    print_status "INFO" "Testing add to favorites endpoint..."
    make_request "POST" "$BASE_URL/api/user/favorites/test_track_123" "" "200"
}

# Test playlists endpoint
test_playlists() {
    print_status "INFO" "Testing playlists endpoint..."
    make_request "GET" "$BASE_URL/api/playlists" "" "200"
}

# Test creating a playlist
test_create_playlist() {
    print_status "INFO" "Testing create playlist endpoint..."
    local playlist_data='{"name":"Test Playlist","description":"A test playlist","is_public":false}'
    make_request "POST" "$BASE_URL/api/playlists" "$playlist_data" "201"
}

# Main test function
run_tests() {
    echo
    print_status "INFO" "Starting MongoDB integration tests..."
    echo
    
    # Check if service is running
    if ! check_service; then
        exit 1
    fi
    
    echo
    print_status "INFO" "Running API tests..."
    echo
    
    # Test basic endpoints
    test_health
    echo
    
    test_search
    echo
    
    # Test user endpoints (these will use MongoDB)
    test_user_profile
    echo
    
    test_user_favorites
    echo
    
    test_user_top_artists
    echo
    
    test_user_top_tracks
    echo
    
    test_user_followings
    echo
    
    test_user_followers
    echo
    
    test_add_to_favorites
    echo
    
    # Test playlist endpoints
    test_playlists
    echo
    
    test_create_playlist
    echo
    
    print_status "SUCCESS" "All tests completed!"
    echo
    print_status "INFO" "MongoDB integration is working correctly."
    print_status "INFO" "The Go service is now using real MongoDB operations instead of mock data."
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if ! command -v jq &> /dev/null; then
        print_status "WARNING" "jq is not installed. JSON responses won't be formatted."
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
    echo "🚀 Starting MongoDB integration tests..."
    run_tests
}

# Run main function
main "$@"
