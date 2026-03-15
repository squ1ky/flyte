package grpcserver

import (
	"context"
	pb "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

func (s *Server) ReserveSeats(ctx context.Context, req *pb.ReserveSeatsRequest) (*pb.ReserveSeatsResponse, error) {
	params := domain.SeatReservationParams{
		FlightID:    req.GetFlightId(),
		SeatNumbers: req.GetSeatNumbers(),
		BookingID:   req.GetBookingId(),
	}

	if err := s.reservation.ReserveSeats(ctx, params); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ReserveSeatsResponse{}, nil
}

func (s *Server) ConfirmReservation(ctx context.Context, req *pb.ConfirmReservationRequest) (*pb.ConfirmReservationResponse, error) {
	params := domain.SeatReservationParams{
		FlightID:    req.GetFlightId(),
		SeatNumbers: req.GetSeatNumbers(),
		BookingID:   req.GetBookingId(),
	}

	if err := s.reservation.ConfirmReservation(ctx, params); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ConfirmReservationResponse{}, nil
}

func (s *Server) CancelReservation(ctx context.Context, req *pb.CancelReservationRequest) (*pb.CancelReservationResponse, error) {
	if err := s.reservation.CancelReservation(ctx, req.GetBookingId()); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CancelReservationResponse{}, nil
}

func (s *Server) GetReservedSeats(ctx context.Context, req *pb.GetReservedSeatsRequest) (*pb.GetReservedSeatsResponse, error) {
	seats, err := s.reservation.GetReservedSeats(ctx, req.GetFlightId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	result := make([]*pb.SeatReservation, 0, len(seats))
	for i := range seats {
		result = append(result, seatReservationToProto(&seats[i]))
	}

	return &pb.GetReservedSeatsResponse{Seats: result}, nil
}
