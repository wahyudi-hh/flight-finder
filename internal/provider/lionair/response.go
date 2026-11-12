package lionair

type SearchResponse struct {
    Success bool    `json:"success"`
    Data    Data  `json:"data"`
}

type Data struct {
    AvailableFlights    []FlightData    `json:"available_flights"`
}

type FlightData struct {
    Id    	    string      `json:"id"`
    Carrier     Carrier     `json:"carrier"`
    Route   	Route       `json:"route"`
    Schedule    Schedule    `json:"schedule"`
    FlightTime  int         `json:"flight_time"`
    IsDirect    bool        `json:"is_direct"`
    StopCount   *int	    `json:"stop_count"`
    Layovers    []Stop      `json:"layovers"`
    Pricing 	Price       `json:"pricing"`
    SeatsLeft   int         `json:"seats_left"`
    PlaneType   string      `json:"plane_type"`
	Services    Services    `json:"services"`
}

type Carrier struct {
    Name    string  `json:"name"`
    Iata    string  `json:"iata"`
}

type Route struct {
    From    Airport `json:"from"`
    To      Airport `json:"to"`
}

type Airport struct {
    Code    string  `json:"code"`
	Name    string  `json:"name"`
    City    string  `json:"city"`
}

type Schedule struct {
    Departure           string  `json:"departure"`
	DepartureTimezone  string  `json:"departure_timezone"`
    Arrival             string  `json:"arrival"`
    ArrivalTimezone    string  `json:"arrival_timezone"`
}

type Price struct {
    Total       int64   `json:"total"`
	Currency    string  `json:"currency"`
    FareType    string  `json:"fare_type"`
}

type Services struct {
    WifiAvailable       bool    `json:"wifi_available"`
    MealsIncluded       bool    `json:"meals_included"`
    BaggageAllowance   Baggage `json:"baggage_allowance"` 
}

type Baggage struct {
    Cabin   string  `json:"cabin"`
    Hold    string  `json:"hold"`
}

type Stop struct {
	Airport			string	`json:"airport"`
	DurationMinutes int		`json:"duration_minutes"`
}