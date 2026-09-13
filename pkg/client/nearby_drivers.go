package client

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

// NearbyDriversRequest filters available drivers by location and sub-service.
type NearbyDriversRequest struct {
	Latitude     float64
	Longitude    float64
	SubServiceID int
	RadiusKM     float64 // Zero uses the server default radius of 5 km.
}

// NearbyDriver is an online, active driver eligible for the requested
// sub-service and within the requested radius.
type NearbyDriver struct {
	UUID               string  `json:"uuid"`
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	Status             string  `json:"status"`
	ServiceID          *int    `json:"service_id"`
	ServiceName        *string `json:"service_name"`
	SubServiceID       *int    `json:"sub_service_id"`
	SubServiceName     *string `json:"sub_service_name"`
	Icon               *string `json:"icon"`
	DistanceMeters     float64 `json:"distance_meters"`
	UpdatedAt          string  `json:"updated_at"`
	LastSeenMinutesAgo float64 `json:"last_seen_minutes_ago"`
	VehicleBrand       *string `json:"vehicle_brand"`
	VehicleModel       *string `json:"vehicle_model"`
	VehicleYear        *int    `json:"vehicle_year"`
	FullName           *string `json:"full_name"`
	Gender             *string `json:"gender"`
	PlateNumber        *string `json:"plate_number"`
	ImageURL           *string `json:"image_url"`
	WorkAreaID         *string `json:"work_area_id"`
	WorkAreaName       *string `json:"work_area_name"`
}

// NearbyDrivers returns available drivers eligible for SubServiceID.
func (c *Client) NearbyDrivers(ctx context.Context, request NearbyDriversRequest) (Response[[]NearbyDriver], error) {
	var response Response[[]NearbyDriver]
	if math.IsNaN(request.Latitude) || math.IsInf(request.Latitude, 0) || request.Latitude < -90 || request.Latitude > 90 ||
		math.IsNaN(request.Longitude) || math.IsInf(request.Longitude, 0) || request.Longitude < -180 || request.Longitude > 180 {
		return response, errors.New("koordinat nearby driver tidak valid")
	}
	if request.SubServiceID <= 0 {
		return response, errors.New("sub-service ID wajib lebih dari nol")
	}
	if math.IsNaN(request.RadiusKM) || math.IsInf(request.RadiusKM, 0) || request.RadiusKM < 0 || request.RadiusKM > 100 {
		return response, errors.New("radius nearby driver harus antara 0 sampai 100 km")
	}

	query := url.Values{}
	query.Set("latitude", strconv.FormatFloat(request.Latitude, 'f', -1, 64))
	query.Set("longitude", strconv.FormatFloat(request.Longitude, 'f', -1, 64))
	query.Set("sub_service_id", strconv.Itoa(request.SubServiceID))
	if request.RadiusKM > 0 {
		query.Set("radius_km", strconv.FormatFloat(request.RadiusKM, 'f', -1, 64))
	}
	err := c.request(ctx, http.MethodGet, "/nearby-drivers", query, nil, &response)
	return response, err
}
