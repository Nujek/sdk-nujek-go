package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
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

type MarkChatReadRequest struct {
	LastReadMessageID int64 `json:"last_read_message_id"`
}

type MarkChatReadResponse struct {
	LastReadMessageID int64 `json:"last_read_message_id"`
}

// SendChatImageRequest contains an image and an optional caption. ContentType
// must be image/jpeg, image/png, image/gif, or image/webp.
type SendChatImageRequest struct {
	FileName    string
	ContentType string
	Image       io.Reader
	Message     string
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

// MarkChatRead marks messages through LastReadMessageID as read by the
// customer. The server keeps the read position monotonic.
func (c *Client) MarkChatRead(ctx context.Context, orderUUID string, request MarkChatReadRequest) (Response[MarkChatReadResponse], error) {
	var response Response[MarkChatReadResponse]
	orderUUID = strings.TrimSpace(orderUUID)
	if orderUUID == "" {
		return response, errors.New("order UUID wajib diisi")
	}
	if request.LastReadMessageID <= 0 {
		return response, errors.New("last read message ID harus lebih dari nol")
	}

	path := "/orders/" + url.PathEscape(orderUUID) + "/chat/" + customerDriverConversation + "/read"
	err := c.postJSON(ctx, path, request, &response)
	return response, err
}

// SendChatImage uploads an image and sends it to the customer-driver chat in
// one signed multipart request.
func (c *Client) SendChatImage(ctx context.Context, orderUUID string, request SendChatImageRequest) (Response[ChatMessage], error) {
	var response Response[ChatMessage]
	orderUUID = strings.TrimSpace(orderUUID)
	if orderUUID == "" {
		return response, errors.New("order UUID wajib diisi")
	}
	request.FileName = strings.TrimSpace(request.FileName)
	if request.FileName == "" {
		return response, errors.New("nama file gambar wajib diisi")
	}
	if request.Image == nil {
		return response, errors.New("data gambar wajib diisi")
	}
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowed[request.ContentType] {
		return response, errors.New("content type gambar harus image/jpeg, image/png, image/gif, atau image/webp")
	}
	if utf8.RuneCountInString(request.Message) > 1000 {
		return response, errors.New("pesan chat maksimal 1000 karakter")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeMultipartValue(request.FileName)))
	fileHeader.Set("Content-Type", request.ContentType)
	part, err := writer.CreatePart(fileHeader)
	if err != nil {
		return response, fmt.Errorf("buat multipart gambar: %w", err)
	}
	if _, err := io.Copy(part, request.Image); err != nil {
		return response, fmt.Errorf("baca gambar: %w", err)
	}
	if request.Message != "" {
		if err := writer.WriteField("message", request.Message); err != nil {
			return response, fmt.Errorf("tulis caption gambar: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return response, fmt.Errorf("tutup multipart gambar: %w", err)
	}

	path := "/orders/" + url.PathEscape(orderUUID) + "/chat/" + customerDriverConversation + "/images"
	err = c.requestBytes(ctx, http.MethodPost, path, nil, body.Bytes(), writer.FormDataContentType(), &response)
	return response, err
}

func escapeMultipartValue(value string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, `\"`, "\r", "", "\n", "").Replace(value)
}
