-- +goose Up
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'));

CREATE TABLE restaurants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
    description TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL,
    country TEXT NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    cuisine TEXT NOT NULL DEFAULT '',
    price_range TEXT NOT NULL DEFAULT '',
    halal_status TEXT NOT NULL DEFAULT 'unknown' CHECK (halal_status IN ('certified','muslim_owned','halal_friendly','no_pork','unknown')),
    halal_verification_note TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((latitude IS NULL) = (longitude IS NULL) AND (latitude IS NULL OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180))),
    CHECK (halal_status <> 'certified' OR (length(btrim(halal_verification_note)) > 0 AND length(btrim(source_url)) > 0))
);
CREATE INDEX restaurants_location_idx ON restaurants (lower(city), lower(country));
CREATE INDEX restaurants_name_idx ON restaurants (lower(name));
CREATE INDEX restaurants_filters_idx ON restaurants (halal_status, price_range);

CREATE TABLE mosques (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
    description TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL,
    country TEXT NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    facilities JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(facilities) = 'array'),
    source_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((latitude IS NULL) = (longitude IS NULL) AND (latitude IS NULL OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)))
);
CREATE INDEX mosques_location_idx ON mosques (lower(city), lower(country));
CREATE INDEX mosques_name_idx ON mosques (lower(name));

CREATE TABLE hotels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
    description TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL,
    country TEXT NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    price_range TEXT NOT NULL DEFAULT '',
    muslim_friendly_features JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(muslim_friendly_features) = 'array'),
    source_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((latitude IS NULL) = (longitude IS NULL) AND (latitude IS NULL OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)))
);
CREATE INDEX hotels_location_idx ON hotels (lower(city), lower(country));
CREATE INDEX hotels_name_idx ON hotels (lower(name));
CREATE INDEX hotels_price_idx ON hotels (price_range);

CREATE TABLE saved_places (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    place_type TEXT NOT NULL CHECK (place_type IN ('restaurant','mosque','hotel')),
    place_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, place_type, place_id)
);

-- +goose Down
DROP TABLE saved_places;
DROP TABLE hotels;
DROP TABLE mosques;
DROP TABLE restaurants;
ALTER TABLE users DROP COLUMN role;
