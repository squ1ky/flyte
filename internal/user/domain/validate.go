package domain

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinPasswordLen = 8
	MaxEmailLen    = 254
)

const (
	emailField       = "email"
	passwordField    = "password"
	phoneNumberField = "phone_number"

	msgRequired = "is required"
	msgTooLong  = "is too long"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

func ValidateRegister(email, password, phoneNumber string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if len(password) < MinPasswordLen {
		return NewValidationError(passwordField,
			fmt.Sprintf("must be at least %d characters long", MinPasswordLen),
		)
	}

	if phoneNumber != "" {
		if !phoneRegex.MatchString(phoneNumber) {
			return NewValidationError(phoneNumberField, "invalid format (use E.164, e.g. +79001234567)")
		}
	}

	return nil
}

func ValidateLogin(email, password string) error {
	if email == "" {
		return NewValidationError(emailField, msgRequired)
	}
	if password == "" {
		return NewValidationError(passwordField, msgRequired)
	}
	return nil
}

func ValidateEmail(email string) error {
	if email == "" {
		return NewValidationError(emailField, msgRequired)
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return NewValidationError(emailField, "is not a valid email address")
	}
	if utf8.RuneCountInString(email) > MaxEmailLen {
		return NewValidationError(emailField, msgTooLong)
	}
	return nil
}

func ValidatePassenger(p *Passenger) error {
	if p == nil {
		return NewValidationError("passenger", "info"+msgRequired)
	}

	if strings.TrimSpace(p.FirstName) == "" {
		return NewValidationError("first_name", msgRequired)
	}
	if strings.TrimSpace(p.LastName) == "" {
		return NewValidationError("last_name", msgRequired)
	}

	if p.BirthDate.IsZero() {
		return NewValidationError("birth_date", msgRequired)
	}
	if p.BirthDate.After(time.Now()) {
		return NewValidationError("birth_date", "cannot be in the future")
	}

	if p.Gender != GenderMale && p.Gender != GenderFemale {
		return NewValidationError("gender",
			fmt.Sprintf("must be '%s' or '%s'", GenderMale, GenderFemale),
		)
	}

	if strings.TrimSpace(p.DocumentNumber) == "" {
		return NewValidationError("document_number", msgRequired)
	}

	if len(p.Citizenship) != 3 {
		return NewValidationError("citizenship", "must be a 3-letter ISO code")
	}

	return nil
}
