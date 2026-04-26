package flight

import (
	flightv1 "github.com/squ1ky/flyte/gen/proto/flight"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	seatClassEconomy  = "economy"
	seatClassBusiness = "business"
	seatClassFirst    = "first"

	flightStatusScheduled = "scheduled"
	flightStatusDeparted  = "departed"
	flightStatusArrived   = "arrived"
	flightStatusCancelled = "cancelled"
)

// seatClassToProto converts a string seat class to the protobuf SeatClass enum.
func seatClassToProto(s string) flightv1.SeatClass {
	switch s {
	case seatClassEconomy:
		return flightv1.SeatClass_SEAT_CLASS_ECONOMY
	case seatClassBusiness:
		return flightv1.SeatClass_SEAT_CLASS_BUSINESS
	case seatClassFirst:
		return flightv1.SeatClass_SEAT_CLASS_FIRST
	default:
		return flightv1.SeatClass_SEAT_CLASS_UNSPECIFIED
	}
}

// seatClassFromProto converts a protobuf SeatClass enum to a string.
func seatClassFromProto(sc flightv1.SeatClass) string {
	switch sc {
	case flightv1.SeatClass_SEAT_CLASS_ECONOMY:
		return seatClassEconomy
	case flightv1.SeatClass_SEAT_CLASS_BUSINESS:
		return seatClassBusiness
	case flightv1.SeatClass_SEAT_CLASS_FIRST:
		return seatClassFirst
	default:
		return ""
	}
}

// flightStatusToProto converts a string status to the protobuf FlightStatus enum.
func flightStatusToProto(s string) flightv1.FlightStatus {
	switch s {
	case flightStatusScheduled:
		return flightv1.FlightStatus_FLIGHT_STATUS_SCHEDULED
	case flightStatusDeparted:
		return flightv1.FlightStatus_FLIGHT_STATUS_DEPARTED
	case flightStatusArrived:
		return flightv1.FlightStatus_FLIGHT_STATUS_ARRIVED
	case flightStatusCancelled:
		return flightv1.FlightStatus_FLIGHT_STATUS_CANCELLED
	default:
		return flightv1.FlightStatus_FLIGHT_STATUS_UNSPECIFIED
	}
}

// flightStatusFromProto converts a protobuf FlightStatus enum to a string.
func flightStatusFromProto(fs flightv1.FlightStatus) string {
	switch fs {
	case flightv1.FlightStatus_FLIGHT_STATUS_SCHEDULED:
		return flightStatusScheduled
	case flightv1.FlightStatus_FLIGHT_STATUS_DEPARTED:
		return flightStatusDeparted
	case flightv1.FlightStatus_FLIGHT_STATUS_ARRIVED:
		return flightStatusArrived
	case flightv1.FlightStatus_FLIGHT_STATUS_CANCELLED:
		return flightStatusCancelled
	default:
		return ""
	}
}

// mapProtoToFlight converts a protobuf Flight message to a FlightResponse DTO.
func mapProtoToFlight(f *flightv1.Flight) FlightResponse {
	resp := FlightResponse{
		ID:             f.Id,
		FlightNumber:   f.FlightNumber,
		Status:         flightStatusFromProto(f.Status),
		MinPriceCents:  f.MinPriceCents,
		AvailableSeats: f.AvailableSeats,
		Currency:       f.Currency,
	}

	if f.Airline != nil {
		resp.Airline = mapProtoToAirline(f.Airline)
	}
	if f.Aircraft != nil {
		resp.Aircraft = mapProtoToAircraft(f.Aircraft)
	}
	if f.DepartureAirport != nil {
		resp.DepartureAirport = mapProtoToAirport(f.DepartureAirport)
	}
	if f.ArrivalAirport != nil {
		resp.ArrivalAirport = mapProtoToAirport(f.ArrivalAirport)
	}
	if f.DepartureTime != nil {
		resp.DepartureTime = f.DepartureTime.AsTime().Format(time.RFC3339)
	}
	if f.ArrivalTime != nil {
		resp.ArrivalTime = f.ArrivalTime.AsTime().Format(time.RFC3339)
	}

	return resp
}

// mapProtoToAirport converts a protobuf Airport message to an AirportResponse DTO.
func mapProtoToAirport(a *flightv1.Airport) AirportResponse {
	return AirportResponse{
		Code:     a.Code,
		Name:     a.Name,
		City:     a.City,
		Country:  a.Country,
		Timezone: a.Timezone,
	}
}

// mapProtoToAirline converts a protobuf Airline message to an AirlineResponse DTO.
func mapProtoToAirline(a *flightv1.Airline) AirlineResponse {
	return AirlineResponse{
		ID:       a.Id,
		IATACode: a.IataCode,
		Name:     a.Name,
		Country:  a.Country,
		LogoURL:  a.LogoUrl,
	}
}

// mapProtoToAircraft converts a protobuf Aircraft message to an AircraftResponse DTO.
func mapProtoToAircraft(a *flightv1.Aircraft) AircraftResponse {
	return AircraftResponse{
		ID:         a.Id,
		Model:      a.Model,
		TotalSeats: a.TotalSeats,
	}
}

// mapProtoToSeat converts a protobuf Seat message to a SeatResponse DTO.
func mapProtoToSeat(s *flightv1.Seat) SeatResponse {
	return SeatResponse{
		ID:          s.Id,
		SeatNumber:  s.SeatNumber,
		SeatClass:   seatClassFromProto(s.SeatClass),
		Row:         s.Row,
		Col:         s.Col,
		IsAvailable: s.IsAvailable,
		PriceCents:  s.PriceCents,
	}
}

// mapProtoToSeatTemplate converts a protobuf SeatTemplate message to a SeatTemplateResponse DTO.
func mapProtoToSeatTemplate(s *flightv1.SeatTemplate) SeatTemplateResponse {
	return SeatTemplateResponse{
		SeatNumber: s.SeatNumber,
		SeatClass:  seatClassFromProto(s.SeatClass),
		Row:        s.Row,
		Col:        s.Col,
	}
}

// mapProtoToFareRule converts a protobuf FareRule message to a FareRuleResponse DTO.
func mapProtoToFareRule(f *flightv1.FareRule) FareRuleResponse {
	return FareRuleResponse{
		SeatClass:  seatClassFromProto(f.SeatClass),
		PriceCents: f.PriceCents,
	}
}

// mapProtoToSeatReservation converts a protobuf SeatReservation message to a SeatReservationResponse DTO.
func mapProtoToSeatReservation(s *flightv1.SeatReservation) SeatReservationResponse {
	resp := SeatReservationResponse{
		FlightID:   s.FlightId,
		SeatNumber: s.SeatNumber,
		BookingID:  s.BookingId,
	}
	if s.ExpiresAt != nil {
		resp.ExpiresAt = s.ExpiresAt.AsTime().Format(time.RFC3339)
	}
	return resp
}

// timestampFromDateStr parses a YYYY-MM-DD date string into a protobuf Timestamp.
func timestampFromDateStr(dateStr string) (*timestamppb.Timestamp, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	return timestamppb.New(t), nil
}

// timestampFromRFC3339 parses an RFC 3339 time string into a protobuf Timestamp.
func timestampFromRFC3339(s string) (*timestamppb.Timestamp, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return timestamppb.New(t), nil
}
