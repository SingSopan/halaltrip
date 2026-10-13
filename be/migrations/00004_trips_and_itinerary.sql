-- +goose Up
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 160),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    origin TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'planning' CHECK (status IN ('planning','active','completed','cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date >= start_date)
);
CREATE INDEX trips_user_dates_idx ON trips (user_id, start_date DESC, created_at DESC);

CREATE TABLE itinerary_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
    description TEXT NOT NULL DEFAULT '',
    place_type TEXT NOT NULL CHECK (place_type IN ('destination','restaurant','mosque','hotel','custom_activity')),
    external_place_id TEXT,
    scheduled_date DATE,
    start_time TIME,
    end_time TIME,
    sequence INTEGER NOT NULL CHECK (sequence >= 1),
    transport_mode TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((start_time IS NULL) = (end_time IS NULL)),
    CHECK (start_time IS NULL OR end_time > start_time),
    UNIQUE (trip_id, sequence) DEFERRABLE INITIALLY DEFERRED
);
CREATE INDEX itinerary_items_trip_date_sequence_idx ON itinerary_items (trip_id, scheduled_date, sequence);

-- +goose Down
DROP TABLE itinerary_items;
DROP TABLE trips;
