# Flight Finder

A Go-based flight search API that aggregates flight offers from multiple airline providers, normalizes their responses into a common flight model, and supports filtering, sorting, best-value ranking, and short-lived in-memory caching.

The current implementation uses local JSON fixtures to simulate airline provider APIs. It does not make external airline API calls.

## Features

- Aggregates search results from Garuda Indonesia, Lion Air, Batik Air, and AirAsia.
- Calls providers concurrently.
- Normalizes provider-specific response formats into a shared `domain.Flight` model.
- Handles partial provider failures and returns results from providers that succeed.
- Simulates provider response latency; AirAsia also simulates a 10% request failure rate.
- Retries failed provider calls once, with a 100 ms backoff between attempts.
- Applies a per-attempt provider timeout and an overall request timeout.
- Supports filtering by price, number of maximum stops, airline, total duration, and local departure/arrival time.
- Supports sorting by price, duration, departure time, arrival time, and best value (score by price and duration).
- Caches complete aggregated provider results in memory for five minutes.

## Tech stack

- Go
- Standard library `net/http` for the HTTP API
- `context` for request cancellation and timeouts
- `sync.RWMutex` for concurrent in-memory cache access
- `go:embed` for bundling mock provider JSON fixtures
- `testify` for tests

## Project structure

```text
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── aggregator/       # Concurrent provider orchestration, retries, timeouts
│   ├── airport/          # Airport-code to city lookup
│   ├── cache/            # In-memory cache with TTL and lazy expiration
│   ├── domain/           # Shared flight and airport models
│   ├── handler/          # HTTP handler, query parsing, validation, response DTOs
│   ├── provider/         # Provider interface and provider-specific clients/mappers
│   │   ├── airasia/
│   │   ├── batikair/
│   │   ├── garuda/
│   │   ├── lionair/
│   │   └── mockdata/     # Embedded JSON fixtures
│   ├── search/           # Filtering and sorting
│   └── service/          # Search orchestration and cache integration
├── go.mod
└── README.md
```

## Requirements

- Go version **1.25.5** (as declared in `go.mod`)

## Run locally

From the project root:

```bash
go mod download
go run ./cmd/api
```

The API listens on:

```text
http://localhost:8080
```

## API

### Search flights

```http
GET /api/v1/flights/search
```

The endpoint accepts query parameters.

#### Required query parameters

| Parameter | Description | Example |
|---|---|---|
| `origin` | Origin airport IATA code | `CGK` |
| `destination` | Destination airport IATA code | `DPS` |
| `departureDate` | Departure date in `YYYY-MM-DD` format | `2025-12-15` |
| `passengers` | Number of passengers; must be greater than zero | `1` |

#### Optional search parameters

| Parameter | Description | Example |
|---|---|---|
| `returnDate` | Return date, if supplied | `2025-12-20` |
| `cabinClass` | Requested cabin class | `economy` |

`returnDate` is accepted and passed through in the provider search request. The current mock provider implementations do not perform a separate return-leg search, so complete round-trip search is not implemented.

#### Filter parameters

| Parameter | Description | Example |
|---|---|---|
| `minPrice` | Minimum price, inclusive; non-negative integer | `500000` |
| `maxPrice` | Maximum price, inclusive; non-negative integer | `1500000` |
| `maxStops` | Maximum number of stops; non-negative integer | `1` |
| `maxDuration` | Maximum total duration in minutes | `300` |
| `airlines` | Comma-separated airline IATA codes | `GA,JT,ID` |
| `departure_time_from` | Earliest local departure time, `HH:MM` | `08:00` |
| `departure_time_to` | Latest local departure time, `HH:MM` | `14:00` |
| `arrival_time_from` | Earliest local arrival time, `HH:MM` | `10:00` |
| `arrival_time_to` | Latest local arrival time, `HH:MM` | `18:00` |

Price filters compare the numeric `Price.Amount` value. They do not convert currencies. The current mock data uses IDR, but this assumption should be revisited if providers return different currencies.

Departure and arrival time filters compare the clock time represented by each flight's parsed `time.Time`, retaining the timezone supplied or resolved by the provider mapper.

#### Sorting parameters

| Parameter | Supported values | Description | Default value
|---|---|---|---|
| `sortBy` | `price`, `duration`, `departure`, `arrival`, `bestValue` | Field or ranking used to sort results | `bestValue`
| `sortOrder` | `asc`, `desc` | Sort direction for ordinary sort fields | `asc`

For `bestValue`, the implementation ranks flights in descending score order.

#### Example request

```bash
curl "http://localhost:8080/api/v1/flights/search?origin=CGK&destination=DPS&departureDate=2025-12-15&passengers=1&cabinClass=economy&minPrice=500000&maxStops=1&sortBy=price&sortOrder=asc"
```

#### Response structure

A successful response has this general structure (flight values depend on the fixture data):

```json
{
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy"
  },
  "metadata": {
    "total_results": 12,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 250,
    "cache_hit": false
  },
  "flights": []
}
```

The metadata fields mean:

- `total_results`: number of flights after filtering.
- `providers_queried`: number of configured providers.
- `providers_succeeded`: providers that returned successfully, including providers that succeeded after a retry.
- `providers_failed`: providers that still failed after the retry limit.
- `search_time_ms`: elapsed service time in integer milliseconds. Very fast cache hits can report `0` because sub-millisecond fractions are truncated.
- `cache_hit`: whether the aggregated provider result was retrieved from cache.

The handler currently returns `400 Bad Request` for invalid query parameters and `503 Service Unavailable` when the search service returns an error (including when all providers fail).

## Architecture and request flow

1. The HTTP handler parses and validates the search query and filter/sort parameters.
2. The handler creates an overall request context with a three-second timeout.
3. The search service creates a cache key from the provider search request (origin, destination, dates, passenger count, and cabin class).
4. On a cache hit, the service retrieves a copy of the cached aggregated result. On a cache miss, it calls the aggregator.
5. The aggregator calls all configured providers concurrently.
6. Each provider call gets a one-second timeout. A failed call is retried once after a 100 ms backoff, unless the overall request context has been cancelled or expired.
7. Successful provider results are combined. If some providers fail, results from the successful providers are still returned. If every provider fails, the aggregator returns an error.
8. Complete aggregated results (with no failed providers) are cached for five minutes.
9. The service applies the requested filters and sorting to the aggregated result and returns it to the handler.

### Timeout and retry configuration

| Setting | Current value |
|---|---:|
| Overall HTTP search context timeout | 3 seconds |
| Timeout per provider attempt | 1 second |
| Maximum provider attempts | 2 (one initial attempt plus one retry) |
| Backoff between attempts | 100 ms |

## Provider fixtures and assumptions

The provider clients read embedded JSON files from `internal/provider/mockdata`:

- `garuda_indonesia_search_response.json`
- `lion_air_search_response.json`
- `batik_air_search_response.json`
- `airasia_search_response.json`

The fixtures simulate provider responses. The implementation assumes that the provider accept a `SearchRequest` and a real airline provider API would already return flights matching the submitted search criteria. Flight Finder then applies the additional filters required by the assignment after aggregating provider results.

The mock fixtures contain sample flight data for a fixed route/date. They should not be interpreted as a live inventory or as dynamically generated results for arbitrary query dates and routes.

### Provider latency and failures

| Provider | Simulated latency | Simulated random failure |
|---|---:|---:|
| Garuda Indonesia | 50–100 ms | No |
| Lion Air | 100–200 ms | No |
| Batik Air | 200–400 ms | No |
| AirAsia | 50–150 ms | 10% |

Provider delays are simulated with timers and respect context cancellation.

### Timezones and duration

Provider-specific mappers normalize timestamps into Go `time.Time` values and retain the supplied offset or resolve the provider's timezone information. Unix timestamps are also populated for chronological sorting.

For connecting flights, the normalized total duration should represent elapsed journey time, including layovers. The Garuda fixture includes segment-level duration and layover data for GA315. Its mapper calculates the total from the segment durations and layover: 90 + 30 + 105 = 225 minutes. The top-level duration value is not used for this flight when segment data is present.

The other providers use their supplied overall duration fields. Their correctness depends on those provider fields representing the complete journey duration, including any connection time.

### Price comparison and provider identity assumptions

The configured providers represent direct airline integrations, not third-party flight aggregators. The implementation therefore assumes the same flight will not be independently listed by multiple providers, and it does not deduplicate or compare duplicate offers for an identical flight across providers.

If a provider is itself an airline aggregator or reseller, duplicate-flight matching and cross-provider offer comparison would need to be added.

Price sorting and best-value scoring currently compare raw numeric price amounts without currency conversion. The mock data is treated as IDR-denominated for this purpose. This should be changed before supporting mixed-currency provider results.

### Best-value ranking

Best value combines normalized price and duration scores with equal weighting:

```text
priceScore    = (maxPrice - flightPrice) / (maxPrice - minPrice)
durationScore = (maxDuration - flightDuration) / (maxDuration - minDuration)
bestValue     = 0.5 * priceScore + 0.5 * durationScore
```

A higher score ranks first. If all flights have the same price or duration, that dimension receives a score of `1.0` for every flight. The score is calculated across the current result set, so rankings are relative to the flights returned by the providers and are not an absolute quality measure.

## Cache management

- The application uses an in-memory cache with a five-minute TTL.
- Cache keys are generated by JSON-marshalling the provider `SearchRequest`.
- Filters and sorting parameters are intentionally excluded from the key because the cache stores the raw aggregated provider result. Different filters or sort orders for the same provider search can reuse the same cached data.
- Only results where all configured providers succeeded are cached. Partial results are returned to the caller but are not cached, allowing a subsequent request to retry providers that previously failed.
- Expired entries are removed lazily when their key is read. There is no background cleanup goroutine, so an expired entry that is never accessed again can remain allocated in memory until the cache is discarded or the key is accessed.
- The cache uses `sync.RWMutex` to protect concurrent map access.
- The cache is local and is cleared when the application restarts. It is suitable for this assignment's single-instance service, not a distributed deployment.

## Testing

Run all tests:

```bash
go test ./... -count=1
```

Generate a coverage profile and view it in a browser:

```bash
go test ./... -count=1 -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Known limitations and assumptions

- Provider responses are static fixtures rather than live airline API responses.
- The mock provider clients do not filter fixtures using the incoming search request; they assume real provider APIs would do that.
- `returnDate` is accepted but a complete outbound-plus-return search is not implemented.
- There is no multi-city search.
- Prices are compared as raw amounts; currency conversion is not implemented.
- Duplicate-flight detection and cross-provider offer comparison are not implemented under the direct-airline-provider assumption.
- The airport lookup currently contains a limited set of airport codes; unknown codes can produce an empty city name in mappers that use the lookup.
- Retry classification currently retries all non-cancellation provider errors, rather than distinguishing transient and permanent errors.
- The cache is in-memory and local. It is suitable for single-instance, not a distributed deployment.
