package raft

import (
	"testing"

	"raft-kv-store/internal/kv"
)

func TestApplyCommittedEntries(t *testing.T) {
	node := NewNodeState("2")
	store := kv.NewStore()

	node.Log = append(node.Log, LogEntry{
		Term:  1,
		Index: 1,
		Key:   "color",
		Value: "blue",
	})

	if err := node.ApplyCommittedEntries(store); err != nil {
		t.Fatal(err)
	}

	if _, exists := store.Get("color"); exists {
		t.Fatal("uncommitted entry was applied")
	}

	node.CommitIndex = 1

	if err := node.ApplyCommittedEntries(store); err != nil {
		t.Fatal(err)
	}

	value, exists := store.Get("color")
	if !exists || value != "blue" {
		t.Fatalf("expected color=blue, got value=%q, exists=%v", value, exists)
	}

	if node.LastApplied != 1 {
		t.Fatalf("expected LastApplied=1, got %d", node.LastApplied)
	}
}
