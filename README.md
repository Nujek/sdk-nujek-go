# Nujek Partner API SDK (Go)

SDK Go untuk Partner API Nujek. SDK otomatis membuat signature HMAC-SHA256
dengan header `X-Client-Key`, `X-Timestamp`, `X-Nonce`, dan `X-Signature`.
Rilis terbaru: `v0.1.6`.

## Instalasi

```bash
go get github.com/Nujek/sdk-nujek-go@v0.1.6
```

## Method SDK dan endpoint upstream

| Method Go | HTTP endpoint |
| --- | --- |
| `Register` | `POST /api/client/register` |
| `PricingPreview` | `GET /api/client/pricing/preview` |
| `RoutingDistance` | `POST /api/client/routing/distance` |
| `CreateOrder` | `POST /api/client/orders` |
| `ListOrders` | `GET /api/client/orders` |
| `ShowOrder` | `GET /api/client/orders/{order_uuid}` |
| `CancelOrder` | `POST /api/client/orders/{order_uuid}/cancel` |
| `ReviewDriver` | `POST /api/client/orders/{order_uuid}/review-driver` |

```go
import (
    "context"
    "github.com/Nujek/sdk-nujek-go/pkg/client"
)

api, err := client.New("https://api.example.com", "client-api-key", "client-api-secret")
result, err := api.Register(context.Background(), "Budi", "budi@example.com", "081234567890")
```

`CreateOrder` menerima `map[string]any` agar field order baru tetap kompatibel.
`PricingPreview` mengembalikan data pricing sebagai `json.RawMessage`.

## Example untuk Postman

Example HTTP lokal tersedia di [`examples/partner_api`](./examples/partner_api).
Endpoint lokal dibuat singkat dan sama dengan SDK Node.js:

| Method | Endpoint lokal |
| --- | --- |
| POST | `/register` |
| GET | `/pricing` |
| POST | `/routing` |
| POST | `/orders` |
| POST | `/orders/{order_uuid}/cancel` |
| POST | `/orders/{order_uuid}/review-driver` |

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
