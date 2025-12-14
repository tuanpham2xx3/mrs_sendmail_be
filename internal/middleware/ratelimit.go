package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"mrs_sendemail_be/internal/models"
	"mrs_sendemail_be/internal/services"
	"mrs_sendemail_be/internal/utils"
)

// RateLimit middleware để kiểm tra rate limiting
func RateLimit(redisService *services.RedisService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy IP của client
		clientIP := utils.GetClientIP(
			c.Request.RemoteAddr,
			c.GetHeader("X-Forwarded-For"),
			c.GetHeader("X-Real-IP"),
		)

		// Kiểm tra rate limit theo IP
		ipAllowed, err := redisService.CheckIPRateLimit(c.Request.Context(), clientIP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to check IP rate limit",
			})
			c.Abort()
			return
		}

		if !ipAllowed {
			// Lấy số lần đã gửi để thông báo chi tiết
			count, _ := redisService.GetIPRateLimitCount(c.Request.Context(), clientIP)
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error:   "Rate Limit Exceeded",
				Message: fmt.Sprintf("IP rate limit exceeded. Current: %d requests per hour", count),
			})
			c.Abort()
			return
		}

		// Lưu IP vào context để sử dụng sau này
		c.Set("client_ip", clientIP)
		c.Next()
	}
}

// EmailRateLimit middleware để kiểm tra rate limiting theo email
// Hỗ trợ cả GenerateRequest (legacy) và GenerateActivationRequest (new)
func EmailRateLimit(redisService *services.RedisService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var email string
		var requestBody interface{}
		
		// Kiểm tra xem đây có phải là activation endpoint không
		isActivationEndpoint := c.Request.URL.Path == "/generate-activation" || 
		                        c.Request.URL.Path == "/resend-activation"
		
		if isActivationEndpoint {
			// Bind vào GenerateActivationRequest cho activation endpoints
			var req models.GenerateActivationRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   "Bad Request",
					Message: err.Error(),
				})
				c.Abort()
				return
			}
			email = req.Email
			requestBody = req
		} else {
			// Bind vào GenerateRequest cho legacy endpoints
			var req models.GenerateRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   "Bad Request",
					Message: err.Error(),
				})
				c.Abort()
				return
			}
			email = req.Email
			requestBody = req
		}

		// Kiểm tra rate limit theo email
		emailAllowed, err := redisService.CheckEmailRateLimit(c.Request.Context(), email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to check email rate limit",
			})
			c.Abort()
			return
		}

		if !emailAllowed {
			// Lấy số lần đã gửi để thông báo chi tiết
			count, _ := redisService.GetEmailRateLimitCount(c.Request.Context(), email)
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error:   "Rate Limit Exceeded",
				Message: fmt.Sprintf("Email rate limit exceeded. Current: %d requests per hour for %s", count, email),
			})
			c.Abort()
			return
		}

		// Lưu request body vào context để tránh bind lại
		c.Set("request_body", requestBody)
		c.Next()
	}
} 