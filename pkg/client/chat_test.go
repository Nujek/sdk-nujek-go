package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
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
	transport := &chatCaptureTransport{responseBody: `{"data":{"items":[{"id":50,"sender_id":7,"sender_role":"driver","message":"otw","message_type":"text","image_path":null,"image_url":null,"created_at":"2026-09-10 14:37:06 +00:00:00","is_read":false}],"total_items":1,"total_pages":1,"current_page":2,"items_per_page":25},"message":"Chat berhasil diambil"}`}
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
	transport := &chatCaptureTransport{responseBody: `{"data":{"id":51,"sender_id":8,"sender_role":"customer","message":"Driver, mohon ke pickup","message_type":"text","image_path":null,"image_url":null,"created_at":"2026-09-10 14:38:06 +00:00:00","is_read":false}}`}
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

func TestMarkChatRead(t *testing.T) {
	transport := &chatCaptureTransport{responseBody: `{"data":{"last_read_message_id":51},"message":"Chat berhasil ditandai telah dibaca"}`}
	c := newChatTestClient(t, transport)

	response, err := c.MarkChatRead(context.Background(), "order-uuid", MarkChatReadRequest{LastReadMessageID: 51})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := transport.request.URL.Path, "/api/client/orders/order-uuid/chat/customer_driver/read"; got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
	if transport.request.Method != http.MethodPost || response.Data.LastReadMessageID != 51 {
		t.Fatalf("unexpected method/response: %s / %+v", transport.request.Method, response.Data)
	}
	var request MarkChatReadRequest
	if err := json.NewDecoder(transport.request.Body).Decode(&request); err != nil {
		t.Fatal(err)
	}
	if request.LastReadMessageID != 51 {
		t.Fatalf("last read message ID = %d, want 51", request.LastReadMessageID)
	}
}

func TestMarkChatReadRejectsInvalidInput(t *testing.T) {
	c := newChatTestClient(t, &chatCaptureTransport{})
	if _, err := c.MarkChatRead(context.Background(), "order-uuid", MarkChatReadRequest{}); err == nil {
		t.Fatal("expected invalid message ID error")
	}
}

func TestSendChatImage(t *testing.T) {
	transport := &chatCaptureTransport{responseBody: `{"data":{"id":52,"sender_id":8,"sender_role":"customer","message":"Lokasi saya","message_type":"image","image_path":"image.jpg","image_url":"https://api.example.com/api/files/image.jpg","created_at":"2026-09-10 14:39:06 +00:00:00","is_read":false},"message":"Gambar chat berhasil dikirim"}`}
	c := newChatTestClient(t, transport)

	response, err := c.SendChatImage(context.Background(), "order-uuid", SendChatImageRequest{
		FileName:    "pickup.jpg",
		ContentType: "image/jpeg",
		Image:       strings.NewReader("jpeg-bytes"),
		Message:     "Lokasi saya",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := transport.request.URL.Path, "/api/client/orders/order-uuid/chat/customer_driver/images"; got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
	if transport.request.Method != http.MethodPost || response.Data.MessageType != "image" {
		t.Fatalf("unexpected method/response: %s / %+v", transport.request.Method, response.Data)
	}
	if response.Data.ImageURL == nil || *response.Data.ImageURL != "https://api.example.com/api/files/image.jpg" {
		t.Fatalf("image URL = %v", response.Data.ImageURL)
	}
	reader, err := multipart.NewReader(transport.request.Body, strings.TrimPrefix(transport.request.Header.Get("Content-Type"), "multipart/form-data; boundary=")).ReadForm(1024)
	if err != nil {
		t.Fatal(err)
	}
	if reader.Value["message"][0] != "Lokasi saya" || reader.File["file"][0].Filename != "pickup.jpg" {
		t.Fatalf("unexpected multipart form: %+v", reader)
	}
}

func TestSendChatImageRejectsInvalidInput(t *testing.T) {
	c := newChatTestClient(t, &chatCaptureTransport{})
	if _, err := c.SendChatImage(context.Background(), "order-uuid", SendChatImageRequest{}); err == nil {
		t.Fatal("expected invalid image request error")
	}
	if _, err := c.SendChatImage(context.Background(), "order-uuid", SendChatImageRequest{
		FileName: "chat.pdf", ContentType: "application/pdf", Image: strings.NewReader("pdf"),
	}); err == nil {
		t.Fatal("expected invalid content type error")
	}
}

func TestChatWebhookVerificationAndParsing(t *testing.T) {
	body := []byte(`{"event":"chat.message","occurred_at":"2026-09-10T14:37:06Z","data":{"order_uuid":"order-uuid","conversation_type":"customer_driver","message_id":50,"sender_id":7,"sender_role":"driver","message":"","message_type":"image","image_path":"image.jpg","image_url":"https://api.example.com/api/files/image.jpg","created_at":"2026-09-10 14:37:06 +00:00:00"}}`)
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
	if event.Data.ImageURL == nil || *event.Data.ImageURL != "https://api.example.com/api/files/image.jpg" {
		t.Fatalf("webhook image URL = %v", event.Data.ImageURL)
	}
}
