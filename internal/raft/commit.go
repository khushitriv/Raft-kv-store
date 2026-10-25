package raft

func (n *NodeState) AdvanceCommit(
	replicatedBy map[string]uint64,
	peers map[string]string,
) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Role != Leader {
		return false
	}

	totalNodes := len(peers) + 1

	for index := uint64(len(n.Log)); index > n.CommitIndex; index-- {
		entry := n.Log[index-1]

		// directly commits entries from the current term
		if entry.Term != n.CurrentTerm {
			continue
		}

		votes := 1 // the leader has its own copy

		for peerID := range peers {
			matchIndex := n.MatchIndex[peerID]

			// include acknowledgements received during this round
			if replicatedBy[peerID] > matchIndex {
				matchIndex = replicatedBy[peerID]
			}

			if matchIndex >= index {
				votes++
			}
		}

		if HasMajority(votes, totalNodes) {
			n.CommitIndex = index
			return true
		}
	}

	return false
}
