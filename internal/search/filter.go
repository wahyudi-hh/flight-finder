package search

import (
	"flight-finder/internal/domain"
	"strings"
	"time"
)

type FlightFilter struct {
	MinPrice           *int64
	MaxPrice           *int64
	MaxStops           *int
	DepartureTimeFrom  *string
	DepartureTimeTo    *string
	ArrivalTimeFrom    *string
	ArrivalTimeTo      *string
	Airlines           []string
	MaxDurationMinutes *int
}

func FilterFlights(flights []domain.Flight, filter FlightFilter) []domain.Flight {
    filtered := make([]domain.Flight, 0)

    for _, flight := range flights {
        if !matchesPrice(flight, filter.MinPrice, filter.MaxPrice) {
            continue
        }
        if !matchesStops(flight, filter.MaxStops) {
            continue
        }
        if !matchesAirline(flight, filter.Airlines) {
            continue
        }
        if !matchesDuration(flight, filter.MaxDurationMinutes) {
            continue
        }
		if !matchesDepartureTime(flight, filter.DepartureTimeFrom, filter.DepartureTimeTo) {
			continue
		}
		if !matchesArrivalTime(flight, filter.ArrivalTimeFrom, filter.ArrivalTimeTo) {
			continue
		}
        filtered = append(filtered, flight)
    }

    return filtered
}

func matchesPrice(flight domain.Flight, minPrice, maxPrice *int64) bool {
	if minPrice != nil && flight.Price.Amount < *minPrice {
		return false
	}
	if maxPrice != nil && flight.Price.Amount > *maxPrice {
		return false
	}
	return true
}

func matchesStops(flight domain.Flight, maxStops *int) bool {
	if maxStops == nil {
		return true
	}
	return flight.Stops <= *maxStops
}

func matchesDepartureTime(flight domain.Flight, from, to *string) bool {
    return matchesTimeRange(flight.Departure.DateTime, from, to)
}

func matchesArrivalTime(flight domain.Flight, from, to *string) bool {
    return matchesTimeRange(flight.Arrival.DateTime, from, to)
}

func matchesTimeRange(flightTime time.Time, from, to *string) bool {
    flightHour := flightTime.Hour()
    flightMinute := flightTime.Minute()
    flightMinutes := flightHour*60 + flightMinute

    if from != nil {
        startTime, err := time.Parse("15:04", *from)
        if err != nil {
            return false
        }

        startMinutes := startTime.Hour()*60 + startTime.Minute()

        if flightMinutes < startMinutes {
            return false
        }
    }

    if to != nil {
        endTime, err := time.Parse("15:04", *to)
        if err != nil {
            return false
        }

        endMinutes := endTime.Hour()*60 + endTime.Minute()

        if flightMinutes > endMinutes {
            return false
        }
    }

    return true
}

func matchesAirline(flight domain.Flight, airlines []string) bool {
    if len(airlines) == 0 {
        return true
    }
    for _, airline := range airlines {
        if strings.EqualFold(flight.Airline.Code, airline) {
            return true
        }
    }
    return false
}

func matchesDuration(flight domain.Flight, maxDuration *int) bool {
    if maxDuration == nil {
        return true
    }
    return flight.Duration.TotalMinutes <= *maxDuration
}