package airasia

import (
	"context"
	"flight-finder/internal/provider"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Name(t *testing.T) {
	client := &Client{failureRate: 0}
	assert.Equal(t, "AirAsia", client.Name())
}

func TestClient_Search_Success(t *testing.T) {
	client := &Client{failureRate: 0}

	response, err := client.Search(context.Background(), provider.SearchRequest{})

	require.NoError(t, err)
	require.Len(t, response, 4)
	assert.Equal(t, "QZ520", response[0].FlightNumber)
	assert.Equal(t, "AirAsia", response[0].Provider)
	assert.Equal(t, "CGK", response[0].Departure.Airport)
	assert.Equal(t, "DPS", response[0].Arrival.Airport)

	assert.NotEmpty(t, response)
}

func TestClient_Search_Failure(t *testing.T) {
	client := &Client{failureRate: 1}

	response, err := client.Search(context.Background(), provider.SearchRequest{})

	require.Error(t, err)
	require.Nil(t, response)
	assert.EqualError(t, err, "AirAsia provider request failed")
}

func TestNewClient_DefaultFailureRate(t *testing.T) {
	client := NewClient()

	assert.Equal(t, 0.10, client.failureRate)
}

func TestClient_Search_ContextCancelled(t *testing.T) {
	client := &Client{failureRate: 0}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	response, err := client.Search(ctx, provider.SearchRequest{})

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, response)
}