package grpcserver

import (
	"context"
	pb "github.com/squ1ky/flyte/gen/proto/user"
	"github.com/squ1ky/flyte/internal/user/application/service"
	"github.com/squ1ky/flyte/internal/user/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Auth interface {
	Register(ctx context.Context, email, password, phone string) (int64, error)
	Login(ctx context.Context, email, password string) (userID int64, token string, err error)
	ValidateToken(ctx context.Context, token string) (*service.UserClaims, error)
	GetUser(ctx context.Context, userID int64) (*domain.User, error)
}

type Passenger interface {
	AddPassenger(ctx context.Context, p *domain.Passenger) (int64, error)
	GetPassengers(ctx context.Context, userID int64) ([]domain.Passenger, error)
	DeletePassenger(ctx context.Context, userID, passengerID int64) error
}

type Server struct {
	pb.UnimplementedUserServiceServer
	auth      Auth
	passenger Passenger
}

func NewServer(auth Auth, passenger Passenger) *Server {
	return &Server{
		auth:      auth,
		passenger: passenger,
	}
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetPhoneNumber())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.RegisterResponse{UserId: userID}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	userID, token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.LoginResponse{
		Token:  token,
		UserId: userID,
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := s.auth.ValidateToken(ctx, req.GetToken())
	if err != nil {
		if domain.IsValidationError(err) {
			return nil, toGRPCError(err)
		}
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		UserId: claims.UserID,
		Role:   mapStringToRole(claims.Role),
		Valid:  true,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.auth.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetUserResponse{
		Id:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        mapStringToRole(user.Role),
		CreatedAt:   timestamppb.New(user.CreatedAt),
	}, nil
}

func (s *Server) AddPassenger(ctx context.Context, req *pb.AddPassengerRequest) (*pb.AddPassengerResponse, error) {
	passenger, err := mapPassengerFromProto(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	}

	id, err := s.passenger.AddPassenger(ctx, passenger)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.AddPassengerResponse{PassengerId: id}, nil
}

func (s *Server) GetPassengers(ctx context.Context, req *pb.GetPassengersRequest) (*pb.GetPassengersResponse, error) {
	passengers, err := s.passenger.GetPassengers(ctx, req.GetUserId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	protoPassengers := make([]*pb.Passenger, 0, len(passengers))
	for _, passenger := range passengers {
		protoPassengers = append(protoPassengers, mapPassengerToProto(&passenger))
	}

	return &pb.GetPassengersResponse{Passengers: protoPassengers}, nil
}

func (s *Server) DeletePassenger(ctx context.Context, req *pb.DeletePassengerRequest) (*pb.DeletePassengerResponse, error) {
	if err := s.passenger.DeletePassenger(ctx, req.GetUserId(), req.GetPassengerId()); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.DeletePassengerResponse{}, nil
}
