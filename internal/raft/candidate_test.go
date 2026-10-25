package raft

import "testing"

func TestStartElection(t *testing.T) {
	node := NewNodeState("1")

	args, err := node.StartElection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.Role != Candidate {
		t.Fatalf("expected candidate, got %q", node.Role)
	}

	if node.CurrentTerm != 1 {
		t.Fatalf("expected term 1, got %d", node.CurrentTerm)
	}

	if node.VotedFor != "1" {
		t.Fatalf("expected self-vote, got %q", node.VotedFor)
	}

	if args.CandidateID != "1" || args.Term != 1 {
		t.Fatalf("unexpected vote request: %+v", args)
	}
}
