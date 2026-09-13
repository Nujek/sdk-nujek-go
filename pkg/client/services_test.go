package client

import (
	"context"
	"math"
	"net/http"
	"testing"
)

func TestServices(t *testing.T) {
	transport := &chatCaptureTransport{responseBody: `{"data":{"latitude":1.4748,"longitude":124.8421,"city":{"id":"7171","name":"Kota Manado","distance_km":1.2},"services":[{"id":1,"name":"Ride","is_available":true,"is_available_register":true,"max_radius":25,"screen":"AppScreen","tariff":{"id":10,"first_km":1,"first_km_price":10000,"next_km_price":3000,"driver_radius_km":5,"commission_percent":10,"commission_flat":0,"max_distance_km":40,"incentive_fee":0,"is_available":true,"scope":"default"},"sub_services":[{"id":1,"service_id":1,"name":"Ride","price_percentage":100,"price_flat":0,"is_available":true,"sort_order":1,"nearby_drivers_count":3}]}]}}`}
	c := newChatTestClient(t, transport)
	response, err := c.Services(context.Background(), ServicesRequest{Latitude: 1.4748, Longitude: 124.8421})
	if err != nil {
		t.Fatal(err)
	}
	if transport.request.Method != http.MethodGet || transport.request.URL.Path != "/api/client/services" {
		t.Fatalf("unexpected request: %s %s", transport.request.Method, transport.request.URL.Path)
	}
	if response.Data.City.ID != "7171" || response.Data.City.Name != "Kota Manado" ||
		len(response.Data.Services) != 1 || response.Data.Services[0].Tariff.Scope != "default" ||
		len(response.Data.Services[0].SubServices) != 1 || response.Data.Services[0].SubServices[0].NearbyDriversCount != 3 {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
	if got := transport.request.URL.Query(); got.Get("latitude") != "1.4748" || got.Get("longitude") != "124.8421" {
		t.Fatalf("unexpected query: %v", got)
	}
}

func TestServicesFiltersByServiceID(t *testing.T) {
	transport := &chatCaptureTransport{responseBody: `{"data":{"services":[{"id":1}]}}`}
	c := newChatTestClient(t, transport)
	serviceID := 1
	if _, err := c.Services(context.Background(), ServicesRequest{Latitude: 1, Longitude: 124, ServiceID: &serviceID}); err != nil {
		t.Fatal(err)
	}
	if got := transport.request.URL.Query().Get("service_id"); got != "1" {
		t.Fatalf("service_id query = %q", got)
	}
}

func TestServicesRejectsInvalidCoordinates(t *testing.T) {
	c := newChatTestClient(t, &chatCaptureTransport{})
	if _, err := c.Services(context.Background(), ServicesRequest{Latitude: math.NaN(), Longitude: 124}); err == nil {
		t.Fatal("expected invalid coordinate error")
	}
}
