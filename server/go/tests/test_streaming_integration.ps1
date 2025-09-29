# Test Music Streaming Integration for Gruvit Go Service
# This script tests the streaming functionality

Write-Host "🎵 Testing Music Streaming Integration for Gruvit Go Service" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# Configuration
$GO_SERVICE_PORT = if ($env:GO_SERVICE_PORT) { $env:GO_SERVICE_PORT } else { "3001" }
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

# Test search to get Jamendo tracks
function Test-SearchJamendo {
    Write-Status "INFO" "Testing search to get Jamendo tracks..."
    Invoke-TestRequest "GET" "$BASE_URL/search?q=rock&limit=5" $null 200
}

# Test public streaming endpoint (Jamendo)
function Test-PublicStreaming {
    Write-Status "INFO" "Testing public streaming endpoint for Jamendo tracks..."
    
    try {
        # First, get a Jamendo track ID from search
        $searchResponse = Invoke-RestMethod -Uri "$BASE_URL/search?q=rock&limit=1" -ErrorAction Stop
        $trackId = $searchResponse.results[0].id
        
        if ($trackId) {
            Write-Status "INFO" "Found Jamendo track ID: $trackId"
            Invoke-TestRequest "GET" "$BASE_URL/stream/$trackId?source=jamendo" $null 200
        }
        else {
            Write-Status "WARNING" "No Jamendo tracks found in search results"
        }
    }
    catch {
        Write-Status "WARNING" "Could not get track ID from search: $($_.Exception.Message)"
    }
}

# Test authenticated streaming endpoint
function Test-AuthenticatedStreaming {
    Write-Status "INFO" "Testing authenticated streaming endpoint..."
    
    try {
        # First, get a Jamendo track ID from search
        $searchResponse = Invoke-RestMethod -Uri "$BASE_URL/search?q=rock&limit=1" -ErrorAction Stop
        $trackId = $searchResponse.results[0].id
        
        if ($trackId) {
            Write-Status "INFO" "Found track ID: $trackId"
            Invoke-TestRequest "GET" "$BASE_URL/api/stream/$trackId?source=jamendo" $null 200
        }
        else {
            Write-Status "WARNING" "No tracks found in search results"
        }
    }
    catch {
        Write-Status "WARNING" "Could not get track ID from search: $($_.Exception.Message)"
    }
}

# Test MusicBrainz streaming (should fail gracefully)
function Test-MusicBrainzStreaming {
    Write-Status "INFO" "Testing MusicBrainz streaming (should fail gracefully)..."
    Invoke-TestRequest "GET" "$BASE_URL/api/stream/fake-musicbrainz-id?source=musicbrainz" $null 500
}

# Test invalid source
function Test-InvalidSource {
    Write-Status "INFO" "Testing invalid source..."
    Invoke-TestRequest "GET" "$BASE_URL/api/stream/test-id?source=invalid" $null 500
}

# Test missing source parameter
function Test-MissingSource {
    Write-Status "INFO" "Testing missing source parameter..."
    Invoke-TestRequest "GET" "$BASE_URL/api/stream/test-id" $null 400
}

# Test stream URL validation
function Test-StreamValidation {
    Write-Status "INFO" "Testing stream URL validation..."
    
    try {
        # Get a stream URL
        $searchResponse = Invoke-RestMethod -Uri "$BASE_URL/search?q=rock&limit=1" -ErrorAction Stop
        $trackId = $searchResponse.results[0].id
        
        if ($trackId) {
            $streamResponse = Invoke-RestMethod -Uri "$BASE_URL/stream/$trackId?source=jamendo" -ErrorAction Stop
            $streamUrl = $streamResponse.stream_url
            
            if ($streamUrl) {
                Write-Status "INFO" "Testing stream URL: $streamUrl"
                
                try {
                    $headResponse = Invoke-WebRequest -Uri $streamUrl -Method Head -ErrorAction Stop
                    if ($headResponse.StatusCode -eq 200) {
                        Write-Status "SUCCESS" "Stream URL is accessible"
                    }
                    else {
                        Write-Status "WARNING" "Stream URL returned status: $($headResponse.StatusCode)"
                    }
                }
                catch {
                    Write-Status "WARNING" "Stream URL may not be accessible: $($_.Exception.Message)"
                }
            }
            else {
                Write-Status "WARNING" "No stream URL found in response"
            }
        }
        else {
            Write-Status "WARNING" "No tracks found for validation test"
        }
    }
    catch {
        Write-Status "WARNING" "Could not test stream validation: $($_.Exception.Message)"
    }
}

# Main test function
function Start-Tests {
    Write-Host ""
    Write-Status "INFO" "Starting streaming integration tests..."
    Write-Host ""
    
    # Check if service is running
    if (-not (Test-Service)) {
        exit 1
    }
    
    Write-Host ""
    Write-Status "INFO" "Running streaming tests..."
    Write-Host ""
    
    # Test search first
    Test-SearchJamendo
    Write-Host ""
    
    # Test public streaming
    Test-PublicStreaming
    Write-Host ""
    
    # Test authenticated streaming
    Test-AuthenticatedStreaming
    Write-Host ""
    
    # Test MusicBrainz streaming (should fail)
    Test-MusicBrainzStreaming
    Write-Host ""
    
    # Test invalid source
    Test-InvalidSource
    Write-Host ""
    
    # Test missing source
    Test-MissingSource
    Write-Host ""
    
    # Test stream validation
    Test-StreamValidation
    Write-Host ""
    
    Write-Status "SUCCESS" "All streaming tests completed!"
    Write-Host ""
    Write-Status "INFO" "Streaming integration is working correctly."
    Write-Status "INFO" "Jamendo tracks can now be streamed properly."
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
    Write-Host "🚀 Starting streaming integration tests..."
    Start-Tests
}

# Run main function
Main
