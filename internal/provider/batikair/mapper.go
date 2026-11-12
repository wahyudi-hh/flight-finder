package batikair

import (
	"flight-finder/internal/airport"
	"flight-finder/internal/domain"
	"fmt"
	"strings"
	"time"
)

func (r SearchResponse) ToFlightDomain() ([]domain.Flight, error) {
	flights := make([]domain.Flight, 0, len(r.Results))

	for _,flight := range r.Results {
		mapped, err := mapFlight(flight)
		if err != nil {
			return nil, err
		}
		flights = append(flights, mapped)
	}
	return flights, nil
}

func mapFlight(flight FlightData) (domain.Flight, error) {
	departure, err := time.Parse("2006-01-02T15:04:05-0700", flight.DepartureDateTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("invalid departure time for %s: %w", flight.DepartureDateTime, err)
	}

	arrival, err := time.Parse("2006-01-02T15:04:05-0700", flight.ArrivalDateTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("invalid arrival time for %s: %w", flight.ArrivalDateTime, err)
	}

	airportData, ok := airport.Lookup(flight.Origin)
	fromCity := ""
	if ok {
		fromCity = airportData.City
	}

	airportData, ok = airport.Lookup(flight.Destination)
	toCity := ""
	if ok {
		toCity = airportData.City
	}

	duration, err := time.ParseDuration(strings.ReplaceAll(flight.TravelTime, " ", ""))
	if err != nil {
        return domain.Flight{}, fmt.Errorf("invalid travel duration for %s: %w", flight.FlightNumber, err)
    }

	return domain.Flight{
		ID: flight.FlightNumber + "_" + strings.ReplaceAll(flight.AirlineName, " ", ""),
		Provider: flight.AirlineName,
		Airline: domain.Airline{
			Name: flight.AirlineName,
			Code: flight.AirlineIATA,
		},
		FlightNumber: flight.FlightNumber,
		Departure: domain.Schedule{
			Airport: flight.Origin,
			City: fromCity,
			DateTime: departure,
			Timestamp: departure.Unix(),
		},
		Arrival: domain.Schedule{
			Airport: flight.Destination,
			City: toCity,
			DateTime: arrival,
			Timestamp: arrival.Unix(),
		},
		Duration: domain.NewDuration(int(duration.Minutes())),
		Stops: flight.NumberOfStops,
		Price: domain.Price{
			Amount: flight.Fare.TotalPrice,
			Currency: flight.Fare.CurrencyCode,
		},
		AvailableSeats: flight.SeatsAvailable,
		CabinClass: mapCabinClass(flight.Fare.Class),
		Aircraft: &flight.AircraftModel,
		Amenities: flight.OnboardServices,
		Baggage: mapBaggage(flight.BaggageInfo),
	}, nil
}

func mapCabinClass(fareClass string) string {
    switch strings.ToUpper(fareClass) {
    case "Y":
        return "economy"
    default:
        return strings.ToLower(fareClass)
    }
}

func mapBaggage(info string) domain.Baggage {
    var baggage domain.Baggage

    for _, item := range strings.Split(info, ",") {
        item = strings.TrimSpace(item)

        switch {
        case strings.Contains(strings.ToLower(item), "cabin"):
            baggage.CarryOn = item
        case strings.Contains(strings.ToLower(item), "checked"):
            baggage.Checked = item
        }
    }

    return baggage
}