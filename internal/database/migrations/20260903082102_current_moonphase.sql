-- +goose Up
ALTER TABLE current_weather
    ADD COLUMN moonphase TEXT;
ALTER TABLE current_weather
    ADD COLUMN moonphase_icon TEXT;
ALTER TABLE current_weather
    ADD COLUMN moonphase_icon_url TEXT;


-- +goose Down
ALTER TABLE current_weather
    DROP COLUMN moonphase_icon_url;
ALTER TABLE current_weather
    DROP COLUMN moonphase_icon;
ALTER TABLE current_weather
    DROP COLUMN moonphase;
