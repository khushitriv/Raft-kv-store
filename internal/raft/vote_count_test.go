package raft

import "testing"

func TestCandidateBecomesLeaderWithMajority(t *testing.T) {
	node := NewNodeState("1")

	_, err := node.StartElection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node.RecordVote("2", true, 3)

	if node.Role != Leader {
		t.Fatalf("expected leader, got %q", node.Role)
	}
}

func TestCandidateDoesNotBecomeLeaderWithoutMajority(t *testing.T) {
	node := NewNodeState("1")

	_, err := node.StartElection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node.RecordVote("2", false, 3)

	if node.Role != Candidate {
		t.Fatalf("expected candidate, got %q", node.Role)
	}
}
func TestDuplicateVoteIsNotCountedTwice(t *testing.T) {
	node := NewNodeState("1")

	_, err := node.StartElection()
	if err != nil {
		t.Fatal(err)
	}

	node.RecordVote("2", true, 3)
	node.RecordVote("2", true, 3)

	if node.VotesReceived != 2 {
		t.Fatalf("expected 2 votes, got %d", node.VotesReceived)
	}
}
