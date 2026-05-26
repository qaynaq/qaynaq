package connection

import (
	"testing"
	"time"
)

func TestInBackoff(t *testing.T) {
	now := time.Now()
	recent := now.Add(-10 * time.Minute)
	old := now.Add(-2 * time.Hour)

	cases := []struct {
		name        string
		failures    int
		lastErrorAt *time.Time
		want        bool
	}{
		{"healthy connection", 0, nil, false},
		{"one failure", 1, &recent, false},
		{"below threshold", backoffThreshold - 1, &recent, false},
		{"at threshold within window", backoffThreshold, &recent, true},
		{"past threshold within window", backoffThreshold + 5, &recent, true},
		{"at threshold past window", backoffThreshold, &old, false},
		{"threshold reached but no timestamp", backoffThreshold, nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := inBackoff(tc.failures, tc.lastErrorAt); got != tc.want {
				t.Errorf("inBackoff(%d, %v) = %v, want %v", tc.failures, tc.lastErrorAt, got, tc.want)
			}
		})
	}
}
