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
`Services`, `NearbyDrivers`, `CreateOrder`, `ListOrders`, `ShowOrder`, `CancelOrder`, `ReviewDriver`,
`ListChatMessages`, `SendChatMessage`, `MarkChatRead`, dan `SendChatImage`.
`CreateOrder` menerima object
JSON apa pun selama memuat `customer_uuid` dan field order yang diwajibkan API.
Jika `routes[].address` kosong atau tidak dikirim, Partner API mengisinya
otomatis dari koordinat sebelum menyimpan order.

Untuk membuat booking, tambahkan `booking_at` sebagai timestamp RFC 3339 dengan
timezone, misalnya `2026-09-14T10:00:00+08:00`. Waktu harus di masa depan dan
`driver_uuid` tidak boleh dikirim. Order akan berstatus `BOOKING` sampai waktu
booking tiba. Client API saat ini mendukung `service_id` `1` dan `2` untuk order.

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

`Services(ctx, ServicesRequest{Latitude: ..., Longitude: ...})` mengembalikan
kota terdekat, tarif layanan yang berlaku (`regional` atau `default`), dan
sub-service beserta persentase serta biaya tetapnya.

`NearbyDrivers` membutuhkan `SubServiceID` dan koordinat. `RadiusKM` opsional;
nilai nol memakai radius default server 5 km. Nilai maksimum radius adalah 100
km. Response hanya berisi driver online dan eligible untuk sub-service tersebut,
serta menggunakan `image_url` untuk link foto driver.

`Services` menerima latitude dan longitude, lalu mengembalikan kota terdekat,
tarif regional/default, daftar service, sub-service, dan
`nearby_drivers_count` pada setiap sub-service. `ServicesRequest.ServiceID`
opsional untuk memfilter satu service berdasarkan ID.
