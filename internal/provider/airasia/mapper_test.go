package airasia

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchResponse_ToFlightDomain(t *testing.T) {
	response := SearchResponse{
		Flights: []FlightData{
			{
				FlightCode:    "QZ123",
				Airline:       "AirAsia",
				FromAirport:   "CGK",
				ToAirport:     "DPS",
				DepartTime:    "2026-01-02T08:30:00+07:00",
				ArriveTime:    "2026-01-02T11:00:00+07:00",
				DurationHours: 2.01,
				DirectFlight:  false,
				Stops:         []Stop{{Airport: "SUB"}, {Airport: "YIA"}},
				PriceIDR:      1250000,
				Seats:         8,
				CabinClass:    "economy",
			},
		},
	}

	got, err := response.ToFlightDomain()

	require.NoError(t, err)
	require.Len(t, got, 1)

	flight := got[0]
	assert.Equal(t, "QZ123_AirAsia", flight.ID)
	assert.Equal(t, "AirAsia", flight.Provider)
	assert.Equal(t, "AirAsia", flight.Airline.Name)
	assert.Equal(t, "QZ", flight.Airline.Code)
	assert.Equal(t, "QZ123", flight.FlightNumber)
	assert.Equal(t, "CGK", flight.Departure.Airport)
	assert.Equal(t, "Jakarta", flight.Departure.City)
	assert.True(t, time.Date(2026, time.January, 2, 8, 30, 0, 0, time.FixedZone("", 7*60*60)).Equal(flight.Departure.DateTime))
	assert.Equal(t, flight.Departure.DateTime.Unix(), flight.Departure.Timestamp)
	assert.Equal(t, "DPS", flight.Arrival.Airport)
	assert.Equal(t, "Denpasar", flight.Arrival.City)
	assert.Equal(t, 121, flight.Duration.TotalMinutes)
	assert.Equal(t, "2h 1m", flight.Duration.Formatted)
	assert.Equal(t, 2, flight.Stops)
	assert.Equal(t, int64(1250000), flight.Price.Amount)
	assert.Equal(t, "IDR", flight.Price.Currency)
	assert.Equal(t, 8, flight.AvailableSeats)
	assert.Equal(t, "economy", flight.CabinClass)
	assert.Nil(t, flight.Aircraft)
	assert.Empty(t, flight.Amenities)
	assert.Equal(t, "Cabin baggage only", flight.Baggage.CarryOn)
	assert.Equal(t, "Additional fee", flight.Baggage.Checked)
}

func TestSearchResponse_ToFlightDomain_DirectFlightHasNoStops(t *testing.T) {
	response := SearchResponse{Flights: []FlightData{
		{
			FlightCode:   "QZ456",
			Airline:      "AirAsia",
			DepartTime:   "2026-01-02T08:30:00Z",
			ArriveTime:   "2026-01-02T10:00:00Z",
			DirectFlight: true,
			Stops:        []Stop{{Airport: "SUB"}},
		},
	}}

	got, err := response.ToFlightDomain()

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Zero(t, got[0].Stops)
}

func TestSearchResponse_ToFlightDomain_InvalidTime(t *testing.T) {
	tests := []struct {
		name   string
		flight FlightData
	}{
		{
			name:   "departure",
			flight: FlightData{FlightCode: "QZ123", DepartTime: "invalid", ArriveTime: "2026-01-02T10:00:00Z"},
		},
		{
			name:   "arrival",
			flight: FlightData{FlightCode: "QZ123", DepartTime: "2026-01-02T08:00:00Z", ArriveTime: "invalid"},
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

func TestExtractAirlineCode(t *testing.T) {
	tests := []struct {
		flightCode string
		want       string
	}{
		{flightCode: "QZ123", want: "QZ"},
		{flightCode: "GA-456", want: "GA-"},
		{flightCode: "  Airline  ", want: "Airline"},
	}

	for _, tt := range tests {
		t.Run(tt.flightCode, func(t *testing.T) {
			assert.Equal(t, tt.want, extractAirlineCode(tt.flightCode))
		})
	}
}
