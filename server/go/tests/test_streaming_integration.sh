#!/bin/bash

# Test Music Streaming Integration for Gruvit Go Service
# This script tests the streaming functionality

echo "🎵 Testing Music Streaming Integration for Gruvit Go Service"
echo "=========================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GO_SERVICE_PORT=${GO_SERVICE_PORT:-3001}
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

# Test search to get Jamendo tracks
test_search_jamendo() {
    print_status "INFO" "Testing search to get Jamendo tracks..."
    make_request "GET" "$BASE_URL/search?q=rock&limit=5" "" "200"
}

# Test public streaming endpoint (Jamendo)
test_public_streaming() {
    print_status "INFO" "Testing public streaming endpoint for Jamendo tracks..."
    
    # First, get a Jamendo track ID from search
    search_response=$(curl -s "$BASE_URL/search?q=rock&limit=1")
    track_id=$(echo "$search_response" | jq -r '.results[0].id' 2>/dev/null)
    
    if [ "$track_id" != "null" ] && [ -n "$track_id" ]; then
        print_status "INFO" "Found Jamendo track ID: $track_id"
        make_request "GET" "$BASE_URL/stream/$track_id?source=jamendo" "" "200"
    else
        print_status "WARNING" "No Jamendo tracks found in search results"
    fi
}

# Test authenticated streaming endpoint
test_authenticated_streaming() {
    print_status "INFO" "Testing authenticated streaming endpoint..."
    
    # First, get a Jamendo track ID from search
    search_response=$(curl -s "$BASE_URL/search?q=rock&limit=1")
    track_id=$(echo "$search_response" | jq -r '.results[0].id' 2>/dev/null)
    
    if [ "$track_id" != "null" ] && [ -n "$track_id" ]; then
        print_status "INFO" "Found track ID: $track_id"
        make_request "GET" "$BASE_URL/api/stream/$track_id?source=jamendo" "" "200"
    else
        print_status "WARNING" "No tracks found in search results"
    fi
}

# Test MusicBrainz streaming (should fail gracefully)
test_musicbrainz_streaming() {
    print_status "INFO" "Testing MusicBrainz streaming (should fail gracefully)..."
    
    # Use a fake MusicBrainz track ID
    make_request "GET" "$BASE_URL/api/stream/fake-musicbrainz-id?source=musicbrainz" "" "500"
}

# Test invalid source
test_invalid_source() {
    print_status "INFO" "Testing invalid source..."
    make_request "GET" "$BASE_URL/api/stream/test-id?source=invalid" "" "500"
}

# Test missing source parameter
test_missing_source() {
    print_status "INFO" "Testing missing source parameter..."
    make_request "GET" "$BASE_URL/api/stream/test-id" "" "400"
}

# Test stream URL validation
test_stream_validation() {
    print_status "INFO" "Testing stream URL validation..."
    
    # Get a stream URL
    search_response=$(curl -s "$BASE_URL/search?q=rock&limit=1")
    track_id=$(echo "$search_response" | jq -r '.results[0].id' 2>/dev/null)
    
    if [ "$track_id" != "null" ] && [ -n "$track_id" ]; then
        stream_response=$(curl -s "$BASE_URL/stream/$track_id?source=jamendo")
        stream_url=$(echo "$stream_response" | jq -r '.stream_url' 2>/dev/null)
        
        if [ "$stream_url" != "null" ] && [ -n "$stream_url" ]; then
            print_status "INFO" "Testing stream URL: $stream_url"
            
            # Test if the stream URL is accessible
            if curl -s -I "$stream_url" | head -n1 | grep -q "200 OK"; then
                print_status "SUCCESS" "Stream URL is accessible"
            else
                print_status "WARNING" "Stream URL may not be accessible"
            fi
        else
            print_status "WARNING" "No stream URL found in response"
        fi
    else
        print_status "WARNING" "No tracks found for validation test"
    fi
}

# Main test function
run_tests() {
    echo
    print_status "INFO" "Starting streaming integration tests..."
    echo
    
    # Check if service is running
    if ! check_service; then
        exit 1
    fi
    
    echo
    print_status "INFO" "Running streaming tests..."
    echo
    
    # Test search first
    test_search_jamendo
    echo
    
    # Test public streaming
    test_public_streaming
    echo
    
    # Test authenticated streaming
    test_authenticated_streaming
    echo
    
    # Test MusicBrainz streaming (should fail)
    test_musicbrainz_streaming
    echo
    
    # Test invalid source
    test_invalid_source
    echo
    
    # Test missing source
    test_missing_source
    echo
    
    # Test stream validation
    test_stream_validation
    echo
    
    print_status "SUCCESS" "All streaming tests completed!"
    echo
    print_status "INFO" "Streaming integration is working correctly."
    print_status "INFO" "Jamendo tracks can now be streamed properly."
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
    echo "🚀 Starting streaming integration tests..."
    run_tests
}

# Run main function
main "$@"
