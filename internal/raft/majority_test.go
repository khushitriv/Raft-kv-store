package raft

import "testing"

func TestHasMajority(t *testing.T) {
	tests := []struct {
		name         string
		votes        int
		totalNodes   int
		wantMajority bool
	}{
		{"one of three", 1, 3, false},
		{"two of three", 2, 3, true},
		{"three of three", 3, 3, true},
		{"two of five", 2, 5, false},
		{"three of five", 3, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasMajority(tt.votes, tt.totalNodes)

			if got != tt.wantMajority {
				t.Fatalf(
					"HasMajority(%d, %d) = %v, want %v",
					tt.votes,
					tt.totalNodes,
					got,
					tt.wantMajority,
				)
			}
		})
	}
}
