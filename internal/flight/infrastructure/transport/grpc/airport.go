package grpcserver

import (
	"context"
	pb "github.com/squ1ky/flyte/gen/proto/flight"
)

func (s *Server) SearchAirports(ctx context.Context, req *pb.SearchAirportsRequest) (*pb.SearchAirportsResponse, error) {
	airports, err := s.airport.Search(ctx, req.GetQuery())
	if err != nil {
		return nil, toGRPCError(err)
	}

	result := make([]*pb.Airport, 0, len(airports))
	for i := range airports {
		result = append(result, airportToProto(&airports[i]))
	}

	return &pb.SearchAirportsResponse{Airports: result}, nil
}

func (s *Server) GetAirport(ctx context.Context, req *pb.GetAirportRequest) (*pb.GetAirportResponse, error) {
	airport, err := s.airport.GetByCode(ctx, req.GetCode())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetAirportResponse{Airport: airportToProto(airport)}, nil
}
