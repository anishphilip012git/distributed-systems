package server

import (
	"errors"
	"fmt"
	"log"
	"time"
	pb "tpc/proto"
)

func (s *TPCServer) CalculateAcceptedTxns(txEntry map[int64]*pb.Transaction) map[int64]*pb.Transaction {
	// for k,v:=range txEntry{
	// 	if _,present:=s.
	// }
	return txEntry

}

// const KEY_DNE = "KEY_DNE"

var rpcCtxDeadlineTimeout time.Duration = 10 * time.Second
var lockTimeout = 0 * time.Millisecond

// now it runs for both sender and receiver
func (s *TPCServer) InitiatePaxosConsensus(index int64, tx *pb.Transaction) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in consensus on %s: %v", s.ServerID, r)
		}
	}()

	// Step 1: Check if the current server is already part of a higher ballot consensus
	// if s.currentBallotNumber > s.acceptedBallotNumber {
	// 	msg := fmt.Sprintf("Server %s is already in a consensus round with a higher ballot %d > %d for txn:%d", s.ServerID, s.currentBallotNumber, s.acceptedBallotNumber, index)
	// 	return false, fmt.Errorf(msg)
	// }

	// Step 2: Increment the current ballot number for this server
	s.currentBallotNumber++

	// Step 3: Prepare phase
	promiseTxnMap, _, err := s.PreparePhaseCall()
	if err != nil {
		s.UnlockResources(tx, index) // Ensure locks are released
		log.Printf("Paxos Prepare failed on %s", s.ServerID)
		return false, err
	}
	// Step 4: Acquire locks
	runPaxos := true
	senderBalance := 0
	if ServerClusters[s.ServerID] == ShardMap[tx.Sender] {
		// Lock the sender's account
		// log.Printf("Lock Sender's account START %d", tx.Sender)

		senderBalance, err = s.BalanceMap.GetAndLock(tx.Sender, index, s.ServerID) //lockTimeout
		if err != nil {
			// s.UnlockResources(tx, index)
			errmSg := fmt.Sprintf(err.Error()+": tx %d : Sender %d, Amount %d", index, tx.Sender, tx.Amount)
			// log.Println(errmSg)
			return false, errors.New(errmSg)
		}
		// log.Printf("PAXOS FAIL CHECK:: balance %d < %d", senderBalance, int(tx.Amount))
		if senderBalance < int(tx.Amount) {
			runPaxos = false
		}
		// log.Printf("LOCK Sender's account SUCCESS %d", tx.Sender)
	}
	if ServerClusters[s.ServerID] == ShardMap[tx.Receiver] {
		// Lock the receiver's account
		// log.Printf("Lock Receiver's account START %d", tx.Sender)
		_, err = s.BalanceMap.GetAndLock(tx.Receiver, index, s.ServerID) //lockTimeout
		if err != nil {
			// s.UnlockResources(tx, index)
			errmSg := fmt.Sprintf(err.Error()+": on %s tx %d : Receiver %d, Amount %d", s.ServerID, index, tx.Sender, tx.Amount)
			log.Println(errmSg)
			return false, errors.New(errmSg)
		}
	}

	if !runPaxos {
		// Abort if insufficient balance
		s.UnlockResources(tx, index)
		errmSg := fmt.Sprintf("Aborting due to insufficient balance: Sender %d, Amount %d", tx.Sender, tx.Amount)
		log.Println(errmSg)
		return false, errors.New(errmSg)
	}

	// Step 5: Accept phase
	acceptedTxnMap := s.CalculateAcceptedTxns(promiseTxnMap)
	acceptedCount, err := s.AcceptPhaseCall(acceptedTxnMap)
	if err != nil {
		s.UnlockResources(tx, index)
		log.Printf("Paxos Accept failed on %s with accepted count %d", s.ServerID, acceptedCount)
		return false, err
	}

	// Step 6: Commit phase
	err = s.DecidePhaseCall(acceptedTxnMap)
	if err != nil {
		s.UnlockResources(tx, index)
		log.Printf("Paxos Decide failed on %s", s.ServerID)
		return false, err
	}

	// Release locks after commit of tpc

	log.Printf("Paxos consensus succeeded on %s", s.ServerID)
	return true, nil
}

// UnlockResources releases locks on both sender and receiver after transaction processing
func (s *TPCServer) UnlockResources(tx *pb.Transaction, index int64) {
	if ServerClusters[s.ServerID] == ShardMap[tx.Sender] {
		s.BalanceMap.UnlockKey(tx.Sender, index, s.ServerID)
		// log.Printf("UNLOCK Sender's account SUCCESS %d for txn:%d", tx.Sender, index)
	}

	if ServerClusters[s.ServerID] == ShardMap[tx.Receiver] {
		s.BalanceMap.UnlockKey(tx.Receiver, index, s.ServerID)
		// log.Printf("UNLOCK Receiver's account SUCCESS %d for txn:%d", tx.Receiver, index)
	}

}
