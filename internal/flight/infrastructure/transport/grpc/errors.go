package grpcserver

import (
	"errors"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrFlightNotFound),
		errors.Is(err, domain.ErrAirportNotFound),
		errors.Is(err, domain.ErrAircraftNotFound),
		errors.Is(err, domain.ErrAirlineNotFound),
		errors.Is(err, domain.ErrReservationNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrFlightAlreadyExists),
		errors.Is(err, domain.ErrAircraftAlreadyExists),
		errors.Is(err, domain.ErrAirlineAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, domain.ErrInvalidFlightID),
		errors.Is(err, domain.ErrInvalidAircraftID),
		errors.Is(err, domain.ErrInvalidAirlineID),
		errors.Is(err, domain.ErrInvalidAircraftModel),
		errors.Is(err, domain.ErrInvalidTotalSeats),
		errors.Is(err, domain.ErrEmptySearchQuery),
		errors.Is(err, domain.ErrInvalidIATACode),
		errors.Is(err, domain.ErrInvalidAirlineName),
		errors.Is(err, domain.ErrEmptySeatsList),
		errors.Is(err, domain.ErrEmptyBookingID),
		errors.Is(err, domain.ErrEmptyFlightNumber),
		errors.Is(err, domain.ErrInvalidAirportCode),
		errors.Is(err, domain.ErrSameAirports),
		errors.Is(err, domain.ErrInvalidFlightTime),
		errors.Is(err, domain.ErrInvalidSearchCriteria),
		errors.Is(err, domain.ErrInvalidSearchDate),
		errors.Is(err, domain.ErrInvalidPassengerCount):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrSeatUnavailable),
		errors.Is(err, domain.ErrReservationExpired):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
