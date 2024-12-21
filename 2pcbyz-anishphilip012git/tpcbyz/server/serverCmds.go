package server

import (
	"log"
	"strings"
	"tpcbyz/db"
	"tpcbyz/metric"
)

func (s *TPCServer) PrintBalance(client int) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	//balance, ok := s.balance[client]
	balance, _ := s.BalanceMap.Read(int64(client)) //s.BalanceMap[int64(client)]
	// /if ok {
	log.Printf("Balance of %s on server %s: %d\n", client, s.ServerID, balance)
	// } else {
	// 	log.Printf("Client %s not found on server %s\n", client, s.serverID)
	// }
}

func (s *TPCServer) PrintLog() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	log.Printf("Log for server %s:\n", s.ServerID)
	// for index, row := range s.T {
	// 	log.Printf("INdex: %d Transaction: %s -> %s, Amount: %d\n", index, tx.Sender, tx.Receiver, tx.Amount)
	// }
	// for index, tx := range s.PrepareLogs {
	// 	log.Printf("INdex: %d Transaction: %s -> %s, Amount: %d\n", index, tx.Sender, tx.Receiver, tx.Amount)
	// }

}
func (s *TPCServer) PrintDB() {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	log.Printf("Datastore for server %s:\n", s.ServerID)

	// Scan and print TBL_EXEC
	record, err := db.ScanTable(s.db, db.TBL_EXEC)
	if err != nil {
		log.Println("Error retrieving datastore for TBL_EXEC:", err)
	} else {
		log.Printf("TBL_EXEC: %s\n", formatRow(record))
	}

	// Scan and print TBL_WAL
	record2, err := db.ScanTable(s.db, db.TBL_WAL)
	if err != nil {
		log.Println("Error retrieving datastore for TBL_WAL:", err)
	} else {
		log.Printf("TBL_WAL:  %s\n", formatRow(record2))
	}
}

// Helper function to format rows into a single line
func formatRow(rows []string) string {
	if len(rows) == 0 {
		return "No records found"
	}
	return strings.Join(rows, " , ")
}

// func (s *TPCServer) PrintDB() {
// 	s.Mu.Lock()
// 	defer s.Mu.Unlock()
// 	log.Printf("Datastore for server %s:\n", s.ServerID)
// 	// for client, balance := range s.balance {
// 	// 	log.Printf("%s: %d\n", client, balance)
// 	// }
// 	record, err := db.ScanTable(s.db, db.TBL_EXEC)
// 	if err != nil {
// 		log.Println("Error in getting datastore")
// 	}
// 	output := ""
// 	for _, txn := range record {
// 		output += fmt.Sprintf("[%s] , ", txn)
// 	}
// 	log.Println("PrintDB output:", output)

// 	output = ""
// 	record2, err := db.ScanTable(s.db, db.TBL_WAL)
// 	if err != nil {
// 		log.Println("Error in getting datastore")
// 	}
// 	for _, txn := range record2 {
// 		output += fmt.Sprintf("[%s] , ", txn)
// 	}
// 	log.Println("PrintDB WAL output:", output)
// }

func (s *TPCServer) Performance() {
	// Placeholder for performance metrics
	metric.PrintMetricsForAllServersAndHandlers(s.ServerID)
}
