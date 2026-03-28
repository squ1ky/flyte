package flightclient

import (
	"context"
	"fmt"
	pb "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type Client struct {
	flight pb.FlightServiceClient
	conn   *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}

	return &Client{
		flight: pb.NewFlightServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// ReserveSeats reserves seats on a flight
// Returns a list of unavailable seats and the reservation expiration time.
func (c *Client) ReserveSeats(ctx context.Context, flightID int64, seatNumbers []string, bookingID string) (unavailableSeats []string, expiresAt time.Time, err error) {
	resp, err := c.flight.ReserveSeats(ctx, &pb.ReserveSeatsRequest{
		FlightId:    flightID,
		SeatNumbers: seatNumbers,
		BookingId:   bookingID,
	})
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("reserve seats: %w", err)
	}

	var expires time.Time
	if resp.ExpiresAt != nil {
		expires = resp.ExpiresAt.AsTime()
	}
	return resp.UnavailableSeats, expires, nil
}

// ConfirmReservation confirms the reservation after successful payment.
func (c *Client) ConfirmReservation(ctx context.Context, flightID int64, seatNumbers []string, bookingID string) error {
	_, err := c.flight.ConfirmReservation(ctx, &pb.ConfirmReservationRequest{
		FlightId:    flightID,
		SeatNumbers: seatNumbers,
		BookingId:   bookingID,
	})
	if err != nil {
		return fmt.Errorf("confirm reservation: %w", err)
	}
	return nil
}

// CancelReservation cancels the seat reservation.
func (c *Client) CancelReservation(ctx context.Context, bookingID string) error {
	_, err := c.flight.CancelReservation(ctx, &pb.CancelReservationRequest{
		BookingId: bookingID,
	})
	if err != nil {
		return fmt.Errorf("cancel reservation: %w", err)
	}
	return nil
}

// GetFlightFares returns flight fares for calculating ticket prices.
func (c *Client) GetFlightFares(ctx context.Context, flightID int64) (map[domain.SeatClass]int64, error) {
	resp, err := c.flight.GetFlightFares(ctx, &pb.GetFlightFaresRequest{
		FlightId: flightID,
	})
	if err != nil {
		return nil, fmt.Errorf("get flight fares: %w", err)
	}

	fares := make(map[domain.SeatClass]int64, len(resp.Fares))
	for _, fare := range resp.Fares {
		seatClass := protoToSeatClass(fare.SeatClass)
		if seatClass == "" {
			continue
		}
		fares[seatClass] = fare.PriceCents
	}

	return fares, nil
}

var protoSeatClassMap = map[pb.SeatClass]domain.SeatClass{
	pb.SeatClass_SEAT_CLASS_ECONOMY:  domain.SeatClassEconomy,
	pb.SeatClass_SEAT_CLASS_BUSINESS: domain.SeatClassBusiness,
	pb.SeatClass_SEAT_CLASS_FIRST:    domain.SeatClassFirst,
}

func protoToSeatClass(pbSeatClass pb.SeatClass) domain.SeatClass {
	return protoSeatClassMap[pbSeatClass]
}
