package command

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	_ "github.com/tazhibayda/OrbittoAuth/internal/domain/auth"
	authMock "github.com/tazhibayda/OrbittoAuth/internal/domain/auth/mock"
	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
	userMock "github.com/tazhibayda/OrbittoAuth/internal/domain/user/mock"
)

func TestLoginHandler_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMock.NewMockRepository(ctrl)
	mockSessionRepo := authMock.NewMockSessionRepository(ctrl)
	handler := NewLoginHandler(mockUserRepo, mockSessionRepo)
	ctx := context.Background()

	validEmail, _ := user.NewEmail("test@example.com")
	validPass, _ := user.NewPassword("ValidPass123!")
	hashedPass, _ := validPass.Hash()

	existingUser := &user.User{
		Email:        validEmail,
		PasswordHash: hashedPass,
	}

	tests := []struct {
		name    string
		cmd     LoginCommand
		setup   func()
		wantErr bool
	}{
		{
			name: "Success login",
			cmd: LoginCommand{
				Email:     "test@example.com",
				Password:  "ValidPass123!",
				IP:        "192.168.1.1",
				UserAgent: "Mozilla/5.0",
			},
			setup: func() {
				mockUserRepo.EXPECT().GetByEmail(ctx, validEmail).Return(existingUser, nil).Times(1)
				mockSessionRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "User not found",
			cmd: LoginCommand{
				Email:    "unknown@example.com",
				Password: "ValidPass123!",
			},
			setup: func() {
				mockUserRepo.EXPECT().GetByEmail(ctx, gomock.Any()).Return(nil, user.ErrUserNotFound).Times(1)
				mockSessionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: true,
		},
		{
			name: "Wrong password",
			cmd: LoginCommand{
				Email:    "test@example.com",
				Password: "WrongPass123!",
			},
			setup: func() {
				mockUserRepo.EXPECT().GetByEmail(ctx, validEmail).Return(existingUser, nil).Times(1)
				mockSessionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := handler.Handle(ctx, tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
