-- name: CreateTrip :one
INSERT INTO trips (name, destination, start_date, end_date, user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTripByID :one
SELECT * FROM trips WHERE id = $1;

-- name: ListTrips :many
SELECT * FROM trips
WHERE (user_id = $1 OR $1 IS NULL)
ORDER BY start_date DESC, created_at DESC;

-- name: UpdateTrip :one
UPDATE trips
SET name = $2, destination = $3, start_date = $4, end_date = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTrip :exec
DELETE FROM trips WHERE id = $1;

-- name: CountEventsByTripAndDateRange :one
SELECT COUNT(*)::int AS event_count
FROM events
WHERE trip_id = $1
  AND (event_date < $2 OR event_date > $3);
