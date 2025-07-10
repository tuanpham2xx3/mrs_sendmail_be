package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mrs_sendemail_be/internal/models"
	"mrs_sendemail_be/internal/services"
)

type HealthHandler struct {
	redisService *services.RedisService
	smtpService  *services.SMTPService
}

func NewHealthHandler(redisService *services.RedisService, smtpService *services.SMTPService) *HealthHandler {
	return &HealthHandler{
		redisService: redisService,
		smtpService:  smtpService,
	}
}

// HealthCheck kiểm tra trạng thái các dịch vụ
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	checks := make(map[string]string)
	
	// Kiểm tra Redis
	if err := h.redisService.Ping(c.Request.Context()); err != nil {
		checks["redis"] = "unhealthy: " + err.Error()
	} else {
		checks["redis"] = "healthy"
	}
	
	// Kiểm tra SMTP
	if err := h.smtpService.TestConnection(); err != nil {
		checks["smtp"] = "unhealthy: " + err.Error()
	} else {
		checks["smtp"] = "healthy"
	}
	
	// Xác định trạng thái tổng thể
	status := "healthy"
	for _, check := range checks {
		if check != "healthy" {
			status = "unhealthy"
			break
		}
	}
	
	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	response := models.HealthCheckResponse{
		Status: status,
		Checks: checks,
	}
	
	c.JSON(statusCode, response)
} 