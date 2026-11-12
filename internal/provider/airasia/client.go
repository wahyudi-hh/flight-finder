package airasia

import (
	"context"
	"encoding/json"
	"flight-finder/internal/provider"
	"flight-finder/internal/provider/mockdata"
	"fmt"
	"math/rand"
	"time"
)

const (
	minDelay = 50 * time.Millisecond
	maxDelay = 150 * time.Millisecond
	// failureRate = 0.10
)

type Client struct {
	failureRate float64
}

func NewClient() *Client{
	return &Client{
		failureRate: 0.10,
	}
}

func (c *Client) Name() string {
	return "AirAsia"
}

func (c *Client) Search (ctx context.Context, request provider.SearchRequest) (*SearchResponse, error) {
	delay := minDelay + time.Duration(rand.Int63n(int64(maxDelay - minDelay)))

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <- ctx.Done():
		return nil, ctx.Err()
	case <- timer.C:
	}

	if isSimulateFailure(c.failureRate) {
		return nil, fmt.Errorf("AirAsia provider request failed")
	}

	data, err := mockdata.Files.ReadFile("airasia_search_response.json")
	if err != nil {
		return nil, fmt.Errorf("read AirAsia mock response: %w", err)
	}

	var response SearchResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("unmarshal AirAsia response: %w", err)
	}

	return &response, nil
}

func isSimulateFailure(failureRate float64) bool {
	return rand.Float64() < failureRate
}