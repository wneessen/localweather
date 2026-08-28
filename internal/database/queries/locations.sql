-- name: LocationByID :one
SELECT *
FROM locations
WHERE id = ?;