-- +goose Up
CREATE TABLE current_weather
(
    address_id           INTEGER NOT NULL REFERENCES addresses (id) ON DELETE CASCADE,
    timestamp            INTEGER,
    temperature          REAL,
    apparent_temperature REAL,
    weather_code         INTEGER,
    wind_speed           REAL,
    wind_gusts           REAL,
    wind_direction       REAL,
    relative_humidity    REAL,
    pressure_msl         REAL,
    temp_day_min         REAL,
    temp_day_max         REAL,
    uv_index             REAL,
    is_day               BOOLEAN,
    sunrise_utc          INTEGER NOT NULL,
    sunset_utc           INTEGER NOT NULL,
    condition            TEXT    NOT NULL,
    category             TEXT    NOT NULL,
    icon                 TEXT    NOT NULL,
    winddir_icon         TEXT    NOT NULL,
    winddir_text         TEXT    NOT NULL,
    temp_unit            TEXT    NOT NULL,
    windspeed_unit       TEXT    NOT NULL,
    humidity_unit        TEXT    NOT NULL,
    pressure_unit        TEXT    NOT NULL,
    winddir_unit         TEXT    NOT NULL,
    timezone             TEXT    NOT NULL,
    timezone_abbr        TEXT,
    base_unit            TEXT    NOT NULL,
    updated_at           INTEGER NOT NULL
);
CREATE UNIQUE INDEX idx_current_weather_address_unit ON current_weather (address_id, base_unit);


-- +goose Down
DROP INDEX idx_current_weather_address;
DROP TABLE current_weather;

