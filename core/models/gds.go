package models

type FlightOfferRequest struct {
	Data FlightOfferRequestData `json:"data"`
}

type FlightOfferRequestData struct {
	Slices           []FlightSliceRequest `json:"slices"`
	Passengers       []Passenger          `json:"passengers"`
	Offers           []FlightOffer        `json:"offers"`
	LiveMode         bool                 `json:"live_mode"`
	ID               string               `json:"id"`
	CreatedAt        string               `json:"created_at"`
	ClientKey        string               `json:"client_key"`
	CabinClass       string               `json:"cabin_class"`
	AirlineCreditIDs []string             `json:"airline_credit_ids"`
}

// --------------------------------------------------
// SLICE REQUEST
// --------------------------------------------------

type FlightSliceRequest struct {
	OriginType      string  `json:"origin_type"`
	Origin          Airport `json:"origin"`
	DestinationType string  `json:"destination_type"`
	Destination     Airport `json:"destination"`
	DepartureDate   string  `json:"departure_date"`
}

// --------------------------------------------------
// AIRPORT / CITY
// --------------------------------------------------

type Airport struct {
	Type            string    `json:"type,omitempty"`
	TimeZone        string    `json:"time_zone"`
	Name            string    `json:"name"`
	Longitude       float64   `json:"longitude"`
	Latitude        float64   `json:"latitude"`
	ID              string    `json:"id"`
	ICAOCode        string    `json:"icao_code"`
	IATACountryCode string    `json:"iata_country_code"`
	IATACode        string    `json:"iata_code"`
	IATACityCode    string    `json:"iata_city_code,omitempty"`
	CityName        string    `json:"city_name,omitempty"`
	City            *City     `json:"city,omitempty"`
	Airports        []Airport `json:"airports,omitempty"`
}

type City struct {
	Name            string    `json:"name"`
	ID              string    `json:"id"`
	IATACountryCode string    `json:"iata_country_code"`
	IATACode        string    `json:"iata_code"`
	Airports        []Airport `json:"airports,omitempty"`
}

// --------------------------------------------------
// PASSENGERS
// --------------------------------------------------

type Passenger struct {
	Type                     string                    `json:"type"`
	LoyaltyProgrammeAccounts []LoyaltyProgrammeAccount `json:"loyalty_programme_accounts,omitempty"`
	ID                       string                    `json:"id"`
	GivenName                string                    `json:"given_name,omitempty"`
	FareType                 string                    `json:"fare_type,omitempty"`
	FamilyName               string                    `json:"family_name,omitempty"`
	Age                      int                       `json:"age,omitempty"`
}

type LoyaltyProgrammeAccount struct {
	AirlineIATACode string `json:"airline_iata_code"`
	AccountNumber   string `json:"account_number"`
}

// --------------------------------------------------
// FLIGHT OFFERS
// --------------------------------------------------

type FlightOffer struct {
	TotalEmissionsKG string `json:"total_emissions_kg"`
	TotalCurrency    string `json:"total_currency"`
	TotalAmount      string `json:"total_amount"`
	TaxCurrency      string `json:"tax_currency"`
	TaxAmount        string `json:"tax_amount"`

	SupportedPassengerIdentityDocumentTypes []string `json:"supported_passenger_identity_document_types"`

	Slices              []OfferSlice        `json:"slices"`
	PrivateFares        []PrivateFare       `json:"private_fares,omitempty"`
	PaymentRequirements PaymentRequirements `json:"payment_requirements"`

	Passengers                         []Passenger `json:"passengers"`
	PassengerIdentityDocumentsRequired bool        `json:"passenger_identity_documents_required"`
	Partial                            bool        `json:"partial"`
	Owner                              Airline     `json:"owner"`
	LiveMode                           bool        `json:"live_mode"`
	ID                                 string      `json:"id"`
	ExpiresAt                          string      `json:"expires_at"`
	CreatedAt                          string      `json:"created_at"`

	Conditions OfferConditions `json:"conditions"`

	BaseCurrency string `json:"base_currency"`
	BaseAmount   string `json:"base_amount"`
}

// --------------------------------------------------
// OFFER SLICES
// --------------------------------------------------

type OfferSlice struct {
	Segments        []Segment       `json:"segments"`
	OriginType      string          `json:"origin_type"`
	Origin          Airport         `json:"origin"`
	NGSShelf        int             `json:"ngs_shelf"`
	ID              string          `json:"id"`
	FareBrandName   string          `json:"fare_brand_name"`
	Duration        string          `json:"duration"`
	DestinationType string          `json:"destination_type"`
	Destination     Airport         `json:"destination"`
	Conditions      SliceConditions `json:"conditions"`
	ComparisonKey   string          `json:"comparison_key"`
}

// --------------------------------------------------
// SEGMENTS
// --------------------------------------------------

type Segment struct {
	Stops                        []Stop             `json:"stops"`
	Passengers                   []SegmentPassenger `json:"passengers"`
	OriginTerminal               string             `json:"origin_terminal,omitempty"`
	Origin                       Airport            `json:"origin"`
	OperatingCarrierFlightNumber string             `json:"operating_carrier_flight_number"`
	OperatingCarrier             Airline            `json:"operating_carrier"`
	MarketingCarrierFlightNumber string             `json:"marketing_carrier_flight_number"`
	MarketingCarrier             Airline            `json:"marketing_carrier"`
	ID                           string             `json:"id"`
	Duration                     string             `json:"duration"`
	Distance                     string             `json:"distance"`
	DestinationTerminal          string             `json:"destination_terminal,omitempty"`
	Destination                  Airport            `json:"destination"`
	DepartingAt                  string             `json:"departing_at"`
	ArrivingAt                   string             `json:"arriving_at"`
	Aircraft                     Aircraft           `json:"aircraft"`
}

// --------------------------------------------------
// STOPS
// --------------------------------------------------

type Stop struct {
	ID          string  `json:"id"`
	Duration    string  `json:"duration"`
	DepartingAt string  `json:"departing_at"`
	ArrivingAt  string  `json:"arriving_at"`
	Airport     Airport `json:"airport"`
}

// --------------------------------------------------
// SEGMENT PASSENGERS
// --------------------------------------------------

type SegmentPassenger struct {
	PassengerID             string    `json:"passenger_id"`
	FareBasisCode           string    `json:"fare_basis_code"`
	CabinClassMarketingName string    `json:"cabin_class_marketing_name"`
	CabinClass              string    `json:"cabin_class"`
	Cabin                   Cabin     `json:"cabin"`
	Baggages                []Baggage `json:"baggages"`
}

type Cabin struct {
	Name          string    `json:"name"`
	MarketingName string    `json:"marketing_name"`
	Amenities     Amenities `json:"amenities"`
}

type Amenities struct {
	Wifi  *WifiAmenity  `json:"wifi,omitempty"`
	Seat  *SeatAmenity  `json:"seat,omitempty"`
	Power *PowerAmenity `json:"power,omitempty"`
}

type WifiAmenity struct {
	Cost      string `json:"cost"`
	Available bool   `json:"available"`
}

type SeatAmenity struct {
	Type    string `json:"type"`
	Pitch   string `json:"pitch"`
	Legroom string `json:"legroom"`
}

type PowerAmenity struct {
	Available bool `json:"available"`
}

type Baggage struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

// --------------------------------------------------
// AIRLINE / CARRIER
// --------------------------------------------------

type Airline struct {
	Name                    string `json:"name"`
	LogoSymbolURL           string `json:"logo_symbol_url"`
	LogoLockupURL           string `json:"logo_lockup_url"`
	ID                      string `json:"id"`
	IATACode                string `json:"iata_code"`
	ConditionsOfCarriageURL string `json:"conditions_of_carriage_url"`
}

// --------------------------------------------------
// AIRCRAFT
// --------------------------------------------------

type Aircraft struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	IATACode string `json:"iata_code"`
}

// --------------------------------------------------
// FARES
// --------------------------------------------------

type PrivateFare struct {
	Type              string `json:"type"`
	TrackingReference string `json:"tracking_reference"`
	TourCode          string `json:"tour_code"`
	CorporateCode     string `json:"corporate_code"`
}

// --------------------------------------------------
// PAYMENT REQUIREMENTS
// --------------------------------------------------

type PaymentRequirements struct {
	RequiresInstantPayment  bool   `json:"requires_instant_payment"`
	PriceGuaranteeExpiresAt string `json:"price_guarantee_expires_at"`
	PaymentRequiredBy       string `json:"payment_required_by"`
}

// --------------------------------------------------
// CONDITIONS
// --------------------------------------------------

type OfferConditions struct {
	RefundBeforeDeparture *PenaltyCondition `json:"refund_before_departure,omitempty"`
	ChangeBeforeDeparture *PenaltyCondition `json:"change_before_departure,omitempty"`
}

type SliceConditions struct {
	PriorityCheckIn       *bool             `json:"priority_check_in"`
	PriorityBoarding      *bool             `json:"priority_boarding"`
	ChangeBeforeDeparture *PenaltyCondition `json:"change_before_departure,omitempty"`
	AdvanceSeatSelection  bool              `json:"advance_seat_selection"`
}

type PenaltyCondition struct {
	PenaltyCurrency string `json:"penalty_currency"`
	PenaltyAmount   string `json:"penalty_amount"`
	Allowed         bool   `json:"allowed"`
}

// --------------------------------------------------
// Flight Raw Order
// --------------------------------------------------
type FlightRawOrder struct {
	ExternalID  string
	Payload     []byte
	OriginalMsg []byte
}

type FlightSearch struct {
	Data struct {
		Slices []struct {
			Origin        string `json:"origin"`
			Destination   string `json:"destination"`
			DepartureDate string `json:"departure_date"`
		} `json:"slices"`
		Passengers []struct {
			Type string `json:"type"`
		} `json:"passengers"`
	} `json:"data"`
}
