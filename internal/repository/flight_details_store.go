package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/repository/sqlcgen"
)

type FlightDetailsStore struct{}

func NewFlightDetailsStore() *FlightDetailsStore {
	return &FlightDetailsStore{}
}

// Create inserts flight_details within the caller's transaction (q is tx-scoped).
func (s *FlightDetailsStore) Create(ctx context.Context, q *sqlcgen.Queries, eventID int, fd *domain.FlightDetails) (*domain.FlightDetails, error) {
	row, err := q.CreateFlightDetails(ctx, sqlcgen.CreateFlightDetailsParams{
		EventID:           int32(eventID),
		Airline:           toPgText(fd.Airline),
		FlightNumber:      toPgText(fd.FlightNumber),
		DepartureAirport:  toPgText(fd.DepartureAirport),
		ArrivalAirport:    toPgText(fd.ArrivalAirport),
		DepartureTerminal: toPgText(fd.DepartureTerminal),
		ArrivalTerminal:   toPgText(fd.ArrivalTerminal),
		DepartureGate:     toPgText(fd.DepartureGate),
		ArrivalGate:       toPgText(fd.ArrivalGate),
		BookingReference:  toPgText(fd.BookingReference),
	})
	if err != nil {
		return nil, fmt.Errorf("inserting flight_details for event %d: %w", eventID, err)
	}
	result := flightRowToDomain(&row)
	return &result, nil
}

// GetByEventID loads flight_details. Returns domain.ErrNotFound if the row doesn't exist.
func (s *FlightDetailsStore) GetByEventID(ctx context.Context, q *sqlcgen.Queries, eventID int) (*domain.FlightDetails, error) {
	row, err := q.GetFlightDetailsByEventID(ctx, int32(eventID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("fetching flight_details for event %d: %w", eventID, err)
	}
	result := flightRowToDomain(&row)
	return &result, nil
}

// Update updates existing flight_details. Uses the caller-provided queries (can be tx-scoped).
func (s *FlightDetailsStore) Update(ctx context.Context, q *sqlcgen.Queries, eventID int, fd *domain.FlightDetails) (*domain.FlightDetails, error) {
	row, err := q.UpdateFlightDetails(ctx, sqlcgen.UpdateFlightDetailsParams{
		EventID:           int32(eventID),
		Airline:           toPgText(fd.Airline),
		FlightNumber:      toPgText(fd.FlightNumber),
		DepartureAirport:  toPgText(fd.DepartureAirport),
		ArrivalAirport:    toPgText(fd.ArrivalAirport),
		DepartureTerminal: toPgText(fd.DepartureTerminal),
		ArrivalTerminal:   toPgText(fd.ArrivalTerminal),
		DepartureGate:     toPgText(fd.DepartureGate),
		ArrivalGate:       toPgText(fd.ArrivalGate),
		BookingReference:  toPgText(fd.BookingReference),
	})
	if err != nil {
		return nil, fmt.Errorf("updating flight_details for event %d: %w", eventID, err)
	}
	result := flightRowToDomain(&row)
	return &result, nil
}

func flightRowToDomain(row *sqlcgen.FlightDetail) domain.FlightDetails {
	return domain.FlightDetails{
		ID:                int(row.ID),
		EventID:           int(row.EventID),
		Airline:           row.Airline.String,
		FlightNumber:      row.FlightNumber.String,
		DepartureAirport:  row.DepartureAirport.String,
		ArrivalAirport:    row.ArrivalAirport.String,
		DepartureTerminal: row.DepartureTerminal.String,
		ArrivalTerminal:   row.ArrivalTerminal.String,
		DepartureGate:     row.DepartureGate.String,
		ArrivalGate:       row.ArrivalGate.String,
		BookingReference:  row.BookingReference.String,
	}
}
