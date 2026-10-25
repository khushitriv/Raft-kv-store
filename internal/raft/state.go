package raft

import (
	"math/rand"
	"sync"
	"time"
)

func randomElectionTimeout() time.Duration {
	return time.Duration(1500+rand.Intn(1500)) * time.Millisecond
}

type Role string

const (
	Follower  Role = "follower"
	Candidate Role = "candidate"
	Leader    Role = "leader"
)

type NodeState struct {
	mu sync.Mutex

	ID            string
	Role          Role
	CurrentTerm   uint64
	VotedFor      string
	VotesReceived int
	VotesFrom     map[string]bool

	Log               []LogEntry
	CommitIndex       uint64
	LastApplied       uint64
	NextIndex         map[string]uint64
	MatchIndex        map[string]uint64
	LastLeaderContact time.Time
	ElectionTimeout   time.Duration
}

func NewNodeState(id string) *NodeState {
	return &NodeState{
		ID:                id,
		Role:              Follower,
		CurrentTerm:       0,
		VotedFor:          "",
		VotesReceived:     0,
		VotesFrom:         make(map[string]bool),
		Log:               make([]LogEntry, 0),
		CommitIndex:       0,
		LastApplied:       0,
		NextIndex:         make(map[string]uint64),
		MatchIndex:        make(map[string]uint64),
		LastLeaderContact: time.Now(),
		ElectionTimeout:   randomElectionTimeout(),
	}
}

func (n *NodeState) Snapshot() NodeState {
	n.mu.Lock()
	defer n.mu.Unlock()

	return NodeState{
		ID:                n.ID,
		Role:              n.Role,
		CurrentTerm:       n.CurrentTerm,
		VotedFor:          n.VotedFor,
		VotesReceived:     n.VotesReceived,
		VotesFrom:         nil,
		Log:               append([]LogEntry(nil), n.Log...),
		CommitIndex:       n.CommitIndex,
		LastApplied:       n.LastApplied,
		NextIndex:         copyIndexMap(n.NextIndex),
		MatchIndex:        copyIndexMap(n.MatchIndex),
		LastLeaderContact: n.LastLeaderContact,
		ElectionTimeout:   n.ElectionTimeout,
	}
}

func copyIndexMap(source map[string]uint64) map[string]uint64 {
	copied := make(map[string]uint64, len(source))

	for key, value := range source {
		copied[key] = value
	}

	return copied
}
