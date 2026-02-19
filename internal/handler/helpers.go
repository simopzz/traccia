package handler

import (
	"net/http"
	"time"
)

func formatDateInput(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatDateTimeInput(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02T15:04")
}

func parseDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseDateAndTime(dateStr, timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04", dateStr+" "+timeStr)
}

func getUserID(r *http.Request) *string {
	// TODO: Extract from Supabase JWT
	return nil // anonymous for now
}
