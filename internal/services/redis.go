package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"mrs_sendemail_be/internal/config"
	"mrs_sendemail_be/internal/models"
)

type RedisService struct {
	client *redis.Client
	config *config.Config
}

func NewRedisService(cfg *config.Config) *RedisService {
	opts := &redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		DB:   cfg.Redis.DB,
	}
	
	// Chỉ set password nếu có giá trị (không rỗng)
	if cfg.Redis.Password != "" {
		opts.Password = cfg.Redis.Password
	}
	
	rdb := redis.NewClient(opts)

	return &RedisService{
		client: rdb,
		config: cfg,
	}
}

// Ping kiểm tra kết nối Redis
func (r *RedisService) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close đóng kết nối Redis
func (r *RedisService) Close() error {
	return r.client.Close()
}

// StoreVerificationCode lưu mã xác thực vào Redis
func (r *RedisService) StoreVerificationCode(ctx context.Context, email, code, system string) error {
	verificationCode := models.VerificationCode{
		Code:      code,
		Email:     email,
		System:    system,
		CreatedAt: time.Now().Unix(),
	}

	data, err := json.Marshal(verificationCode)
	if err != nil {
		return fmt.Errorf("failed to marshal verification code: %w", err)
	}

	key := fmt.Sprintf("verify:%s", email)
	expiration := time.Duration(r.config.Code.ExpireMinutes) * time.Minute

	return r.client.Set(ctx, key, data, expiration).Err()
}

// GetVerificationCode lấy mã xác thực từ Redis
func (r *RedisService) GetVerificationCode(ctx context.Context, email string) (*models.VerificationCode, error) {
	key := fmt.Sprintf("verify:%s", email)
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("verification code not found or expired")
		}
		return nil, fmt.Errorf("failed to get verification code: %w", err)
	}

	var verificationCode models.VerificationCode
	if err := json.Unmarshal([]byte(data), &verificationCode); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification code: %w", err)
	}

	return &verificationCode, nil
}

// DeleteVerificationCode xóa mã xác thực từ Redis
func (r *RedisService) DeleteVerificationCode(ctx context.Context, email string) error {
	key := fmt.Sprintf("verify:%s", email)
	return r.client.Del(ctx, key).Err()
}

// CheckEmailRateLimit kiểm tra rate limit theo email
func (r *RedisService) CheckEmailRateLimit(ctx context.Context, email string) (bool, error) {
	key := fmt.Sprintf("genlimit:email:%s", email)
	
	count, err := r.client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("failed to get email rate limit: %w", err)
	}

	return count < r.config.RateLimit.EmailPerHour, nil
}

// IncrementEmailRateLimit tăng counter rate limit theo email
func (r *RedisService) IncrementEmailRateLimit(ctx context.Context, email string) error {
	key := fmt.Sprintf("genlimit:email:%s", email)
	
	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Hour)
	
	_, err := pipe.Exec(ctx)
	return err
}

// CheckIPRateLimit kiểm tra rate limit theo IP
func (r *RedisService) CheckIPRateLimit(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("genlimit:ip:%s", ip)
	
	count, err := r.client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("failed to get IP rate limit: %w", err)
	}

	return count < r.config.RateLimit.IPPerHour, nil
}

// IncrementIPRateLimit tăng counter rate limit theo IP
func (r *RedisService) IncrementIPRateLimit(ctx context.Context, ip string) error {
	key := fmt.Sprintf("genlimit:ip:%s", ip)
	
	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Hour)
	
	_, err := pipe.Exec(ctx)
	return err
}

// GetEmailRateLimitCount lấy số lần đã gửi theo email
func (r *RedisService) GetEmailRateLimitCount(ctx context.Context, email string) (int, error) {
	key := fmt.Sprintf("genlimit:email:%s", email)
	
	count, err := r.client.Get(ctx, key).Int()
	if err != nil && err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// GetIPRateLimitCount lấy số lần đã gửi theo IP
func (r *RedisService) GetIPRateLimitCount(ctx context.Context, ip string) (int, error) {
	key := fmt.Sprintf("genlimit:ip:%s", ip)
	
	count, err := r.client.Get(ctx, key).Int()
	if err != nil && err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// ===== ACTIVATION TOKEN METHODS =====

// StoreActivationToken lưu activation token vào Redis
func (r *RedisService) StoreActivationToken(ctx context.Context, token *models.ActivationToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal activation token: %w", err)
	}

	// Store by token for verification
	tokenKey := fmt.Sprintf("activation:token:%s", token.Token)
	
	// Store by email+action for resend logic
	emailKey := fmt.Sprintf("activation:email:%s:%s", token.Email, token.Action)
	
	// 30 minutes expiration
	expiration := 30 * time.Minute

	pipe := r.client.Pipeline()
	pipe.Set(ctx, tokenKey, data, expiration)
	pipe.Set(ctx, emailKey, token.Token, expiration) // Store token reference
	
	_, err = pipe.Exec(ctx)
	return err
}

// GetActivationToken lấy activation token từ Redis bằng token
func (r *RedisService) GetActivationToken(ctx context.Context, token string) (*models.ActivationToken, error) {
	key := fmt.Sprintf("activation:token:%s", token)
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("activation token not found or expired")
		}
		return nil, fmt.Errorf("failed to get activation token: %w", err)
	}

	var activationToken models.ActivationToken
	if err := json.Unmarshal([]byte(data), &activationToken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal activation token: %w", err)
	}

	return &activationToken, nil
}

// GetActivationTokenByEmail lấy activation token từ Redis bằng email và action
func (r *RedisService) GetActivationTokenByEmail(ctx context.Context, email, action string) (*models.ActivationToken, error) {
	emailKey := fmt.Sprintf("activation:email:%s:%s", email, action)
	
	// Get token reference
	tokenRef, err := r.client.Get(ctx, emailKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no activation token found for email and action")
		}
		return nil, fmt.Errorf("failed to get token reference: %w", err)
	}

	// Get actual token data
	return r.GetActivationToken(ctx, tokenRef)
}

// UpdateActivationToken cập nhật activation token (cho resend logic)
func (r *RedisService) UpdateActivationToken(ctx context.Context, token *models.ActivationToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal activation token: %w", err)
	}

	tokenKey := fmt.Sprintf("activation:token:%s", token.Token)
	
	// Calculate remaining TTL
	ttl, err := r.client.TTL(ctx, tokenKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}
	
	if ttl <= 0 {
		return fmt.Errorf("token has expired")
	}

	// Update with remaining TTL
	return r.client.Set(ctx, tokenKey, data, ttl).Err()
}

// DeleteActivationToken xóa activation token từ Redis
func (r *RedisService) DeleteActivationToken(ctx context.Context, token *models.ActivationToken) error {
	tokenKey := fmt.Sprintf("activation:token:%s", token.Token)
	emailKey := fmt.Sprintf("activation:email:%s:%s", token.Email, token.Action)
	
	pipe := r.client.Pipeline()
	pipe.Del(ctx, tokenKey)
	pipe.Del(ctx, emailKey)
	
	_, err := pipe.Exec(ctx)
	return err
}

// CheckActivationResendLimit kiểm tra xem có thể gửi lại activation email không
func (r *RedisService) CheckActivationResendLimit(ctx context.Context, email, action string) (bool, int64, error) {
	token, err := r.GetActivationTokenByEmail(ctx, email, action)
	if err != nil {
		// No existing token, can send
		return true, 0, nil
	}

	now := time.Now().Unix()
	
	// Check if max sends reached (3 times)
	if token.SendCount >= 3 {
		return false, 0, fmt.Errorf("maximum resend limit reached")
	}
	
	// Check if 60 seconds have passed since last send
	if now-token.LastSentAt < 60 {
		nextAllowedTime := token.LastSentAt + 60
		return false, nextAllowedTime, fmt.Errorf("must wait 60 seconds between resends")
	}
	
	return true, 0, nil
} 