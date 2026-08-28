-- +goose Up
CREATE TABLE addresses
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    latitude      float64 NOT NULL,
    longitude     float64 NOT NULL,
    altitude      float64 NULL,
    accuracy      float64 NOT NULL,
    display_name  string  NOT NULL,
    country       string  NULL,
    state         string  NULL,
    municipality  string  NULL,
    city_district string  NULL,
    postcode      string  NULL,
    city          string  NULL,
    suburb        string  NULL,
    street        string  NULL,
    house_number  string  NULL,
    provider      string  NOT NULL,
    created_at    INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE UNIQUE INDEX idx_locations_name ON addresses (display_name);
CREATE UNIQUE INDEX idx_locations_coords ON addresses (latitude, longitude);

CREATE TABLE current_address
(
    lock       INTEGER NOT NULL DEFAULT 1 UNIQUE CHECK (lock = 1),
    address_id INTEGER NOT NULL REFERENCES addresses (id) ON DELETE CASCADE,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;


-- +goose Down
DROP TABLE current_address;
DROP TABLE addresses;
