package client

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"testing"
)

func TestReverseGeocode(t *testing.T) {
	transport := &chatCaptureTransport{responseBody: `{"data":{"formatted":"Jalan Sam Ratulangi, Manado, Sulawesi Utara, Indonesia","address_line1":"Jalan Sam Ratulangi","address_line2":"Manado, Sulawesi Utara, Indonesia","street":"Jalan Sam Ratulangi","house_number":null,"city":"Manado","district":"Wenang","state":"Sulawesi Utara","postcode":"95111","country":"Indonesia","country_code":"id","latitude":1.4748,"longitude":124.8421,"result_type":"street","provider":"geoapify"},"message":"Alamat berhasil ditemukan"}`}
	c := newChatTestClient(t, transport)

	response, err := c.ReverseGeocode(context.Background(), ReverseGeocodeRequest{
		Latitude: 1.4748, Longitude: 124.8421,
	})
	if err != nil {
		t.Fatal(err)
	}
	if transport.request.Method != http.MethodPost || transport.request.URL.Path != "/api/client/geocoding/reverse" {
		t.Fatalf("unexpected request: %s %s", transport.request.Method, transport.request.URL.Path)
	}
	var request ReverseGeocodeRequest
	if err := json.NewDecoder(transport.request.Body).Decode(&request); err != nil {
		t.Fatal(err)
	}
	if request.Latitude != 1.4748 || response.Data.Formatted == "" || response.Data.Provider != "geoapify" {
		t.Fatalf("unexpected request/response: %+v / %+v", request, response.Data)
	}
}

func TestReverseGeocodeRejectsInvalidCoordinates(t *testing.T) {
	c := newChatTestClient(t, &chatCaptureTransport{})
	for _, request := range []ReverseGeocodeRequest{
		{Latitude: 91, Longitude: 124},
		{Latitude: 1, Longitude: 181},
		{Latitude: math.NaN(), Longitude: 124},
	} {
		if _, err := c.ReverseGeocode(context.Background(), request); err == nil {
			t.Fatalf("expected invalid coordinate error for %+v", request)
		}
	}
}
