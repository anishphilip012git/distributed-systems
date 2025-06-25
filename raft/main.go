package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Define the Raft states
const (
	Follower = iota
	Candidate
	Leader
)

// Raft node structure
type Raft struct {
	mu             sync.Mutex
	state          int       // Current state: Follower, Candidate, or Leader
	currentTerm    int       // Latest term server has seen
	votedFor       int       // Candidate ID that received vote in current term
	votesReceived  int       // Number of votes received in the election
	log            []LogEntry
	commitIndex    int
	lastApplied    int
	peers          []int     // Other nodes in the cluster
	id             int       // Server ID
	electionTimer  *time.Timer
	heartbeatTimer *time.Timer
}

// LogEntry represents a single log entry
type LogEntry struct {
	Term    int
	Command interface{}
}

// Election timeout duration
var electionTimeout = 1500 * time.Millisecond

// Create a new Raft node
func NewRaft(id int, peers []int) *Raft {
	r := &Raft{
		state:         Follower,
		currentTerm:   0,
		votedFor:      -1,
		peers:         peers,
		id:            id,
		electionTimer: time.NewTimer(randomElectionTimeout()),
	}
	go r.run()
	return r
}

// Randomize election timeout
func randomElectionTimeout() time.Duration {
	return time.Duration(1500+rand.Intn(1500)) * time.Millisecond
}

// Run Raft node: handles state transitions
func (r *Raft) run() {
	for {
		switch r.state {
		case Follower:
			r.runFollower()
		case Candidate:
			r.runCandidate()
		case Leader:
			r.runLeader()
		}
	}
}

// Follower state: wait for election timeout or leader's heartbeat
func (r *Raft) runFollower() {
	r.mu.Lock()
	fmt.Printf("Node %d is a Follower\n", r.id)
	r.mu.Unlock()

	select {
	case <-r.electionTimer.C:
		r.mu.Lock()
		r.state = Candidate
		r.mu.Unlock()
	}
}

// Candidate state: start election and request votes
func (r *Raft) runCandidate() {
	r.mu.Lock()
	fmt.Printf("Node %d is a Candidate\n", r.id)
	r.currentTerm++
	r.votedFor = r.id
	r.votesReceived = 1 // Vote for itself
	r.resetElectionTimer()
	r.mu.Unlock()

	// Send RequestVote RPCs to other peers (simulated)
	for _, peer := range r.peers {
		if peer != r.id {
			go r.sendRequestVote(peer)
		}
	}

	// Collect votes and either become leader or restart election
	for r.votesReceived < len(r.peers)/2+1 {
		select {
		case <-r.electionTimer.C:
			r.mu.Lock()
			r.state = Candidate
			r.mu.Unlock()
			return
		}
	}
	r.mu.Lock()
	r.state = Leader
	r.mu.Unlock()
}

// Leader state: send heartbeats to followers
func (r *Raft) runLeader() {
	r.mu.Lock()
	fmt.Printf("Node %d is the Leader\n", r.id)
	r.mu.Unlock()

	// Start sending heartbeats (AppendEntries)
	for {
		r.mu.Lock()
		r.sendHeartbeats()
		r.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

// Simulate sending heartbeats (AppendEntries)
func (r *Raft) sendHeartbeats() {
	for _, peer := range r.peers {
		if peer != r.id {
			go r.sendAppendEntries(peer)
		}
	}
}

// Simulate RequestVote RPC
func (r *Raft) sendRequestVote(peer int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Printf("Node %d is requesting vote from Node %d\n", r.id, peer)
	// Simulate vote granting process
	voteGranted := rand.Intn(2) == 0
	if voteGranted {
		r.votesReceived++
		fmt.Printf("Node %d received vote from Node %d\n", r.id, peer)
	}
}

// Simulate AppendEntries (heartbeat) RPC
func (r *Raft) sendAppendEntries(peer int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Printf("Leader %d is sending heartbeat to Node %d\n", r.id, peer)
	// Simulate response
}

// Reset election timer
func (r *Raft) resetElectionTimer() {
	r.electionTimer.Reset(randomElectionTimeout())
}

func main() {
	// Create a Raft cluster with 5 nodes
	nodes := []int{1, 2, 3, 4, 5}
	for _, id := range nodes {
		go NewRaft(id, nodes)
	}

	// Run indefinitely
	select {}
}
