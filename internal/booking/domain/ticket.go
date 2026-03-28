package domain

type TicketStatus string

const (
	TicketStatusReserved TicketStatus = "RESERVED"
	TicketStatusIssued   TicketStatus = "ISSUED"
)

type SeatClass string

const (
	SeatClassEconomy  SeatClass = "ECONOMY"
	SeatClassBusiness SeatClass = "BUSINESS"
	SeatClassFirst    SeatClass = "FIRST"
)

var validSeatClasses = map[SeatClass]struct{}{
	SeatClassEconomy:  {},
	SeatClassBusiness: {},
	SeatClassFirst:    {},
}

func (c SeatClass) IsValid() bool {
	_, ok := validSeatClasses[c]
	return ok
}

type Ticket struct {
	ID                 string       `db:"id"`
	BookingID          string       `db:"booking_id"`
	PassengerFirstName string       `db:"passenger_first_name"`
	PassengerLastName  string       `db:"passenger_last_name"`
	PassengerDocNum    string       `db:"passenger_doc_num"`
	SeatNumber         string       `db:"seat_number"`
	PriceCents         int64        `db:"price_cents"`
	TicketNumber       string       `db:"ticket_number"`
	Status             TicketStatus `db:"status"`
}
