package domain

import (
	"fmt"
	"time"
)

type Flight struct {
	ID				string		`json:"id"`
	Provider		string		`json:"provider"`
	Airline			Airline		`json:"airline"`
	FlightNumber	string		`json:"flight_number"`
	Departure		Schedule	`json:"departure"`
	Arrival			Schedule	`json:"arrival"`
	Duration		Duration	`json:"duration"`
	Stops			int			`json:"stops"`
	Price			Price		`json:"price"`
	AvailableSeats	int			`json:"available_seats"`
	CabinClass		string		`json:"cabin_class"`
	Aircraft		*string		`json:"aircraft"`
	Amenities		[]string	`json:"amenities"`
	Baggage			Baggage		`json:"baggage"`

}

type Airline struct {
	Name	string	`json:"name"`
	Code	string	`json:"code"`
}

type Schedule struct {
	Airport		string		`json:"airport"`
	City		string		`json:"city"`
	DateTime	time.Time	`json:"datetime"`
	Timestamp	int64		`json:"timestamp"`
}

type Duration struct {
	TotalMinutes	int		`json:"total_minutes"`
	Formatted		string	`json:"formatted"`
}

type Price struct {
	Amount		int64	`json:"amount"`
	Currency	string	`json:"currency"`
}

type Baggage struct {
	CarryOn	string	`json:"carry_on"`
	Checked	string	`json:"checked"`
}

func NewDuration(totalMinutes int) Duration {
    hours := totalMinutes / 60
    minutes := totalMinutes % 60

    return Duration{
        TotalMinutes: totalMinutes,
        Formatted:    fmt.Sprintf("%dh %dm", hours, minutes),
    }
}