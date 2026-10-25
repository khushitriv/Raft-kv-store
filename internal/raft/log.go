package raft

type LogEntry struct {
	Term  uint64 `json:"term"`
	Index uint64 `json:"index"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
