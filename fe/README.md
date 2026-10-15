# HalalTrip frontend

Frontend Next.js ini menggunakan backend Go HalalTrip sebagai sumber data untuk katalog lokal, akun, preferensi, saved places, trips, dan itinerary. Rekomendasi kota/destinasi dan katalog destinasi provider hanya diakses lewat proxy backend. Frontend tidak mengisi data contoh ketika request gagal.

## Menjalankan secara lokal

Jalankan PostgreSQL terpasang secara lokal. Di terminal pertama, dari direktori `be`:

    Copy-Item .env.example .env
    $env:DATABASE_URL = "postgres://postgres:PASSWORD_ANDA@localhost:5432/halaltrip?sslmode=disable"
    $env:JWT_SECRET = "GANTI_DENGAN_RANDOM_SECRET_MINIMAL_32_KARAKTER"
    go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
    goose -dir migrations postgres "$env:DATABASE_URL" up
    go run ./cmd/api

Buat database `halaltrip` sebelum menjalankan migration. Backend menerima origin frontend lokal melalui `CORS_ALLOWED_ORIGINS=http://localhost:3000`. Jika frontend berjalan pada origin lain, ubah environment backend lalu restart.

Di terminal kedua, dari direktori `fe`:

    npm ci
    npm run dev

Buka `http://localhost:3000`. Nilai lokal `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1` tersedia di `.env.local` dan `.env.example`. Ubah nilainya lalu restart Next.js jika backend berada di host/port berbeda. Untuk production, atur variable yang sama ke URL backend HTTPS publik dan atur `CORS_ALLOWED_ORIGINS` backend ke origin frontend production yang tepat.

`NEXT_PUBLIC_*` hanya berisi URL publik API. Jangan menaruh `JWT_SECRET`, password database, atau secret provider di sana.

## Integrasi yang aktif

| Fitur | Endpoint backend | Catatan |
| --- | --- | --- |
| Register dan login | `POST /auth/register`, `POST /auth/login` | Memakai respons `user` dan `access_token` backend. |
| Profil dan preferensi | `GET/PATCH /users/me`, `GET/PATCH /users/me/preferences` | Preferensi memakai enum backend: `culinary`/`culture`/`nature`/`shopping`/`adventure`/`history`, activity dan travel type dengan kapitalisasi sesuai kontrak. |
| Rekomendasi kota/destinasi | `POST /recommendations/cities`, `POST /recommendations/destinations` | Request dikirim ke backend; backend yang meneruskan ke provider. Tidak ada request provider langsung dari browser. |
| Destination provider | `GET /destinations`, `GET /destinations/:id` | Format provider opaque. UI membaca field yang dikenali; field skor hanya ditampilkan jika benar-benar ada pada payload. |
| Direktori lokal | `GET /restaurants`, `/mosques`, `/hotels` serta `/:id` | Filter kota, pencarian dan filter spesifik diteruskan sesuai dukungan endpoint backend. |
| Saved places | `GET/POST /saved-places`, `DELETE /saved-places/:place_type/:place_id` | Terbatas ke restoran, masjid, dan hotel lokal sesuai backend; destination provider tidak dapat disimpan melalui endpoint ini. |
| Perjalanan | `GET/POST /trips`, `GET/PATCH/DELETE /trips/:tripId` | Memerlukan token; backend membatasi data menurut pemilik trip. |
| Itinerary | `GET/POST /trips/:tripId/itinerary-items`, `PATCH/DELETE .../:itemId`, `PUT .../order` | Mendukung tambah, edit, hapus, dan ubah urutan item. |

Katalog mempunyai loading, error/retry, dan empty state. Status halal ditampilkan sesuai field backend; badge **Bersertifikat halal** hanya muncul bila `halal_status` bernilai `certified`. Data provider tanpa ID stabil dapat ditampilkan, tetapi tidak mendapat link detail atau fungsi simpan.

## Sesi autentikasi

Backend saat ini mengembalikan access token JWT di JSON dan belum menyediakan cookie `HttpOnly`. Frontend menyimpan token hanya di memori aplikasi; token tidak ditulis ke `localStorage`, `sessionStorage`, atau log. Akibatnya, reload penuh mengakhiri sesi dan pengguna harus login kembali. Respons `401` dari request terproteksi membersihkan cache/sesi dan mengarahkan ke halaman login.

## Fitur yang belum terhubung / batas API

- CRUD direktori `POST/PUT/DELETE` tersedia di backend untuk admin, tetapi belum ada UI administrasi.
- Rekomendasi mengikuti kontrak di [README provider](https://github.com/blasphemous20/travel-recommendation-api): `/recommend/cities` mengirim daftar `{ city, country, score }`, dan `/recommend/destinations` mengirim destinasi yang dapat berisi `name`, `city`, `country`, `rating`, `review_count`, dan `score`. Planner memilih kota dahulu dari dropdown katalog/populer, lalu mengirim preferensi beserta kota tersebut ke rekomendasi destinasi; hasilnya digabung dengan katalog lokal untuk membuat draf itinerary. Field tambahan yang tidak dikenal diabaikan.
- Saved places backend tidak mencakup destination provider.
- Backend tidak menyediakan peta, geocoding, atau kalkulasi rute. Tampilan peta dekoratif yang belum terhubung ke data lokasi tidak digunakan sebagai sumber hasil.
- Backend belum mendukung cookie sesi; untuk sesi persisten yang lebih aman diperlukan perubahan backend terencana untuk cookie `HttpOnly`, `Secure` di production, `SameSite`, dan dukungan credentialed CORS.

## Verifikasi frontend

    npm run test
    npm run lint
    npm run typecheck
    npm run build

Test API client menggunakan mock `fetch` hanya di test dan mencakup respons JSON/204, error HTTP (400, 401, 404, 500, 502, 504), respons JSON invalid, network error, dan timeout. Aplikasi saat dijalankan tetap menggunakan backend yang dikonfigurasi.

## Troubleshooting

- Pesan `NEXT_PUBLIC_API_BASE_URL belum dikonfigurasi`: isi `.env.local` dengan URL yang memuat prefix `/api/v1`, lalu restart Next.js.
- Backend tidak dapat dijangkau: pastikan PostgreSQL, migration, dan proses Go berjalan serta port pada URL benar.
- Error CORS di browser: tambahkan origin frontend persis, termasuk scheme dan port, ke `CORS_ALLOWED_ORIGINS` backend. Jangan gunakan wildcard.
- `401`: login lagi; token hanya hidup selama halaman frontend tidak dimuat ulang.
- `502`/`504` untuk rekomendasi/destinasi: provider eksternal sedang gagal atau timeout; frontend menampilkan error dan menyediakan retry.
