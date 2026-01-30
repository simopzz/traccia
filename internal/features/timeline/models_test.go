package timeline_test

import (
	"testing"
	"traccia/internal/features/timeline"
)

func TestEventStructFields(t *testing.T) {
	lat := 1.0
	lng := 2.0
	cat := "Activity"
	e := timeline.Event{
		Title:    "Test",
		Category: &cat,
		GeoLat:   &lat,
		GeoLng:   &lng,
		IsPinned: true,
	}
	if *e.Category != "Activity" {
		t.Error("Category mismatch")
	}
	if !e.IsPinned {
		t.Error("IsPinned mismatch")
	}
}
