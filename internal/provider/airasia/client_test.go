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
    require.NotNil(t, response)

    assert.NotEmpty(t, response)
}

func TestClient_Search_Failure(t *testing.T) {
    client := &Client{failureRate: 1}

    response, err := client.Search(context.Background(), provider.SearchRequest{})

    require.Error(t, err)
    require.Nil(t, response)
}

func TestClient_Search_ContextCancelled(t *testing.T) {
    client := &Client{failureRate: 0}

    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    response, err := client.Search(ctx, provider.SearchRequest{})

    require.ErrorIs(t, err, context.Canceled)
    assert.Nil(t, response)
}