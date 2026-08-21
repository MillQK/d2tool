package update

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name           string
		latest         string
		current        string
		wantComparison int
		wantError      bool
	}{
		{name: "newer patch", latest: "0.0.13", current: "0.0.12", wantComparison: 1},
		{name: "optional v prefix", latest: "v1.2.0", current: "1.1.9", wantComparison: 1},
		{name: "equal", latest: "1.2.3", current: "v1.2.3", wantComparison: 0},
		{name: "older", latest: "1.9.9", current: "2.0.0", wantComparison: -1},
		{name: "prerelease older", latest: "2.0.0-beta.1", current: "2.0.0", wantComparison: -1},
		{name: "invalid latest", latest: "Release 2", current: "1.0.0", wantError: true},
		{name: "invalid current", latest: "2.0.0", current: "dev", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareVersions(tt.latest, tt.current)
			if (err != nil) != tt.wantError {
				t.Fatalf("compareVersions() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && got != tt.wantComparison {
				t.Fatalf("compareVersions() = %d, want %d", got, tt.wantComparison)
			}
		})
	}
}
