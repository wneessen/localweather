-- +goose Up
CREATE TABLE addresses
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    latitude      REAL NOT NULL,
    longitude     REAL NOT NULL,
    altitude      REAL,
    accuracy      REAL NOT NULL,
    display_name  TEXT NOT NULL,
    country       TEXT,
    state         TEXT,
    municipality  TEXT,
    city_district TEXT,
    postcode      TEXT,
    city          TEXT,
    suburb        TEXT,
    street        TEXT,
    house_number  TEXT,
    provider      TEXT NOT NULL,
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
