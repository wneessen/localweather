-- +goose Up
CREATE TABLE locations
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    latitude   REAL    NOT NULL,
    longitude  REAL    NOT NULL,
    altitude   REAL    NULL,
    accuracy   REAL    NOT NULL,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE UNIQUE INDEX idx_locations_name ON locations (name);
CREATE UNIQUE INDEX idx_locations_coords ON locations (latitude, longitude);

-- +goose Down
DROP TABLE locations;
