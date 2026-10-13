-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_lower_unique_idx ON users (lower(email));
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    interests TEXT[] NOT NULL CHECK (cardinality(interests) BETWEEN 1 AND 6 AND interests <@ ARRAY['culinary','culture','nature','shopping','adventure','history']::TEXT[]),
    activity TEXT NOT NULL CHECK (activity IN ('Relaxing','Exploration','Adventure')),
    travel_type TEXT NOT NULL CHECK (travel_type IN ('Solo','Couple','Family','Friends')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose Down
DROP TABLE user_preferences;
DROP TABLE users;
