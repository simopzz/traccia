-- name: CreateFlightDetails :one
INSERT INTO flight_details (event_id, airline, flight_number, departure_airport, arrival_airport, departure_terminal, arrival_terminal, departure_gate, arrival_gate, booking_reference)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetFlightDetailsByEventID :one
SELECT * FROM flight_details WHERE event_id = $1;

-- name: GetFlightDetailsByEventIDs :many
SELECT * FROM flight_details WHERE event_id = ANY(@event_ids::int[]);

-- name: UpdateFlightDetails :one
UPDATE flight_details
SET airline = $2, flight_number = $3, departure_airport = $4, arrival_airport = $5,
    departure_terminal = $6, arrival_terminal = $7, departure_gate = $8, arrival_gate = $9,
    booking_reference = $10
WHERE event_id = $1
RETURNING *;
