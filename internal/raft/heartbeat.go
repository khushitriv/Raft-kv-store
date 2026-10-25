package raft

import (
	"context"
	"log"
	"time"
)

func (n *NodeState) RunHeartbeatLoop(
	ctx context.Context,
	peers map[string]string,
) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	log.Printf("node %s: heartbeat loop started", n.ID)

	lastRole := Role("")

	for {
		select {
		case <-ctx.Done():
			log.Printf("node %s: heartbeat loop stopped", n.ID)
			return

		case <-ticker.C:
			n.mu.Lock()

			role := n.Role
			timeoutExpired := time.Since(n.LastLeaderContact) >= n.ElectionTimeout

			n.mu.Unlock()

			if role != lastRole {
				log.Printf("node %s: role is now %s", n.ID, role)
				lastRole = role
			}

			if role == Leader {
				if err := n.ReplicateLogOnce(peers); err != nil {
					log.Printf(
						"node %s: heartbeat/replication failed: %v",
						n.ID,
						err,
					)
				}
				continue
			}

			if timeoutExpired {
				if err := n.RunElection(peers); err != nil {
					log.Printf("node %s: election failed: %v", n.ID, err)
				}
			}
		}
	}
}
