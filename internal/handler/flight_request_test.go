package handler

import (
	"net/url"
	"testing"

	"flight-finder/internal/search"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlightSearchRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request FlightSearchRequest
		wantErr string
	}{
		{
			name:    "valid request",
			request: FlightSearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2026-12-01", Passengers: 1},
		},
		{
			name:    "missing origin",
			request: FlightSearchRequest{Destination: "DPS", DepartureDate: "2026-12-01", Passengers: 1},
			wantErr: "origin is required",
		},
		{
			name:    "missing destination",
			request: FlightSearchRequest{Origin: "CGK", DepartureDate: "2026-12-01", Passengers: 1},
			wantErr: "destination is required",
		},
		{
			name:    "same origin and destination",
			request: FlightSearchRequest{Origin: "CGK", Destination: "CGK", DepartureDate: "2026-12-01", Passengers: 1},
			wantErr: "origin and destination must be different",
		},
		{
			name:    "missing departure date",
			request: FlightSearchRequest{Origin: "CGK", Destination: "DPS", Passengers: 1},
			wantErr: "departure_date is required",
		},
		{
			name:    "invalid departure date",
			request: FlightSearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "01-12-2026", Passengers: 1},
			wantErr: "departure_date must use YYYY-MM-DD format",
		},
		{
			name:    "zero passengers",
			request: FlightSearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2026-12-01"},
			wantErr: "passengers must be greater than zero",
		},
		{
			name:    "negative passengers",
			request: FlightSearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2026-12-01", Passengers: -1},
			wantErr: "passengers must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestFlightSearchRequest_ToFilter(t *testing.T) {
	query := url.Values{
		"minPrice":            {"100"},
		"maxPrice":            {"500"},
		"maxStops":            {"2"},
		"maxDuration":         {"360"},
		"airlines":            {"GA,JT"},
		"departure_time_from": {"08:00"},
		"departure_time_to":   {"12:00"},
		"arrival_time_from":   {"10:00"},
		"arrival_time_to":     {"15:00"},
	}

	got, err := (FlightSearchRequest{}).ToFilter(query)

	require.NoError(t, err)
	require.NotNil(t, got.MinPrice)
	require.NotNil(t, got.MaxPrice)
	require.NotNil(t, got.MaxStops)
	require.NotNil(t, got.MaxDurationMinutes)
	require.NotNil(t, got.DepartureTimeFrom)
	require.NotNil(t, got.DepartureTimeTo)
	require.NotNil(t, got.ArrivalTimeFrom)
	require.NotNil(t, got.ArrivalTimeTo)
	assert.Equal(t, int64(100), *got.MinPrice)
	assert.Equal(t, int64(500), *got.MaxPrice)
	assert.Equal(t, 2, *got.MaxStops)
	assert.Equal(t, 360, *got.MaxDurationMinutes)
	assert.Equal(t, []string{"GA", "JT"}, got.Airlines)
	assert.Equal(t, "08:00", *got.DepartureTimeFrom)
	assert.Equal(t, "12:00", *got.DepartureTimeTo)
	assert.Equal(t, "10:00", *got.ArrivalTimeFrom)
	assert.Equal(t, "15:00", *got.ArrivalTimeTo)
}

func TestFlightSearchRequest_ToFilterEmptyQuery(t *testing.T) {
	got, err := (FlightSearchRequest{}).ToFilter(url.Values{})

	require.NoError(t, err)
	assert.Equal(t, search.FlightFilter{}, got)
}

func TestFlightSearchRequest_ToFilterRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		query   url.Values
		wantErr string
	}{
		{name: "invalid minimum price", query: url.Values{"minPrice": {"abc"}}, wantErr: "minPrice must be a non-negative integer"},
		{name: "negative minimum price", query: url.Values{"minPrice": {"-1"}}, wantErr: "minPrice must be a non-negative integer"},
		{name: "invalid maximum price", query: url.Values{"maxPrice": {"abc"}}, wantErr: "maxPrice must be a non-negative integer"},
		{name: "negative maximum price", query: url.Values{"maxPrice": {"-1"}}, wantErr: "maxPrice must be a non-negative integer"},
		{name: "minimum exceeds maximum", query: url.Values{"minPrice": {"20"}, "maxPrice": {"10"}}, wantErr: "min_price cannot exceed max_price"},
		{name: "invalid maximum stops", query: url.Values{"maxStops": {"x"}}, wantErr: "maxStops must be a non-negative integer"},
		{name: "negative maximum stops", query: url.Values{"maxStops": {"-1"}}, wantErr: "maxStops must be a non-negative integer"},
		{name: "invalid maximum duration", query: url.Values{"maxDuration": {"x"}}, wantErr: "maxDuration must be a non-negative integer"},
		{name: "negative maximum duration", query: url.Values{"maxDuration": {"-1"}}, wantErr: "maxDuration must be a non-negative integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (FlightSearchRequest{}).ToFilter(tt.query)

			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestFlightSearchRequest_ToSort(t *testing.T) {
	tests := []struct {
		name    string
		query   url.Values
		want    search.FlightSort
		wantErr string
	}{
		{
			name:  "default sorting",
			query: url.Values{},
			want:  search.FlightSort{By: search.SortByBestValue, Order: search.Ascending},
		},
		{
			name:  "custom sorting",
			query: url.Values{"sortBy": {string(search.SortByPrice)}, "sortOrder": {string(search.Descending)}},
			want:  search.FlightSort{By: search.SortByPrice, Order: search.Descending},
		},
		{
			name:    "invalid sort key",
			query:   url.Values{"sortBy": {"unknown"}},
			wantErr: "invalid sorting parameter(s)",
		},
		{
			name:    "invalid sort order",
			query:   url.Values{"sortOrder": {"sideways"}},
			wantErr: "invalid sorting parameter(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (FlightSearchRequest{}).ToSort(tt.query)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Equal(t, search.FlightSort{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
