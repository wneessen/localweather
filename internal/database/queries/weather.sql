-- name: CurrentWeatherByAddressID :one
SELECT *
FROM current_weather
WHERE address_id = ?;

-- name: UpdateCurrentWeather :exec
INSERT INTO current_weather (address_id,
                             timestamp,
                             temperature,
                             apparent_temperature,
                             weather_code,
                             wind_speed,
                             wind_gusts,
                             wind_direction,
                             relative_humidity,
                             pressure_msl,
                             is_day,
                             updated_at)
VALUES (:address_id,
        :timestamp,
        :temperature,
        :apparent_temperature,
        :weather_code,
        :wind_speed,
        :wind_gusts,
        :wind_direction,
        :relative_humidity,
        :pressure_msl,
        :is_day,
        :updated_at)
ON CONFLICT (address_id) DO UPDATE SET timestamp            = excluded.timestamp,
                                       temperature          = excluded.temperature,
                                       apparent_temperature = excluded.apparent_temperature,
                                       weather_code         = excluded.weather_code,
                                       wind_speed           = excluded.wind_speed,
                                       wind_gusts           = excluded.wind_gusts,
                                       wind_direction       = excluded.wind_direction,
                                       relative_humidity    = excluded.relative_humidity,
                                       pressure_msl         = excluded.pressure_msl,
                                       is_day               = excluded.is_day,
                                       updated_at           = excluded.updated_at
WHERE excluded.timestamp > current_weather.timestamp;

