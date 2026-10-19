package experiment

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	runnerEmail    = "experiment-runner@flyte.local"
	runnerPassword = "experiment-runner-pass"
	runnerPhone    = "+70000000000"

	airlineIATA   = "XE"
	aircraftModel = "flyte-experiment-180"
	seatRows      = 30
	seatCols      = 6
)

// PrepareAdmin makes sure the runner's user exists and has the admin role (needed to create
// flights), then signs in. Role is baked into the JWT, so sign-in must happen after promotion.
func PrepareAdmin(ctx context.Context, gw *GatewayClient, userDBURL string) error {
	// Fails when the user is already registered by a previous experiment; that is checked below
	// by whether the promotion actually found the user.
	signUpErr := gw.SignUp(ctx, runnerEmail, runnerPassword, runnerPhone)

	conn, err := pgx.Connect(ctx, userDBURL)
	if err != nil {
		return fmt.Errorf("connect user-db: %w", err)
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, `UPDATE users SET role = 'admin' WHERE email = $1`, runnerEmail)
	if err != nil {
		return fmt.Errorf("promote runner user to admin: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("runner user %s not found after sign up: %w", runnerEmail, signUpErr)
	}

	if err := gw.SignIn(ctx, runnerEmail, runnerPassword); err != nil {
		return fmt.Errorf("sign in runner user: %w", err)
	}
	return nil
}

// SeatPool hands out never-used (flight, seat) pairs. Paid bookings occupy seats forever,
// so over many runs the pool keeps creating fresh flights on a dedicated experiment aircraft
// instead of exhausting whatever flights exist in the database.
type SeatPool struct {
	gw         *GatewayClient
	logger     *slog.Logger
	airlineID  int64
	aircraftID int64
	seats      []string

	mu        sync.Mutex
	flightID  int64
	nextSeat  int
	flightSeq int
}

func NewSeatPool(ctx context.Context, gw *GatewayClient, logger *slog.Logger) (*SeatPool, error) {
	airlineID, err := ensureAirline(ctx, gw)
	if err != nil {
		return nil, err
	}
	aircraftID, seats, err := ensureAircraft(ctx, gw)
	if err != nil {
		return nil, err
	}
	return &SeatPool{gw: gw, logger: logger, airlineID: airlineID, aircraftID: aircraftID, seats: seats}, nil
}

// Next returns a seat that has not been handed out before. Creates a new flight when needed.
func (p *SeatPool) Next(ctx context.Context) (int64, string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.flightID == 0 || p.nextSeat >= len(p.seats) {
		p.flightSeq++
		number := fmt.Sprintf("XE%d", time.Now().Unix()%100_000*10+int64(p.flightSeq%10))
		departure := time.Now().Add(30 * 24 * time.Hour).Truncate(time.Minute)
		id, err := p.gw.CreateFlight(ctx, number, p.airlineID, p.aircraftID, departure)
		if err != nil {
			return 0, "", fmt.Errorf("create experiment flight: %w", err)
		}
		p.logger.Info("created experiment flight", "flight_id", id, "seats", len(p.seats))
		p.flightID, p.nextSeat = id, 0
	}

	seat := p.seats[p.nextSeat]
	p.nextSeat++
	return p.flightID, seat, nil
}

func ensureAirline(ctx context.Context, gw *GatewayClient) (int64, error) {
	airlines, err := gw.ListAirlines(ctx)
	if err != nil {
		return 0, fmt.Errorf("list airlines: %w", err)
	}
	for _, a := range airlines {
		if a.IATACode == airlineIATA {
			return a.ID, nil
		}
	}
	id, err := gw.CreateAirline(ctx, airlineIATA, "Flyte Experiment Air")
	if err != nil {
		return 0, fmt.Errorf("create airline: %w", err)
	}
	return id, nil
}

func ensureAircraft(ctx context.Context, gw *GatewayClient) (int64, []string, error) {
	aircrafts, err := gw.ListAircrafts(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("list aircrafts: %w", err)
	}

	var id int64
	for _, a := range aircrafts {
		if a.Model == aircraftModel {
			id = a.ID
			break
		}
	}
	if id == 0 {
		if id, err = gw.CreateAircraft(ctx, aircraftModel, seatRows*seatCols); err != nil {
			return 0, nil, fmt.Errorf("create aircraft: %w", err)
		}
	}

	templates, err := gw.GetAircraftSeats(ctx, id)
	if err != nil {
		return 0, nil, fmt.Errorf("get aircraft seats: %w", err)
	}
	if len(templates) == 0 {
		templates = make([]seatTemplate, 0, seatRows*seatCols)
		for row := 1; row <= seatRows; row++ {
			for col := 1; col <= seatCols; col++ {
				templates = append(templates, seatTemplate{
					SeatNumber: fmt.Sprintf("%d%c", row, 'A'+col-1),
					SeatClass:  "economy",
					Row:        row,
					Col:        col,
				})
			}
		}
		if err := gw.AddAircraftSeats(ctx, id, templates); err != nil {
			return 0, nil, fmt.Errorf("add aircraft seats: %w", err)
		}
	}

	seats := make([]string, len(templates))
	for i, t := range templates {
		seats[i] = t.SeatNumber
	}
	return id, seats, nil
}
