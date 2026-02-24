-- name: CreateTransitDetails :one
INSERT INTO transit_details (event_id, origin, destination, transport_mode)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTransitDetailsByEventID :one
SELECT * FROM transit_details WHERE event_id = $1;

-- name: GetTransitDetailsByEventIDs :many
SELECT * FROM transit_details WHERE event_id = ANY(@event_ids::int[]);

-- name: UpdateTransitDetails :one
UPDATE transit_details
SET origin = $2, destination = $3, transport_mode = $4
WHERE event_id = $1
RETURNING *;
