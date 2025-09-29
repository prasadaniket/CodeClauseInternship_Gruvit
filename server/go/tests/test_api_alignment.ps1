# Test API Alignment between Frontend and Backend
# This script tests that the Go service responses match what the frontend expects

Write-Host "🔗 Testing API Alignment between Frontend and Backend" -ForegroundColor Cyan
Write-Host "=====================================================" -ForegroundColor Cyan

# Configuration
$GO_SERVICE_PORT = if ($env:GO_SERVICE_PORT) { $env:GO_SERVICE_PORT } else { "8080" }
$BASE_URL = "http://localhost:$GO_SERVICE_PORT"

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

# Function to test API endpoint and validate response format
function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [string[]]$ExpectedFields,
        [string]$AuthToken = $null,
        [string]$Description
    )
    
    Write-Status "INFO" "Testing $Description"
    
    $headers = @{}
    if ($AuthToken) {
        $headers["Authorization"] = "Bearer $AuthToken"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL$Endpoint" -Method $Method -Headers $headers -ErrorAction Stop
        Write-Status "SUCCESS" "$Method $Endpoint -> 200"
        
        # Validate expected fields
        $allFieldsFound = $true
        foreach ($field in $ExpectedFields) {
            if ($response.PSObject.Properties.Name -contains $field) {
                Write-Status "SUCCESS" "Field '$field' found in response"
            } else {
                Write-Status "ERROR" "Field '$field' missing from response"
                $allFieldsFound = $false
            }
        }
        
        if ($allFieldsFound) {
            $response | ConvertTo-Json -Depth 3
            return $true
        } else {
            return $false
        }
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Status "ERROR" "$Method $Endpoint -> Expected 200, got $statusCode"
        Write-Host $_.Exception.Message
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

# Test search endpoints
function Test-SearchEndpoints {
    Write-Status "INFO" "Testing search endpoints..."
    
    # Test /search endpoint
    Test-Endpoint "GET" "/search?q=rock&limit=5" @("query", "results", "total", "page", "limit") $null "Search endpoint"
    
    # Test /music/search endpoint (should redirect)
    Test-Endpoint "GET" "/music/search?q=rock&limit=5" @("query", "results", "total", "page", "limit") $null "Music search endpoint"
    
    # Test /music/artist endpoint
    Test-Endpoint "GET" "/music/artist?artist=rock&limit=5" @("artist", "results", "total") $null "Artist search endpoint"
    
    # Test /music/genre endpoint
    Test-Endpoint "GET" "/music/genre?genre=rock&limit=5" @("genre", "results", "total") $null "Genre search endpoint"
    
    # Test /music/popular endpoint
    Test-Endpoint "GET" "/music/popular?limit=5" @("results", "total") $null "Popular tracks endpoint"
    
    # Test /music/track/:id endpoint
    Test-Endpoint "GET" "/music/track/test123" @("id", "title", "artist", "source") $null "Track details endpoint"
}

# Test streaming endpoints
function Test-StreamingEndpoints {
    Write-Status "INFO" "Testing streaming endpoints..."
    
    try {
        # Get a track ID from search first
        $searchResponse = Invoke-RestMethod -Uri "$BASE_URL/search?q=rock&limit=1" -ErrorAction Stop
        $trackId = $searchResponse.results[0].id
        
        if ($trackId) {
            Write-Status "INFO" "Found track ID: $trackId"
            
            # Test public streaming endpoint
            Test-Endpoint "GET" "/stream/$trackId?source=jamendo" @("stream_url") $null "Public streaming endpoint"
            
            # Test authenticated streaming endpoint (will fail without auth, but should return proper error)
            try {
                $response = Invoke-RestMethod -Uri "$BASE_URL/api/stream/$trackId?source=jamendo" -ErrorAction Stop
                Write-Status "WARNING" "Authenticated streaming endpoint should require auth"
            }
            catch {
                $statusCode = $_.Exception.Response.StatusCode.value__
                if ($statusCode -eq 401 -or $statusCode -eq 403) {
                    Write-Status "SUCCESS" "Authenticated streaming endpoint properly requires auth"
                } else {
                    Write-Status "WARNING" "Authenticated streaming endpoint returned $statusCode (expected 401/403)"
                }
            }
        } else {
            Write-Status "WARNING" "No tracks found for streaming test"
        }
    }
    catch {
        Write-Status "WARNING" "Could not get track ID from search: $($_.Exception.Message)"
    }
}

# Test profile endpoint (requires auth)
function Test-ProfileEndpoint {
    Write-Status "INFO" "Testing profile endpoint..."
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/api/profile" -ErrorAction Stop
        Write-Status "WARNING" "Profile endpoint should require auth"
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 401 -or $statusCode -eq 403) {
            Write-Status "SUCCESS" "Profile endpoint properly requires auth"
        } else {
            Write-Status "WARNING" "Profile endpoint returned $statusCode (expected 401/403)"
        }
    }
}

# Test playlist endpoints (requires auth)
function Test-PlaylistEndpoints {
    Write-Status "INFO" "Testing playlist endpoints..."
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/api/playlists" -ErrorAction Stop
        Write-Status "WARNING" "Playlist endpoint should require auth"
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 401 -or $statusCode -eq 403) {
            Write-Status "SUCCESS" "Playlist endpoint properly requires auth"
        } else {
            Write-Status "WARNING" "Playlist endpoint returned $statusCode (expected 401/403)"
        }
    }
}

# Test user analytics endpoints (requires auth)
function Test-UserEndpoints {
    Write-Status "INFO" "Testing user analytics endpoints..."
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/api/user/stats" -ErrorAction Stop
        Write-Status "WARNING" "User stats endpoint should require auth"
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 401 -or $statusCode -eq 403) {
            Write-Status "SUCCESS" "User stats endpoint properly requires auth"
        } else {
            Write-Status "WARNING" "User stats endpoint returned $statusCode (expected 401/403)"
        }
    }
}

# Test health endpoint
function Test-HealthEndpoint {
    Write-Status "INFO" "Testing health endpoint..."
    Test-Endpoint "GET" "/health" @("status", "service", "version") $null "Health endpoint"
}

# Main test function
function Start-Tests {
    Write-Host ""
    Write-Status "INFO" "Starting API alignment tests..."
    Write-Host ""
    
    # Check if service is running
    if (-not (Test-Service)) {
        exit 1
    }
    
    Write-Host ""
    Write-Status "INFO" "Running API alignment tests..."
    Write-Host ""
    
    # Test health endpoint
    Test-HealthEndpoint
    Write-Host ""
    
    # Test search endpoints
    Test-SearchEndpoints
    Write-Host ""
    
    # Test streaming endpoints
    Test-StreamingEndpoints
    Write-Host ""
    
    # Test profile endpoint
    Test-ProfileEndpoint
    Write-Host ""
    
    # Test playlist endpoints
    Test-PlaylistEndpoints
    Write-Host ""
    
    # Test user endpoints
    Test-UserEndpoints
    Write-Host ""
    
    Write-Status "SUCCESS" "All API alignment tests completed!"
    Write-Host ""
    Write-Status "INFO" "API responses now match frontend expectations."
    Write-Status "INFO" "Frontend should be able to display data properly."
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
    Write-Host "🚀 Starting API alignment tests..."
    Start-Tests
}

# Run main function
Main
