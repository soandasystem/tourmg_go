package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"tourmanager/config"
	"tourmanager/core/models"
	"tourmanager/core/ports"
	"tourmanager/util"
)

type mpagoService struct {
	config                 config.Config
	gatewaysRepo           ports.GatewaysRepository
	gatewayscRepo          ports.GatewayscRepository
	saleRepo               ports.SaleRepository
	cursoRepo              ports.CursoRepository
	paymentRepo            ports.PaymentRepository
	installmentsRepo       ports.InstallmentsRepository
	paymentInstallmentRepo ports.PaymentInstallmentRepository
	schemaRegistryRepo     ports.SchemaRegistryRepository
}

// NewMpagoService creates a new MercadoPago service
func NewMPagoService(cfg config.Config, gatewaysRepo ports.GatewaysRepository, gatewayscRepo ports.GatewayscRepository, saleRepo ports.SaleRepository, cursoRepo ports.CursoRepository, paymentRepo ports.PaymentRepository, installmentsRepo ports.InstallmentsRepository, paymentInstallmentRepo ports.PaymentInstallmentRepository, schemaRegistryRepo ports.SchemaRegistryRepository) ports.MPagoService {
	return &mpagoService{
		config:                 cfg,
		gatewaysRepo:           gatewaysRepo,
		gatewayscRepo:          gatewayscRepo,
		saleRepo:               saleRepo,
		cursoRepo:              cursoRepo,
		paymentRepo:            paymentRepo,
		installmentsRepo:       installmentsRepo,
		paymentInstallmentRepo: paymentInstallmentRepo,
		schemaRegistryRepo:     schemaRegistryRepo,
	}
}

func (s *mpagoService) InitPayment(ctx context.Context, req models.InitMPagoPaymentReq) (models.InitMPagoResp, error) {
	// 1. Buscar la venta
	saleIDStr := strconv.FormatInt(req.SaleID, 10)
	saleResult, err := s.saleRepo.GetByID(ctx, saleIDStr)
	if err != nil {
		return models.InitMPagoResp{}, fmt.Errorf("error fetching sale: %v", err)
	}
	if saleResult == nil {
		return models.InitMPagoResp{}, fmt.Errorf("sale not found")
	}
	_, ok := saleResult.(*models.SaleResp)
	if !ok {
		return models.InitMPagoResp{}, fmt.Errorf("sale no encontrado en response")
	}

	// 2. Buscar el curso
	cursoFilter := map[string]interface{}{"id": req.CursoID}
	cursoResult, err := s.cursoRepo.Get(ctx, cursoFilter, nil, nil)
	if err != nil {
		return models.InitMPagoResp{}, fmt.Errorf("error fetching curso: %v", err)
	}
	if len(cursoResult) == 0 {
		return models.InitMPagoResp{}, fmt.Errorf("curso not found")
	}
	cursoList, ok := cursoResult[0].(models.CursoListResponse)
	if !ok || len(cursoList.Items) == 0 {
		return models.InitMPagoResp{}, fmt.Errorf("curso no encontrado en response")
	}

	// 3. Obtener configuración del gateway de Mercado Pago (gateway_id = 1)
	gatewayFilter := map[string]interface{}{"gateway_id": 1, "company_id": req.CompanyID}
	gatewayResult, err := s.gatewaysRepo.Get(ctx, gatewayFilter, nil, nil)
	if err != nil {
		return models.InitMPagoResp{}, fmt.Errorf("error fetching gateway: %v", err)
	}
	if len(gatewayResult) == 0 {
		return models.InitMPagoResp{}, fmt.Errorf("gateway not found")
	}

	gatewayList, ok := gatewayResult[0].(models.GatewaysListResponse)
	if !ok || len(gatewayList.Items) == 0 {
		return models.InitMPagoResp{}, fmt.Errorf("gateway no encontrado en response")
	}
	gateway := gatewayList.Items[0]

	publicKey := gateway.AdditionalConfig.MpPublickey
	accessToken := gateway.AdditionalConfig.MpAccesstoken

	if accessToken == "" {
		return models.InitMPagoResp{}, fmt.Errorf("access token de Mercado Pago no configurado")
	}

	// 4. Crear preferencia en Mercado Pago
	mpAPI := util.NewMPagoAPI(accessToken)

	preferenceData := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"id":          req.Identificador,
				"title":       "Pago cuota viaje estudio",
				"quantity":    1,
				"unit_price":  req.Monto,
				"currency_id": "CLP",
			},
		},
		"external_reference": req.Identificador,
		"back_urls": map[string]string{
			"success": "https://tourmg-go.onrender.com/api/v3.5/mpago/verificar",
			"failure": "https://tourmg-go.onrender.com/api/v3.5/mpago/verificar",
			"pending": "https://tourmg-go.onrender.com/api/v3.5/mpago/verificar",
		},
		"auto_return": "approved",
		"binary_mode": true,
	}

	prefResp, err := mpAPI.CreatePreference(preferenceData)
	if err != nil {
		return models.InitMPagoResp{}, fmt.Errorf("error comunicándose con Mercado Pago: %v", err)
	}

	preferenceID, ok := prefResp["id"].(string)
	if !ok || preferenceID == "" {
		return models.InitMPagoResp{}, fmt.Errorf("respuesta inválida de Mercado Pago: id de preferencia no encontrado")
	}

	// 5. Grabar el ingreso en payments con estado "En Proceso"
	ingreso := models.CreatePaymentReq{
		PassengerId:     req.CursoID,
		Amount:          float32(req.Monto),
		PaymentMethod:   "MP",
		PaymentDate:     time.Now(),
		Identifier:      req.Identificador,
		Notes:           "Pagos",
		TransactionRef:  preferenceID,
		TransactionType: "",
		CardNumber:      "",
		AuthCode:        "",
		PaymentToken:    preferenceID,
		CompanyId:       req.CompanyID,
		SaleId:          req.SaleID,
		Author:          "admin",
		State:           "En Proceso",
	}

	insertedID, err := s.paymentRepo.Create(ctx, ingreso)
	if err != nil {
		fmt.Printf("Error al registrar pago en payments: %v\n", err)
	}

	// 6. Registrar en schemaRegistry en schema global
	schemaVal, _ := ctx.Value("schema").(string)
	if schemaVal == "" {
		schemaVal = "global"
	}

	tokenRegistry := models.TokenSchemaRegistry{
		Token:      preferenceID,
		SchemaName: schemaVal,
		CompanyId:  req.CompanyID,
		Active:     true,
	}

	ctxGlobal := context.WithValue(ctx, "schema", "global")
	_, err = s.schemaRegistryRepo.Create(ctxGlobal, tokenRegistry)
	if err != nil {
		return models.InitMPagoResp{}, fmt.Errorf("error creando registro en schema: %v", err)
	}

	// 7. Retornar modelo InitMPagoResp
	return models.InitMPagoResp{
		IDIngreso:     insertedID,
		Identificador: req.Identificador,
		MontoPagado:   int64(req.Monto),
		Preference:    preferenceID,
		PublicKey:     publicKey,
	}, nil
}

func (s *mpagoService) VerifyPayment(ctx context.Context, req models.VerifyMPagoReq) (models.VerifyMPagoResp, error) {
	// 1. Obtener configuración del gateway de Mercado Pago (gateway_id = 1)
	gatewayFilter := map[string]interface{}{"gateway_id": 1}
	if req.CompanyID > 0 {
		gatewayFilter["company_id"] = req.CompanyID
	}
	gatewayResult, err := s.gatewaysRepo.Get(ctx, gatewayFilter, nil, nil)
	if err != nil || len(gatewayResult) == 0 {
		return models.VerifyMPagoResp{}, fmt.Errorf("gateway de Mercado Pago no configurado")
	}

	gatewayList, ok := gatewayResult[0].(models.GatewaysListResponse)
	if !ok || len(gatewayList.Items) == 0 {
		return models.VerifyMPagoResp{}, fmt.Errorf("datos de gateway no encontrados")
	}
	gateway := gatewayList.Items[0]
	accessToken := gateway.AdditionalConfig.MpAccesstoken
	expectedUserID := gateway.AdditionalConfig.MpUsersid

	if accessToken == "" {
		return models.VerifyMPagoResp{}, fmt.Errorf("access token de Mercado Pago no configurado")
	}

	// 2. Consultar el pago en la API de Mercado Pago
	mpAPI := util.NewMPagoAPI(accessToken)
	var payment *models.MPagoPaymentData

	if req.PaymentID != "" {
		payment, err = mpAPI.GetPayment(req.PaymentID)
	} else if req.PreferenceID != "" {
		payment, err = mpAPI.SearchPaymentByPreference(req.PreferenceID)
	} else {
		return models.VerifyMPagoResp{}, fmt.Errorf("parámetros requeridos no encontrados (payment_id o preference_id)")
	}

	if err != nil {
		return models.VerifyMPagoResp{}, fmt.Errorf("error al consultar el pago en Mercado Pago: %v", err)
	}

	// 3. Validación de seguridad (collector_id)
	if expectedUserID != "" && strconv.FormatInt(payment.CollectorID, 10) != expectedUserID {
		return models.VerifyMPagoResp{}, fmt.Errorf("pago no válido para esta cuenta")
	}

	// 4. Buscar el ingreso en la base de datos por external_reference (identificador)
	paymentFilter := map[string]interface{}{"identifier": payment.ExternalReference}
	pResult, err := s.paymentRepo.Get(ctx, paymentFilter, nil, nil)
	if err != nil || len(pResult) == 0 {
		return models.VerifyMPagoResp{}, fmt.Errorf("pago no encontrado con identificador: %s", payment.ExternalReference)
	}

	pList, ok := pResult[0].(models.PaymentListResponse)
	if !ok || len(pList.Items) == 0 {
		return models.VerifyMPagoResp{}, fmt.Errorf("ingreso no encontrado en base de datos")
	}
	dbPayment := pList.Items[0]

	// 4.1 con el compay_id buscar en company la empresa con el schema=global
	companyFilter := map[string]interface{}{"id": dbPayment.CompanyId, "schema": "global"}
	companyResult, err := s.companyRepo.Get(ctx, companyFilter, nil, nil)
	if err != nil || len(companyResult) == 0 {
		return models.VerifyMPagoResp{}, fmt.Errorf("empresa no encontrada con identificador: %s", dbPayment.CompanyId)
	}
	companyList, ok := companyResult[0].(models.CompanyListResponse)
	if !ok || len(companyList.Items) == 0 {
		return models.VerifyMPagoResp{}, fmt.Errorf("ingreso no encontrado en base de datos")
	}
	company := companyList.Items[0]

	// 5. Procesar según el estado de Mercado Pago
	var message string
	switch payment.Status {
	case "approved":
		message = fmt.Sprintf("Pago aprobado! ID: %d", payment.ID)
		state := "Pagado"
		media := payment.PaymentMethodID
		authDate, _ := time.Parse(time.RFC3339, payment.DateApproved)

		updateReq := models.UpdatePaymentReq{
			TransactionType: &media,
			AuthDate:        &authDate,
			State:           &state,
		}
		_ = s.paymentRepo.Update(ctx, dbPayment.ID, updateReq)

		// Actualizar installments si existen
		instFilter := map[string]interface{}{
			"passenger_id": dbPayment.PassengerId,
			"company_id":   dbPayment.CompanyId,
			"sale_id":      dbPayment.SaleId,
		}
		instResult, err := s.installmentsRepo.Get(ctx, instFilter, nil, nil)
		if err == nil && len(instResult) > 0 {
			if list, ok := instResult[0].(models.InstallmentListResponse); ok {
				for _, inst := range list.Items {
					if inst.Balance != 0 {
						newPaid := float32(inst.PaidAmount) + float32(payment.TransactionAmount)
						newBal := float32(inst.Amount) - newPaid
						status := "PARTIAL"
						if newBal <= 0 {
							status = "PAID"
							newBal = 0
						}
						instID := fmt.Sprint(inst.ID)
						s.installmentsRepo.Update(ctx, instID, models.UpdateInstallmentReq{
							PaidAmount: &newPaid,
							Balance:    &newBal,
							Status:     &status,
						})

						pID, _ := strconv.ParseInt(dbPayment.ID, 10, 64)
						iID, _ := strconv.ParseInt(instID, 10, 64)
						s.paymentInstallmentRepo.Create(ctx, models.CreatePaymentInstallmentReq{
							PaymentId:     pID,
							InstallmentId: iID,
							AppliedAmount: float32(payment.TransactionAmount),
						})
						break
					}
				}
			}
		}

	case "pending":
		message = "Pago pendiente: " + payment.StatusDetail
		state := "Pendiente"
		_ = s.paymentRepo.Update(ctx, dbPayment.ID, models.UpdatePaymentReq{State: &state})

	case "rejected":
		message = "Pago rechazado: " + payment.StatusDetail
		state := "Rechazado"
		_ = s.paymentRepo.Update(ctx, dbPayment.ID, models.UpdatePaymentReq{State: &state})

	default:
		message = "Estado desconocido: " + payment.Status
	}

	return models.VerifyMPagoResp{
		Status:            payment.Status,
		Message:           message,
		Subdominio:  	   company.Subdominio,         
		PaymentID:         payment.ID,
		TransactionAmount: payment.TransactionAmount,
		ExternalReference: payment.ExternalReference,
	}, nil
}
