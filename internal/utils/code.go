package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"
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