package models

type InitMPagoPaymentReq struct {
	Monto         int    `json:"mpagar"`
	Identificador string `json:"identificador"`
	ValorCuota    int    `json:"valorcuota"`
	NroCuotas     int    `json:"nrocuotas"`
	FechaInicial  string `json:"fechainicial"`
	CompanyID     int64  `json:"company_id"`
	SaleID        int64  `json:"sale_id"`
	CursoID       int64  `json:"curso_id"`
	UserRut       string `json:"user_rut"`
}

type InitMPagoPaymentResp struct {
	RedirectURL string `json:"redirect_url"`
}

type InitMPagoResp struct {
	IDIngreso     string `json:"id_ingreso"`
	Identificador string `json:"identificador"`
	MontoPagado   int64  `json:"monto_pagado"`
	Preference    string `json:"preference"`
	PublicKey     string `json:"public_key"`
}

type TokenMpRequest struct {
	Token string `json:"token"`
}

type TokenMpResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MPagoWebhookReq representa el payload recibido en el webhook de Mercado Pago
type MPagoWebhookReq struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

// VerifyMPagoReq representa los datos necesarios para verificar un pago
type VerifyMPagoReq struct {
	PaymentID    string `json:"payment_id"`
	PreferenceID string `json:"preference_id"`
	Source       string `json:"source"` // "webhook" o "redirect"
	CompanyID    int64  `json:"company_id"`
	Subdominio   string `json:"subdominio"`
}

// MPagoPaymentData representa la información de un pago retornado por la API de Mercado Pago
type MPagoPaymentData struct {
	ID                int64   `json:"id"`
	DateCreated       string  `json:"date_created"`
	DateApproved      string  `json:"date_approved"`
	Status            string  `json:"status"` // approved, pending, rejected
	StatusDetail      string  `json:"status_detail"`
	PaymentMethodID   string  `json:"payment_method_id"`
	TransactionAmount float64 `json:"transaction_amount"`
	ExternalReference string  `json:"external_reference"`
	CollectorID       int64   `json:"collector_id"`
	AuthorizationCode string  `json:"authorization_code"`
	Order             struct {
		ID string `json:"id"`
	} `json:"order"`
	Card struct {
		LastFourDigits string `json:"last_four_digits"`
	} `json:"card"`
}

// MPagoSearchResponse representa la respuesta de búsqueda de pagos de Mercado Pago
type MPagoSearchResponse struct {
	Results []MPagoPaymentData `json:"results"`
}

// VerifyMPagoResp representa la respuesta tras procesar y verificar el pago
type VerifyMPagoResp struct {
	Status            string  `json:"status"`
	Message           string  `json:"message"`
	PaymentID         int64   `json:"payment_id,omitempty"`
	TransactionAmount float64 `json:"transaction_amount,omitempty"`
	ExternalReference string  `json:"external_reference,omitempty"`
	RedirectURL       string  `json:"redirect_url,omitempty"`
}
