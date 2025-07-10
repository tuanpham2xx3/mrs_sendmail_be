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
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

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