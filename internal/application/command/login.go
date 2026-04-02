package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tazhibayda/OrbittoAuth/internal/domain/auth"
	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
)

type LoginCommand struct {
	Email     string
	Password  string
	IP        string
	UserAgent string
}

type LoginResult struct {
	SessionID string
	UserID    string
}

type LoginHandler struct {
	userRepo    user.Repository
	sessionRepo auth.SessionRepository
}

func NewLoginHandler(userRepo user.Repository, sessionRepo auth.SessionRepository) *LoginHandler {
	return &LoginHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCommand) (LoginResult, error) {
	emailVO, err := user.NewEmail(cmd.Email)
	if err != nil {
		return LoginResult{}, user.ErrUserNotFound
	}

	u, err := h.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		return LoginResult{}, user.ErrUserNotFound
	}

	rawPassword, err := user.NewPassword(cmd.Password)
	if err != nil {
		return LoginResult{}, errors.New("invalid credentials")
	}

	if err := u.PasswordHash.Compare(rawPassword); err != nil {
		return LoginResult{}, errors.New("invalid credentials")
	}

	sessionDuration := 30 * 24 * time.Hour // 30 days refresh token
	session := auth.NewSession(u.ID, cmd.IP, cmd.UserAgent, sessionDuration)

	if err := h.sessionRepo.Create(ctx, session); err != nil {
		return LoginResult{}, fmt.Errorf("failed to create session: %w", err)
	}

	return LoginResult{
		SessionID: session.ID.String(),
		UserID:    u.ID.String(),
	}, nil
}
