package grpc

import (
	"context"
	bookingv1 "github.com/squ1ky/flyte/gen/go/booking"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

type Server struct {
	bookingv1.UnimplementedBookingServiceServer

	svc     *service.BookingService
	timeout time.Duration
}

func NewServer(svc *service.BookingService, timeout time.Duration) *Server {
	return &Server{
		svc:     svc,
		timeout: timeout,
	}
}

func (s *Server) Register(gRPCServer *grpc.Server) {
	bookingv1.RegisterBookingServiceServer(gRPCServer, s)
}

func (s *Server) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.CreateBookingResponse, error) {
	if err := validateCreateBookingRequest(req); err != nil {
		return nil, err
	}

	dto := service.CreateBookingDTO{
		UserID:            req.UserId,
		FlightID:          req.FlightId,
		SeatNumber:        strings.TrimSpace(req.SeatNumber),
		PriceCents:        req.PriceCents,
		Currency:          strings.TrimSpace(req.Currency),
		PassengerName:     strings.TrimSpace(req.PassengerName),
		PassengerPassport: strings.TrimSpace(req.PassengerPassport),
	}
	if dto.Currency == "" {
		dto.Currency = "RUB"
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	id, err := s.svc.CreateBooking(ctx, dto)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create booking: %v", err)
	}

	return &bookingv1.CreateBookingResponse{BookingId: id}, nil
}

func (s *Server) GetBooking(ctx context.Context, req *bookingv1.GetBookingRequest) (*bookingv1.GetBookingResponse, error) {
	if err := validateGetBookingRequest(req); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	b, err := s.svc.GetBooking(ctx, strings.TrimSpace(req.BookingId))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "booking not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get booking: %v", err)
	}

	return &bookingv1.GetBookingResponse{Booking: mapBookingToProto(b)}, nil
}

func (s *Server) ListBookings(ctx context.Context, req *bookingv1.ListBookingsRequest) (*bookingv1.ListBookingsResponse, error) {
	if err := validateListBookingsRequest(req); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	list, err := s.svc.ListBookings(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list bookings: %v", err)
	}

	out := make([]*bookingv1.Booking, 0, len(list))
	for i := range list {
		out = append(out, mapBookingToProto(&list[i]))
	}

	return &bookingv1.ListBookingsResponse{Bookings: out}, nil
}

func (s *Server) CancelBooking(ctx context.Context, req *bookingv1.CancelBookingRequest) (*bookingv1.CancelBookingResponse, error) {
	if err := validateCancelBookingRequest(req); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.svc.CancelBooking(ctx, req.BookingId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cancel booking: %v", err)
	}

	return &bookingv1.CancelBookingResponse{}, nil
}

func mapBookingToProto(b *domain.Booking) *bookingv1.Booking {
	if b == nil {
		return nil
	}
	return &bookingv1.Booking{
		Id:                b.ID,
		UserId:            b.UserID,
		FlightId:          b.FlightID,
		SeatNumber:        b.SeatNumber,
		PassengerName:     b.PassengerName,
		PassengerPassport: b.PassengerPassport,
		Status:            string(b.Status),
		PriceCents:        b.PriceCents,
		Currency:          b.Currency,
		CreatedAt:         timestamppb.New(b.CreatedAt),
		UpdatedAt:         timestamppb.New(b.UpdatedAt),
	}
}
