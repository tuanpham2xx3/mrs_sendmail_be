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

// ActivationToken represents stored activation token in Redis
type ActivationToken struct {
	Token      string `json:"token"`        // UUID token
	Email      string `json:"email"`        // Email address
	Action     string `json:"action"`       // "registration", "password_reset"
	System     string `json:"system"`       // System name
	CreatedAt  int64  `json:"created_at"`   // Unix timestamp
	ExpiresAt  int64  `json:"expires_at"`   // Unix timestamp (30 minutes from creation)
	SendCount  int    `json:"send_count"`   // Number of times email was sent (max 3)
	LastSentAt int64  `json:"last_sent_at"` // Last time email was sent
}

// GenerateActivationRequest represents request payload for /generate-activation endpoint
type GenerateActivationRequest struct {
	Email      string                 `json:"email" binding:"required,email"`
	Action     string                 `json:"action" binding:"required"` // "registration", "password_reset"
	System     string                 `json:"system,omitempty"`
	BaseURL    string                 `json:"baseUrl" binding:"required"` // Frontend base URL
	CustomData map[string]interface{} `json:"customData,omitempty"`
}

// VerifyActivationRequest represents request payload for /verify-activation endpoint
type VerifyActivationRequest struct {
	Token string `json:"token" binding:"required"`
}

// ResendActivationRequest represents request payload for /resend-activation endpoint
type ResendActivationRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Action  string `json:"action" binding:"required"` // "registration", "password_reset"
	BaseURL string `json:"baseUrl,omitempty"`         // Frontend base URL (optional, defaults to config)
	System  string `json:"system,omitempty"`          // System name (optional)
}

// ActivationResponse represents successful activation generation response
type ActivationResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
	Token        string `json:"token,omitempty"`          // Only for development/testing
	CanResend    bool   `json:"can_resend"`               // Whether user can resend email
	NextResendAt int64  `json:"next_resend_at,omitempty"` // Unix timestamp when next resend is allowed
	SendCount    int    `json:"send_count"`               // Current send count
	MaxSends     int    `json:"max_sends"`                // Maximum allowed sends (3)
}
