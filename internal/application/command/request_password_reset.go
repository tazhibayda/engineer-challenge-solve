package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/tazhibayda/OrbittoAuth/internal/domain/auth"
	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
)

type RequestPasswordResetCommand struct {
	Email string
}

type RequestPasswordResetHandler struct {
	userRepo  user.Repository
	resetRepo auth.ResetTokenRepository
	logger    *zap.SugaredLogger
}

func NewRequestPasswordResetHandler(
	userRepo user.Repository,
	resetRepo auth.ResetTokenRepository,
	logger *zap.SugaredLogger,
) *RequestPasswordResetHandler {
	return &RequestPasswordResetHandler{
		userRepo:  userRepo,
		resetRepo: resetRepo,
		logger:    logger,
	}
}

func (h *RequestPasswordResetHandler) Handle(ctx context.Context, cmd RequestPasswordResetCommand) error {
	emailVO, err := user.NewEmail(cmd.Email)
	if err != nil {
		return nil
	}

	u, err := h.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		return nil
	}

	token := uuid.New().String()

	err = h.resetRepo.SaveToken(ctx, token, u.ID, 15*time.Minute)
	if err != nil {
		h.logger.Errorf("Failed to save reset token in Redis: %v", err)
		return err
	}

	h.logger.Infof("🔑 RESET PASSWORD LINK: http://localhost:8080/v1/auth/password-reset/confirm?token=%s", token)
	// Временно выводим ссылку в лог (заглушка для EmailSender)
	return nil
}
