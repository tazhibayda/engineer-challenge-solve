package command

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	authMock "github.com/tazhibayda/OrbittoAuth/internal/domain/auth/mock"
	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
	userMock "github.com/tazhibayda/OrbittoAuth/internal/domain/user/mock"
)

func TestRequestPasswordResetHandler_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockRepository(ctrl)
	mockResetRepo := authMock.NewMockResetTokenRepository(ctrl)
	logger := zaptest.NewLogger(t).Sugar()

	handler := NewRequestPasswordResetHandler(mockUserRepo, mockResetRepo, logger)
	ctx := context.Background()

	validEmail, _ := user.NewEmail("test@example.com")
	existingUser := &user.User{Email: validEmail}

	tests := []struct {
		name    string
		cmd     RequestPasswordResetCommand
		setup   func()
		wantErr bool
	}{
		{
			name: "Success - token generated and saved",
			cmd:  RequestPasswordResetCommand{Email: "test@example.com"},
			setup: func() {
				mockUserRepo.EXPECT().GetByEmail(ctx, validEmail).Return(existingUser, nil).Times(1)
				mockResetRepo.EXPECT().SaveToken(ctx, gomock.Any(), existingUser.ID, gomock.Any()).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "User not found - returns nil to prevent enumeration",
			cmd:  RequestPasswordResetCommand{Email: "unknown@example.com"},
			setup: func() {
				mockUserRepo.EXPECT().GetByEmail(ctx, gomock.Any()).Return(nil, user.ErrUserNotFound).Times(1)
				mockResetRepo.EXPECT().SaveToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: false,
		},
		{
			name: "Invalid email format - returns nil silently",
			cmd:  RequestPasswordResetCommand{Email: "invalid-email"},
			setup: func() {
				mockUserRepo.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: false,
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
