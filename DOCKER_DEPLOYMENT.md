# RcloneStorage Docker Deployment Guide

This guide provides a complete Docker deployment setup for RcloneStorage with port 5601.

## Quick Start

```bash
# Clone and deploy in one command
./deploy.sh
```

That's it! The service will be available at http://localhost:5601

## What's Included

- **Complete Docker setup** with multi-stage build
- **Docker Compose** configuration with health checks
- **Automated deployment script** (`./deploy.sh`)
- **Comprehensive testing** (`./test-deployment.sh`)
- **Pre-configured rclone** with local storage for testing
- **All dependencies** included in the container

## Files Created

- `Dockerfile` - Multi-stage Docker build
- `docker-compose.yml` - Service orchestration
- `deploy.sh` - One-click deployment script
- `test-deployment.sh` - Comprehensive testing
- `.env.docker` - Docker-optimized environment
- `DOCKER_DEPLOYMENT.md` - This guide

## Service Information

| Component | URL | Description |
|-----------|-----|-------------|
| Web Interface | http://localhost:5601 | Main application |
| API Documentation | http://localhost:5601/swagger/index.html | Swagger UI |
| Monitoring Dashboard | http://localhost:5601/dashboard.html | System monitoring |
| Health Check | http://localhost:5601/health | Service status |

## Default Credentials

- **Email**: admin@rclonestorage.local
- **Password**: Admin123!

## Directory Structure

```
├── configs/
│   └── rclone.conf          # Storage provider configuration
├── data/
│   └── auth.db             # User database (auto-created)
├── cache/
│   ├── files/              # File cache
│   ├── metadata/           # Metadata cache
│   └── temp/               # Temporary files
└── logs/                   # Application logs
```

## Configuration

### Storage Providers

Edit `configs/rclone.conf` to add your cloud storage:

```ini
# Example Mega configuration
[mega1]
type = mega
user = your-email@example.com
pass = your-encrypted-password

# Union to combine providers
[union]
type = union
upstreams = mega1: mega2: mega3:
```

### Environment Variables

Key settings in `.env`:

```bash
API_PORT=5601              # Service port
CACHE_MAX_SIZE=10737418240 # 10GB cache limit
STORAGE_PROVIDERS=mega1,mega2,mega3
JWT_SECRET=your-secret-key
```

## Commands

### Deployment
```bash
./deploy.sh                # Deploy everything
./test-deployment.sh       # Test deployment
```

### Management
```bash
docker-compose up -d       # Start services
docker-compose down        # Stop services
docker-compose restart     # Restart services
docker-compose logs -f     # View logs
```

### Development
```bash
docker-compose exec rclonestorage sh  # Container shell
docker-compose build --no-cache       # Rebuild image
```

## Testing

The deployment includes comprehensive testing:

```bash
./test-deployment.sh
```

Tests include:
- ✅ Health endpoint
- ✅ Web interface
- ✅ API documentation
- ✅ Container status
- ✅ Rclone functionality
- ✅ File permissions
- ✅ Database initialization
- ✅ User registration API
- ✅ Resource monitoring

## Troubleshooting

### Service Won't Start
```bash
# Check logs
docker-compose logs

# Check container status
docker-compose ps

# Rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Port Already in Use
```bash
# Check what's using port 5601
sudo netstat -tulpn | grep 5601

# Change port in docker-compose.yml
ports:
  - "5602:5601"  # Use different external port
```

### Rclone Issues
```bash
# Test rclone in container
docker-compose exec rclonestorage rclone version
docker-compose exec rclonestorage rclone listremotes

# Check rclone config
docker-compose exec rclonestorage cat /app/configs/rclone.conf
```

### Permission Issues
```bash
# Fix directory permissions
sudo chown -R $USER:$USER data cache logs configs

# Check container user
docker-compose exec rclonestorage id
```

## Security Notes

- Change default JWT secret in production
- Use strong passwords for admin account
- Configure proper firewall rules
- Use HTTPS in production (add reverse proxy)
- Regularly update the container image

## Performance Tuning

### Cache Settings
```bash
# Increase cache size (in bytes)
CACHE_MAX_SIZE=21474836480  # 20GB

# Adjust cache TTL
CACHE_TTL=48h  # 48 hours
```

### Resource Limits
```yaml
# In docker-compose.yml
services:
  rclonestorage:
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '1.0'
```

## Backup

### Important Data
```bash
# Backup user database
cp data/auth.db data/auth.db.backup

# Backup configuration
cp configs/rclone.conf configs/rclone.conf.backup
```

### Full Backup
```bash
# Create backup archive
tar -czf rclonestorage-backup-$(date +%Y%m%d).tar.gz \
  data/ configs/ .env docker-compose.yml
```

## Updates

```bash
# Pull latest code
git pull

# Rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Test after update
./test-deployment.sh
```

## Support

- Check logs: `docker-compose logs -f`
- Test deployment: `./test-deployment.sh`
- Health check: `curl http://localhost:5601/health`
- API docs: http://localhost:5601/swagger/index.html

## Features

- 🚀 **One-click deployment** with `./deploy.sh`
- 🐳 **Docker containerized** with health checks
- 🔒 **Authentication system** with JWT tokens
- 📁 **Multi-provider storage** (Mega, Google Drive, etc.)
- 🎥 **Video streaming** capabilities
- 📊 **Monitoring dashboard** built-in
- 📚 **API documentation** with Swagger
- 🔄 **Auto-restart** on failure
- 💾 **Persistent data** with volumes
- 🧪 **Comprehensive testing** included