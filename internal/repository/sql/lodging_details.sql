-- name: CreateLodgingDetails :one
INSERT INTO lodging_details (event_id, check_in_time, check_out_time, booking_reference)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLodgingDetailsByEventID :one
SELECT * FROM lodging_details WHERE event_id = $1;

-- name: GetLodgingDetailsByEventIDs :many
SELECT * FROM lodging_details WHERE event_id = ANY(@event_ids::int[]);

-- name: UpdateLodgingDetails :one
UPDATE lodging_details
SET check_in_time = $2, check_out_time = $3, booking_reference = $4
WHERE event_id = $1
RETURNING *;
