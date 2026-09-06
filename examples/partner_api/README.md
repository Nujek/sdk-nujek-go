# Contoh lengkap Partner API Nujek Go

Example ini menunjukkan semua method SDK:

| Method | Endpoint |
| --- | --- |
| `Register` | `POST {{baseUrl}}/api/client/register` |
| `PricingPreview` | `GET {{baseUrl}}/api/client/pricing/preview` |
| `RoutingDistance` | `POST {{baseUrl}}/api/client/routing/distance` |
| `CreateOrder` | `POST {{baseUrl}}/api/client/orders` |
| `CancelOrder` | `POST {{baseUrl}}/api/client/orders/{order_uuid}/cancel` |
| `ReviewDriver` | `POST {{baseUrl}}/api/client/orders/{order_uuid}/review-driver` |

Untuk Postman, `baseUrl` adalah server example lokal: `http://localhost:8088`.
Server example membaca credentials upstream dari `.env`, sehingga Postman tidak
perlu mengirim API key atau signature.

## Persiapan

Pastikan Go 1.24 atau lebih baru:

```bash
cd core/sdk/sdk-nujek-go/examples/partner_api
cp .env.example .env
```

Isi minimal `.env`:

```dotenv
CLIENT_API_BASE_URL=https://staging-api.example.com
CLIENT_API_KEY=client-key-dari-admin
CLIENT_API_SECRET=client-secret-dari-admin
EXAMPLE_PORT=8088
```

Request body customer/order dikirim dari Postman. Jangan commit `.env`; file
tersebut sudah di-ignore.

## Import ke Postman

Import file [`openapi.json`](./openapi.json) melalui **Import → File**. Setelah
diimpor, buat environment Postman dengan variable berikut:

```text
baseUrl       = http://localhost:8088
```

Jalankan `go run .`, lalu jalankan request yang diimpor. Example akan meneruskan
request ke upstream dan menambahkan signature menggunakan credentials dari `.env`.

## Menjalankan server example

```bash
set -a
source .env
set +a
go run .
```

Server membuka port `8088` secara default. Gunakan `EXAMPLE_PORT` untuk port
lain. Setelah muncul log `listening`, buka Postman dan gunakan base URL lokal:

```text
http://localhost:8088
```

Server ini meneruskan request ke Partner API dengan signature dari SDK.

### Register dari Postman

```http
POST http://localhost:8088/api/client/register
Content-Type: application/json
```

```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "phone": "081234567890"
}
```

### Pricing preview dari Postman

```http
GET http://localhost:8088/api/client/pricing/preview?service_id=1&sub_service_id=1&regency_id=7171&distance_km=5.5
```

### Routing distance dari Postman

```http
POST http://localhost:8088/api/client/routing/distance
Content-Type: application/json
```

```json
{
  "mode": "motorcycle",
  "routes": [
    {"latitude": -7.250445, "longitude": 112.768845},
    {"latitude": -7.260000, "longitude": 112.780000}
  ]
}
```

Route lokal lainnya adalah `POST /orders`, `POST /orders/{order_uuid}/cancel`,
dan `POST /orders/{order_uuid}/review-driver`; body-nya sama dengan contoh
SDK di bawah.

## Penggunaan setiap API

### Register customer

```go
result, err := api.Register(ctx, "Budi", "budi@example.com", "081234567890")
```

Response berisi `uuid` customer. Simpan UUID tersebut untuk `CUSTOMER_UUID`.
Request identik bersifat idempotent untuk client yang sama.

### Pricing preview

```go
pricing, message, err := api.PricingPreview(ctx, client.PricingPreviewParams{
    "service_id": {"1"}, "sub_service_id": {"1"},
    "regency_id": {"7171"}, "distance_km": {"5.5"},
})
```

`service_id` wajib dan hanya mendukung `1` atau `2`. Parameter lain: 
`sub_service_id`, `regency_id`, `regency_name`, `distance_km`, `is_simple`,
`latitude`, dan `longitude`. Response pricing dikembalikan sebagai JSON mentah.

### Routing distance

```go
result, err := api.RoutingDistance(ctx, client.RoutingRequest{
    Mode: "motorcycle", // atau "drive"
    Routes: []client.Waypoint{
        {Latitude: -7.250445, Longitude: 112.768845},
        {Latitude: -7.260000, Longitude: 112.780000},
    },
})
```

`Routes` harus berisi 2–6 titik. Response berisi jarak meter, durasi detik,
provider, mode, dan rincian legs.

### Create order

```go
order := map[string]any{
    "customer_uuid": os.Getenv("CUSTOMER_UUID"),
    "client_request_id": "uuid-request-unik",
    "service_id": 1, "sub_service_id": 1, "payment_method_id": 1,
    "regency_id": "7171",
    "routes": []any{
        map[string]any{"latitude": -7.250445, "longitude": 112.768845},
        map[string]any{"latitude": -7.260000, "longitude": 112.780000},
    },
}
created, message, err := api.CreateOrder(ctx, order)
```

`client_request_id` wajib unik untuk idempotensi, `service_id` hanya `1` atau
`2`, dan `routes` harus 2–6 titik. Field `items`, `voucher_code`,
`merchant_uuid`, `driver_uuid`, dan `delivery_detail` mengikuti kontrak order.

### Cancel order

```go
message, err := api.CancelOrder(ctx, orderUUID,
    &client.CancelRequest{Reason: "Alasan pembatalan"})
```

Request body boleh `nil`. Order harus milik client yang sama.

### Review driver

```go
result, message, err := api.ReviewDriver(ctx, orderUUID, client.ReviewRequest{
    Rating: 5, Comment: "Pelayanan baik",
})
```

Rating harus 1–5 dan order harus sudah selesai.

## Variable opsional

```dotenv
CUSTOMER_UUID=uuid-customer-dari-register
ORDER_JSON={"client_request_id":"uuid-request-unik","service_id":1,"sub_service_id":1,"payment_method_id":1,"regency_id":"7171","routes":[{"latitude":-7.250445,"longitude":112.768845},{"latitude":-7.260000,"longitude":112.780000}]}
ORDER_UUID=uuid-order-untuk-cancel
REVIEW_ORDER_UUID=uuid-order-selesai
REVIEW_RATING=5
REVIEW_COMMENT=Pelayanan baik
```

## Error dan keamanan

SDK otomatis mengirim `X-Client-Key`, `X-Timestamp`, `X-Nonce`, dan
`X-Signature` HMAC-SHA256. Error HTTP bertipe `*client.APIError` dengan field
`StatusCode`, `Code`, `Message`, dan `Fields`.

- `401/403`: credentials, signature, timestamp, nonce, atau IP whitelist salah.
- `422`: payload tidak memenuhi validasi backend.
- `404`: UUID tidak ditemukan atau bukan milik client.
- Jangan menaruh API secret di source code atau commit `.env`.
