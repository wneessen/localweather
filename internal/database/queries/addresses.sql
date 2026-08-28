-- name: AddressByCoords :one
SELECT *
FROM addresses
WHERE latitude = ?
  AND longitude = ?;

-- name: NewAddress :one
INSERT INTO addresses
(latitude, longitude, altitude, accuracy, display_name, country, state, municipality, city_district, postcode, city,
 suburb, street, house_number, provider)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateCurrentAddress :exec
INSERT INTO current_address (address_id)
VALUES (?)
ON CONFLICT (lock) DO UPDATE SET address_id = excluded.address_id,
                                 updated_at = unixepoch();

-- name: CurrentAddress :one
SELECT *
FROM current_address
         JOIN addresses ON addresses.id = current_address.address_id
LIMIT 1;

-- name: ClearCurrentAddress :exec
DELETE
FROM current_address;