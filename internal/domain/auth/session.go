package auth

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrSessionExpired = errors.New("session is expired")
	ErrSessionRevoked = errors.New("session is revoked")
)

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FamilyID  uuid.UUID
	ClientIP  string
	UserAgent string
	IsRevoked bool
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewSession(userID, familyID uuid.UUID, clientIP, userAgent string, duriation time.Duration) *Session {
	now := time.Now().UTC()
	sessionId := uuid.New()

	return &Session{
		ID:        sessionId,
		UserID:    userID,
		FamilyID:  familyID,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		IsRevoked: false,
		ExpiresAt: now.Add(duriation),
		CreatedAt: now,
	}
}

func (s *Session) Rotate(ip, userAgent string, duration time.Duration) (*Session, error) {
	if err := s.CheckValidity(); err != nil {
		return nil, err
	}
	s.IsRevoked = true

	now := time.Now().UTC()
	return &Session{
		ID:        uuid.New(),
		UserID:    s.UserID,
		FamilyID:  s.FamilyID,
		ClientIP:  ip,
		UserAgent: userAgent,
		IsRevoked: false,
		ExpiresAt: now.Add(duration),
		CreatedAt: now,
	}, nil

}

func (s *Session) CheckValidity() error {
	if s.IsRevoked {
		return ErrSessionRevoked
	}
	if time.Now().UTC().After(s.ExpiresAt) {
		return ErrSessionExpired
	}
	return nil
}
