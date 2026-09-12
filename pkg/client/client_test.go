package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDecodeAPIErrorSupportsBackendEnvelopes(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		message    string
		fieldCount int
	}{
		{name: "message response", body: `{"message":"Resource tidak ditemukan"}`, message: "Resource tidak ditemukan"},
		{name: "validation response", body: `{"status":"error","message":"Validation failed","errors":{"email":["invalid email"]}}`, message: "Validation failed", fieldCount: 1},
		{name: "nested proxy response", body: `{"error":{"code":"INVALID_SIGNATURE","message":"Signature salah","fields":{"signature":"invalid"}}}`, message: "Signature salah", fieldCount: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := decodeAPIError(http.StatusBadRequest, []byte(test.body))
			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("error type = %T, want *APIError", err)
			}
			if apiErr.Message != test.message || len(apiErr.Fields) != test.fieldCount {
				t.Fatalf("unexpected API error: %+v", apiErr)
			}
		})
	}
}

type captureTransport struct{ request *http.Request }

func (t *captureTransport) Do(request *http.Request) (*http.Response, error) {
	t.request = request
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"data":{"ok":true}}`)),
		Header:     make(http.Header),
	}, nil
}

func TestRequestSignatureIncludesQueryAndBody(t *testing.T) {
	transport := &captureTransport{}
	c, err := New("https://example.test", "public-key", "secret",
		WithHTTPClient(transport),
		WithClock(func() time.Time { return time.Unix(1700000000, 0) }),
		WithNonceGenerator(func() (string, error) { return "nonce-123456789012", nil }),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.PricingPreview(context.Background(), PricingPreviewParams{"service_id": {"1"}, "regency_id": {"7171"}}); err != nil {
		t.Fatal(err)
	}

	request := transport.request
	bodyHash := sha256.Sum256(nil)
	canonical := "1700000000\nnonce-123456789012\nGET\n/api/client/pricing/preview?regency_id=7171&service_id=1\n" + hex.EncodeToString(bodyHash[:])
	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write([]byte(canonical))
	if got, want := request.Header.Get("X-Signature"), hex.EncodeToString(mac.Sum(nil)); got != want {
		t.Fatalf("signature = %s, want %s", got, want)
	}
}
