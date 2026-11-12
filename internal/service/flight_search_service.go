package service

import (
	"context"
	"flight-finder/internal/aggregator"
	"flight-finder/internal/provider"
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
) (*aggregator.SearchResult, error){
	return s.aggregator.Search(ctx, request)
}