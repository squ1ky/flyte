package grpc

import (
	"context"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/user/domain"
	"github.com/squ1ky/flyte/internal/user/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
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
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	userID, err := s.auth.Register(ctx, req.Email, req.Password, req.PhoneNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &userv1.RegisterResponse{UserId: userID}, nil
}

func (s *Server) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	token, err := s.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	return &userv1.LoginResponse{Token: token}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *userv1.ValidateTokenRequest) (*userv1.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
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
		return nil, status.Error(codes.NotFound, "user not found")
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
	info := req.Info
	if info == nil {
		return nil, status.Error(codes.InvalidArgument, "passenger info is required")
	}

	birthDate, err := time.Parse("2006-01-02", info.BirthDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid birth date format: %v", err)
	}

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
