package garuda

import (
	"context"
	"flight-finder/internal/provider"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Search_Success(t *testing.T) {
    client := NewClient()

    response, err := client.Search(context.Background(), provider.SearchRequest{})

    require.NoError(t, err)
    require.NotNil(t, response)

    assert.Equal(t, "success", response.Status)
    assert.NotEmpty(t, response.Flights)
}

func TestClient_Search_ContextCancelled(t *testing.T) {
    client := NewClient()

    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    response, err := client.Search(ctx, provider.SearchRequest{})

    require.ErrorIs(t, err, context.Canceled)
    assert.Nil(t, response)
}