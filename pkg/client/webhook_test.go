package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func signWebhookForTest(secret, timestamp, deliveryID string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + deliveryID + "\n" + string(body)))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookAtChecksSignatureAndFreshness(t *testing.T) {
	const secret = "webhook-secret"
	const timestamp = "1700000000"
	const deliveryID = "3e1a6e70-3602-4a57-a092-078b2d8f22a1"
	body := []byte(`{"event":"order.created","occurred_at":"2023-11-14T22:13:20Z","data":{"order_uuid":"order-uuid","status":"PENDING","driver_uuid":null}}`)
	signature := signWebhookForTest(secret, timestamp, deliveryID, body)

	if err := VerifyWebhookAt(secret, timestamp, deliveryID, body, signature, time.Unix(1700000120, 0), 5*time.Minute); err != nil {
		t.Fatalf("valid webhook rejected: %v", err)
	}
	if err := VerifyWebhookAt(secret, timestamp, deliveryID, body, signature, time.Unix(1700000600, 0), 5*time.Minute); err == nil {
		t.Fatal("expected stale webhook to be rejected")
	}
	if err := VerifyWebhookAt(secret, timestamp, deliveryID, []byte("tampered"), signature, time.Unix(1700000120, 0), 5*time.Minute); err == nil {
		t.Fatal("expected modified body to be rejected")
	}
}

func TestParseAndDecodeKnownWebhook(t *testing.T) {
	body := []byte(`{"event":"order.created","occurred_at":"2026-09-10T14:37:06Z","data":{"order_uuid":"order-uuid","status":"BOOKING","driver_uuid":null,"booking_at":"2026-09-11T01:00:00Z"}}`)
	event, err := ParseWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if !IsKnownWebhookEvent(event.Event) {
		t.Fatalf("expected %q to be known", event.Event)
	}
	data, err := DecodeWebhookData[OrderWebhookData](event)
	if err != nil {
		t.Fatal(err)
	}
	if data.OrderUUID != "order-uuid" || data.Status != "BOOKING" || data.DriverUUID != nil || data.BookingAt == nil {
		t.Fatalf("unexpected payload: %+v", data)
	}
}

func TestParseWebhookAllowsFutureEvents(t *testing.T) {
	event, err := ParseWebhook([]byte(`{"event":"order.future_event","occurred_at":"2026-09-10T14:37:06Z","data":{"value":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	if IsKnownWebhookEvent(event.Event) {
		t.Fatal("future event should not be reported as known")
	}
}
