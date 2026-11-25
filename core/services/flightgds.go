package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tourmanager/config"
	"tourmanager/core/models"
	"tourmanager/core/ports"
)

type flightgdsService struct {
	config          config.Config
	flightgdsClient ports.FlightgdsClient
}

// NewDefontanaService crea un nuevo servicio para la integración con Defontana
func NewFlightgdsService(cfg config.Config,
	flightgdsClient ports.FlightgdsClient,
) ports.FlightgdsService {
	return &flightgdsService{
		config:          cfg,
		flightgdsClient: flightgdsClient,
	}
}

// GetOrders orquesta la extracción de órdenes de Wocommerce, el guardado del JSON crudo y su posterior normalización
func (s *flightgdsService) GetFlight(ctx context.Context, flightsearch models.FlightSearch) ([]models.FlightOfferNorm, error) {
	key := s.config.ApikeyDuffel

	rawFlights, err := s.flightgdsClient.FetchFlightsGDS(ctx, key, flightsearch)
	if err != nil {
		return nil, fmt.Errorf("error al consultar vuelos: %w", err)
	}

	var processedFlight []models.FlightOfferNorm

	// 4. Procesar cada orden: almacenar crudo en IntegrationOrder y normalizar en Order

	for _, rawFlight := range rawFlights {

		// Normalizar el JSON crudo al modelo de dominio Order

		normalizedFlight, err := normalizeFlightgds(rawFlight.Payload)
		if err != nil {
			//	errMsg := fmt.Sprintf("error al persistir orden normalizada: %v", err)
			continue
		}

		processedFlight = append(processedFlight, normalizedFlight...)

	}

	return processedFlight, nil

}

// normalizeDefontanaOrder normaliza el payload JSON crudo a los structs Order y OrderItem
func normalizeFlightgds(payload []byte) ([]models.FlightOfferNorm, error) {
	// Estructura del JSON Defontana
	var result []models.FlightOfferNorm

	var flightOffer models.FlightOfferRequest

	if err := json.Unmarshal(payload, &flightOffer); err != nil {
		return nil, fmt.Errorf(
			"error parseando respuesta Duffel: %w",
			err,
		)
	}

	for _, offer := range flightOffer.Data.Offers {

		normalized, err := MapDuffelOffer(offer)
		if err != nil {
			return nil, fmt.Errorf(
				"error convirtiendo oferta %s: %w",
				offer.ID, err,
			)
		}

		result = append(result, normalized)
	}

	return result, nil
}

// MapDuffelOffer convierte una oferta individual.
func MapDuffelOffer(
	offer models.FlightOffer,
) (models.FlightOfferNorm, error) {

	price, err := strconv.ParseFloat(offer.TotalAmount, 64)
	if err != nil {
		return models.FlightOfferNorm{},
			fmt.Errorf("precio inválido: %w", err)
	}

	normalized := models.FlightOfferNorm{
		ID:              offer.ID,
		Provider:        "DUFFEL",
		ProviderOfferID: offer.ID,
		Price:           price,
		Currency:        offer.TotalCurrency,
	}

	// Convertir los slices.
	for i, slice := range offer.Slices {

		normalizedSlice, err := mapSlice(slice)
		if err != nil {
			return models.FlightOfferNorm{},
				fmt.Errorf("slice %d: %w", i, err)
		}

		switch i {
		case 0:
			normalized.Outbound = normalizedSlice

		case 1:
			normalized.Return = &normalizedSlice
		}

		// La duración del slice incluye el tiempo total
		// del trayecto, incluidas las conexiones.
		duration, err := parseISODuration(slice.Duration)
		if err != nil {
			return models.FlightOfferNorm{},
				fmt.Errorf("duración del slice: %w", err)
		}

		normalized.TotalDuration += duration
	}

	// Condiciones generales de la oferta.
	normalized.Conditions = mapOfferConditions(offer)

	return normalized, nil
}

// --------------------------------------------------
// SLICE
// --------------------------------------------------

func mapSlice(
	slice models.OfferSlice,
) (models.FlightSlice, error) {

	result := models.FlightSlice{
		Segments: make([]models.FlightSegment, 0),
	}

	for _, segment := range slice.Segments {

		normalized, err := mapSegment(segment)
		if err != nil {
			return result, err
		}

		result.Segments = append(
			result.Segments,
			normalized,
		)
	}

	// Utilizar el primer segmento para el origen
	// y el último segmento para el destino.
	// Esto evita inconsistencias en los datos del ejemplo.
	if len(result.Segments) > 0 {

		result.Origin = result.Segments[0].Origin

		last := len(result.Segments) - 1

		result.Destination =
			result.Segments[last].Destination
	}

	return result, nil
}

// --------------------------------------------------
// SEGMENT
// --------------------------------------------------

func mapSegment(
	segment models.Segment,
) (models.FlightSegment, error) {

	departure, err := parseDateTime(
		segment.DepartingAt,
		segment.Origin.TimeZone,
	)
	if err != nil {
		return models.FlightSegment{},
			fmt.Errorf("fecha de salida: %w", err)
	}

	arrival, err := parseDateTime(
		segment.ArrivingAt,
		segment.Destination.TimeZone,
	)
	if err != nil {
		return models.FlightSegment{},
			fmt.Errorf("fecha de llegada: %w", err)
	}

	duration, err := parseISODuration(segment.Duration)
	if err != nil {
		return models.FlightSegment{},
			fmt.Errorf("duración del vuelo: %w", err)
	}

	result := models.FlightSegment{
		Airline: models.AirlineNorm{
			Code:    segment.MarketingCarrier.IATACode,
			Name:    segment.MarketingCarrier.Name,
			LogoURL: segment.MarketingCarrier.LogoSymbolURL,
		},

		FlightNumber: segment.MarketingCarrierFlightNumber,

		Origin:      mapAirport(segment.Origin),
		Destination: mapAirport(segment.Destination),

		DepartureTime: departure,
		ArrivalTime:   arrival,
		Duration:      duration,

		Stops:   make([]models.FlightStop, 0),
		Baggage: make([]models.BaggageNorm, 0),
	}

	// Escalas.
	for _, stop := range segment.Stops {

		result.Stops = append(
			result.Stops,
			models.FlightStop{
				Airport: mapAirport(stop.Airport),
			},
		)
	}

	// Equipaje asociado a cada pasajero del segmento.
	for _, passenger := range segment.Passengers {

		for _, baggage := range passenger.Baggages {

			result.Baggage = append(
				result.Baggage,
				models.BaggageNorm{
					Type:     baggage.Type,
					Quantity: baggage.Quantity,
				},
			)
		}
	}

	return result, nil
}

// --------------------------------------------------
// AIRPORT
// --------------------------------------------------

func mapAirport(
	airport models.Airport,
) models.AirportNorm {

	return models.AirportNorm{
		Code:        airport.IATACode,
		Name:        airport.Name,
		City:        airport.CityName,
		CountryCode: airport.IATACountryCode,
		TimeZone:    airport.TimeZone,
	}
}

// --------------------------------------------------
// CONDITIONS
// --------------------------------------------------

func mapOfferConditions(
	offer models.FlightOffer,
) []models.FlightCondition {

	var result []models.FlightCondition

	if condition := offer.Conditions.RefundBeforeDeparture; condition != nil {

		result = append(result, models.FlightCondition{
			Type:            "refund_before_departure",
			Allowed:         condition.Allowed,
			PenaltyAmount:   parseAmount(condition.PenaltyAmount),
			PenaltyCurrency: condition.PenaltyCurrency,
		})
	}

	if condition := offer.Conditions.ChangeBeforeDeparture; condition != nil {

		result = append(result, models.FlightCondition{
			Type:            "change_before_departure",
			Allowed:         condition.Allowed,
			PenaltyAmount:   parseAmount(condition.PenaltyAmount),
			PenaltyCurrency: condition.PenaltyCurrency,
		})
	}

	// Condiciones adicionales de cada slice.
	for i, slice := range offer.Slices {

		if condition := slice.Conditions.ChangeBeforeDeparture; condition != nil {

			result = append(result, models.FlightCondition{
				Type: fmt.Sprintf(
					"slice_%d_change_before_departure", i,
				),
				Allowed:         condition.Allowed,
				PenaltyAmount:   parseAmount(condition.PenaltyAmount),
				PenaltyCurrency: condition.PenaltyCurrency,
			})
		}
	}

	return result
}

// --------------------------------------------------
// HELPERS
// --------------------------------------------------

func parseAmount(value string) float64 {
	result, _ := strconv.ParseFloat(value, 64)
	return result
}

// Convierte PT02H26M, PT45M, PT30S, etc.
// a time.Duration.
var isoDurationRegex = regexp.MustCompile(
	`^PT(?:(\d+(?:\.\d+)?)H)?(?:(\d+(?:\.\d+)?)M)?(?:(\d+(?:\.\d+)?)S)?$`,
)

func parseISODuration(value string) (time.Duration, error) {

	matches := isoDurationRegex.FindStringSubmatch(
		strings.ToUpper(value),
	)

	if matches == nil {
		return 0, fmt.Errorf(
			"formato de duración no válido: %s", value,
		)
	}

	var duration time.Duration

	units := []time.Duration{
		time.Hour,
		time.Minute,
		time.Second,
	}

	for i, match := range matches[1:] {

		if match == "" {
			continue
		}

		number, err := strconv.ParseFloat(match, 64)
		if err != nil {
			return 0, err
		}

		duration += time.Duration(
			number * float64(units[i]),
		)
	}

	return duration, nil
}

// Convierte fechas respetando la zona horaria del aeropuerto.
func parseDateTime(
	value string,
	timezone string,
) (time.Time, error) {

	if timezone == "" {
		timezone = "UTC"
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"zona horaria %s: %w", timezone, err,
		)
	}

	// Primero intenta RFC3339 con zona horaria explícita.
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed, nil
	}

	// Duffel también puede entregar fechas sin offset.
	return time.ParseInLocation(
		"2006-01-02T15:04:05",
		value,
		location,
	)
}
