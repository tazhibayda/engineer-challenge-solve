package user

import (
	"testing"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr error
	}{
		{"Valid email", "test@example.com", nil},
		{"Valid email with tags", "user+tag@domain.co.uk", nil},
		{"Empty string", "", ErrInvalidEmail},
		{"Missing domain", "test@", ErrInvalidEmail},
		{"Missing local part", "@example.com", ErrInvalidEmail},
		{"Plain text", "not-an-email", ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEmail(tt.address)
			if err != tt.wantErr {
				t.Errorf("NewEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
