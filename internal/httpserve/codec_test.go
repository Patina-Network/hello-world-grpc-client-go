package httpserve

import (
	"testing"
	"time"
)

func TestIsJSONContentType(t *testing.T) {
	for _, tc := range []struct {
		name   string
		header string
		want   bool
	}{
		{"JSON", "application/json", true},
		{"parameters", "application/json; charset=utf-8", true},
		{"structured suffix", "application/cloudevents+json", true},
		{"case insensitive", "Application/JSON", true},
		{"empty", "", false},
		{"wrong type", "text/json", false},
		{"wrong type with suffix", "text/example+json", false},
		{"wrong subtype", "application/jsonp", false},
		{"suffix not at end", "application/json+xml", false},
		{"missing subtype", "application/", false},
		{"malformed parameter", "application/json; charset", false},
		{"multiple types", "application/json, text/plain", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isJSONContentType(tc.header); got != tc.want {
				t.Errorf("isJSONContentType(%q) = %t, want %t", tc.header, got, tc.want)
			}
		})
	}
}

func TestTimestampsUseShortestSubsecondPrecision(t *testing.T) {
	for _, tc := range []struct {
		at   time.Time
		want string
	}{
		{time.Unix(0, 0), "1970-01-01T00:00:00Z"},
		{time.Date(2026, 1, 2, 3, 4, 5, 500_000_000, time.UTC), "2026-01-02T03:04:05.500Z"},
		{time.Date(2026, 1, 2, 3, 4, 5, 1_000, time.UTC), "2026-01-02T03:04:05.000001Z"},
		{time.Date(2026, 1, 2, 3, 4, 5, 1, time.UTC), "2026-01-02T03:04:05.000000001Z"},
		{time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("EST", -5*3600)), "2026-01-02T08:04:05Z"},
		{time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC), "+10000-01-01T00:00:00Z"},
		{time.Date(-1, 1, 1, 0, 0, 0, 0, time.UTC), "-0001-01-01T00:00:00Z"},
	} {
		if got := formatTimestamp(tc.at); got != tc.want {
			t.Errorf("formatTimestamp(%v) = %s, want %s", tc.at, got, tc.want)
		}
	}
}
