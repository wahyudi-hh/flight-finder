package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"flight-finder/internal/aggregator"
	"flight-finder/internal/cache"
	"flight-finder/internal/domain"
	"flight-finder/internal/provider"
	"flight-finder/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type handlerTestProvider struct {
	flights []domain.Flight
	err     error
	calls   atomic.Int32
	request provider.SearchRequest
}

func (p *handlerTestProvider) Name() string {
	return "test provider"
}

func (p *handlerTestProvider) Search(_ context.Context, request provider.SearchRequest) ([]domain.Flight, error) {
	p.calls.Add(1)
	p.request = request
	return p.flights, p.err
}

func newTestFlightHandler(providers ...provider.Provider) *FlightHandler {
	flightAggregator := aggregator.NewAggregator(providers...)
	flightCache := cache.NewInMemoryCache()
	flightService := service.NewFlightSearchService(flightAggregator, flightCache)
	return NewFlightHandler(flightService)
}

func TestFlightHandler_SearchSuccess(t *testing.T) {
	fakeProvider := &handlerTestProvider{
		flights: []domain.Flight{
			{ID: "expensive", Price: domain.Price{Amount: 200}},
			{ID: "cheap", Price: domain.Price{Amount: 100}},
		},
	}
	handler := newTestFlightHandler(fakeProvider)
	query := url.Values{
		"origin":        {"CGK"},
		"destination":   {"DPS"},
		"departureDate": {"2026-12-01"},
		"returnDate":    {"2026-12-10"},
		"passengers":    {"2"},
		"cabinClass":    {"economy"},
		"minPrice":      {"100"},
		"maxPrice":      {"200"},
		"sortBy":        {"price"},
		"sortOrder":     {"asc"},
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/flights/search?"+query.Encode(), nil)
	response := httptest.NewRecorder()

	handler.Search(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
	var body FlightSearchResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, FlightSearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2026-12-01",
		ReturnDate:    "2026-12-10",
		Passengers:    2,
		CabinClass:    "economy",
	}, body.SearchCriteria)
	assert.Equal(t, 2, body.Metadata.TotalResults)
	assert.Equal(t, 1, body.Metadata.ProvidersQueried)
	assert.Equal(t, 1, body.Metadata.ProvidersSucceeded)
	assert.Zero(t, body.Metadata.ProvidersFailed)
	assert.False(t, body.Metadata.CacheHit)
	require.Len(t, body.Flights, 2)
	assert.Equal(t, []string{"cheap", "expensive"}, []string{body.Flights[0].ID, body.Flights[1].ID})
	require.NotNil(t, fakeProvider.request.ReturnDate)
	assert.Equal(t, "2026-12-10", *fakeProvider.request.ReturnDate)
	assert.Equal(t, 2, fakeProvider.request.Passengers)

	response = httptest.NewRecorder()
	handler.Search(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.True(t, body.Metadata.CacheHit)
	assert.EqualValues(t, 1, fakeProvider.calls.Load())
}

func TestFlightHandler_SearchBadRequest(t *testing.T) {
	tests := []struct {
		name  string
		query url.Values
		body  string
	}{
		{
			name:  "invalid passenger count",
			query: url.Values{"passengers": {"many"}},
			body:  "passengers must be a valid integer",
		},
		{
			name:  "missing required fields",
			query: url.Values{"passengers": {"1"}},
			body:  "origin is required",
		},
		{
			name: "invalid filter",
			query: url.Values{
				"origin":        {"CGK"},
				"destination":   {"DPS"},
				"departureDate": {"2026-12-01"},
				"passengers":    {"1"},
				"minPrice":      {"bad"},
			},
			body: "minPrice must be a non-negative integer",
		},
		{
			name: "invalid sort",
			query: url.Values{
				"origin":        {"CGK"},
				"destination":   {"DPS"},
				"departureDate": {"2026-12-01"},
				"passengers":    {"1"},
				"sortBy":        {"unknown"},
			},
			body: "invalid sorting parameter(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestFlightHandler(&handlerTestProvider{})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/flights/search?"+tt.query.Encode(), nil)
			response := httptest.NewRecorder()

			handler.Search(response, request)

			assert.Equal(t, http.StatusBadRequest, response.Code)
			assert.Contains(t, response.Body.String(), tt.body)
		})
	}
}

func TestFlightHandler_SearchServiceUnavailable(t *testing.T) {
	handler := newTestFlightHandler()
	query := url.Values{
		"origin":        {"CGK"},
		"destination":   {"DPS"},
		"departureDate": {"2026-12-01"},
		"passengers":    {"1"},
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/flights/search?"+query.Encode(), nil)
	response := httptest.NewRecorder()

	handler.Search(response, request)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.Equal(t, "no flight providers configured\n", response.Body.String())
}

func TestFlightHandler_SearchProviderFailure(t *testing.T) {
	fakeProvider := &handlerTestProvider{err: errors.New("provider unavailable")}
	handler := newTestFlightHandler(fakeProvider)
	query := url.Values{
		"origin":        {"CGK"},
		"destination":   {"DPS"},
		"departureDate": {"2026-12-01"},
		"passengers":    {"1"},
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/flights/search?"+query.Encode(), nil)
	response := httptest.NewRecorder()

	handler.Search(response, request)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.Equal(t, "all flight providers failed\n", response.Body.String())
}
