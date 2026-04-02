package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSession_Rotate(t *testing.T) {
	userID := uuid.New()

	t.Run("Successful rotation", func(t *testing.T) {
		session := NewSession(userID, "127.0.0.1", "TestAgent", time.Hour)

		newSession, err := session.Rotate("127.0.0.1", "TestAgent", time.Hour)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !session.IsRevoked {
			t.Errorf("old session should be revoked")
		}
		if newSession.FamilyID != session.FamilyID {
			t.Errorf("family ID should remain the same across rotations")
		}
	})

	t.Run("Cannot rotate revoked session", func(t *testing.T) {
		session := NewSession(userID, "127.0.0.1", "TestAgent", time.Hour)
		session.IsRevoked = true

		_, err := session.Rotate("127.0.0.1", "TestAgent", time.Hour)
		if err != ErrSessionRevoked {
			t.Errorf("expected ErrSessionRevoked, got %v", err)
		}
	})
}
