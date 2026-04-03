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
	registerHandler *command.RegisterUserHandler
	loginHandler    *command.LoginHandler
	jwtManager      port.JWTManager
}

func NewAuthServer(rh *command.RegisterUserHandler, lh *command.LoginHandler, jwtManager port.JWTManager) *AuthServer {
	return &AuthServer{
		registerHandler: rh,
		loginHandler:    lh,
		jwtManager:      jwtManager,
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
