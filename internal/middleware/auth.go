package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mrs_sendemail_be/internal/config"
	"mrs_sendemail_be/internal/models"
)

// APIKeyAuth middleware để xác thực API key
func APIKeyAuth(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "API key is required",
			})
			c.Abort()
			return
		}

		// Kiểm tra API key có hợp lệ không
		validKey := false
		for _, validAPIKey := range config.Security.APIKeys {
			if apiKey == validAPIKey {
				validKey = true
				break
			}
		}

		if !validKey {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized", 
				Message: "Invalid API key",
			})
			c.Abort()
			return
		}

		// Lưu API key vào context để sử dụng sau này nếu cần
		c.Set("api_key", apiKey)
		c.Next()
	}
} 