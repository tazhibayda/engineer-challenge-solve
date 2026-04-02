package command

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
	"github.com/tazhibayda/OrbittoAuth/internal/domain/user/mock"
)

func TestRegisterUserHandler_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	handler := NewRegisterUserHandler(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name    string
		cmd     RegisterUserCommand
		setup   func()
		wantErr bool
	}{
		{
			name: "Success registration",
			cmd: RegisterUserCommand{
				Email:     "test@example.com",
				Password:  "StrongPass123!",
				FirstName: "John",
				LastName:  "Doe",
			},
			setup: func() {
				mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Validation error (weak password)",
			cmd: RegisterUserCommand{
				Email:     "test@example.com",
				Password:  "123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setup: func() {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: true,
		},
		{
			name: "Repository error (email already exists)",
			cmd: RegisterUserCommand{
				Email:     "exist@example.com",
				Password:  "StrongPass123!",
				FirstName: "Jane",
				LastName:  "Doe",
			},
			setup: func() {
				mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(user.ErrUserAlreadyExists).Times(1)
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
