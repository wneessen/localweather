-- name: AddressByCoords :one
SELECT *
FROM addresses
WHERE lat_trunc = ?
  AND lon_trunc = ?
  AND locale = ?;

-- name: NewAddress :one
INSERT INTO addresses
(latitude, longitude, lat_trunc, lon_trunc, altitude, accuracy, display_name, country, state, municipality,
 city_district, postcode, city,
 suburb, street, house_number, provider, locale, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateCurrentAddress :exec
INSERT INTO current_address (address_id, created_at, updated_at)
VALUES (?, ?, ?)
ON CONFLICT (lock) DO UPDATE SET
                                 address_id = excluded.address_id,
                                 updated_at = excluded.updated_at;

-- name: CurrentAddress :one
SELECT *
FROM current_address
         JOIN addresses ON addresses.id = current_address.address_id
LIMIT 1;

-- name: ClearCurrentAddress :exec
DELETE
FROM current_address;