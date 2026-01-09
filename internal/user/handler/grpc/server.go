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
	"time"
)

const (
	ErrTokenRequired = "token is required"

	ErrInvalidCredentials = "invalid email or password"
	ErrUserAlreadyExists  = "user with this email already exists"
	ErrUserNotFound       = "user not found"
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

	userID, err := s.auth.Register(ctx, req.Email, req.Password, req.PhoneNumber)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, ErrUserAlreadyExists)
		}
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &userv1.RegisterResponse{UserId: userID}, nil
}

func (s *Server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if err := validator.ValidateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, ErrInvalidCredentials)
	}

	return &userv1.LoginResponse{Token: token}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *userv1.ValidateTokenRequest) (*userv1.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, ErrTokenRequired)
	}

	claims, err := s.auth.ValidateToken(ctx, req.Token)
	if err != nil {
		return &userv1.ValidateTokenResponse{Valid: false}, nil
	}

	return &userv1.ValidateTokenResponse{
		UserId: claims.UserID,
		Role:   claims.Role,
		Valid:  true,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := s.auth.GetUser(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, ErrUserNotFound)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &userv1.GetUserResponse{
		Id:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Server) AddPassenger(ctx context.Context, req *userv1.AddPassengerRequest) (*userv1.AddPassengerResponse, error) {
	if err := validator.ValidatePassenger(req.Info); err != nil {
		return nil, err
	}

	info := req.Info
	birthDate, _ := time.Parse("2006-01-02", info.BirthDate)

	passenger := &domain.Passenger{
		UserID:         req.UserId,
		FirstName:      info.FirstName,
		LastName:       info.LastName,
		MiddleName:     info.MiddleName,
		BirthDate:      birthDate,
		Gender:         info.Gender,
		DocumentNumber: info.DocumentNumber,
		DocumentType:   info.DocumentType,
		Citizenship:    info.Citizenship,
	}

	id, err := s.passenger.AddPassenger(ctx, passenger)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add passenger: %v", err)
	}

	return &userv1.AddPassengerResponse{PassengerId: id}, nil
}

func (s *Server) GetPassengers(ctx context.Context, req *userv1.GetPassengersRequest) (*userv1.GetPassengersResponse, error) {
	passengers, err := s.passenger.GetPassengers(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get passengers: %v", err)
	}

	var protoPassengers []*userv1.Passenger
	for _, p := range passengers {
		protoPassengers = append(protoPassengers, &userv1.Passenger{
			Id:             p.ID,
			FirstName:      p.FirstName,
			LastName:       p.LastName,
			MiddleName:     p.MiddleName,
			BirthDate:      p.BirthDate.Format("2006-01-02"),
			Gender:         p.Gender,
			DocumentNumber: p.DocumentNumber,
			DocumentType:   p.DocumentType,
			Citizenship:    p.Citizenship,
		})
	}

	return &userv1.GetPassengersResponse{Passengers: protoPassengers}, nil
}
