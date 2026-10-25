package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func SendAppendEntries(
	address string,
	args AppendEntriesArgs,
) (AppendEntriesReply, error) {
	var reply AppendEntriesReply

	body, err := json.Marshal(args)
	if err != nil {
		return reply, err
	}

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Post(
		"http://"+address+"/raft/append-entries",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return reply, fmt.Errorf("sending append entries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return reply, fmt.Errorf(
			"peer returned HTTP status %d",
			resp.StatusCode,
		)
	}

	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return reply, fmt.Errorf(
			"decoding append entries response: %w",
			err,
		)
	}

	return reply, nil
}

func (n *NodeState) ReplicateLogOnce(
	peers map[string]string,
) error {
	n.mu.Lock()

	if n.Role != Leader {
		n.mu.Unlock()
		return fmt.Errorf("only a leader can replicate its log")
	}

	term := n.CurrentTerm
	leaderID := n.ID
	leaderCommit := n.CommitIndex

	if n.NextIndex == nil {
		n.NextIndex = make(map[string]uint64)
	}
	if n.MatchIndex == nil {
		n.MatchIndex = make(map[string]uint64)
	}

	for peerID := range peers {
		if _, exists := n.NextIndex[peerID]; !exists {
			n.NextIndex[peerID] = 1
		}
		if _, exists := n.MatchIndex[peerID]; !exists {
			n.MatchIndex[peerID] = 0
		}
	}

	n.mu.Unlock()

	replicatedBy := make(map[string]uint64)

	for peerID, address := range peers {
		n.mu.Lock()

		if n.Role != Leader || n.CurrentTerm != term {
			n.mu.Unlock()
			return fmt.Errorf("leadership changed during replication")
		}

		nextIndex := n.NextIndex[peerID]
		lastIndex := uint64(len(n.Log))

		// Keeping nextIndex within the valid range of [1, lastIndex+1]
		if nextIndex < 1 {
			nextIndex = 1
		}
		if nextIndex > lastIndex+1 {
			nextIndex = lastIndex + 1
		}

		prevLogIndex := nextIndex - 1
		var prevLogTerm uint64

		if prevLogIndex > 0 {
			prevLogTerm = n.Log[prevLogIndex-1].Term
		}

		// to copy only entries the follower still needs
		entries := append(
			[]LogEntry(nil),
			n.Log[nextIndex-1:]...,
		)

		args := AppendEntriesArgs{
			Term:         term,
			LeaderID:     leaderID,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      entries,
			LeaderCommit: leaderCommit,
		}

		n.mu.Unlock()

		reply, err := SendAppendEntries(address, args)
		if err != nil {
			log.Printf("replication to peer %s failed: %v", peerID, err)
			continue
		}

		log.Printf(
			"peer %s replied: term=%d success=%v prevLogIndex=%d entries=%d",
			peerID,
			reply.Term,
			reply.Success,
			args.PrevLogIndex,
			len(args.Entries),
		)

		n.mu.Lock()

		if reply.Term > n.CurrentTerm {
			n.CurrentTerm = reply.Term
			n.VotedFor = ""
			n.Role = Follower
			n.VotesReceived = 0
			n.VotesFrom = make(map[string]bool)
			n.mu.Unlock()

			return fmt.Errorf(
				"peer %s has newer term %d; stepping down",
				peerID,
				reply.Term,
			)
		}

		if n.Role != Leader || n.CurrentTerm != term {
			n.mu.Unlock()
			return fmt.Errorf("leadership changed during replication")
		}

		if reply.Success && reply.Term == term {
			matchIndex := prevLogIndex + uint64(len(entries))

			n.MatchIndex[peerID] = matchIndex
			n.NextIndex[peerID] = matchIndex + 1
			replicatedBy[peerID] = matchIndex
		} else if reply.Term == term {
			// the follower rejected the previous-log position
			// Move backward and retry on a another replication
			if n.NextIndex[peerID] > 1 {
				n.NextIndex[peerID]--
			}
		}

		n.mu.Unlock()
	}

	n.AdvanceCommit(replicatedBy, peers)

	return nil
}
