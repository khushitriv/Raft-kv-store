package raft

import (
	"fmt"

	"raft-kv-store/internal/kv"
)

func (n *NodeState) SetValue(
	key string,
	value string,
	peers map[string]string,
	store *kv.Store,
) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	n.mu.Lock()

	if n.Role != Leader {
		n.mu.Unlock()
		return fmt.Errorf("this node is not the leader")
	}

	entry := LogEntry{
		Term:  n.CurrentTerm,
		Index: uint64(len(n.Log) + 1),
		Key:   key,
		Value: value,
	}

	n.Log = append(n.Log, entry)
	entryIndex := entry.Index

	n.mu.Unlock()

	if err := n.ReplicateLogOnce(peers); err != nil {
		return fmt.Errorf("replication failed: %w", err)
	}

	n.mu.Lock()
	committed := n.CommitIndex >= entryIndex
	n.mu.Unlock()

	if !committed {
		return fmt.Errorf(
			"entry %d was not committed: majority unavailable",
			entryIndex,
		)
	}

	if err := n.ApplyCommittedEntries(store); err != nil {
		return fmt.Errorf("applying committed entries: %w", err)
	}

	_ = n.ReplicateLogOnce(peers)

	return nil
}
