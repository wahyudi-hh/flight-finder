package batikair

import (
	"context"
	"flight-finder/internal/provider"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Name(t *testing.T) {
	client := NewClient()
	assert.Equal(t, "Batik Air", client.Name())
}

func TestClient_Search_Success(t *testing.T) {
	client := NewClient()

	response, err := client.Search(context.Background(), provider.SearchRequest{})

	require.NoError(t, err)
	require.Len(t, response, 3)
	assert.Equal(t, "ID6514", response[0].FlightNumber)
	assert.Equal(t, "Batik Air", response[0].Provider)
	assert.Equal(t, "CGK", response[0].Departure.Airport)
	assert.Equal(t, "DPS", response[0].Arrival.Airport)
}

func TestClient_Search_ContextCancelled(t *testing.T) {
	client := NewClient()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	response, err := client.Search(ctx, provider.SearchRequest{})

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, response)
}