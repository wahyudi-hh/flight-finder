package aggregator

import (
	"context"
	"errors"
	"flight-finder/internal/domain"
	"flight-finder/internal/provider"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	name    string
	flights []domain.Flight
	err		error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Search(ctx context.Context, request provider.SearchRequest) ([]domain.Flight, error) {
    return m.flights, m.err
}

func TestAggregator_Search_AllProvidersSucceed(t *testing.T) {
    garuda := &mockProvider{
        name: "Garuda Indonesia",
        flights: []domain.Flight{
            {ID: "GA001"},
            {ID: "GA002"},
        },
    }

    lion := &mockProvider{
        name: "Lion Air",
        flights: []domain.Flight{
            {ID: "JT001"},
        },
    }

    aggregator := NewAggregator(garuda, lion)

    result, err := aggregator.Search(context.Background(), provider.SearchRequest{})

    require.NoError(t, err)
    require.NotNil(t, result)
    assert.Len(t, result.Flights, 3)
    assert.Empty(t, result.FailedProviders)
}

func TestAggregator_Search_PartialFailure(t *testing.T) {
    garuda := &mockProvider{
        name: "Garuda Indonesia",
        flights: []domain.Flight{
            {ID: "GA001"},
        },
    }

    lion := &mockProvider{
        name: "Lion Air",
        err: errors.New("provider unavailable"),
    }

    aggregator := NewAggregator(garuda, lion)

    result, err := aggregator.Search(context.Background(), provider.SearchRequest{})

    require.NoError(t, err)
    require.NotNil(t, result)

    assert.Len(t, result.Flights, 1)
    assert.Equal(t, "GA001", result.Flights[0].ID)
    assert.Equal(t, []string{"Lion Air"}, result.FailedProviders)
}

func TestAggregator_Search_AllProvidersFail(t *testing.T) {
    garuda := &mockProvider{
        name: "Garuda Indonesia",
        err: errors.New("provider unavailable"),
    }

    lion := &mockProvider{
        name: "Lion Air",
        err: errors.New("provider unavailable"),
    }

    aggregator := NewAggregator(garuda, lion)

    result, err := aggregator.Search(context.Background(), provider.SearchRequest{})

    require.Error(t, err)
    assert.Nil(t, result)
    assert.EqualError(t, err, "all flight providers failed")
}