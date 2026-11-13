package batikair

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchResponse_ToFlightDomain(t *testing.T) {
	response := SearchResponse{
		Results: []FlightData{
			{
				FlightNumber:      "ID-6500",
				AirlineName:       "Batik Air",
				AirlineIATA:       "ID",
				Origin:            "CGK",
				Destination:       "DPS",
				DepartureDateTime: "2026-01-02T08:30:00+0700",
				ArrivalDateTime:   "2026-01-02T11:00:00+0700",
				TravelTime:        "2 h 30 m",
				NumberOfStops:     1,
				Fare:              Fare{TotalPrice: 1450000, CurrencyCode: "IDR", Class: "Y"},
				SeatsAvailable:    12,
				AircraftModel:     "Airbus A320",
				BaggageInfo:       "7 kg Cabin baggage, 20 kg Checked baggage",
				OnboardServices:   []string{"Meal", "Entertainment"},
			},
		},
	}

	got, err := response.ToFlightDomain()

	require.NoError(t, err)
	require.Len(t, got, 1)

	flight := got[0]
	assert.Equal(t, "ID-6500_BatikAir", flight.ID)
	assert.Equal(t, "Batik Air", flight.Provider)
	assert.Equal(t, "Batik Air", flight.Airline.Name)
	assert.Equal(t, "ID", flight.Airline.Code)
	assert.Equal(t, "ID-6500", flight.FlightNumber)
	assert.Equal(t, "CGK", flight.Departure.Airport)
	assert.Equal(t, "Jakarta", flight.Departure.City)
	assert.True(t, time.Date(2026, time.January, 2, 8, 30, 0, 0, time.FixedZone("", 7*60*60)).Equal(flight.Departure.DateTime))
	assert.Equal(t, flight.Departure.DateTime.Unix(), flight.Departure.Timestamp)
	assert.Equal(t, "DPS", flight.Arrival.Airport)
	assert.Equal(t, "Denpasar", flight.Arrival.City)
	assert.Equal(t, 150, flight.Duration.TotalMinutes)
	assert.Equal(t, "2h 30m", flight.Duration.Formatted)
	assert.Equal(t, 1, flight.Stops)
	assert.Equal(t, int64(1450000), flight.Price.Amount)
	assert.Equal(t, "IDR", flight.Price.Currency)
	assert.Equal(t, 12, flight.AvailableSeats)
	assert.Equal(t, "economy", flight.CabinClass)
	require.NotNil(t, flight.Aircraft)
	assert.Equal(t, "Airbus A320", *flight.Aircraft)
	assert.Equal(t, []string{"Meal", "Entertainment"}, flight.Amenities)
	assert.Equal(t, "7 kg Cabin baggage", flight.Baggage.CarryOn)
	assert.Equal(t, "20 kg Checked baggage", flight.Baggage.Checked)
}

func TestSearchResponse_ToFlightDomain_InvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		flight FlightData
	}{
		{
			name: "departure time",
			flight: FlightData{
				FlightNumber:      "ID-6500",
				DepartureDateTime: "invalid",
				ArrivalDateTime:   "2026-01-02T11:00:00+0700",
				TravelTime:        "2h",
			},
		},
		{
			name: "arrival time",
			flight: FlightData{
				FlightNumber:      "ID-6500",
				DepartureDateTime: "2026-01-02T08:30:00+0700",
				ArrivalDateTime:   "invalid",
				TravelTime:        "2h",
			},
		},
		{
			name: "travel duration",
			flight: FlightData{
				FlightNumber:      "ID-6500",
				DepartureDateTime: "2026-01-02T08:30:00+0700",
				ArrivalDateTime:   "2026-01-02T11:00:00+0700",
				TravelTime:        "invalid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (SearchResponse{Results: []FlightData{tt.flight}}).ToFlightDomain()

			require.Error(t, err)
			assert.Nil(t, got)
		})
	}
}

func TestMapCabinClass(t *testing.T) {
	assert.Equal(t, "economy", mapCabinClass("Y"))
	assert.Equal(t, "economy", mapCabinClass("y"))
	assert.Equal(t, "business", mapCabinClass("BUSINESS"))
}

func TestMapBaggage(t *testing.T) {
	got := mapBaggage("20 kg checked baggage, 7 kg cabin baggage")

	assert.Equal(t, "7 kg cabin baggage", got.CarryOn)
	assert.Equal(t, "20 kg checked baggage", got.Checked)
	assert.Empty(t, mapBaggage("").CarryOn)
	assert.Empty(t, mapBaggage("").Checked)
}
