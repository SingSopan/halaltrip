-- +goose Up
ALTER TABLE destinations ADD COLUMN latitude DOUBLE PRECISION;
ALTER TABLE destinations ADD COLUMN longitude DOUBLE PRECISION;
ALTER TABLE destinations ADD COLUMN source_url TEXT NOT NULL DEFAULT '';
ALTER TABLE destinations ADD COLUMN image_attribution TEXT NOT NULL DEFAULT '';
ALTER TABLE destinations ADD COLUMN image_license_url TEXT NOT NULL DEFAULT '';

ALTER TABLE restaurants ADD COLUMN image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE hotels ADD COLUMN image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE mosques ADD COLUMN image_url TEXT NOT NULL DEFAULT '';

INSERT INTO destinations (slug, name, country, city, description, image_url, latitude, longitude, source_url, image_attribution, image_license_url)
VALUES
    ('monumen-nasional-jakarta', 'Monumen Nasional (Monas)', 'Indonesia', 'Jakarta', 'Monumen nasional di Lapangan Merdeka, pusat Jakarta. Area monumen dan museum memberi konteks sejarah kemerdekaan Indonesia.', 'https://upload.wikimedia.org/wikipedia/commons/f/f3/National_Monument_Jakarta.jpg', -6.175382, 106.827153, 'https://sinar.big.go.id/Pencarian/Detail/163788', 'Michael Meer — public domain', 'https://commons.wikimedia.org/wiki/File:National_Monument_Jakarta.jpg'),
    ('kota-tua-jakarta', 'Kota Tua Jakarta', 'Indonesia', 'Jakarta', 'Kawasan kota lama Jakarta dengan bangunan bersejarah dan area Fatahillah. Waktu kunjungan dapat dipakai untuk museum dan berjalan di sekitar alun-alun.', 'https://upload.wikimedia.org/wikipedia/commons/thumb/3/30/Kota_tua_Jakarta.jpg/1280px-Kota_tua_Jakarta.jpg', -6.134327, 106.813582, 'https://commons.wikimedia.org/wiki/File:Kota_tua_Jakarta.jpg', 'Chitra arti — CC0 1.0', 'https://creativecommons.org/publicdomain/zero/1.0/'),
    ('masjid-istiqlal-jakarta', 'Masjid Istiqlal', 'Indonesia', 'Jakarta', 'Masjid negara di kawasan Pasar Baru. Direktori SIMAS Kementerian Agama mencatat alamat, koordinat, dan fasilitas masjid.', 'https://upload.wikimedia.org/wikipedia/commons/f/f8/Masjid_Istiqlal_Jakarta.jpg', -6.170150, 106.831103, 'https://simas.kemenag.go.id/profil/masjid/01.0.11.03.02.000001', 'AR El Syarif — CC BY-SA 4.0', 'https://creativecommons.org/licenses/by-sa/4.0/');

INSERT INTO restaurants (name, description, city, country, address, cuisine, price_range, halal_status, source_url)
VALUES ('Sate Khas Senayan — Menteng', 'Cabang restoran Indonesia di Menteng. Status sertifikasi halal tidak ditetapkan oleh seed ini; konfirmasi langsung ke restoran bila menjadi syarat perjalanan.', 'Jakarta', 'Indonesia', 'Jl. HOS Cokroaminoto No. 76/78, Menteng, Jakarta Pusat', 'Masakan Indonesia', 'mid', 'unknown', 'https://satekhas.sarirasa.co.id/location');

INSERT INTO hotels (name, description, city, country, address, price_range, muslim_friendly_features, source_url)
VALUES ('Quds Express Syariah Hotel', 'Hotel di Tanah Abang yang menyebut konsep halal hospitality pada situs resminya. Detail fasilitas kamar dan ketersediaan perlu dikonfirmasi langsung ke hotel.', 'Jakarta', 'Indonesia', 'Jl. K.H. Mas Mansyur No. 64, Tanah Abang, Jakarta Pusat', 'mid', '[{"name":"Konsep halal hospitality","verified":false}]'::jsonb, 'https://qudshoteljakarta.com/id-id/');

INSERT INTO mosques (name, description, city, country, address, latitude, longitude, facilities, source_url)
VALUES ('Masjid Istiqlal', 'Masjid negara di Pasar Baru. Daftar fasilitas berikut berasal dari profil SIMAS Kementerian Agama.', 'Jakarta', 'Indonesia', 'Jl. Taman Wijaya Kusuma, Pasar Baru, Sawah Besar, Jakarta Pusat', -6.170150, 106.831103, '["Tempat wudhu","Kamar mandi/WC","Akses disabilitas","Parkir","Perpustakaan","Museum Masjid"]'::jsonb, 'https://simas.kemenag.go.id/profil/masjid/01.0.11.03.02.000001');

-- +goose Down
DELETE FROM mosques WHERE name = 'Masjid Istiqlal' AND city = 'Jakarta' AND source_url = 'https://simas.kemenag.go.id/profil/masjid/01.0.11.03.02.000001';
DELETE FROM hotels WHERE name = 'Quds Express Syariah Hotel' AND city = 'Jakarta' AND source_url = 'https://qudshoteljakarta.com/id-id/';
DELETE FROM restaurants WHERE name = 'Sate Khas Senayan — Menteng' AND city = 'Jakarta' AND source_url = 'https://satekhas.sarirasa.co.id/location';
DELETE FROM destinations WHERE slug IN ('monumen-nasional-jakarta', 'kota-tua-jakarta', 'masjid-istiqlal-jakarta');
ALTER TABLE mosques DROP COLUMN image_url;
ALTER TABLE hotels DROP COLUMN image_url;
ALTER TABLE restaurants DROP COLUMN image_url;
ALTER TABLE destinations DROP COLUMN image_license_url;
ALTER TABLE destinations DROP COLUMN image_attribution;
ALTER TABLE destinations DROP COLUMN source_url;
ALTER TABLE destinations DROP COLUMN longitude;
ALTER TABLE destinations DROP COLUMN latitude;
