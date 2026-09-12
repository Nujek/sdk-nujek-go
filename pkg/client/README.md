# Partner API client

Package ini memanggil seluruh endpoint partner di `/api/client` dan otomatis
membuat header `X-Client-Key`, `X-Timestamp`, `X-Nonce`, serta signature
HMAC-SHA256 sesuai kontrak API.

```go
partner, err := client.New(
    "https://api.example.com",
    os.Getenv("CLIENT_API_KEY"),
    os.Getenv("CLIENT_API_SECRET"),
)
if err != nil { log.Fatal(err) }

registered, err := partner.Register(ctx, "Budi", "budi@example.com", "081234567890")
route, err := partner.RoutingDistance(ctx, client.RoutingRequest{
    Mode: "motorcycle",
    Routes: []client.Waypoint{{Latitude: -7.250445, Longitude: 112.768845},
        {Latitude: -7.260000, Longitude: 112.780000}},
})
```

Method yang tersedia: `Register`, `PricingPreview`, `RoutingDistance`, `ReverseGeocode`,
`CreateOrder`, `ListOrders`, `ShowOrder`, `CancelOrder`, `ReviewDriver`,
`ListChatMessages`, `SendChatMessage`, `MarkChatRead`, dan `SendChatImage`.
`CreateOrder` menerima object
JSON apa pun selama memuat `customer_uuid` dan field order yang diwajibkan API.
Jika `routes[].address` kosong atau tidak dikirim, Partner API mengisinya
otomatis dari koordinat sebelum menyimpan order.

Contoh response sukses dan error untuk setiap method tersedia di
[`API_RESPONSES.md`](../../API_RESPONSES.md). Dokumentasi tersebut juga
menjelaskan method yang mengembalikan tipe konkret dan method yang
mengembalikan `json.RawMessage`.

Chat Partner API menggunakan percakapan `customer_driver`. Pesan baru dari
driver dikirim melalui webhook event `chat.message`; SDK menyediakan
`VerifyWebhook`, `ParseWebhook`, dan model payload untuk seluruh event Nujek.
Panduan lengkap tersedia di [`WEBHOOKS.md`](../../WEBHOOKS.md).

Contoh lengkap seluruh method tersedia di [`examples/partner_api`](../../examples/partner_api).
Jalankan dengan `CLIENT_API_BASE_URL`, `CLIENT_API_KEY`, dan `CLIENT_API_SECRET`;
`ORDER_JSON`, `ORDER_UUID`, dan `REVIEW_ORDER_UUID` bersifat opsional.

`New` menerima `WithHTTPClient`, `WithClock`, dan `WithNonceGenerator` untuk
custom transport atau testing.
