package search

import (
	"cmp"
	"flight-finder/internal/domain"
	"math"
	"slices"
)

type SortBy string
type SortOrder string

const (
	SortByBestValue	SortBy = "bestValue"
	SortByPrice     SortBy = "price"
	SortByDuration  SortBy = "duration"
	SortByDeparture	SortBy = "departure"
	SortByArrival   SortBy = "arrival"
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

	if SortByBestValue == sort.By {
		return sortByBestValue(sorted)
	}

	slices.SortFunc(sorted, func(a, b domain.Flight) int {
		var comparison int

		switch sort.By {
		case SortByPrice : comparison = cmp.Compare(a.Price.Amount, b.Price.Amount)
		case SortByDuration : comparison = cmp.Compare(a.Duration.TotalMinutes, b.Duration.TotalMinutes)
		case SortByDeparture : comparison = cmp.Compare(a.Departure.Timestamp, b.Departure.Timestamp)
		case SortByArrival : comparison = cmp.Compare(a.Arrival.Timestamp, b.Arrival.Timestamp)
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
	case SortByBestValue, SortByPrice, SortByDuration, SortByDeparture, SortByArrival:
	default: return false
	}

	switch s.Order {
	case Ascending, Descending: return true
	default: return false
	}
}

type RankedFlight struct {
    Flight domain.Flight
    Score  float64
}

func sortByBestValue(flights []domain.Flight) []domain.Flight {
	if len(flights) == 0 {
		return nil
	}
	
	ranked := scoreFlights(flights)
	sorted := make([]domain.Flight, 0, len(flights))

	slices.SortStableFunc(ranked, func(a, b RankedFlight) int {
		return cmp.Compare(b.Score, a.Score)
	})

	for _, scoredFlight := range ranked {
		sorted = append(sorted, scoredFlight.Flight)
	}
	return sorted
}

func scoreFlights(flights []domain.Flight) []RankedFlight {
	ranked := make([]RankedFlight, 0, len(flights))

	minPrice := int64(math.MaxInt64)
	maxPrice := int64(0)
	minDuration := math.MaxInt
	maxDuration := 0

	for _, flight := range flights {
		if minPrice > flight.Price.Amount {
			minPrice = flight.Price.Amount
		}
		if maxPrice < flight.Price.Amount {
			maxPrice = flight.Price.Amount
		}
		if minDuration > flight.Duration.TotalMinutes {
			minDuration = flight.Duration.TotalMinutes
		}
		if maxDuration < flight.Duration.TotalMinutes {
			maxDuration = flight.Duration.TotalMinutes
		}
	}

	for _, flight := range flights {
		priceScore := 1.0
		if maxPrice != minPrice {
			priceScore = (float64(maxPrice - flight.Price.Amount)) / float64(maxPrice - minPrice)
		}

		durationScore := 1.0
		if maxDuration != minDuration {
			durationScore = float64(maxDuration - flight.Duration.TotalMinutes) / float64(maxDuration - minDuration)
		}
		ranked = append(ranked, RankedFlight{
			Flight: flight,
			Score: (0.5 * priceScore) + (0.5 * durationScore),
		})
	}
	return ranked	
}