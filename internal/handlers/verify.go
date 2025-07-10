package handlers

import (
	"log"
	"net/http"

	"mrs_sendemail_be/internal/models"
	"mrs_sendemail_be/internal/services"

	"github.com/gin-gonic/gin"
)

type VerifyHandler struct {
	redisService *services.RedisService
}

func NewVerifyHandler(redisService *services.RedisService) *VerifyHandler {
	return &VerifyHandler{
		redisService: redisService,
	}
}

// Verify kiểm tra mã xác thực
func (h *VerifyHandler) Verify(c *gin.Context) {
	var req models.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
		})
		return
	}

	// Lấy mã xác thực từ Redis
	storedCode, err := h.redisService.GetVerificationCode(c.Request.Context(), req.Email)
	if err != nil {
		log.Printf("Error getting verification code for %s: %v", req.Email, err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid or Expired Code",
			Message: "Verification code not found or has expired",
		})
		return
	}

	// So sánh mã xác thực
	if storedCode.Code != req.Code {
		log.Printf("Invalid verification code attempt for %s", req.Email)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid Code",
			Message: "The verification code provided is incorrect",
		})
		return
	}

	// Mã xác thực đúng, xóa khỏi Redis để tránh sử dụng lại
	if err := h.redisService.DeleteVerificationCode(c.Request.Context(), req.Email); err != nil {
		log.Printf("Error deleting verification code for %s: %v", req.Email, err)
		// Không trả lỗi ở đây vì việc xác thực đã thành công
	}

	// Log thành công
	log.Printf("Verification successful for %s with system %s", req.Email, storedCode.System)

	// Trả về response thành công
	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Verification successful",
	})
}
