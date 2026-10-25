package raft

import (
	"fmt"
	"time"
)

func (n *NodeState) RequestVote(args RequestVoteArgs) RequestVoteReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	reply := RequestVoteReply{
		Term:        n.CurrentTerm,
		VoteGranted: false,
	}

	// reject requests from an older term
	if args.Term < n.CurrentTerm {
		return reply
	}

	// new term means we must step down and update the term
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.VotedFor = ""
		n.Role = Follower
		n.VotesReceived = 0
		n.VotesFrom = make(map[string]bool)
	}

	reply.Term = n.CurrentTerm

	// To find this node's latest log index and term
	var lastLogIndex uint64
	var lastLogTerm uint64

	if len(n.Log) > 0 {
		lastEntry := n.Log[len(n.Log)-1]
		lastLogIndex = lastEntry.Index
		lastLogTerm = lastEntry.Term
	}

	// considers the candidate's log up to date if its
	// last log term is greater, or the terms match and its log index is at least as large
	candidateLogIsUpToDate :=
		args.LastLogTerm > lastLogTerm ||
			(args.LastLogTerm == lastLogTerm &&
				args.LastLogIndex >= lastLogIndex)

	// grant at most one vote per term, and only to an up-to-date candidate.
	if (n.VotedFor == "" || n.VotedFor == args.CandidateID) &&
		candidateLogIsUpToDate {

		n.VotedFor = args.CandidateID
		n.Role = Follower
		n.LastLeaderContact = time.Now()
		n.ElectionTimeout = randomElectionTimeout()
		reply.VoteGranted = true
	}

	return reply
}

func (n *NodeState) StartElection() (RequestVoteArgs, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Role == Leader {
		return RequestVoteArgs{}, fmt.Errorf("leader cannot start an election")
	}

	n.CurrentTerm++
	n.Role = Candidate
	n.VotedFor = n.ID
	n.VotesReceived = 1
	n.VotesFrom = make(map[string]bool)
	n.LastLeaderContact = time.Now()
	n.ElectionTimeout = randomElectionTimeout()

	// including latest log position in the vote request
	var lastLogIndex uint64
	var lastLogTerm uint64

	if len(n.Log) > 0 {
		lastEntry := n.Log[len(n.Log)-1]
		lastLogIndex = lastEntry.Index
		lastLogTerm = lastEntry.Term
	}

	return RequestVoteArgs{
		Term:         n.CurrentTerm,
		CandidateID:  n.ID,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}, nil
}

func HasMajority(votesReceived, totalNodes int) bool {
	return totalNodes > 0 && votesReceived > totalNodes/2
}

func (n *NodeState) RecordVoteResponse(
	peerID string,
	requestTerm uint64,
	reply RequestVoteReply,
	totalNodes int,
) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if reply.Term > n.CurrentTerm {
		n.CurrentTerm = reply.Term
		n.Role = Follower
		n.VotedFor = ""
		n.VotesReceived = 0
		n.VotesFrom = make(map[string]bool)
		return
	}

	if n.Role != Candidate || requestTerm != n.CurrentTerm {
		return
	}

	if !reply.VoteGranted || reply.Term != requestTerm {
		return
	}

	if n.VotesFrom[peerID] {
		return
	}

	n.VotesFrom[peerID] = true
	n.VotesReceived++

	if HasMajority(n.VotesReceived, totalNodes) {
		n.Role = Leader
	}
}

func (n *NodeState) RecordVote(
	peerID string,
	granted bool,
	totalNodes int,
) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Role != Candidate || !granted {
		return
	}

	if n.VotesFrom[peerID] {
		return
	}

	n.VotesFrom[peerID] = true
	n.VotesReceived++

	if HasMajority(n.VotesReceived, totalNodes) {
		n.Role = Leader
	}
}
