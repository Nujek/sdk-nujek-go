// Example penggunaan seluruh endpoint partner API.
//
// Jalankan dengan:
//   CLIENT_API_BASE_URL=https://api.example.com \
//   CLIENT_API_KEY=... CLIENT_API_SECRET=... \
//   go run ./examples/partner_api
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Nujek/sdk-nujek-go/pkg/client"
)

func main() {
	ctx := context.Background()
	api, err := client.New(os.Getenv("CLIENT_API_BASE_URL"), os.Getenv("CLIENT_API_KEY"), os.Getenv("CLIENT_API_SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	// Register customer (idempotent untuk client yang sama).
	registered, err := api.Register(ctx, os.Getenv("CUSTOMER_NAME"), os.Getenv("CUSTOMER_EMAIL"), os.Getenv("CUSTOMER_PHONE"))
	if err != nil {
		log.Fatal(err)
	}
	printJSON("register", registered)

	pricing, message, err := api.PricingPreview(ctx, client.PricingPreviewParams{
		"service_id": {"1"},
		"origin_lat": {"-7.250445"}, "origin_long": {"112.768845"},
		"destination_lat": {"-7.260000"}, "destination_long": {"112.780000"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("pricing (%s): %s\n", message, pricing)

	route, err := api.RoutingDistance(ctx, client.RoutingRequest{
		Mode: "motorcycle",
		Routes: []client.Waypoint{
			{Latitude: -7.250445, Longitude: 112.768845},
			{Latitude: -7.260000, Longitude: 112.780000},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	printJSON("routing", route)

	// ORDER_JSON harus berisi field order sesuai kontrak backend. customer_uuid
	// ditambahkan oleh contoh ini agar tidak perlu menduplikasi struktur order.
	if raw := os.Getenv("ORDER_JSON"); raw != "" {
		var order map[string]any
		if err := json.Unmarshal([]byte(raw), &order); err != nil {
			log.Fatalf("ORDER_JSON tidak valid: %v", err)
		}
		order["customer_uuid"] = os.Getenv("CUSTOMER_UUID")
		created, message, err := api.CreateOrder(ctx, order)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("create order (%s): %s\n", message, created)
	}

	if orderUUID := os.Getenv("ORDER_UUID"); orderUUID != "" {
		message, err := api.CancelOrder(ctx, orderUUID, &client.CancelRequest{Reason: "Dibatalkan dari contoh SDK"})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("cancel order: %s\n", message)
	}

	if orderUUID := os.Getenv("REVIEW_ORDER_UUID"); orderUUID != "" {
		rating, _ := strconv.Atoi(os.Getenv("REVIEW_RATING"))
		if rating == 0 {
			rating = 5
		}
		review, message, err := api.ReviewDriver(ctx, orderUUID, client.ReviewRequest{Rating: rating, Comment: os.Getenv("REVIEW_COMMENT")})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("review driver (%s): %s\n", message, review)
	}
}

func printJSON(name string, value any) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %s\n", name, raw)
}
