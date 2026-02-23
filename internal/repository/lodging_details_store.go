package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/repository/sqlcgen"
)

type LodgingDetailsStore struct{}

func NewLodgingDetailsStore() *LodgingDetailsStore {
	return &LodgingDetailsStore{}
}

// Create inserts lodging_details within the caller's transaction (q is tx-scoped).
func (s *LodgingDetailsStore) Create(ctx context.Context, q *sqlcgen.Queries, eventID int, ld *domain.LodgingDetails) (*domain.LodgingDetails, error) {
	row, err := q.CreateLodgingDetails(ctx, sqlcgen.CreateLodgingDetailsParams{
		EventID:          int32(eventID),
		CheckInTime:      toOptionalPgTimestamptz(ld.CheckInTime),
		CheckOutTime:     toOptionalPgTimestamptz(ld.CheckOutTime),
		BookingReference: toPgText(ld.BookingReference),
	})
	if err != nil {
		return nil, fmt.Errorf("inserting lodging_details for event %d: %w", eventID, err)
	}
	result := lodgingRowToDomain(&row)
	return &result, nil
}

// GetByEventID loads lodging_details. Returns domain.ErrNotFound if the row doesn't exist.
func (s *LodgingDetailsStore) GetByEventID(ctx context.Context, q *sqlcgen.Queries, eventID int) (*domain.LodgingDetails, error) {
	row, err := q.GetLodgingDetailsByEventID(ctx, int32(eventID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("fetching lodging_details for event %d: %w", eventID, err)
	}
	result := lodgingRowToDomain(&row)
	return &result, nil
}

// GetByEventIDs fetches lodging_details for multiple events in a single query.
func (s *LodgingDetailsStore) GetByEventIDs(ctx context.Context, q *sqlcgen.Queries, eventIDs []int) (map[int]*domain.LodgingDetails, error) {
	if len(eventIDs) == 0 {
		return nil, nil
	}
	ids := make([]int32, len(eventIDs))
	for i, id := range eventIDs {
		ids[i] = int32(id)
	}
	rows, err := q.GetLodgingDetailsByEventIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("fetching lodging_details by ids: %w", err)
	}
	results := make(map[int]*domain.LodgingDetails)
	for i := range rows {
		ld := lodgingRowToDomain(&rows[i])
		results[int(rows[i].EventID)] = &ld
	}
	return results, nil
}

// Update updates existing lodging_details. Uses the caller-provided queries (can be tx-scoped).
func (s *LodgingDetailsStore) Update(ctx context.Context, q *sqlcgen.Queries, eventID int, ld *domain.LodgingDetails) (*domain.LodgingDetails, error) {
	row, err := q.UpdateLodgingDetails(ctx, sqlcgen.UpdateLodgingDetailsParams{
		EventID:          int32(eventID),
		CheckInTime:      toOptionalPgTimestamptz(ld.CheckInTime),
		CheckOutTime:     toOptionalPgTimestamptz(ld.CheckOutTime),
		BookingReference: toPgText(ld.BookingReference),
	})
	if err != nil {
		return nil, fmt.Errorf("updating lodging_details for event %d: %w", eventID, err)
	}
	result := lodgingRowToDomain(&row)
	return &result, nil
}

func lodgingRowToDomain(row *sqlcgen.LodgingDetail) domain.LodgingDetails {
	return domain.LodgingDetails{
		ID:               int(row.ID),
		EventID:          int(row.EventID),
		CheckInTime:      fromPgTimestamptz(row.CheckInTime),
		CheckOutTime:     fromPgTimestamptz(row.CheckOutTime),
		BookingReference: row.BookingReference.String,
	}
}
