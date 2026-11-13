package handler

import (
	"context"
	"encoding/json"
	"flight-finder/internal/provider"
	"flight-finder/internal/service"
	"net/http"
	"strconv"
	"time"
)

const requestTimeout = 3 * time.Second

type FlightHandler struct {
	service *service.FlightSearchService
}

func NewFlightHandler(service *service.FlightSearchService) *FlightHandler {
	return &FlightHandler{
		service: service,
	}
}

func (h *FlightHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	request := FlightSearchRequest{
		Origin: query.Get("origin"),
		Destination: query.Get("destination"),
		DepartureDate: query.Get("departureDate"),
		ReturnDate: query.Get("returnDate"),
		CabinClass: query.Get("cabinClass"),
	}

	passengers, err := strconv.Atoi(query.Get("passengers"))
	if err != nil {
		http.Error(w, "passengers must be a valid integer", http.StatusBadRequest)
		return
	}
	request.Passengers = passengers

	if err := request.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	searchRequest := provider.SearchRequest{
		Origin: request.Origin,
		Destination: request.Destination,
		DepartureDate: request.DepartureDate,
		Passengers: request.Passengers,
		CabinClass: request.CabinClass,
	}
	if request.ReturnDate != "" {
		searchRequest.ReturnDate = &request.ReturnDate
	}

	filter, err := request.ToFilter(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sort, err := request.ToSort(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()
	result, searchTimeMs, cacheHit, err := h.service.Search(ctx, searchRequest, filter, sort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := FlightSearchResponse{
		SearchCriteria: request,
		Metadata: SearchMetadata{
			TotalResults: len(result.Flights),
			ProvidersQueried: result.TotalProviders,
			ProvidersSucceeded: result.TotalProviders - result.FailedProviders,
			ProvidersFailed: result.FailedProviders,
			SearchTimeMs: searchTimeMs,
			CacheHit: cacheHit,
		},
		Flights: result.Flights,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}