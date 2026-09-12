package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

const ChatMessageEvent = "chat.message"

type WebhookEvent[T any] struct {
	Event      string `json:"event"`
	OccurredAt string `json:"occurred_at"`
	Data       T      `json:"data"`
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

// VerifyWebhookSignature verifies X-Webhook-Signature against the exact raw
// request body. timestamp and deliveryID are the values of X-Webhook-Timestamp
// and X-Webhook-Id.
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

func ParseChatMessageWebhook(body []byte) (WebhookEvent[ChatMessageWebhookData], error) {
	var event WebhookEvent[ChatMessageWebhookData]
	if err := json.Unmarshal(body, &event); err != nil {
		return event, err
	}
	if event.Event != ChatMessageEvent {
		return event, errors.New("event webhook bukan chat.message")
	}
	return event, nil
}
