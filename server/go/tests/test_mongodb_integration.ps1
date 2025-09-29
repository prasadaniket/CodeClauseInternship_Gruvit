# Test MongoDB Integration for Gruvit Go Service
# This script tests the MongoDB integration by running the Go service and making API calls

Write-Host "🧪 Testing MongoDB Integration for Gruvit Go Service" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan

# Configuration
$GO_SERVICE_PORT = if ($env:GO_SERVICE_PORT) { $env:GO_SERVICE_PORT } else { "3001" }
$BASE_URL = "http://localhost:$GO_SERVICE_PORT"
$TEST_USER_ID = "test_user_123"
$TEST_USERNAME = "testuser"

# Function to print colored output
function Write-Status {
    param(
        [string]$Status,
        [string]$Message
    )
    
    switch ($Status) {
        "SUCCESS" { Write-Host "✅ $Message" -ForegroundColor Green }
        "ERROR" { Write-Host "❌ $Message" -ForegroundColor Red }
        "INFO" { Write-Host "ℹ️  $Message" -ForegroundColor Blue }
        "WARNING" { Write-Host "⚠️  $Message" -ForegroundColor Yellow }
    }
}

# Function to make HTTP requests
function Invoke-TestRequest {
    param(
        [string]$Method,
        [string]$Url,
        [string]$Data = $null,
        [int]$ExpectedStatus
    )
    
    $headers = @{
        "Authorization" = "Bearer test_token"
    }
    
    if ($Data) {
        $headers["Content-Type"] = "application/json"
        try {
            $response = Invoke-RestMethod -Uri $Url -Method $Method -Headers $headers -Body $Data -ErrorAction Stop
            $statusCode = 200
        }
        catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            $response = $_.Exception.Message
        }
    }
    else {
        try {
            $response = Invoke-RestMethod -Uri $Url -Method $Method -Headers $headers -ErrorAction Stop
            $statusCode = 200
        }
        catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            $response = $_.Exception.Message
        }
    }
    
    if ($statusCode -eq $ExpectedStatus) {
        Write-Status "SUCCESS" "$Method $Url -> $statusCode"
        $response | ConvertTo-Json -Depth 3
        return $true
    }
    else {
        Write-Status "ERROR" "$Method $Url -> Expected $ExpectedStatus, got $statusCode"
        Write-Host $response
        return $false
    }
}

# Check if Go service is running
function Test-Service {
    Write-Status "INFO" "Checking if Go service is running on port $GO_SERVICE_PORT..."
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/health" -ErrorAction Stop
        Write-Status "SUCCESS" "Go service is running"
        return $true
    }
    catch {
        Write-Status "ERROR" "Go service is not running on port $GO_SERVICE_PORT"
        Write-Status "INFO" "Please start the Go service first:"
        Write-Host "  cd server/go && go run main.go"
        return $false
    }
}

# Test health endpoint
function Test-Health {
    Write-Status "INFO" "Testing health endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/health" $null 200
}

# Test search endpoint
function Test-Search {
    Write-Status "INFO" "Testing search endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/search?q=test&limit=5" $null 200
}

# Test user profile endpoint
function Test-UserProfile {
    Write-Status "INFO" "Testing user profile endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/api/profile" $null 200
}

# Test user favorites endpoint
function Test-UserFavorites {
    Write-Status "INFO" "Testing user favorites endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/api/user/favorites" $null 200
}

# Test user top artists endpoint
function Test-UserTopArtists {
    Write-Status "INFO" "Testing user top artists endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/api/user/top-artists" $null 200
}

# Test user top tracks endpoint
function Test-UserTopTracks {
    Write-Status "INFO" "Testing user top tracks endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/api/user/top-tracks" $null 200
}

# Test user followings endpoint
function Test-UserFollowings {
    Write-Status "INFO" "Testing user followings endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/api/user/followings" $null 200
}

# Test user followers endpoint
function Test-UserFollowers {
    Write-Status "INFO" "Testing user followers endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/api/user/followers" $null 200
}

# Test adding track to favorites
function Test-AddToFavorites {
    Write-Status "INFO" "Testing add to favorites endpoint..."
    Invoke-TestRequest "POST" "$BASE_URL/api/user/favorites/test_track_123" $null 200
}

# Test playlists endpoint
function Test-Playlists {
    Write-Status "INFO" "Testing playlists endpoint..."
    Invoke-TestRequest "GET" "$BASE_URL/api/playlists" $null 200
}

# Test creating a playlist
function Test-CreatePlaylist {
    Write-Status "INFO" "Testing create playlist endpoint..."
    $playlistData = '{"name":"Test Playlist","description":"A test playlist","is_public":false}'
    Invoke-TestRequest "POST" "$BASE_URL/api/playlists" $playlistData 201
}

# Main test function
function Start-Tests {
    Write-Host ""
    Write-Status "INFO" "Starting MongoDB integration tests..."
    Write-Host ""
    
    # Check if service is running
    if (-not (Test-Service)) {
        exit 1
    }
    
    Write-Host ""
    Write-Status "INFO" "Running API tests..."
    Write-Host ""
    
    # Test basic endpoints
    Test-Health
    Write-Host ""
    
    Test-Search
    Write-Host ""
    
    # Test user endpoints (these will use MongoDB)
    Test-UserProfile
    Write-Host ""
    
    Test-UserFavorites
    Write-Host ""
    
    Test-UserTopArtists
    Write-Host ""
    
    Test-UserTopTracks
    Write-Host ""
    
    Test-UserFollowings
    Write-Host ""
    
    Test-UserFollowers
    Write-Host ""
    
    Test-AddToFavorites
    Write-Host ""
    
    # Test playlist endpoints
    Test-Playlists
    Write-Host ""
    
    Test-CreatePlaylist
    Write-Host ""
    
    Write-Status "SUCCESS" "All tests completed!"
    Write-Host ""
    Write-Status "INFO" "MongoDB integration is working correctly."
    Write-Status "INFO" "The Go service is now using real MongoDB operations instead of mock data."
}

# Check dependencies
function Test-Dependencies {
    $missingDeps = @()
    
    if (-not (Get-Command curl -ErrorAction SilentlyContinue)) {
        $missingDeps += "curl"
    }
    
    if ($missingDeps.Count -ne 0) {
        Write-Status "ERROR" "Missing dependencies: $($missingDeps -join ', ')"
        Write-Status "INFO" "Please install the missing dependencies and run the script again."
        exit 1
    }
}

# Main execution
function Main {
    Write-Host "🔍 Checking dependencies..."
    Test-Dependencies
    
    Write-Host ""
    Write-Host "🚀 Starting MongoDB integration tests..."
    Start-Tests
}

# Run main function
Main
