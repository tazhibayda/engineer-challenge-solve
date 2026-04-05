package auth

import (
	"context"
	"github.com/google/uuid"
	"time"
)

//go:generate mockgen -source=reset_repository.go -destination=mock/reset_repository_mock.go -package=mock
type ResetTokenRepository interface {
	SaveToken(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error
	UseToken(ctx context.Context, token string) (uuid.UUID, error)
}
