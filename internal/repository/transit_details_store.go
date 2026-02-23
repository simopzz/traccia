package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/repository/sqlcgen"
)

type TransitDetailsStore struct{}

func NewTransitDetailsStore() *TransitDetailsStore {
	return &TransitDetailsStore{}
}

// Create inserts transit_details within the caller's transaction (q is tx-scoped).
func (s *TransitDetailsStore) Create(ctx context.Context, q *sqlcgen.Queries, eventID int, td *domain.TransitDetails) (*domain.TransitDetails, error) {
	row, err := q.CreateTransitDetails(ctx, sqlcgen.CreateTransitDetailsParams{
		EventID:       int32(eventID),
		Origin:        toPgText(td.Origin),
		Destination:   toPgText(td.Destination),
		TransportMode: toPgText(td.TransportMode),
	})
	if err != nil {
		return nil, fmt.Errorf("inserting transit_details for event %d: %w", eventID, err)
	}
	result := transitRowToDomain(&row)
	return &result, nil
}

// GetByEventID loads transit_details. Returns domain.ErrNotFound if the row doesn't exist.
func (s *TransitDetailsStore) GetByEventID(ctx context.Context, q *sqlcgen.Queries, eventID int) (*domain.TransitDetails, error) {
	row, err := q.GetTransitDetailsByEventID(ctx, int32(eventID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("fetching transit_details for event %d: %w", eventID, err)
	}
	result := transitRowToDomain(&row)
	return &result, nil
}

// GetByEventIDs fetches transit_details for multiple events in a single query.
func (s *TransitDetailsStore) GetByEventIDs(ctx context.Context, q *sqlcgen.Queries, eventIDs []int) (map[int]*domain.TransitDetails, error) {
	if len(eventIDs) == 0 {
		return nil, nil
	}
	ids := make([]int32, len(eventIDs))
	for i, id := range eventIDs {
		ids[i] = int32(id)
	}
	rows, err := q.GetTransitDetailsByEventIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("fetching transit_details by ids: %w", err)
	}
	results := make(map[int]*domain.TransitDetails)
	for i := range rows {
		td := transitRowToDomain(&rows[i])
		results[int(rows[i].EventID)] = &td
	}
	return results, nil
}

// Update updates existing transit_details. Uses the caller-provided queries (can be tx-scoped).
func (s *TransitDetailsStore) Update(ctx context.Context, q *sqlcgen.Queries, eventID int, td *domain.TransitDetails) (*domain.TransitDetails, error) {
	row, err := q.UpdateTransitDetails(ctx, sqlcgen.UpdateTransitDetailsParams{
		EventID:       int32(eventID),
		Origin:        toPgText(td.Origin),
		Destination:   toPgText(td.Destination),
		TransportMode: toPgText(td.TransportMode),
	})
	if err != nil {
		return nil, fmt.Errorf("updating transit_details for event %d: %w", eventID, err)
	}
	result := transitRowToDomain(&row)
	return &result, nil
}

func transitRowToDomain(row *sqlcgen.TransitDetail) domain.TransitDetails {
	return domain.TransitDetails{
		ID:            int(row.ID),
		EventID:       int(row.EventID),
		Origin:        row.Origin.String,
		Destination:   row.Destination.String,
		TransportMode: row.TransportMode.String,
	}
}
