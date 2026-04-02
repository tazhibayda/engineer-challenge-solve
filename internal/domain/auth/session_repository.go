package auth

import (
	"context"
	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	SessionByID(ctx context.Context, id uuid.UUID) (*Session, error)
	RevokeRefresh(ctx context.Context, id uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	MarkRotated(ctx context.Context, id uuid.UUID) error
}
