package raft

import (
	"encoding/json"
	"net/http"

	"raft-kv-store/internal/kv"
)

type AppendEntriesHandler struct {
	Node  *NodeState
	Store *kv.Store
}

func (h *AppendEntriesHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var args AppendEntriesArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, "invalid AppendEntries request", http.StatusBadRequest)
		return
	}

	reply := h.Node.AppendEntries(args)

	// to apply entries only after AppendEntries succeeds
	if reply.Success && h.Store != nil {
		if err := h.Node.ApplyCommittedEntries(h.Store); err != nil {
			http.Error(w, "failed to apply committed entries", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reply); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
