#!/bin/bash

echo "=== RcloneStorage Setup Script ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on Arch Linux
if ! command -v pacman &> /dev/null; then
    print_warning "This script is optimized for Arch Linux. You may need to adapt it for your system."
fi

# Step 1: Install rclone
print_status "Step 1: Installing rclone..."
if command -v rclone &> /dev/null; then
    print_status "rclone is already installed: $(rclone version | head -n1)"
else
    print_status "Installing rclone via pacman..."
    sudo pacman -S rclone
    
    if [ $? -eq 0 ]; then
        print_status "rclone installed successfully!"
    else
        print_error "Failed to install rclone via pacman. Trying manual installation..."
        curl https://rclone.org/install.sh | sudo bash
    fi
fi

# Step 2: Create necessary directories
print_status "Step 2: Creating project directories..."
mkdir -p cache/{files,metadata,temp}
mkdir -p configs
mkdir -p logs

# Step 3: Copy configuration files
print_status "Step 3: Setting up configuration files..."
if [ ! -f ".env" ]; then
    cp .env.example .env
    print_status "Created .env file from example"
else
    print_warning ".env file already exists, skipping..."
fi

if [ ! -f "configs/rclone.conf" ]; then
    cp configs/rclone.conf.example configs/rclone.conf
    print_status "Created rclone.conf from example"
    print_warning "Please edit configs/rclone.conf with your actual Mega credentials"
else
    print_warning "rclone.conf already exists, skipping..."
fi

# Step 4: Install Go dependencies
print_status "Step 4: Installing Go dependencies..."
go mod tidy

if [ $? -eq 0 ]; then
    print_status "Go dependencies installed successfully!"
else
    print_error "Failed to install Go dependencies"
    exit 1
fi

# Step 5: Build the application
print_status "Step 5: Building the application..."
go build -o bin/rclonestorage cmd/server/main.go

if [ $? -eq 0 ]; then
    print_status "Application built successfully!"
else
    print_error "Failed to build application"
    exit 1
fi

# Step 6: Test rclone configuration
print_status "Step 6: Testing rclone configuration..."
if [ -f "configs/rclone.conf" ]; then
    export RCLONE_CONFIG="$(pwd)/configs/rclone.conf"
    
    # Test if any mega remotes are configured
    if rclone listremotes | grep -q "mega"; then
        print_status "Mega remotes found in rclone configuration"
        
        # Test connection to first mega remote
        FIRST_MEGA=$(rclone listremotes | grep "mega" | head -n1)
        if [ ! -z "$FIRST_MEGA" ]; then
            print_status "Testing connection to $FIRST_MEGA..."
            if timeout 10 rclone lsd "$FIRST_MEGA" > /dev/null 2>&1; then
                print_status "Connection to $FIRST_MEGA successful!"
            else
                print_warning "Failed to connect to $FIRST_MEGA. Please check your credentials."
            fi
        fi
    else
        print_warning "No Mega remotes found. Please configure rclone with your Mega accounts."
        print_status "Run: rclone config"
    fi
fi

print_status "=== Setup completed! ==="
echo ""
print_status "Next steps:"
echo "1. Edit configs/rclone.conf with your actual Mega credentials"
echo "2. Run 'rclone config' to setup your cloud storage accounts"
echo "3. Test the application: ./bin/rclonestorage"
echo "4. Access the API at http://localhost:8080"
echo ""
print_status "Useful commands:"
echo "- Test rclone: rclone lsd mega1:"
echo "- Run server: go run cmd/server/main.go"
echo "- Build: go build -o bin/rclonestorage cmd/server/main.go"