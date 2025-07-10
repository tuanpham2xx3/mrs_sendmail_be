# Thiết Kế Microservice Gửi Mã Xác Thực Qua Gmail

## 1. Công Nghệ Sử Dụng
- **Gửi mail:** Gmail SMTP (yêu cầu App password)
- **Lưu code, rate limit:** Redis (in-memory DB)
- **Bảo mật:** API key, giới hạn theo email & IP

---

## 2. Chức Năng & API
Là micro service dùng cho nhiều hệ thống
mặc định là hệ thống fix4home
### 2.1. Health Check
- `GET /health`  
  Kiểm tra trạng thái kết nối Redis, SMTP.

### 2.2. Sinh Mã & Gửi Email
- `POST /generate`
- **Request:**
    ```json
    {
      "email": "user@example.com",
      "system": "Hệ Thống A/B",
      "customData": { ... }
    }
    ```
- **Luồng hoạt động:**
    1. Kiểm tra API key (header `x-api-key`)
    2. Rate limit: tối đa 5 lần/giờ/email, 30 lần/giờ/IP (dùng Redis)
    3. Sinh mã ngẫu nhiên, lưu vào Redis (`verify:{email}`, expire 30p)
    4. Gửi email HTML (nội dung tùy biến theo system)
    5. Phản hồi `{ success: true }`

### 2.3. Kiểm Tra Mã
- `POST /verify`
- **Request:**
    ```json
    {
      "email": "user@example.com",
      "code": "123456"
    }
    ```
- **Luồng hoạt động:**
    1. Lấy mã từ Redis (`verify:{email}`)
    2. So sánh, khớp thì xóa key và trả thành công, sai báo lỗi

---

## 3. Bảo Mật
- **API key:** Mỗi hệ thống được cấp 1 key riêng (header `x-api-key`)
- **Rate limit:** Redis giới hạn email & IP (hạn chế spam/brute force)
- **Redis:** Chỉ mở private, đặt password, không public Internet
- **Gmail:** Dùng App password, lưu .env, không hard-code
- **Chỉ expose các API cần thiết, trả thông báo tối giản**
- **Log, cảnh báo các bất thường về rate limit, truy cập**

---

## 4. Cấu Hình Redis (Ví dụ)
- `verify:{email}`: chứa mã xác thực, hết hạn 30 phút
- `genlimit:email:{email}`: đếm số lần gửi mã, reset sau 1h
- `genlimit:ip:{ip}`: đếm số lần gửi mã theo IP, reset sau 1h

---

## 5. Mẫu Middleware Kiểm Tra API Key (Node.js)
```js
function checkApiKey(req, res, next) {
  const apiKey = req.headers['x-api-key'];
  if (!apiKey || !VALID_API_KEYS.includes(apiKey)) {
    return res.status(401).json({ error: 'Invalid API key' });
  }
  next();
}
