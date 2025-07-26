package level1_tests

import (
	"testing"
)

func TestParseRemappedRowsResults(t *testing.T) {
	expectedBusIDs := []string{
		"0000:0f:00.0",
		"0000:2d:00.0",
		"0000:44:00.0",
	}

	tests := []struct {
		name         string
		output       string
		threshold    int
		wantFailures int
		wantFailed   []string
		wantMissing  []string
	}{
		{
			name: "no failures",
			output: `00000000:0f:00.0, 0
00000000:2d:00.0, 0
00000000:44:00.0, 0`,
			threshold:    0,
			wantFailures: 0,
			wantFailed:   []string{},
			wantMissing:  []string{},
		},
		{
			name: "one failure above threshold",
			output: `00000000:0f:00.0, 5
00000000:2d:00.0, 0
00000000:44:00.0, 0`,
			threshold:    0,
			wantFailures: 1,
			wantFailed:   []string{"0000:0f:00.0"},
			wantMissing:  []string{},
		},
		{
			name: "missing GPU",
			output: `00000000:0f:00.0, 0
00000000:2d:00.0, 0`,
			threshold:    0,
			wantFailures: 0,
			wantFailed:   []string{},
			wantMissing:  []string{"0000:44:00.0"},
		},
		{
			name: "case insensitive matching",
			output: `00000000:0F:00.0, 0
00000000:2D:00.0, 1
00000000:44:00.0, 0`,
			threshold:    0,
			wantFailures: 1,
			wantFailed:   []string{"0000:2d:00.0"},
			wantMissing:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failureCount, failedBusIDs, missingBusIDs, err := parseRemappedRowsResults(tt.output, expectedBusIDs, tt.threshold)
			
			if err != nil {
				t.Errorf("parseRemappedRowsResults() error = %v", err)
				return
			}
			
			if failureCount != tt.wantFailures {
				t.Errorf("parseRemappedRowsResults() failureCount = %v, want %v", failureCount, tt.wantFailures)
			}
			
			if len(failedBusIDs) != len(tt.wantFailed) {
				t.Errorf("parseRemappedRowsResults() failed count = %v, want %v", len(failedBusIDs), len(tt.wantFailed))
			}
			
			if len(missingBusIDs) != len(tt.wantMissing) {
				t.Errorf("parseRemappedRowsResults() missing count = %v, want %v", len(missingBusIDs), len(tt.wantMissing))
			}
			
			for _, expected := range tt.wantFailed {
				found := false
				for _, actual := range failedBusIDs {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("parseRemappedRowsResults() missing failed bus ID %v", expected)
				}
			}
		})
	}
}