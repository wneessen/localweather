-- name: AddressByCoords :one
SELECT *
FROM addresses
WHERE latitude = ? AND longitude = ?;

-- name: NewAddress :one
INSERT INTO addresses
(latitude, longitude, altitude, accuracy, display_name, country, state, municipality, city_district, postcode, city, suburb, street, house_number)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;