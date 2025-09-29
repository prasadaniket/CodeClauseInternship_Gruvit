#!/bin/bash

# Script to switch from mock auth to integrated auth with Java service

echo "🔄 Switching to integrated authentication..."

# Backup the current main.go
if [ -f "main.go" ]; then
    cp main.go main_mock_auth.go.backup
    echo "✅ Backed up current main.go to main_mock_auth.go.backup"
fi

# Switch to integrated version
if [ -f "integrated_main.go" ]; then
    cp integrated_main.go main.go
    echo "✅ Switched to integrated authentication"
else
    echo "❌ integrated_main.go not found!"
    exit 1
fi

echo ""
echo "🚀 Integrated Authentication Features:"
echo "   • Real user validation via Java auth service"
echo "   • Proper JWT token validation"
echo "   • User registration and login"
echo "   • Token refresh functionality"
echo "   • User profile management"
echo ""
echo "📋 Next Steps:"
echo "   1. Start Java auth service: cd ../java && ./mvnw spring-boot:run"
echo "   2. Start Go music service: go run main.go"
echo "   3. Test authentication endpoints"
echo ""
echo "🔄 To switch back to mock auth: ./switch-to-mock.sh"
