package command

import (
	"context"
	"fmt"

	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
)

type RegisterUserCommand struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterUserHandler struct {
	userRepo user.Repository
}

func NewRegisterUserHandler(repo user.Repository) *RegisterUserHandler {
	return &RegisterUserHandler{userRepo: repo}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) error {
	u, err := user.NewUser(cmd.Email, cmd.Password, cmd.FirstName, cmd.LastName)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	err = h.userRepo.Create(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}
