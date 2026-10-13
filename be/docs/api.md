# API backend HalalTrip

Semua endpoint internal berada di prefix `/api/v1`. Endpoint provider berada di root base URL eksternal. Backend meneruskan JSON response provider tanpa mengubah bentuknya.

## Health dan rekomendasi

### `GET /api/v1/health/recommendation-provider`

Memeriksa `GET /health` provider.

### `POST /api/v1/recommendations/cities`

Request:

    {"interests":["culture","history"],"activity":"Exploration","travel_type":"Solo","top_n":5}

### `POST /api/v1/recommendations/destinations`

Request:

    {"city":"Yogyakarta","interests":["culture","history"],"activity":"Exploration","travel_type":"Solo","top_n":5}

`interests` berisi minimal satu dari `culinary`, `culture`, `nature`, `shopping`, `adventure`, `history`. `activity`: `Relaxing`, `Exploration`, `Adventure`. `travel_type`: `Solo`, `Couple`, `Family`, `Friends`. `city` wajib pada rekomendasi destinasi. `top_n` opsional dengan batas lokal backend 1–100; batas provider belum dapat diverifikasi karena Swagger tidak dapat dijangkau ketika implementasi.

Contoh response yang diteruskan:

    {"recommendations":[]}

Array kosong adalah hasil rekomendasi yang valid. Response aktual mengikuti kontrak provider dan tidak dibungkus envelope baru.

## CRUD destinasi

| Endpoint internal | Endpoint eksternal |
| --- | --- |
| `GET /api/v1/destinations` | `GET /destinations` |
| `GET /api/v1/destinations/:id` | `GET /destinations/{id}` |
| `POST /api/v1/destinations` | `POST /destinations` |
| `PUT /api/v1/destinations/:id` | `PUT /destinations/{id}` |
| `DELETE /api/v1/destinations/:id` | `DELETE /destinations/{id}` |

Body POST/PUT dan response diteruskan sebagai JSON tanpa perubahan agar field provider tetap terjaga. Periksa Swagger untuk field CRUD provider: [Swagger UI](https://travel-recommendation-api-production-75f5.up.railway.app/docs). Backend hanya menyediakan endpoint yang tercantum di tabel dan tidak menyimpan salinan destinasi di PostgreSQL.

## Place directory

Data restoran, masjid/fasilitas salat, dan hotel disimpan oleh HalalTrip. Endpoint baca publik:

| Resource | List | Detail | Category filters |
| --- | --- | --- | --- |
| Restaurants | `GET /api/v1/restaurants` | `GET /api/v1/restaurants/:id` | `cuisine`, `halal_status`, `price_range` |
| Mosques | `GET /api/v1/mosques` | `GET /api/v1/mosques/:id` | `facility` |
| Hotels | `GET /api/v1/hotels` | `GET /api/v1/hotels/:id` | `price_range`, `feature` |

Semua list menerima `page`, `page_size` (maksimum 100), `q`, `city`, `country`, `sort` (`name`, `city`, `created_at`, `updated_at`), dan `order` (`asc`, `desc`). Respons list memiliki `data` dan `pagination` (`page`, `page_size`, `total`, `total_pages`). Filter kota/negara dan kategori memakai kecocokan case-insensitive.

Admin dengan bearer token dan role `admin` dari kolom database dapat memakai `POST /api/v1/{restaurants|mosques|hotels}`, `PUT /api/v1/{resource}/:id`, dan `DELETE /api/v1/{resource}/:id`. Role tidak berasal dari request/token; role diberikan operator melalui akses database tepercaya. Tidak ada route promosi role di API.

Restaurant body memuat `name`, `city`, `country`, dan opsional `description`, `address`, pasangan `latitude`/`longitude`, `cuisine`, `price_range`, `halal_status`, `halal_verification_note`, `source_url`. Nilai `halal_status`: `certified`, `muslim_owned`, `halal_friendly`, `no_pork`, `unknown`; status `certified` harus disertai catatan verifikasi dan URL bukti.

Mosque body memakai field lokasi umum serta `facilities` berupa array string dan `source_url`. Hotel body memakai field lokasi umum, `price_range`, `muslim_friendly_features` berupa array `{name, verified, verification_note}`, dan `source_url`. Fitur hotel terverifikasi wajib memiliki catatan dan URL sumber. Koordinat harus diberikan berpasangan dan berada pada rentang latitude -90..90, longitude -180..180.

Saved places bersifat privat per user: `GET /api/v1/saved-places`, `POST /api/v1/saved-places` dengan body `{"place_type":"restaurant","place_id":"<uuid>"}`, dan `DELETE /api/v1/saved-places/:place_type/:place_id`. Semua memerlukan bearer token; response hanya berisi data tempat publik yang disimpan pemilik token.

Tidak ada sumber peta atau places eksternal yang diasumsikan aktif. Tidak ada data seed atau klaim real-time.

## Trip planner dan itinerary

Semua route trip memerlukan bearer token; query selalu dibatasi ke pemilik trip dari subject token.

| Method dan endpoint | Fungsi |
| --- | --- |
| `POST /api/v1/trips` | Membuat trip |
| `GET /api/v1/trips` | Daftar trip pengguna |
| `GET /api/v1/trips/:tripId` | Detail trip milik pengguna |
| `PATCH /api/v1/trips/:tripId` | Memperbarui field trip |
| `DELETE /api/v1/trips/:tripId` | Menghapus trip beserta item (cascade) |
| `POST /api/v1/trips/:tripId/itinerary-items` | Menambahkan item, sequence berikutnya diberikan server |
| `GET /api/v1/trips/:tripId/itinerary-items` | Daftar item menurut sequence |
| `PATCH /api/v1/trips/:tripId/itinerary-items/:itemId` | Memperbarui item |
| `DELETE /api/v1/trips/:tripId/itinerary-items/:itemId` | Menghapus item |
| `PUT /api/v1/trips/:tripId/itinerary-items/order` | Menyusun ulang item secara transaksional |

Contoh create trip: `{"title":"Liburan Kyoto","start_date":"2026-11-01","end_date":"2026-11-05","origin":"Bangkok","notes":""}`. Tanggal wajib `YYYY-MM-DD`; end date tidak boleh sebelum start date. Status opsional: `planning`, `active`, `completed`, `cancelled`.

Contoh create item: `{"title":"Kunjungi Kiyomizu-dera","place_type":"destination","external_place_id":"provider-id-as-opaque-string","scheduled_date":"2026-11-02","start_time":"09:00","end_time":"11:00","transport_mode":"train"}`. `place_type` menerima `destination`, `restaurant`, `mosque`, `hotel`, `custom_activity`. ID eksternal adalah nilai opaque; identifier provider rekomendasi tidak diasumsikan sama dengan UUID place lokal. Item kustom boleh tanpa `external_place_id`.

Order body harus berupa daftar penuh tanpa duplikasi: `{"item_ids":["uuid-item-a","uuid-item-b"]}`. Repository mengunci trip dan seluruh item dalam transaksi lalu menerapkan sequence 1..N; order parsial atau stale ditolak.

Jadwal item opsional, tetapi jika ada, tanggal harus berada di dalam rentang trip. `start_time` dan `end_time` harus sama-sama tersedia atau kosong; formatnya `HH:MM` atau `HH:MM:SS`, dan waktu selesai harus setelah waktu mulai. Mengubah tanggal trip juga ditolak jika akan mengeluarkan jadwal item yang sudah ada dari rentang trip.

Endpoint rekomendasi eksternal tetap hanya menyediakan pilihan kota/destinasi sesuai kontrak provider. HalalTrip menyimpan pilihan pengguna setelahnya sebagai itinerary; API tidak mengklaim membuat itinerary otomatis, rute optimal, estimasi jarak, maupun data transportasi.

## Status error

- Request JSON atau pilihan rekomendasi tidak valid: `400`.
- Provider melaporkan data tidak ditemukan: `404`.
- Timeout provider: `504`.
- Koneksi, status gagal provider, response kosong/tidak valid, atau response terlalu besar: `502`.
- Kesalahan internal: `500`.

Pesan error tidak menyertakan body, credential, atau detail internal provider. Bentuk error: `{"error":{"code":"provider_error","message":"Layanan rekomendasi sedang tidak tersedia."}}`.

## Konfigurasi dan testing

- `TRAVEL_RECOMMENDATION_API_URL`: base URL provider. Default `https://travel-recommendation-api-production-75f5.up.railway.app`.
- `HTTP_CLIENT_TIMEOUT`: timeout HTTP client; default `15s`.

Test memakai `httptest` mock server:

    go test ./...
