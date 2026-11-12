package handler

import (
	"fmt"
	"time"
)

type FlightSearchRequest struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	ReturnDate    string `json:"return_date,omitempty"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabin_class"`
}

func (r FlightSearchRequest) Validate() error {
	if r.Origin == "" {
		return fmt.Errorf("origin is required")
	}

	if r.Destination == "" {
		return fmt.Errorf("destination is required")
	}

	if r.Origin == r.Destination {
		return fmt.Errorf("origin and destination must be different")
	}

	if r.DepartureDate == "" {
		return fmt.Errorf("departure_date is required")
	}

	if _, err := time.Parse("2006-01-02", r.DepartureDate); err != nil {
		return fmt.Errorf("departure_date must use YYYY-MM-DD format")
	}

	if r.Passengers <= 0 {
		return fmt.Errorf("passengers must be greater than zero")
	}

	return nil
}