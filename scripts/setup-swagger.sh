#!/bin/bash

# Setup script for Swagger documentation in RcloneStorage

set -e

echo "🚀 Setting up Swagger documentation for RcloneStorage..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

# Install Swagger tools
echo "📦 Installing Swagger tools..."
go install github.com/swaggo/swag/cmd/swag@latest

# Add swag to PATH if not already there
if ! command -v swag &> /dev/null; then
    echo "⚠️  swag command not found in PATH. Adding Go bin to PATH..."
    export PATH=$PATH:$(go env GOPATH)/bin
    echo "export PATH=\$PATH:\$(go env GOPATH)/bin" >> ~/.bashrc
fi

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

# Generate Swagger documentation
echo "📝 Generating Swagger documentation..."
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# Create docs directory if it doesn't exist
mkdir -p docs

echo "✅ Swagger setup completed successfully!"
echo ""
echo "📖 How to use:"
echo "  1. Start the server: make run"
echo "  2. Access Swagger UI: http://localhost:8080/swagger/index.html"
echo "  3. Access API docs redirect: http://localhost:8080/docs"
echo ""
echo "🔧 Development commands:"
echo "  - Regenerate docs: make swagger-gen"
echo "  - Install tools: make swagger-install"
echo "  - Serve docs locally: make swagger-serve"
echo ""
echo "📚 Available endpoints:"
echo "  - Swagger UI: /swagger/index.html"
echo "  - Monitoring Dashboard: /dashboard.html"
echo "  - Health Check: /health"
echo "  - API Base: /api/v1/"