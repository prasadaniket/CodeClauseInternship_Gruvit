#!/bin/bash

# Jamendo API Setup Script for Gruvit Music Platform
# This script sets up the Jamendo API integration with your credentials

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Your Jamendo API credentials
JAMENDO_CLIENT_ID="be6cb53f"
JAMENDO_CLIENT_SECRET="94b8586b8053ee3e2bb1ff3606e0e7d5"

echo -e "${GREEN}🎵 Setting up Jamendo API for Gruvit Music Platform${NC}"
echo "=================================================="

# Function to print status
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "server/go/main.go" ]; then
    print_error "Please run this script from the Gruvit project root directory"
    exit 1
fi

print_status "Setting up Jamendo API credentials..."

# 1. Create local environment file
print_status "Creating local environment file..."
cat > server/go/.env << EOF
# Gruvit Music API Environment Variables
# Database Configuration
MONGO_URI=mongodb://localhost:27017
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-jwt-secret-key-change-this-in-production

# Jamendo API Configuration
JAMENDO_API_KEY=${JAMENDO_CLIENT_ID}
JAMENDO_CLIENT_SECRET=${JAMENDO_CLIENT_SECRET}

# Spotify API Configuration (optional)
SPOTIFY_CLIENT_ID=your-spotify-client-id
SPOTIFY_CLIENT_SECRET=your-spotify-client-secret

# Deezer API Configuration (optional)
DEEZER_APP_ID=your-deezer-app-id

# Server Configuration
PORT=8080
EOF

print_status "Environment file created: server/go/.env"

# 2. Test the Jamendo API
print_status "Testing Jamendo API integration..."

cd server/go

# Set environment variables for testing
export JAMENDO_API_KEY=${JAMENDO_CLIENT_ID}
export JAMENDO_CLIENT_SECRET=${JAMENDO_CLIENT_SECRET}

# Run the test
print_status "Running Jamendo API test..."
go run test_jamendo.go indie

if [ $? -eq 0 ]; then
    print_status "✅ Jamendo API test successful!"
else
    print_error "❌ Jamendo API test failed!"
    exit 1
fi

# 3. Start the Go service
print_status "Starting Go music service..."
print_warning "Make sure MongoDB and Redis are running before starting the service"

# Check if MongoDB is running
if ! pgrep -x "mongod" > /dev/null; then
    print_warning "MongoDB is not running. Please start MongoDB first:"
    print_warning "  - macOS: brew services start mongodb-community"
    print_warning "  - Linux: sudo systemctl start mongod"
    print_warning "  - Windows: net start MongoDB"
fi

# Check if Redis is running
if ! pgrep -x "redis-server" > /dev/null; then
    print_warning "Redis is not running. Please start Redis first:"
    print_warning "  - macOS: brew services start redis"
    print_warning "  - Linux: sudo systemctl start redis"
    print_warning "  - Windows: redis-server"
fi

print_status "Starting Go service on port 8080..."
print_warning "Press Ctrl+C to stop the service"

# Start the service
go run main.go &

# Wait a moment for the service to start
sleep 3

# Test the service
print_status "Testing the music service..."

# Test basic search
echo "Testing basic search endpoint..."
curl -s "http://localhost:8080/search?q=indie" | head -c 200
echo "..."

# Test enhanced search
echo "Testing enhanced music search endpoint..."
curl -s "http://localhost:8080/music/search?q=electronic" | head -c 200
echo "..."

# Test health endpoint
echo "Testing health endpoint..."
curl -s "http://localhost:8080/health"

echo ""
print_status "🎉 Jamendo API setup completed successfully!"
echo ""
print_status "Your Gruvit music service is now running with Jamendo integration!"
print_status "API Endpoints:"
print_status "  - Basic search: http://localhost:8080/search?q=query"
print_status "  - Enhanced search: http://localhost:8080/music/search?q=query"
print_status "  - Health check: http://localhost:8080/health"
print_status "  - Cache stats: http://localhost:8080/music/stats"
echo ""
print_status "Next steps:"
print_status "  1. Test the API endpoints above"
print_status "  2. Add Spotify API credentials for mainstream music"
print_status "  3. Add Deezer API credentials for regional hits"
print_status "  4. Deploy to production using Kubernetes"
echo ""
print_warning "Remember to keep your API credentials secure!"
