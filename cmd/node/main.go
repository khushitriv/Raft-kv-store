package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"context"
	"raft-kv-store/internal/kv"
	"raft-kv-store/internal/raft"
)

func main() {
	nodeID := flag.String("id", "1", "Unique ID for this node")
	port := flag.Int("port", 8001, "Port for this node")
	flag.Parse()

	allNodes := map[string]string{
		"1": "localhost:8001",
		"2": "localhost:8002",
		"3": "localhost:8003",
	}

	peers := make(map[string]string)
	for id, address := range allNodes {
		if id != *nodeID {
			peers[id] = address
		}
	}

	config := raft.Config{
		ID:    *nodeID,
		Port:  *port,
		Peers: peers,
	}

	node := raft.NewNodeState(config.ID)
	go node.RunHeartbeatLoop(context.Background(), config.Peers)
	store := kv.NewStore()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Node %s is running\n", config.ID)
	})

	mux.Handle("/raft/request-vote", &raft.VoteHandler{
		Node: node,
	})

	mux.Handle("/raft/append-entries", &raft.AppendEntriesHandler{
		Node:  node,
		Store: store,
	})

	mux.HandleFunc("/raft/election", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := node.RunElection(config.Peers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(node.Snapshot())
	})

	mux.HandleFunc("/raft/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(node.Snapshot())
	})

	mux.HandleFunc("/kv/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key query parameter", http.StatusBadRequest)
			return
		}

		value, exists := store.Get(key)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"key":   key,
			"found": exists,
			"value": value,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("encoding key-value response: %v", err)
		}
	})

	mux.HandleFunc("/kv/set", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid JSON request", http.StatusBadRequest)
			return
		}

		if err := node.SetValue(request.Key, request.Value, config.Peers, store); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"key":     request.Key,
			"value":   request.Value,
		})
	})

	address := fmt.Sprintf(":%d", config.Port)
	log.Printf("Starting node %s on port %d", config.ID, config.Port)
	log.Printf("Known peers: %v", config.Peers)

	log.Fatal(http.ListenAndServe(address, mux))
}
