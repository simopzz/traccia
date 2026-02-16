-- name: CreateEvent :one
INSERT INTO events (trip_id, event_date, title, category, location, latitude, longitude, start_time, end_time, pinned, position, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetEventByID :one
SELECT * FROM events WHERE id = $1;

-- name: ListEventsByTrip :many
SELECT * FROM events
WHERE trip_id = $1
ORDER BY event_date ASC, position ASC;

-- name: ListEventsByTripAndDate :many
SELECT * FROM events
WHERE trip_id = $1 AND event_date = $2
ORDER BY position ASC;

-- name: UpdateEvent :one
UPDATE events
SET title = $2, category = $3, location = $4, latitude = $5, longitude = $6,
    start_time = $7, end_time = $8, pinned = $9, position = $10,
    event_date = $11, notes = $12, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEvent :execrows
DELETE FROM events WHERE id = $1;

-- name: GetMaxPositionByTripAndDate :one
SELECT COALESCE(MAX(position), 0)::int AS max_position
FROM events
WHERE trip_id = $1 AND event_date = $2;

-- name: GetLastEventByTrip :one
SELECT * FROM events
WHERE trip_id = $1
ORDER BY event_date DESC, end_time DESC
LIMIT 1;

-- name: CountEventsByTrip :one
SELECT COUNT(*)::int AS event_count FROM events WHERE trip_id = $1;

-- name: CountEventsByTripGroupedByDate :many
SELECT event_date, COUNT(*)::int AS event_count
FROM events
WHERE trip_id = $1
  AND (event_date < $2 OR event_date > $3)
GROUP BY event_date
ORDER BY event_date;
