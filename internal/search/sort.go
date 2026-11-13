package search

import (
	"cmp"
	"flight-finder/internal/domain"
	"slices"
)

type SortBy string
type SortOrder string

const (
	Price     SortBy = "price"
	Duration  SortBy = "duration"
	Departure SortBy = "departure"
	Arrival   SortBy = "arrival"
)

const (
	Ascending  SortOrder = "asc"
	Descending SortOrder = "desc"
)

type FlightSort struct {
	By    SortBy
	Order SortOrder
}

func SortFlights(flights []domain.Flight, sort FlightSort) []domain.Flight {
	sorted := make([]domain.Flight, len(flights))

	copy(sorted, flights)

	slices.SortFunc(sorted, func(a, b domain.Flight) int {
		var comparison int

		switch sort.By {
		case Price : comparison = cmp.Compare(a.Price.Amount, b.Price.Amount)
		case Duration : comparison = cmp.Compare(a.Duration.TotalMinutes, b.Duration.TotalMinutes)
		case Departure : comparison = cmp.Compare(a.Departure.Timestamp, b.Departure.Timestamp)
		case Arrival : comparison = cmp.Compare(a.Arrival.Timestamp, b.Arrival.Timestamp)
		}

		if sort.Order == Descending {
			return -comparison
		}
		return comparison
	})
	return sorted
}

func (s FlightSort)IsValid() bool {
	switch s.By {
	case Price, Duration, Departure, Arrival:
	default: return false
	}

	switch s.Order {
	case Ascending, Descending: return true
	default: return false
	}
}