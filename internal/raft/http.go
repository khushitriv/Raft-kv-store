package raft

import (
	"encoding/json"
	"net/http"
)

type VoteHandler struct {
	Node *NodeState
}

func (h *VoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var args RequestVoteArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, "invalid vote request", http.StatusBadRequest)
		return
	}

	reply := h.Node.RequestVote(args)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reply); err != nil {
		http.Error(w, "failed to encode vote response", http.StatusInternalServerError)
		return
	}
}
