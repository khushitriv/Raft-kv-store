package raft

import "testing"

func TestAppendEntriesAddsLogEntry(t *testing.T) {
	node := NewNodeState("2")

	args := AppendEntriesArgs{
		Term:     1,
		LeaderID: "1",
		Entries: []LogEntry{
			{
				Term:  1,
				Index: 1,
				Key:   "color",
				Value: "blue",
			},
		},
	}

	reply := node.AppendEntries(args)

	if !reply.Success {
		t.Fatal("expected AppendEntries to succeed")
	}

	if len(node.Log) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(node.Log))
	}

	if node.Log[0].Key != "color" || node.Log[0].Value != "blue" {
		t.Fatalf("unexpected log entry: %+v", node.Log[0])
	}

	if node.Role != Follower {
		t.Fatalf("expected follower role, got %s", node.Role)
	}
}

func TestAppendEntriesRejectsOlderTerm(t *testing.T) {
	node := NewNodeState("2")
	node.CurrentTerm = 2

	reply := node.AppendEntries(AppendEntriesArgs{
		Term:     1,
		LeaderID: "1",
	})

	if reply.Success {
		t.Fatal("expected older-term AppendEntries to be rejected")
	}
}

func TestAppendEntriesRejectsIncorrectEntryIndex(t *testing.T) {
	node := NewNodeState("2")

	reply := node.AppendEntries(AppendEntriesArgs{
		Term:     1,
		LeaderID: "1",
		Entries: []LogEntry{
			{
				Term:  1,
				Index: 2,
				Key:   "color",
				Value: "blue",
			},
		},
	})

	if reply.Success {
		t.Fatal("expected entry with an incorrect index to be rejected")
	}

	if len(node.Log) != 0 {
		t.Fatalf("expected log to remain empty, got %d entries", len(node.Log))
	}
}

func TestAppendEntriesRejectsMismatchedPreviousLogTerm(t *testing.T) {
	node := NewNodeState("2")
	node.Log = []LogEntry{
		{Term: 1, Index: 1, Key: "color", Value: "blue"},
	}

	reply := node.AppendEntries(AppendEntriesArgs{
		Term:         2,
		LeaderID:     "1",
		PrevLogIndex: 1,
		PrevLogTerm:  2,
		Entries: []LogEntry{
			{Term: 2, Index: 2, Key: "color", Value: "red"},
		},
	})

	if reply.Success {
		t.Fatal("expected mismatched previous log term to be rejected")
	}

	if len(node.Log) != 1 || node.Log[0].Value != "blue" {
		t.Fatal("expected existing log to remain unchanged")
	}
}

func TestAppendEntriesReplacesConflictingUncommittedEntry(t *testing.T) {
	node := NewNodeState("2")
	node.Log = []LogEntry{
		{Term: 1, Index: 1, Key: "color", Value: "blue"},
		{Term: 1, Index: 2, Key: "shape", Value: "circle"},
	}

	reply := node.AppendEntries(AppendEntriesArgs{
		Term:         2,
		LeaderID:     "1",
		PrevLogIndex: 1,
		PrevLogTerm:  1,
		Entries: []LogEntry{
			{Term: 2, Index: 2, Key: "shape", Value: "square"},
		},
	})

	if !reply.Success {
		t.Fatal("expected conflicting uncommitted entry to be replaced")
	}

	if len(node.Log) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(node.Log))
	}

	if node.Log[1].Term != 2 || node.Log[1].Value != "square" {
		t.Fatalf("expected replacement entry, got %+v", node.Log[1])
	}
}

func TestAppendEntriesDoesNotOverwriteCommittedEntry(t *testing.T) {
	node := NewNodeState("2")
	node.Log = []LogEntry{
		{Term: 1, Index: 1, Key: "color", Value: "blue"},
	}
	node.CommitIndex = 1

	reply := node.AppendEntries(AppendEntriesArgs{
		Term:         2,
		LeaderID:     "1",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries: []LogEntry{
			{Term: 2, Index: 1, Key: "color", Value: "red"},
		},
	})

	if reply.Success {
		t.Fatal("expected overwrite of committed entry to be rejected")
	}

	if len(node.Log) != 1 || node.Log[0].Value != "blue" {
		t.Fatal("committed entry should remain unchanged")
	}
}
