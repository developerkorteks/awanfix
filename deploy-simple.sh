#!/bin/bash

# Simple RcloneStorage Deployment Script
# Usage: ./deploy-simple.sh

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
echo "=== RcloneStorage Simple Deployment ==="
echo ""

# Step 1: Check prerequisites
print_step "1. Checking prerequisites..."

if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

if ! command -v rclone &> /dev/null; then
    print_error "Rclone is not installed. Please install rclone first."
    exit 1
fi

print_status "✅ Go and rclone are available"

# Step 2: Create directories
print_step "2. Creating necessary directories..."
mkdir -p cache/{files,metadata,temp}
mkdir -p configs
mkdir -p data
mkdir -p logs
mkdir -p bin

print_status "✅ Directories created"

# Step 3: Setup environment
print_step "3. Setting up environment configuration..."
if [ ! -f ".env" ]; then
    cp .env.example .env
fi
sed -i 's/API_PORT=.*/API_PORT=5601/' .env
print_status "✅ Environment configured for port 5601"

# Step 4: Setup rclone config
print_step "4. Setting up rclone configuration..."
if [ ! -f "configs/rclone.conf" ]; then
    print_status "Creating default rclone.conf with local storage for testing..."
    cat > configs/rclone.conf << 'EOF'
# RcloneStorage Configuration
# For testing with local filesystem

[local]
type = local
nounc = true

[union]
type = union
upstreams = local:

# Add your cloud storage providers here:
# [mega1]
# type = mega
# user = your-email@example.com
# pass = your-encrypted-password

# [gdrive]
# type = drive
# client_id = your-client-id
# client_secret = your-client-secret
# token = your-oauth-token
EOF
    print_status "✅ Created rclone.conf with local storage"
else
    print_status "✅ rclone.conf already exists"
fi

# Step 5: Install dependencies and build
print_step "5. Installing dependencies and building..."
go mod tidy
go build -o bin/rclonestorage cmd/server/main.go

if [ $? -eq 0 ]; then
    print_status "✅ Application built successfully"
else
    print_error "❌ Build failed"
    exit 1
fi

# Step 6: Test rclone
print_step "6. Testing rclone configuration..."
export RCLONE_CONFIG="$(pwd)/configs/rclone.conf"
REMOTES=$(rclone listremotes)
if [ ! -z "$REMOTES" ]; then
    print_status "✅ Rclone remotes configured:"
    echo "$REMOTES" | sed 's/^/     /'
else
    print_warning "⚠️  No rclone remotes found"
fi

# Step 7: Stop any existing instance
print_step "7. Stopping any existing instance..."
pkill -f "rclonestorage" || true
sleep 2

# Step 8: Start the service
print_step "8. Starting RcloneStorage service..."
API_PORT=5601 nohup ./bin/rclonestorage > logs/rclonestorage.log 2>&1 &
SERVICE_PID=$!
echo $SERVICE_PID > rclonestorage.pid

# Step 9: Wait for service to start
print_step "9. Waiting for service to start..."
echo -n "Starting"
for i in {1..15}; do
    if curl -s http://localhost:5601/health > /dev/null 2>&1; then
        echo ""
        print_status "✅ Service started successfully!"
        break
    fi
    echo -n "."
    sleep 2
done

if ! curl -s http://localhost:5601/health > /dev/null 2>&1; then
    echo ""
    print_error "❌ Service failed to start. Check logs/rclonestorage.log"
    exit 1
fi

# Step 10: Test the service
print_step "10. Testing the deployment..."

# Health check
HEALTH=$(curl -s http://localhost:5601/health | jq -r '.status // "failed"')
if [ "$HEALTH" = "ok" ]; then
    print_status "✅ Health check passed"
else
    print_error "❌ Health check failed"
fi

# Test rclone in service
STATS=$(curl -s http://localhost:5601/api/v1/public/stats | jq -r '.status // "failed"')
if [ "$STATS" = "ok" ]; then
    print_status "✅ Rclone integration working"
else
    print_warning "⚠️  Rclone integration may need configuration"
fi

# Final success message
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
print_status "Service Management:"
echo "  📋 View logs: tail -f logs/rclonestorage.log"
echo "  🛑 Stop service: kill \$(cat rclonestorage.pid)"
echo "  🔄 Restart: ./deploy-simple.sh"
echo ""
print_status "Configuration Files:"
echo "  📁 configs/rclone.conf - Storage provider configuration"
echo "  📁 .env - Environment variables"
echo "  📁 data/ - Database and persistent data"
echo "  📁 cache/ - File cache directory"
echo "  📁 logs/ - Application logs"
echo ""
print_warning "Next Steps:"
echo "  1. Configure your cloud storage in configs/rclone.conf"
echo "  2. Restart the service: ./deploy-simple.sh"
echo "  3. Test file upload/download functionality"
echo ""
print_status "Process ID: $SERVICE_PID (saved to rclonestorage.pid)"
echo ""