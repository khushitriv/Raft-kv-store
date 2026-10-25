package raft

import "testing"

func TestAdvanceCommitRequiresMajority(t *testing.T) {
	node := NewNodeState("1")
	node.Role = Leader
	node.CurrentTerm = 1
	node.Log = []LogEntry{
		{Term: 1, Index: 1, Key: "color", Value: "blue"},
	}

	peers := map[string]string{
		"2": "localhost:8002",
		"3": "localhost:8003",
	}

	if node.AdvanceCommit(map[string]uint64{}, peers) {
		t.Fatal("commit advanced without a majority")
	}

	if node.CommitIndex != 0 {
		t.Fatalf("expected commit index 0, got %d", node.CommitIndex)
	}

	if !node.AdvanceCommit(map[string]uint64{"2": 1}, peers) {
		t.Fatal("expected commit to advance with a majority")
	}

	if node.CommitIndex != 1 {
		t.Fatalf("expected commit index 1, got %d", node.CommitIndex)
	}
}
