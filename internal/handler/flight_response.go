package handler

import "flight-finder/internal/domain"

type FlightSearchResponse struct {
	SearchCriteria FlightSearchRequest `json:"search_criteria"`
	Metadata       SearchMetadata      `json:"metadata"`
	Flights        []domain.Flight     `json:"flights"`
}

type SearchMetadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}