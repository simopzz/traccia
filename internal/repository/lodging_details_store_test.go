package repository

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/simopzz/traccia/internal/repository/sqlcgen"
)

func Test_lodgingRowToDomain(t *testing.T) {
	checkIn := time.Date(2026, 6, 1, 15, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 6, 5, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		row  *sqlcgen.LodgingDetail
		name string
	}{
		{
			name: "full lodging details",
			row: &sqlcgen.LodgingDetail{
				ID:               1,
				EventID:          2,
				CheckInTime:      pgtype.Timestamptz{Time: checkIn, Valid: true},
				CheckOutTime:     pgtype.Timestamptz{Time: checkOut, Valid: true},
				BookingReference: pgtype.Text{String: "REF123", Valid: true},
			},
		},
		{
			name: "nil lodging details",
			row: &sqlcgen.LodgingDetail{
				ID:               1,
				EventID:          2,
				CheckInTime:      pgtype.Timestamptz{Valid: false},
				CheckOutTime:     pgtype.Timestamptz{Valid: false},
				BookingReference: pgtype.Text{String: "", Valid: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lodgingRowToDomain(tt.row)

			if got.ID != int(tt.row.ID) {
				t.Errorf("ID = %d, want %d", got.ID, tt.row.ID)
			}
			if got.EventID != int(tt.row.EventID) {
				t.Errorf("EventID = %d, want %d", got.EventID, tt.row.EventID)
			}
			if tt.row.CheckInTime.Valid {
				if got.CheckInTime == nil || !got.CheckInTime.Equal(tt.row.CheckInTime.Time) {
					t.Errorf("CheckInTime mismatch: got %v, want %v", got.CheckInTime, tt.row.CheckInTime.Time)
				}
			} else if got.CheckInTime != nil {
				t.Error("CheckInTime should be nil")
			}
			if tt.row.CheckOutTime.Valid {
				if got.CheckOutTime == nil || !got.CheckOutTime.Equal(tt.row.CheckOutTime.Time) {
					t.Errorf("CheckOutTime mismatch: got %v, want %v", got.CheckOutTime, tt.row.CheckOutTime.Time)
				}
			} else if got.CheckOutTime != nil {
				t.Error("CheckOutTime should be nil")
			}
			if got.BookingReference != tt.row.BookingReference.String {
				t.Errorf("BookingReference = %q, want %q", got.BookingReference, tt.row.BookingReference.String)
			}
		})
	}
}
