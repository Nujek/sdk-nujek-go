# Nujek Partner API SDK (Go)

SDK Go untuk Partner API Nujek. SDK otomatis membuat signature HMAC-SHA256
dengan header `X-Client-Key`, `X-Timestamp`, `X-Nonce`, dan `X-Signature`.
Rilis terbaru: `v0.1.12`.

Contoh JSON response sukses untuk **setiap method SDK**, response kosong, dan
seluruh bentuk error tersedia di [API_RESPONSES.md](./API_RESPONSES.md).

## Instalasi

```bash
go get github.com/Nujek/sdk-nujek-go@v0.1.12
```

## Method SDK dan endpoint upstream

| Method Go | HTTP endpoint |
| --- | --- |
| `Register` | `POST /api/client/register` |
| `RoutingDistance` | `POST /api/client/routing/distance` |
| `ReverseGeocode` | `POST /api/client/geocoding/reverse` |
| `PricingPreview` | `GET /api/client/pricing/preview` |
| `ReviewApplication` | `POST /api/client/reviews` |
| `CreateOrder` | `POST /api/client/orders` |
| `ListOrders` | `GET /api/client/orders` |
| `ShowOrder` | `GET /api/client/orders/{order_uuid}` |
| `CancelOrder` | `POST /api/client/orders/{order_uuid}/cancel` |
| `ReviewDriver` | `POST /api/client/orders/{order_uuid}/review-driver` |
| `ListChatMessages` | `GET /api/client/orders/{order_uuid}/chat/customer_driver/messages` |
| `SendChatMessage` | `POST /api/client/orders/{order_uuid}/chat/customer_driver/messages` |
| `MarkChatRead` | `POST /api/client/orders/{order_uuid}/chat/customer_driver/read` |
| `SendChatImage` | `POST /api/client/orders/{order_uuid}/chat/customer_driver/images` |

Seluruh 14 method di atas memiliki contoh response yang dapat langsung dipakai
sebagai fixture di [API_RESPONSES.md](./API_RESPONSES.md).

```go
import (
    "context"
    "github.com/Nujek/sdk-nujek-go/pkg/client"
)

api, err := client.New("https://api.example.com", "client-api-key", "client-api-secret")
result, err := api.Register(context.Background(), "Budi", "budi@example.com", "081234567890")
```

```go
rating := int16(5)
review, err := api.ReviewApplication(ctx, client.AppReviewRequest{
    CustomerUUID: customerUUID,
    Category: "service",
    Rating: &rating,
    Review: "Aplikasi sangat membantu",
})
```

`CreateOrder` menerima `map[string]any` agar field order baru tetap kompatibel.
`PricingPreview` mengembalikan data pricing sebagai `json.RawMessage`.

`CreateOrder` otomatis mengisi `routes[].address` melalui reverse geocoding
ketika address tidak dikirim, kosong, atau hanya berisi spasi.

```go
address, err := api.ReverseGeocode(ctx, client.ReverseGeocodeRequest{
    Latitude: 1.4748, Longitude: 124.8421,
})
fmt.Println(address.Data.Formatted)
```

```go
messages, err := api.ListChatMessages(ctx, orderUUID, client.ChatMessagesParams{
    Page: 1, Limit: 50,
})
sent, err := api.SendChatMessage(ctx, orderUUID, client.SendChatMessageRequest{
    Message: "Driver, mohon ke pickup",
    MessageType: "text",
})

read, err := api.MarkChatRead(ctx, orderUUID, client.MarkChatReadRequest{
    LastReadMessageID: 51,
})

file, err := os.Open("pickup.jpg")
if err != nil { log.Fatal(err) }
defer file.Close()
image, err := api.SendChatImage(ctx, orderUUID, client.SendChatImageRequest{
    FileName: "pickup.jpg", ContentType: "image/jpeg",
    Image: file, Message: "Lokasi pickup saya",
})
```

Pesan driver diteruskan ke webhook partner sebagai event `chat.message`. Gunakan
`VerifyWebhook` pada raw request body sebelum `ParseWebhook`. Daftar seluruh
event, model payload, dan contoh receiver tersedia di [WEBHOOKS.md](./WEBHOOKS.md).

## Example untuk Postman

Example HTTP lokal tersedia di [`examples/partner_api`](./examples/partner_api).
Endpoint lokal dibuat singkat dan sama dengan SDK Node.js:

| Method | Endpoint lokal |
| --- | --- |
| POST | `/register` |
| GET | `/pricing` |
| POST | `/routing` |
| POST | `/reviews` |
| POST | `/geocoding/reverse` |
| POST | `/orders` |
| POST | `/orders/{order_uuid}/cancel` |
| POST | `/orders/{order_uuid}/review-driver` |
| GET | `/orders/{order_uuid}/chat/messages` |
| POST | `/orders/{order_uuid}/chat/messages` |
| POST | `/orders/{order_uuid}/chat/read` |
| POST | `/orders/{order_uuid}/chat/images` (multipart) |

Jalankan:

```bash
cp examples/partner_api/.env.example examples/partner_api/.env
# isi CLIENT_API_BASE_URL, CLIENT_API_KEY, CLIENT_API_SECRET
go run ./examples/partner_api
```

Server default berjalan pada `http://localhost:8088`. Credential hanya dibaca
dari `.env`, sehingga Postman tidak perlu mengirim API key atau signature.
Dokumentasi request dan file OpenAPI tersedia di folder example.

## Testing

```bash
go test ./...
```

Repository upstream: `git@github.com:Nujek/sdk-nujek-go.git`.
