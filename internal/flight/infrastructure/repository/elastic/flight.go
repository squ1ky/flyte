package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"io"
	"time"
)

const (
	indexName = "flights"
)

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

func (r *FlightSearchRepo) IndexFlight(ctx context.Context, doc *domain.FlightDocument) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal doc: %w", err)
	}

	res, err := r.client.Index(
		indexName,
		bytes.NewReader(data),
		r.client.Index.WithDocumentID(fmt.Sprintf("%d", doc.ID)),
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

func (r *FlightSearchRepo) Search(ctx context.Context, criteria domain.FlightSearchCriteria) ([]domain.FlightDocument, error) {
	query := r.buildSearchQuery(criteria)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encode query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(indexName),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithSize(100),
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

func (r *FlightSearchRepo) buildSearchQuery(c domain.FlightSearchCriteria) map[string]interface{} {
	mustConditions := []map[string]interface{}{}

	if !c.Date.IsZero() {
		dateStart := time.Date(c.Date.Year(), c.Date.Month(), c.Date.Day(), 0, 0, 0, 0, time.UTC)
		dateEnd := dateStart.Add(24 * time.Hour)
		mustConditions = append(mustConditions, map[string]interface{}{
			"range": map[string]interface{}{
				"departure_time": map[string]interface{}{
					"gte": dateStart,
					"lt":  dateEnd,
				},
			},
		})
	}

	if c.Origin != "" {
		mustConditions = append(mustConditions, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  c.Origin,
				"fields": []string{"departure.code", "departure.city", "departure.name"},
			},
		})
	}

	if c.Destination != "" {
		mustConditions = append(mustConditions, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  c.Destination,
				"fields": []string{"arrival.code", "arrival.city", "arrival.name"},
			},
		})
	}

	if c.NumPassengers > 0 {
		mustConditions = append(mustConditions, map[string]interface{}{
			"range": map[string]interface{}{
				"available_seats": map[string]interface{}{
					"gte": c.NumPassengers,
				},
			},
		})
	}

	mustConditions = append(mustConditions, map[string]interface{}{
		"terms": map[string]interface{}{
			"status": []string{string(domain.FlightStatusScheduled)},
		},
	})

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustConditions,
			},
		},
		"sort": []map[string]interface{}{
			{"min_price_cents": "asc"},
			{"departure_time": "asc"},
		},
	}
}

func (r *FlightSearchRepo) parseSearchResponse(body io.ReadCloser) ([]domain.FlightDocument, error) {
	var response struct {
		Hits struct {
			Hits []struct {
				Source domain.FlightDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("json decode response: %w", err)
	}

	docs := make([]domain.FlightDocument, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		docs = append(docs, hit.Source)
	}

	return docs, nil
}
