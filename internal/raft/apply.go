package raft

import (
	"fmt"

	"raft-kv-store/internal/kv"
)

func (n *NodeState) ApplyCommittedEntries(store *kv.Store) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.CommitIndex > uint64(len(n.Log)) {
		return fmt.Errorf(
			"commit index %d exceeds log length %d",
			n.CommitIndex,
			len(n.Log),
		)
	}
	for n.LastApplied < n.CommitIndex {

		entry := n.Log[n.LastApplied]

		store.Set(entry.Key, entry.Value)
		n.LastApplied++
	}

	return nil
}
