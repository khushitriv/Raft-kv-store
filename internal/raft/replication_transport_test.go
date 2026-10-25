package raft

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReplicateLogOnceCommitsWithMajority(t *testing.T) {
	follower := NewNodeState("2")

	mux := http.NewServeMux()
	mux.Handle("/raft/append-entries", &AppendEntriesHandler{
		Node: follower,
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	leader := NewNodeState("1")
	leader.Role = Leader
	leader.CurrentTerm = 1
	leader.Log = []LogEntry{
		{
			Term:  1,
			Index: 1,
			Key:   "color",
			Value: "blue",
		},
	}

	peers := map[string]string{
		"2": server.Listener.Addr().String(),
		// this peer is notreachable
		"3": "127.0.0.1:1",
	}

	if err := leader.ReplicateLogOnce(peers); err != nil {
		t.Fatalf("replication returned an error: %v", err)
	}

	followerSnapshot := follower.Snapshot()
	if len(followerSnapshot.Log) != 1 {
		t.Fatalf("expected follower log length 1, got %d", len(followerSnapshot.Log))
	}

	if followerSnapshot.Log[0].Key != "color" ||
		followerSnapshot.Log[0].Value != "blue" {
		t.Fatalf("follower received unexpected entry: %+v", followerSnapshot.Log[0])
	}

	leaderSnapshot := leader.Snapshot()
	if leaderSnapshot.CommitIndex != 1 {
		t.Fatalf("expected leader commit index 1, got %d", leaderSnapshot.CommitIndex)
	}
}
