#!/bin/bash

# Fixed RcloneStorage Docker Deployment Script
# Usage: ./deploy-fixed.sh

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
echo "=== RcloneStorage Fixed Docker Deployment ==="
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
fi
# Update port to 5601
sed -i 's/API_PORT=.*/API_PORT=5601/' .env
print_status "✅ Environment configured for port 5601"

# Step 4: Setup rclone configuration
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

# Step 5: Stop any existing containers
print_step "5. Stopping existing containers..."
docker-compose down 2>/dev/null || true
docker stop rclonestorage 2>/dev/null || true
docker rm rclonestorage 2>/dev/null || true

# Step 6: Build and start containers
print_step "6. Building Docker image (this may take a few minutes)..."
docker-compose build --no-cache

print_step "7. Starting containers..."
docker-compose up -d

# Step 7.5: Fix permission issues
print_step "7.5. Fixing permission issues..."
sleep 5  # Wait for container to start
if docker-compose exec --user root rclonestorage chown -R appuser:appgroup /app/cache /app/data /app/logs 2>/dev/null; then
    print_status "✅ Permissions fixed successfully"
else
    print_warning "⚠️  Could not fix permissions automatically"
fi

# Step 8: Wait for service to be ready
print_step "8. Waiting for service to be ready..."
echo -n "Waiting for RcloneStorage to start"
for i in {1..60}; do
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
    print_error "Service failed to start. Checking logs..."
    docker-compose logs --tail=20
    exit 1
fi

# Step 9: Test the deployment
print_step "9. Testing deployment..."

# Test health endpoint
HEALTH_RESPONSE=$(curl -s http://localhost:5601/health)
if echo "$HEALTH_RESPONSE" | grep -q '"status":"ok"'; then
    print_status "✅ Health check passed"
else
    print_warning "Health check returned unexpected response"
fi

# Test API routes
print_status "Testing API routes..."
if curl -s http://localhost:5601/api/auth/register > /dev/null 2>&1; then
    print_status "✅ Auth routes accessible"
else
    print_warning "Auth routes may need configuration"
fi

# Test upload functionality
print_status "Testing upload functionality..."
echo "test upload" > /tmp/deploy_test.txt
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:5601/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@rclonestorage.local","password":"Admin123!"}')

if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
    UPLOAD_RESPONSE=$(curl -s -X POST http://localhost:5601/api/v1/upload \
        -H "Authorization: Bearer $TOKEN" \
        -F "file=@/tmp/deploy_test.txt")
    
    if echo "$UPLOAD_RESPONSE" | grep -q '"uploaded_to_cloud"'; then
        print_status "✅ Upload functionality working"
    else
        print_warning "⚠️  Upload test failed - check permissions"
    fi
else
    print_warning "⚠️  Could not test upload - login failed"
fi

rm -f /tmp/deploy_test.txt

# Step 10: Display deployment information
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
print_status "API Endpoints (Fixed Routes):"
echo "  🔐 Auth: /api/auth/*"
echo "  👤 User: /api/user/* (requires JWT token)"
echo "  🔑 API Keys: /api/user/api-keys (NOT /api/v1/user/api-keys)"
echo "  👑 Admin: /api/admin/*"
echo "  📁 Files: /api/v1/*"
echo ""
print_status "Useful Commands:"
echo "  📋 View logs: docker-compose logs -f"
echo "  🔄 Restart: docker-compose restart"
echo "  🛑 Stop: docker-compose down"
echo "  🔧 Shell access: docker-compose exec rclonestorage sh"
echo ""
print_warning "Important Notes:"
echo "  ⚠️  Use /api/user/api-keys NOT /api/v1/user/api-keys"
echo "  ⚠️  JWT token required for user/admin endpoints"
echo "  ⚠️  Configure storage providers in configs/rclone.conf"
echo ""