package lionair

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
	minDelay = 100 * time.Millisecond
	maxDelay = 200 * time.Millisecond
)

type Client struct {}

func NewClient() *Client{
	return &Client{}
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

	data, err := mockdata.Files.ReadFile("lion_air_search_response.json")
	if err != nil {
		return nil, fmt.Errorf("read Lion Air mock response: %w", err)
	}

	var response SearchResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("unmarshal Lion Air response: %w", err)
	}

	return &response, nil
}