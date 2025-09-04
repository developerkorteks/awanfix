#!/bin/bash

# Complete setup script for RcloneStorage with monitoring and documentation

set -e

echo "🚀 Setting up RcloneStorage with Monitoring Dashboard and Swagger Documentation..."
echo "=================================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check prerequisites
echo "🔍 Checking prerequisites..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi
print_status "Go is installed: $(go version)"

# Check if rclone is installed
if ! command -v rclone &> /dev/null; then
    print_warning "rclone is not installed. Installing..."
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v pacman &> /dev/null; then
            sudo pacman -S rclone
        elif command -v apt &> /dev/null; then
            sudo apt update && sudo apt install rclone
        else
            curl https://rclone.org/install.sh | sudo bash
        fi
    else
        curl https://rclone.org/install.sh | sudo bash
    fi
fi
print_status "rclone is installed: $(rclone version | head -1)"

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p cache/{files,metadata,temp}
mkdir -p configs
mkdir -p logs
mkdir -p bin
mkdir -p docs
print_status "Directories created"

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod tidy
print_status "Go dependencies installed"

# Install Swagger tools
echo "📚 Installing Swagger tools..."
go install github.com/swaggo/swag/cmd/swag@latest

# Add swag to PATH if not already there
if ! command -v swag &> /dev/null; then
    print_warning "swag command not found in PATH. Adding Go bin to PATH..."
    export PATH=$PATH:$(go env GOPATH)/bin
    echo "export PATH=\$PATH:\$(go env GOPATH)/bin" >> ~/.bashrc
    print_info "Added Go bin to PATH. You may need to restart your terminal."
fi
print_status "Swagger tools installed"

# Generate Swagger documentation
echo "📝 Generating Swagger documentation..."
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
print_status "Swagger documentation generated"

# Build the application
echo "🔨 Building application..."
go build -o bin/rclonestorage cmd/server/main.go
print_status "Application built successfully"

# Check if rclone config exists
if [ ! -f "configs/rclone.conf" ]; then
    print_warning "No rclone configuration found."
    echo "📋 To complete setup, you need to configure rclone:"
    echo "   1. Run: make rclone-setup"
    echo "   2. Add your cloud storage providers (mega1, mega2, mega3, gdrive)"
    echo "   3. Test connections with: make test-rclone"
else
    print_status "rclone configuration found"
fi

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    print_info "Creating .env file from template..."
    cp .env.example .env
    print_status ".env file created"
else
    print_status ".env file already exists"
fi

echo ""
echo "🎉 Setup completed successfully!"
echo "=================================================================="
echo ""
echo "📖 Available Features:"
echo "   ✅ Multi-provider cloud storage (Mega, Google Drive)"
echo "   ✅ Authentication system with JWT and API keys"
echo "   ✅ Video streaming capabilities"
echo "   ✅ Real-time monitoring dashboard"
echo "   ✅ Swagger API documentation"
echo "   ✅ File caching system"
echo ""
echo "🚀 Quick Start:"
echo "   1. Configure storage providers: make rclone-setup"
echo "   2. Start the server: make run"
echo "   3. Access web interface: http://localhost:8080"
echo ""
echo "📚 Available Endpoints:"
echo "   🌐 Main Dashboard: http://localhost:8080"
echo "   📊 Monitoring Dashboard: http://localhost:8080/dashboard.html"
echo "   📖 API Documentation: http://localhost:8080/swagger/index.html"
echo "   🔍 Health Check: http://localhost:8080/health"
echo ""
echo "🔧 Development Commands:"
echo "   make run              - Start the server"
echo "   make build            - Build the application"
echo "   make test             - Run tests"
echo "   make swagger-gen      - Regenerate API documentation"
echo "   make rclone-setup     - Configure cloud storage"
echo "   make test-rclone      - Test storage connections"
echo ""
echo "📋 Default Admin Credentials:"
echo "   Email: admin@rclonestorage.local"
echo "   Password: Admin123!"
echo ""
print_info "For production deployment, make sure to:"
echo "   - Set JWT_SECRET environment variable"
echo "   - Configure proper host/domain in Swagger annotations"
echo "   - Set up proper SSL/TLS certificates"
echo "   - Configure firewall and security settings"