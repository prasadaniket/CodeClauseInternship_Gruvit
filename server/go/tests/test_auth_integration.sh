#!/bin/bash

# Test script for authentication integration
# This script tests the complete auth flow between Go and Java services

echo "🧪 Testing Authentication Integration"
echo "====================================="

# Configuration
GO_SERVICE_URL="http://localhost:3001"
JAVA_SERVICE_URL="http://localhost:8081"
TEST_USERNAME="testuser_$(date +%s)"
TEST_EMAIL="test_$(date +%s)@example.com"
TEST_PASSWORD="password123"

echo "📋 Test Configuration:"
echo "   Go Service: $GO_SERVICE_URL"
echo "   Java Service: $JAVA_SERVICE_URL"
echo "   Test User: $TEST_USERNAME"
echo ""

# Function to check if service is running
check_service() {
    local url=$1
    local name=$2
    
    echo "🔍 Checking $name..."
    if curl -s "$url/health" > /dev/null 2>&1; then
        echo "   ✅ $name is running"
        return 0
    else
        echo "   ❌ $name is not running"
        return 1
    fi
}

# Function to make API call and check response
api_call() {
    local method=$1
    local url=$2
    local data=$3
    local headers=$4
    local expected_status=$5
    
    echo "📡 $method $url"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -H "$headers" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "$headers")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        echo "   ✅ Status: $http_code (expected: $expected_status)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        echo "   ❌ Status: $http_code (expected: $expected_status)"
        echo "$body"
        return 1
    fi
}

# Check if services are running
echo "🔍 Checking Services..."
if ! check_service "$GO_SERVICE_URL" "Go Music Service"; then
    echo "❌ Go service is not running. Please start it with: go run main.go"
    exit 1
fi

if ! check_service "$JAVA_SERVICE_URL" "Java Auth Service"; then
    echo "❌ Java service is not running. Please start it with: ./mvnw spring-boot:run"
    exit 1
fi

echo ""

# Test 1: User Registration
echo "🧪 Test 1: User Registration"
echo "----------------------------"
registration_data='{
    "username": "'$TEST_USERNAME'",
    "email": "'$TEST_EMAIL'",
    "password": "'$TEST_PASSWORD'",
    "firstName": "Test",
    "lastName": "User"
}'

if api_call "POST" "$GO_SERVICE_URL/auth/register" "$registration_data" "" "201"; then
    echo "✅ Registration successful"
    REGISTRATION_SUCCESS=true
else
    echo "❌ Registration failed"
    REGISTRATION_SUCCESS=false
fi

echo ""

# Test 2: User Login
echo "🧪 Test 2: User Login"
echo "---------------------"
login_data='{
    "username": "'$TEST_USERNAME'",
    "password": "'$TEST_PASSWORD'"
}'

if api_call "POST" "$GO_SERVICE_URL/auth/login" "$login_data" "" "200"; then
    echo "✅ Login successful"
    LOGIN_SUCCESS=true
    
    # Extract access token
    ACCESS_TOKEN=$(echo "$body" | jq -r '.access_token' 2>/dev/null)
    if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
        echo "   🔑 Access token extracted"
    else
        echo "   ❌ Failed to extract access token"
        LOGIN_SUCCESS=false
    fi
else
    echo "❌ Login failed"
    LOGIN_SUCCESS=false
fi

echo ""

# Test 3: Token Validation
echo "🧪 Test 3: Token Validation"
echo "---------------------------"
if [ "$LOGIN_SUCCESS" = true ] && [ -n "$ACCESS_TOKEN" ]; then
    if api_call "POST" "$GO_SERVICE_URL/auth/validate" "" "Authorization: Bearer $ACCESS_TOKEN" "200"; then
        echo "✅ Token validation successful"
        VALIDATION_SUCCESS=true
    else
        echo "❌ Token validation failed"
        VALIDATION_SUCCESS=false
    fi
else
    echo "⏭️  Skipping token validation (login failed)"
    VALIDATION_SUCCESS=false
fi

echo ""

# Test 4: Protected Endpoint Access
echo "🧪 Test 4: Protected Endpoint Access"
echo "------------------------------------"
if [ "$LOGIN_SUCCESS" = true ] && [ -n "$ACCESS_TOKEN" ]; then
    if api_call "GET" "$GO_SERVICE_URL/api/profile" "" "Authorization: Bearer $ACCESS_TOKEN" "200"; then
        echo "✅ Protected endpoint access successful"
        PROTECTED_ACCESS_SUCCESS=true
    else
        echo "❌ Protected endpoint access failed"
        PROTECTED_ACCESS_SUCCESS=false
    fi
else
    echo "⏭️  Skipping protected endpoint test (login failed)"
    PROTECTED_ACCESS_SUCCESS=false
fi

echo ""

# Test 5: Token Refresh
echo "🧪 Test 5: Token Refresh"
echo "------------------------"
if [ "$LOGIN_SUCCESS" = true ]; then
    REFRESH_TOKEN=$(echo "$body" | jq -r '.refresh_token' 2>/dev/null)
    if [ "$REFRESH_TOKEN" != "null" ] && [ -n "$REFRESH_TOKEN" ]; then
        refresh_data='{
            "refresh_token": "'$REFRESH_TOKEN'"
        }'
        
        if api_call "POST" "$GO_SERVICE_URL/auth/refresh" "$refresh_data" "" "200"; then
            echo "✅ Token refresh successful"
            REFRESH_SUCCESS=true
        else
            echo "❌ Token refresh failed"
            REFRESH_SUCCESS=false
        fi
    else
        echo "❌ No refresh token available"
        REFRESH_SUCCESS=false
    fi
else
    echo "⏭️  Skipping token refresh test (login failed)"
    REFRESH_SUCCESS=false
fi

echo ""

# Test 6: Music Search (Public Endpoint)
echo "🧪 Test 6: Music Search (Public Endpoint)"
echo "------------------------------------------"
if api_call "GET" "$GO_SERVICE_URL/search?q=test" "" "" "200"; then
    echo "✅ Music search successful"
    SEARCH_SUCCESS=true
else
    echo "❌ Music search failed"
    SEARCH_SUCCESS=false
fi

echo ""

# Summary
echo "📊 Test Summary"
echo "==============="
echo "Registration:     $([ "$REGISTRATION_SUCCESS" = true ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "Login:            $([ "$LOGIN_SUCCESS" = true ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "Token Validation: $([ "$VALIDATION_SUCCESS" = true ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "Protected Access: $([ "$PROTECTED_ACCESS_SUCCESS" = true ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "Token Refresh:    $([ "$REFRESH_SUCCESS" = true ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "Music Search:     $([ "$SEARCH_SUCCESS" = true ] && echo "✅ PASS" || echo "❌ FAIL")"

echo ""

# Overall result
PASSED_TESTS=0
TOTAL_TESTS=6

[ "$REGISTRATION_SUCCESS" = true ] && ((PASSED_TESTS++))
[ "$LOGIN_SUCCESS" = true ] && ((PASSED_TESTS++))
[ "$VALIDATION_SUCCESS" = true ] && ((PASSED_TESTS++))
[ "$PROTECTED_ACCESS_SUCCESS" = true ] && ((PASSED_TESTS++))
[ "$REFRESH_SUCCESS" = true ] && ((PASSED_TESTS++))
[ "$SEARCH_SUCCESS" = true ] && ((PASSED_TESTS++))

echo "🎯 Overall Result: $PASSED_TESTS/$TOTAL_TESTS tests passed"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo "🎉 All tests passed! Authentication integration is working correctly."
    exit 0
else
    echo "⚠️  Some tests failed. Please check the logs and configuration."
    exit 1
fi
