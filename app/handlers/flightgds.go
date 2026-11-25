package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"tourmanager/config"
	"tourmanager/core/models"
	"tourmanager/core/ports"
	"tourmanager/util"

	"github.com/gin-gonic/gin"
)

// SetFlowRoutes creates flow payment routes
func SetFlightGDSRoutes(ctx context.Context, cfg config.Config, r *gin.Engine, p ports.FlightgdsService) {
	r.GET("/api/v3.5/flightgds", getFlightgds(p))

}

// @Summary getFlightgds
// @Description obtener las ordenes de defontana
// @Tags Flightgds
// @Success 200 {object} string "OK"
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /api/v1/getFlightgds/ [get]
func getFlightgds(p ports.FlightgdsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Leer el cuerpo de la solicitud
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response := util.NewErrorResponse(err, http.StatusInternalServerError)
			c.JSON(response.StatusCode, response)
			return
		}

		if len(bytes.TrimSpace(body)) == 0 {
			response := util.NewErrorResponse(err, http.StatusInternalServerError)
			c.JSON(response.StatusCode, response)
			return
		}

		// Parsear el JSON al modelo adecuado
		var flightsearch models.FlightSearch
		err = json.Unmarshal(body, &flightsearch)
		if err != nil {
			response := util.NewErrorResponse(err, http.StatusInternalServerError)
			c.JSON(response.StatusCode, response)
			return
		}

		resp, err := p.GetFlight(ctx, flightsearch)
		if err != nil {
			response := util.NewErrorResponse(err, http.StatusInternalServerError)
			c.JSON(response.StatusCode, response)
			return
		}

		response := util.NewSuccessResponse(resp, http.StatusOK)
		c.JSON(response.StatusCode, response)

	}
}
