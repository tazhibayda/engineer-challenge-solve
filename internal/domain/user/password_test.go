package user

import (
	"strings"
	"testing"
)

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr error
	}{
		{"Valid password", "ValidPass123!", nil},
		{"Too short", "Short1!", ErrPasswordTooShort},
		{"No letters", "1234567890", ErrPasswordNoLetter},
		{"No digits", "NoDigitsHere", ErrPasswordNoDigit},
		{"Too long", strings.Repeat("a", 70) + "123", ErrPasswordTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPassword(tt.raw)
			if err != tt.wantErr {
				t.Errorf("NewPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
