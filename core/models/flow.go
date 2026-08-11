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
