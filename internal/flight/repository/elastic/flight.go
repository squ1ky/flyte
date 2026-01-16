package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"github.com/squ1ky/flyte/internal/flight/repository"
	"io"
	"time"
)

const (
	indexName = "flights"
	
	fieldDepAirport     = "departure_airport"
	fieldArrAirport     = "arrival_airport"
	fieldDepTime        = "departure_time"
	fieldPriceCents     = "price_cents"
	fieldAvailableSeats = "available_seats"
)

type flightDocument struct {
	ID               int64     `json:"id"`
	DepartureAirport string    `json:"departure_airport"`
	ArrivalAirport   string    `json:"arrival_airport"`
	DepartureTime    time.Time `json:"departure_time"`
	PriceCents       int64     `json:"price_cents"`
	AvailableSeats   int       `json:"available_seats"`
}

type FlightSearchRepo struct {
	client *elasticsearch.Client
}

func NewFlightSearchRepo(url string) (*FlightSearchRepo, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating elastic client: %w", err)
	}

	if _, err := es.Info(); err != nil {
		return nil, fmt.Errorf("elastic connection check: %w", err)
	}

	return &FlightSearchRepo{client: es}, nil
}

func (r *FlightSearchRepo) IndexFlight(ctx context.Context, f *domain.Flight) error {
	doc := flightDocument{
		ID:               f.ID,
		DepartureAirport: f.DepartureAirport,
		ArrivalAirport:   f.ArrivalAirport,
		DepartureTime:    f.DepartureTime,
		PriceCents:       f.PriceCents,
		AvailableSeats:   f.AvailableSeats,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal doc: %w", err)
	}

	res, err := r.client.Index(
		indexName,
		bytes.NewReader(data),
		r.client.Index.WithDocumentID(fmt.Sprintf("%d", f.ID)),
		r.client.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elastic index request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elastic index response error: %s", res.String())
	}
	return nil
}

func (r *FlightSearchRepo) UpdateAvailableSeats(ctx context.Context, flightID int64, newCount int) error {
	payload := fmt.Sprintf(`{"doc": {"%s": %d}}`, fieldAvailableSeats, newCount)

	res, err := r.client.Update(
		indexName,
		fmt.Sprintf("%d", flightID),
		bytes.NewReader([]byte(payload)),
		r.client.Update.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elastic update request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elastic update response error: %s", res.String())
	}
	return nil
}

func (r *FlightSearchRepo) Search(ctx context.Context, filter repository.SearchFilter) ([]domain.Flight, error) {
	query := r.buildSearchQuery(filter)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encode query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(indexName),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("elastic search request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elastic search response error: %s", res.String())
	}

	return r.parseSearchResponse(res.Body)
}

func (r *FlightSearchRepo) buildSearchQuery(f repository.SearchFilter) map[string]interface{} {
	dateStart := time.Date(f.Date.Year(), f.Date.Month(), f.Date.Day(), 0, 0, 0, 0, time.UTC)
	dateEnd := dateStart.Add(24 * time.Hour)

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"match": map[string]interface{}{fieldDepAirport: f.FromAirport}},
					{"match": map[string]interface{}{fieldArrAirport: f.ToAirport}},
					{"range": map[string]interface{}{
						fieldDepTime: map[string]interface{}{
							"gte": dateStart,
							"lt":  dateEnd,
						},
					}},
					{"range": map[string]interface{}{
						fieldAvailableSeats: map[string]interface{}{
							"gte": f.PassengerCount,
						},
					}},
				},
			},
		},
		"sort": []map[string]interface{}{
			{fieldPriceCents: "asc"},
		},
	}
}

func (r *FlightSearchRepo) parseSearchResponse(body io.ReadCloser) ([]domain.Flight, error) {
	var response struct {
		Hits struct {
			Hits []struct {
				Source flightDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("json decode response: %w", err)
	}

	flights := make([]domain.Flight, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		flights = append(flights, domain.Flight{
			ID:               hit.Source.ID,
			DepartureAirport: hit.Source.DepartureAirport,
			ArrivalAirport:   hit.Source.ArrivalAirport,
			DepartureTime:    hit.Source.DepartureTime,
			PriceCents:       hit.Source.PriceCents,
			AvailableSeats:   hit.Source.AvailableSeats,
		})
	}

	return flights, nil
}
