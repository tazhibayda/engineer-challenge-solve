package grpc

import (
	"context"
	"github.com/tazhibayda/OrbittoAuth/internal/application/port"
	"time"

	"github.com/tazhibayda/OrbittoAuth/internal/application/command"
	authv1 "github.com/tazhibayda/OrbittoAuth/pkg/api/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	registerHandler      *command.RegisterUserHandler
	loginHandler         *command.LoginHandler
	requestResetHandler  *command.RequestPasswordResetHandler
	resetPasswordHandler *command.ResetPasswordHandler
	jwtManager           port.JWTManager
}

func NewAuthServer(
	rh *command.RegisterUserHandler,
	lh *command.LoginHandler,
	rrh *command.RequestPasswordResetHandler,
	rph *command.ResetPasswordHandler,
	jwtManager port.JWTManager,
) *AuthServer {
	return &AuthServer{
		registerHandler:      rh,
		loginHandler:         lh,
		requestResetHandler:  rrh,
		resetPasswordHandler: rph,
		jwtManager:           jwtManager,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	cmd := command.RegisterUserCommand{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}

	err := s.registerHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	return &authv1.RegisterResponse{
		Message: "User registered successfully",
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	var clientIP string
	if p, ok := peer.FromContext(ctx); ok {
		clientIP = p.Addr.String()
	}

	cmd := command.LoginCommand{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		IP:        clientIP,
		UserAgent: "gRPC-Client",
	}

	result, err := s.loginHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
	}

	claims := port.AuthClaims{UserID: result.UserID}
	accessToken, err := s.jwtManager.Generate(claims, 15*time.Minute)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	return &authv1.LoginResponse{
		AccessToken: accessToken,
		SessionId:   result.SessionID,
		UserId:      result.UserID,
	}, nil
}

func (s *AuthServer) RequestPasswordReset(ctx context.Context, req *authv1.RequestPasswordResetRequest) (*authv1.RequestPasswordResetResponse, error) {
	cmd := command.RequestPasswordResetCommand{
		Email: req.GetEmail(),
	}

	token, err := s.requestResetHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process request")
	}

	return &authv1.RequestPasswordResetResponse{
		Message:    "If your email is in our system, you will receive a reset link.",
		ResetToken: token, // Мы возвращаем токен в ответе для тестирования, потом уберем его и будем отправлять только по email
	}, nil
}

func (s *AuthServer) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*authv1.ResetPasswordResponse, error) {
	cmd := command.ResetPasswordCommand{
		Token:       req.GetToken(),
		NewPassword: req.GetNewPassword(),
	}

	err := s.resetPasswordHandler.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to reset password: %v", err)
	}

	return &authv1.ResetPasswordResponse{
		Message: "Password has been reset successfully. All active sessions revoked.",
	}, nil
}
