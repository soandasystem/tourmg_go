package models

type InitFlowPaymentReq struct {
	Monto           int    `json:"mpagar"`
	Identificador   string `json:"identificador"`
	ValorCuota      int    `json:"valorcuota"`
	NroCuotas       int    `json:"nrocuotas"`
	FechaInicial    string `json:"fechainicial"`
	CompanyID       int64  `json:"company_id"`
	SaleID          int64  `json:"sale_id"`
	CursoID         int64  `json:"curso_id"`
	UserRut         string `json:"user_rut"`
	Urlreturn       string `json:"urlreturn"`
	Urlconfirmation string `json:"urlconfirmation"`
}

type InitFlowPaymentResp struct {
	RedirectURL string `json:"redirect_url"`
}

type TokenRequest struct {
	Token string `json:"token"`
}

type TokenResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type FlowResponse struct {
	Amount        string `json:"amount"`
	CommerceOrder string `json:"commerceOrder"`
	Currency      string `json:"currency"`
	FlowOrder     string `json:"flowOrder"`
	Merchantid    string `json:"merchantId"`
	Optional      struct {
		Venta  string `json:"venta"`
		Alumno string `json:"alumno"`
	} `json:"optional"`
	Payer       string `json:"payer"`
	PaymentData struct {
		Amount         string `json:"amount"`
		Balance        string `json:"balance"`
		Conversiondate string `json:"conversionDate"`
		Conversionrate string `json:"conversionRate"`
		Currency       string `json:"currency"`
		Date           string `json:"date"`
		Fee            string `json:"fee"`
		Media          string `json:"media"`
		Transferdate   string `json:"transferDate"`
	} `json:"paymentData"`
	PendingInfo struct {
		Date  string `json:"date"`
		Media string `json:"media"`
	} `json:"pending_info"`
	RequestDate string `json:"requestDate"`
	Status      string `json:"status"`
	Subject     string `json:"subject"`
}

type FlowListResponse struct {
	Items      []FlowResponse `json:"items"`
	TotalCount int64          `json:"totalCount"`
}
