package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"tourmanager/config"
	"tourmanager/core/models"
	"tourmanager/core/ports"
	"tourmanager/util"

	"github.com/gin-gonic/gin"
)

// SetMpagoRoutes creates MercadoPago payment routes
func SetMpagoRoutes(ctx context.Context, cfg config.Config, r *gin.Engine, p ports.MPagoService) {
	r.POST("/api/v3.5/iniciopagomp", initMpagoPayment(ctx, cfg, p))
	r.POST("/api/v3.5/mpago/webhook", mpagoWebhook(ctx, cfg, p))
	r.GET("/api/v3.5/mpago/verificar", mpagoRedirect(ctx, cfg, p))
	r.POST("/api/v3.5/mpago/verificar", mpagoWebhook(ctx, cfg, p))
}

// @Summary Init Mercado Pago Payment
// @Description Inicializa el pago a través de Mercado Pago y devuelve los datos de preferencia y clave pública
// @Tags mpago
// @Param request body models.InitMPagoPaymentReq true "Configuración inicial del pago"
// @Success 200 {object} models.InitMPagoResp "OK"
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /api/v3.5/iniciopagomp [post]
func initMpagoPayment(ctx context.Context, cfg config.Config, p ports.MPagoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response := util.NewErrorResponse(err, http.StatusInternalServerError)
			c.JSON(response.StatusCode, response)
			return
		}

		var req models.InitMPagoPaymentReq
		if err := json.Unmarshal(body, &req); err != nil {
			response := util.NewErrorResponse(err, http.StatusBadRequest)
			c.JSON(response.StatusCode, response)
			return
		}

		resp, err := p.InitPayment(ctx, req)
		if err != nil {
			response := util.NewErrorResponse(err, http.StatusInternalServerError)
			c.JSON(response.StatusCode, response)
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Webhook Mercado Pago
// @Description Procesa las notificaciones webhook enviadas por Mercado Pago tras eventos de pago
// @Tags mpago
// @Accept json
// @Produce json
// @Param request body models.MPagoWebhookReq true "Datos de la notificación webhook"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /api/v3.5/mpago/webhook [post]
func mpagoWebhook(ctx context.Context, cfg config.Config, p ports.MPagoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error al leer el cuerpo de la solicitud"})
			return
		}

		var webhook models.MPagoWebhookReq
		if err := json.Unmarshal(body, &webhook); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
			return
		}

		// Validar que sea una notificación de pago
		paymentID := webhook.Data.ID
		if webhook.Type != "payment" || paymentID == "" {
			// Si es otro evento, respondemos OK para que Mercado Pago no reintente
			c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "tipo de notificación no soportado"})
			return
		}

		resp, err := p.VerifyPayment(c.Request.Context(), models.VerifyMPagoReq{
			PaymentID: paymentID,
			Source:    "webhook",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": resp.Message,
		})
	}
}

// @Summary Retorno y verificación Mercado Pago
// @Description Recibe el retorno del usuario desde Mercado Pago, verifica el pago y redirige a la pantalla de resultado
// @Tags mpago
// @Param payment_id query string false "ID del pago en Mercado Pago"
// @Param preference_id query string false "ID de la preferencia en Mercado Pago"
// @Success 302 {string} string "Redirección a /mpagopagos/resultado"
// @Failure 400 {object} object
// @Router /api/v3.5/mpago/verificar [get]
func mpagoRedirect(ctx context.Context, cfg config.Config, p ports.MPagoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentID := c.Query("payment_id")
		preferenceID := c.Query("preference_id")

		if paymentID == "" && preferenceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros requeridos no encontrados (payment_id o preference_id)"})
			return
		}

		resp, err := p.VerifyPayment(c.Request.Context(), models.VerifyMPagoReq{
			PaymentID:    paymentID,
			PreferenceID: preferenceID,
			Source:       "redirect",
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Redirigir a la vista de resultado en el frontend
		redirectURL := fmt.Sprintf("/mpagopagos/resultado?payment_id=%d&status=%s&external_reference=%s",
			resp.PaymentID, url.QueryEscape(resp.Status), url.QueryEscape(resp.ExternalReference))
		c.Redirect(http.StatusFound, redirectURL)
	}
}
