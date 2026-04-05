package command

import (
	"context"
	"errors"
	"time"

	"github.com/tazhibayda/OrbittoAuth/internal/domain/auth"
	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
)

type ResetPasswordCommand struct {
	Token       string
	NewPassword string
}

type ResetPasswordHandler struct {
	userRepo    user.Repository
	resetRepo   auth.ResetTokenRepository
	sessionRepo auth.SessionRepository
}

func NewResetPasswordHandler(
	userRepo user.Repository,
	resetRepo auth.ResetTokenRepository,
	sessionRepo auth.SessionRepository,
) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		userRepo:    userRepo,
		resetRepo:   resetRepo,
		sessionRepo: sessionRepo,
	}
}

func (h *ResetPasswordHandler) Handle(ctx context.Context, cmd ResetPasswordCommand) error {
	userID, err := h.resetRepo.UseToken(ctx, cmd.Token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	newPassVO, err := user.NewPassword(cmd.NewPassword)
	if err != nil {
		return err
	}

	newHash, err := newPassVO.Hash()
	if err != nil {
		return err
	}

	u.PasswordHash = newHash
	u.UpdatedAt = time.Now().UTC()

	if err := h.userRepo.Update(ctx, u); err != nil {
		return err
	}

	// 7. КРИТИЧНО: Разлогиниваем пользователя со всех устройств
	// Вызываем метод, который найдет и удалит все сессии этого пользователя
	if err := h.sessionRepo.RevokeAllForUser(ctx, u.ID); err != nil {
		// Пароль уже изменен, поэтому сессии нужно удалить хотя бы асинхронно,
		// но для надежности возвращаем ошибку, если Redis упал
		return err
	}

	return nil
}
