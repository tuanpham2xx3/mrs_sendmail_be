package handlers

import (
	"log"
	"net/http"
	"time"

	"mrs_sendemail_be/internal/config"
	"mrs_sendemail_be/internal/models"
	"mrs_sendemail_be/internal/services"
	"mrs_sendemail_be/internal/utils"

	"github.com/gin-gonic/gin"
)

type ActivationHandler struct {
	config       *config.Config
	redisService *services.RedisService
	smtpService  *services.SMTPService
}

func NewActivationHandler(config *config.Config, redisService *services.RedisService, smtpService *services.SMTPService) *ActivationHandler {
	return &ActivationHandler{
		config:       config,
		redisService: redisService,
		smtpService:  smtpService,
	}
}

// GenerateActivation tạo và gửi activation link
func (h *ActivationHandler) GenerateActivation(c *gin.Context) {
	// Lấy request từ context (đã được validate bởi middleware)
	reqBody, exists := c.Get("request_body")
	if !exists {
		// Fallback: bind lại nếu middleware không lưu
		var req models.GenerateActivationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		reqBody = req
	}

	req := reqBody.(models.GenerateActivationRequest)

	clientIP, _ := c.Get("client_ip")

	// Sử dụng system name mặc định nếu không có
	system := req.System
	if system == "" {
		system = h.config.Code.DefaultSystemName
	}

	now := time.Now().Unix()

	// Kiểm tra xem có thể gửi lại activation email không
	canResend, nextResendAt, err := h.redisService.CheckActivationResendLimit(c.Request.Context(), req.Email, req.Action)
	if !canResend {
		var message string
		var response models.ActivationResponse

		if err.Error() == "maximum resend limit reached" {
			message = "Đã đạt giới hạn tối đa 3 lần gửi email. Vui lòng thử lại sau."
			response = models.ActivationResponse{
				Success:   false,
				Message:   message,
				CanResend: false,
				SendCount: 3,
				MaxSends:  3,
			}
		} else {
			message = "Vui lòng chờ 60 giây trước khi gửi lại email."
			response = models.ActivationResponse{
				Success:      false,
				Message:      message,
				CanResend:    false,
				NextResendAt: nextResendAt,
				SendCount:    0, // Will be updated below
				MaxSends:     3,
			}
		}

		c.JSON(http.StatusTooManyRequests, response)
		return
	}

	// Kiểm tra xem đã có token cho email và action này chưa
	existingToken, err := h.redisService.GetActivationTokenByEmail(c.Request.Context(), req.Email, req.Action)

	var token *models.ActivationToken

	if err != nil {
		// Không có token cũ, tạo token mới
		tokenStr, err := utils.GenerateActivationToken()
		if err != nil {
			log.Printf("Error generating activation token: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to generate activation token",
			})
			return
		}

		token = &models.ActivationToken{
			Token:      tokenStr,
			Email:      req.Email,
			Action:     req.Action,
			System:     system,
			CreatedAt:  now,
			ExpiresAt:  now + (30 * 60), // 30 minutes
			SendCount:  1,
			LastSentAt: now,
		}
	} else {
		// Sử dụng lại token cũ, cập nhật send count
		existingToken.SendCount++
		existingToken.LastSentAt = now
		token = existingToken
	}

	// Tạo activation URL
	activationURL := utils.GenerateActivationURL(req.BaseURL, req.Action, token.Token)

	// Lưu/cập nhật token vào Redis
	if existingToken == nil {
		if err := h.redisService.StoreActivationToken(c.Request.Context(), token); err != nil {
			log.Printf("Error storing activation token to Redis: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to store activation token",
			})
			return
		}
	} else {
		if err := h.redisService.UpdateActivationToken(c.Request.Context(), token); err != nil {
			log.Printf("Error updating activation token in Redis: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to update activation token",
			})
			return
		}
	}

	// Gửi email
	if err := h.smtpService.SendActivationEmail(req.Email, activationURL, req.Action, system, req.CustomData); err != nil {
		log.Printf("Error sending activation email: %v", err)

		// Nếu là token mới và gửi email thất bại, xóa token
		if existingToken == nil {
			_ = h.redisService.DeleteActivationToken(c.Request.Context(), token)
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to send activation email",
		})
		return
	}

	// Tăng rate limit counters
	_ = h.redisService.IncrementEmailRateLimit(c.Request.Context(), req.Email)
	if clientIPStr, ok := clientIP.(string); ok {
		_ = h.redisService.IncrementIPRateLimit(c.Request.Context(), clientIPStr)
	}

	// Log thành công
	log.Printf("Activation email sent successfully to %s for action %s (send count: %d)", req.Email, req.Action, token.SendCount)

	// Tính toán next resend time
	nextResendTime := token.LastSentAt + 60

	// Trả về response thành công
	response := models.ActivationResponse{
		Success:      true,
		Message:      "Activation email sent successfully",
		CanResend:    token.SendCount < 3,
		NextResendAt: nextResendTime,
		SendCount:    token.SendCount,
		MaxSends:     3,
	}

	// Chỉ trả về token trong development mode
	if gin.Mode() != gin.ReleaseMode {
		response.Token = token.Token
	}

	c.JSON(http.StatusOK, response)
}

// VerifyActivation xác thực activation token
func (h *ActivationHandler) VerifyActivation(c *gin.Context) {
	var req models.VerifyActivationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
		})
		return
	}

	// Lấy activation token từ Redis
	token, err := h.redisService.GetActivationToken(c.Request.Context(), req.Token)
	if err != nil {
		log.Printf("Error getting activation token %s: %v", req.Token, err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid or Expired Token",
			Message: "Activation token not found or has expired",
		})
		return
	}

	// Kiểm tra xem token đã hết hạn chưa
	now := time.Now().Unix()
	if now > token.ExpiresAt {
		log.Printf("Activation token %s has expired", req.Token)
		// Xóa token đã hết hạn
		_ = h.redisService.DeleteActivationToken(c.Request.Context(), token)

		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Expired Token",
			Message: "Activation token has expired",
		})
		return
	}

	// Token hợp lệ, xóa khỏi Redis để tránh sử dụng lại
	if err := h.redisService.DeleteActivationToken(c.Request.Context(), token); err != nil {
		log.Printf("Error deleting activation token %s: %v", req.Token, err)
		// Không trả lỗi ở đây vì việc xác thực đã thành công
	}

	// Log thành công
	log.Printf("Activation successful for email %s with action %s", token.Email, token.Action)

	// Trả về response thành công với thông tin token
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Activation successful",
		"data": gin.H{
			"email":  token.Email,
			"action": token.Action,
			"system": token.System,
		},
	})
}

// ResendActivation gửi lại activation email
func (h *ActivationHandler) ResendActivation(c *gin.Context) {
	// Lấy request từ context (đã được validate bởi middleware)
	reqBody, exists := c.Get("request_body")
	if !exists {
		// Fallback: bind lại nếu middleware không lưu
		var req models.ResendActivationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		reqBody = req
	}

	req := reqBody.(models.ResendActivationRequest)

	// Kiểm tra xem có thể gửi lại không
	canResend, nextResendAt, err := h.redisService.CheckActivationResendLimit(c.Request.Context(), req.Email, req.Action)
	if !canResend {
		var message string
		var response models.ActivationResponse

		if err.Error() == "maximum resend limit reached" {
			message = "Đã đạt giới hạn tối đa 3 lần gửi email."
			response = models.ActivationResponse{
				Success:   false,
				Message:   message,
				CanResend: false,
				SendCount: 3,
				MaxSends:  3,
			}
		} else {
			message = "Vui lòng chờ 60 giây trước khi gửi lại email."
			response = models.ActivationResponse{
				Success:      false,
				Message:      message,
				CanResend:    false,
				NextResendAt: nextResendAt,
				SendCount:    0, // Will be updated if we can get existing token
				MaxSends:     3,
			}
		}

		c.JSON(http.StatusTooManyRequests, response)
		return
	}

	// Lấy token hiện tại
	existingToken, err := h.redisService.GetActivationTokenByEmail(c.Request.Context(), req.Email, req.Action)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "No Active Token",
			Message: "No activation token found for this email and action",
		})
		return
	}

	// Cập nhật send count và last sent time
	now := time.Now().Unix()
	existingToken.SendCount++
	existingToken.LastSentAt = now

	// Tạo activation URL
	// BaseURL sẽ được set từ config hoặc request khi gửi email

	// Cập nhật token trong Redis
	if err := h.redisService.UpdateActivationToken(c.Request.Context(), existingToken); err != nil {
		log.Printf("Error updating activation token in Redis: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update activation token",
		})
		return
	}

	// Gửi lại email (sử dụng baseURL từ request hoặc config)
	system := existingToken.System
	if system == "" {
		if req.System != "" {
			system = req.System
		} else {
			system = h.config.Code.DefaultSystemName
		}
	}

	// Sử dụng baseURL từ request, nếu không có thì dùng default từ config
	baseURL := req.BaseURL
	if baseURL == "" {
		// Default to localhost for development if not provided
		baseURL = "http://localhost:3000"
	}
	fullActivationURL := utils.GenerateActivationURL(baseURL, req.Action, existingToken.Token)

	if err := h.smtpService.SendActivationEmail(req.Email, fullActivationURL, req.Action, system, nil); err != nil {
		log.Printf("Error resending activation email: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to resend activation email",
		})
		return
	}

	// Log thành công
	log.Printf("Activation email resent successfully to %s for action %s (send count: %d)", req.Email, req.Action, existingToken.SendCount)

	// Tính toán next resend time
	nextResendTime := existingToken.LastSentAt + 60

	// Trả về response thành công
	response := models.ActivationResponse{
		Success:      true,
		Message:      "Activation email resent successfully",
		CanResend:    existingToken.SendCount < 3,
		NextResendAt: nextResendTime,
		SendCount:    existingToken.SendCount,
		MaxSends:     3,
	}

	c.JSON(http.StatusOK, response)
}
