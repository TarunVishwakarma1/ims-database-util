package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionData struct {
	Sub       string
	Sid       string
	IP        string
	UserAgent string
}

type SessionRepository interface {
	SaveRefreshToken(ctx context.Context, hashedToken string, data SessionData, expiresIn time.Duration) error
	GetSession(ctx context.Context, hashedToken string) (*SessionData, error)
	RevokeSession(ctx context.Context, hashedToken string) error
	DenyListAccessToken(ctx context.Context, jti string, expiresIn time.Duration) error
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

type redisSessionRepo struct {
	client *redis.Client
}

// NewSessionRepository creates a SessionRepository that uses the provided Redis client for persistence.
func NewSessionRepository(client *redis.Client) SessionRepository {
	return &redisSessionRepo{client: client}
}

func (r *redisSessionRepo) SaveRefreshToken(ctx context.Context, hashedToken string, data SessionData, expiresIn time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	key := fmt.Sprintf("refresh: %s", hashedToken)
	if err := r.client.Set(ctx, key, bytes, expiresIn).Err(); err != nil {
		return fmt.Errorf("failed to save refresh token to redis: %w", err)
	}
	return nil
}

func (r *redisSessionRepo) GetSession(ctx context.Context, hashedToken string) (*SessionData, error) {
	key := fmt.Sprintf("refresh: %s", hashedToken)
	val, err := r.client.Get(ctx, key).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Token not found (expired or invalid)
		}
		return nil, fmt.Errorf("failed to get session from redis: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

func (r *redisSessionRepo) RevokeSession(ctx context.Context, hashedToken string) error {
	key := fmt.Sprintf("refresh:%s", hashedToken)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *redisSessionRepo) DenyListAccessToken(ctx context.Context, jti string, expiresIn time.Duration) error {
	key := fmt.Sprintf("denylist:%s", jti)
	if err := r.client.Set(ctx, key, "revoked", expiresIn).Err(); err != nil {
		return fmt.Errorf("failed to denylist access token: %w", err)
	}
	return nil
}

func (r *redisSessionRepo) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("denylist:%s", jti)
	err := r.client.Get(ctx, key).Err()
	if err == nil {
		return true, nil // Key exists, token is revoked
	}
	if errors.Is(err, redis.Nil) {
		return false, nil // Key does not exist, token is valid
	}
	return false, fmt.Errorf("failed to check denylist: %w", err)
}
