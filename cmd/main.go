package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tazhibayda/OrbittoAuth/internal/infrastructure/db/postgres"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tazhibayda/OrbittoAuth/internal/application/command"
	redisAdapter "github.com/tazhibayda/OrbittoAuth/internal/infrastructure/db/redis"
	"github.com/tazhibayda/OrbittoAuth/internal/infrastructure/security"
	transportGrpc "github.com/tazhibayda/OrbittoAuth/internal/infrastructure/transport/grpc"
	authv1 "github.com/tazhibayda/OrbittoAuth/pkg/api/auth/v1"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Info("Starting Auth Service...")

	ctx := context.Background()
	// TODO: Вынести конфигурацию в отдельный файл .env или использовать флаги командной строки
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:1234@192.168.32.123:5432/auth_db?sslmode=disable"
	}
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		sugar.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()
	if err := dbPool.Ping(ctx); err != nil {
		sugar.Fatalf("Database ping failed: %v", err)
	}
	sugar.Info("Connected to PostgreSQL")

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "192.168.32.123:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "192.168.32.123:6379", // пока что для локальной разработки - wsl redis
		Password: "1234",
	})
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		sugar.Fatalf("Redis ping failed: %v", err)
	}
	sugar.Info("Connected to Redis")

	jwtManager, err := security.NewRSAJWTManager()
	if err != nil {
		sugar.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	userRepo := postgres.NewUserRepository(dbPool)
	sessionRepo := redisAdapter.NewSessionRepository(redisClient)

	registerHandler := command.NewRegisterUserHandler(userRepo)
	loginHandler := command.NewLoginHandler(userRepo, sessionRepo)
	resetRepo := redisAdapter.NewResetRepository(redisClient)

	requestResetHandler := command.NewRequestPasswordResetHandler(userRepo, resetRepo, sugar)
	resetPasswordHandler := command.NewResetPasswordHandler(userRepo, resetRepo, sessionRepo)

	authInterceptor := transportGrpc.NewAuthInterceptor(jwtManager)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)

	authServer := transportGrpc.NewAuthServer(registerHandler, loginHandler, requestResetHandler, resetPasswordHandler, jwtManager)
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
	time.Sleep(1 * time.Second)
	sugar.Info("Servers stopped gracefully")
}
