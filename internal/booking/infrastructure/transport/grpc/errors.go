package grpcserver

import (
	"errors"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrBookingNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrBookingAlreadyCancelled),
		errors.Is(err, domain.ErrBookingAlreadyPaid),
		errors.Is(err, domain.ErrBookingExpired),
		errors.Is(err, domain.ErrSeatsUnavailable):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, domain.ErrNoPassengers),
		errors.Is(err, domain.ErrInvalidFlightID),
		errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidSeatNumber),
		errors.Is(err, domain.ErrInvalidSeatClass),
		errors.Is(err, domain.ErrInvalidDocument),
		errors.Is(err, domain.ErrInvalidName):
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return status.Error(codes.Internal, "internal error")
	}
}
