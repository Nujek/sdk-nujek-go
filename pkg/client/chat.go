package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

const customerDriverConversation = "customer_driver"

// ChatMessage is a message in the customer-driver conversation of an order.
type ChatMessage struct {
	ID          int64   `json:"id"`
	SenderID    int64   `json:"sender_id"`
	SenderRole  string  `json:"sender_role"`
	Message     string  `json:"message"`
	MessageType string  `json:"message_type"`
	ImagePath   *string `json:"image_path"`
	CreatedAt   string  `json:"created_at"`
	IsRead      bool    `json:"is_read"`
}

type ChatMessagesPage struct {
	Items        []ChatMessage `json:"items"`
	TotalItems   uint64        `json:"total_items"`
	TotalPages   uint64        `json:"total_pages"`
	CurrentPage  uint64        `json:"current_page"`
	ItemsPerPage uint64        `json:"items_per_page"`
}

// ChatMessagesParams controls pagination. Zero values use the API defaults
// (page 1 and limit 50).
type ChatMessagesParams struct {
	Page  uint64
	Limit uint64
}

type SendChatMessageRequest struct {
	Message     string  `json:"message"`
	MessageType string  `json:"message_type,omitempty"`
	ImagePath   *string `json:"image_path,omitempty"`
}

// ListChatMessages returns newest messages first. Partner chat is limited to
// the customer-driver conversation of an order owned by the client.
func (c *Client) ListChatMessages(ctx context.Context, orderUUID string, params ChatMessagesParams) (Response[ChatMessagesPage], error) {
	var response Response[ChatMessagesPage]
	orderUUID = strings.TrimSpace(orderUUID)
	if orderUUID == "" {
		return response, errors.New("order UUID wajib diisi")
	}

	query := make(url.Values)
	if params.Page > 0 {
		query.Set("page", strconv.FormatUint(params.Page, 10))
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.FormatUint(params.Limit, 10))
	}
	path := "/orders/" + url.PathEscape(orderUUID) + "/chat/" + customerDriverConversation + "/messages"
	err := c.request(ctx, http.MethodGet, path, query, nil, &response)
	return response, err
}

// SendChatMessage sends a customer message to the driver of an order owned by
// the client. MessageType defaults to "text" when omitted.
func (c *Client) SendChatMessage(ctx context.Context, orderUUID string, request SendChatMessageRequest) (Response[ChatMessage], error) {
	var response Response[ChatMessage]
	orderUUID = strings.TrimSpace(orderUUID)
	if orderUUID == "" {
		return response, errors.New("order UUID wajib diisi")
	}
	if strings.TrimSpace(request.Message) == "" {
		return response, errors.New("pesan chat wajib diisi")
	}
	if utf8.RuneCountInString(request.Message) > 1000 {
		return response, errors.New("pesan chat maksimal 1000 karakter")
	}

	path := "/orders/" + url.PathEscape(orderUUID) + "/chat/" + customerDriverConversation + "/messages"
	err := c.postJSON(ctx, path, request, &response)
	return response, err
}
