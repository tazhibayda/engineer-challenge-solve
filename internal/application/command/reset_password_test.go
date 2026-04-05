package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	authMock "github.com/tazhibayda/OrbittoAuth/internal/domain/auth/mock"
	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
	userMock "github.com/tazhibayda/OrbittoAuth/internal/domain/user/mock"
)

func TestResetPasswordHandler_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockRepository(ctrl)
	mockResetRepo := authMock.NewMockResetTokenRepository(ctrl)
	mockSessionRepo := authMock.NewMockSessionRepository(ctrl)

	handler := NewResetPasswordHandler(mockUserRepo, mockResetRepo, mockSessionRepo)
	ctx := context.Background()

	userID := uuid.New()
	existingUser := &user.User{ID: userID}

	tests := []struct {
		name    string
		cmd     ResetPasswordCommand
		setup   func()
		wantErr bool
	}{
		{
			name: "Success - password changed and sessions revoked",
			cmd: ResetPasswordCommand{
				Token:       "valid-token",
				NewPassword: "NewStrongPassword123!",
			},
			setup: func() {
				mockResetRepo.EXPECT().UseToken(ctx, "valid-token").Return(userID, nil).Times(1)
				mockUserRepo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil).Times(1)
				mockUserRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil).Times(1)        // Проверяем обновление
				mockSessionRepo.EXPECT().RevokeAllForUser(ctx, userID).Return(nil).Times(1) // Проверяем логаут
			},
			wantErr: false,
		},
		{
			name: "Invalid or expired token",
			cmd: ResetPasswordCommand{
				Token:       "invalid-token",
				NewPassword: "NewStrongPassword123!",
			},
			setup: func() {
				mockResetRepo.EXPECT().UseToken(ctx, "invalid-token").Return(uuid.Nil, errors.New("not found")).Times(1)
				mockUserRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: true,
		},
		{
			name: "Weak new password",
			cmd: ResetPasswordCommand{
				Token:       "valid-token",
				NewPassword: "123",
			},
			setup: func() {
				mockResetRepo.EXPECT().UseToken(ctx, "valid-token").Return(userID, nil).Times(1)
				mockUserRepo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil).Times(1)
				mockUserRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)
				mockSessionRepo.EXPECT().RevokeAllForUser(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := handler.Handle(ctx, tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
