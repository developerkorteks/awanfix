# 🚀 RcloneStorage Deployment Guide

## Quick Deploy (Recommended)

```bash
git clone <your-repo>
cd <project-directory>
./deploy-fixed.sh
```

## What's Fixed

### ✅ Permission Issues
- **Dockerfile**: User ID 1000 matches host user
- **Deploy Script**: Auto-fixes permissions after container start
- **Volume Mounts**: SELinux compatible with `:Z` flag

### ✅ Upload Functionality
- **Cache Directory**: Proper write permissions
- **Rclone Integration**: Working with all providers
- **Auto-Test**: Deploy script tests upload functionality

### ✅ API Documentation
- **Swagger Paths**: Fixed routing for auth/user/admin endpoints
- **BasePath**: Correct `/api/v1` for file operations
- **Authentication**: JWT tokens working properly

## Deployment Features

### 🔧 Automatic Setup
- Creates necessary directories
- Fixes file permissions
- Tests all functionality
- Provides comprehensive status

### 🧪 Built-in Testing
- Health check validation
- Rclone connectivity test
- Upload functionality test
- API endpoint verification

### 📊 Monitoring
- Container health checks
- Service status monitoring
- Real-time logs access

## File Structure

```
├── deploy-fixed.sh          # Main deployment script
├── Dockerfile.fixed         # Optimized Docker build
├── docker-compose.yml       # Service orchestration
├── configs/rclone.conf      # Storage provider config
├── cache/                   # File cache (auto-created)
├── data/                    # Database storage
└── logs/                    # Application logs
```

## Service URLs

| Service | URL | Description |
|---------|-----|-------------|
| Web App | http://localhost:5601 | Main interface |
| Swagger | http://localhost:5601/swagger/index.html | API docs |
| Health | http://localhost:5601/health | Status check |

## API Endpoints

### Authentication (no v1)
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration

### User Management (no v1)
- `GET /api/user/profile` - Get profile
- `POST /api/user/api-keys` - Create API key
- `GET /api/user/api-keys` - List API keys
- `DELETE /api/user/api-keys/{id}` - Delete API key

### File Operations (with v1)
- `POST /api/v1/upload` - Upload file
- `GET /api/v1/files` - List files
- `GET /api/v1/download/{id}` - Download file

## Troubleshooting

### Permission Issues
```bash
# Fix manually if needed
docker-compose exec --user root rclonestorage \
  chown -R appuser:appgroup /app/cache /app/data /app/logs
```

### Container Issues
```bash
# Check logs
docker-compose logs -f

# Restart service
docker-compose restart

# Rebuild if needed
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Upload Problems
```bash
# Test rclone in container
docker-compose exec rclonestorage \
  sh -c "RCLONE_CONFIG=/app/configs/rclone.conf rclone listremotes"

# Check cache permissions
docker-compose exec rclonestorage ls -la /app/cache/
```

## Default Credentials

- **Email**: admin@rclonestorage.local
- **Password**: Admin123!

## Security Notes

- Change default JWT secret in production
- Update admin password after first login
- Configure proper firewall rules
- Use HTTPS in production

## Performance Tips

- Increase cache size for better performance
- Use SSD storage for cache directory
- Monitor container resource usage
- Regular cleanup of old cache files