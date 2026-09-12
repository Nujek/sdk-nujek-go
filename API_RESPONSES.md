# Contoh Response Partner API

Dokumen ini memuat contoh response untuk **seluruh method publik API** pada SDK
Go. Contoh menggunakan nilai ilustratif; UUID, nominal, jarak, waktu, dan pesan
dapat berbeda sesuai data dan bahasa request.

Method yang mengembalikan `json.RawMessage` tetap menerima JSON pada properti
`data`. Gunakan `json.Unmarshal` ke model aplikasi partner jika membutuhkan tipe
yang lebih ketat.

## 1. Register

```go
response, err := api.Register(ctx, "Budi Santoso", "budi@example.com", "081234567890")
```

```json
{
  "data": {
    "uuid": "24a55c29-5be8-4381-b915-3d658685a5a8",
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "phone": "081234567890"
  },
  "message": "Customer berhasil didaftarkan"
}
```

Tipe hasil: `client.Response[client.RegisterResponse]`.

## 2. PricingPreview

```go
data, message, err := api.PricingPreview(ctx, client.PricingPreviewParams{
    "service_id": {"1"},
    "sub_service_id": {"1"},
    "regency_id": {"7171"},
    "distance_km": {"5.5"},
})
```

```json
{
  "data": {
    "service_id": 1,
    "service_name": "Ride",
    "service_description": "Transportasi penumpang",
    "screen": "ride",
    "sub_service_id": 1,
    "regency_id": "7171",
    "regency_name": "Kota Manado",
    "distance_km": 5.5,
    "base_price": 17500,
    "service_percent": 100,
    "service_flat": 0,
    "sub_service_percent": 100,
    "sub_service_flat": 0,
    "sub_service_adjustment": 0,
    "commission_fee": 3500,
    "incentive_fee": 0,
    "is_available": true,
    "incentive_expires_at": null,
    "total_price": 21000,
    "tariff_used": "Regional",
    "driver_radius_km": 10,
    "max_distance_km": 30,
    "sub_services": [
      {
        "sub_service_id": 1,
        "service_id": 1,
        "name": "Motor",
        "description": null,
        "icon": null,
        "service_percent": 100,
        "service_flat": 0,
        "sub_service_percent": 100,
        "sub_service_flat": 0,
        "base_price": 17500,
        "sub_service_adjustment": 0,
        "adjusted_price": 17500,
        "commission_fee": 3500,
        "incentive_fee": 0,
        "total_price": 21000,
        "nearby_drivers_count": 4
      }
    ]
  },
  "message": "Estimasi harga berhasil dihitung"
}
```

Tipe hasil: `json.RawMessage`, `string`, `error`. Bentuk `data` dapat berbeda
ketika `is_simple=true`.

## 3. RoutingDistance

```go
response, err := api.RoutingDistance(ctx, client.RoutingRequest{
    Mode: "motorcycle",
    Routes: []client.Waypoint{
        {Latitude: -7.250445, Longitude: 112.768845},
        {Latitude: -7.260000, Longitude: 112.780000},
    },
})
```

```json
{
  "data": {
    "distance_meters": 2450,
    "duration_seconds": 420,
    "provider": "geoapify",
    "mode": "motorcycle",
    "legs": [
      {
        "index": 0,
        "distance_meters": 2450,
        "duration_seconds": 420
      }
    ]
  },
  "message": "Jarak routing berhasil dihitung"
}
```

Tipe hasil: `client.Response[client.RoutingResponse]`.

## 4. CreateOrder

Contoh ini menggunakan order booking. Untuk order instan, hilangkan
`booking_at`.

```go
data, message, err := api.CreateOrder(ctx, map[string]any{
    "customer_uuid": "24a55c29-5be8-4381-b915-3d658685a5a8",
    "client_request_id": "11111111-1111-4111-8111-111111111111",
    "service_id": 1,
    "sub_service_id": 1,
    "payment_method_id": 1,
    "regency_id": "7171",
    "booking_at": "2026-09-13T09:00:00+08:00",
    "routes": []map[string]any{
        {"latitude": -7.250445, "longitude": 112.768845, "address": "Pickup"},
        {"latitude": -7.260000, "longitude": 112.780000, "address": "Tujuan"},
    },
})
```

```json
{
  "data": {
    "uuid": "f4851143-4f1e-4326-9ccd-bfd70fe01ab8",
    "user": {
      "uuid": "24a55c29-5be8-4381-b915-3d658685a5a8",
      "name": "Budi Santoso",
      "phone": "081234567890",
      "rating": 4.9,
      "total_rating": 12,
      "phone_verification_at": null
    },
    "driver": null,
    "merchant": null,
    "service": {"id": 1, "name": "Ride"},
    "sub_service": {"id": 1, "name": "Motor", "icon": null},
    "payment_method": {
      "id": 1,
      "name": "Cash",
      "code": "CASH",
      "qr_string": null
    },
    "payment_status": null,
    "regency": {"id": "7171", "name": "Kota Manado"},
    "logs": [
      {
        "from_status": null,
        "to_status": "BOOKING",
        "description": "Pesanan booking dibuat",
        "timestamp": "2026-09-12T04:00:00Z"
      }
    ],
    "trackings": [],
    "total_distance_meters": 2450,
    "total_price": 21000,
    "status": "BOOKING",
    "booking_at": "2026-09-13T01:00:00Z",
    "created_at": "2026-09-12T04:00:00Z",
    "updated_at": "2026-09-12T04:00:00Z",
    "routes": [
      {
        "id": 501,
        "order_id": 1001,
        "stop_index": 0,
        "latitude": -7.250445,
        "longitude": 112.768845,
        "distance_from_previous_meters": 0,
        "address": "Pickup",
        "note": null,
        "arrived_at": null,
        "created_at": "2026-09-12T04:00:00Z"
      },
      {
        "id": 502,
        "order_id": 1001,
        "stop_index": 1,
        "latitude": -7.26,
        "longitude": 112.78,
        "distance_from_previous_meters": 2450,
        "address": "Tujuan",
        "note": null,
        "arrived_at": null,
        "created_at": "2026-09-12T04:00:00Z"
      }
    ],
    "items": [
      {
        "id": 701,
        "order_id": 1001,
        "name": "Perjalanan (Motor)",
        "amount": 17500,
        "quantity": 1,
        "item_type": "ride_fee",
        "product_id": null,
        "created_at": "2026-09-12T04:00:00Z"
      },
      {
        "id": 702,
        "order_id": 1001,
        "name": "Biaya Layanan",
        "amount": 3500,
        "quantity": 1,
        "item_type": "application_commission",
        "product_id": null,
        "created_at": "2026-09-12T04:00:00Z"
      }
    ],
    "ratings": [],
    "delivery_detail": null,
    "delivery_items": [],
    "balance_transfers": []
  },
  "message": "Order berhasil dibuat"
}
```

Tipe hasil: `json.RawMessage`, `string`, `error`.

## 5. ListOrders

```go
data, message, err := api.ListOrders(ctx, client.PricingPreviewParams{
    "page": {"1"},
    "limit": {"10"},
    "status": {"BOOKING"},
})
```

Response `data` berupa array detail order, bukan envelope pagination:

```json
{
  "data": [
    {
      "uuid": "f4851143-4f1e-4326-9ccd-bfd70fe01ab8",
      "user": {
        "uuid": "24a55c29-5be8-4381-b915-3d658685a5a8",
        "name": "Budi Santoso",
        "phone": "081234567890",
        "rating": 4.9,
        "total_rating": 12,
        "phone_verification_at": null
      },
      "driver": null,
      "merchant": null,
      "service": {"id": 1, "name": "Ride"},
      "sub_service": {"id": 1, "name": "Motor", "icon": null},
      "payment_method": {
        "id": 1,
        "name": "Cash",
        "code": "CASH",
        "qr_string": null
      },
      "payment_status": null,
      "regency": {"id": "7171", "name": "Kota Manado"},
      "logs": [],
      "trackings": [],
      "total_distance_meters": 2450,
      "total_price": 21000,
      "status": "BOOKING",
      "booking_at": "2026-09-13T01:00:00Z",
      "created_at": "2026-09-12T04:00:00Z",
      "updated_at": "2026-09-12T04:00:00Z",
      "routes": [],
      "items": [],
      "ratings": [],
      "delivery_detail": null,
      "delivery_items": [],
      "balance_transfers": []
    }
  ],
  "message": "Daftar order berhasil diambil"
}
```

Jika tidak ada order:

```json
{
  "data": [],
  "message": "Daftar order berhasil diambil"
}
```

Tipe hasil: `json.RawMessage`, `string`, `error`.

## 6. ShowOrder

```go
data, message, err := api.ShowOrder(ctx, orderUUID)
```

```json
{
  "data": {
    "uuid": "f4851143-4f1e-4326-9ccd-bfd70fe01ab8",
    "user": {
      "uuid": "24a55c29-5be8-4381-b915-3d658685a5a8",
      "name": "Budi Santoso",
      "phone": "081234567890",
      "rating": 4.9,
      "total_rating": 12,
      "phone_verification_at": null
    },
    "driver": {
      "uuid": "70df822b-2252-4141-8d64-b95feadf3698",
      "name": "Andi Driver",
      "plate_number": "DB 1234 AB",
      "rating": 4.8,
      "total_rating": 48,
      "phone_verification_at": "2026-08-01T02:00:00Z",
      "image_path": "https://cdn.example.com/drivers/andi.jpg"
    },
    "merchant": null,
    "service": {"id": 1, "name": "Ride"},
    "sub_service": {"id": 1, "name": "Motor", "icon": null},
    "payment_method": {
      "id": 1,
      "name": "Cash",
      "code": "CASH",
      "qr_string": null
    },
    "payment_status": null,
    "regency": {"id": "7171", "name": "Kota Manado"},
    "logs": [
      {
        "from_status": "BOOKING",
        "to_status": "BOOKING_ACCEPTED",
        "description": "Driver accepted the booking order",
        "timestamp": "2026-09-12T04:15:00Z"
      }
    ],
    "trackings": [],
    "total_distance_meters": 2450,
    "total_price": 21000,
    "status": "BOOKING_ACCEPTED",
    "booking_at": "2026-09-13T01:00:00Z",
    "created_at": "2026-09-12T04:00:00Z",
    "updated_at": "2026-09-12T04:15:00Z",
    "routes": [],
    "items": [],
    "ratings": [],
    "delivery_detail": null,
    "delivery_items": [],
    "balance_transfers": []
  },
  "message": "Detail order berhasil diambil"
}
```

Struktur lengkap `data` sama dengan response `CreateOrder`. Tipe hasil:
`json.RawMessage`, `string`, `error`.

## 7. CancelOrder

```go
message, err := api.CancelOrder(ctx, orderUUID, &client.CancelRequest{
    Reason: "Customer membatalkan order",
})
```

```json
{
  "message": "Order berhasil dibatalkan"
}
```

Tipe hasil: `string`, `error`.

## 8. ReviewDriver

```go
data, message, err := api.ReviewDriver(ctx, orderUUID, client.ReviewRequest{
    Rating: 5,
    Comment: "Pelayanan baik",
})
```

```json
{
  "data": {
    "order_id": 1001,
    "rater_type": "customer",
    "rater_id": 201,
    "ratee_type": "driver",
    "ratee_id": 301,
    "overall_rating": 5,
    "reviews": ["Pelayanan baik"],
    "cleanliness": null,
    "neatness": null,
    "fragrance": null,
    "punctuality": null,
    "attitude": null,
    "service_quality": null,
    "payment_punctuality": null,
    "order_clarity": null,
    "created_at": "2026-09-13T02:30:00Z"
  },
  "message": "Penilaian driver berhasil disimpan"
}
```

Tipe hasil: `json.RawMessage`, `string`, `error`.

## 9. ListChatMessages

```go
response, err := api.ListChatMessages(ctx, orderUUID, client.ChatMessagesParams{
    Page: 1,
    Limit: 50,
})
```

```json
{
  "data": {
    "items": [
      {
        "id": 50,
        "sender_id": 301,
        "sender_role": "driver",
        "message": "Saya menuju lokasi pickup",
        "message_type": "text",
        "image_path": null,
        "image_url": null,
        "created_at": "2026-09-12T14:37:06Z",
        "is_read": false
      }
    ],
    "total_items": 1,
    "total_pages": 1,
    "current_page": 1,
    "items_per_page": 50
  },
  "message": "Chat berhasil diambil"
}
```

Tipe hasil: `client.Response[client.ChatMessagesPage]`.

## 10. SendChatMessage

```go
response, err := api.SendChatMessage(ctx, orderUUID, client.SendChatMessageRequest{
    Message: "Driver, mohon ke pickup",
    MessageType: "text",
})
```

```json
{
  "data": {
    "id": 51,
    "sender_id": 201,
    "sender_role": "customer",
    "message": "Driver, mohon ke pickup",
    "message_type": "text",
    "image_path": null,
    "image_url": null,
    "created_at": "2026-09-12T14:38:06Z",
    "is_read": false
  },
  "message": "Chat berhasil dikirim"
}
```

Tipe hasil: `client.Response[client.ChatMessage]`.

## 11. MarkChatRead

```go
response, err := api.MarkChatRead(ctx, orderUUID, client.MarkChatReadRequest{
    LastReadMessageID: 51,
})
```

```json
{
  "data": {
    "last_read_message_id": 51
  },
  "message": "Chat berhasil ditandai telah dibaca"
}
```

Tipe hasil: `client.Response[client.MarkChatReadResponse]`.

## 12. SendChatImage

```go
file, err := os.Open("pickup.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

response, err := api.SendChatImage(ctx, orderUUID, client.SendChatImageRequest{
    FileName:    "pickup.jpg",
    ContentType: "image/jpeg",
    Image:       file,
    Message:     "Lokasi pickup saya",
})
```

```json
{
  "data": {
    "id": 52,
    "sender_id": 201,
    "sender_role": "customer",
    "message": "Lokasi pickup saya",
    "message_type": "image",
    "image_path": "83ca7ca8-a2ba-4efe-a247-89e15e23b312.jpg",
    "image_url": "https://api.example.com/api/files/83ca7ca8-a2ba-4efe-a247-89e15e23b312.jpg",
    "created_at": "2026-09-12T14:39:06Z",
    "is_read": false
  },
  "message": "Gambar chat berhasil dikirim"
}
```

Tipe hasil: `client.Response[client.ChatMessage]`.

## 13. ReverseGeocode

```go
response, err := api.ReverseGeocode(ctx, client.ReverseGeocodeRequest{
    Latitude:  1.4748,
    Longitude: 124.8421,
})
```

```json
{
  "data": {
    "formatted": "Jalan Sam Ratulangi, Wenang, Manado, Sulawesi Utara, Indonesia",
    "address_line1": "Jalan Sam Ratulangi",
    "address_line2": "Wenang, Manado, Sulawesi Utara, Indonesia",
    "street": "Jalan Sam Ratulangi",
    "house_number": null,
    "city": "Manado",
    "district": "Wenang",
    "state": "Sulawesi Utara",
    "postcode": "95111",
    "country": "Indonesia",
    "country_code": "id",
    "latitude": 1.4748,
    "longitude": 124.8421,
    "result_type": "street",
    "provider": "geoapify"
  },
  "message": "Alamat berhasil ditemukan"
}
```

Tipe hasil: `client.Response[client.ReverseGeocodeResponse]`.

## 14. ReviewApplication

```go
rating := int16(5)
response, err := api.ReviewApplication(ctx, client.AppReviewRequest{
    CustomerUUID: "24a55c29-5be8-4381-b915-3d658685a5a8",
    Category:     "service",
    Rating:       &rating,
    Review:       "Aplikasi sangat membantu",
})
```

```json
{
  "data": {
    "uuid": "76d64027-87e9-4566-bac3-e3a7f24f65ca",
    "customer_uuid": "24a55c29-5be8-4381-b915-3d658685a5a8",
    "category": "service",
    "rating": 5,
    "review": "Aplikasi sangat membantu",
    "status": "pending",
    "created_at": "2026-09-12T12:00:00Z",
    "updated_at": "2026-09-12T12:00:00Z"
  },
  "message": "Review aplikasi berhasil dikirim"
}
```

Tipe hasil: `client.Response[client.AppReview]`.

## Response error

Semua HTTP non-2xx dikembalikan sebagai `*client.APIError`. SDK mendukung
response upstream sederhana, response validasi, dan envelope dari proxy.

Response error umum:

```json
{
  "message": "Resource tidak ditemukan"
}
```

Response validasi:

```json
{
  "status": "error",
  "message": "Validation failed",
  "errors": {
    "email": ["invalid email"],
    "routes": ["length must be between 2 and 6"]
  }
}
```

Response error dari server example lokal:

```json
{
  "error": {
    "status_code": 422,
    "message": "Validation failed",
    "fields": {
      "email": ["invalid email"]
    }
  }
}
```

Cara membaca error:

```go
var apiErr *client.APIError
if errors.As(err, &apiErr) {
    log.Printf("HTTP=%d code=%s message=%s fields=%v",
        apiErr.StatusCode, apiErr.Code, apiErr.Message, apiErr.Fields)
}
```

Nilai `Code` dapat kosong karena backend utama saat ini tidak selalu mengirim
kode error terpisah. `StatusCode`, `Message`, dan field validasi tetap tersedia.
