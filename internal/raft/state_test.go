package raft

import "testing"

func TestNewNodeStatsAsFollowe(t *testing.T) {
	node := NewNodeState("1")

	if node.ID != "1" {
		t.Fatalf("expected ID 1, got %q", node.ID)
	}

	if node.Role != Follower {
		t.Fatalf("expected follower, got %q", node.Role)
	}

	if node.CurrentTerm != 0 {
		t.Fatalf("expected term 0, got %d", node.CurrentTerm)
	}

	if node.VotedFor != "" {
		t.Fatalf("expected no vote, got %q", node.VotedFor)
	}
}
