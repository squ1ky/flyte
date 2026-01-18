package grpc

import (
	"context"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateAircraft(ctx context.Context, req *flightv1.CreateAircraftRequest) (*flightv1.CreateAircraftResponse, error) {
	if err := validateCreateAircraftRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	id, err := s.aircraftService.CreateAircraft(ctx, req.Model, int(req.TotalSeats))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create aircraft: %v", err)
	}

	return &flightv1.CreateAircraftResponse{AircraftId: id}, nil
}

func (s *Server) AddAircraftSeats(ctx context.Context, req *flightv1.AddAircraftSeatsRequest) (*flightv1.AddAircraftSeatsResponse, error) {
	if err := validateAddAircraftSeatsRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	domainSeats := make([]domain.AircraftSeat, 0, len(req.Seats))
	for _, seat := range req.Seats {
		domainSeats = append(domainSeats, domain.AircraftSeat{
			SeatNumber:      seat.SeatNumber,
			SeatClass:       domain.SeatClass(seat.SeatClass),
			PriceMultiplier: seat.PriceMultiplier,
		})
	}

	if err := s.aircraftService.ConfigureSeats(ctx, req.AircraftId, domainSeats); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add seats: %v", err)
	}

	return &flightv1.AddAircraftSeatsResponse{Success: true}, nil
}

func (s *Server) ListAircrafts(ctx context.Context, req *flightv1.ListAircraftsRequest) (*flightv1.ListAircraftsResponse, error) {
	list, err := s.aircraftService.GetAircrafts(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list aircrafts: %v", err)
	}

	pbList := make([]*flightv1.Aircraft, 0, len(list))
	for _, a := range list {
		pbList = append(pbList, &flightv1.Aircraft{
			Id:         a.ID,
			Model:      a.Model,
			TotalSeats: int32(a.TotalSeats),
		})
	}

	return &flightv1.ListAircraftsResponse{Aircrafts: pbList}, nil
}
