package models

// GenerateRequest represents request payload for /generate endpoint
type GenerateRequest struct {
	Email      string                 `json:"email" binding:"required,email"`
	System     string                 `json:"system,omitempty"`
	CustomData map[string]interface{} `json:"customData,omitempty"`
}

// VerifyRequest represents request payload for /verify endpoint
type VerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// SuccessResponse represents successful API response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse represents error API response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HealthCheckResponse represents health check response
type HealthCheckResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// VerificationCode represents stored verification code in Redis
type VerificationCode struct {
	Code      string `json:"code"`
	Email     string `json:"email"`
	System    string `json:"system"`
	CreatedAt int64  `json:"created_at"`
} 