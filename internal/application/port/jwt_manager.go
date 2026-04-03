package port

import (
	"time"
)

type AuthClaims struct {
	UserID string
	//Role  string
}

type JWTManager interface {
	Generate(claims AuthClaims, duration time.Duration) (string, error)
	Verify(accessToken string) (*AuthClaims, error)
}
