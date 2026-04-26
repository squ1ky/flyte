package booking

import (
	bookingv1 "github.com/squ1ky/flyte/gen/proto/booking"
	"time"
)

const (
	bookingStatusPending   = "pending"
	bookingStatusPaid      = "paid"
	bookingStatusCancelled = "cancelled"

	ticketStatusReserved = "reserved"
	ticketStatusIssued   = "issued"

	seatClassEconomy  = "economy"
	seatClassBusiness = "business"
	seatClassFirst    = "first"
)

func bookingStatusFromProto(s bookingv1.BookingStatus) string {
	switch s {
	case bookingv1.BookingStatus_BOOKING_STATUS_PENDING:
		return bookingStatusPending
	case bookingv1.BookingStatus_BOOKING_STATUS_PAID:
		return bookingStatusPaid
	case bookingv1.BookingStatus_BOOKING_STATUS_CANCELLED:
		return bookingStatusCancelled
	default:
		return ""
	}
}

func ticketStatusFromProto(s bookingv1.TicketStatus) string {
	switch s {
	case bookingv1.TicketStatus_TICKET_STATUS_RESERVED:
		return ticketStatusReserved
	case bookingv1.TicketStatus_TICKET_STATUS_ISSUED:
		return ticketStatusIssued
	default:
		return ""
	}
}

func seatClassToProto(s string) bookingv1.SeatClass {
	switch s {
	case seatClassEconomy:
		return bookingv1.SeatClass_SEAT_CLASS_ECONOMY
	case seatClassBusiness:
		return bookingv1.SeatClass_SEAT_CLASS_BUSINESS
	case seatClassFirst:
		return bookingv1.SeatClass_SEAT_CLASS_FIRST
	default:
		return bookingv1.SeatClass_SEAT_CLASS_UNSPECIFIED
	}
}

func seatClassFromProto(sc bookingv1.SeatClass) string {
	switch sc {
	case bookingv1.SeatClass_SEAT_CLASS_ECONOMY:
		return seatClassEconomy
	case bookingv1.SeatClass_SEAT_CLASS_BUSINESS:
		return seatClassBusiness
	case bookingv1.SeatClass_SEAT_CLASS_FIRST:
		return seatClassFirst
	default:
		return ""
	}
}

func mapProtoToBooking(b *bookingv1.Booking) BookingResponse {
	resp := BookingResponse{
		ID:              b.Id,
		UserID:          b.UserId,
		FlightID:        b.FlightId,
		RefCode:         b.RefCode,
		Status:          bookingStatusFromProto(b.Status),
		TotalPriceCents: b.TotalPriceCents,
		Currency:        b.Currency,
		ContactEmail:    b.ContactEmail,
	}

	if b.ExpiresAt != nil {
		resp.ExpiresAt = b.ExpiresAt.AsTime().Format(time.RFC3339)
	}
	if b.CreatedAt != nil {
		resp.CreatedAt = b.CreatedAt.AsTime().Format(time.RFC3339)
	}

	resp.Tickets = make([]TicketResponse, len(b.Tickets))
	for i, t := range b.Tickets {
		resp.Tickets[i] = mapProtoToTicket(t)
	}

	return resp
}

func mapProtoToTicket(t *bookingv1.Ticket) TicketResponse {
	return TicketResponse{
		ID:                 t.Id,
		PassengerFirstName: t.PassengerFirstName,
		PassengerLastName:  t.PassengerLastName,
		PassengerDocNum:    t.PassengerDocNum,
		SeatNumber:         t.SeatNumber,
		SeatClass:          seatClassFromProto(t.SeatClass),
		PriceCents:         t.PriceCents,
		TicketNumber:       t.TicketNumber,
		Status:             ticketStatusFromProto(t.Status),
	}
}
