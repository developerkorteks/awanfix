# RcloneStorage - Multi-Provider Cloud Storage System

RcloneStorage adalah sistem cloud storage yang menggabungkan multiple provider cloud storage menggunakan rclone dengan fitur authentication, caching, monitoring, dan video streaming.

## Features

- Multi-provider cloud storage (Mega, Google Drive, OneDrive, dll)
- JWT dan API Key authentication
- Video streaming dengan range request support
- Intelligent caching system (24 jam TTL)
- Real-time monitoring dashboard
- Interactive Swagger API documentation
- File upload/download dengan ownership tracking
- Admin panel untuk user management

## Quick Start

### 1. Prerequisites

```bash
# Install rclone
curl https://rclone.org/install.sh | sudo bash

# Install Go 1.21+
# Install ffmpeg (optional, untuk video processing)
```

### 2. Installation

```bash
git clone https://github.com/nabilulilalbab/rclonestorage.git
cd rclonestorage
go mod tidy
```

### 3. Setup Rclone Providers

```bash
# Setup script otomatis
./scripts/setup-complete.sh

# Atau manual
rclone config
```

### 4. Environment Configuration

```bash
cp .env.example .env
```

Edit .env file:
```bash
# Server Configuration
API_HOST=0.0.0.0
API_PORT=8080

# Security
JWT_SECRET=your-super-secure-jwt-secret-here
ADMIN_EMAIL=admin@yourdomain.com
ADMIN_PASSWORD=YourSecurePassword123!

# Storage
CACHE_DIR=./cache
CACHE_TTL=24h
CACHE_MAX_SIZE=10737418240

# Rclone
RCLONE_CONFIG_PATH=./configs/rclone.conf
```

### 5. Start Server

```bash
# Development
go run cmd/server/main.go

# Production
make build
./bin/rclonestorage
```

## Adding New Storage Providers

### Supported Providers

RcloneStorage mendukung semua provider yang didukung rclone:
- Mega
- Google Drive
- OneDrive
- Dropbox
- Amazon S3
- Backblaze B2
- pCloud
- Dan 40+ provider lainnya

### Adding Mega Provider

```bash
# 1. Run rclone config
rclone config

# 2. Pilih "n" untuk new remote
# 3. Nama remote: mega1 (atau mega2, mega3, dst)
# 4. Storage type: mega
# 5. Username: your-mega-email@example.com
# 6. Password: your-mega-password
# 7. Confirm: y

# 8. Test connection
rclone ls mega1:

# 9. Restart server untuk apply changes
```

### Adding Google Drive Provider

```bash
# 1. Setup OAuth credentials di Google Cloud Console
# 2. Download client_secret.json
# 3. Run rclone config
rclone config

# 4. Pilih "n" untuk new remote
# 5. Nama remote: gdrive (atau gdrive1, gdrive2)
# 6. Storage type: drive
# 7. Client ID: (dari Google Cloud Console)
# 8. Client Secret: (dari Google Cloud Console)
# 9. Scope: drive
# 10. Root folder: (kosongkan untuk root)
# 11. Service account: n
# 12. Auto config: y (akan buka browser)
# 13. Authorize di browser
# 14. Confirm: y

# 15. Test connection
rclone ls gdrive:
```

### Adding OneDrive Provider

```bash
rclone config
# Nama remote: onedrive1
# Storage type: onedrive
# Client ID: (kosongkan untuk default)
# Client Secret: (kosongkan untuk default)
# Region: global
# Auto config: y
# Drive type: onedrive
# Confirm: y
```

### Multiple Accounts Same Provider

```bash
# Untuk multiple akun provider yang sama
rclone config

# Mega account 1
# Remote name: mega1
# Storage: mega
# User: account1@email.com

# Mega account 2  
# Remote name: mega2
# Storage: mega
# User: account2@email.com

# Mega account 3
# Remote name: mega3
# Storage: mega
# User: account3@email.com
```

### Union Configuration

RcloneStorage menggunakan union backend untuk menggabungkan semua provider:

```bash
# Edit configs/rclone.conf
[union]
type = union
upstreams = mega1:uploads mega2:uploads mega3:uploads gdrive:uploads onedrive1:uploads
```

## API Documentation

### Base URL
```
http://localhost:8080/api
```

### Authentication

#### 1. User Registration
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "role": "user"
  }'
```

Response:
```json
{
  "message": "User registered successfully"
}
```

#### 2. User Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@rclonestorage.local",
    "password": "Admin123!"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "admin@rclonestorage.local",
    "role": "admin",
    "storage_quota": 1073741824,
    "storage_used": 104857600,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 3. Create API Key
```bash
curl -X POST http://localhost:8080/api/user/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "My API Key"
  }'
```

Response:
```json
{
  "id": "api_key_123",
  "name": "My API Key",
  "key": "rcs_1234567890abcdef",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### File Management

#### 1. Upload File
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "X-API-Key: rcs_1234567890abcdef" \
  -F "file=@/path/to/video.mp4" \
  -F "description=My video file"
```

Response:
```json
{
  "file_id": "abc123def456",
  "filename": "video.mp4",
  "message": "File uploaded successfully to cloud",
  "mime_type": "video/mp4",
  "owner": "admin@rclonestorage.local",
  "remote_path": "union:uploads/abc123def456_video.mp4",
  "size": 1048576,
  "status": "uploaded_to_cloud",
  "uploaded_at": "2024-01-01T00:00:00Z"
}
```

#### 2. List Files
```bash
curl -X GET http://localhost:8080/api/v1/files \
  -H "X-API-Key: rcs_1234567890abcdef" \
  -G \
  -d "page=1" \
  -d "limit=20" \
  -d "search=video"
```

Response:
```json
{
  "files": [
    {
      "id": "abc123def456",
      "filename": "video.mp4",
      "size": 1048576,
      "size_human": "1.0 MB",
      "mime_type": "video/mp4",
      "description": "My video file",
      "upload_date": "2024-01-01T00:00:00Z",
      "owner_id": 1,
      "provider": "mega1",
      "download_url": "/api/v1/download/abc123def456",
      "stream_url": "/api/v1/stream/abc123def456"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```

#### 3. Get File Info
```bash
curl -X GET http://localhost:8080/api/v1/files/abc123def456 \
  -H "X-API-Key: rcs_1234567890abcdef"
```

Response:
```json
{
  "file": {
    "id": "abc123def456",
    "filename": "abc123def456_video.mp4",
    "name": "video.mp4",
    "size": 1048576,
    "size_human": "1.0 MB",
    "extension": ".mp4",
    "type": "video",
    "streamable": true,
    "downloadable": true,
    "modified": "2024-01-01T00:00:00Z",
    "provider": "union"
  },
  "actions": {
    "download": "/api/v1/download/abc123def456",
    "stream": "/api/v1/stream/abc123def456",
    "delete": "/api/v1/files/abc123def456"
  },
  "message": "File info retrieved successfully"
}
```

#### 4. Download File
```bash
curl -X GET http://localhost:8080/api/v1/download/abc123def456 \
  -H "X-API-Key: rcs_1234567890abcdef" \
  -o downloaded_file.mp4
```

Headers:
```
Content-Type: application/octet-stream
Content-Length: 1048576
Content-Disposition: attachment; filename="video.mp4"
X-Cache: HIT
```

#### 5. Delete File
```bash
curl -X DELETE http://localhost:8080/api/v1/files/abc123def456 \
  -H "X-API-Key: rcs_1234567890abcdef"
```

Response:
```json
{
  "message": "File deleted successfully"
}
```

### Video Streaming

#### 1. Stream Video
```bash
# Full video stream
curl -X GET http://localhost:8080/api/v1/stream/abc123def456 \
  -H "X-API-Key: rcs_1234567890abcdef"

# Range request (untuk seeking)
curl -X GET http://localhost:8080/api/v1/stream/abc123def456 \
  -H "X-API-Key: rcs_1234567890abcdef" \
  -H "Range: bytes=0-1023"
```

Headers:
```
Content-Type: video/mp4
Content-Length: 1048576
Accept-Ranges: bytes
X-Cache: HIT
```

Range Response Headers:
```
HTTP/1.1 206 Partial Content
Content-Range: bytes 0-1023/1048576
Content-Length: 1024
```

#### 2. Get Stream Info
```bash
curl -X GET http://localhost:8080/api/v1/stream/abc123def456/info \
  -H "X-API-Key: rcs_1234567890abcdef"
```

Response:
```json
{
  "file_id": "abc123def456",
  "filename": "video.mp4",
  "size": 1048576,
  "mime_type": "video/mp4",
  "duration": 120.5,
  "bitrate": 1000000,
  "resolution": "1920x1080"
}
```

### System Monitoring

#### 1. Public Statistics
```bash
curl -X GET http://localhost:8080/api/v1/public/stats
```

Response:
```json
{
  "status": "ok",
  "public_stats": {
    "total_files": 150,
    "total_size": 1073741824,
    "size_human": "1.0 GB",
    "providers": ["mega1", "mega2", "mega3", "gdrive"],
    "provider_count": 4,
    "features": [
      "multi-provider storage",
      "video streaming",
      "authentication",
      "api keys"
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 2. System Statistics (Admin)
```bash
curl -X GET http://localhost:8080/api/v1/stats \
  -H "X-API-Key: rcs_1234567890abcdef"
```

Response:
```json
{
  "status": "ok",
  "stats": {
    "storage": {
      "total_files": 150,
      "total_size": 1073741824,
      "size_human": "1.0 GB",
      "providers": ["mega1", "mega2", "mega3", "gdrive"],
      "provider_count": 4
    },
    "cache": {
      "total_files": 25,
      "total_size": 104857600,
      "hit_rate": 0.85
    },
    "system": {
      "uptime": "24h30m15s",
      "cache_enabled": true,
      "cache_ttl": "24h",
      "max_cache_size": "10GB"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 3. Real-time Monitoring
```bash
curl -X GET http://localhost:8080/api/v1/monitoring/realtime \
  -H "X-API-Key: rcs_1234567890abcdef"
```

Response:
```json
{
  "status": "success",
  "data": {
    "system": {
      "version": "1.0.0",
      "go_version": "go1.21.0",
      "os": "linux",
      "arch": "amd64",
      "num_cpu": 8,
      "num_goroutine": 45
    },
    "storage": {
      "total_files": 150,
      "total_size": 1073741824,
      "total_size_human": "1.0 GB",
      "providers": ["mega1", "mega2", "mega3", "gdrive"],
      "provider_count": 4
    },
    "cache": {
      "total_files": 25,
      "total_size": 104857600,
      "total_size_human": "100 MB",
      "hit_rate": 0.85,
      "usage_percent": 1.0
    },
    "performance": {
      "memory_usage": 67108864,
      "memory_usage_human": "64 MB",
      "cache_hit_rate": 0.85,
      "requests_per_second": 10,
      "avg_response_time": 150
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 4. Cache Management
```bash
# Clear cache
curl -X POST http://localhost:8080/api/v1/cache/clear \
  -H "X-API-Key: rcs_1234567890abcdef"
```

Response:
```json
{
  "message": "Cache cleared successfully",
  "removed_count": 25,
  "removed_files": ["file1.mp4", "file2.jpg"],
  "cache_dir": "./cache/temp"
}
```

### User Management (Admin)

#### 1. List Users
```bash
curl -X GET http://localhost:8080/api/admin/users \
  -H "X-API-Key: rcs_1234567890abcdef"
```

Response:
```json
{
  "users": [
    {
      "id": 1,
      "email": "admin@rclonestorage.local",
      "role": "admin",
      "storage_quota": 1073741824,
      "storage_used": 104857600,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 10
}
```

#### 2. Get User
```bash
curl -X GET http://localhost:8080/api/admin/users/1 \
  -H "X-API-Key: rcs_1234567890abcdef"
```

#### 3. Create User (Admin)
```bash
curl -X POST http://localhost:8080/api/admin/users \
  -H "Content-Type: application/json" \
  -H "X-API-Key: rcs_1234567890abcdef" \
  -d '{
    "email": "newuser@example.com",
    "password": "SecurePass123!",
    "role": "user"
  }'
```

## API Structure

### Authentication Methods

1. **JWT Token**
   ```bash
   -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
   ```

2. **API Key**
   ```bash
   -H "X-API-Key: rcs_1234567890abcdef"
   ```

### Response Format

#### Success Response
```json
{
  "status": "success",
  "data": { ... },
  "message": "Operation completed successfully",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### Error Response
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Detailed error information",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### HTTP Status Codes

- `200` - Success
- `201` - Created
- `206` - Partial Content (range requests)
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

### Rate Limiting

- **Default**: 100 requests per minute per IP
- **Authenticated**: 1000 requests per minute per user
- **Admin**: Unlimited

### File Upload Limits

- **Max file size**: 5GB per file
- **Supported formats**: All formats supported
- **Concurrent uploads**: 5 per user

## Cache System

### Cache Configuration

```bash
# Cache TTL: 24 hours
# Max cache size: 10GB
# Auto cleanup: Every 12 hours
# Cache directory: ./cache/files
```

### Cache Behavior

#### Download Cache
```bash
# First download
curl http://localhost:8080/api/v1/download/abc123
# Response: X-Cache: MISS

# Subsequent downloads
curl http://localhost:8080/api/v1/download/abc123
# Response: X-Cache: HIT
```

#### Streaming Cache
```bash
# First stream (full file)
curl http://localhost:8080/api/v1/stream/abc123
# Response: X-Cache: MISS

# Subsequent streams
curl http://localhost:8080/api/v1/stream/abc123
# Response: X-Cache: HIT

# Range requests (seeking)
curl -H "Range: bytes=0-1023" http://localhost:8080/api/v1/stream/abc123
# Response: X-Cache: MISS (always from cloud for seeking)
```

### Cache Statistics
```bash
curl -X GET http://localhost:8080/api/v1/monitoring/cache \
  -H "X-API-Key: rcs_1234567890abcdef"
```

Response:
```json
{
  "total_files": 25,
  "total_size": 104857600,
  "total_size_human": "100 MB",
  "hit_rate": 0.85,
  "max_size": 10737418240,
  "usage_percent": 1.0,
  "ttl": "24h",
  "status": "active"
}
```

## Web Interface

### Available Pages

- **Main Dashboard**: http://localhost:8080
- **Login**: http://localhost:8080/login.html
- **Register**: http://localhost:8080/register.html
- **File Manager**: http://localhost:8080/files.html
- **Upload**: http://localhost:8080/upload.html
- **Profile**: http://localhost:8080/profile.html
- **Monitoring Dashboard**: http://localhost:8080/dashboard.html

### API Documentation

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **API Docs**: http://localhost:8080/docs

## Development

### Project Structure

```
rclonestorage/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── api/                    # API handlers
│   │   ├── handlers.go
│   │   ├── upload.go
│   │   ├── download.go
│   │   ├── stream.go
│   │   └── cache.go
│   ├── auth/                   # Authentication
│   │   ├── auth.go
│   │   ├── handlers.go
│   │   ├── middleware.go
│   │   ├── jwt.go
│   │   └── models.go
│   ├── cache/                  # Cache management
│   │   └── manager.go
│   ├── config/                 # Configuration
│   │   └── config.go
│   ├── monitoring/             # Monitoring
│   │   └── dashboard.go
│   └── storage/                # Storage providers
│       ├── interface.go
│       ├── union.go
│       ├── mega.go
│       └── gdrive.go
├── web/                        # Web interface
│   ├── static/
│   │   ├── css/
│   │   └── js/
│   └── templates/
├── configs/                    # Configuration files
│   ├── rclone.conf
│   └── rclone.conf.example
├── docs/                       # Swagger documentation
├── scripts/                    # Setup scripts
└── cache/                      # Cache directory
    ├── files/
    ├── metadata/
    └── temp/
```

### Build Commands

```bash
# Development
make run

# Build
make build

# Test
make test

# Clean
make clean

# Setup
make setup

# Swagger
make swagger-gen
make swagger-install
```

### Environment Variables

```bash
# Server
API_HOST=0.0.0.0
API_PORT=8080

# Security
JWT_SECRET=your-jwt-secret
ADMIN_EMAIL=admin@domain.com
ADMIN_PASSWORD=password

# Storage
CACHE_DIR=./cache
CACHE_TTL=24h
CACHE_MAX_SIZE=10737418240
RCLONE_CONFIG_PATH=./configs/rclone.conf

# Database
DB_PATH=./data/auth.db

# Logging
LOG_LEVEL=info
LOG_FILE=./logs/app.log
```

## Production Deployment

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/rclonestorage cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates rclone
COPY --from=builder /app/bin/rclonestorage .
COPY --from=builder /app/web ./web
COPY --from=builder /app/docs ./docs
EXPOSE 8080
CMD ["./rclonestorage"]
```

### Nginx Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    client_max_body_size 5G;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Systemd Service

```ini
[Unit]
Description=RcloneStorage Service
After=network.target

[Service]
Type=simple
User=rclonestorage
WorkingDirectory=/opt/rclonestorage
ExecStart=/opt/rclonestorage/bin/rclonestorage
Restart=always
EnvironmentFile=/opt/rclonestorage/.env

[Install]
WantedBy=multi-user.target
```

## Troubleshooting

### Common Issues

1. **Storage Provider Connection**
   ```bash
   # Test rclone connection
   rclone ls mega1:
   rclone ls gdrive:
   ```

2. **Cache Issues**
   ```bash
   # Clear cache
   curl -X POST http://localhost:8080/api/v1/cache/clear \
     -H "X-API-Key: YOUR_API_KEY"
   ```

3. **Authentication Issues**
   ```bash
   # Check JWT token
   curl -X GET http://localhost:8080/api/user/profile \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

### Debug Mode

```bash
export LOG_LEVEL=debug
go run cmd/server/main.go
```

### Health Check

```bash
curl http://localhost:8080/health
```

## License

MIT License

## Support

- GitHub Issues: https://github.com/nabilulilalbab/rclonestorage/issues
- Documentation: http://localhost:8080/swagger/index.html
- Monitoring: http://localhost:8080/dashboard.html