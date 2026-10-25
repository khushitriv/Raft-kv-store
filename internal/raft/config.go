package raft

type Config struct {
	ID    string
	Port  int
	Peers map[string]string
}
