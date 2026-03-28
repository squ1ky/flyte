package grpcserver

import (
	"context"
	pb "github.com/squ1ky/flyte/gen/proto/booking"
	"github.com/squ1ky/flyte/internal/booking/application/service"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
)

type Handler struct {
	pb.UnimplementedBookingServiceServer

	saga          *service.BookingSaga
	queryService  *service.BookingQueryService
	cancelService *service.CancelService
	logger        *slog.Logger
}

func NewHandler(
	saga *service.BookingSaga,
	queryService *service.BookingQueryService,
	cancelService *service.CancelService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		saga:          saga,
		queryService:  queryService,
		cancelService: cancelService,
		logger:        logger,
	}
}

func (h *Handler) CreateBooking(ctx context.Context, req *pb.CreateBookingRequest) (*pb.CreateBookingResponse, error) {
	passengers := make([]service.PassengerInput, len(req.GetPassengers()))
	for i, passenger := range req.GetPassengers() {
		passengers[i] = service.PassengerInput{
			FirstName:      passenger.GetFirstName(),
			LastName:       passenger.GetLastName(),
			DocumentNumber: passenger.GetDocumentNumber(),
			SeatNumber:     passenger.GetSeatNumber(),
			SeatClass:      protoToSeatClass(passenger.GetSeatClass()),
		}
	}

	input := service.CreateBookingInput{
		UserID:       req.GetUserId(),
		FlightID:     req.GetFlightId(),
		ContactEmail: req.GetContactEmail(),
		Passengers:   passengers,
	}

	booking, err := h.saga.CreateBooking(ctx, input)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateBookingResponse{
		BookingId: booking.ID,
		RefCode:   booking.RefCode,
		ExpiresAt: timestamppb.New(booking.ExpiresAt),
	}, nil
}

func (h *Handler) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.GetBookingResponse, error) {
	booking, err := h.queryService.GetBooking(ctx, req.GetBookingId(), req.GetRefCode())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetBookingResponse{
		Booking: bookingToProto(booking),
	}, nil
}

func (h *Handler) ListBookings(ctx context.Context, req *pb.ListBookingsRequest) (*pb.ListBookingsResponse, error) {
	bookings, err := h.queryService.ListBookings(ctx, req.GetUserId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbBookings := make([]*pb.Booking, len(bookings))
	for i := range bookings {
		pbBookings[i] = bookingToProto(&bookings[i])
	}

	return &pb.ListBookingsResponse{
		Bookings: pbBookings,
	}, nil
}

func (h *Handler) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
	if err := h.cancelService.CancelBooking(ctx, req.GetBookingId()); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CancelBookingResponse{}, nil
}
