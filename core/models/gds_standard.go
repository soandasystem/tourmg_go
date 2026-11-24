package models

import (
	"time"
)

type FlightOfferNorm struct {
	ID              string
	Provider        string
	ProviderOfferID string

	Price    float64
	Currency string

	TotalDuration time.Duration
	Outbound      FlightSlice
	Return        *FlightSlice

	Conditions []FlightCondition
}

type FlightSlice struct {
	Origin      AirportNorm     `json:"origin"`
	Destination AirportNorm     `json:"destination"`
	Segments    []FlightSegment `json:"segments"`
}

type FlightSegment struct {
	Airline       AirlineNorm   `json:"airline"`
	FlightNumber  string        `json:"flight_number"`
	Origin        AirportNorm   `json:"origin"`
	Destination   AirportNorm   `json:"destination"`
	DepartureTime time.Time     `json:"departure_time"`
	ArrivalTime   time.Time     `json:"arrival_time"`
	Duration      time.Duration `json:"duration"`
	Stops         []FlightStop  `json:"stops"`
	Baggage       []BaggageNorm `json:"baggage"`
}

type AirportNorm struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	City        string `json:"city"`
	CountryCode string `json:"country_code"`
	TimeZone    string `json:"time_zone"`
}

type AirlineNorm struct {
	Code    string
	Name    string
	LogoURL string
}

type FlightStop struct {
	Airport AirportNorm `json:"airport"`
}

type BaggageNorm struct {
	Type     string
	Quantity int
}

type FlightCondition struct {
	Type            string
	Allowed         bool
	PenaltyAmount   float64
	PenaltyCurrency string
}
