package auth

import (
	"context"
	"github.com/google/uuid"
)

//go:generate mockgen -source=session_repository.go -destination=mock/session_repository_mock.go -package=mock
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	SessionByID(ctx context.Context, id uuid.UUID) (*Session, error)
	RevokeRefresh(ctx context.Context, id uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	MarkRotated(ctx context.Context, id uuid.UUID) error
}
