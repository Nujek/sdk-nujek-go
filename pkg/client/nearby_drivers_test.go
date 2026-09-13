package client

import (
	"context"
	"math"
	"net/http"
	"testing"
)

func TestNearbyDrivers(t *testing.T) {
	transport := &chatCaptureTransport{responseBody: `{"data":[{"uuid":"driver-uuid","status":"ONLINE","sub_service_id":1,"distance_meters":250,"image_url":"https://staging-api.nujek.co.id/api/files/driver.jpg"}]}`}
	c := newChatTestClient(t, transport)
	response, err := c.NearbyDrivers(context.Background(), NearbyDriversRequest{
		Latitude: 1.4748, Longitude: 124.8421, SubServiceID: 1, RadiusKM: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if transport.request.Method != http.MethodGet || transport.request.URL.Path != "/api/client/nearby-drivers" {
		t.Fatalf("unexpected request: %s %s", transport.request.Method, transport.request.URL.Path)
	}
	query := transport.request.URL.Query()
	if query.Get("sub_service_id") != "1" || query.Get("radius_km") != "3" {
		t.Fatalf("unexpected query: %v", query)
	}
	if len(response.Data) != 1 || response.Data[0].UUID != "driver-uuid" ||
		response.Data[0].ImageURL == nil || *response.Data[0].ImageURL != "https://staging-api.nujek.co.id/api/files/driver.jpg" {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
}

func TestNearbyDriversRejectsInvalidRequest(t *testing.T) {
	c := newChatTestClient(t, &chatCaptureTransport{})
	for _, request := range []NearbyDriversRequest{
		{Latitude: math.NaN(), Longitude: 124, SubServiceID: 1},
		{Latitude: 1, Longitude: 124, SubServiceID: 0},
		{Latitude: 1, Longitude: 124, SubServiceID: 1, RadiusKM: 101},
	} {
		if _, err := c.NearbyDrivers(context.Background(), request); err == nil {
			t.Fatalf("expected validation error for %+v", request)
		}
	}
}
