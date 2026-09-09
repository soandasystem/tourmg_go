package ports

import (
	"context"
	"tourmanager/core/models"
)

type MPagoService interface {
	InitPayment(ctx context.Context, req models.InitMPagoPaymentReq) (models.InitMPagoResp, error)
	VerifyPayment(ctx context.Context, req models.VerifyMPagoReq) (models.VerifyMPagoResp, error)
}
