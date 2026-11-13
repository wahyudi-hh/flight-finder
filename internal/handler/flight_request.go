package handler

import (
	"flight-finder/internal/search"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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

func (r FlightSearchRequest) ToFilter(query url.Values) (search.FlightFilter, error) {
	filter := search.FlightFilter{}

	if value := query.Get("minPrice"); value != "" {
		price, err := strconv.ParseInt(value, 10, 64)
		if err != nil || price < 0 {
			return filter, fmt.Errorf("minPrice must be a non-negative integer")
		}
		filter.MinPrice = &price
	}
	if value := query.Get("maxPrice"); value != "" {
		price, err := strconv.ParseInt(value, 10, 64)
		if err != nil || price < 0 {
			return filter, fmt.Errorf("maxPrice must be a non-negative integer")
		}
		filter.MaxPrice = &price
	}
	if filter.MinPrice != nil && filter.MaxPrice != nil && *filter.MinPrice > *filter.MaxPrice {
		return filter, fmt.Errorf("min_price cannot exceed max_price")
	}

	if value := query.Get("maxStops"); value != "" {
		stops, err := strconv.Atoi(value)
		if err != nil || stops < 0 {
			return filter, fmt.Errorf("maxStops must be a non-negative integer")
		}
		filter.MaxStops = &stops
	}

	if value := query.Get("maxDuration"); value != "" {
		duration, err := strconv.Atoi(value)
		if err != nil || duration < 0 {
			return filter, fmt.Errorf("maxDuration must be a non-negative integer")
		}
		filter.MaxDurationMinutes = &duration
	}

	if value := query.Get("airlines"); value != "" {
		filter.Airlines = strings.Split(value, ",")
	}

	if value := query.Get("departure_time_from"); value != "" {
		filter.DepartureTimeFrom = &value
	}
	if value := query.Get("departure_time_to"); value != "" {
		filter.DepartureTimeTo = &value
	}

	if value := query.Get("arrival_time_from"); value != "" {
		filter.ArrivalTimeFrom = &value
	}
	if value := query.Get("arrival_time_to"); value != "" {
		filter.ArrivalTimeTo = &value
	}

	return filter, nil
}

func (r FlightSearchRequest) ToSort(query url.Values) (search.FlightSort, error) {
	sortBy := query.Get("sortBy")
	if sortBy == "" {
		sortBy = string(search.SortByBestValue)
	}

	sortOrder := query.Get("sortOrder")
	if sortOrder == "" {
		sortOrder = string(search.Ascending)
	}

	sort := search.FlightSort{
		By: search.SortBy(sortBy),
		Order: search.SortOrder(sortOrder),
	}

	if !sort.IsValid() {
		return search.FlightSort{}, fmt.Errorf("invalid sorting parameter(s)")
	}
	return sort, nil
}