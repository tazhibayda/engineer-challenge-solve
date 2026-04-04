package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	authv1 "github.com/tazhibayda/OrbittoAuth/pkg/api/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestAuthFlow_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := authv1.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	email := "e2e@test.com"
	pass := "StrongPass123!"

	regResp, err := client.Register(ctx, &authv1.RegisterRequest{
		Email:     email,
		Password:  pass,
		FirstName: "E2E",
		LastName:  "Tester",
	})
	assert.NoError(t, err)
	assert.Contains(t, regResp.Message, "successfully")

	loginResp, err := client.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: pass,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.SessionId)
}
