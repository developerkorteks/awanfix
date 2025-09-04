# Makefile for RcloneStorage

.PHONY: help build run test clean setup install-deps swagger-install swagger-gen swagger-serve

# Default target
help:
	@echo "Available targets:"
	@echo "  setup        - Run full setup (install rclone, deps, build)"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts and cache"
	@echo "  install-deps - Install Go dependencies"
	@echo "  rclone-setup - Setup rclone configuration"
	@echo "  swagger-install - Install Swagger tools"
	@echo "  swagger-gen  - Generate Swagger documentation"
	@echo "  swagger-serve - Serve Swagger UI locally"

# Setup everything
setup:
	@echo "Running full setup..."
	./scripts/setup.sh

# Install rclone on Arch Linux
install-rclone:
	@echo "Installing rclone..."
	./scripts/install-rclone-arch.sh

# Install Go dependencies
install-deps:
	@echo "Installing Go dependencies..."
	go mod tidy

# Install Swagger tools
swagger-install:
	@echo "Installing Swagger tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Swagger tools installed successfully!"

# Generate Swagger documentation
swagger-gen:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation generated in docs/ directory"

# Serve Swagger UI locally (for development)
swagger-serve:
	@echo "Starting Swagger UI server on http://localhost:8081"
	@echo "Make sure to run 'make swagger-gen' first"
	@if command -v swagger > /dev/null; then \
		swagger serve docs/swagger.yaml -p 8081; \
	else \
		echo "Swagger CLI not installed. Install with: npm install -g swagger"; \
		echo "Or access Swagger UI at http://localhost:8080/swagger/index.html when server is running"; \
	fi

# Build the application
build:
	@echo "Building application..."
	mkdir -p bin
	go build -o bin/rclonestorage cmd/server/main.go

# Run the application
run:
	@echo "Starting RcloneStorage server..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts and cache
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf cache/files/*
	rm -rf cache/temp/*
	go clean

# Setup rclone configuration
rclone-setup:
	@echo "Setting up rclone configuration..."
	@echo "This will open rclone config. Add your Mega accounts as mega1, mega2, mega3"
	rclone config

# Test rclone connection
test-rclone:
	@echo "Testing rclone connections..."
	@if [ -f "configs/rclone.conf" ]; then \
		export RCLONE_CONFIG="$(pwd)/configs/rclone.conf"; \
		for remote in $$(rclone listremotes | grep mega); do \
			echo "Testing $$remote..."; \
			rclone lsd "$$remote" || echo "Failed to connect to $$remote"; \
		done; \
	else \
		echo "No rclone.conf found. Run 'make rclone-setup' first."; \
	fi

# Development server with auto-reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without auto-reload..."; \
		make run; \
	fi

# Create directories
dirs:
	@echo "Creating necessary directories..."
	mkdir -p cache/{files,metadata,temp}
	mkdir -p configs
	mkdir -p logs
	mkdir -p bin

# Quick test upload (requires running server)
test-upload:
	@echo "Testing file upload..."
	echo "Hello, RcloneStorage!" > /tmp/test.txt
	curl -X POST -F "file=@/tmp/test.txt" http://localhost:8080/api/v1/upload
	rm /tmp/test.txt

# Show server status
status:
	@echo "Checking server status..."
	curl -s http://localhost:8080/health | jq . || echo "Server not running or jq not installed"

# Show API endpoints
endpoints:
	@echo "Available API endpoints:"
	@echo "  GET  /health                 - Health check"
	@echo "  POST /api/v1/upload          - Upload file"
	@echo "  GET  /api/v1/files           - List files"
	@echo "  GET  /api/v1/files/:id       - Get file info"
	@echo "  GET  /api/v1/download/:id    - Download file"
	@echo "  GET  /api/v1/stream/:id      - Stream file"
	@echo "  DELETE /api/v1/files/:id     - Delete file"
	@echo "  GET  /api/v1/stats           - Get statistics"
	@echo "  POST /api/v1/cache/clear     - Clear cache"
	@echo "  GET  /swagger/index.html     - Swagger API Documentation"
	@echo "  GET  /dashboard.html         - Monitoring Dashboard"