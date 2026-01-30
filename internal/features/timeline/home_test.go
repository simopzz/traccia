package timeline

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestHome_SmartDefaults(t *testing.T) {
	component := Home()
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render home: %v", err)
	}
	output := buf.String()

	// Check for Alpine.js x-data to handle date logic
	if !strings.Contains(output, "x-data") {
		t.Errorf("expected x-data in Home component for smart defaults")
	}

	// Check for start_date and end_date binding
	if !strings.Contains(output, "start_date") || !strings.Contains(output, "end_date") {
		t.Errorf("expected start_date and end_date binding in x-data")
	}

	// Check for logic to update end date (e.g., +7 days)
	// We expect some JS logic here.
	if !strings.Contains(output, "setDate") && !strings.Contains(output, "7") {
		t.Errorf("expected logic to set default end date (+7 days)")
	}
}
