package airasia

import (
	"flight-finder/internal/airport"
	"flight-finder/internal/domain"
	"fmt"
	"math"
	"strings"
	"time"
)

func (r SearchResponse) ToFlightDomain() ([]domain.Flight, error) {
	flights := make([]domain.Flight, 0, len(r.Flights))

	for _,flight := range r.Flights {
		mapped, err := mapFlight(flight)
		if err != nil {
			return nil, err
		}
		flights = append(flights, mapped)
	}
	return flights, nil
}

func mapFlight(flight FlightData) (domain.Flight, error) {
	airlineCode := extractAirlineCode(flight.FlightCode)

	departure, err := time.Parse(time.RFC3339, flight.DepartTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("invalid departure time for %s: %w", flight.FlightCode, err)
	}

	arrival, err := time.Parse(time.RFC3339, flight.ArriveTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("invalid arrival time for %s: %w", flight.FlightCode, err)
	}

	airportData, ok := airport.Lookup(flight.FromAirport)
	fromCity := ""
	if ok {
		fromCity = airportData.City
	}

	airportData, ok = airport.Lookup(flight.ToAirport)
	toCity := ""
	if ok {
		toCity = airportData.City
	}

	durationMinutes := int(math.Ceil(flight.DurationHours * 60))

	stops := 0
	if !flight.DirectFlight {
		stops = len(flight.Stops)
	}

	return domain.Flight{
		ID: flight.FlightCode + "_" + flight.Airline,
		Provider: flight.Airline,
		Airline: domain.Airline{
			Name: flight.Airline,
			Code: airlineCode,
		},
		FlightNumber: flight.FlightCode,
		Departure: domain.Schedule{
			Airport: flight.FromAirport,
			City: fromCity,
			DateTime: departure,
		},
		Arrival: domain.Schedule{
			Airport: flight.ToAirport,
			City: toCity, 
			DateTime: arrival,
		},
		DurationMinute: durationMinutes,
		Stops: stops,
		Price: domain.Price{
			Amount: flight.PriceIDR,
			Currency: "IDR",
		},
		AvailableSeats: flight.Seats,
		CabinClass: flight.CabinClass,
		Aircraft: nil,
		Amenities: []string{},
		Baggage: domain.Baggage{
			CarryOn: "Cabin baggage only",
			Checked: "Additional fee",
		},
	}, nil
}

func extractAirlineCode(flightCode string) string {
	for i,char := range flightCode {
		if char >= '0' && char <= '9' {
			return flightCode[:i]
		}
	}
	return strings.TrimSpace(flightCode)
}