# 🔧 API Routes Fix - Solusi Masalah 404

## 🚨 Masalah yang Ditemukan:

1. **Docker Issue**: Binary tidak ada di container
2. **API Route Issue**: Path yang salah di Swagger

## ✅ Solusi:

### 1. Docker Fix
- Gunakan `Dockerfile.fixed` yang build binary dengan benar
- Update `docker-compose.yml` untuk menggunakan Dockerfile yang benar

### 2. API Routes yang Benar:

| ❌ Path Salah (404) | ✅ Path Benar |
|---------------------|---------------|
| `/api/v1/user/api-keys` | `/api/user/api-keys` |
| `/api/v1/user/profile` | `/api/user/profile` |
| `/api/v1/admin/users` | `/api/admin/users` |

### 3. Struktur API yang Benar:

```
/api/auth/*          - Public authentication (register, login)
/api/user/*          - User endpoints (requires JWT)
/api/admin/*         - Admin endpoints (requires JWT + admin role)
/api/v1/*            - File operations (upload, download, etc)
```

## 🔑 Cara Test API Keys yang Benar:

### Via Swagger:
```
POST /api/user/api-keys
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "name": "My API Key"
}
```

### Via cURL:
```bash
curl -X POST http://localhost:5601/api/user/api-keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "My API Key"}'
```

## 🚀 Deploy yang Benar:

```bash
# Clone repository
git clone <your-repo>
cd <your-project>

# Deploy dengan fix
./deploy-fixed.sh
```

## 📝 Catatan Penting:

1. **JWT Token**: Ambil dari login response atau web interface
2. **API Path**: Jangan gunakan `/api/v1/` untuk user/admin endpoints
3. **Docker**: Gunakan `Dockerfile.fixed` untuk build yang benar
4. **Testing**: Test di http://localhost:5601/swagger/index.html

## 🔍 Debug Commands:

```bash
# Check container logs
docker-compose logs -f

# Check if binary exists in container
docker-compose exec rclonestorage ls -la /app/

# Test API routes
curl http://localhost:5601/health
curl http://localhost:5601/api/auth/register
```