package grpc

import (
	"context"
	"errors"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/user/domain"
	"github.com/squ1ky/flyte/internal/user/service"
	"github.com/squ1ky/flyte/internal/user/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	MsgTokenRequired      = "token is required"
	MsgInvalidCredentials = "invalid email or password"
	MsgUserAlreadyExists  = "user with this email already exists"
	MsgUserNotFound       = "user not found"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	auth      service.Auth
	passenger service.Passenger
}

func NewServer(auth service.Auth, passenger service.Passenger) *Server {
	return &Server{
		auth:      auth,
		passenger: passenger,
	}
}

func (s *Server) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	if err := validator.ValidateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetPhoneNumber())
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, MsgUserAlreadyExists)
		}
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &userv1.RegisterResponse{UserId: userID}, nil
}

func (s *Server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if err := validator.ValidateLogin(req); err != nil {
		return nil, err
	}

	userID, token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, MsgInvalidCredentials)
	}

	return &userv1.LoginResponse{
		Token:  token,
		UserId: userID,
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *userv1.ValidateTokenRequest) (*userv1.ValidateTokenResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, MsgTokenRequired)
	}

	claims, err := s.auth.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return &userv1.ValidateTokenResponse{Valid: false}, nil
	}

	return &userv1.ValidateTokenResponse{
		UserId: claims.UserID,
		Role:   mapStringToRole(claims.Role),
		Valid:  true,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := s.auth.GetUser(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, MsgUserNotFound)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &userv1.GetUserResponse{
		Id:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        mapStringToRole(user.Role),
		CreatedAt:   timestamppb.New(user.CreatedAt),
	}, nil
}

func (s *Server) AddPassenger(ctx context.Context, req *userv1.AddPassengerRequest) (*userv1.AddPassengerResponse, error) {
	if err := validator.ValidatePassengerInfo(req.GetInfo()); err != nil {
		return nil, err
	}

	info := req.GetInfo()
	birthDate, _ := time.Parse("2006-01-02", info.GetBirthDate())

	passenger := &domain.Passenger{
		UserID:         req.GetUserId(),
		FirstName:      info.GetFirstName(),
		LastName:       info.GetLastName(),
		MiddleName:     info.GetMiddleName(),
		BirthDate:      birthDate,
		Gender:         mapGenderToString(info.GetGender()),
		DocumentNumber: info.GetDocumentNumber(),
		DocumentType:   info.GetDocumentType(),
		Citizenship:    info.GetCitizenship(),
	}

	id, err := s.passenger.AddPassenger(ctx, passenger)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add passenger: %v", err)
	}

	return &userv1.AddPassengerResponse{PassengerId: id}, nil
}

func (s *Server) GetPassengers(ctx context.Context, req *userv1.GetPassengersRequest) (*userv1.GetPassengersResponse, error) {
	passengers, err := s.passenger.GetPassengers(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get passengers: %v", err)
	}

	var protoPassengers []*userv1.Passenger
	for _, p := range passengers {
		protoPassengers = append(protoPassengers, &userv1.Passenger{
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
		})
	}

	return &userv1.GetPassengersResponse{Passengers: protoPassengers}, nil
}

func (s *Server) DeletePassenger(ctx context.Context, req *userv1.DeletePassengerRequest) (*userv1.DeletePassengerResponse, error) {
	if err := s.passenger.DeletePassenger(ctx, req.GetUserId(), req.GetPassengerId()); err != nil {
		if errors.Is(err, domain.ErrPassengerNotFound) {
			return nil, status.Error(codes.NotFound, "passenger not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete passenger: %v", err)
	}

	return &userv1.DeletePassengerResponse{}, nil
}

func mapStringToRole(r string) userv1.Role {
	switch r {
	case domain.RoleUser:
		return userv1.Role_ROLE_USER
	case domain.RoleAdmin:
		return userv1.Role_ROLE_ADMIN
	default:
		return userv1.Role_ROLE_UNSPECIFIED
	}
}

func mapStringToGender(g string) userv1.Gender {
	switch g {
	case domain.GenderMale:
		return userv1.Gender_GENDER_MALE
	case domain.GenderFemale:
		return userv1.Gender_GENDER_FEMALE
	default:
		return userv1.Gender_GENDER_UNSPECIFIED
	}
}

func mapGenderToString(g userv1.Gender) string {
	switch g {
	case userv1.Gender_GENDER_MALE:
		return domain.GenderMale
	case userv1.Gender_GENDER_FEMALE:
		return domain.GenderFemale
	default:
		return "unknown"
	}
}
