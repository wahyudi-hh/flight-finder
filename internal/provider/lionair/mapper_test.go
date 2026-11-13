package lionair

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchResponse_ToFlightDomain(t *testing.T) {
	stops := 1
	response := SearchResponse{
		Data: Data{
			AvailableFlights: []FlightData{
				{
					Id:      "JT-123",
					Carrier: Carrier{Name: "Lion Air", Iata: "JT"},
					Route: Route{
						From: Airport{Code: "CGK", Name: "Soekarno-Hatta", City: "Jakarta"},
						To:   Airport{Code: "DPS", Name: "Ngurah Rai", City: "Denpasar"},
					},
					Schedule: Schedule{
						Departure:         "2026-01-02T08:30:00",
						DepartureTimezone: "UTC",
						Arrival:           "2026-01-02T12:00:00",
						ArrivalTimezone:   "UTC",
					},
					FlightTime: 210,
					StopCount:  &stops,
					Pricing:    Price{Total: 1250000, Currency: "IDR", FareType: "ECONOMY"},
					SeatsLeft:  9,
					PlaneType:  "Boeing 737",
					Services: Services{
						BaggageAllowance: Baggage{Cabin: "7 kg", Hold: "20 kg"},
					},
				},
			},
		},
	}

	got, err := response.ToFlightDomain()

	require.NoError(t, err)
	require.Len(t, got, 1)

	flight := got[0]
	assert.Equal(t, "JT-123_LionAir", flight.ID)
	assert.Equal(t, "Lion Air", flight.Provider)
	assert.Equal(t, "Lion Air", flight.Airline.Name)
	assert.Equal(t, "JT", flight.Airline.Code)
	assert.Equal(t, "JT", flight.FlightNumber)
	assert.Equal(t, "CGK", flight.Departure.Airport)
	assert.Equal(t, "Jakarta", flight.Departure.City)
	assert.Equal(t, time.Date(2026, time.January, 2, 8, 30, 0, 0, time.UTC), flight.Departure.DateTime)
	assert.Equal(t, flight.Departure.DateTime.Unix(), flight.Departure.Timestamp)
	assert.Equal(t, "DPS", flight.Arrival.Airport)
	assert.Equal(t, "Denpasar", flight.Arrival.City)
	assert.Equal(t, time.Date(2026, time.January, 2, 12, 0, 0, 0, time.UTC), flight.Arrival.DateTime)
	assert.Equal(t, flight.Arrival.DateTime.Unix(), flight.Arrival.Timestamp)
	assert.Equal(t, 210, flight.Duration.TotalMinutes)
	assert.Equal(t, "3h 30m", flight.Duration.Formatted)
	assert.Equal(t, 1, flight.Stops)
	assert.Equal(t, int64(1250000), flight.Price.Amount)
	assert.Equal(t, "IDR", flight.Price.Currency)
	assert.Equal(t, 9, flight.AvailableSeats)
	assert.Equal(t, "economy", flight.CabinClass)
	require.NotNil(t, flight.Aircraft)
	assert.Equal(t, "Boeing 737", *flight.Aircraft)
	assert.Empty(t, flight.Amenities)
	assert.Equal(t, "7 kg", flight.Baggage.CarryOn)
	assert.Equal(t, "20 kg", flight.Baggage.Checked)
}

func TestSearchResponse_ToFlightDomain_DefaultsMissingStops(t *testing.T) {
	response := SearchResponse{Data: Data{AvailableFlights: []FlightData{{
		Schedule: Schedule{
			Departure:         "2026-01-02T08:30:00",
			DepartureTimezone: "UTC",
			Arrival:           "2026-01-02T10:00:00",
			ArrivalTimezone:   "UTC",
		},
	}}}}

	got, err := response.ToFlightDomain()

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Zero(t, got[0].Stops)
}

func TestSearchResponse_ToFlightDomain_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
	}{
		{
			name: "invalid departure timezone",
			schedule: Schedule{
				Departure:         "2026-01-02T08:30:00",
				DepartureTimezone: "Invalid/Timezone",
				Arrival:           "2026-01-02T10:00:00",
				ArrivalTimezone:   "UTC",
			},
		},
		{
			name: "invalid departure datetime",
			schedule: Schedule{
				Departure:         "invalid",
				DepartureTimezone: "UTC",
				Arrival:           "2026-01-02T10:00:00",
				ArrivalTimezone:   "UTC",
			},
		},
		{
			name: "invalid arrival datetime",
			schedule: Schedule{
				Departure:         "2026-01-02T08:30:00",
				DepartureTimezone: "UTC",
				Arrival:           "invalid",
				ArrivalTimezone:   "UTC",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flight := FlightData{Id: "JT-123", Schedule: tt.schedule}
			got, err := (SearchResponse{Data: Data{AvailableFlights: []FlightData{flight}}}).ToFlightDomain()

			require.Error(t, err)
			assert.Nil(t, got)
		})
	}
}

func TestParseLocalDateTime(t *testing.T) {
	got, err := parseLocalDateTime("2026-01-02T08:30:00", "UTC")

	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, time.January, 2, 8, 30, 0, 0, time.UTC), got)

	_, err = parseLocalDateTime("2026-01-02T08:30:00", "Invalid/Timezone")
	require.Error(t, err)
}
