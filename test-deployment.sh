#!/bin/bash

# RcloneStorage Deployment Test Script
# Usage: ./test-deployment.sh

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
    echo -e "${BLUE}[TEST]${NC} $1"
}

BASE_URL="http://localhost:5601"
TEST_FAILED=0

echo ""
echo "=== RcloneStorage Deployment Test Suite ==="
echo ""

# Test 1: Health Check
print_step "1. Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health" || echo "FAILED")
if echo "$HEALTH_RESPONSE" | grep -q '"status":"ok"'; then
    print_status "✅ Health check passed"
    echo "   Response: $(echo $HEALTH_RESPONSE | jq -r '.status // "N/A"') - $(echo $HEALTH_RESPONSE | jq -r '.service // "N/A"')"
else
    print_error "❌ Health check failed"
    echo "   Response: $HEALTH_RESPONSE"
    TEST_FAILED=1
fi

# Test 2: Web Interface
print_step "2. Testing web interface..."
WEB_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/" || echo "000")
if [ "$WEB_RESPONSE" = "200" ]; then
    print_status "✅ Web interface accessible"
else
    print_error "❌ Web interface failed (HTTP $WEB_RESPONSE)"
    TEST_FAILED=1
fi

# Test 3: API Documentation
print_step "3. Testing Swagger API documentation..."
SWAGGER_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/swagger/index.html" || echo "000")
if [ "$SWAGGER_RESPONSE" = "200" ]; then
    print_status "✅ Swagger documentation accessible"
else
    print_error "❌ Swagger documentation failed (HTTP $SWAGGER_RESPONSE)"
    TEST_FAILED=1
fi

# Test 4: Container Status
print_step "4. Testing container status..."
if docker-compose ps | grep -q "Up"; then
    print_status "✅ Container is running"
    CONTAINER_STATUS=$(docker-compose ps --format "table {{.Service}}\t{{.Status}}")
    echo "   $CONTAINER_STATUS"
else
    print_error "❌ Container is not running properly"
    TEST_FAILED=1
fi

# Test 5: Rclone in Container
print_step "5. Testing rclone functionality..."
RCLONE_VERSION=$(docker-compose exec -T rclonestorage rclone version 2>/dev/null | head -n1 || echo "FAILED")
if echo "$RCLONE_VERSION" | grep -q "rclone"; then
    print_status "✅ Rclone is working"
    echo "   Version: $RCLONE_VERSION"
else
    print_error "❌ Rclone test failed"
    TEST_FAILED=1
fi

# Test 6: Rclone Configuration
print_step "6. Testing rclone configuration..."
RCLONE_REMOTES=$(docker-compose exec -T rclonestorage rclone listremotes 2>/dev/null || echo "FAILED")
if [ "$RCLONE_REMOTES" != "FAILED" ] && [ ! -z "$RCLONE_REMOTES" ]; then
    print_status "✅ Rclone remotes configured"
    echo "   Available remotes:"
    echo "$RCLONE_REMOTES" | sed 's/^/     /'
else
    print_warning "⚠️  No rclone remotes configured (this is expected for initial setup)"
fi

# Test 7: File System Permissions
print_step "7. Testing file system permissions..."
WRITE_TEST=$(docker-compose exec -T rclonestorage sh -c "echo 'test' > /app/cache/test_write.tmp && rm /app/cache/test_write.tmp && echo 'OK'" 2>/dev/null || echo "FAILED")
if [ "$WRITE_TEST" = "OK" ]; then
    print_status "✅ File system permissions correct"
else
    print_error "❌ File system permission test failed"
    TEST_FAILED=1
fi

# Test 8: Database Initialization
print_step "8. Testing database initialization..."
if docker-compose exec -T rclonestorage sh -c "test -f /app/data/auth.db" 2>/dev/null; then
    print_status "✅ Database file exists"
else
    print_warning "⚠️  Database file not found (will be created on first API call)"
fi

# Test 9: API Registration Test
print_step "9. Testing user registration API..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "test@example.com",
        "password": "TestPassword123!",
        "name": "Test User"
    }' || echo "FAILED")

if echo "$REGISTER_RESPONSE" | grep -q '"token"'; then
    print_status "✅ User registration API working"
    # Extract token for further tests
    TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token // empty')
    if [ ! -z "$TOKEN" ]; then
        print_status "   JWT token received"
    fi
elif echo "$REGISTER_RESPONSE" | grep -q "already exists"; then
    print_status "✅ User registration API working (user already exists)"
else
    print_error "❌ User registration failed"
    echo "   Response: $REGISTER_RESPONSE"
    TEST_FAILED=1
fi

# Test 10: Memory and CPU Usage
print_step "10. Testing resource usage..."
STATS=$(docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep rclonestorage || echo "FAILED")
if [ "$STATS" != "FAILED" ]; then
    print_status "✅ Resource monitoring available"
    echo "   $STATS"
else
    print_warning "⚠️  Could not get resource stats"
fi

# Summary
echo ""
echo "=== Test Summary ==="
echo ""

if [ $TEST_FAILED -eq 0 ]; then
    print_status "🎉 All critical tests passed!"
    echo ""
    print_status "Your RcloneStorage deployment is ready to use:"
    echo "  🌐 Web Interface: $BASE_URL"
    echo "  📚 API Docs: $BASE_URL/swagger/index.html"
    echo "  📊 Dashboard: $BASE_URL/dashboard.html"
    echo ""
    print_status "Default admin credentials:"
    echo "  📧 Email: admin@rclonestorage.local"
    echo "  🔑 Password: Admin123!"
    echo ""
    print_warning "Next steps:"
    echo "  1. Configure your cloud storage in configs/rclone.conf"
    echo "  2. Restart the service: docker-compose restart"
    echo "  3. Test file upload/download functionality"
else
    print_error "❌ Some tests failed. Please check the logs:"
    echo "  📋 Container logs: docker-compose logs"
    echo "  🔍 Service status: docker-compose ps"
    echo ""
    exit 1
fi

echo ""