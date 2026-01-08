package grpc

import (
	"context"
	"errors"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"github.com/squ1ky/flyte/internal/flight/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	ErrAirportsRequired = "airports are required"
	ErrFlightIDRequired = "flight_id is required"

	ErrFlightNotFound    = "flight not found"
	ErrSeatNotFound      = "seat not found"
	ErrSeatAlreadyBooked = "seat already booked"
)

type Server struct {
	flightv1.UnimplementedFlightServiceServer
	flightService *service.FlightService
}

func NewServer(flightService *service.FlightService) *Server {
	return &Server{
		flightService: flightService,
	}
}

func (s *Server) SearchFlights(ctx context.Context, req *flightv1.SearchFlightsRequest) (*flightv1.SearchFlightsResponse, error) {
	if req.FromAirport == "" || req.ToAirport == "" {
		return nil, status.Error(codes.InvalidArgument, ErrAirportsRequired)
	}

	flights, err := s.flightService.SearchFlights(
		ctx,
		req.FromAirport,
		req.ToAirport,
		req.Date.AsTime(),
		int(req.PassengerCount),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search failed: %v", err)
	}

	var pbFlights []*flightv1.Flight
	for _, f := range flights {
		pbFlights = append(pbFlights, mapFlightToProto(&f))
	}

	return &flightv1.SearchFlightsResponse{Flights: pbFlights}, nil
}

func (s *Server) CreateFlight(ctx context.Context, req *flightv1.CreateFlightRequest) (*flightv1.CreateFlightResponse, error) {
	flight := &domain.Flight{
		FlightNumber:     req.FlightNumber,
		DepartureAirport: req.DepartureAirport,
		ArrivalAirport:   req.ArrivalAirport,
		DepartureTime:    req.DepartureTime.AsTime(),
		ArrivalTime:      req.ArrivalTime.AsTime(),
		Price:            req.Price,
		TotalSeats:       int(req.TotalSeats),
		Status:           "scheduled",
	}

	id, err := s.flightService.CreateFlight(ctx, flight)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create flight: %v", err)
	}

	return &flightv1.CreateFlightResponse{FlightId: id}, nil
}

func (s *Server) GetFlightDetails(ctx context.Context, req *flightv1.GetFlightDetailsRequest) (*flightv1.GetFlightDetailsResponse, error) {
	if req.FlightId == 0 {
		return nil, status.Error(codes.InvalidArgument, ErrFlightIDRequired)
	}

	flight, err := s.flightService.GetFlightDetails(ctx, req.FlightId)
	if err != nil {
		if errors.Is(err, domain.ErrFlightNotFound) {
			return nil, status.Error(codes.NotFound, ErrFlightNotFound)
		}
		return nil, status.Errorf(codes.Internal, "failed to get flight details: %v", err)
	}

	return &flightv1.GetFlightDetailsResponse{Flight: mapFlightToProto(flight)}, nil
}

func (s *Server) GetFlightSeats(ctx context.Context, req *flightv1.GetFlightSeatsRequest) (*flightv1.GetFlightSeatsResponse, error) {
	if req.FlightId == 0 {
		return nil, status.Error(codes.InvalidArgument, ErrFlightIDRequired)
	}

	seats, err := s.flightService.GetFlightSeats(ctx, req.FlightId)
	if err != nil {
		if errors.Is(err, domain.ErrFlightNotFound) {
			return nil, status.Error(codes.NotFound, ErrFlightNotFound)
		}
		return nil, status.Errorf(codes.Internal, "failed to get seats: %v", err)
	}

	var pbSeats []*flightv1.Seat
	for _, seat := range seats {
		pbSeats = append(pbSeats, &flightv1.Seat{
			Id:              seat.ID,
			SeatNumber:      seat.SeatNumber,
			IsBooked:        seat.IsBooked,
			PriceMultiplier: seat.PriceMultiplier,
		})
	}

	return &flightv1.GetFlightSeatsResponse{Seats: pbSeats}, nil
}

func (s *Server) ListAirports(ctx context.Context, req *flightv1.ListAirportsRequest) (*flightv1.ListAirportsResponse, error) {
	airports, err := s.flightService.GetAirports(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list airports: %v", err)
	}

	var pbAirports []*flightv1.Airport
	for _, a := range airports {
		pbAirports = append(pbAirports, &flightv1.Airport{
			Code:    a.Code,
			Name:    a.Name,
			City:    a.City,
			Country: a.Country,
		})
	}

	return &flightv1.ListAirportsResponse{Airports: pbAirports}, nil
}

func (s *Server) ReserveSeat(ctx context.Context, req *flightv1.ReserveSeatRequest) (*flightv1.ReserveSeatResponse, error) {
	if req.FlightId == 0 {
		return nil, status.Error(codes.InvalidArgument, ErrFlightIDRequired)
	}

	seatID, err := s.flightService.ReserveSeat(ctx, req.FlightId, req.SeatNumber)
	if err != nil {
		if errors.Is(err, domain.ErrSeatAlreadyBooked) {
			return nil, status.Error(codes.AlreadyExists, ErrSeatAlreadyBooked)
		}
		if errors.Is(err, domain.ErrSeatNotFound) {
			return nil, status.Error(codes.NotFound, ErrSeatNotFound)
		}
		return nil, status.Errorf(codes.Internal, "failed to reserve seat: %v", err)
	}

	return &flightv1.ReserveSeatResponse{Success: true, SeatId: seatID}, nil
}

func (s *Server) ReleaseSeat(ctx context.Context, req *flightv1.ReleaseSeatRequest) (*flightv1.ReleaseSeatResponse, error) {
	if err := s.flightService.ReleaseSeat(ctx, req.FlightId, req.SeatNumber); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to release seat: %v", err)
	}
	return &flightv1.ReleaseSeatResponse{Success: true}, nil
}

func mapFlightToProto(f *domain.Flight) *flightv1.Flight {
	return &flightv1.Flight{
		Id:               f.ID,
		FlightNumber:     f.FlightNumber,
		DepartureAirport: f.DepartureAirport,
		ArrivalAirport:   f.ArrivalAirport,
		DepartureTime:    timestamppb.New(f.DepartureTime),
		ArrivalTime:      timestamppb.New(f.ArrivalTime),
		Price:            f.Price,
		Status:           f.Status,
		TotalSeats:       int32(f.TotalSeats),
		AvailableSeats:   int32(f.AvailableSeats),
	}
}
