package user

import (
	"context"
	"github.com/google/uuid"
)

//go:generate mockgen -source=user_repository.go -destination=mock/user_repository_mock.go -package=mock
type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
}
