package client

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

type ServicesRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ServicesResponse contains the city resolved from the coordinates and the
// service catalog with the tariff selected for that city.
type ServicesResponse struct {
	Latitude  float64              `json:"latitude"`
	Longitude float64              `json:"longitude"`
	City      ServiceCity          `json:"city"`
	Services  []ServiceWithPricing `json:"services"`
}

type ServiceCity struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	DistanceKM float64 `json:"distance_km"`
}

// ServiceWithPricing is one service and all of its sub-services.
type ServiceWithPricing struct {
	ID                  int             `json:"id"`
	Name                string          `json:"name"`
	Description         *string         `json:"description"`
	IsAvailable         bool            `json:"is_available"`
	IsAvailableRegister bool            `json:"is_available_register"`
	MaxRadius           float64         `json:"max_radius"`
	Screen              string          `json:"screen"`
	Tariff              *ServiceTariff  `json:"tariff"`
	SubServices         []SubServiceFee `json:"sub_services"`
}

// ServiceTariff contains the regional tariff, or the default tariff when the
// city has no tariff override for this service. Scope is "regional" or
// "default".
type ServiceTariff struct {
	ID                int     `json:"id"`
	FirstKM           float64 `json:"first_km"`
	FirstKMPrice      float64 `json:"first_km_price"`
	NextKMPrice       float64 `json:"next_km_price"`
	DriverRadiusKM    float64 `json:"driver_radius_km"`
	CommissionPercent float64 `json:"commission_percent"`
	CommissionFlat    float64 `json:"commission_flat"`
	MaxDistanceKM     float64 `json:"max_distance_km"`
	IncentiveFee      float64 `json:"incentive_fee"`
	IsAvailable       bool    `json:"is_available"`
	Scope             string  `json:"scope"`
}

// SubServiceFee contains the sub-service adjustment applied to the service
// tariff. PricePercentage is a multiplier in percent and PriceFlat is a fixed
// amount.
type SubServiceFee struct {
	ID                 int     `json:"id"`
	ServiceID          int     `json:"service_id"`
	Name               string  `json:"name"`
	Description        *string `json:"description"`
	Icon               *string `json:"icon"`
	PricePercentage    float64 `json:"price_percentage"`
	PriceFlat          float64 `json:"price_flat"`
	IsAvailable        bool    `json:"is_available"`
	SortOrder          int     `json:"sort_order"`
	NearbyDriversCount int64   `json:"nearby_drivers_count"`
}

// Services resolves coordinates to the nearest supported city and returns
// every service, its selected tariff, and its sub-services.
func (c *Client) Services(ctx context.Context, request ServicesRequest) (Response[ServicesResponse], error) {
	var response Response[ServicesResponse]
	if math.IsNaN(request.Latitude) || math.IsInf(request.Latitude, 0) || request.Latitude < -90 || request.Latitude > 90 ||
		math.IsNaN(request.Longitude) || math.IsInf(request.Longitude, 0) || request.Longitude < -180 || request.Longitude > 180 {
		return response, errors.New("koordinat layanan tidak valid")
	}
	query := url.Values{}
	query.Set("latitude", strconv.FormatFloat(request.Latitude, 'f', -1, 64))
	query.Set("longitude", strconv.FormatFloat(request.Longitude, 'f', -1, 64))
	err := c.request(ctx, http.MethodGet, "/services", query, nil, &response)
	return response, err
}
