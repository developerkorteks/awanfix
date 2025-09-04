#!/bin/bash

# RcloneStorage Docker Deployment Script
# Usage: ./deploy.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

echo ""
echo "=== RcloneStorage Docker Deployment ==="
echo ""

# Step 1: Check prerequisites
print_step "1. Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

print_status "✅ Docker and Docker Compose are available"

# Step 2: Create necessary directories
print_step "2. Creating necessary directories..."
mkdir -p cache/{files,metadata,temp}
mkdir -p configs
mkdir -p data
mkdir -p logs

print_status "✅ Directories created"

# Step 3: Setup environment file
print_step "3. Setting up environment configuration..."
if [ ! -f ".env" ]; then
    cp .env.example .env
    print_status "✅ Created .env file with port 5601"
else
    print_warning ".env file already exists, updating port..."
    sed -i 's/API_PORT=.*/API_PORT=5601/' .env
fi

# Step 4: Setup rclone configuration
print_step "4. Setting up rclone configuration..."
if [ ! -f "configs/rclone.conf" ]; then
    print_warning "Creating default rclone.conf - you'll need to configure your storage providers"
    cat > configs/rclone.conf << 'EOF'
# RcloneStorage Configuration
# Configure your storage providers here

# Example Mega configuration (uncomment and configure)
# [mega1]
# type = mega
# user = your-email1@example.com
# pass = your-encrypted-password1

# [mega2]
# type = mega
# user = your-email2@example.com
# pass = your-encrypted-password2

# [mega3]
# type = mega
# user = your-email3@example.com
# pass = your-encrypted-password3

# Union configuration to combine all accounts
# [union]
# type = union
# upstreams = mega1: mega2: mega3:

# For testing without real storage (local filesystem)
[local]
type = local
nounc = true

[union]
type = union
upstreams = local:
EOF
    print_status "✅ Created default rclone.conf with local storage for testing"
else
    print_status "✅ rclone.conf already exists"
fi

# Step 5: Build and start containers
print_step "5. Building and starting Docker containers..."

# Stop any existing containers
print_status "Stopping existing containers..."
docker-compose down 2>/dev/null || true

# Build and start
print_status "Building Docker image..."
docker-compose build

print_status "Starting containers..."
docker-compose up -d

# Step 6: Wait for service to be ready
print_step "6. Waiting for service to be ready..."
echo -n "Waiting for RcloneStorage to start"
for i in {1..30}; do
    if curl -s http://localhost:5601/health > /dev/null 2>&1; then
        echo ""
        print_status "✅ Service is ready!"
        break
    fi
    echo -n "."
    sleep 2
done

if ! curl -s http://localhost:5601/health > /dev/null 2>&1; then
    echo ""
    print_error "Service failed to start. Check logs with: docker-compose logs"
    exit 1
fi

# Step 7: Test the deployment
print_step "7. Testing deployment..."

# Test health endpoint
HEALTH_RESPONSE=$(curl -s http://localhost:5601/health)
if echo "$HEALTH_RESPONSE" | grep -q '"status":"ok"'; then
    print_status "✅ Health check passed"
else
    print_warning "Health check returned unexpected response"
fi

# Test rclone functionality
print_status "Testing rclone functionality..."
if docker-compose exec -T rclonestorage rclone version > /dev/null 2>&1; then
    print_status "✅ Rclone is working inside container"
else
    print_warning "Rclone test failed"
fi

# Step 8: Display deployment information
echo ""
print_step "🎉 Deployment completed successfully!"
echo ""
print_status "Service Information:"
echo "  🌐 Web Interface: http://localhost:5601"
echo "  📚 API Documentation: http://localhost:5601/swagger/index.html"
echo "  📊 Monitoring Dashboard: http://localhost:5601/dashboard.html"
echo "  🔍 Health Check: http://localhost:5601/health"
echo ""
print_status "Default Admin Credentials:"
echo "  📧 Email: admin@rclonestorage.local"
echo "  🔑 Password: Admin123!"
echo ""
print_status "Useful Commands:"
echo "  📋 View logs: docker-compose logs -f"
echo "  🔄 Restart: docker-compose restart"
echo "  🛑 Stop: docker-compose down"
echo "  🔧 Shell access: docker-compose exec rclonestorage sh"
echo ""
print_warning "Next Steps:"
echo "  1. Configure your storage providers in configs/rclone.conf"
echo "  2. Restart the service: docker-compose restart"
echo "  3. Test file upload/download functionality"
echo ""
print_status "Configuration Files:"
echo "  📁 configs/rclone.conf - Storage provider configuration"
echo "  📁 .env - Environment variables"
echo "  📁 data/ - Database and persistent data"
echo "  📁 cache/ - File cache directory"
echo "  📁 logs/ - Application logs"
echo ""