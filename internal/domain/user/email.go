package user

import (
	"errors"
	"net/mail"
)

var ErrInvalidEmail = errors.New("invalid email address")

type Email struct {
	value string
}

func NewEmail(address string) (Email, error) {
	parsed, err := mail.ParseAddress(address)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: parsed.Address}, nil
}

func (e Email) String() string {
	return e.value
}
