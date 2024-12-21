package main

import (
	"context"
	"fmt"
	"log"
	pb "paxos1/proto" // Import the generated proto code
	"sync"
	"time"

	"google.golang.org/grpc"
)

type PaxosServer struct {
	pb.UnimplementedPaxosServer
	mu                        sync.Mutex
	serverID                  string
	addr                      string
	currentBallotNumber       int64
	acceptedBallotNumber      int64
	lastCommittedBallotNumber int64
	acceptedValue             map[int64]*pb.Transaction // Changed to handle transactions
	uncommittedLogs           map[int64]*pb.Transaction // Uncommitted logs now hold a list of transactions

	peerServers     map[string]pb.PaxosClient // gRPC clients to other servers
	peerConnections map[string]*grpc.ClientConn
	isDown          bool // Server down status
	balance         int  //map[string]int
	committedLogs   map[int64]*pb.Transaction
	buffer          map[int64]*pb.Transaction
	isLeader        bool // This will indicate if this server is the leader (Proposer)
	balanceMap      map[string]int
}

// Prepare phase: leader proposes a ballot number
func (s *PaxosServer) Prepare(ctx context.Context, req *pb.PrepareRequest) (*pb.PromiseResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := &pb.PromiseResponse{
		Promise: true,
	}

	// s.balanceMap[req.LeaderId] = int(req.LastSvrBalance) //Sync Balance Map
	if req.BallotNumber <= s.currentBallotNumber {
		response.Promise = false
		return response, nil
	}
	// to maintain datastore consistency ,when leader is outdated
	if req.LastCommittedBallot < s.lastCommittedBallotNumber {
		response.Promise = false
		return response, fmt.Errorf("LastCommittedBallot of leader is less than server.Need to SYNC")
	}
	var err error
	if req.LastCommittedBallot > s.lastCommittedBallotNumber {
		err = fmt.Errorf("SYNC")
	}

	s.currentBallotNumber = req.BallotNumber
	response.LastAcceptedBallot = s.acceptedBallotNumber
	response.LastAcceptedValue = s.acceptedValue // No need for accepted value as it's a transaction list
	response.UncommittedTransactions = s.uncommittedLogs
	// response.LastCommittedBallot = s.lastCommittedBallotNumber
	log.Printf("Server %v: Promise for from %s for ballot %v gave %v", s.serverID, req.LeaderId, req.BallotNumber, response)
	return response, err
}

// Accept phase: leader sends a list of transactions for consensus
func (s *PaxosServer) Accept(ctx context.Context, req *pb.AcceptRequest) (*pb.AcceptedResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := &pb.AcceptedResponse{
		Accepted: true,
	}

	if req.BallotNumber < s.currentBallotNumber {
		response.Accepted = false
		return response, nil
	}

	s.acceptedBallotNumber = req.BallotNumber
	s.acceptedValue = req.AccpetedTransactions // Update the accepted transactions

	log.Printf("Server %v accepted transactions with ballot %v", s.serverID, req.BallotNumber)
	return response, nil
}

// Decide phase: commit the transaction block and clean up logs
func (s *PaxosServer) Decide(ctx context.Context, req *pb.DecisionRequest) (*pb.DecisionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Committing transactions on server %v", s.serverID)
	// Data store logs

	s.CommitLogs(req.TransactionsEntries)
	for k, _ := range req.TransactionsEntries {
		_, ok := s.uncommittedLogs[k]
		if ok {
			delete(s.uncommittedLogs, k)
		}
	}
	s.lastCommittedBallotNumber = max(s.lastCommittedBallotNumber, req.BallotNumber)
	response := &pb.DecisionResponse{
		Success: true,
	}
	return response, nil
}

func (s *PaxosServer) CommitLogs(logs map[int64]*pb.Transaction) {
	if s.committedLogs == nil {
		s.committedLogs = map[int64]*pb.Transaction{}

	}
	for k, txn := range logs {
		// log.Printf("GetRecord [%d] on server [%s]: \n", k, s.serverID)
		val, _ := GetRecord(DB, s.serverID, k)
		if val != "" {
			log.Printf("Already executed log %d on server %s: %+v , balance :%d\n", k, s.serverID, txn, s.balance)
			continue
		}

		if s.serverID == txn.Receiver {
			PaxosServerMapLock.Lock()

			PaxosServerMap[txn.Receiver].balance += int(txn.Amount)
			PaxosServerMapLock.Unlock()
		}
		if s.serverID != txn.Sender {
			s.balanceMap[txn.Sender] -= int(txn.Amount)
		}
		s.balanceMap[txn.Receiver] += int(txn.Amount)

		s.committedLogs[k] = txn
		delete(s.uncommittedLogs, k)

		log.Printf("Executing log %d on server %s: %+v , balance :%d , bMap: %v \n", k, s.serverID, txn, s.balance, s.balanceMap)
		err := StoreRecord(DB, s.serverID, k, txn)
		if err != nil {
			log.Println("Error in committing ", err)
		}
	}

}

// Implement the SendTransaction RPC method in paxosServer
func (s *PaxosServer) SendTransaction(ctx context.Context, req *pb.TransactionRequest) (returnVal *pb.TransactionResponse, err error) {
	log.SetPrefix("Server " + s.serverID + "::")
	s.mu.Lock()
	defer s.mu.Unlock()

	// Process the transaction here, e.g., validate the transaction, apply to balance, etc.
	for index, tx := range req.TransactionEntry {
		if s.isDown {
			log.Printf("Server %s is down.Buffering transaction", s.serverID)
			s.buffer[index] = tx
			return &pb.TransactionResponse{
				Success: false,
				Message: "Server is Down",
			}, nil
		} else {
			// if len(s.buffer) > 0 {
			// 	log.Printf("Server %s Buffer is not empty processing buffer first ", s.serverID)

			// 	for bufIndex, buffTxn := range s.buffer {
			// 		s.ProcessTransaction(bufIndex, buffTxn)
			// 	}
			// }
		}
		log.Printf("Server %s received transaction from %s to %s of amount %d\n", s.serverID, tx.Sender, tx.Receiver, tx.Amount)
		returnVal, err = s.ProcessTransaction(index, tx)
		if returnVal != nil {
			return returnVal, err
		}
	}
	// Respond with success
	return &pb.TransactionResponse{
		Success: true,
		Message: "Transaction successful",
	}, nil
}

// gRPC Service: Start Paxos and return immediately.
func (s *PaxosServer) ProcessTransaction(index int64, tx *pb.Transaction) (*pb.TransactionResponse, error) {

	// Asynchronously run Paxos in a goroutine.
	go s.runPaxosUntilSuccess(index, tx)
	time.Sleep(50 * time.Millisecond)
	// Respond immediately to the client, telling it the transaction is being processed.
	return &pb.TransactionResponse{
		Success: true,
		Message: "Transaction processed. Please check logs for further status.",
	}, nil
}

// Separate function to run Paxos in an infinite loop.
func (s *PaxosServer) runPaxosUntilSuccess(index int64, tx *pb.Transaction) {
	start := time.Now() // Record the start time
	if s.serverID == tx.Sender && s.balance < int(tx.Amount) {
		log.Println("Starting Paxos consensus...")
		success, err := s.InitiatePaxosConsensus(index)
		if err != nil || !success {
			log.Printf("Paxos consensus failed: %v", err)
		}
	}
	cnt := 0
	for {
		// Check if balance is sufficient after Paxos.
		if s.serverID == tx.Sender && s.balance >= int(tx.Amount) {
			s.balance -= int(tx.Amount)
			s.balanceMap[s.serverID] -= int(tx.Amount)
			s.addToUncommittedLogs(index, tx)
			delete(s.buffer, index)
			log.Printf("Backlog paxos: Transaction %d processed successfully on %s", index, s.serverID)
			break // Exit the loop after a successful transaction.
		} else {
			if cnt%10 == 2 {
				log.Printf("Backlog paxos: Insufficient balance on %s even after Paxos consensus iteration.", s.serverID)
			}
			cnt++
			s.addToBufferLogs(index, tx)
		}
		time.Sleep(1 * time.Second)
	}
	elapsed := time.Since(start) // Calculate the elapsed time
	TnxSum.Mutex.Lock()
	TnxSum.TotalElapsedTime += elapsed
	TnxSum.TransactionCount++
	TnxSum.Mutex.Unlock()
	StoreTime(DB, index, tx, elapsed)
}

// Helper function to add a transaction to uncommitted logs.
func (s *PaxosServer) addToBufferLogs(index int64, tx *pb.Transaction) {
	if s.buffer == nil {
		s.buffer = make(map[int64]*pb.Transaction)
	}
	s.buffer[index] = tx
}

// Helper function to add a transaction to uncommitted logs.
func (s *PaxosServer) addToUncommittedLogs(index int64, tx *pb.Transaction) {
	if s.uncommittedLogs == nil {
		s.uncommittedLogs = make(map[int64]*pb.Transaction)
	}
	s.uncommittedLogs[index] = tx
}

// Implement the SendTransaction RPC method in paxosServer
func (s *PaxosServer) SyncDataStore(ctx context.Context, req *pb.TransactionRequest) (returnVal *pb.TransactionResponse, err error) {
	log.SetPrefix("Server " + s.serverID + "::")
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Syncing cmmitted transactions")
	s.CommitLogs(req.TransactionEntry)
	// Respond with success
	return &pb.TransactionResponse{
		Success: true,
		Message: "Sync successful",
	}, nil
}
