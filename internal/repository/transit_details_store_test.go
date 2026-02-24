package repository

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/simopzz/traccia/internal/repository/sqlcgen"
)

func Test_transitRowToDomain(t *testing.T) {
	tests := []struct {
		row  *sqlcgen.TransitDetail
		name string
	}{
		{
			name: "full transit details",
			row: &sqlcgen.TransitDetail{
				ID:            1,
				EventID:       2,
				Origin:        pgtype.Text{String: "Shibuya Station", Valid: true},
				Destination:   pgtype.Text{String: "Asakusa Station", Valid: true},
				TransportMode: pgtype.Text{String: "Metro", Valid: true},
			},
		},
		{
			name: "empty transit details",
			row: &sqlcgen.TransitDetail{
				ID:            3,
				EventID:       4,
				Origin:        pgtype.Text{String: "", Valid: true},
				Destination:   pgtype.Text{String: "", Valid: true},
				TransportMode: pgtype.Text{String: "", Valid: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transitRowToDomain(tt.row)

			if got.ID != int(tt.row.ID) {
				t.Errorf("ID = %d, want %d", got.ID, tt.row.ID)
			}
			if got.EventID != int(tt.row.EventID) {
				t.Errorf("EventID = %d, want %d", got.EventID, tt.row.EventID)
			}
			if got.Origin != tt.row.Origin.String {
				t.Errorf("Origin = %q, want %q", got.Origin, tt.row.Origin.String)
			}
			if got.Destination != tt.row.Destination.String {
				t.Errorf("Destination = %q, want %q", got.Destination, tt.row.Destination.String)
			}
			if got.TransportMode != tt.row.TransportMode.String {
				t.Errorf("TransportMode = %q, want %q", got.TransportMode, tt.row.TransportMode.String)
			}
		})
	}
}
