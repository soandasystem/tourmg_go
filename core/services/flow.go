package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"tourmanager/config"
	"tourmanager/core/models"
	"tourmanager/core/ports"
	"tourmanager/util"
)

type flowService struct {
	config                 config.Config
	gatewaysRepo           ports.GatewaysRepository
	gatewayscRepo          ports.GatewayscRepository
	saleRepo               ports.SaleRepository
	cursoRepo              ports.CursoRepository
	paymentRepo            ports.PaymentRepository
	installmentsRepo       ports.InstallmentsRepository
	paymentInstallmentRepo ports.PaymentInstallmentRepository
}

// NewFlowService creates a new flow service
func NewFlowService(cfg config.Config, gatewaysRepo ports.GatewaysRepository, gatewayscRepo ports.GatewayscRepository, saleRepo ports.SaleRepository, cursoRepo ports.CursoRepository, paymentRepo ports.PaymentRepository, installmentsRepo ports.InstallmentsRepository, paymentInstallmentRepo ports.PaymentInstallmentRepository) ports.FlowService {
	return &flowService{
		config:                 cfg,
		gatewaysRepo:           gatewaysRepo,
		gatewayscRepo:          gatewayscRepo,
		saleRepo:               saleRepo,
		cursoRepo:              cursoRepo,
		paymentRepo:            paymentRepo,
		installmentsRepo:       installmentsRepo,
		paymentInstallmentRepo: paymentInstallmentRepo,
	}
}

func (s *flowService) InitPayment(ctx context.Context, req models.InitFlowPaymentReq) (models.InitFlowPaymentResp, error) {
	// 1. Buscar la venta
	saleIDStr := strconv.FormatInt(req.SaleID, 10)
	saleResult, err := s.saleRepo.GetByID(ctx, saleIDStr)
	if err != nil {
		return models.InitFlowPaymentResp{}, fmt.Errorf("error fetching sale: %v", err)
	}
	if saleResult == nil {
		return models.InitFlowPaymentResp{}, fmt.Errorf("sale not found")
	}
	_, ok := saleResult.(*models.SaleResp)
	if !ok {
		return models.InitFlowPaymentResp{}, fmt.Errorf("sale no encontrado en response")
	}

	// 2. Buscar el curso
	cursoFilter := map[string]interface{}{"id": req.CursoID}
	cursoResult, err := s.cursoRepo.Get(ctx, cursoFilter, nil, nil)
	if err != nil {
		return models.InitFlowPaymentResp{}, fmt.Errorf("error fetching curso: %v", err)
	}
	if len(cursoResult) == 0 {
		return models.InitFlowPaymentResp{}, fmt.Errorf("curso not found")
	}
	cursoList, ok := cursoResult[0].(models.CursoListResponse)
	if !ok || len(cursoList.Items) == 0 {
		return models.InitFlowPaymentResp{}, fmt.Errorf("curso no encontrado en response")
	}
	curso := cursoList.Items[0]

	// 2. Get Gateway Config
	gatewayFilter := map[string]interface{}{"gateway_id": 3, "company_id": req.CompanyID} // 3 = Flow
	gatewayResult, err := s.gatewaysRepo.Get(ctx, gatewayFilter, nil, nil)
	if err != nil {
		return models.InitFlowPaymentResp{}, fmt.Errorf("error fetching gateway: %v", err)
	}
	if len(gatewayResult) == 0 {
		return models.InitFlowPaymentResp{}, fmt.Errorf("gateway not found")
	}

	gatewayList, ok := gatewayResult[0].(models.GatewaysListResponse)
	if !ok || len(gatewayList.Items) == 0 {
		return models.InitFlowPaymentResp{}, fmt.Errorf("gateway no encontrado en response")
	}
	gateway := gatewayList.Items[0]

	flowAPIKey := gateway.AdditionalConfig.FlowAPIKey
	flowSecretKey := gateway.AdditionalConfig.FlowSecretKey

	// 3. Prepare Flow params
	optionalData := map[string]string{
		"venta":  strconv.FormatInt(req.SaleID, 10),
		"alumno": req.UserRut,
	}
	optionalJSON, _ := json.Marshal(optionalData)

	params := map[string]string{
		"commerceOrder":   req.Identificador,
		"subject":         "Pago cuota viaje estudio",
		"currency":        "CLP",
		"amount":          strconv.Itoa(req.Monto),
		"email":           curso.Correo,
		"paymentMethod":   "9",
		"urlConfirmation": "https://flowresponse.onrender.com/token", // Webhook
		"urlReturn":       req.Urlreturn,                             // Return URL
		"optional":        string(optionalJSON),
	}
	flowAPI := util.NewFlowAPI(flowAPIKey, flowSecretKey, s.config.FlowAPIURL)
	response, err := flowAPI.Send("payment/create", params, "POST")

	if err != nil {
		return models.InitFlowPaymentResp{}, fmt.Errorf("error comunicándose con Flow: %v", err)
	}

	flowURL, hasUrl := response["url"].(string)
	flowToken, hasToken := response["token"].(string)
	if !hasUrl || !hasToken {
		return models.InitFlowPaymentResp{}, fmt.Errorf("respuesta inválida de flow: %v", response)
	}
	//grabar el pago solo en payments y en notes dejarlo como PENDIENTE
	ingreso := models.CreatePaymentReq{
		PassengerId:     req.CursoID,
		Amount:          float32(req.Monto),
		PaymentMethod:   "flow",
		PaymentDate:     time.Now(),
		Identifier:      req.Identificador,
		Notes:           "Pendiente",
		TransactionRef:  "",
		TransactionType: "",
		CardNumber:      "",
		AuthCode:        "",
		PaymentToken:    flowToken,
		CompanyId:       req.CompanyID,
		SaleId:          req.SaleID,
	}

	insertedID, err := s.paymentRepo.Create(ctx, ingreso)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Pago creado exitosamente con ID: " + insertedID)
	}

	destinationURL := fmt.Sprintf("%s?token=%s", flowURL, flowToken)
	return models.InitFlowPaymentResp{RedirectURL: destinationURL}, nil
}

func (s *flowService) FlowToken(ctx context.Context, token string) (models.TokenResponse, error) {

	//variables de respuesta
	status_servicio := "Ok"
	error_servicio := "terminado correctamente"

	//----------------------
	//Buscar el ingreso por token flow
	var result []interface{}
	var err error
	maxRetries := 7
	found := false

	for i := 0; i < maxRetries; i++ {
		tokenStr := map[string]interface{}{"payment_token": token}
		result, err = s.paymentRepo.Get(ctx, tokenStr, nil, nil)
		if err == nil {
			found = true
			break
		}

		fmt.Printf("Intento %d/%d: ingreso no encontrado, esperando...\n", i+1, maxRetries)
		time.Sleep(500 * time.Millisecond)
	}

	if !found {
		fmt.Println("No se encontró ingreso con el token tras varios intentos")
		status_servicio = "No"
		error_servicio = "Ingreso no encontrado tras varios intentos"
		return models.TokenResponse{
			Status:  status_servicio,
			Message: error_servicio,
		}, nil
	}

	paymentResult, ok := result[0].(models.CreatePaymentReq)
	if !ok {
		return models.TokenResponse{
			Status:  "Error",
			Message: "No se pudo convertir el ingreso",
		}, nil
	}
	//---------------------
	//con el company id del ingeso buscar las key de flow
	gatewayFilter := map[string]interface{}{"gateway_id": 3, "company_id": paymentResult.CompanyId} // 3 = Flow
	gatewayResult, err := s.gatewaysRepo.Get(ctx, gatewayFilter, nil, nil)
	if err != nil {
		return models.TokenResponse{
			Status:  "Error",
			Message: "error fetching gateway: %v",
		}, nil
	}
	if len(gatewayResult) == 0 {
		return models.TokenResponse{
			Status:  "Error",
			Message: "gateway not found",
		}, nil
	}

	gatewayList, ok := gatewayResult[0].(models.GatewaysListResponse)
	if !ok || len(gatewayList.Items) == 0 {
		return models.TokenResponse{
			Status:  "Error",
			Message: "gateway no encontrado en response",
		}, nil
	}
	gateway := gatewayList.Items[0]

	flowAPIKey := gateway.AdditionalConfig.FlowAPIKey
	flowSecretKey := gateway.AdditionalConfig.FlowSecretKey

	// Parámetros para el método Send
	// Llamar a la API de Flow
	params := map[string]string{"token": token}

	flowAPI := util.NewFlowAPI(flowAPIKey, flowSecretKey, s.config.FlowAPIURL)
	response, err := flowAPI.Send("payment/getStatus", params, "GET")

	if err != nil {
		fmt.Println("Error:", err)

	}

	fmt.Println("respuesta ", response)
	// Extraer los valores del response y asignarlos a la estructura PaymentResponse
	continuaOperacion := true
	paymentResponse, err := parseResponse(response)
	if err != nil {
		fmt.Println("Error parsing response", err)
		status_servicio = "No"
		error_servicio = "Error al convertir el body a JSON "
		continuaOperacion = false
	}

	// Procesar la respuesta
	if continuaOperacion {
		status := paymentResponse.Status
		switch status {
		case "1":
			fmt.Println("Pendiente")
			paymentNotes := "Pendiente"
			ingreso := models.UpdatePaymentReq{
				Notes: &paymentNotes,
			}
			s.paymentRepo.Update(ctx, paymentResult.ID, ingreso)
		case "2":
			fmt.Println("Pagado")
			// Convertir la cadena a float64
			fechaAuto := paymentResponse.RequestDate
			media := paymentResponse.PaymentData.Media

			fecha_Auto, err := time.Parse("2006-01-02 15:04:05", fechaAuto)
			if err != nil {
				fmt.Println("Error al parsear la fecha:", err)
				status_servicio = "No"
				error_servicio = "Error al parsear la fecha fechaauto"

			}
			//busco payment que este note = pendiente y identifier = identifier
			paymentNotes := "Pagado"
			ingreso := models.UpdatePaymentReq{
				TransactionType: &media,
				AuthDate:        &fecha_Auto,
				Notes:           &paymentNotes,
			}
			s.paymentRepo.Update(ctx, paymentResult.ID, ingreso)
			//en installments busco el balance <> 0 y sumo al paid_amount
			//balace = amount - paid_amount
			installmentsFilter := map[string]interface{}{"passenger_id": paymentResult.PassengerId, "company_id": paymentResult.CompanyId, "sale_id": paymentResult.SaleId} // 3 = Flow
			installmentsResult, err := s.gatewaysRepo.Get(ctx, installmentsFilter, nil, nil)

			var amount float32
			var newpaidAmount float32
			var newBalance float32
			var ID string
			for _, item := range installmentsResult {
				installment := item.(map[string]interface{})

				balance, _ := installment["balance"].(float64)

				if balance != 0 {
					ID, _ = installment["id"].(string)
					amount, _ = installment["amount"].(float32)
					newpaidAmount, _ = installment["paid_amount"].(float32)
					found = true
					break
				}
			}
			newpaidAmount += float32(paymentResult.Amount)
			newBalance = amount - newpaidAmount
			if newBalance != 0 {
				status = "PARTIAL"
			} else {
				status = "PAID"
			}

			installmentUpdate := models.UpdateInstallmentReq{
				PaidAmount: &newpaidAmount,
				Balance:    &newBalance,
				Status:     &status,
			}
			s.installmentsRepo.Update(ctx, ID, installmentUpdate)
			//grabo payment_installments con el monto pagado
			id_ip, err := strconv.ParseInt(paymentResult.ID, 10, 64)
			id_in, err := strconv.ParseInt(ID, 10, 64)
			payment_installments := models.CreatePaymentInstallmentReq{
				PaymentId:     id_ip,
				InstallmentId: id_in,
				AppliedAmount: float32(paymentResult.Amount),
			}
			s.paymentInstallmentRepo.Create(ctx, payment_installments)

		case "3":
			fmt.Println("Transacción Rechazada")
			paymentNotes := "Rechazado"
			ingreso := models.UpdatePaymentReq{
				Notes: &paymentNotes,
			}
			s.paymentRepo.Update(ctx, paymentResult.ID, ingreso)
		case "4":
			fmt.Println("Transacción Anulada")
			paymentNotes := "Anulado"
			ingreso := models.UpdatePaymentReq{
				Notes: &paymentNotes,
			}
			s.paymentRepo.Update(ctx, paymentResult.ID, ingreso)
		default:
			fmt.Println("Estado desconocido")
			paymentNotes := "Desconocido"
			ingreso := models.UpdatePaymentReq{
				Notes: &paymentNotes,
			}
			s.paymentRepo.Update(ctx, paymentResult.ID, ingreso)
		}

	}
	return models.TokenResponse{
		Status:  "Success",
		Message: "datos guardados ok",
	}, nil
}

func parseResponse(response map[string]interface{}) (models.PaymentResponse, error) {

	var paymentResponse models.PaymentResponse

	// Asignar valores directamente desde el map
	paymentResponse.Amount = response["amount"].(string)
	paymentResponse.CommerceOrder = response["commerceOrder"].(string)
	paymentResponse.Currency = response["currency"].(string)
	paymentResponse.FlowOrder = fmt.Sprint(response["flowOrder"])   // Convertir a string
	paymentResponse.Merchantid = fmt.Sprint(response["merchantId"]) // Si es nil, se asigna un valor vacío

	// Optional
	optional := response["optional"].(map[string]interface{})
	paymentResponse.Optional.Venta = fmt.Sprint(optional["venta"])
	paymentResponse.Optional.Alumno = optional["alumno"].(string)

	// Payer
	paymentResponse.Payer = response["payer"].(string)

	// PaymentData
	paymentData := response["paymentData"].(map[string]interface{})
	paymentResponse.PaymentData.Amount = fmt.Sprint(paymentData["amount"])
	paymentResponse.PaymentData.Balance = fmt.Sprint(paymentData["balance"])
	paymentResponse.PaymentData.Conversiondate = fmt.Sprint(paymentData["conversionDate"])
	paymentResponse.PaymentData.Conversionrate = fmt.Sprint(paymentData["conversionRate"])
	paymentResponse.PaymentData.Currency = fmt.Sprint(paymentData["currency"])
	paymentResponse.PaymentData.Date = fmt.Sprint(paymentData["date"])
	paymentResponse.PaymentData.Fee = fmt.Sprint(paymentData["fee"])
	paymentResponse.PaymentData.Media = fmt.Sprint(paymentData["media"])
	paymentResponse.PaymentData.Transferdate = fmt.Sprint(paymentData["transferDate"])

	// PendingInfo
	pendingInfo := response["pending_info"].(map[string]interface{})
	paymentResponse.PendingInfo.Date = fmt.Sprint(pendingInfo["date"])
	paymentResponse.PendingInfo.Media = fmt.Sprint(pendingInfo["media"])

	// RequestDate
	paymentResponse.RequestDate = response["requestDate"].(string)

	// Status
	paymentResponse.Status = fmt.Sprint(response["status"])

	// Subject
	paymentResponse.Subject = response["subject"].(string)

	return paymentResponse, nil

}
