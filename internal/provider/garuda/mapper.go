package garuda

import (
	"flight-finder/internal/domain"
	"fmt"
	"strconv"
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
	departure, err := time.Parse(time.RFC3339, flight.Departure.Time)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("invalid departure time for %s: %w", flight.FlightId, err)
	}

	arrival, err := time.Parse(time.RFC3339, flight.Arrival.Time)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("invalid arrival time for %s: %w", flight.FlightId, err)
	}
	arrivalSchedule := domain.Schedule {
		Airport: flight.Arrival.Airport,
		City: flight.Arrival.City,
		DateTime: arrival,
	}

	durationMinutes := flight.DurationMinutes
	stops := flight.Stops

	if flight.Segments != nil {
		stops = len(flight.Segments) - 1
		lastSegment := flight.Segments[len(flight.Segments) - 1]

		arrival, err := time.Parse(time.RFC3339, lastSegment.Arrival.Time)
		if err != nil {
			return domain.Flight{}, fmt.Errorf("invalid arrival time for %s: %w", lastSegment.FlightNumber, err)
		}

		arrivalSchedule.Airport = lastSegment.Arrival.Airport
		arrivalSchedule.City = lastSegment.Arrival.City
		arrivalSchedule.DateTime = arrival
		arrivalSchedule.Timestamp = arrival.Unix()

		durationMinutes = 0
		for _, segment := range flight.Segments {
			durationMinutes += segment.DurationMinutes
			if segment.LayoverMinutes != nil {
				durationMinutes += *segment.LayoverMinutes
			}
		}
	}

	return domain.Flight{
		ID: flight.FlightId + "_" + strings.ReplaceAll(flight.Airline, " ", ""),
		Provider: flight.Airline,
		Airline: domain.Airline{
			Name: flight.Airline,
			Code: flight.AirlineCode,
		},
		FlightNumber: flight.FlightId,
		Departure: domain.Schedule{
			Airport: flight.Departure.Airport,
			City: flight.Departure.City,
			DateTime: departure,
			Timestamp: departure.Unix(),
		},
		Arrival: arrivalSchedule,
		Duration: domain.NewDuration(durationMinutes),
		Stops: stops,
		Price: domain.Price{
			Amount: flight.Price.Amount,
			Currency: flight.Price.Currency,
		},
		AvailableSeats: flight.AvailableSeats,
		CabinClass: flight.FareClass,
		Aircraft: &flight.Aircraft,
		Amenities: flight.Amenities,
		Baggage: domain.Baggage{
			CarryOn: strconv.Itoa(flight.Baggage.CarryOn),
			Checked: strconv.Itoa(flight.Baggage.Checked),
		},
	}, nil
}