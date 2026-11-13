package main

import (
	"flight-finder/internal/aggregator"
	"flight-finder/internal/cache"
	"flight-finder/internal/handler"
	"flight-finder/internal/middleware"
	"flight-finder/internal/provider/airasia"
	"flight-finder/internal/provider/batikair"
	"flight-finder/internal/provider/garuda"
	"flight-finder/internal/provider/lionair"
	"flight-finder/internal/service"
	"log"
	"net/http"
)

func main() {

	flightAggregator := aggregator.NewAggregator(
		airasia.NewClient(),
		batikair.NewClient(),
		garuda.NewClient(),
		lionair.NewClient(),
	)

	flightCache := cache.NewInMemoryCache()

	flightService := service.NewFlightSearchService(flightAggregator, flightCache)

	flightHandler := handler.NewFlightHandler(flightService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/flights/search", flightHandler.Search)

	rateLimiter := middleware.NewIPRateLimiter(1, 10)
	httpHandler := rateLimiter.Middleware(mux)

	server := &http.Server{
		Addr: ":8080",
		Handler: httpHandler,
	}

	log.Println("Server listening on :8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}