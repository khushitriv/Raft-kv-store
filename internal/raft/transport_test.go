package raft

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestVoteFromPeer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/raft/request-vote" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			var args RequestVoteArgs
			if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
				t.Error(err)
				return
			}

			json.NewEncoder(w).Encode(RequestVoteReply{
				Term:        args.Term,
				VoteGranted: true,
			})
		},
	))
	defer server.Close()

	// remove "http://" to obtain the expected address
	address := server.Listener.Addr().String()

	reply, err := RequestVoteFromPeer(address, RequestVoteArgs{
		Term:        1,
		CandidateID: "1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reply.VoteGranted {
		t.Fatal("expected vote to be granted")
	}

	if reply.Term != 1 {
		t.Fatalf("expected term 1, got %d", reply.Term)
	}
}
