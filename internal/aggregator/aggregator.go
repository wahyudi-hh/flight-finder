package aggregator

import (
	"context"
	"errors"
	"flight-finder/internal/domain"
	"flight-finder/internal/provider"
	"fmt"
	"time"
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
			flights, err := searchWithRetry(ctx, currProvider, request)
			
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

const (
	maxAttempts = 2
	initialBackoff = 100 * time.Millisecond
	providerTimeout = 1 * time.Second
)

func searchWithRetry(
	ctx context.Context,
	currProvider provider.Provider,
	request provider.SearchRequest,
) ([]domain.Flight, error) {
	var flights []domain.Flight
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		flights, err = searchWithTimeout(ctx, currProvider, request)
		if err == nil {
			return flights, nil
		}

		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if attempt == maxAttempts {
			break;
		}
		backoff := initialBackoff * time.Duration(1 << (attempt - 1))

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, err
}

func searchWithTimeout(
    ctx context.Context,
    currProvider provider.Provider,
    request provider.SearchRequest,
) ([]domain.Flight, error) {
    attemptCtx, cancel := context.WithTimeout(ctx, providerTimeout)
    defer cancel()

    return currProvider.Search(attemptCtx, request)
}