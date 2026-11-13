package aggregator

import (
	"context"
	"errors"
	"flight-finder/internal/domain"
	"flight-finder/internal/provider"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	name    string
	flights []domain.Flight
	err     error
	calls   int
	search  func(context.Context, provider.SearchRequest) ([]domain.Flight, error)
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Search(ctx context.Context, request provider.SearchRequest) ([]domain.Flight, error) {
	m.calls++
	if m.search != nil {
		return m.search(ctx, request)
	}
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
	assert.Equal(t, 2, result.TotalProviders)
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
		err:  errors.New("provider unavailable"),
	}

	aggregator := NewAggregator(garuda, lion)

	result, err := aggregator.Search(context.Background(), provider.SearchRequest{})

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Len(t, result.Flights, 1)
	assert.Equal(t, "GA001", result.Flights[0].ID)
	assert.Equal(t, 1, result.FailedProviders)
	assert.Equal(t, 2, result.TotalProviders)
	assert.Equal(t, 2, lion.calls, "failed providers should be retried once")
}

func TestAggregator_Search_AllProvidersFail(t *testing.T) {
	garuda := &mockProvider{
		name: "Garuda Indonesia",
		err:  errors.New("provider unavailable"),
	}

	lion := &mockProvider{
		name: "Lion Air",
		err:  errors.New("provider unavailable"),
	}

	aggregator := NewAggregator(garuda, lion)

	result, err := aggregator.Search(context.Background(), provider.SearchRequest{})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "all flight providers failed")
	assert.Equal(t, 2, garuda.calls)
	assert.Equal(t, 2, lion.calls)
}

func TestAggregator_Search_RetriesAndSucceeds(t *testing.T) {
	transient := errors.New("temporary provider failure")
	attempts := 0
	mock := &mockProvider{
		name: "recovering provider",
		search: func(ctx context.Context, request provider.SearchRequest) ([]domain.Flight, error) {
			attempts++
			_, hasDeadline := ctx.Deadline()
			if !hasDeadline {
				t.Error("provider call has no timeout deadline")
			}
			if request.Origin != "CGK" {
				t.Errorf("provider request origin = %q, want CGK", request.Origin)
			}
			if attempts == 1 {
				return nil, transient
			}
			return []domain.Flight{{ID: "recovered"}}, nil
		},
	}

	result, err := NewAggregator(mock).Search(context.Background(), provider.SearchRequest{Origin: "CGK"})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Flights, 1)
	assert.Equal(t, "recovered", result.Flights[0].ID)
	assert.Zero(t, result.FailedProviders)
	assert.Equal(t, 1, result.TotalProviders)
	assert.Equal(t, 2, mock.calls)
}

func TestAggregator_Search_NoProviders(t *testing.T) {
	result, err := NewAggregator().Search(context.Background(), provider.SearchRequest{})

	require.Error(t, err)
	assert.EqualError(t, err, "no flight providers configured")
	assert.Nil(t, result)
}

func TestAggregator_Search_CancellationDuringRetryBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	firstAttempt := make(chan struct{})
	mock := &mockProvider{
		name: "provider",
		search: func(context.Context, provider.SearchRequest) ([]domain.Flight, error) {
			close(firstAttempt)
			return nil, errors.New("temporary provider failure")
		},
	}
	done := make(chan error, 1)

	go func() {
		_, err := NewAggregator(mock).Search(ctx, provider.SearchRequest{})
		done <- err
	}()

	select {
	case <-firstAttempt:
	case <-time.After(time.Second):
		t.Fatal("provider was not called")
	}
	cancel()

	select {
	case err := <-done:
		require.EqualError(t, err, "all flight providers failed")
	case <-time.After(time.Second):
		t.Fatal("search did not stop after cancellation")
	}
	assert.Equal(t, 1, mock.calls, "cancellation during backoff should prevent a retry")
}
