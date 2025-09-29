#!/bin/bash

# Script to switch from integrated auth back to mock auth

echo "🔄 Switching to mock authentication..."

# Backup the current main.go
if [ -f "main.go" ]; then
    cp main.go main_integrated.go.backup
    echo "✅ Backed up current main.go to main_integrated.go.backup"
fi

# Switch to mock version
if [ -f "main_mock_auth.go.backup" ]; then
    cp main_mock_auth.go.backup main.go
    echo "✅ Switched to mock authentication"
else
    echo "❌ main_mock_auth.go.backup not found!"
    echo "   Please restore from git or create a new mock implementation"
    exit 1
fi

echo ""
echo "⚠️  Mock Authentication Features:"
echo "   • Simple username/password validation"
echo "   • No real user database"
echo "   • Basic JWT token generation"
echo "   • No token refresh"
echo ""
echo "📋 Next Steps:"
echo "   1. Start Go music service: go run main.go"
echo "   2. Test with any username/password"
echo ""
echo "🔄 To switch to integrated auth: ./switch-to-integrated.sh"
