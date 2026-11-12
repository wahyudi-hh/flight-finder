package garuda

type SearchResponse struct {
    Status  string        `json:"status"`
    Flights []FlightData  `json:"flights"`
}

type FlightData struct {
    FlightId    	string      `json:"flight_id"`
    Airline       	string      `json:"airline"`
    AirlineCode   	string      `json:"airline_code"`
    Departure     	Schedule    `json:"departure"`
    Arrival    	    Schedule    `json:"arrival"`
    DurationMinutes int       `json:"duration_minutes"`
    Stops			int	        `json:"stops"`
    Aircraft  	    string      `json:"aircraft"`
    Price      	    Price       `json:"price"`
    Segments        []Segment  `json:"segments"`
	AvailableSeats	int    	    `json:"available_seats"`
	FareClass		string      `json:"fare_class"`
	Baggage		    Baggage     `json:"baggage"`
    Amenities		[]string    `json:"amenities"`
}

type Schedule struct {
    Airport     string  `json:"airport"`
    City       	string `json:"city"`
    Time       	string  `json:"time"`
    Terminal    string `json:"terminal"`
}

type Price struct {
	Amount		int64   `json:"amount"`
	Currency	string  `json:"currency"`
}

type Segment struct {
    FlightNumber    string      `json:"flight_number"`
    Departure     	Schedule    `json:"departure"`
    Arrival    	    Schedule    `json:"arrival"`
    DurationMinutes int     `json:"duration_minutes"`
    LayoverMinutes  *int     `json:"layover_minutes"`
}

type Baggage struct {
	CarryOn	    int `json:"carry_on"`
	Checked	    int	`json:"checked"`
}