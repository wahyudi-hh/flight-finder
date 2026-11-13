package main

import (
	"flight-finder/internal/aggregator"
	"flight-finder/internal/cache"
	"flight-finder/internal/handler"
	"flight-finder/internal/provider/airasia"
	"flight-finder/internal/provider/batikair"
	"flight-finder/internal/provider/garuda"
	"flight-finder/internal/provider/lionair"
	"flight-finder/internal/service"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	flightAggregator := aggregator.NewAggregator(
		airasia.NewClient(),
		batikair.NewClient(),
		garuda.NewClient(),
		lionair.NewClient(),
	)

	flightCache := cache.NewInMemoryCache()

	flightService := service.NewFlightSearchService(flightAggregator, flightCache)

	flightHandler := handler.NewFlightHandler(flightService)

	mux.HandleFunc("GET /api/v1/flights/search", flightHandler.Search)

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("Server listening on :8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}