package security

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tazhibayda/OrbittoAuth/internal/application/port"
)

type userClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type RSAJWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewRSAJWTManager() (*RSAJWTManager, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("cannot generate RSA key: %w", err)
	}

	return &RSAJWTManager{
		privateKey: privKey,
		publicKey:  &privKey.PublicKey,
	}, nil
}

func (m *RSAJWTManager) Generate(claims port.AuthClaims, duration time.Duration) (string, error) {
	now := time.Now().UTC()

	c := userClaims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth-service",
			Subject:   claims.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	token.Header["kid"] = "main-key-v1"

	return token.SignedString(m.privateKey)
}

func (m *RSAJWTManager) Verify(accessToken string) (*port.AuthClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&userClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return m.publicKey, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*userClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &port.AuthClaims{
		UserID: claims.UserID,
	}, nil
}
