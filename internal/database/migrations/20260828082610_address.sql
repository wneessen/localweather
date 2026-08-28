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
    created_at    INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE UNIQUE INDEX idx_locations_name ON addresses (display_name);
CREATE UNIQUE INDEX idx_locations_coords ON addresses (latitude, longitude);

-- +goose Down
DROP TABLE addresses;
