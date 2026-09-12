# Webhook Partner API

Nujek mengirim request `POST` ke `webhook_url` milik partner dengan header:

| Header | Keterangan |
| --- | --- |
| `Content-Type` | `application/json` |
| `X-Webhook-Id` | UUID unik delivery; gunakan sebagai idempotency key |
| `X-Webhook-Timestamp` | Unix timestamp saat delivery dikirim |
| `X-Webhook-Signature` | HMAC-SHA256 hex menggunakan webhook secret |

Signature dihitung dari `<timestamp>\n<delivery_id>\n<raw_body>`. Gunakan raw body
persis seperti yang diterima, sebelum JSON diubah atau di-encode ulang.

```go
body, err := io.ReadAll(r.Body)
if err != nil {
    http.Error(w, "invalid body", http.StatusBadRequest)
    return
}
deliveryID := r.Header.Get("X-Webhook-Id")
if err := client.VerifyWebhook(
    webhookSecret,
    r.Header.Get("X-Webhook-Timestamp"),
    deliveryID,
    body,
    r.Header.Get("X-Webhook-Signature"),
); err != nil {
    http.Error(w, "invalid webhook", http.StatusUnauthorized)
    return
}

// Tolak jika deliveryID sudah pernah diproses, lalu simpan secara atomik.
event, err := client.ParseWebhook(body)
if err != nil {
    http.Error(w, "invalid webhook", http.StatusBadRequest)
    return
}

switch event.Event {
case client.WebhookEventChatMessage:
    message, err := client.DecodeWebhookData[client.ChatMessageWebhookData](event)
    _ = message
    _ = err
case client.WebhookEventOrderSOSCreated:
    sos, err := client.DecodeWebhookData[client.OrderSOSWebhookData](event)
    _ = sos
    _ = err
default:
    order, err := client.DecodeWebhookData[client.OrderWebhookData](event)
    _ = order
    _ = err
}

w.WriteHeader(http.StatusNoContent)
```

## Event yang dikirim

| Event | Bentuk `data` |
| --- | --- |
| `order.created` | `OrderWebhookData` |
| `driver.accepted` | `OrderWebhookData` |
| `driver.rejected` | `OrderWebhookData` |
| `driver.cancelled` | `OrderWebhookData` |
| `driver.arrived` | `OrderWebhookData` |
| `driver.picked_up` | `OrderWebhookData` |
| `driver.timeout` | `OrderWebhookData` |
| `order.finished` | `OrderWebhookData` |
| `order.finished_by_admin` | `OrderWebhookData` |
| `order.cancelled_by_admin` | `OrderWebhookData` |
| `order.cancelled_by_user` | `OrderWebhookData` |
| `order.driver_changed` | `OrderWebhookData` |
| `order.timeout` | `OrderWebhookData` |
| `chat.message` | `ChatMessageWebhookData` |
| `order.sos_created` | `OrderSOSWebhookData` |
| `user_client_revoked` | `UserClientRevokedWebhookData` |

`ParseWebhook` tetap menerima nama event yang belum dikenal agar penambahan event
backend di masa depan tidak langsung merusak receiver. Gunakan
`IsKnownWebhookEvent` bila aplikasi perlu membedakan event yang sudah didukung.

Balas dengan status HTTP `2xx` hanya setelah event berhasil diterima. Nujek akan
mencoba kembali delivery yang gagal, sehingga `X-Webhook-Id` wajib diproses secara
idempoten. Webhook secret berbeda dari API secret dan diberikan saat client dibuat
atau secret dirotasi.
