package provider

import (
	"context"
	"flight-finder/internal/domain"
)

type SearchRequest struct {
	Origin			string
	Destination		string
	DepartureDate	string
	ReturnDate		*string
	Passengers		int
	CabinClass		string
}

type Provider interface {
	Name() string
	Search(ctx context.Context, request SearchRequest) ([]domain.Flight, error)
}