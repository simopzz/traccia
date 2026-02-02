-- name: CreateEvent :one
INSERT INTO events (trip_id, title, category, location, latitude, longitude, start_time, end_time, pinned, position)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetEventByID :one
SELECT * FROM events WHERE id = $1;

-- name: ListEventsByTrip :many
SELECT * FROM events
WHERE trip_id = $1
ORDER BY position ASC, start_time ASC;

-- name: UpdateEvent :one
UPDATE events
SET title = $2, category = $3, location = $4, latitude = $5, longitude = $6,
    start_time = $7, end_time = $8, pinned = $9, position = $10, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = $1;

-- name: GetMaxPositionByTrip :one
SELECT COALESCE(MAX(position), -1)::int AS max_position FROM events WHERE trip_id = $1;

-- name: GetLastEventByTrip :one
SELECT * FROM events
WHERE trip_id = $1
ORDER BY end_time DESC
LIMIT 1;
