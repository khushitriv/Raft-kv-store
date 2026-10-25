package raft

import "time"

func (n *NodeState) AppendEntries(args AppendEntriesArgs) AppendEntriesReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	reply := AppendEntriesReply{
		Term:    n.CurrentTerm,
		Success: false,
	}

	if args.Term < n.CurrentTerm {
		return reply
	}

	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.VotedFor = ""
		n.VotesReceived = 0
		n.VotesFrom = make(map[string]bool)
	}

	n.Role = Follower
	n.LastLeaderContact = time.Now()
	n.ElectionTimeout = randomElectionTimeout()
	reply.Term = n.CurrentTerm

	if args.PrevLogIndex > uint64(len(n.Log)) {
		return reply
	}

	if args.PrevLogIndex > 0 {
		previous := n.Log[args.PrevLogIndex-1]
		if previous.Term != args.PrevLogTerm {
			return reply
		}
	}

	for i, entry := range args.Entries {
		expectedIndex := args.PrevLogIndex + uint64(i) + 1

		if entry.Index != expectedIndex {
			return reply
		}

		if entry.Index <= uint64(len(n.Log)) {
			existing := n.Log[entry.Index-1]

			if existing.Term == entry.Term {
				continue
			}

			if entry.Index <= n.CommitIndex {
				return reply
			}

			n.Log = n.Log[:entry.Index-1]
		}

		n.Log = append(n.Log, entry)
	}

	if args.LeaderCommit > n.CommitIndex {
		lastIndex := uint64(len(n.Log))

		if args.LeaderCommit < lastIndex {
			n.CommitIndex = args.LeaderCommit
		} else {
			n.CommitIndex = lastIndex
		}
	}

	reply.Success = true
	return reply
}
