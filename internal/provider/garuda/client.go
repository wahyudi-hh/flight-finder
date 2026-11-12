package garuda

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
	minDelay = 50 * time.Millisecond
	maxDelay = 100 * time.Millisecond
)

type Client struct {}

func NewClient() *Client{
	return &Client{}
}

func (c *Client) Name() string {
	return "Garuda Indonesia"
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

	data, err := mockdata.Files.ReadFile("garuda_indonesia_search_response.json")
	if err != nil {
		return nil, fmt.Errorf("read Garuda mock response: %w", err)
	}

	response := &SearchResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("unmarshal Garuda response: %w", err)
	}

	return response.ToFlightDomain()
}