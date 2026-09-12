package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	WebhookEventOrderCreated          = "order.created"
	WebhookEventDriverAccepted        = "driver.accepted"
	WebhookEventDriverRejected        = "driver.rejected"
	WebhookEventDriverCancelled       = "driver.cancelled"
	WebhookEventDriverArrived         = "driver.arrived"
	WebhookEventDriverPickedUp        = "driver.picked_up"
	WebhookEventDriverTimeout         = "driver.timeout"
	WebhookEventOrderFinished         = "order.finished"
	WebhookEventOrderFinishedByAdmin  = "order.finished_by_admin"
	WebhookEventOrderCancelledByAdmin = "order.cancelled_by_admin"
	WebhookEventOrderCancelledByUser  = "order.cancelled_by_user"
	WebhookEventOrderDriverChanged    = "order.driver_changed"
	WebhookEventOrderTimeout          = "order.timeout"
	WebhookEventChatMessage           = "chat.message"
	WebhookEventOrderSOSCreated       = "order.sos_created"
	WebhookEventUserClientRevoked     = "user_client_revoked"

	// Deprecated: use WebhookEventChatMessage.
	ChatMessageEvent = WebhookEventChatMessage

	DefaultWebhookTolerance = 5 * time.Minute
)

var knownWebhookEvents = map[string]struct{}{
	WebhookEventOrderCreated: {}, WebhookEventDriverAccepted: {},
	WebhookEventDriverRejected: {}, WebhookEventDriverCancelled: {},
	WebhookEventDriverArrived: {}, WebhookEventDriverPickedUp: {},
	WebhookEventDriverTimeout: {}, WebhookEventOrderFinished: {},
	WebhookEventOrderFinishedByAdmin: {}, WebhookEventOrderCancelledByAdmin: {},
	WebhookEventOrderCancelledByUser: {}, WebhookEventOrderDriverChanged: {},
	WebhookEventOrderTimeout: {}, WebhookEventChatMessage: {},
	WebhookEventOrderSOSCreated: {}, WebhookEventUserClientRevoked: {},
}

type WebhookEvent[T any] struct {
	Event      string `json:"event"`
	OccurredAt string `json:"occurred_at"`
	Data       T      `json:"data"`
}

type OrderWebhookData struct {
	OrderUUID  string  `json:"order_uuid"`
	Status     string  `json:"status"`
	DriverUUID *string `json:"driver_uuid"`
	BookingAt  *string `json:"booking_at"`
}

type ChatMessageWebhookData struct {
	OrderUUID        string  `json:"order_uuid"`
	ConversationType string  `json:"conversation_type"`
	MessageID        int64   `json:"message_id"`
	SenderID         int64   `json:"sender_id"`
	SenderRole       string  `json:"sender_role"`
	Message          string  `json:"message"`
	MessageType      string  `json:"message_type"`
	ImagePath        *string `json:"image_path"`
	CreatedAt        string  `json:"created_at"`
}

type OrderSOSWebhookData struct {
	SOSUUID      string   `json:"sos_uuid"`
	OrderUUID    string   `json:"order_uuid"`
	ReporterRole string   `json:"reporter_role"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Status       string   `json:"status"`
}

type UserClientRevokedWebhookData struct {
	UserUUID   string `json:"user_uuid"`
	ClientUUID string `json:"client_uuid"`
}

// VerifyWebhookSignature verifies X-Webhook-Signature against the exact raw
// request body. timestamp and deliveryID are the values of X-Webhook-Timestamp
// and X-Webhook-Id. It does not check timestamp freshness.
func VerifyWebhookSignature(webhookSecret, timestamp, deliveryID string, body []byte, signature string) bool {
	if webhookSecret == "" || timestamp == "" || deliveryID == "" || signature == "" {
		return false
	}
	want, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	canonical := timestamp + "\n" + deliveryID + "\n" + string(body)
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	_, _ = mac.Write([]byte(canonical))
	return hmac.Equal(mac.Sum(nil), want)
}

// VerifyWebhook verifies both signature and timestamp freshness. Store each
// deliveryID after accepting it to prevent duplicate processing during retries.
func VerifyWebhook(webhookSecret, timestamp, deliveryID string, body []byte, signature string) error {
	return VerifyWebhookAt(webhookSecret, timestamp, deliveryID, body, signature, time.Now(), DefaultWebhookTolerance)
}

// VerifyWebhookAt is the deterministic variant of VerifyWebhook for tests and
// applications that use a custom clock or tolerance.
func VerifyWebhookAt(webhookSecret, timestamp, deliveryID string, body []byte, signature string, now time.Time, tolerance time.Duration) error {
	if !VerifyWebhookSignature(webhookSecret, timestamp, deliveryID, body, signature) {
		return errors.New("signature webhook tidak valid")
	}
	if tolerance <= 0 {
		return errors.New("toleransi webhook harus lebih dari nol")
	}
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("timestamp webhook tidak valid")
	}
	difference := now.Sub(time.Unix(seconds, 0))
	if difference < 0 {
		difference = -difference
	}
	if difference > tolerance {
		return errors.New("timestamp webhook di luar toleransi")
	}
	return nil
}

// ParseWebhook parses any Nujek webhook without rejecting future event names.
func ParseWebhook(body []byte) (WebhookEvent[json.RawMessage], error) {
	var event WebhookEvent[json.RawMessage]
	if err := json.Unmarshal(body, &event); err != nil {
		return event, err
	}
	if event.Event == "" || event.OccurredAt == "" || len(event.Data) == 0 || string(event.Data) == "null" {
		return event, errors.New("envelope webhook tidak lengkap")
	}
	return event, nil
}

func IsKnownWebhookEvent(event string) bool {
	_, known := knownWebhookEvents[event]
	return known
}

func DecodeWebhookData[T any](event WebhookEvent[json.RawMessage]) (T, error) {
	var data T
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return data, fmt.Errorf("decode data webhook %s: %w", event.Event, err)
	}
	return data, nil
}

func ParseChatMessageWebhook(body []byte) (WebhookEvent[ChatMessageWebhookData], error) {
	var result WebhookEvent[ChatMessageWebhookData]
	event, err := ParseWebhook(body)
	if err != nil {
		return result, err
	}
	if event.Event != WebhookEventChatMessage {
		return result, errors.New("event webhook bukan chat.message")
	}
	data, err := DecodeWebhookData[ChatMessageWebhookData](event)
	if err != nil {
		return result, err
	}
	return WebhookEvent[ChatMessageWebhookData]{Event: event.Event, OccurredAt: event.OccurredAt, Data: data}, nil
}
