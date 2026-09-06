# SDK Nujek Go

Source SDK untuk seluruh Partner API Nujek (`/api/client`). Versi ini setara
dengan rilis GitHub `v0.1.1`.

Package utama berada di `pkg/client`; contoh lengkap tersedia di
`examples/partner_api`.

```bash
cd sdk-nujek-go
go test ./...
cp examples/partner_api/.env.example examples/partner_api/.env
# isi secret di examples/partner_api/.env, lalu:
set -a; source examples/partner_api/.env; set +a
go run ./examples/partner_api
```

Repository upstream: `git@github.com:Nujek/sdk-nujek-go.git`.
