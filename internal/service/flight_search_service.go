package service

import (
	"context"
	"encoding/json"
	"flight-finder/internal/aggregator"
	"flight-finder/internal/cache"
	"flight-finder/internal/domain"
	"flight-finder/internal/provider"
	"flight-finder/internal/search"
	"time"
)

const flightSearchCacheTTL = 5 * time.Minute

type FlightSearchService struct {
	aggregator 	*aggregator.Aggregator
	cache 		cache.Cache
}

func NewFlightSearchService(aggregator *aggregator.Aggregator, cache cache.Cache) *FlightSearchService {
	return &FlightSearchService{
		aggregator: aggregator,
		cache: cache,
	}
}

func (s *FlightSearchService) Search(
	ctx context.Context,
	request provider.SearchRequest,
	filter search.FlightFilter,
	sort search.FlightSort,
) (*aggregator.SearchResult, int64, bool, error){
	start := time.Now()

	cacheKey, err := generateCacheKey(request)
	if err != nil {
		return nil, int64(time.Since(start).Milliseconds()), false, err
	}
	
	var result *aggregator.SearchResult
	cacheHit := false
	cached, exist := s.cache.Get(cacheKey)
	if exist {
		if res, valid := cached.(*aggregator.SearchResult); valid {
			result = cloneSearchResult(res)
			cacheHit = true
		}
	}

	if !cacheHit {
    	result, err = s.aggregator.Search(ctx, request)
		if err != nil {
			return nil, time.Since(start).Milliseconds(), false, err
    	}
		if result != nil && result.FailedProviders == 0 {
			s.cache.Set(cacheKey, cloneSearchResult(result), flightSearchCacheTTL)
		}
	}

	result.Flights = search.FilterFlights(result.Flights, filter)
	result.Flights = search.SortFlights(result.Flights, sort)

	return result, time.Since(start).Milliseconds(), cacheHit, nil
}

func generateCacheKey(request provider.SearchRequest) (string, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func cloneSearchResult(result *aggregator.SearchResult) *aggregator.SearchResult {
	if result == nil {
		return nil
	}

	clone := *result
	clone.Flights = append([]domain.Flight(nil), result.Flights...)

	return &clone
}