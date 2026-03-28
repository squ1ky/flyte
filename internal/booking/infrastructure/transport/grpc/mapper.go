package grpcserver

import (
	pb "github.com/squ1ky/flyte/gen/proto/booking"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func bookingToProto(b *domain.Booking) *pb.Booking {
	pbBooking := &pb.Booking{
		Id:              b.ID,
		UserId:          b.UserID,
		FlightId:        b.FlightID,
		RefCode:         b.RefCode,
		Status:          bookingStatusToProto(b.Status),
		TotalPriceCents: b.TotalPriceCents,
		Currency:        b.Currency,
		ContactEmail:    b.ContactEmail,
		ExpiresAt:       timestamppb.New(b.ExpiresAt),
		CreatedAt:       timestamppb.New(b.CreatedAt),
	}

	if len(b.Tickets) > 0 {
		pbBooking.Tickets = make([]*pb.Ticket, len(b.Tickets))
		for i := range b.Tickets {
			pbBooking.Tickets[i] = ticketToProto(&b.Tickets[i])
		}
	}

	return pbBooking
}

func ticketToProto(t *domain.Ticket) *pb.Ticket {
	return &pb.Ticket{
		Id:                 t.ID,
		PassengerFirstName: t.PassengerFirstName,
		PassengerLastName:  t.PassengerLastName,
		PassengerDocNum:    t.PassengerDocNum,
		SeatNumber:         t.SeatNumber,
		PriceCents:         t.PriceCents,
		TicketNumber:       t.TicketNumber,
		Status:             ticketStatusToProto(t.Status),
	}
}

var bookingStatusMap = map[domain.BookingStatus]pb.BookingStatus{
	domain.BookingStatusPending:   pb.BookingStatus_BOOKING_STATUS_PENDING,
	domain.BookingStatusPaid:      pb.BookingStatus_BOOKING_STATUS_PAID,
	domain.BookingStatusCancelled: pb.BookingStatus_BOOKING_STATUS_CANCELLED,
}

func bookingStatusToProto(s domain.BookingStatus) pb.BookingStatus {
	if ps, ok := bookingStatusMap[s]; ok {
		return ps
	}
	return pb.BookingStatus_BOOKING_STATUS_UNSPECIFIED
}

var ticketStatusMap = map[domain.TicketStatus]pb.TicketStatus{
	domain.TicketStatusReserved: pb.TicketStatus_TICKET_STATUS_RESERVED,
	domain.TicketStatusIssued:   pb.TicketStatus_TICKET_STATUS_ISSUED,
}

func ticketStatusToProto(s domain.TicketStatus) pb.TicketStatus {
	if ps, ok := ticketStatusMap[s]; ok {
		return ps
	}
	return pb.TicketStatus_TICKET_STATUS_UNSPECIFIED
}

var protoSeatClassMap = map[pb.SeatClass]domain.SeatClass{
	pb.SeatClass_SEAT_CLASS_ECONOMY:  domain.SeatClassEconomy,
	pb.SeatClass_SEAT_CLASS_BUSINESS: domain.SeatClassBusiness,
	pb.SeatClass_SEAT_CLASS_FIRST:    domain.SeatClassFirst,
}

func protoToSeatClass(sc pb.SeatClass) domain.SeatClass {
	if dc, ok := protoSeatClassMap[sc]; ok {
		return dc
	}
	return ""
}
