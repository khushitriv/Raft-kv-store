package raft

import "testing"

func TestGrantVoteInCurrentTerm(t *testing.T) {
	node := NewNodeState("2")

	reply := node.RequestVote(RequestVoteArgs{
		Term:        1,
		CandidateID: "1",
	})

	if !reply.VoteGranted {
		t.Fatal("expected vote to be granted")
	}

	if node.VotedFor != "1" {
		t.Fatalf("expected vote for node 1, got %q", node.VotedFor)
	}
}

func TestRejectOlderTerm(t *testing.T) {
	node := NewNodeState("2")
	node.CurrentTerm = 2

	reply := node.RequestVote(RequestVoteArgs{
		Term:        1,
		CandidateID: "1",
	})

	if reply.VoteGranted {
		t.Fatal("expected old-term vote request to be rejected")
	}
}

func TestCannotVoteForTwoCandidatesInSameTerm(t *testing.T) {
	node := NewNodeState("2")

	first := node.RequestVote(RequestVoteArgs{
		Term:        1,
		CandidateID: "1",
	})
	second := node.RequestVote(RequestVoteArgs{
		Term:        1,
		CandidateID: "3",
	})

	if !first.VoteGranted {
		t.Fatal("expected first vote to be granted")
	}

	if second.VoteGranted {
		t.Fatal("expected second candidate to be rejected")
	}
}

func TestGrantVoteToCandidateWithMoreUpToDateLog(t *testing.T) {
	node := NewNodeState("2")
	node.Log = []LogEntry{
		{Term: 1, Index: 1, Key: "a", Value: "1"},
	}

	reply := node.RequestVote(RequestVoteArgs{
		Term:         2,
		CandidateID:  "1",
		LastLogIndex: 2,
		LastLogTerm:  2,
	})

	if !reply.VoteGranted {
		t.Fatal("expected vote for candidate with a more up-to-date log")
	}
}

func TestRejectVoteForCandidateWithOlderLogTerm(t *testing.T) {
	node := NewNodeState("2")
	node.Log = []LogEntry{
		{Term: 2, Index: 1, Key: "a", Value: "1"},
	}

	reply := node.RequestVote(RequestVoteArgs{
		Term:         3,
		CandidateID:  "1",
		LastLogIndex: 1,
		LastLogTerm:  1,
	})

	if reply.VoteGranted {
		t.Fatal("expected candidate with an older log term to be rejected")
	}
}

func TestRejectVoteForShorterLogWithSameTerm(t *testing.T) {
	node := NewNodeState("2")
	node.Log = []LogEntry{
		{Term: 1, Index: 1, Key: "a", Value: "1"},
		{Term: 2, Index: 2, Key: "b", Value: "2"},
	}

	reply := node.RequestVote(RequestVoteArgs{
		Term:         3,
		CandidateID:  "1",
		LastLogIndex: 1,
		LastLogTerm:  2,
	})

	if reply.VoteGranted {
		t.Fatal("expected candidate with a shorter log in the same term to be rejected")
	}
}
