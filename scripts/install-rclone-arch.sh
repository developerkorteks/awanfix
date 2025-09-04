#!/bin/bash

echo "=== Installing rclone on Arch Linux ==="

# Method 1: Using pacman (recommended)
echo "Installing rclone via pacman..."
sudo pacman -S rclone

# Verify installation
echo "Verifying rclone installation..."
rclone version

echo "=== Rclone installation completed! ==="
echo "Next steps:"
echo "1. Run 'rclone config' to setup your cloud storage accounts"
echo "2. For Mega: rclone config -> New remote -> mega -> follow prompts"
echo "3. Test connection: rclone ls mega-remote-name:"