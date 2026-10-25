package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func RequestVoteFromPeer(
	address string,
	args RequestVoteArgs,
) (RequestVoteReply, error) {
	var reply RequestVoteReply

	body, err := json.Marshal(args)
	if err != nil {
		return reply, err
	}

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Post(
		"http://"+address+"/raft/request-vote",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return reply, fmt.Errorf("sending vote request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return reply, fmt.Errorf(
			"peer returned HTTP status %d",
			resp.StatusCode,
		)
	}

	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return reply, fmt.Errorf("decoding vote response: %w", err)
	}

	return reply, nil
}

func (n *NodeState) RunElection(peers map[string]string) error {
	args, err := n.StartElection()
	if err != nil {
		return err
	}

	totalNodes := len(peers) + 1

	// a single-node cluster can elect itself
	if HasMajority(1, totalNodes) {
		n.mu.Lock()
		if n.Role == Candidate && n.CurrentTerm == args.Term {
			n.Role = Leader
		}
		n.mu.Unlock()
		return nil
	}

	// Ask each peer for a vote
	for peerID, address := range peers {
		reply, err := RequestVoteFromPeer(address, args)
		if err != nil {
			// unreachable peer cannot contribute in voting
			continue
		}

		n.RecordVoteResponse(peerID, args.Term, reply, totalNodes)

		n.mu.Lock()
		role := n.Role
		n.mu.Unlock()

		if role == Leader {
			return nil
		}
		if role != Candidate {
			return fmt.Errorf("election ended without becoming leader")
		}
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Role == Leader {
		return nil
	}

	return fmt.Errorf(
		"election in term %d failed to obtain a majority (%d/%d votes)",
		args.Term,
		n.VotesReceived,
		totalNodes,
	)
}
