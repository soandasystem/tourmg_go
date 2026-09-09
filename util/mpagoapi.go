package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"tourmanager/core/models"
)

// MPagoAPI encapsula la comunicación HTTP con la API de Mercado Pago
type MPagoAPI struct {
	AccessToken string
	BaseURL     string
	Client      *http.Client
}

// NewMPagoAPI crea una nueva instancia del cliente de Mercado Pago
func NewMPagoAPI(accessToken string) *MPagoAPI {
	return &MPagoAPI{
		AccessToken: accessToken,
		BaseURL:     "https://api.mercadopago.com",
		Client:      &http.Client{Timeout: 15 * time.Second},
	}
}

// GetPayment consulta la información de un pago por su payment_id
func (api *MPagoAPI) GetPayment(paymentID string) (*models.MPagoPaymentData, error) {
	reqURL := fmt.Sprintf("%s/v1/payments/%s", api.BaseURL, paymentID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+api.AccessToken)

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error mercadopago code %d: %s", resp.StatusCode, string(body))
	}

	var payment models.MPagoPaymentData
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

// SearchPaymentByPreference busca pagos asociados al external_reference (preference_id)
func (api *MPagoAPI) SearchPaymentByPreference(preferenceID string) (*models.MPagoPaymentData, error) {
	reqURL := fmt.Sprintf("%s/v1/payments/search?sort=date_created&criteria=desc&external_reference=%s", api.BaseURL, preferenceID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+api.AccessToken)

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error mercadopago search code %d: %s", resp.StatusCode, string(body))
	}

	var searchResp models.MPagoSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	if len(searchResp.Results) == 0 {
		return nil, fmt.Errorf("no se encontraron pagos para la preferencia %s", preferenceID)
	}

	return &searchResp.Results[0], nil
}

// CreatePreference crea una preferencia de pago en Mercado Pago (/checkout/preferences)
func (api *MPagoAPI) CreatePreference(data map[string]interface{}) (map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s/checkout/preferences", api.BaseURL)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api.AccessToken)

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error mercadopago create preference code %d: %s", resp.StatusCode, string(body))
	}

	var preference map[string]interface{}
	if err := json.Unmarshal(body, &preference); err != nil {
		return nil, err
	}

	return preference, nil
}

