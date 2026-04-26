package user

import (
	userv1 "github.com/squ1ky/flyte/gen/proto/user"
	"strings"
)

const (
	genderMaleString   = "male"
	genderFemaleString = "female"
)

func genderToProto(genderStr string) userv1.Gender {
	switch strings.ToLower(genderStr) {
	case genderMaleString:
		return userv1.Gender_GENDER_MALE
	case genderFemaleString:
		return userv1.Gender_GENDER_FEMALE
	default:
		return userv1.Gender_GENDER_UNSPECIFIED
	}
}

func genderFromProto(gender userv1.Gender) string {
	switch gender {
	case userv1.Gender_GENDER_MALE:
		return genderMaleString
	case userv1.Gender_GENDER_FEMALE:
		return genderFemaleString
	default:
		return ""
	}
}

func mapProtoToPassenger(p *userv1.Passenger) PassengerResponse {
	return PassengerResponse{
		ID:             p.Id,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		MiddleName:     p.MiddleName,
		BirthDate:      p.BirthDate,
		Gender:         genderFromProto(p.Gender),
		DocumentNumber: p.DocumentNumber,
		DocumentType:   p.DocumentType,
		Citizenship:    p.Citizenship,
	}
}
