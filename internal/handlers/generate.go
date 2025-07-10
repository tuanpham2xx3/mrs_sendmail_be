package handlers

import (
	"log"
	"net/http"

	"mrs_sendemail_be/internal/config"
	"mrs_sendemail_be/internal/models"
	"mrs_sendemail_be/internal/services"
	"mrs_sendemail_be/internal/utils"

	"github.com/gin-gonic/gin"
)

type GenerateHandler struct {
	config       *config.Config
	redisService *services.RedisService
	smtpService  *services.SMTPService
}

func NewGenerateHandler(config *config.Config, redisService *services.RedisService, smtpService *services.SMTPService) *GenerateHandler {
	return &GenerateHandler{
		config:       config,
		redisService: redisService,
		smtpService:  smtpService,
	}
}

// Generate sinh mã xác thực và gửi email
func (h *GenerateHandler) Generate(c *gin.Context) {
	// Lấy request từ context (đã được validate bởi middleware)
	reqBody, exists := c.Get("request_body")
	if !exists {
		// Fallback: bind lại nếu middleware không lưu
		var req models.GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		reqBody = req
	}

	req := reqBody.(models.GenerateRequest)
	clientIP, _ := c.Get("client_ip")

	// Sinh mã xác thực
	code, err := utils.GenerateVerificationCode(h.config.Code.Length)
	if err != nil {
		log.Printf("Error generating verification code: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate verification code",
		})
		return
	}

	// Sử dụng system name mặc định nếu không có
	system := req.System
	if system == "" {
		system = h.config.Code.DefaultSystemName
	}

	// Lưu mã xác thực vào Redis
	if err := h.redisService.StoreVerificationCode(c.Request.Context(), req.Email, code, system); err != nil {
		log.Printf("Error storing verification code to Redis: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to store verification code",
		})
		return
	}

	// Gửi email
	if err := h.smtpService.SendVerificationEmail(req.Email, code, system, req.CustomData); err != nil {
		log.Printf("Error sending verification email: %v", err)

		// Xóa mã khỏi Redis nếu gửi email thất bại
		_ = h.redisService.DeleteVerificationCode(c.Request.Context(), req.Email)

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to send verification email",
		})
		return
	}

	// Tăng rate limit counters
	_ = h.redisService.IncrementEmailRateLimit(c.Request.Context(), req.Email)
	if clientIPStr, ok := clientIP.(string); ok {
		_ = h.redisService.IncrementIPRateLimit(c.Request.Context(), clientIPStr)
	}

	// Log thành công
	log.Printf("Verification code sent successfully to %s from system %s", req.Email, system)

	// Trả về response thành công
	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Verification code sent successfully",
	})
}
