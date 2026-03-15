package grpcserver

import (
	"context"
	pb "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

func (s *Server) CreateAircraft(ctx context.Context, req *pb.CreateAircraftRequest) (*pb.CreateAircraftResponse, error) {
	aircraft := &domain.Aircraft{
		Model:      req.GetModel(),
		TotalSeats: int(req.GetTotalSeats()),
	}

	id, err := s.aircraft.Create(ctx, aircraft)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateAircraftResponse{AircraftId: int32(id)}, nil
}

func (s *Server) AddAircraftSeats(ctx context.Context, req *pb.AddAircraftSeatsRequest) (*pb.AddAircraftSeatsResponse, error) {
	seats := make([]domain.AircraftSeat, 0, len(req.GetSeats()))
	for _, seat := range req.GetSeats() {
		seats = append(seats, seatTemplateToDomain(seat))
	}

	if err := s.aircraft.AddSeats(ctx, req.GetAircraftId(), seats); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.AddAircraftSeatsResponse{SeatsAdded: int32(len(seats))}, nil
}

func (s *Server) GetAircraftSeats(ctx context.Context, req *pb.GetAircraftSeatsRequest) (*pb.GetAircraftSeatsResponse, error) {
	seats, err := s.aircraft.GetSeats(ctx, req.GetAircraftId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	result := make([]*pb.SeatTemplate, 0, len(seats))
	for i := range seats {
		result = append(result, seatTemplateToProto(&seats[i]))
	}

	return &pb.GetAircraftSeatsResponse{Seats: result}, nil
}
