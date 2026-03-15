package grpcserver

import (
	"context"
	pb "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

func (s *Server) SearchFlights(ctx context.Context, req *pb.SearchFlightsRequest) (*pb.SearchFlightsResponse, error) {
	criteria := domain.FlightSearchCriteria{
		Origin:        req.GetFromAirportCode(),
		Destination:   req.GetToAirportCode(),
		Date:          req.GetDepartureDate().AsTime(),
		NumPassengers: int(req.GetPassengerCount()),
		SeatClass:     seatClassToDomain(req.GetSeatClass()),
	}

	results, err := s.flight.Search(ctx, criteria)
	if err != nil {
		return nil, toGRPCError(err)
	}

	flights := make([]*pb.Flight, 0, len(results))
	for i := range results {
		flights = append(flights, flightDocToProto(&results[i]))
	}

	return &pb.SearchFlightsResponse{Flights: flights}, nil
}

func (s *Server) GetFlightDetails(ctx context.Context, req *pb.GetFlightDetailsRequest) (*pb.GetFlightDetailsResponse, error) {
	flight, err := s.flight.GetByID(ctx, req.GetFlightId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetFlightDetailsResponse{Flight: flightToProto(flight)}, nil
}

func (s *Server) GetFlightFares(ctx context.Context, req *pb.GetFlightFaresRequest) (*pb.GetFlightFaresResponse, error) {
	fares, err := s.flight.GetFares(ctx, req.GetFlightId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	result := make([]*pb.FareRule, 0, len(fares))
	for i := range fares {
		result = append(result, fareToProto(&fares[i]))
	}

	return &pb.GetFlightFaresResponse{Fares: result}, nil
}

func (s *Server) UpdateFlightStatus(ctx context.Context, req *pb.UpdateFlightStatusRequest) (*pb.UpdateFlightStatusResponse, error) {
	if err := s.flight.UpdateStatus(ctx, req.GetFlightId(), flightStatusToDomain(req.GetStatus())); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UpdateFlightStatusResponse{}, nil
}

func (s *Server) CreateFlight(ctx context.Context, req *pb.CreateFlightRequest) (*pb.CreateFlightResponse, error) {
	flight := &domain.Flight{
		FlightNumber:         req.GetFlightNumber(),
		AircraftID:           req.GetAircraftId(),
		AirlineID:            req.GetAirlineId(),
		DepartureAirportCode: req.GetDepartureAirportCode(),
		ArrivalAirportCode:   req.GetArrivalAirportCode(),
		DepartureTime:        req.GetDepartureTime().AsTime(),
		ArrivalTime:          req.GetArrivalTime().AsTime(),
	}

	fares := make([]domain.FlightFare, 0, len(req.GetFareRules()))
	for _, fare := range req.GetFareRules() {
		fares = append(fares, fareRuleToDomain(fare))
	}

	id, err := s.flight.Create(ctx, flight, fares)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateFlightResponse{FlightId: id}, nil
}
