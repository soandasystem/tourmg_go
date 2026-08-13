package ports

import (
	"context"
	"tourmanager/core/models"
)

type FlowService interface {
	InitPayment(ctx context.Context, req models.InitFlowPaymentReq) (models.InitFlowPaymentResp, error)
	FlowToken(ctx context.Context, token string) (models.TokenResponse, error)
	ConsultaToken(ctx context.Context, token string) (*models.FlowListResponse, error)
	ReturnFlow(ctx context.Context, token string) (*models.FlowListResponse, error)
}
