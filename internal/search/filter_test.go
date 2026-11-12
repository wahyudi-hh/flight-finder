package search

import (
	"testing"
	"time"

	"flight-finder/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestFilterFlights(t *testing.T) {
	flights := []domain.Flight{
		{
			ID:        "GA001",
			Airline:   domain.Airline{Code: "GA"},
			Departure: domain.Schedule{DateTime: time.Date(2026, time.January, 1, 8, 30, 0, 0, time.UTC)},
			Arrival:   domain.Schedule{DateTime: time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)},
			Duration:  domain.Duration{TotalMinutes: 90},
			Stops:     0,
			Price:     domain.Price{Amount: 100},
		},
		{
			ID:        "JT002",
			Airline:   domain.Airline{Code: "JT"},
			Departure: domain.Schedule{DateTime: time.Date(2026, time.January, 1, 11, 15, 0, 0, time.UTC)},
			Arrival:   domain.Schedule{DateTime: time.Date(2026, time.January, 1, 13, 15, 0, 0, time.UTC)},
			Duration:  domain.Duration{TotalMinutes: 120},
			Stops:     1,
			Price:     domain.Price{Amount: 200},
		},
		{
			ID:        "GA003",
			Airline:   domain.Airline{Code: "GA"},
			Departure: domain.Schedule{DateTime: time.Date(2026, time.January, 1, 18, 0, 0, 0, time.UTC)},
			Arrival:   domain.Schedule{DateTime: time.Date(2026, time.January, 1, 20, 30, 0, 0, time.UTC)},
			Duration:  domain.Duration{TotalMinutes: 150},
			Stops:     2,
			Price:     domain.Price{Amount: 300},
		},
	}

	int64Pointer := func(value int64) *int64 { return &value }
	intPointer := func(value int) *int { return &value }
	stringPointer := func(value string) *string { return &value }

	tests := []struct {
		name    string
		filter  FlightFilter
		wantIDs []string
	}{
		{
			name:    "no filters",
			filter:  FlightFilter{},
			wantIDs: []string{"GA001", "JT002", "GA003"},
		},
		{
			name:    "minimum price is inclusive",
			filter:  FlightFilter{MinPrice: int64Pointer(200)},
			wantIDs: []string{"JT002", "GA003"},
		},
		{
			name:    "maximum price is inclusive",
			filter:  FlightFilter{MaxPrice: int64Pointer(200)},
			wantIDs: []string{"GA001", "JT002"},
		},
		{
			name:    "price range",
			filter:  FlightFilter{MinPrice: int64Pointer(150), MaxPrice: int64Pointer(250)},
			wantIDs: []string{"JT002"},
		},
		{
			name:    "maximum stops is inclusive",
			filter:  FlightFilter{MaxStops: intPointer(1)},
			wantIDs: []string{"GA001", "JT002"},
		},
		{
			name:    "departure time range includes boundaries",
			filter:  FlightFilter{DepartureTimeFrom: stringPointer("08:30"), DepartureTimeTo: stringPointer("11:15")},
			wantIDs: []string{"GA001", "JT002"},
		},
		{
			name:    "arrival time range",
			filter:  FlightFilter{ArrivalTimeFrom: stringPointer("13:15"), ArrivalTimeTo: stringPointer("20:30")},
			wantIDs: []string{"JT002", "GA003"},
		},
		{
			name:    "airline matching ignores case",
			filter:  FlightFilter{Airlines: []string{"jt"}},
			wantIDs: []string{"JT002"},
		},
		{
			name:    "maximum duration is inclusive",
			filter:  FlightFilter{MaxDurationMinutes: intPointer(120)},
			wantIDs: []string{"GA001", "JT002"},
		},
		{
			name: "combined filters",
			filter: FlightFilter{
				MinPrice:           int64Pointer(150),
				MaxPrice:           int64Pointer(250),
				MaxStops:           intPointer(1),
				DepartureTimeFrom:  stringPointer("11:15"),
				DepartureTimeTo:    stringPointer("11:15"),
				ArrivalTimeFrom:    stringPointer("13:15"),
				ArrivalTimeTo:      stringPointer("13:15"),
				Airlines:           []string{"JT"},
				MaxDurationMinutes: intPointer(120),
			},
			wantIDs: []string{"JT002"},
		},
		{
			name:    "invalid departure time excludes flights",
			filter:  FlightFilter{DepartureTimeFrom: stringPointer("not-a-time")},
			wantIDs: []string{},
		},
		{
			name:    "invalid arrival time excludes flights",
			filter:  FlightFilter{ArrivalTimeTo: stringPointer("25:00")},
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterFlights(flights, tt.filter)

			gotIDs := make([]string, 0, len(got))
			for _, flight := range got {
				gotIDs = append(gotIDs, flight.ID)
			}

			assert.Equal(t, tt.wantIDs, gotIDs)
		})
	}
}
