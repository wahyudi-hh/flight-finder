package service

import (
	"context"
	"flight-finder/internal/aggregator"
	"flight-finder/internal/provider"
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
) (*aggregator.SearchResult, int64, error){
	start := time.Now()
	
	result, err := s.aggregator.Search(ctx, request)

	searchTimeMs := time.Since(start).Milliseconds();

	if err != nil {
		return nil, searchTimeMs, err
	}
	return result, searchTimeMs, nil
}