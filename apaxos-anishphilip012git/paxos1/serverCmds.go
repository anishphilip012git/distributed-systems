package main

import (
	"fmt"
	"log"
)

func (s *PaxosServer) PrintBalance(client string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	//balance, ok := s.balance[client]
	balance := s.balance
	// /if ok {
	log.Printf("Balance of %s on server %s: %d\n", client, s.serverID, balance)
	// } else {
	// 	log.Printf("Client %s not found on server %s\n", client, s.serverID)
	// }
}

func (s *PaxosServer) PrintLog() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("Log for server %s:\n", s.serverID)
	for index, tx := range s.uncommittedLogs {
		log.Printf("INdex: %d Transaction: %s -> %s, Amount: %d\n", index, tx.Sender, tx.Receiver, tx.Amount)
	}
	for index, tx := range s.buffer {
		log.Printf("INdex: %d Transaction: %s -> %s, Amount: %d\n", index, tx.Sender, tx.Receiver, tx.Amount)
	}

}

func (s *PaxosServer) PrintDB() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("Datastore for server %s:\n", s.serverID)
	// for client, balance := range s.balance {
	// 	log.Printf("%s: %d\n", client, balance)
	// }
	record, err := ScanTable(DB, s.serverID)
	if err != nil {
		log.Println("Error in getting datastore")
	}
	output := ""
	for index, txn := range record {
		output += fmt.Sprintf("[%d: %s] , ", index, txn)
	}
	log.Println("PrintDB output:", output)
}

func (s *PaxosServer) Performance() {
	// Placeholder for performance metrics
	PrintMetricsForAllServersAndHandlers(s.serverID)
}
