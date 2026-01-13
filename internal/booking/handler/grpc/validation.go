package grpc

import (
	bookingv1 "github.com/squ1ky/flyte/gen/go/booking"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func validateCreateBookingRequest(req *bookingv1.CreateBookingRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.UserId <= 0 {
		return status.Error(codes.InvalidArgument, "user_id must be > 0")
	}
	if req.FlightId <= 0 {
		return status.Error(codes.InvalidArgument, "flight_id must be > 0")
	}
	if strings.TrimSpace(req.SeatNumber) == "" {
		return status.Error(codes.InvalidArgument, "seat_number is required")
	}
	if strings.TrimSpace(req.PassengerName) == "" {
		return status.Error(codes.InvalidArgument, "passenger_name is required")
	}
	if strings.TrimSpace(req.PassengerPassport) == "" {
		return status.Error(codes.InvalidArgument, "passenger_passport is required")
	}
	if req.Price <= 0 {
		return status.Error(codes.InvalidArgument, "price must be > 0")
	}
	return nil
}

func validateGetBookingRequest(req *bookingv1.GetBookingRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if strings.TrimSpace(req.BookingId) == "" {
		return status.Error(codes.InvalidArgument, "booking_id is required")
	}
	return nil
}

func validateListBookingsRequest(req *bookingv1.ListBookingsRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.UserId <= 0 {
		return status.Error(codes.InvalidArgument, "user_id must be > 0")
	}
	return nil
}

func validateCancelBookingRequest(req *bookingv1.CancelBookingRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if strings.TrimSpace(req.BookingId) == "" {
		return status.Error(codes.InvalidArgument, "booking_id is required")
	}
	return nil
}
