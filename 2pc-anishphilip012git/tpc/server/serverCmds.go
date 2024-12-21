package server

import (
	"fmt"
	"log"
	"tpc/db"
	"tpc/metric"
)

func (s *TPCServer) PrintBalance(client int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	//balance, ok := s.balance[client]
	balance, _ := s.BalanceMap.Read(int64(client)) //s.BalanceMap[int64(client)]
	// /if ok {
	log.Printf("Balance of %s on server %s: %d\n", client, s.ServerID, balance)
	// } else {
	// 	log.Printf("Client %s not found on server %s\n", client, s.serverID)
	// }
}

func (s *TPCServer) PrintLog() {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("Log for server %s:\n", s.ServerID)
	for index, tx := range s.uncommittedLogs {
		log.Printf("INdex: %d Transaction: %s -> %s, Amount: %d\n", index, tx.Sender, tx.Receiver, tx.Amount)
	}
	for index, tx := range s.buffer {
		log.Printf("INdex: %d Transaction: %s -> %s, Amount: %d\n", index, tx.Sender, tx.Receiver, tx.Amount)
	}

}
func (s *TPCServer) PrintDB() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Datastore for server %s:\n", s.ServerID)

	// Scan the main database table (TBL_EXEC)
	record, err := db.ScanTable(s.db, db.TBL_EXEC)
	if err != nil {
		log.Println("Error retrieving datastore for TBL_EXEC:", err)
	} else {
		log.Println("+------------------------------------------------------------+")
		log.Println("| Database Table (TBL_EXEC):                                      |")
		log.Println("+------------------------------------------------------------+")
		// for _, txn := range record {
		// 	log.Printf("| %-20s |\t", txn) //log.Printf("| %-16s |\t", txn)
		// }
		output := ""
		for i, txn := range record {
			output += fmt.Sprintf("%-20s", txn)     // Add transaction data to the line
			if (i+1)%4 == 0 || i == len(record)-1 { // Wrap every 4 transactions or at the end
				log.Printf("| %s |\n", output)
				output = "" // Reset the output for the next line
			}
		}
		log.Println("+------------------------------------------------------------+")
	}

	// Scan the Write-Ahead Log table (TBL_WAL)
	record2, err := db.ScanTable(s.db, db.TBL_WAL)
	if err != nil {
		log.Println("Error retrieving datastore for TBL_WAL:", err)
	} else {
		log.Println("+------------------------------------------------------------+")
		log.Println("| Write-Ahead Log Table (TBL_WAL)                                         |")
		log.Println("+------------------------------------------------------------+")
		// output := ""
		for _, txn := range record2 {
			log.Printf("| %-16s |\t", txn)
		}
		log.Println("+------------------------------------------------------------+")
	}
}

// func (s *TPCServer) PrintDB() {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
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
