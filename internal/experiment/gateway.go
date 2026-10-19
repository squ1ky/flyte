package experiment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GatewayClient talks to the flyte gateway REST API exactly like the frontend does,
// so that generated traffic goes through the real saga (flight -> booking -> payment).
type GatewayClient struct {
	baseURL string
	http    *http.Client
	token   string
}

func NewGatewayClient(baseURL string) *GatewayClient {
	return &GatewayClient{baseURL: baseURL, http: &http.Client{Timeout: 15 * time.Second}}
}

type HTTPError struct {
	Status int
	Body   string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("http %d: %s", e.Status, e.Body) }

func (g *GatewayClient) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, g.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	}

	resp, err := g.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return &HTTPError{Status: resp.StatusCode, Body: string(raw)}
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func (g *GatewayClient) SignUp(ctx context.Context, email, password, phone string) error {
	return g.do(ctx, http.MethodPost, "/auth/sign-up",
		map[string]string{"email": email, "password": password, "phone_number": phone}, nil)
}

func (g *GatewayClient) SignIn(ctx context.Context, email, password string) error {
	var resp struct {
		Token string `json:"token"`
	}
	if err := g.do(ctx, http.MethodPost, "/auth/sign-in",
		map[string]string{"email": email, "password": password}, &resp); err != nil {
		return err
	}
	g.token = resp.Token
	return nil
}

type airline struct {
	ID       int64  `json:"id"`
	IATACode string `json:"iata_code"`
}

func (g *GatewayClient) ListAirlines(ctx context.Context) ([]airline, error) {
	var out []airline
	err := g.do(ctx, http.MethodGet, "/airlines?limit=100", nil, &out)
	return out, err
}

func (g *GatewayClient) CreateAirline(ctx context.Context, iata, name string) (int64, error) {
	var out struct {
		ID int64 `json:"id"`
	}
	err := g.do(ctx, http.MethodPost, "/airlines",
		map[string]string{"iata_code": iata, "name": name, "country": "Russia"}, &out)
	return out.ID, err
}

type aircraft struct {
	ID    int64  `json:"id"`
	Model string `json:"model"`
}

func (g *GatewayClient) ListAircrafts(ctx context.Context) ([]aircraft, error) {
	var out []aircraft
	err := g.do(ctx, http.MethodGet, "/aircrafts?limit=100", nil, &out)
	return out, err
}

func (g *GatewayClient) CreateAircraft(ctx context.Context, model string, totalSeats int) (int64, error) {
	var out struct {
		ID int64 `json:"id"`
	}
	err := g.do(ctx, http.MethodPost, "/aircrafts",
		map[string]any{"model": model, "total_seats": totalSeats}, &out)
	return out.ID, err
}

type seatTemplate struct {
	SeatNumber string `json:"seat_number"`
	SeatClass  string `json:"seat_class"`
	Row        int    `json:"row"`
	Col        int    `json:"col"`
}

func (g *GatewayClient) GetAircraftSeats(ctx context.Context, aircraftID int64) ([]seatTemplate, error) {
	var out []seatTemplate
	err := g.do(ctx, http.MethodGet, fmt.Sprintf("/aircrafts/%d/seats", aircraftID), nil, &out)
	return out, err
}

func (g *GatewayClient) AddAircraftSeats(ctx context.Context, aircraftID int64, seats []seatTemplate) error {
	return g.do(ctx, http.MethodPost, fmt.Sprintf("/aircrafts/%d/seats", aircraftID),
		map[string]any{"seats": seats}, nil)
}

func (g *GatewayClient) CreateFlight(ctx context.Context, number string, airlineID, aircraftID int64, departure time.Time) (int64, error) {
	var out struct {
		ID int64 `json:"id"`
	}
	err := g.do(ctx, http.MethodPost, "/flights", map[string]any{
		"flight_number":          number,
		"aircraft_id":            aircraftID,
		"airline_id":             airlineID,
		"departure_airport_code": "DME",
		"arrival_airport_code":   "LED",
		"departure_time":         departure.Format(time.RFC3339),
		"arrival_time":           departure.Add(90 * time.Minute).Format(time.RFC3339),
		"fare_rules":             []map[string]any{{"seat_class": "economy", "price_cents": 100000}},
	}, &out)
	return out.ID, err
}

func (g *GatewayClient) CreateBooking(ctx context.Context, flightID int64, seat string, n int64) (string, error) {
	var out struct {
		BookingID string `json:"booking_id"`
	}
	err := g.do(ctx, http.MethodPost, "/bookings", map[string]any{
		"flight_id":     flightID,
		"contact_email": "experiment@flyte.local",
		"passengers": []map[string]string{{
			"first_name":      "Experiment",
			"last_name":       "Passenger",
			"document_number": fmt.Sprintf("EXP%07d", n%10_000_000),
			"seat_number":     seat,
			"seat_class":      "economy",
		}},
	}, &out)
	return out.BookingID, err
}

func (g *GatewayClient) CancelBooking(ctx context.Context, bookingID string) error {
	return g.do(ctx, http.MethodPost, "/bookings/"+bookingID+"/cancel", nil, nil)
}
