// internal/domain/user/user.go
package user

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID
	Email        Email
	PasswordHash HashedPassword
	FirstName    string
	LastName     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(rawEmail, rawPassword, firstName, lastName string) (*User, error) {
	email, err := NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}

	pass, err := NewPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	hash, err := pass.Hash()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
