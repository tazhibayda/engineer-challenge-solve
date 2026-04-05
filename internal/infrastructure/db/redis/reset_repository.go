package redis

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const resetPrefix = "reset_token:"

type ResetRepository struct {
	client *redis.Client
}

func NewResetRepository(client *redis.Client) *ResetRepository {
	return &ResetRepository{client: client}
}

func (r *ResetRepository) SaveToken(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error {
	key := resetPrefix + token
	return r.client.Set(ctx, key, userID.String(), ttl).Err()
}

func (r *ResetRepository) UseToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := resetPrefix + token

	userIDStr, err := r.client.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, errors.New("token not found or expired")
		}
		return uuid.Nil, err
	}

	return uuid.Parse(userIDStr)
}
