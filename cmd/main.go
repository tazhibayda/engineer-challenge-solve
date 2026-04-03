package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tazhibayda/OrbittoAuth/internal/application/command"
	"github.com/tazhibayda/OrbittoAuth/internal/infrastructure/security"
	transportGrpc "github.com/tazhibayda/OrbittoAuth/internal/infrastructure/transport/grpc"
	authv1 "github.com/tazhibayda/OrbittoAuth/pkg/api/auth/v1"

	authMock "github.com/tazhibayda/OrbittoAuth/internal/domain/auth/mock"
	userMock "github.com/tazhibayda/OrbittoAuth/internal/domain/user/mock"
	"go.uber.org/mock/gomock"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Info("Starting Auth Service...")

	jwtManager, err := security.NewRSAJWTManager()
	if err != nil {
		sugar.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	// TODO: Здесь будет подключение к PostgreSQL и Redis.
	ctrl := gomock.NewController(nil)
	userRepo := userMock.NewMockRepository(ctrl)
	sessionRepo := authMock.NewMockSessionRepository(ctrl)

	registerHandler := command.NewRegisterUserHandler(userRepo)
	loginHandler := command.NewLoginHandler(userRepo, sessionRepo)

	authInterceptor := transportGrpc.NewAuthInterceptor(jwtManager)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)

	authServer := transportGrpc.NewAuthServer(registerHandler, loginHandler, jwtManager)
	authv1.RegisterAuthServiceServer(grpcServer, authServer)

	reflection.Register(grpcServer)

	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			sugar.Fatalf("Failed to listen: %v", err)
		}
		sugar.Infof("gRPC server is running on port 50051")
		if err := grpcServer.Serve(listener); err != nil {
			sugar.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	go func() {
		ctx := context.Background()
		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithInsecure()} // Для локальной разработки

		err := authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
		if err != nil {
			sugar.Fatalf("Failed to register gateway: %v", err)
		}

		sugar.Infof("REST Gateway is running on port 8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			sugar.Fatalf("Failed to serve REST gateway: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sugar.Info("Shutting down servers...")
	grpcServer.GracefulStop()
	sugar.Info("Servers stopped gracefully")
}
