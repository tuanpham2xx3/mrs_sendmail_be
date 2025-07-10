# Microservice Gửi Mã Xác Thực Qua Gmail

Microservice được xây dựng bằng Golang để gửi mã xác thực qua Gmail với tính năng rate limiting và bảo mật API key.

## 🚀 Tính Năng

- ✅ Gửi mã xác thực 6 số qua Gmail SMTP
- ✅ Rate limiting: 5 lần/giờ/email, 30 lần/giờ/IP
- ✅ Xác thực API key bảo mật
- ✅ Lưu trữ mã xác thực trong Redis (expire 30 phút)
- ✅ Health check endpoint
- ✅ Email template đẹp với HTML
- ✅ Multi-system support
- ✅ Docker & Docker Compose ready

## 🛠️ Công Nghệ Sử Dụng

- **Backend**: Golang 1.21, Gin Framework
- **Database**: Redis (in-memory storage)
- **Email**: Gmail SMTP với App Password
- **Container**: Docker & Docker Compose

## 📋 Yêu Cầu Hệ Thống

- Go 1.21+
- Redis Server
- Gmail account với App Password
- Docker & Docker Compose (optional)

## ⚙️ Cấu Hình

### 1. Tạo file `.env` từ `config.example`

```bash
cp config.example .env
```

### 2. Cấu hình Gmail App Password

1. Vào [Google Account Settings](https://myaccount.google.com/)
2. Bật 2-Factor Authentication
3. Tạo App Password cho ứng dụng
4. Cập nhật `SMTP_USERNAME` và `SMTP_PASSWORD` trong `.env`

### 3. Cấu hình API Keys

Cập nhật `API_KEYS` trong `.env` với các key bảo mật:

```env
API_KEYS=fix4home_secret_key,partner_system_key,admin_key_xyz
```

### 4. Cấu hình Redis

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
```

## 🚀 Chạy Ứng Dụng

### Cách 1: Sử dụng Docker Compose (Khuyến nghị)

```bash
# Clone repository
git clone <repository-url>
cd mrs_sendemail_be

# Cấu hình environment
cp config.example .env
# Chỉnh sửa .env với thông tin thực

# Chạy toàn bộ stack
docker-compose up -d

# Kiểm tra logs
docker-compose logs -f sendemail_service
```

### Cách 2: Chạy Local

```bash
# Install dependencies
go mod tidy

# Start Redis
redis-server

# Run application
go run cmd/server/main.go
```

## 📚 API Documentation

### Base URL
```
http://localhost:8080
```

### Headers Required
```
Content-Type: application/json
x-api-key: your-api-key
```

---

### 1. Health Check

**GET** `/health`

Kiểm tra trạng thái Redis và SMTP connection.

**Response Success:**
```json
{
  "status": "healthy",
  "checks": {
    "redis": "healthy",
    "smtp": "healthy"
  }
}
```

---

### 2. Sinh Mã và Gửi Email

**POST** `/generate`

**Request Body:**
```json
{
  "email": "user@example.com",
  "system": "Fix4Home App",
  "customData": {
    "userId": "12345",
    "action": "login"
  }
}
```

**Response Success:**
```json
{
  "success": true,
  "message": "Verification code sent successfully"
}
```

**Response Error - Rate Limit:**
```json
{
  "error": "Rate Limit Exceeded",
  "message": "Email rate limit exceeded. Current: 5 requests per hour for user@example.com"
}
```

---

### 3. Kiểm Tra Mã Xác Thực

**POST** `/verify`

**Request Body:**
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

**Response Success:**
```json
{
  "success": true,
  "message": "Verification successful"
}
```

**Response Error:**
```json
{
  "error": "Invalid Code",
  "message": "The verification code provided is incorrect"
}
```

## 🔒 Bảo Mật

### API Key Authentication
- Tất cả endpoints (trừ `/health`) yêu cầu `x-api-key` header
- API keys được cấu hình trong environment variables
- Mỗi hệ thống nên có key riêng

### Rate Limiting
- **Email**: Tối đa 5 lần gửi mã/giờ/email
- **IP**: Tối đa 30 lần gửi mã/giờ/IP
- Sử dụng Redis để track và reset mỗi giờ

### SMTP Security
- Sử dụng Gmail App Password thay vì password thông thường
- TLS encryption cho kết nối SMTP
- Không lưu password trong code

## 🧪 Test API

### Sử dụng curl

```bash
# Health check
curl -X GET http://localhost:8080/health

# Generate code
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -H "x-api-key: fix4home_secret_key" \
  -d '{
    "email": "test@example.com",
    "system": "Fix4Home Test"
  }'

# Verify code
curl -X POST http://localhost:8080/verify \
  -H "Content-Type: application/json" \
  -H "x-api-key: fix4home_secret_key" \
  -d '{
    "email": "test@example.com",
    "code": "123456"
  }'
```

### Sử dụng Postman

Import collection với các endpoint trên và thêm:
- Base URL: `http://localhost:8080`
- Header: `x-api-key: your-api-key`

## 🐳 Production Deployment

### 1. Cấu hình Production

```bash
# Set production environment
export GIN_MODE=release

# Update docker-compose.yml environment variables
# - Strong API keys
# - Real Gmail credentials
# - Strong Redis password
```

### 2. SSL/TLS (Khuyến nghị)

Sử dụng reverse proxy như Nginx hoặc Traefik để handle SSL:

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 3. Monitoring & Logging

```bash
# View logs
docker-compose logs -f sendemail_service

# Monitor Redis
docker exec -it mrs_sendemail_redis redis-cli -a your_redis_password monitor
```

## 🔧 Troubleshooting

### 1. SMTP Connection Failed
- Kiểm tra Gmail App Password
- Verify Gmail account settings
- Check firewall/network

### 2. Redis Connection Failed
- Kiểm tra Redis service đang chạy
- Verify Redis password
- Check Redis port accessibility

### 3. Rate Limit Issues
- Adjust rate limit values in config
- Clear Redis rate limit keys manually:
```bash
redis-cli -a password
> DEL genlimit:email:user@example.com
> DEL genlimit:ip:192.168.1.1
```

### 4. Email Not Received
- Check spam/junk folder
- Verify email address format
- Check SMTP logs for errors

## 📞 Support

- **GitHub Issues**: [Create Issue](https://github.com/your-repo/issues)
- **Email**: admin@fix4home.com
- **Documentation**: [Wiki](https://github.com/your-repo/wiki)

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Fix4Home Email Verification Microservice** - Built with ❤️ in Vietnam 