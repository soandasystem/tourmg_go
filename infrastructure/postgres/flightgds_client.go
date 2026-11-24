package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"tourmanager/core/models"
	"tourmanager/core/ports"
)

type flightgdsClient struct {
	httpClient *http.Client
}

// NewPrestashopClient crea un nuevo cliente para la API de Paris
func NewFlightgdsClient() ports.FlightgdsClient {
	return &flightgdsClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchOrders obtiene las órdenes desde la API de Shopify usando la configuración provista y paginación con nextCursor
func (c *flightgdsClient) FetchFlightsGDS(ctx context.Context, key string) ([]models.FlightRawOrder, error) {

	baseURL := "https://api.duffel.com/air/offer_requests"

	now := time.Now()
	departureDate := now.AddDate(0, 0, +5).Format("2006-01-02")
	returnDate := now.AddDate(0, 0, +8).Format("2006-01-02")
	// Construir solicitud
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"slices": []map[string]interface{}{
				{
					"origin":         "SCL",
					"destination":    "MIA",
					"departure_date": departureDate,
				},
				{
					"origin":         "MIA",
					"destination":    "SCL",
					"departure_date": returnDate,
				},
			},
			"passengers": []map[string]interface{}{
				{
					"type": "adult",
				},
			},
		},
	}

	// Convertir estructura a JSON
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error convirtiendo solicitud a JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error al crear solicitud HTTP para prestashop: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Duffel-Version", "v2")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al contactar la API de prestashop: %w", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta de prestashop: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API prestashop respondió con status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	pageOrders, err := c.parseFlightgdsPage(bodyBytes, key, ctx)
	if err != nil {
		return nil, err
	}

	return pageOrders, nil
}

// parseParisOrdersPage divide la respuesta en órdenes
func (c *flightgdsClient) parseFlightgdsPage(raw []byte, key string, ctx context.Context) ([]models.FlightRawOrder, error) {
	var response struct {
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf(
			"error parseando respuesta FlightGDS: %w",
			err,
		)
	}

	// Paris respondió correctamente pero no hay órdenes
	if len(response.Data) == 0 {
		fmt.Println("no hay vuelos")
		return []models.FlightRawOrder{}, nil
	}

	Data, err := c.rawMessagesFlightgds(
		response.Data, key, ctx,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"error procesando offer requests: %w",
			err,
		)
	}

	return Data, nil
}

func (c *flightgdsClient) rawMessagesFlightgds(messages json.RawMessage, key string, ctx context.Context) ([]models.FlightRawOrder, error) {

	//originalMsg := msg // Guardamos el payload original
	var idFinder struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(messages, &idFinder); err != nil {
		return nil, fmt.Errorf("error parseando vuelos: %w", err)
	}

	if idFinder.ID == "" {
		return nil, fmt.Errorf("Duffel no devolvió el ID de offer request")
	}
	// Busque la orden segun el id obtenido
	customURL := "https://api.duffel.com/air/offer_requests/" + idFinder.ID
	flightOfferResp, err := c.getOfferRequestDuffel(customURL, key, ctx)
	if err != nil {
		fmt.Printf("advertencia: no se pudo obtener la orden %s: %v\n", idFinder.ID, err)
	}

	// Guardar la respuesta completa
	var flightgds models.FlightOfferRequest

	if err := json.Unmarshal(flightOfferResp, &flightgds); err != nil {
		return nil, fmt.Errorf("error procesando respuesta Duffel: %w", err)
	}

	unifiedPayload, err := json.Marshal(flightgds)
	if err != nil {
		return nil, fmt.Errorf("error serializando respuesta Duffel: %w", err)
	}

	rawflights := []models.FlightRawOrder{
		{
			ExternalID:  idFinder.ID,
			Payload:     unifiedPayload,
			OriginalMsg: flightOfferResp,
		},
	}

	return rawflights, nil
}

func (c *flightgdsClient) getOfferRequestDuffel(orderURL string, key string, ctx context.Context) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, orderURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error al preparar solicitud de orden: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Duffel-Version", "v2")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar petición de token: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta de token: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("echazó la autenticación (status %d): %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, nil

}
