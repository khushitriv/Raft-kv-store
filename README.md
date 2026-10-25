# Distributed key value store with Raft inspired consensus

This project Is a distributed key-value store built in Go Golang to explore leader election, replicated logs, consensus, and fault tolerance. It uses a three-node cluster with HTTP-based communication and implements core Raft-inspired mechanisms.

## Overview

In distributed systems, multiple nodes must be coordinated while handling failures and keeping replicated data consistent. This project explores the challenges by implementing a small key-value store in which nodes elect a leader, replicate log entries, and use a majority-based commitment.

The focus throughout the development of this project has been on understanding the fundamentals of distributed consensus through implementation, automated testing and manual failure experiments.

## Project Features

In the current state, this project consists of the following features:

Leader election: term-based elections and voting between nodes 
Voting safeguards: It rejects stale term requests and checks candidate log freshness
-Log replication: It sends log entries from the leader to followers over HTTP -Log consistency: It validates previous log indexes and terms, handles conflicting uncommitted entries and protects the committed entries -Majority-based commitment: It uses replication acknowledgements and the current term commitment rule 
-Key value state machine: It applies committed operations to the key- value store
-Heartbeats and failure detection: It periodically contacts peers and starts elections after the leader contact times out
-Follower catch-up: It allows a restarted follower to catch up through log replication 
- Automated tests: It covers core election, replication, key value and commitment behavior

## Tech Stack

-Language: Go/Golang
-Communication: HTTP 
-Concurrency: Goroutines, mutexes, context cancellation
-Testing: Go testing package and race detector
-Concepts: Distributed consensus, leader election, replicated logs, majority quorums and fault tolerance

## Architechture

This system runs three nodes and each maintains consensus state and communicates with peers over HTTP.

                              Client
                                |
                                v
                            Leader Node
                            /       \
                           v         v
                      Follower  Follower
                            \       /
                          HTTP messages

                         Replicated Log
                               |
                         Majority Commit
                               |
                               v
                  Key-Value State Machine

The leader coordinates writes and replicates log entries to the followers. Commitment depends on replication by a majority and the implemented current term rule.

## Project structure:
```
raft-kv-store/
├── cmd/
│   └── node/                 
├── internal/
│   ├── kv/                   
│   └── raft/                 
├── go.mod
├── go.sum
└── README.md
```

## Requirements

- Go installed and available on your path
- A terminal such as command prompt, PowerShell or a Unix shell

Verify your installation: 

go version 

Run the Tests From the project root:
go test ./... 

Run the test suite with Go's race detector 
go test -race ./... 

Both commands passed during development. Run them again on your checkout to verify its current state. 

## Running the Cluster 

The development setup uses three nodes on ports 8001, 8002, and 8003. 

Start each node using the startup configuration and commands defined in the project. Keep each node running in a separate terminal to observe elections, heartbeats, and replication. 

The health endpoint can be accessed at: http://localhost:8001/health 

The key-value API includes a write endpoint: 
POST http://localhost:8002/kv/set 

The exact startup commands and request/response formats are to be added later. Please refer to doc section for further updates.

## Testing and Failure Experiments 

The following scenarios were exercised during development: 

1. Data replication A value was written through one node and successfully read from other nodes
2. Follower recovery A follower was stopped while the remaining nodes continued operating. After writes were replicated to the available majority, the follower was restarted and caught up through log replication 
3. Leader failover After the current leader was stopped, another node was elected leader in a later term. A subsequent write was readable from another running node 
4. Majority availability With one node stopped, the remaining two-node majority continued operating and replicated a write
5. Automated testing Tests cover key-value operations, term-based voting, candidate log freshness, replication consistency, conflicting entries, committed-entry protection, and majority-based commitment. 

## Limitations 

This is an educational prototype, not a production-ready distributed database. 

- In-memory state: Logs, terms, and key-value data are not persisted to disk. Restarting all nodes loses the previous cluster state. 

- Limited fault testing: The tested scenarios do not cover every possible network partition, message ordering, or crash timing. 

- Protocol correctness: Passing tests and the race detector does not prove full Raft safety and liveness under all conditions. 

- No log compaction: Snapshots and log compaction are not implemented. 

- No dynamic membership: Safe cluster membership changes are outside the current scope. 

- Production guarantees: Durable storage, security, operational monitoring, and comprehensive linearizability guarantees have not been established. 

## Key Learnings 

- Implementing term-based leader election and voting. 
- Understanding log replication, consistency checks, and commit indexes. - Applying majority quorums to fault-tolerant coordination. 
- Managing concurrent Go code using mutexes, goroutines, and contexts. 
- Testing distributed-system components and investigating concurrency issues. 

## Future Improvements 

- Potential extensions include persistent logs, crash recovery, network-partition testing, broader protocol-level tests, and snapshotting.