package garuda

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchResponse_ToFlightDomain_WithoutSegments(t *testing.T) {
	response := SearchResponse{
		Flights: []FlightData{
			{
				FlightId:        "GA-123",
				Airline:         "Garuda Indonesia",
				AirlineCode:     "GA",
				Departure:       Schedule{Airport: "CGK", City: "Jakarta", Time: "2026-01-02T08:30:00+07:00"},
				Arrival:         Schedule{Airport: "DPS", City: "Denpasar", Time: "2026-01-02T11:00:00+07:00"},
				DurationMinutes: 150,
				Stops:           0,
				Aircraft:        "Boeing 737",
				Price:           Price{Amount: 1750000, Currency: "IDR"},
				AvailableSeats:  10,
				FareClass:       "Economy",
				Baggage:         Baggage{CarryOn: 7, Checked: 20},
				Amenities:       []string{"Meal", "Wi-Fi"},
			},
		},
	}

	got, err := response.ToFlightDomain()

	require.NoError(t, err)
	require.Len(t, got, 1)

	flight := got[0]
	assert.Equal(t, "GA-123_GarudaIndonesia", flight.ID)
	assert.Equal(t, "Garuda Indonesia", flight.Provider)
	assert.Equal(t, "Garuda Indonesia", flight.Airline.Name)
	assert.Equal(t, "GA", flight.Airline.Code)
	assert.Equal(t, "GA-123", flight.FlightNumber)
	assert.Equal(t, "CGK", flight.Departure.Airport)
	assert.Equal(t, "Jakarta", flight.Departure.City)
	assert.True(t, time.Date(2026, time.January, 2, 8, 30, 0, 0, time.FixedZone("", 7*60*60)).Equal(flight.Departure.DateTime))
	assert.Equal(t, flight.Departure.DateTime.Unix(), flight.Departure.Timestamp)
	assert.Equal(t, "DPS", flight.Arrival.Airport)
	assert.Equal(t, "Denpasar", flight.Arrival.City)
	assert.Equal(t, 150, flight.Duration.TotalMinutes)
	assert.Equal(t, "2h 30m", flight.Duration.Formatted)
	assert.Zero(t, flight.Stops)
	assert.Equal(t, int64(1750000), flight.Price.Amount)
	assert.Equal(t, "IDR", flight.Price.Currency)
	assert.Equal(t, 10, flight.AvailableSeats)
	assert.Equal(t, "Economy", flight.CabinClass)
	require.NotNil(t, flight.Aircraft)
	assert.Equal(t, "Boeing 737", *flight.Aircraft)
	assert.Equal(t, []string{"Meal", "Wi-Fi"}, flight.Amenities)
	assert.Equal(t, "7", flight.Baggage.CarryOn)
	assert.Equal(t, "20", flight.Baggage.Checked)
}

func TestSearchResponse_ToFlightDomain_SegmentsOverrideSummary(t *testing.T) {
	layover := 45
	response := SearchResponse{Flights: []FlightData{
		{
			FlightId:        "GA-456",
			Airline:         "Garuda Indonesia",
			Departure:       Schedule{Airport: "CGK", City: "Jakarta", Time: "2026-01-02T08:00:00Z"},
			Arrival:         Schedule{Airport: "DPS", City: "Denpasar", Time: "2026-01-02T10:00:00Z"},
			DurationMinutes: 120,
			Stops:           0,
			Segments: []Segment{
				{
					FlightNumber:    "GA-456",
					DurationMinutes: 90,
					Departure:       Schedule{Airport: "CGK", City: "Jakarta", Time: "2026-01-02T08:00:00Z"},
					Arrival:         Schedule{Airport: "SUB", City: "Surabaya", Time: "2026-01-02T09:30:00Z"},
					LayoverMinutes:  &layover,
				},
				{
					FlightNumber:    "GA-789",
					DurationMinutes: 120,
					Departure:       Schedule{Airport: "SUB", City: "Surabaya", Time: "2026-01-02T10:15:00Z"},
					Arrival:         Schedule{Airport: "DPS", City: "Denpasar", Time: "2026-01-02T12:15:00Z"},
				},
			},
		},
	}}

	got, err := response.ToFlightDomain()

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].Stops)
	assert.Equal(t, 255, got[0].Duration.TotalMinutes)
	assert.Equal(t, "DPS", got[0].Arrival.Airport)
	assert.Equal(t, "Denpasar", got[0].Arrival.City)
	assert.Equal(t, time.Date(2026, time.January, 2, 12, 15, 0, 0, time.UTC), got[0].Arrival.DateTime)
}

func TestSearchResponse_ToFlightDomain_InvalidTime(t *testing.T) {
	tests := []struct {
		name   string
		flight FlightData
	}{
		{
			name: "departure",
			flight: FlightData{
				FlightId:  "GA-123",
				Departure: Schedule{Time: "invalid"},
				Arrival:   Schedule{Time: "2026-01-02T10:00:00Z"},
			},
		},
		{
			name: "arrival",
			flight: FlightData{
				FlightId:  "GA-123",
				Departure: Schedule{Time: "2026-01-02T08:00:00Z"},
				Arrival:   Schedule{Time: "invalid"},
			},
		},
		{
			name: "last segment arrival",
			flight: FlightData{
				FlightId:  "GA-123",
				Departure: Schedule{Time: "2026-01-02T08:00:00Z"},
				Arrival:   Schedule{Time: "2026-01-02T10:00:00Z"},
				Segments: []Segment{{
					FlightNumber: "GA-123",
					Arrival:      Schedule{Time: "invalid"},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (SearchResponse{Flights: []FlightData{tt.flight}}).ToFlightDomain()

			require.Error(t, err)
			assert.Nil(t, got)
		})
	}
}
