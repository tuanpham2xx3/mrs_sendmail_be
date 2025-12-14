package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"
	"github.com/google/uuid"
)

// GenerateVerificationCode sinh mã xác thực ngẫu nhiên
func GenerateVerificationCode(length int) (string, error) {
	const digits = "0123456789"
	
	code := make([]byte, length)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		code[i] = digits[num.Int64()]
	}
	
	return string(code), nil
}

// GetClientIP lấy IP address của client từ request
func GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string {
	// Kiểm tra X-Forwarded-For header (có thể chứa nhiều IP)
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" && ip != "unknown" {
				return ip
			}
		}
	}
	
	// Kiểm tra X-Real-IP header
	if xRealIP != "" && xRealIP != "unknown" {
		return xRealIP
	}
	
	// Fallback to RemoteAddr
	if remoteAddr != "" {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			return remoteAddr
		}
		return host
	}
	
	return "unknown"
}

// GenerateActivationToken sinh UUID token cho activation
func GenerateActivationToken() (string, error) {
	token, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID token: %w", err)
	}
	return token.String(), nil
}

// GenerateActivationURL tạo URL activation từ base URL và token
func GenerateActivationURL(baseURL, action, token string) string {
	switch action {
	case "registration":
		return fmt.Sprintf("%s/activate?token=%s", strings.TrimRight(baseURL, "/"), token)
	case "password_reset":
		return fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(baseURL, "/"), token)
	default:
		return fmt.Sprintf("%s/verify?token=%s", strings.TrimRight(baseURL, "/"), token)
	}
} 