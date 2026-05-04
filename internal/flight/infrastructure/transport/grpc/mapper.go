package grpcserver

import (
	pb "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func seatClassToProto(c domain.SeatClass) pb.SeatClass {
	switch c {
	case domain.SeatClassEconomy:
		return pb.SeatClass_SEAT_CLASS_ECONOMY
	case domain.SeatClassBusiness:
		return pb.SeatClass_SEAT_CLASS_BUSINESS
	case domain.SeatClassFirst:
		return pb.SeatClass_SEAT_CLASS_FIRST
	default:
		return pb.SeatClass_SEAT_CLASS_UNSPECIFIED
	}
}

func seatClassToDomain(c pb.SeatClass) domain.SeatClass {
	switch c {
	case pb.SeatClass_SEAT_CLASS_ECONOMY:
		return domain.SeatClassEconomy
	case pb.SeatClass_SEAT_CLASS_BUSINESS:
		return domain.SeatClassBusiness
	case pb.SeatClass_SEAT_CLASS_FIRST:
		return domain.SeatClassFirst
	default:
		return ""
	}
}

func flightStatusToProto(s domain.FlightStatus) pb.FlightStatus {
	switch s {
	case domain.FlightStatusScheduled:
		return pb.FlightStatus_FLIGHT_STATUS_SCHEDULED
	case domain.FlightStatusDeparted:
		return pb.FlightStatus_FLIGHT_STATUS_DEPARTED
	case domain.FlightStatusArrived:
		return pb.FlightStatus_FLIGHT_STATUS_ARRIVED
	case domain.FlightStatusCancelled:
		return pb.FlightStatus_FLIGHT_STATUS_CANCELLED
	default:
		return pb.FlightStatus_FLIGHT_STATUS_UNSPECIFIED
	}
}

func flightStatusToDomain(s pb.FlightStatus) domain.FlightStatus {
	switch s {
	case pb.FlightStatus_FLIGHT_STATUS_SCHEDULED:
		return domain.FlightStatusScheduled
	case pb.FlightStatus_FLIGHT_STATUS_DEPARTED:
		return domain.FlightStatusDeparted
	case pb.FlightStatus_FLIGHT_STATUS_ARRIVED:
		return domain.FlightStatusArrived
	case pb.FlightStatus_FLIGHT_STATUS_CANCELLED:
		return domain.FlightStatusCancelled
	default:
		return ""
	}
}

func airportToProto(a *domain.Airport) *pb.Airport {
	if a == nil {
		return nil
	}
	return &pb.Airport{
		Code:     a.Code,
		Name:     a.Name,
		City:     a.City,
		Country:  a.Country,
		Timezone: a.Timezone,
	}
}

func aircraftToProto(a *domain.Aircraft) *pb.Aircraft {
	if a == nil {
		return nil
	}
	return &pb.Aircraft{
		Id:         a.ID,
		Model:      a.Model,
		TotalSeats: int32(a.TotalSeats),
	}
}

func airlineToProto(a *domain.Airline) *pb.Airline {
	if a == nil {
		return nil
	}
	return &pb.Airline{
		Id:       a.ID,
		IataCode: a.IATACode,
		Name:     a.Name,
		Country:  a.Country,
		LogoUrl:  a.LogoURL,
	}
}

func flightDocToProto(f *domain.FlightDocument) *pb.Flight {
	if f == nil {
		return nil
	}
	return &pb.Flight{
		Id:               f.ID,
		FlightNumber:     f.FlightNumber,
		Airline:          airlineToProto(&f.Airline),
		DepartureAirport: airportToProto(&f.Departure),
		ArrivalAirport:   airportToProto(&f.Arrival),
		DepartureTime:    timestamppb.New(f.DepartureTime),
		ArrivalTime:      timestamppb.New(f.ArrivalTime),
		Status:           flightStatusToProto(f.Status),
		MinPriceCents:    f.MinPriceCents,
		AvailableSeats:   int32(f.AvailableSeats),
	}
}

func flightToProto(f *domain.Flight) *pb.Flight {
	if f == nil {
		return nil
	}
	return &pb.Flight{
		Id:            f.ID,
		FlightNumber:  f.FlightNumber,
		DepartureTime: timestamppb.New(f.DepartureTime),
		ArrivalTime:   timestamppb.New(f.ArrivalTime),
		Status:        flightStatusToProto(f.Status),
	}
}

func seatTemplateToProto(s *domain.AircraftSeat) *pb.SeatTemplate {
	if s == nil {
		return nil
	}
	return &pb.SeatTemplate{
		SeatNumber: s.SeatNumber,
		SeatClass:  seatClassToProto(s.SeatClass),
		Row:        int32(s.Row),
		Col:        int32(s.Col),
	}
}

func fareToProto(f *domain.FlightFare) *pb.FareRule {
	if f == nil {
		return nil
	}
	return &pb.FareRule{
		SeatClass:  seatClassToProto(f.SeatClass),
		PriceCents: f.PriceCents,
	}
}

func seatReservationToProto(r *domain.SeatReservation) *pb.SeatReservation {
	if r == nil {
		return nil
	}
	res := &pb.SeatReservation{
		FlightId:   r.FlightID,
		SeatNumber: r.SeatNumber,
		BookingId:  r.BookingID,
	}
	if r.ExpiresAt != nil {
		res.ExpiresAt = timestamppb.New(*r.ExpiresAt)
	}
	return res
}

func seatTemplateToDomain(s *pb.SeatTemplate) domain.AircraftSeat {
	return domain.AircraftSeat{
		SeatNumber: s.GetSeatNumber(),
		SeatClass:  seatClassToDomain(s.GetSeatClass()),
		Row:        int(s.GetRow()),
		Col:        int(s.GetCol()),
	}
}

func fareRuleToDomain(f *pb.FareRule) domain.FlightFare {
	return domain.FlightFare{
		SeatClass:  seatClassToDomain(f.GetSeatClass()),
		PriceCents: f.GetPriceCents(),
	}
}
