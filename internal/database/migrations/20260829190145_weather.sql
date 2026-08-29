-- +goose Up
CREATE TABLE current_weather
(
    address_id           INTEGER NOT NULL REFERENCES addresses (id) ON DELETE CASCADE,
    timestamp            INTEGER NOT NULL,
    temperature          REAL    NOT NULL,
    apparent_temperature REAL    NOT NULL,
    weather_code         INTEGER NOT NULL,
    wind_speed           REAL    NOT NULL,
    wind_gusts           REAL    NOT NULL,
    wind_direction       REAL NOT NULL,
    relative_humidity    REAL    NOT NULL,
    pressure_msl         REAL    NOT NULL,
    is_day               BOOLEAN NOT NULL,
    updated_at           INTEGER NOT NULL
);
CREATE UNIQUE INDEX idx_current_weather_address ON current_weather (address_id);


-- +goose Down
DROP INDEX idx_current_weather_address;
DROP TABLE current_weather;
