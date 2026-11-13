package service

import (
	"context"
	"flight-finder/internal/aggregator"
	"flight-finder/internal/provider"
	"flight-finder/internal/search"
	"time"
)

type FlightSearchService struct {
	aggregator *aggregator.Aggregator
}

func NewFlightSearchService(aggregator *aggregator.Aggregator) *FlightSearchService {
	return &FlightSearchService{
		aggregator: aggregator,
	}
}

func (s *FlightSearchService) Search(
	ctx context.Context,
	request provider.SearchRequest,
	filter search.FlightFilter,
	sort search.FlightSort,
) (*aggregator.SearchResult, int64, error){
	start := time.Now()
	
	result, err := s.aggregator.Search(ctx, request)

	searchTimeMs := time.Since(start).Milliseconds();

	if err != nil {
		return nil, searchTimeMs, err
	}

	result.Flights = search.FilterFlights(result.Flights, filter)
	result.Flights = search.SortFlights(result.Flights, sort)

	return result, searchTimeMs, nil
}