# Backend HalalTrip

Backend HalalTrip menggunakan Go, Gin, PostgreSQL, dan pgx. Rekomendasi serta data destinasi yang dikelola provider diteruskan ke Travel Recommendation API. Backend tidak menyalin destinasi ke PostgreSQL dan tidak membuat rekomendasi lokal.

## Persyaratan dan setup

- Go 1.24 atau lebih baru
- PostgreSQL terpasang dan Goose CLI untuk migration

Buat database `halaltrip`, salin `.env.example` menjadi `.env`, lalu sesuaikan `DATABASE_URL` dengan password PostgreSQL lokal. Migration membuat tabel destinations, users, user_preferences, direktori tempat, saved_places, trips, dan itinerary_items tanpa data contoh. Auth memerlukan `JWT_SECRET` minimal 32 karakter; `JWT_ACCESS_TOKEN_TTL` mengatur masa berlaku access token (default `1h`). `JWT_ISSUER` dan `JWT_AUDIENCE` opsional; jika diisi, token harus cocok dengan keduanya.

Dari direktori `be/`, jalankan migration dan backend:

    go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
    $env:DATABASE_URL = "postgres://postgres:PASSWORD_ANDA@localhost:5432/halaltrip?sslmode=disable"
    goose -dir migrations postgres "$env:DATABASE_URL" up
    go run ./cmd/api

## Authentication dan profil

- `POST /api/v1/auth/register`: menerima `name`, `email`, `password` (minimal 8 karakter). Respons memberi profil tanpa password dan access token.
- `POST /api/v1/auth/login`: menerima `email` dan `password`; kredensial salah selalu memakai pesan generik.
- `GET /api/v1/users/me`: membaca profil berdasarkan identitas dari JWT.
- `PATCH /api/v1/users/me`: memperbarui `name` dan/atau `email`.
- `GET /api/v1/users/me/preferences`: mengambil preferensi; 404 berarti belum disimpan.
- `PATCH /api/v1/users/me/preferences`: menyimpan `interests`, `activity`, dan `travel_type`.

Contoh request:

    curl -X POST http://localhost:8080/api/v1/auth/register -H "Content-Type: application/json" -d '{"name":"Alya","email":"alya@example.com","password":"password-kuat"}'
    curl http://localhost:8080/api/v1/users/me -H "Authorization: Bearer TOKEN"
    curl -X PATCH http://localhost:8080/api/v1/users/me/preferences -H "Authorization: Bearer TOKEN" -H "Content-Type: application/json" -d '{"interests":["culture","culinary"],"activity":"Exploration","travel_type":"Solo"}'

Endpoint rekomendasi tetap meneruskan perhitungan ke Travel Recommendation API eksternal. Token pada endpoint rekomendasi bersifat opsional. Jika ketiga field preferensi tidak dikirim, token valid dengan preferensi tersimpan akan mengisi ketiganya. Jika field tidak dikirim dan token atau preferensi tersimpan tidak tersedia, backend membalas 400; payload preferensi parsial juga ditolak. Preferensi eksplisit yang lengkap tetap dipakai.

## Endpoint internal

- `GET /health` — status proses API.
- `GET /health/ready` — kesiapan database; HTTP 503 bila database tidak siap.
- `GET /api/v1/health/recommendation-provider` — proxy pemeriksaan provider.
- `POST /api/v1/recommendations/cities` — proxy rekomendasi kota.
- `POST /api/v1/recommendations/destinations` — proxy rekomendasi destinasi.
- `GET /api/v1/destinations` dan `GET /api/v1/destinations/:id` — membaca destinasi provider.
- `POST /api/v1/destinations`, `PUT /api/v1/destinations/:id`, `DELETE /api/v1/destinations/:id` — CRUD destinasi provider.

## Place directory

Direktori restoran, masjid/fasilitas salat, dan hotel disimpan HalalTrip di PostgreSQL; data destinasi dari provider eksternal tetap tidak disalin. Endpoint baca bersifat publik: `GET /api/v1/restaurants`, `/mosques`, `/hotels`, beserta `/:id`. Daftar mendukung `page` (default 1), `page_size` (default 20, maksimal 100), `q`, `city`, `country`, `sort` (`name`, `city`, `created_at`, `updated_at`), dan `order` (`asc`/`desc`). Filter kategori: restoran `cuisine`, `halal_status`, `price_range`; masjid `facility`; hotel `price_range`, `feature`.

CRUD direktori (`POST`, `PUT /:id`, `DELETE /:id`) hanya untuk admin terautentikasi. Tidak ada endpoint untuk menaikkan role. Setelah membuat akun operator, berikan role admin melalui akses database tepercaya, misalnya:

    UPDATE users SET role='admin' WHERE lower(email)=lower('operator@example.com');

Pengguna terautentikasi dapat mengelola daftar pribadi melalui `GET/POST /api/v1/saved-places` dan `DELETE /api/v1/saved-places/:place_type/:place_id`. `place_type` adalah `restaurant`, `mosque`, atau `hotel`.

Status restoran `certified` wajib disertai `halal_verification_note` dan `source_url`. Fitur hotel dengan `verified: true` wajib menyertakan `verification_note` dan `source_url`. Field fasilitas masjid hanya menyimpan informasi yang dikirim; backend tidak mengisi jadwal atau fasilitas otomatis. Belum ada provider peta/tempat yang dikonfigurasi dan tidak ada seed data; sumber lokasi harus disertakan sebagai `source_url` bila tersedia.

## Trip planner dan itinerary

Semua endpoint `/api/v1/trips` memerlukan bearer token dan hanya menampilkan atau mengubah trip milik subject token. Trip memakai `title`, `start_date`, `end_date` (format `YYYY-MM-DD`), `origin`, `notes`, dan status `planning`, `active`, `completed`, atau `cancelled`.

- `POST /api/v1/trips`, `GET /api/v1/trips`
- `GET`, `PATCH`, `DELETE /api/v1/trips/:tripId`
- `POST`, `GET /api/v1/trips/:tripId/itinerary-items`
- `PATCH`, `DELETE /api/v1/trips/:tripId/itinerary-items/:itemId`
- `PUT /api/v1/trips/:tripId/itinerary-items/order` dengan body `{"item_ids":["<item-id-1>","<item-id-2>"]}`

Item menerima `title`, `description`, `place_type` (`destination`, `restaurant`, `mosque`, `hotel`, `custom_activity`), `external_place_id`, `scheduled_date`, `start_time`, `end_time`, `transport_mode`, dan `notes`. `external_place_id` adalah ID opaque dari sumber terkait dan tidak menyamakan ID provider destinasi dengan ID direktori lokal. Sequence dibuat berurutan saat item ditambahkan; endpoint order harus mengirim semua ID item tepat satu kali dan menyimpan perubahan dalam transaksi.

Tanggal item harus berada dalam rentang trip. Waktu memakai `HH:MM` atau `HH:MM:SS`; awal dan akhir harus diberikan bersama dan waktu akhir harus setelah waktu mulai. Backend tidak menghitung rute, jarak, atau estimasi transportasi. Endpoint rekomendasi eksternal tetap hanya membantu memilih kota/destinasi; pilihan pengguna disimpan sebagai trip itinerary terpisah.

Contoh request kota:

    curl -X POST http://localhost:8080/api/v1/recommendations/cities -H "Content-Type: application/json" -d '{"interests":["culture","history"],"activity":"Exploration","travel_type":"Solo","top_n":5}'

Contoh request destinasi:

    curl -X POST http://localhost:8080/api/v1/recommendations/destinations -H "Content-Type: application/json" -d '{"city":"Yogyakarta","interests":["culture","history"],"activity":"Exploration","travel_type":"Solo","top_n":5}'

`interests` harus berisi setidaknya satu dari `culinary`, `culture`, `nature`, `shopping`, `adventure`, atau `history`. `activity` menerima `Relaxing`, `Exploration`, atau `Adventure`; `travel_type` menerima `Solo`, `Couple`, `Family`, atau `Friends`. Field `city` wajib untuk rekomendasi destinasi. `top_n` opsional dan dibatasi backend pada rentang 1–100; batas ini adalah batas lokal sementara karena Swagger provider tidak dapat diakses saat implementasi.

Contoh response sukses diteruskan tanpa perubahan, misalnya `{"recommendations":[]}`. Provider dapat mengembalikan daftar rekomendasi kosong; kondisi ini tetap sukses. Request dan response CRUD diteruskan sebagai JSON provider tanpa membuang field. Backend tidak menambahkan envelope `data` pada response provider. Error provider disamarkan; timeout menjadi HTTP 504, respons provider gagal menjadi HTTP 502, dan 404 provider menjadi 404.

## Provider eksternal dan konfigurasi

Backend meneruskan hanya endpoint berikut: `GET /health`, `POST /recommend/cities`, `POST /recommend/destinations`, dan `GET`, `POST`, `PUT`, `DELETE /destinations` (dengan `/{id}` untuk operasi satu destinasi). Kontrak payload dan response provider dipertahankan. Backend tidak mengklaim endpoint hotel, restoran, masjid, transportasi, atau itinerary.

Base URL default provider adalah `https://travel-recommendation-api-production-75f5.up.railway.app`; ubah melalui environment:

    TRAVEL_RECOMMENDATION_API_URL=https://travel-recommendation-api-production-75f5.up.railway.app
    HTTP_CLIENT_TIMEOUT=15s

Swagger provider: [https://travel-recommendation-api-production-75f5.up.railway.app/docs](https://travel-recommendation-api-production-75f5.up.railway.app/docs). Endpoint internal memakai prefix `/api/v1`, sedangkan endpoint eksternal tidak memakai prefix tersebut.

## Format error dan verifikasi

Error internal memakai bentuk `{"error":{"code":"not_found","message":"Data tidak ditemukan."}}`. Test integrasi menggunakan mock HTTP server dan tidak bergantung pada koneksi internet atau API production.

    gofmt -w .
    go test ./...
    go vet ./...
    go build -o tmp/halaltrip-api.exe ./cmd/api

## Struktur backend dan environment

Susunan kode utama:

- `cmd/api`: entrypoint HTTP server.
- `internal/auth`, `middleware`, `config`, `router`: token, validasi konfigurasi, middleware, dan route.
- `internal/users`, `directory`, `trips`: handler, aturan bisnis, dan penyimpanan PostgreSQL per domain.
- `internal/recommendation`, `travelprovider`: request rekomendasi dan proxy provider eksternal.
- `migrations`: skema database berurutan; `docs/openapi.yaml` adalah kontrak OpenAPI 3.1.

Konfigurasi wajib adalah `DATABASE_URL` dan `JWT_SECRET` (minimal 32 karakter acak). `JWT_ACCESS_TOKEN_TTL` default `1h`; `JWT_ISSUER` dan `JWT_AUDIENCE` opsional di kode, tetapi contoh konfigurasi mengisinya agar validasi token membatasi penerbit dan API penerima yang diharapkan. Gunakan nilai audience yang konsisten dengan API ini pada semua instance backend. `CORS_ALLOWED_ORIGINS` default `http://localhost:3000`; `HTTP_CLIENT_TIMEOUT` default `15s`; `TRAVEL_RECOMMENDATION_API_URL` menunjuk provider eksternal. Jangan menyimpan secret production ke Git.

Untuk frontend lokal, set `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1`. Health endpoint berada di root server (`/health`, `/health/ready`), di luar prefix API.

## PostgreSQL lokal tanpa container

Jalankan PostgreSQL yang terpasang pada mesin lokal, buat database `halaltrip`, lalu dari PowerShell di folder `be`:

    Copy-Item .env.example .env
    $env:DATABASE_URL = "postgres://postgres:PASSWORD_ANDA@localhost:5432/halaltrip?sslmode=disable"
    $env:JWT_SECRET = "GANTI_DENGAN_RANDOM_SECRET_MINIMAL_32_KARAKTER"
    go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
    goose -dir migrations postgres "$env:DATABASE_URL" up
    go run ./cmd/api

`go run` memuat `.env` saat tersedia. Alternatifnya isi environment shell secara eksplisit. Backend tidak menjalankan migrasi otomatis saat startup. File `docker-compose.yml` hanya diperiksa secara statis; langkah pengembangan dan verifikasi proyek ini menggunakan PostgreSQL lokal tanpa Docker.

## OpenAPI dan integration test database

Kontrak endpoint tersedia di [`docs/openapi.yaml`](docs/openapi.yaml); panduan ringkas tambahan berada di [`docs/api.md`](docs/api.md). Skema request/response CRUD destinasi provider sengaja ditandai opaque karena backend meneruskannya tanpa transformasi dan tidak mengasumsikan format provider.

`go test ./...` menjalankan unit/handler tests tanpa database. Integration test repository PostgreSQL dilewati jika `TEST_DATABASE_URL` tidak disetel. Untuk menjalankannya, arahkan variabel tersebut ke database PostgreSQL **khusus test** yang boleh membuat dan menghapus schema, lalu jalankan `go test ./integration -v`. Test menggunakan schema acak terpisah dan menghapusnya saat selesai. Contoh PowerShell:

    $env:TEST_DATABASE_URL = "postgres://postgres:PASSWORD_ANDA@localhost:5432/halaltrip_test?sslmode=disable"
    go test ./integration -v

Jangan arahkan `TEST_DATABASE_URL` ke database production atau database yang menyimpan data yang perlu dipertahankan.
