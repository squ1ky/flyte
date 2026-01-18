package flight

import (
	"context"
	"fmt"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type Client struct {
	api flightv1.FlightServiceClient
}

func NewClient(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}

	return &Client{
		api: flightv1.NewFlightServiceClient(conn),
	}, nil
}

func (c *Client) ReserveSeat(ctx context.Context, flightID int64, seatNumber string) error {
	_, err := c.api.ReserveSeat(ctx, &flightv1.ReserveSeatRequest{
		FlightId:   flightID,
		SeatNumber: seatNumber,
	})
	if err != nil {
		return fmt.Errorf("failed to reserve seat: %w", err)
	}
	return nil
}

func (c *Client) ReleaseSeat(ctx context.Context, flightID int64, seatNumber string) error {
	_, err := c.api.ReleaseSeat(ctx, &flightv1.ReleaseSeatRequest{
		FlightId:   flightID,
		SeatNumber: seatNumber,
	})
	if err != nil {
		return fmt.Errorf("failed to release seat: %w", err)
	}
	return nil
}

func (c *Client) ConfirmSeat(ctx context.Context, flightID int64, seatNumber string) error {
	req := &flightv1.ConfirmSeatRequest{
		FlightId:   flightID,
		SeatNumber: seatNumber,
	}

	_, err := c.api.ConfirmSeat(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to confirm seat: %w", err)
	}

	return nil
}
