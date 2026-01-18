package grpc

import (
	"errors"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

var (
	errFlightIDRequired     = errors.New("flight ID is required")
	errFlightNumberRequired = errors.New("flight number is required")
	errAirportsRequired     = errors.New("departure and arrival airports are required")
	errSeatIsRequired       = errors.New("seat number is required")
	errSameAirports         = errors.New("departure and arrival airports must be different")
	errInvalidTime          = errors.New("arrival time must be after departure time")
	errInvalidPrice         = errors.New("price must be positive")
	errInvalidPassenger     = errors.New("passenger count must be positive")

	errAircraftIDRequired    = errors.New("aircraft ID is required")
	errAircraftModelRequired = errors.New("aircraft model is required")
	errTotalSeatsInvalid     = errors.New("total seats must be positive")
	errSeatsListEmpty        = errors.New("seats list is empty")
)

func validateCreateFlightRequest(req *flightv1.CreateFlightRequest) error {
	if strings.TrimSpace(req.FlightNumber) == "" {
		return status.Error(codes.InvalidArgument, errFlightNumberRequired.Error())
	}
	if req.AircraftId <= 0 {
		return status.Error(codes.InvalidArgument, errAircraftIDRequired.Error())
	}
	if strings.TrimSpace(req.DepartureAirport) == "" || strings.TrimSpace(req.ArrivalAirport) == "" {
		return status.Error(codes.InvalidArgument, errAirportsRequired.Error())
	}
	if req.DepartureAirport == req.ArrivalAirport {
		return status.Error(codes.InvalidArgument, errSameAirports.Error())
	}
	if req.BasePriceCents <= 0 {
		return status.Error(codes.InvalidArgument, errInvalidPrice.Error())
	}

	depTime := req.DepartureTime.AsTime()
	arrTime := req.ArrivalTime.AsTime()
	if !arrTime.After(depTime) {
		return status.Error(codes.InvalidArgument, errInvalidTime.Error())
	}
	return nil
}

func validateSearchFlightsRequest(req *flightv1.SearchFlightsRequest) error {
	if strings.TrimSpace(req.FromAirport) == "" || strings.TrimSpace(req.ToAirport) == "" {
		return status.Error(codes.InvalidArgument, errAirportsRequired.Error())
	}
	if req.PassengerCount <= 0 {
		return status.Error(codes.InvalidArgument, errInvalidPassenger.Error())
	}
	return nil
}

func validateGetFlightDetailsRequest(req *flightv1.GetFlightDetailsRequest) error {
	if req.FlightId <= 0 {
		return status.Error(codes.InvalidArgument, errFlightIDRequired.Error())
	}
	return nil
}

func validateGetFlightSeatsRequest(req *flightv1.GetFlightSeatsRequest) error {
	if req.FlightId <= 0 {
		return status.Error(codes.InvalidArgument, errFlightIDRequired.Error())
	}
	return nil
}

func validateReserveSeatRequest(req *flightv1.ReserveSeatRequest) error {
	if req.FlightId <= 0 {
		return status.Error(codes.InvalidArgument, errFlightIDRequired.Error())
	}
	if strings.TrimSpace(req.SeatNumber) == "" {
		return status.Error(codes.InvalidArgument, errSeatIsRequired.Error())
	}
	return nil
}

func validateReleaseSeatRequest(req *flightv1.ReleaseSeatRequest) error {
	if req.FlightId <= 0 {
		return status.Error(codes.InvalidArgument, errFlightIDRequired.Error())
	}
	if strings.TrimSpace(req.SeatNumber) == "" {
		return status.Error(codes.InvalidArgument, errSeatIsRequired.Error())
	}
	return nil
}

func validateConfirmSeatRequest(req *flightv1.ConfirmSeatRequest) error {
	if req.FlightId <= 0 {
		return status.Error(codes.InvalidArgument, errFlightIDRequired.Error())
	}
	if strings.TrimSpace(req.SeatNumber) == "" {
		return status.Error(codes.InvalidArgument, errSeatIsRequired.Error())
	}
	return nil
}

func validateCreateAircraftRequest(req *flightv1.CreateAircraftRequest) error {
	if strings.TrimSpace(req.Model) == "" {
		return status.Error(codes.InvalidArgument, errAircraftModelRequired.Error())
	}
	if req.TotalSeats <= 0 {
		return status.Error(codes.InvalidArgument, errTotalSeatsInvalid.Error())
	}
	return nil
}

func validateAddAircraftSeatsRequest(req *flightv1.AddAircraftSeatsRequest) error {
	if req.AircraftId <= 0 {
		return status.Error(codes.InvalidArgument, errAircraftIDRequired.Error())
	}
	if len(req.Seats) == 0 {
		return status.Error(codes.InvalidArgument, errSeatsListEmpty.Error())
	}
	return nil
}
