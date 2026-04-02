package user

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"unicode"
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong  = errors.New("password cannot exceed 72 characters")
	ErrPasswordNoLetter = errors.New("password must contain at least one letter")
	ErrPasswordNoDigit  = errors.New("password must contain at least one digit")
)

type Password struct {
	value string
}
type HashedPassword string

func NewPassword(raw string) (Password, error) {
	if len(raw) < 8 {
		return Password{}, ErrPasswordTooShort
	}
	if len(raw) > 72 {
		return Password{}, ErrPasswordTooLong
	}
	var hasLetter, hasDigit bool
	for _, char := range raw {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}
	if !hasLetter {
		return Password{}, ErrPasswordNoLetter
	}
	if !hasDigit {
		return Password{}, ErrPasswordNoDigit
	}
	return Password{value: raw}, nil
}
func (p Password) String() string {
	return p.value
}

func (p Password) Hash() (HashedPassword, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p.value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return HashedPassword(bytes), nil
}

func (h HashedPassword) Compare(raw Password) error {
	return bcrypt.CompareHashAndPassword([]byte(h), []byte(raw.value))
}
