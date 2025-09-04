#!/bin/bash

echo "=== Google Drive Setup Script ==="

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
print_step "Google Drive Integration Setup"
echo ""

print_status "Current system status:"
echo "- Mega accounts: 3 (mega1, mega2, mega3)"
echo "- Total storage: 45GB"
echo "- Target: Add Google Drive for 60GB total"
echo ""

print_step "Prerequisites Check:"
echo ""

# Check if rclone supports Google Drive
if rclone config providers | grep -q "drive"; then
    print_status "✅ Rclone supports Google Drive"
else
    print_error "❌ Rclone doesn't support Google Drive"
    print_warning "Please update rclone to latest version"
    exit 1
fi

# Check current rclone config
if [ -f "configs/rclone.conf" ]; then
    print_status "✅ Rclone config file exists"
else
    print_error "❌ Rclone config file not found"
    exit 1
fi

print_step "Google Cloud Console Setup Required:"
echo ""
echo "1. Go to: https://console.cloud.google.com/"
echo "2. Create new project or select existing"
echo "3. Enable Google Drive API"
echo "4. Create OAuth 2.0 credentials"
echo "5. Download credentials JSON"
echo ""

print_step "Rclone Configuration:"
echo ""
echo "Run this command to add Google Drive:"
echo "RCLONE_CONFIG=\"\$(pwd)/configs/rclone.conf\" rclone config"
echo ""
echo "Then select:"
echo "- n) New remote"
echo "- name> gdrive"
echo "- Storage> drive"
echo "- Follow OAuth2 flow"
echo ""

print_warning "Manual steps required:"
echo "1. Google Cloud Console setup (OAuth2 credentials)"
echo "2. Rclone config (OAuth2 authorization)"
echo "3. Test connection"
echo ""

print_status "After setup, run:"
echo "- Test: RCLONE_CONFIG=\"\$(pwd)/configs/rclone.conf\" rclone lsd gdrive:"
echo "- Integration: Update application config"
echo ""

print_step "Ready to proceed with Google Cloud Console setup?"
echo ""