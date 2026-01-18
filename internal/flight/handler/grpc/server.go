package grpc

import (
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"github.com/squ1ky/flyte/internal/flight/service"
)

type Server struct {
	flightv1.UnimplementedFlightServiceServer
	flightService   *service.FlightService
	aircraftService *service.AircraftService
}

func NewServer(flightService *service.FlightService, aircraftService *service.AircraftService) *Server {
	return &Server{
		flightService:   flightService,
		aircraftService: aircraftService,
	}
}
