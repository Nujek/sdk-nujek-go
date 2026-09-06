// Package client provides a small, dependency-free Go client for the partner
// API exposed under /api/client.
package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Option func(*Client)

func WithHTTPClient(httpClient HTTPDoer) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.http = httpClient
		}
	}
}
func WithClock(clock func() time.Time) Option {
	return func(c *Client) {
		if clock != nil {
			c.clock = clock
		}
	}
}
func WithNonceGenerator(generator func() (string, error)) Option {
	return func(c *Client) {
		if generator != nil {
			c.nonce = generator
		}
	}
}

type Client struct {
	baseURL   string
	apiKey    string
	apiSecret []byte
	http      HTTPDoer
	clock     func() time.Time
	nonce     func() (string, error)
}

const apiPrefix = "/api/client"

func New(baseURL, apiKey, apiSecret string, options ...Option) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(apiKey) == "" || apiSecret == "" {
		return nil, errors.New("client API base URL, key, dan secret wajib diisi")
	}
	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("base URL tidak valid")
	}
	c := &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, apiSecret: []byte(apiSecret), http: http.DefaultClient, clock: time.Now, nonce: randomNonce}
	for _, option := range options {
		option(c)
	}
	return c, nil
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Fields     map[string]any
}

func (e *APIError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("client API HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("client API %s: %s", e.Code, e.Message)
}

type Response[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message,omitempty"`
}
type RegisterResponse struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
type RoutingRequest struct {
	Mode   string     `json:"mode"`
	Routes []Waypoint `json:"routes"`
}
type Waypoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type RoutingResponse struct {
	DistanceMeters  int64        `json:"distance_meters"`
	DurationSeconds int64        `json:"duration_seconds"`
	Provider        string       `json:"provider"`
	Mode            string       `json:"mode"`
	Legs            []RoutingLeg `json:"legs"`
}
type RoutingLeg struct {
	Index           int   `json:"index"`
	DistanceMeters  int64 `json:"distance_meters"`
	DurationSeconds int64 `json:"duration_seconds"`
}
type CancelRequest struct {
	Reason string `json:"reason,omitempty"`
}
type ReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment,omitempty"`
}

type PricingPreviewParams url.Values

func (c *Client) Register(ctx context.Context, name, email, phone string) (Response[RegisterResponse], error) {
	var response Response[RegisterResponse]
	err := c.postJSON(ctx, "/register", map[string]string{"name": name, "email": email, "phone": phone}, &response)
	return response, err
}

// PricingPreview returns the unmodified pricing response because its shape is
// shared with the internal pricing API and can evolve independently.
func (c *Client) PricingPreview(ctx context.Context, params PricingPreviewParams) (json.RawMessage, string, error) {
	var response Response[json.RawMessage]
	err := c.request(ctx, http.MethodGet, "/pricing/preview", url.Values(params), nil, &response)
	return response.Data, response.Message, err
}

func (c *Client) RoutingDistance(ctx context.Context, request RoutingRequest) (Response[RoutingResponse], error) {
	var response Response[RoutingResponse]
	err := c.postJSON(ctx, "/routing/distance", request, &response)
	return response, err
}

// CreateOrder accepts any JSON object containing customer_uuid and the order
// fields expected by the server. This keeps the SDK compatible with new order
// fields without a release for every backend change.
func (c *Client) CreateOrder(ctx context.Context, payload any) (json.RawMessage, string, error) {
	var response Response[json.RawMessage]
	err := c.postJSON(ctx, "/orders", payload, &response)
	return response.Data, response.Message, err
}

func (c *Client) ListOrders(ctx context.Context, params PricingPreviewParams) (json.RawMessage, string, error) {
	var response Response[json.RawMessage]
	err := c.request(ctx, http.MethodGet, "/orders", url.Values(params), nil, &response)
	return response.Data, response.Message, err
}

func (c *Client) ShowOrder(ctx context.Context, orderUUID string) (json.RawMessage, string, error) {
	var response Response[json.RawMessage]
	err := c.request(ctx, http.MethodGet, "/orders/"+url.PathEscape(orderUUID), nil, nil, &response)
	return response.Data, response.Message, err
}

func (c *Client) CancelOrder(ctx context.Context, orderUUID string, request *CancelRequest) (string, error) {
	var response Response[json.RawMessage]
	err := c.postJSON(ctx, "/orders/"+url.PathEscape(orderUUID)+"/cancel", request, &response)
	return response.Message, err
}

func (c *Client) ReviewDriver(ctx context.Context, orderUUID string, request ReviewRequest) (json.RawMessage, string, error) {
	var response Response[json.RawMessage]
	err := c.postJSON(ctx, "/orders/"+url.PathEscape(orderUUID)+"/review-driver", request, &response)
	return response.Data, response.Message, err
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, target any) error {
	return c.request(ctx, http.MethodPost, path, nil, payload, target)
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values, payload any, target any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if path == "" || path[0] != '/' {
		return errors.New("path API harus diawali slash")
	}
	var body []byte
	var err error
	if payload != nil && method != http.MethodGet {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
	}
	endpointPath := apiPrefix + path
	u, err := url.Parse(c.baseURL + endpointPath)
	if err != nil {
		return err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	nonce, err := c.nonce()
	if err != nil {
		return fmt.Errorf("buat nonce: %w", err)
	}
	timestamp := strconv.FormatInt(c.clock().Unix(), 10)
	bodyHash := sha256.Sum256(body)
	pathAndQuery := u.EscapedPath()
	if u.RawQuery != "" {
		pathAndQuery += "?" + u.RawQuery
	}
	canonical := strings.Join([]string{timestamp, nonce, strings.ToUpper(method), pathAndQuery, hex.EncodeToString(bodyHash[:])}, "\n")
	mac := hmac.New(sha256.New, c.apiSecret)
	_, _ = mac.Write([]byte(canonical))
	signature := hex.EncodeToString(mac.Sum(nil))
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-Client-Key", c.apiKey)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Signature", signature)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request client API: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeAPIError(resp.StatusCode, raw)
	}
	if target == nil || len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode client API response: %w", err)
	}
	return nil
}

func decodeAPIError(status int, raw []byte) error {
	var envelope struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Fields  map[string]any `json:"fields"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &envelope)
	message := envelope.Error.Message
	if message == "" {
		message = strings.TrimSpace(string(raw))
		if message == "" {
			message = http.StatusText(status)
		}
	}
	return &APIError{StatusCode: status, Code: envelope.Error.Code, Message: message, Fields: envelope.Error.Fields}
}

func randomNonce() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
