package raft

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVoteHandlerGrantsVote(t *testing.T) {
	node := NewNodeState("2")
	handler := &VoteHandler{Node: node}

	args := RequestVoteArgs{
		Term:        1,
		CandidateID: "1",
	}

	body, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/raft/request-vote",
		bytes.NewReader(body),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var reply RequestVoteReply
	if err := json.NewDecoder(recorder.Body).Decode(&reply); err != nil {
		t.Fatal(err)
	}

	if !reply.VoteGranted {
		t.Fatal("expected vote to be granted")
	}

	if node.VotedFor != "1" {
		t.Fatalf("expected vote for node 1, got %q", node.VotedFor)
	}
}
