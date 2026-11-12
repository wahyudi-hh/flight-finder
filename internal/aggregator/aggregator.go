package aggregator

import (
	"context"
	"flight-finder/internal/domain"
	"flight-finder/internal/provider"
	"fmt"
)

type Aggregator struct {
	providers []provider.Provider
}

func NewAggregator(providers ...provider.Provider) *Aggregator {
	return &Aggregator{
		providers: providers,
	}
}

type providerResult struct {
	index 	int
	name	string
	flights []domain.Flight
	err 	error
}

type SearchResult struct {
	Flights 		[]domain.Flight
	FailedProviders	int
	TotalProviders	int
}

func (a *Aggregator) Search(ctx context.Context, request provider.SearchRequest) (*SearchResult, error) {
	if len(a.providers) == 0 {
		return nil, fmt.Errorf("no flight providers configured")
	}

	results := make(chan providerResult, len(a.providers))

	for i, p := range a.providers {
		go func (index int, currProvider provider.Provider) {
			flights, err := currProvider.Search(ctx, request)
			
			results <- providerResult{
				index: index,
				name: currProvider.Name(),
				flights: flights,
				err: err,
			}
		}(i, p)
	}

	providerResults := make([]providerResult, len(a.providers))

	for range a.providers {
		result := <-results
		providerResults[result.index] = result
	}

	allFlights := make([]domain.Flight, 0)
	failedProviders := make([]string, 0)

	for _, result := range providerResults {
		if result.err != nil {
			failedProviders = append(failedProviders, result.name)
			continue
		}
		allFlights = append(allFlights, result.flights...)
	}

	if len(failedProviders) == len(a.providers) {
		return nil, fmt.Errorf("all flight providers failed")
	}

	return &SearchResult {
		Flights: allFlights,
		FailedProviders: len(failedProviders),
		TotalProviders: len(a.providers),
	}, nil
}