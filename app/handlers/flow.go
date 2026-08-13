package handlers

import (
	"bytes"
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

// SetFlowRoutes creates flow payment routes
func SetFlowRoutes(ctx context.Context, cfg config.Config, r *gin.Engine, p ports.FlowService) {
	r.POST("/api/v3.5/iniciopagoflow", initFlowPayment(ctx, cfg, p))
	r.POST("/api/v3.5/token", tokenflow(ctx, cfg, p))
	r.POST("/api/v3.5/consulta-token", consultatoken(ctx, cfg, p))
	r.POST("/api/v3.5/returnflow", returnflow(ctx, cfg, p))
}

// @Summary Init Flow Payment
// @Description Inicializa el pago a través de Flow y devuelve la URL de redirección
// @Tags flow
// @Param request body models.InitFlowPaymentReq true "Configuración inicial del pago"
// @Success 200 {object} models.InitFlowPaymentResp "OK"
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /api/v3.5/gateways/flow/init [post]
func initFlowPayment(ctx context.Context, cfg config.Config, p ports.FlowService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response := util.NewErrorResponse(err, http.StatusInternalServerError)
			c.JSON(response.StatusCode, response)
			return
		}

		var req models.InitFlowPaymentReq
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

// @Summary Init Flow Payment
// @Description Inicializa el pago a través de Flow y devuelve la URL de redirección
// @Tags flow
// @Param request body models.InitFlowPaymentReq true "Configuración inicial del pago"
// @Success 200 {object} models.InitFlowPaymentResp "OK"
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /api/v3.5/gateways/flow/init [post]
func tokenflow(ctx context.Context, cfg config.Config, p ports.FlowService) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method != http.MethodPost {
			response := util.NewErrorResponse(fmt.Errorf("Método no permitido"), http.StatusMethodNotAllowed)
			c.JSON(response.StatusCode, response)
			return
		}

		var tokenRequest models.TokenRequest
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		if err := json.Unmarshal(body, &tokenRequest); err != nil || tokenRequest.Token == "" {
			// Si falla, probamos leerlo desde el formulario (x-www-form-urlencoded)
			token := c.PostForm("token")
			if token == "" {
				response := util.NewErrorResponse(fmt.Errorf("Token no encontrado en JSON ni en formulario"), http.StatusBadRequest)
				c.JSON(response.StatusCode, response)
				return
			}
			tokenRequest.Token = token
		}

		// Aquí puedes llamar a tu función principal con el token recibido
		// Simulamos que el token es válido y devolvemos una respuesta JSON
		response, err := p.FlowToken(ctx, tokenRequest.Token)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, response)

	}
}

func consultatoken(ctx context.Context, cfg config.Config, p ports.FlowService) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method != http.MethodPost {
			response := util.NewErrorResponse(fmt.Errorf("Método no permitido"), http.StatusMethodNotAllowed)
			c.JSON(response.StatusCode, response)
			return
		}

		var tokenRequest models.TokenRequest
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		if err := json.Unmarshal(body, &tokenRequest); err != nil || tokenRequest.Token == "" {
			// Si falla, probamos leerlo desde el formulario (x-www-form-urlencoded)
			token := c.PostForm("token")
			if token == "" {
				response := util.NewErrorResponse(fmt.Errorf("Token no encontrado en JSON ni en formulario"), http.StatusBadRequest)
				c.JSON(response.StatusCode, response)
				return
			}
			tokenRequest.Token = token
		}

		// Aquí puedes llamar a tu función principal con el token recibido
		// Simulamos que el token es válido y devolvemos una respuesta JSON
		response, err := p.ConsultaToken(ctx, tokenRequest.Token)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, response)

	}
}

func returnflow(ctx context.Context, cfg config.Config, p ports.FlowService) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method != http.MethodPost {
			response := util.NewErrorResponse(fmt.Errorf("Método no permitido"), http.StatusMethodNotAllowed)
			c.JSON(response.StatusCode, response)
			return
		}

		var tokenRequest models.TokenRequest
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		if err := json.Unmarshal(body, &tokenRequest); err != nil || tokenRequest.Token == "" {
			// Si falla, probamos leerlo desde el formulario (x-www-form-urlencoded)
			token := c.PostForm("token")
			if token == "" {
				response := util.NewErrorResponse(fmt.Errorf("Token no encontrado en JSON ni en formulario"), http.StatusBadRequest)
				c.JSON(response.StatusCode, response)
				return
			}
			tokenRequest.Token = token
		}

		// Aquí puedes llamar a tu función principal con el token recibido
		_, err := p.ReturnFlow(ctx, tokenRequest.Token)
		if err != nil {
			fmt.Println(err)
		}

		// Realizar la redirección HTTP aquí en el handler
		redirectURL := "/flowpagos/resultado?commerceOrder=" + url.QueryEscape(tokenRequest.Token)
		c.Redirect(http.StatusFound, redirectURL)

	}
}
