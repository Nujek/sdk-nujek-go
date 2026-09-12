package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type chatCaptureTransport struct {
	request      *http.Request
	responseBody string
}

func (t *chatCaptureTransport) Do(request *http.Request) (*http.Response, error) {
	t.request = request
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(t.responseBody)),
		Header:     make(http.Header),
	}, nil
}

func newChatTestClient(t *testing.T, transport HTTPDoer) *Client {
	t.Helper()
	c, err := New("https://example.test", "public-key", "secret", WithHTTPClient(transport))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestListChatMessages(t *testing.T) {
	transport := &chatCaptureTransport{responseBody: `{"data":{"items":[{"id":50,"sender_id":7,"sender_role":"driver","message":"otw","message_type":"text","image_path":null,"created_at":"2026-09-10 14:37:06 +00:00:00","is_read":false}],"total_items":1,"total_pages":1,"current_page":2,"items_per_page":25},"message":"Chat berhasil diambil"}`}
	c := newChatTestClient(t, transport)

	response, err := c.ListChatMessages(context.Background(), "order-uuid", ChatMessagesParams{Page: 2, Limit: 25})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := transport.request.URL.RequestURI(), "/api/client/orders/order-uuid/chat/customer_driver/messages?limit=25&page=2"; got != want {
		t.Fatalf("request URI = %q, want %q", got, want)
	}
	if transport.request.Method != http.MethodGet {
		t.Fatalf("method = %s, want GET", transport.request.Method)
	}
	if len(response.Data.Items) != 1 || response.Data.Items[0].SenderRole != "driver" {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
}

func TestSendChatMessage(t *testing.T) {
	transport := &chatCaptureTransport{responseBody: `{"data":{"id":51,"sender_id":8,"sender_role":"customer","message":"Driver, mohon ke pickup","message_type":"text","image_path":null,"created_at":"2026-09-10 14:38:06 +00:00:00","is_read":false}}`}
	c := newChatTestClient(t, transport)

	response, err := c.SendChatMessage(context.Background(), "order-uuid", SendChatMessageRequest{
		Message:     "Driver, mohon ke pickup",
		MessageType: "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := transport.request.URL.Path, "/api/client/orders/order-uuid/chat/customer_driver/messages"; got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
	if transport.request.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", transport.request.Method)
	}
	body, err := io.ReadAll(transport.request.Body)
	if err != nil {
		t.Fatal(err)
	}
	var request SendChatMessageRequest
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	if request.Message != "Driver, mohon ke pickup" || response.Data.ID != 51 {
		t.Fatalf("unexpected request/response: %+v / %+v", request, response.Data)
	}
}

func TestSendChatMessageRejectsInvalidInput(t *testing.T) {
	c := newChatTestClient(t, &chatCaptureTransport{})
	if _, err := c.SendChatMessage(context.Background(), "", SendChatMessageRequest{Message: "hello"}); err == nil {
		t.Fatal("expected empty order UUID error")
	}
	if _, err := c.SendChatMessage(context.Background(), "order-uuid", SendChatMessageRequest{}); err == nil {
		t.Fatal("expected empty message error")
	}
}

func TestChatWebhookVerificationAndParsing(t *testing.T) {
	body := []byte(`{"event":"chat.message","occurred_at":"2026-09-10T14:37:06Z","data":{"order_uuid":"order-uuid","conversation_type":"customer_driver","message_id":50,"sender_id":7,"sender_role":"driver","message":"otw","message_type":"text","image_path":null,"created_at":"2026-09-10 14:37:06 +00:00:00"}}`)
	const secret = "webhook-secret"
	const timestamp = "1789051026"
	const deliveryID = "3e1a6e70-3602-4a57-a092-078b2d8f22a1"
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + deliveryID + "\n" + string(body)))
	signature := hex.EncodeToString(mac.Sum(nil))

	if !VerifyWebhookSignature(secret, timestamp, deliveryID, body, signature) {
		t.Fatal("expected valid webhook signature")
	}
	if VerifyWebhookSignature(secret, timestamp, deliveryID, []byte("tampered"), signature) {
		t.Fatal("expected tampered body to fail verification")
	}
	event, err := ParseChatMessageWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if event.Data.MessageID != 50 || event.Data.SenderRole != "driver" {
		t.Fatalf("unexpected event: %+v", event)
	}
}
