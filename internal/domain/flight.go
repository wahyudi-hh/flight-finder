package domain

import "time"

type Flight struct {
	ID					string
	Provider			string
	Airline				Airline
	FlightNumber		string
	Departure			Schedule
	Arrival				Schedule
	DurationMinute		int
	Stops				int
	Price				Price
	AvailableSeats		int
	CabinClass			string
	Aircraft			*string
	Amenities			[]string
	Baggage				Baggage
}

type Airline struct {
	Name	string
	Code	string
}

type Schedule struct {
	Airport		string
	City		string
	DateTime	time.Time
}

type Price struct {
	Amount		int64
	Currency	string
}

type Baggage struct {
	CarryOn	string
	Checked	string
}

