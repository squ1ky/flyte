package grpcserver

import (
	pb "github.com/squ1ky/flyte/gen/proto/user"
	"github.com/squ1ky/flyte/internal/user/domain"
	"time"
)

func mapPassengerFromProto(req *pb.AddPassengerRequest) (*domain.Passenger, error) {
	info := req.GetInfo()
	if info == nil {
		return &domain.Passenger{UserID: req.GetUserId()}, nil
	}

	var birthDate time.Time
	if info.GetBirthDate() != "" {
		var err error
		birthDate, err = time.Parse("2006-01-02", info.GetBirthDate())
		if err != nil {
			return nil, err
		}
	}

	return &domain.Passenger{
		UserID:         req.GetUserId(),
		FirstName:      info.GetFirstName(),
		LastName:       info.GetLastName(),
		MiddleName:     info.GetMiddleName(),
		BirthDate:      birthDate,
		Gender:         mapGenderToString(info.GetGender()),
		DocumentNumber: info.GetDocumentNumber(),
		DocumentType:   info.GetDocumentType(),
		Citizenship:    info.GetCitizenship(),
	}, nil
}

func mapPassengerToProto(p *domain.Passenger) *pb.Passenger {
	return &pb.Passenger{
		Id:             p.ID,
		UserId:         p.UserID,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		MiddleName:     p.MiddleName,
		BirthDate:      p.BirthDate.Format("2006-01-02"),
		Gender:         mapStringToGender(p.Gender),
		DocumentNumber: p.DocumentNumber,
		DocumentType:   p.DocumentType,
		Citizenship:    p.Citizenship,
	}
}

func mapStringToRole(r string) pb.Role {
	switch r {
	case domain.RoleUser:
		return pb.Role_ROLE_USER
	case domain.RoleAdmin:
		return pb.Role_ROLE_ADMIN
	default:
		return pb.Role_ROLE_UNSPECIFIED
	}
}

func mapStringToGender(g string) pb.Gender {
	switch g {
	case domain.GenderMale:
		return pb.Gender_GENDER_MALE
	case domain.GenderFemale:
		return pb.Gender_GENDER_FEMALE
	default:
		return pb.Gender_GENDER_UNSPECIFIED
	}
}

func mapGenderToString(g pb.Gender) string {
	switch g {
	case pb.Gender_GENDER_MALE:
		return domain.GenderMale
	case pb.Gender_GENDER_FEMALE:
		return domain.GenderFemale
	default:
		return "unknown"
	}
}
