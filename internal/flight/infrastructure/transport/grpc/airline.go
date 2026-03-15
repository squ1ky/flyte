package grpcserver

import (
	"context"
	pb "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

func (s *Server) CreateAirline(ctx context.Context, req *pb.CreateAirlineRequest) (*pb.CreateAirlineResponse, error) {
	airline := &domain.Airline{
		IATACode: req.GetIataCode(),
		Name:     req.GetName(),
		Country:  req.GetCountry(),
		LogoURL:  req.GetLogoUrl(),
	}

	id, err := s.airline.Create(ctx, airline)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateAirlineResponse{AirlineId: id}, nil
}

func (s *Server) GetAirline(ctx context.Context, req *pb.GetAirlineRequest) (*pb.GetAirlineResponse, error) {
	airline, err := s.airline.GetByID(ctx, req.GetAirlineId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetAirlineResponse{Airline: airlineToProto(airline)}, nil
}

func (s *Server) ListAirlines(ctx context.Context, req *pb.ListAirlinesRequest) (*pb.ListAirlinesResponse, error) {
	airlines, err := s.airline.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, toGRPCError(err)
	}

	result := make([]*pb.Airline, 0, len(airlines))
	for i := range airlines {
		result = append(result, airlineToProto(&airlines[i]))
	}

	return &pb.ListAirlinesResponse{Airlines: result}, nil
}
