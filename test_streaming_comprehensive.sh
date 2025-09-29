#!/bin/bash

# Comprehensive Streaming Functionality Test Script
# This script tests the complete streaming flow from backend to frontend

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_URL="http://localhost:3001"
JAMENDO_API_KEY="be6cb53f"
TEST_TRACK_ID="12345"  # We'll get a real one from search

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Helper functions
print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} $message"
            ;;
    esac
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    
    print_status "INFO" "Testing: $description"
    print_status "INFO" "Endpoint: $method $endpoint"
    
    response=$(curl -s -w "\n%{http_code}" "$BACKEND_URL$endpoint" -X "$method" 2>/dev/null || echo "000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        print_status "SUCCESS" "$description - Status: $http_code"
        if [ -n "$body" ] && [ "$body" != "null" ]; then
            print_status "INFO" "Response: $body"
        fi
        return 0
    else
        print_status "ERROR" "$description - Expected: $expected_status, Got: $http_code"
        if [ -n "$body" ]; then
            print_status "ERROR" "Response: $body"
        fi
        return 1
    fi
}

# Test 1: Backend Health Check
test_backend_health() {
    echo -e "\n${BLUE}=== TEST 1: Backend Health Check ===${NC}"
    
    if curl -s "$BACKEND_URL/health" > /dev/null 2>&1; then
        print_status "SUCCESS" "Backend is running and accessible"
    else
        print_status "ERROR" "Backend is not accessible at $BACKEND_URL"
        print_status "INFO" "Make sure the backend server is running on port 3001"
        return 1
    fi
}

# Test 2: Search for Real Tracks
test_search_tracks() {
    echo -e "\n${BLUE}=== TEST 2: Search for Real Tracks ===${NC}"
    
    print_status "INFO" "Searching for tracks to get real track IDs..."
    
    search_response=$(curl -s "$BACKEND_URL/search?q=rock&limit=5" 2>/dev/null || echo "{}")
    
    if echo "$search_response" | jq -e '.results' > /dev/null 2>&1; then
        track_count=$(echo "$search_response" | jq '.results | length')
        if [ "$track_count" -gt 0 ]; then
            # Get the first track ID
            TEST_TRACK_ID=$(echo "$search_response" | jq -r '.results[0].id' 2>/dev/null)
            track_title=$(echo "$search_response" | jq -r '.results[0].title' 2>/dev/null)
            track_artist=$(echo "$search_response" | jq -r '.results[0].artist' 2>/dev/null)
            
            print_status "SUCCESS" "Found $track_count tracks"
            print_status "INFO" "Using track: '$track_title' by '$track_artist' (ID: $TEST_TRACK_ID)"
            return 0
        else
            print_status "ERROR" "No tracks found in search results"
            return 1
        fi
    else
        print_status "ERROR" "Invalid search response format"
        print_status "ERROR" "Response: $search_response"
        return 1
    fi
}

# Test 3: Public Streaming Endpoint
test_public_streaming() {
    echo -e "\n${BLUE}=== TEST 3: Public Streaming Endpoint ===${NC}"
    
    if [ -z "$TEST_TRACK_ID" ] || [ "$TEST_TRACK_ID" = "null" ]; then
        print_status "ERROR" "No valid track ID available for testing"
        return 1
    fi
    
    print_status "INFO" "Testing public streaming endpoint with track ID: $TEST_TRACK_ID"
    
    stream_response=$(curl -s "$BACKEND_URL/stream/$TEST_TRACK_ID?source=jamendo" 2>/dev/null || echo "{}")
    
    if echo "$stream_response" | jq -e '.stream_url' > /dev/null 2>&1; then
        stream_url=$(echo "$stream_response" | jq -r '.stream_url')
        if [ "$stream_url" != "null" ] && [ -n "$stream_url" ]; then
            print_status "SUCCESS" "Public streaming endpoint returned valid stream URL"
            print_status "INFO" "Stream URL: $stream_url"
            
            # Test if the stream URL is accessible
            print_status "INFO" "Testing stream URL accessibility..."
            if curl -s -I "$stream_url" | head -n1 | grep -q "200 OK"; then
                print_status "SUCCESS" "Stream URL is accessible and returns 200 OK"
            else
                print_status "WARNING" "Stream URL may not be accessible (this could be normal for some APIs)"
            fi
            return 0
        else
            print_status "ERROR" "Stream URL is null or empty"
            return 1
        fi
    else
        print_status "ERROR" "Invalid streaming response format"
        print_status "ERROR" "Response: $stream_response"
        return 1
    fi
}

# Test 4: Authenticated Streaming Endpoint (without auth - should fail)
test_authenticated_streaming_no_auth() {
    echo -e "\n${BLUE}=== TEST 4: Authenticated Streaming (No Auth) ===${NC}"
    
    if [ -z "$TEST_TRACK_ID" ] || [ "$TEST_TRACK_ID" = "null" ]; then
        print_status "ERROR" "No valid track ID available for testing"
        return 1
    fi
    
    print_status "INFO" "Testing authenticated streaming endpoint without authentication (should fail)"
    
    response=$(curl -s -w "\n%{http_code}" "$BACKEND_URL/api/stream/$TEST_TRACK_ID?source=jamendo" 2>/dev/null || echo "000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
        print_status "SUCCESS" "Authenticated endpoint properly requires authentication (Status: $http_code)"
        return 0
    else
        print_status "ERROR" "Authenticated endpoint should require auth, but got status: $http_code"
        print_status "ERROR" "Response: $body"
        return 1
    fi
}

# Test 5: Invalid Track ID
test_invalid_track_id() {
    echo -e "\n${BLUE}=== TEST 5: Invalid Track ID Handling ===${NC}"
    
    print_status "INFO" "Testing with invalid track ID (should fail gracefully)"
    
    response=$(curl -s -w "\n%{http_code}" "$BACKEND_URL/stream/invalid-track-id?source=jamendo" 2>/dev/null || echo "000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "500" ] || [ "$http_code" = "400" ]; then
        print_status "SUCCESS" "Invalid track ID handled gracefully (Status: $http_code)"
        return 0
    else
        print_status "WARNING" "Unexpected response for invalid track ID (Status: $http_code)"
        print_status "INFO" "Response: $body"
        return 1
    fi
}

# Test 6: Invalid Source Parameter
test_invalid_source() {
    echo -e "\n${BLUE}=== TEST 6: Invalid Source Parameter ===${NC}"
    
    print_status "INFO" "Testing with invalid source parameter"
    
    response=$(curl -s -w "\n%{http_code}" "$BACKEND_URL/stream/$TEST_TRACK_ID?source=invalid" 2>/dev/null || echo "000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "500" ] || [ "$http_code" = "400" ]; then
        print_status "SUCCESS" "Invalid source parameter handled gracefully (Status: $http_code)"
        return 0
    else
        print_status "WARNING" "Unexpected response for invalid source (Status: $http_code)"
        print_status "INFO" "Response: $body"
        return 1
    fi
}

# Test 7: Jamendo API Key Validation
test_jamendo_api_key() {
    echo -e "\n${BLUE}=== TEST 7: Jamendo API Key Validation ===${NC}"
    
    print_status "INFO" "Testing Jamendo API key validity..."
    
    # Test direct Jamendo API call
    jamendo_response=$(curl -s "https://api.jamendo.com/v3.0/tracks/?client_id=$JAMENDO_API_KEY&format=json&limit=1" 2>/dev/null || echo "{}")
    
    if echo "$jamendo_response" | jq -e '.results' > /dev/null 2>&1; then
        print_status "SUCCESS" "Jamendo API key is valid and working"
        return 0
    else
        print_status "ERROR" "Jamendo API key may be invalid or API is down"
        print_status "ERROR" "Response: $jamendo_response"
        return 1
    fi
}

# Test 8: Frontend API Service Test
test_frontend_api_service() {
    echo -e "\n${BLUE}=== TEST 8: Frontend API Service Test ===${NC}"
    
    print_status "INFO" "Testing if frontend can build successfully..."
    
    cd ../client
    if npm run build > /dev/null 2>&1; then
        print_status "SUCCESS" "Frontend builds successfully"
        cd ..
        return 0
    else
        print_status "ERROR" "Frontend build failed"
        cd ..
        return 1
    fi
}

# Main test execution
main() {
    echo -e "${BLUE}🎵 GRUVIT STREAMING FUNCTIONALITY TEST${NC}"
    echo -e "${BLUE}=====================================${NC}"
    echo -e "Backend URL: $BACKEND_URL"
    echo -e "Jamendo API Key: $JAMENDO_API_KEY"
    echo -e "Test Track ID: $TEST_TRACK_ID"
    echo ""
    
    # Run all tests
    test_backend_health
    test_search_tracks
    test_public_streaming
    test_authenticated_streaming_no_auth
    test_invalid_track_id
    test_invalid_source
    test_jamendo_api_key
    test_frontend_api_service
    
    # Summary
    echo -e "\n${BLUE}=== TEST SUMMARY ===${NC}"
    echo -e "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}🎉 ALL TESTS PASSED! Streaming functionality should work properly.${NC}"
        exit 0
    else
        echo -e "\n${RED}❌ SOME TESTS FAILED! Check the issues above.${NC}"
        exit 1
    fi
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is not installed. Please install jq to run this test script.${NC}"
    echo -e "${YELLOW}On Ubuntu/Debian: sudo apt-get install jq${NC}"
    echo -e "${YELLOW}On macOS: brew install jq${NC}"
    echo -e "${YELLOW}On Windows: Download from https://stedolan.github.io/jq/${NC}"
    exit 1
fi

# Run the tests
main "$@"
