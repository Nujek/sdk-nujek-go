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

Method yang tersedia: `Register`, `PricingPreview`, `RoutingDistance`,
`CreateOrder`, `CancelOrder`, dan `ReviewDriver`. `CreateOrder` menerima object
JSON apa pun selama memuat `customer_uuid` dan field order yang diwajibkan API.

`New` menerima `WithHTTPClient`, `WithClock`, dan `WithNonceGenerator` untuk
custom transport atau testing.
