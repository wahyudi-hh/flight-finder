package batikair

import (
	"context"
	"encoding/json"
	"flight-finder/internal/domain"
	"flight-finder/internal/provider"
	"flight-finder/internal/provider/mockdata"
	"fmt"
	"math/rand"
	"time"
)

const (
	minDelay = 200 * time.Millisecond
	maxDelay = 400 * time.Millisecond
)

type Client struct {}

func NewClient() *Client{
	return &Client{}
}

func (c *Client) Name() string {
	return "Batik Air"
}

func (c *Client) Search (ctx context.Context, request provider.SearchRequest) ([]domain.Flight, error) {
	delay := minDelay + time.Duration(rand.Int63n(int64(maxDelay - minDelay)))

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <- ctx.Done():
		return nil, ctx.Err()
	case <- timer.C:
	}

	data, err := mockdata.Files.ReadFile("batik_air_search_response.json")
	if err != nil {
		return nil, fmt.Errorf("read Batik Air mock response: %w", err)
	}

	response := &SearchResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("unmarshal Batik Air response: %w", err)
	}

	return response.ToFlightDomain()
}