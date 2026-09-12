package client

import (
	"context"
	"errors"
	"math"
)

type ReverseGeocodeRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ReverseGeocodeResponse struct {
	Formatted    string  `json:"formatted"`
	AddressLine1 *string `json:"address_line1"`
	AddressLine2 *string `json:"address_line2"`
	Street       *string `json:"street"`
	HouseNumber  *string `json:"house_number"`
	City         *string `json:"city"`
	District     *string `json:"district"`
	State        *string `json:"state"`
	Postcode     *string `json:"postcode"`
	Country      *string `json:"country"`
	CountryCode  *string `json:"country_code"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ResultType   *string `json:"result_type"`
	Provider     string  `json:"provider"`
}

// ReverseGeocode resolves coordinates into a display-ready and structured
// address. Order creation also applies this automatically to routes whose
// address is omitted or blank.
func (c *Client) ReverseGeocode(ctx context.Context, request ReverseGeocodeRequest) (Response[ReverseGeocodeResponse], error) {
	var response Response[ReverseGeocodeResponse]
	if math.IsNaN(request.Latitude) || math.IsInf(request.Latitude, 0) ||
		math.IsNaN(request.Longitude) || math.IsInf(request.Longitude, 0) ||
		request.Latitude < -90 || request.Latitude > 90 ||
		request.Longitude < -180 || request.Longitude > 180 {
		return response, errors.New("koordinat reverse geocoding tidak valid")
	}
	err := c.postJSON(ctx, "/geocoding/reverse", request, &response)
	return response, err
}
