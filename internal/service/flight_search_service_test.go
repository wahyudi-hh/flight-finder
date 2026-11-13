package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"flight-finder/internal/aggregator"
	"flight-finder/internal/cache"
	"flight-finder/internal/domain"
	"flight-finder/internal/provider"
	"flight-finder/internal/search"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serviceTestProvider struct {
	name    string
	flights []domain.Flight
	err     error
	calls   int
	request provider.SearchRequest
	search  func(context.Context, provider.SearchRequest) ([]domain.Flight, error)
}

func (p *serviceTestProvider) Name() string {
	return p.name
}

func (p *serviceTestProvider) Search(ctx context.Context, request provider.SearchRequest) ([]domain.Flight, error) {
	p.calls++
	p.request = request
	if p.search != nil {
		return p.search(ctx, request)
	}
	return p.flights, p.err
}

func newTestFlightSearchService(providers ...provider.Provider) *FlightSearchService {
	return NewFlightSearchService(
		aggregator.NewAggregator(providers...),
		cache.NewInMemoryCache(),
	)
}

func TestFlightSearchService_SearchFiltersSortsAndCaches(t *testing.T) {
	request := provider.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2026-12-01",
		Passengers:    1,
	}
	flights := []domain.Flight{
		{ID: "cheap", Price: domain.Price{Amount: 100}},
		{ID: "middle", Price: domain.Price{Amount: 200}},
		{ID: "expensive", Price: domain.Price{Amount: 300}},
	}
	fakeProvider := &serviceTestProvider{name: "test", flights: flights}
	flightService := newTestFlightSearchService(fakeProvider)

	minPrice := int64(150)
	filter := search.FlightFilter{MinPrice: &minPrice}
	sort := search.FlightSort{By: search.SortByPrice, Order: search.Descending}

	firstResult, elapsed, cacheHit, err := flightService.Search(context.Background(), request, filter, sort)

	require.NoError(t, err)
	require.NotNil(t, firstResult)
	assert.GreaterOrEqual(t, elapsed, int64(0))
	assert.False(t, cacheHit)
	assert.Equal(t, 1, firstResult.TotalProviders)
	assert.Zero(t, firstResult.FailedProviders)
	require.Len(t, firstResult.Flights, 2)
	assert.Equal(t, []string{"expensive", "middle"}, []string{firstResult.Flights[0].ID, firstResult.Flights[1].ID})
	assert.Equal(t, request, fakeProvider.request)
	assert.Equal(t, []string{"cheap", "middle", "expensive"}, []string{flights[0].ID, flights[1].ID, flights[2].ID})

	firstResult.Flights[0].ID = "mutated response"
	maxPrice := int64(200)
	cachedResult, _, cacheHit, err := flightService.Search(
		context.Background(),
		request,
		search.FlightFilter{MaxPrice: &maxPrice},
		search.FlightSort{By: search.SortByPrice, Order: search.Ascending},
	)

	require.NoError(t, err)
	require.NotNil(t, cachedResult)
	assert.True(t, cacheHit)
	assert.Equal(t, 1, fakeProvider.calls)
	require.Len(t, cachedResult.Flights, 2)
	assert.Equal(t, []string{"cheap", "middle"}, []string{cachedResult.Flights[0].ID, cachedResult.Flights[1].ID})
}

func TestFlightSearchService_SearchDoesNotCachePartialFailure(t *testing.T) {
	request := provider.SearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2026-12-01", Passengers: 1}
	successfulProvider := &serviceTestProvider{
		name:    "working",
		flights: []domain.Flight{{ID: "available"}},
	}
	failedProvider := &serviceTestProvider{name: "unavailable", err: errors.New("provider unavailable")}
	flightService := newTestFlightSearchService(successfulProvider, failedProvider)
	sort := search.FlightSort{By: search.SortByPrice, Order: search.Ascending}

	for i := 0; i < 2; i++ {
		result, _, cacheHit, err := flightService.Search(context.Background(), request, search.FlightFilter{}, sort)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, cacheHit)
		assert.Equal(t, 1, result.FailedProviders)
		assert.Equal(t, 2, result.TotalProviders)
		assert.Equal(t, []string{"available"}, []string{result.Flights[0].ID})
	}

	assert.Equal(t, 2, successfulProvider.calls)
	assert.Equal(t, 4, failedProvider.calls, "failed provider should be retried on each search")
}

func TestFlightSearchService_SearchAllProvidersFail(t *testing.T) {
	fakeProvider := &serviceTestProvider{name: "unavailable", err: errors.New("provider unavailable")}
	flightService := newTestFlightSearchService(fakeProvider)

	result, elapsed, cacheHit, err := flightService.Search(
		context.Background(),
		provider.SearchRequest{Origin: "CGK"},
		search.FlightFilter{},
		search.FlightSort{By: search.SortByBestValue, Order: search.Ascending},
	)

	require.Error(t, err)
	assert.EqualError(t, err, "all flight providers failed")
	assert.Nil(t, result)
	assert.GreaterOrEqual(t, elapsed, int64(0))
	assert.False(t, cacheHit)
	assert.Equal(t, 2, fakeProvider.calls, "provider should be retried once")
}

func TestFlightSearchService_SearchNoProviders(t *testing.T) {
	flightService := newTestFlightSearchService()

	result, _, cacheHit, err := flightService.Search(
		context.Background(),
		provider.SearchRequest{},
		search.FlightFilter{},
		search.FlightSort{By: search.SortByBestValue, Order: search.Ascending},
	)

	require.Error(t, err)
	assert.EqualError(t, err, "no flight providers configured")
	assert.Nil(t, result)
	assert.False(t, cacheHit)
}

func TestFlightSearchService_SearchRetriesTransientProviderError(t *testing.T) {
	request := provider.SearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2026-12-01", Passengers: 1}
	attempts := 0
	fakeProvider := &serviceTestProvider{
		name: "recovering provider",
		search: func(_ context.Context, _ provider.SearchRequest) ([]domain.Flight, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("temporary provider failure")
			}
			return []domain.Flight{{ID: "recovered"}}, nil
		},
	}
	flightService := newTestFlightSearchService(fakeProvider)

	result, _, cacheHit, err := flightService.Search(
		context.Background(),
		request,
		search.FlightFilter{},
		search.FlightSort{By: search.SortByBestValue, Order: search.Ascending},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, cacheHit)
	assert.Equal(t, 2, fakeProvider.calls)
	assert.Zero(t, result.FailedProviders)
	require.Len(t, result.Flights, 1)
	assert.Equal(t, "recovered", result.Flights[0].ID)
}

func TestCloneSearchResult(t *testing.T) {
	assert.Nil(t, cloneSearchResult(nil))

	original := &aggregator.SearchResult{
		Flights:         []domain.Flight{{ID: "flight"}},
		FailedProviders: 1,
		TotalProviders:  2,
	}
	clone := cloneSearchResult(original)

	require.NotNil(t, clone)
	assert.Equal(t, original, clone)
	clone.Flights[0].ID = "changed"
	assert.Equal(t, "flight", original.Flights[0].ID)
}

func TestFlightSearchService_SearchUsesCacheTTL(t *testing.T) {
	assert.Equal(t, 5*time.Minute, flightSearchCacheTTL)
}
