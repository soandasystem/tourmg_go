package ports

import (
	"context"
	"tourmanager/core/models"
)

// FlightgdsClient define la comunicación externa con la API de Defontana
type FlightgdsClient interface {
	FetchFlightsGDS(ctx context.Context, key string) ([]models.FlightRawOrder, error)
}

// FlightgdsService define el caso de uso del dominio para la integración de Defontana
type FlightgdsService interface {
	GetFlight(ctx context.Context) ([]models.FlightOfferNorm, error)
}
