package lionair

import (
	"flight-finder/internal/domain"
	"fmt"
	"strings"
	"time"
)

func (r SearchResponse) ToFlightDomain() ([]domain.Flight, error) {
	flights := make([]domain.Flight, 0, len(r.Data.AvailableFlights))

	for _,flight := range r.Data.AvailableFlights {
		mapped, err := mapFlight(flight)
		if err != nil {
			return nil, err
		}
		flights = append(flights, mapped)
	}
	return flights, nil
}

func mapFlight(flight FlightData) (domain.Flight, error) {
	departure, err := parseLocalDateTime(flight.Schedule.Departure, flight.Schedule.DepartureTimezone)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("invalid departure time for %s: %w", flight.Id, err)
	}

	arrival, err := parseLocalDateTime(flight.Schedule.Arrival, flight.Schedule.ArrivalTimezone)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("invalid arrival time for %s: %w", flight.Id, err)
	}

	stops := 0
	if flight.StopCount != nil {
		stops = *flight.StopCount
	}

	return domain.Flight{
		ID: flight.Id + "_" + strings.ReplaceAll(flight.Carrier.Name, " ", ""),
		Provider: flight.Carrier.Name,
		Airline: domain.Airline{
			Name: flight.Carrier.Name,
			Code: flight.Carrier.Iata,
		},
		FlightNumber: flight.Carrier.Iata,
		Departure: domain.Schedule{
			Airport: flight.Route.From.Code,
			City: flight.Route.From.City,
			DateTime: departure,
		},
		Arrival: domain.Schedule{
			Airport: flight.Route.From.Code,
			City: flight.Route.To.City,
			DateTime: arrival,
		},
		DurationMinute: flight.FlightTime,
		Stops: stops,
		Price: domain.Price{
			Amount: flight.Pricing.Total,
			Currency: flight.Pricing.Currency,
		},
		AvailableSeats: flight.SeatsLeft,
		CabinClass: strings.ToLower(flight.Pricing.FareType),
		Aircraft: &flight.PlaneType,
		Amenities: []string{},
		Baggage: domain.Baggage{
			CarryOn: flight.Services.BaggageAllowance.Cabin,
			Checked: flight.Services.BaggageAllowance.Hold,
		},
	}, nil
}


func parseLocalDateTime(dateTime string, timezone string) (time.Time, error) {
    location, err := time.LoadLocation(timezone)
    if err != nil {
        return time.Time{}, fmt.Errorf("invalid timezone %q: %w", timezone, err)
    }

    parsedTime, err := time.ParseInLocation("2006-01-02T15:04:05", dateTime, location)
    if err != nil {
        return time.Time{}, fmt.Errorf("invalid datetime %q: %w", dateTime, err)
    }

    return parsedTime, nil
}
