package validator

import (
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	minPasswordLen = 8
	maxEmailLen    = 254
	dateLayout     = "2006-01-02"
)

const (
	msgEmailRequired    = "email is required"
	msgInvalidEmail     = "invalid email format"
	msgEmailTooLong     = "email is too long"
	msgPasswordRequired = "password is required"
	msgPasswordTooShort = "password must be at least %d characters long"
	msgInvalidPhone     = "invalid phone number format (use E.164, e.g., +79001234567)"

	msgPassengerInfoRequired = "passenger info is required"
	msgFirstNameRequired     = "first name is required"
	msgLastNameRequired      = "last name is required"
	msgInvalidBirthDate      = "invalid birth date format (expected %s): %v"
	msgFutureBirthDate       = "birth date cannot be in the future"
	msgInvalidGender         = "gender must be 'M' or 'F'"
	msgDocNumberRequired     = "document number is required"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

func ValidateRegister(req *userv1.RegisterRequest) error {
	if err := validateEmail(req.GetEmail()); err != nil {
		return err
	}

	if len(req.GetPassword()) < minPasswordLen {
		return status.Errorf(codes.InvalidArgument, msgPasswordTooShort, minPasswordLen)
	}

	if req.GetPhoneNumber() != "" {
		if !phoneRegex.MatchString(req.GetPhoneNumber()) {
			return status.Error(codes.InvalidArgument, msgInvalidPhone)
		}
	}

	return nil
}

func ValidateLogin(req *userv1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, msgEmailRequired)
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, msgPasswordRequired)
	}
	return nil
}

func ValidatePassenger(info *userv1.Passenger) error {
	if info == nil {
		return status.Error(codes.InvalidArgument, msgPassengerInfoRequired)
	}

	if strings.TrimSpace(info.GetFirstName()) == "" {
		return status.Error(codes.InvalidArgument, msgFirstNameRequired)
	}
	if strings.TrimSpace(info.GetLastName()) == "" {
		return status.Error(codes.InvalidArgument, msgLastNameRequired)
	}

	birthDate, err := time.Parse(dateLayout, info.GetBirthDate())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, msgInvalidBirthDate, dateLayout, info.GetBirthDate())
	}
	if birthDate.After(time.Now()) {
		return status.Errorf(codes.InvalidArgument, msgFutureBirthDate)
	}

	gender := strings.ToUpper(info.GetGender())
	if gender != "M" && gender != "F" {
		return status.Errorf(codes.InvalidArgument, msgInvalidGender)
	}

	if info.GetDocumentNumber() == "" {
		return status.Error(codes.InvalidArgument, msgDocNumberRequired)
	}

	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, msgEmailRequired)
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(codes.InvalidArgument, msgInvalidEmail)
	}
	if utf8.RuneCountInString(email) > maxEmailLen {
		return status.Error(codes.InvalidArgument, msgEmailTooLong)
	}
	return nil
}
