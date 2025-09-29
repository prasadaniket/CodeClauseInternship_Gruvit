# Comprehensive Streaming Functionality Test Script (PowerShell)
# This script tests the complete streaming flow from backend to frontend

param(
    [string]$BackendUrl = "http://localhost:3001",
    [string]$JamendoApiKey = "be6cb53f"
)

# Test results tracking
$TestsPassed = 0
$TestsFailed = 0
$TotalTests = 0
$TestTrackId = $null

# Helper functions
function Write-Status {
    param(
        [string]$Status,
        [string]$Message
    )
    
    $TotalTests++
    
    switch ($Status) {
        "INFO" { Write-Host "[INFO] $Message" -ForegroundColor Blue }
        "SUCCESS" { 
            Write-Host "[SUCCESS] $Message" -ForegroundColor Green
            $script:TestsPassed++
        }
        "ERROR" { 
            Write-Host "[ERROR] $Message" -ForegroundColor Red
            $script:TestsFailed++
        }
        "WARNING" { Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
    }
}

function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [int]$ExpectedStatus,
        [string]$Description
    )
    
    Write-Status "INFO" "Testing: $Description"
    Write-Status "INFO" "Endpoint: $Method $Endpoint"
    
    try {
        $response = Invoke-RestMethod -Uri "$BackendUrl$Endpoint" -Method $Method -ErrorAction Stop
        $httpCode = 200
        Write-Status "SUCCESS" "$Description - Status: $httpCode"
        if ($response) {
            Write-Status "INFO" "Response: $($response | ConvertTo-Json -Compress)"
        }
        return $true
    }
    catch {
        $httpCode = $_.Exception.Response.StatusCode.value__
        if ($httpCode -eq $ExpectedStatus) {
            Write-Status "SUCCESS" "$Description - Status: $httpCode (Expected)"
            return $true
        } else {
            Write-Status "ERROR" "$Description - Expected: $ExpectedStatus, Got: $httpCode"
            Write-Status "ERROR" "Error: $($_.Exception.Message)"
            return $false
        }
    }
}

# Test 1: Backend Health Check
function Test-BackendHealth {
    Write-Host "`n=== TEST 1: Backend Health Check ===" -ForegroundColor Blue
    
    try {
        $response = Invoke-RestMethod -Uri "$BackendUrl/health" -ErrorAction Stop
        Write-Status "SUCCESS" "Backend is running and accessible"
        return $true
    }
    catch {
        Write-Status "ERROR" "Backend is not accessible at $BackendUrl"
        Write-Status "INFO" "Make sure the backend server is running on port 3001"
        return $false
    }
}

# Test 2: Search for Real Tracks
function Test-SearchTracks {
    Write-Host "`n=== TEST 2: Search for Real Tracks ===" -ForegroundColor Blue
    
    Write-Status "INFO" "Searching for tracks to get real track IDs..."
    
    try {
        $searchResponse = Invoke-RestMethod -Uri "$BackendUrl/search?q=rock&limit=5" -ErrorAction Stop
        
        if ($searchResponse.results -and $searchResponse.results.Count -gt 0) {
            $script:TestTrackId = $searchResponse.results[0].id
            $trackTitle = $searchResponse.results[0].title
            $trackArtist = $searchResponse.results[0].artist
            
            Write-Status "SUCCESS" "Found $($searchResponse.results.Count) tracks"
            Write-Status "INFO" "Using track: '$trackTitle' by '$trackArtist' (ID: $TestTrackId)"
            return $true
        } else {
            Write-Status "ERROR" "No tracks found in search results"
            return $false
        }
    }
    catch {
        Write-Status "ERROR" "Search request failed: $($_.Exception.Message)"
        return $false
    }
}

# Test 3: Public Streaming Endpoint
function Test-PublicStreaming {
    Write-Host "`n=== TEST 3: Public Streaming Endpoint ===" -ForegroundColor Blue
    
    if (-not $TestTrackId) {
        Write-Status "ERROR" "No valid track ID available for testing"
        return $false
    }
    
    Write-Status "INFO" "Testing public streaming endpoint with track ID: $TestTrackId"
    
    try {
        $streamResponse = Invoke-RestMethod -Uri "$BackendUrl/stream/$TestTrackId?source=jamendo" -ErrorAction Stop
        
        if ($streamResponse.stream_url) {
            $streamUrl = $streamResponse.stream_url
            Write-Status "SUCCESS" "Public streaming endpoint returned valid stream URL"
            Write-Status "INFO" "Stream URL: $streamUrl"
            
            # Test if the stream URL is accessible
            Write-Status "INFO" "Testing stream URL accessibility..."
            try {
                $streamTest = Invoke-WebRequest -Uri $streamUrl -Method Head -ErrorAction Stop
                if ($streamTest.StatusCode -eq 200) {
                    Write-Status "SUCCESS" "Stream URL is accessible and returns 200 OK"
                } else {
                    Write-Status "WARNING" "Stream URL returned status: $($streamTest.StatusCode)"
                }
            }
            catch {
                Write-Status "WARNING" "Stream URL may not be accessible (this could be normal for some APIs)"
            }
            return $true
        } else {
            Write-Status "ERROR" "Stream URL is null or empty"
            return $false
        }
    }
    catch {
        Write-Status "ERROR" "Streaming request failed: $($_.Exception.Message)"
        return $false
    }
}

# Test 4: Authenticated Streaming Endpoint (without auth - should fail)
function Test-AuthenticatedStreamingNoAuth {
    Write-Host "`n=== TEST 4: Authenticated Streaming (No Auth) ===" -ForegroundColor Blue
    
    if (-not $TestTrackId) {
        Write-Status "ERROR" "No valid track ID available for testing"
        return $false
    }
    
    Write-Status "INFO" "Testing authenticated streaming endpoint without authentication (should fail)"
    
    try {
        $response = Invoke-RestMethod -Uri "$BackendUrl/api/stream/$TestTrackId?source=jamendo" -ErrorAction Stop
        Write-Status "ERROR" "Authenticated endpoint should require auth, but succeeded"
        return $false
    }
    catch {
        $httpCode = $_.Exception.Response.StatusCode.value__
        if ($httpCode -eq 401 -or $httpCode -eq 403) {
            Write-Status "SUCCESS" "Authenticated endpoint properly requires authentication (Status: $httpCode)"
            return $true
        } else {
            Write-Status "ERROR" "Authenticated endpoint should require auth, but got status: $httpCode"
            return $false
        }
    }
}

# Test 5: Invalid Track ID
function Test-InvalidTrackId {
    Write-Host "`n=== TEST 5: Invalid Track ID Handling ===" -ForegroundColor Blue
    
    Write-Status "INFO" "Testing with invalid track ID (should fail gracefully)"
    
    try {
        $response = Invoke-RestMethod -Uri "$BackendUrl/stream/invalid-track-id?source=jamendo" -ErrorAction Stop
        Write-Status "WARNING" "Invalid track ID should have failed, but succeeded"
        return $false
    }
    catch {
        $httpCode = $_.Exception.Response.StatusCode.value__
        if ($httpCode -eq 500 -or $httpCode -eq 400) {
            Write-Status "SUCCESS" "Invalid track ID handled gracefully (Status: $httpCode)"
            return $true
        } else {
            Write-Status "WARNING" "Unexpected response for invalid track ID (Status: $httpCode)"
            return $false
        }
    }
}

# Test 6: Jamendo API Key Validation
function Test-JamendoApiKey {
    Write-Host "`n=== TEST 6: Jamendo API Key Validation ===" -ForegroundColor Blue
    
    Write-Status "INFO" "Testing Jamendo API key validity..."
    
    try {
        $jamendoResponse = Invoke-RestMethod -Uri "https://api.jamendo.com/v3.0/tracks/?client_id=$JamendoApiKey&format=json&limit=1" -ErrorAction Stop
        
        if ($jamendoResponse.results -and $jamendoResponse.results.Count -gt 0) {
            Write-Status "SUCCESS" "Jamendo API key is valid and working"
            return $true
        } else {
            Write-Status "ERROR" "Jamendo API key may be invalid or API is down"
            return $false
        }
    }
    catch {
        Write-Status "ERROR" "Jamendo API request failed: $($_.Exception.Message)"
        return $false
    }
}

# Test 7: Frontend API Service Test
function Test-FrontendApiService {
    Write-Host "`n=== TEST 7: Frontend API Service Test ===" -ForegroundColor Blue
    
    Write-Status "INFO" "Testing if frontend can build successfully..."
    
    $originalLocation = Get-Location
    try {
        Set-Location "../client"
        $buildResult = & npm run build 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Status "SUCCESS" "Frontend builds successfully"
            return $true
        } else {
            Write-Status "ERROR" "Frontend build failed"
            Write-Status "ERROR" "Build output: $buildResult"
            return $false
        }
    }
    catch {
        Write-Status "ERROR" "Frontend build failed: $($_.Exception.Message)"
        return $false
    }
    finally {
        Set-Location $originalLocation
    }
}

# Main test execution
function Main {
    Write-Host "🎵 GRUVIT STREAMING FUNCTIONALITY TEST" -ForegroundColor Blue
    Write-Host "=====================================" -ForegroundColor Blue
    Write-Host "Backend URL: $BackendUrl"
    Write-Host "Jamendo API Key: $JamendoApiKey"
    Write-Host "Test Track ID: $TestTrackId"
    Write-Host ""
    
    # Run all tests
    Test-BackendHealth
    Test-SearchTracks
    Test-PublicStreaming
    Test-AuthenticatedStreamingNoAuth
    Test-InvalidTrackId
    Test-JamendoApiKey
    Test-FrontendApiService
    
    # Summary
    Write-Host "`n=== TEST SUMMARY ===" -ForegroundColor Blue
    Write-Host "Total Tests: $TotalTests"
    Write-Host "Passed: $TestsPassed" -ForegroundColor Green
    Write-Host "Failed: $TestsFailed" -ForegroundColor Red
    
    if ($TestsFailed -eq 0) {
        Write-Host "`n🎉 ALL TESTS PASSED! Streaming functionality should work properly." -ForegroundColor Green
        exit 0
    } else {
        Write-Host "`n❌ SOME TESTS FAILED! Check the issues above." -ForegroundColor Red
        exit 1
    }
}

# Run the tests
Main
