# 🐳 Hướng Dẫn Cài Đặt Docker

## 📋 Cấu Hình Docker Hiện Tại

### Redis Database
```yaml
redis:
  image: redis:7
  container_name: sendmail_db
  ports:
    - "6379:6379"
  command: ["redis-server", "--requirepass", "mypassword"]
  volumes:
    - ./redis_data:/data
```

### Microservice
```yaml
sendemail_service:
  container_name: mrs_sendemail_service
  ports:
    - "8200:8200"
  environment:
    REDIS_PASSWORD: mypassword
    # ... other configs
```

## 🚀 Cách Chạy

### 1. Chạy Chỉ Redis
```bash
# Chạy container Redis riêng lẻ
docker run -d \
  --name sendmail_db \
  -p 6379:6379 \
  -v ./redis_data:/data \
  redis:7 redis-server --requirepass mypassword
```

### 2. Chạy Toàn Bộ Stack
```bash
# Chạy cả Redis và Microservice
docker-compose up -d

# Xem logs
docker-compose logs -f sendemail_service

# Stop services
docker-compose down
```

### 3. Chạy Riêng Microservice (Redis bên ngoài)
```bash
# Nếu Redis đã chạy riêng
go run cmd/server/main.go
```

## 🔧 Test Kết Nối

### Test Redis
```bash
# Test Redis connection
docker exec -it sendmail_db redis-cli -a mypassword ping
# Response: PONG

# Xem các keys
docker exec -it sendmail_db redis-cli -a mypassword keys "*"
```

### Test Microservice
```bash
# Health check
curl http://localhost:8200/health

# PowerShell
Invoke-RestMethod -Uri "http://localhost:8200/health" -Method GET
```

## 📂 Cấu Trúc Files

```
mrs_sendemail_be/
├── docker-compose.yml    # Cấu hình Docker
├── Dockerfile           # Build microservice
├── .env                # Environment variables
├── environment.template # Template config
└── redis_data/         # Redis data volume (tự tạo)
```

## ⚙️ Environment Variables

File `.env` quan trọng:
```env
# Server
SERVER_PORT=8200

# Redis  
REDIS_PASSWORD=mypassword

# Gmail (cần thay đổi)
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# API Keys (cần thay đổi)
API_KEYS=your-api-keys-here
```

## 🐛 Troubleshooting

### 1. Port đã được sử dụng
```bash
# Kiểm tra port 8200
netstat -an | findstr :8200

# Kill process nếu cần
taskkill /f /pid <PID>
```

### 2. Redis connection failed
```bash
# Kiểm tra Redis container
docker ps | findstr redis

# Restart Redis
docker restart sendmail_db
```

### 3. Build error
```bash
# Rebuild image
docker-compose build --no-cache

# Xem logs chi tiết
docker-compose logs sendemail_service
```

## 📝 Notes

- **Port**: Microservice chạy trên port **8200** (thay vì 8080)
- **Redis Password**: `mypassword` (có thể thay đổi trong docker-compose.yml)
- **Data Volume**: `./redis_data` trong thư mục project
- **Container Names**: 
  - Redis: `sendmail_db`
  - Microservice: `mrs_sendemail_service` 