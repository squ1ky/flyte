package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"github.com/squ1ky/flyte/internal/flight/infrastructure/repository/outbox"
	"log/slog"
	"time"
)

const (
	eventsLimit = 50
)

type ElasticSync struct {
	outbox         outbox.Storage
	flightRepo     domain.FlightStorage
	airlineRepo    domain.AirlineStorage
	airportRepo    domain.AirportStorage
	flightSearcher domain.FlightSearcher
	logger         *slog.Logger
}

func NewElasticSync(
	outbox outbox.Storage,
	flightRepo domain.FlightStorage,
	airlineRepo domain.AirlineStorage,
	airportRepo domain.AirportStorage,
	searcher domain.FlightSearcher,
	logger *slog.Logger,
) *ElasticSync {
	return &ElasticSync{
		outbox:         outbox,
		flightRepo:     flightRepo,
		airlineRepo:    airlineRepo,
		airportRepo:    airportRepo,
		flightSearcher: searcher,
		logger:         logger,
	}
}

func (w *ElasticSync) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	w.logger.InfoContext(ctx, "starting elastic sync worker")

	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "stopping elastic sync worker")
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.ErrorContext(ctx, "batch processing error",
					slog.Any("error", err),
				)
			}
		}
	}
}

func (w *ElasticSync) processBatch(ctx context.Context) error {
	events, err := w.outbox.FetchPending(ctx, eventsLimit)
	if err != nil {
		return fmt.Errorf("fetch pending events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	var processedIDs []string
	var failedIDs []string

	for _, event := range events {
		if err := w.handleEvent(ctx, event); err != nil {
			w.logger.ErrorContext(ctx, "handle event failed",
				slog.String("id", event.ID),
				slog.String("type", string(event.EventType)),
				slog.Any("error", err),
			)
			failedIDs = append(failedIDs, event.ID)
			continue
		}
		processedIDs = append(processedIDs, event.ID)
	}

	if len(failedIDs) > 0 {
		if err := w.outbox.MarkFailed(ctx, failedIDs); err != nil {
			w.logger.ErrorContext(ctx, "failed to mark events as failed",
				slog.Any("error", err),
			)
		}
	}

	if len(processedIDs) > 0 {
		if err := w.outbox.MarkProcessed(ctx, processedIDs); err != nil {
			return fmt.Errorf("mark events processed: %w", err)
		}
	}

	return nil
}

func (w *ElasticSync) handleEvent(ctx context.Context, event outbox.Event) error {
	switch event.EventType {
	case domain.EventTypeFlightCreated:
		return w.handleFlightCreated(ctx, event.Payload)
	case domain.EventTypeSeatsReserved,
		domain.EventTypeSeatsConfirmed,
		domain.EventTypeSeatsReleased:
		return w.handleSeatsChanged(ctx, event.Payload)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

func (w *ElasticSync) handleFlightCreated(ctx context.Context, payload json.RawMessage) error {
	var data domain.FlightCreatedPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal flight created: %w", err)
	}

	doc, err := w.buildFlightDocument(ctx, data.FlightID)
	if err != nil {
		return err
	}
	return w.flightSearcher.IndexFlight(ctx, doc)
}

func (w *ElasticSync) handleSeatsChanged(ctx context.Context, payload json.RawMessage) error {
	var data domain.SeatsStatusPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal seats event: %w", err)
	}

	doc, err := w.buildFlightDocument(ctx, data.FlightID)
	if err != nil {
		return err
	}
	return w.flightSearcher.IndexFlight(ctx, doc)
}

func (w *ElasticSync) buildFlightDocument(ctx context.Context, flightID int64) (*domain.FlightDocument, error) {
	flight, err := w.flightRepo.GetByID(ctx, flightID)
	if err != nil {
		return nil, fmt.Errorf("get flight: %w", err)
	}

	fares, err := w.flightRepo.GetFaresByFlightID(ctx, flight.ID)
	if err != nil {
		return nil, fmt.Errorf("get fares: %w", err)
	}

	airline, err := w.airlineRepo.GetByID(ctx, flight.AirlineID)
	if err != nil {
		return nil, fmt.Errorf("get airline: %w", err)
	}

	departure, err := w.airportRepo.GetByCode(ctx, flight.DepartureAirportCode)
	if err != nil {
		return nil, fmt.Errorf("get departure airport: %w", err)
	}

	arrival, err := w.airportRepo.GetByCode(ctx, flight.ArrivalAirportCode)
	if err != nil {
		return nil, fmt.Errorf("get arrival airport: %w", err)
	}

	var minPrice int64
	var availableSeats int
	for i, fare := range fares {
		available := fare.SeatsTotal - fare.SeatsBooked
		availableSeats += available
		if i == 0 || fare.PriceCents < minPrice {
			minPrice = fare.PriceCents
		}
	}

	return &domain.FlightDocument{
		ID:             flight.ID,
		FlightNumber:   flight.FlightNumber,
		DepartureTime:  flight.DepartureTime,
		ArrivalTime:    flight.ArrivalTime,
		MinPriceCents:  minPrice,
		AvailableSeats: availableSeats,
		Status:         flight.Status,
		Airline:        *airline,
		Departure:      *departure,
		Arrival:        *arrival,
	}, nil
}
