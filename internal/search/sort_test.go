package search

import (
	"testing"

	"flight-finder/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestSortFlights(t *testing.T) {
	flights := []domain.Flight{
		{
			ID:        "middle",
			Price:     domain.Price{Amount: 200},
			Duration:  domain.Duration{TotalMinutes: 120},
			Departure: domain.Schedule{Timestamp: 200},
			Arrival:   domain.Schedule{Timestamp: 500},
		},
		{
			ID:        "first",
			Price:     domain.Price{Amount: 100},
			Duration:  domain.Duration{TotalMinutes: 60},
			Departure: domain.Schedule{Timestamp: 100},
			Arrival:   domain.Schedule{Timestamp: 300},
		},
		{
			ID:        "last",
			Price:     domain.Price{Amount: 300},
			Duration:  domain.Duration{TotalMinutes: 180},
			Departure: domain.Schedule{Timestamp: 300},
			Arrival:   domain.Schedule{Timestamp: 700},
		},
	}

	tests := []struct {
		name    string
		sort    FlightSort
		wantIDs []string
	}{
		{
			name:    "price ascending",
			sort:    FlightSort{By: SortByPrice, Order: Ascending},
			wantIDs: []string{"first", "middle", "last"},
		},
		{
			name:    "price descending",
			sort:    FlightSort{By: SortByPrice, Order: Descending},
			wantIDs: []string{"last", "middle", "first"},
		},
		{
			name:    "duration ascending",
			sort:    FlightSort{By: SortByDuration, Order: Ascending},
			wantIDs: []string{"first", "middle", "last"},
		},
		{
			name:    "duration descending",
			sort:    FlightSort{By: SortByDuration, Order: Descending},
			wantIDs: []string{"last", "middle", "first"},
		},
		{
			name:    "departure ascending",
			sort:    FlightSort{By: SortByDeparture, Order: Ascending},
			wantIDs: []string{"first", "middle", "last"},
		},
		{
			name:    "departure descending",
			sort:    FlightSort{By: SortByDeparture, Order: Descending},
			wantIDs: []string{"last", "middle", "first"},
		},
		{
			name:    "arrival ascending",
			sort:    FlightSort{By: SortByArrival, Order: Ascending},
			wantIDs: []string{"first", "middle", "last"},
		},
		{
			name:    "arrival descending",
			sort:    FlightSort{By: SortByArrival, Order: Descending},
			wantIDs: []string{"last", "middle", "first"},
		},
		{
			name:    "best value",
			sort:    FlightSort{By: SortByBestValue, Order: Ascending},
			wantIDs: []string{"first", "middle", "last"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortFlights(flights, tt.sort)

			gotIDs := make([]string, 0, len(got))
			for _, flight := range got {
				gotIDs = append(gotIDs, flight.ID)
			}

			assert.Equal(t, tt.wantIDs, gotIDs)
		})
	}
}

func TestSortFlights_BestValueRanksPriceAndDuration(t *testing.T) {
	flights := []domain.Flight{
		{ID: "price-value", Price: domain.Price{Amount: 100}, Duration: domain.Duration{TotalMinutes: 300}},
		{ID: "duration-value", Price: domain.Price{Amount: 300}, Duration: domain.Duration{TotalMinutes: 100}},
		{ID: "best", Price: domain.Price{Amount: 100}, Duration: domain.Duration{TotalMinutes: 100}},
		{ID: "worst", Price: domain.Price{Amount: 300}, Duration: domain.Duration{TotalMinutes: 300}},
	}

	got := SortFlights(flights, FlightSort{By: SortByBestValue, Order: Descending})

	gotIDs := make([]string, 0, len(got))
	for _, flight := range got {
		gotIDs = append(gotIDs, flight.ID)
	}

	assert.Equal(t, []string{"best", "price-value", "duration-value", "worst"}, gotIDs)
}

func TestSortFlights_EmptyInput(t *testing.T) {
	assert.Empty(t, SortFlights(nil, FlightSort{By: SortByPrice, Order: Ascending}))
	assert.Empty(t, SortFlights(nil, FlightSort{By: SortByBestValue, Order: Ascending}))
}

func TestFlightSort_IsValid(t *testing.T) {
	tests := []struct {
		name string
		sort FlightSort
		want bool
	}{
		{name: "best value ascending", sort: FlightSort{By: SortByBestValue, Order: Ascending}, want: true},
		{name: "price descending", sort: FlightSort{By: SortByPrice, Order: Descending}, want: true},
		{name: "duration ascending", sort: FlightSort{By: SortByDuration, Order: Ascending}, want: true},
		{name: "departure descending", sort: FlightSort{By: SortByDeparture, Order: Descending}, want: true},
		{name: "arrival ascending", sort: FlightSort{By: SortByArrival, Order: Ascending}, want: true},
		{name: "unknown sort key", sort: FlightSort{By: "unknown", Order: Ascending}, want: false},
		{name: "unknown sort order", sort: FlightSort{By: SortByPrice, Order: "sideways"}, want: false},
		{name: "missing sort key", sort: FlightSort{Order: Ascending}, want: false},
		{name: "missing sort order", sort: FlightSort{By: SortByPrice}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sort.IsValid())
		})
	}
}
