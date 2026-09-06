// Example HTTP server untuk mengetes seluruh Partner API SDK melalui Postman.
package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Nujek/sdk-nujek-go/pkg/client"
)

type server struct{ api *client.Client }

func main() {
	if err := loadDotEnv(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("warning: .env tidak dimuat: %v", err)
	}
	api, err := client.New(os.Getenv("CLIENT_API_BASE_URL"), os.Getenv("CLIENT_API_KEY"), os.Getenv("CLIENT_API_SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("EXAMPLE_PORT")
	if port == "" {
		port = "8088"
	}
	h := &server{api: api}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /register", h.register)
	mux.HandleFunc("GET /pricing", h.pricingPreview)
	mux.HandleFunc("POST /routing", h.routingDistance)
	mux.HandleFunc("POST /orders", h.createOrder)
	mux.HandleFunc("POST /orders/{orderUUID}/cancel", h.cancelOrder)
	mux.HandleFunc("POST /orders/{orderUUID}/review-driver", h.reviewDriver)

	address := ":" + port
	log.Printf("Partner API SDK example listening on http://localhost:%s", port)
	log.Printf("Postman base URL: http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(address, logging(mux)))
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	result, err := s.api.Register(r.Context(), payload.Name, payload.Email, payload.Phone)
	writeResult(w, result, err)
}

func (s *server) pricingPreview(w http.ResponseWriter, r *http.Request) {
	result, message, err := s.api.PricingPreview(r.Context(), client.PricingPreviewParams(r.URL.Query()))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result, "message": message})
}

func (s *server) routingDistance(w http.ResponseWriter, r *http.Request) {
	var payload client.RoutingRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	result, err := s.api.RoutingDistance(r.Context(), payload)
	writeResult(w, result, err)
}

func (s *server) createOrder(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if !decodeJSON(w, r, &payload) {
		return
	}
	result, message, err := s.api.CreateOrder(r.Context(), payload)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result, "message": message})
}

func (s *server) cancelOrder(w http.ResponseWriter, r *http.Request) {
	var payload *client.CancelRequest
	if r.ContentLength != 0 {
		var request client.CancelRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		payload = &request
	}
	message, err := s.api.CancelOrder(r.Context(), r.PathValue("orderUUID"), payload)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": message})
}

func (s *server) reviewDriver(w http.ResponseWriter, r *http.Request) {
	var payload client.ReviewRequest
	if !decodeJSON(w, r, &payload) {
		return
	}
	result, message, err := s.api.ReviewDriver(r.Context(), r.PathValue("orderUUID"), payload)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result, "message": message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"code": "INVALID_JSON", "message": err.Error()}})
		return false
	}
	return true
}

func writeResult[T any](w http.ResponseWriter, result client.Response[T], err error) {
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		status = apiErr.StatusCode
		writeJSON(w, status, map[string]any{"error": apiErr})
		return
	}
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": "SDK_ERROR", "message": err.Error()}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func loadDotEnv(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "export "))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		if _, exists := os.LookupEnv(strings.TrimSpace(key)); !exists {
			_ = os.Setenv(strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), "\"'"))
		}
	}
	return nil
}
