# 🚀 RcloneStorage - One-Click Deployment

## Quick Start

```bash
git clone <your-repo>
cd <project-directory>
./deploy
```

**That's it!** Service akan tersedia di `http://localhost:5601`

## ✅ What's Fixed & Automated

### 🔧 Permission Issues
- ✅ **Auto-fix permissions** dalam deploy script
- ✅ **Dockerfile optimized** dengan user ID yang benar
- ✅ **Volume mounts** dengan SELinux support
- ✅ **Cache directory** writable permissions

### 🧪 Comprehensive Testing
- ✅ **Health check** validation
- ✅ **Upload functionality** test
- ✅ **Authentication** test
- ✅ **Rclone connectivity** verification

### 📚 API Documentation Fixed
- ✅ **Swagger paths** corrected
- ✅ **Auth endpoints**: `/api/auth/*` (no v1)
- ✅ **User endpoints**: `/api/user/*` (no v1)  
- ✅ **File endpoints**: `/api/v1/*` (with v1)

## 📁 Files Created

| File | Purpose |
|------|---------|
| `deploy` | **Main deployment script** (recommended) |
| `deploy-fixed.sh` | Advanced deployment with detailed logging |
| `Dockerfile.fixed` | Optimized Docker build |
| `docker-compose.yml` | Service orchestration |
| `DEPLOYMENT_GUIDE.md` | Comprehensive documentation |

## 🌐 Service Access

| Service | URL | Credentials |
|---------|-----|-------------|
| **Web App** | http://localhost:5601 | admin@rclonestorage.local / Admin123! |
| **Swagger** | http://localhost:5601/swagger/index.html | Same as above |
| **Health** | http://localhost:5601/health | Public |

## 🔑 API Endpoints Structure

```
/api/auth/*          # Authentication (login, register)
/api/user/*          # User management (profile, api-keys)
/api/admin/*         # Admin functions
/api/v1/*            # File operations (upload, download)
```

## 🛠️ Management Commands

```bash
# View logs
docker-compose logs -f

# Restart service
docker-compose restart

# Stop service
docker-compose down

# Rebuild if needed
docker-compose build --no-cache && docker-compose up -d
```

## 🔧 Troubleshooting

### Upload Issues
```bash
# Check permissions
docker-compose exec rclonestorage ls -la /app/cache/

# Fix permissions manually
docker-compose exec --user root rclonestorage \
  chown -R appuser:appgroup /app/cache /app/data /app/logs
```

### Rclone Issues
```bash
# Test rclone
docker-compose exec rclonestorage \
  sh -c "RCLONE_CONFIG=/app/configs/rclone.conf rclone listremotes"
```

## 📝 Next Steps After Deployment

1. **Configure Storage Providers** in `configs/rclone.conf`
2. **Change Admin Password** via web interface
3. **Test Upload/Download** functionality
4. **Create API Keys** for external access

## 🎯 Production Deployment

For production, update:
- JWT secret in `.env`
- Admin password
- Firewall rules
- HTTPS setup (reverse proxy)
- Storage provider credentials

---

**🎉 Deployment is now fully automated with all permission issues resolved!**