package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var PeerAddresses = map[string]string{
	"S1": "localhost:50051",
	"S2": "localhost:50052",
	"S3": "localhost:50053",
	"S4": "localhost:50054",
	"S5": "localhost:50055",
}

var PeerIdMap = map[string]string{
	"localhost:50051": "S1",
	"localhost:50052": "S2",
	"localhost:50053": "S3",
	"localhost:50054": "S4",
	"localhost:50055": "S5",
}

// Global variables
var (
	AllServers = []string{"S1", "S2", "S3", "S4", "S5"}                                      // List of all servers
	LiveSvrMap = map[string]bool{"S1": true, "S2": true, "S3": true, "S4": true, "S5": true} // Map of server status
	// 	CrashSvrMap=map[string]bool{"S1": false, "S2": false, "S3": false, "S4": false, "S5": false}
	mu sync.Mutex
	DB *badger.DB // Mutex to ensure thread-safe access
)
var PaxosServerMapLock = sync.RWMutex{}
var PaxosServerMap = map[string]*PaxosServer{}

func main() {
	// // Initialize serverMappings A, B, C, D, E
	//serverMappings := []*ServerMappingAtClient{}
	currentTime := time.Now()
	// Format the time as ddmmyyyy_hh_mm
	timestamp := currentTime.Format("02012006_15_04")
	fileName := fmt.Sprintf("output_%s.log", timestamp)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	// Create a MultiWriter that writes to both stdout and the file
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Set the log output to the MultiWriter
	log.SetOutput(multiWriter)
	log.SetPrefix("main ::")
	clients := []*Client{}
	for i, serverID := range AllServers {
		svrMapping := NewServerMapping(serverID, PeerAddresses[serverID])
		//serverMappings = append(serverMappings, svrMapping)
		clientID := fmt.Sprintf("Client %d", i+1)
		clients = append(clients, NewClient(clientID, svrMapping))

	}
	DB = StartDB()
	//Paxos starting of individual servers
	StartAllServers(AllServers)
	log.Println("Paxos Server length :", len(PaxosServerMap))

	MakePeerConnections(AllServers)

	// Read transactions from CSV
	csvRows, err := ReadCSV("test.csv")
	if err != nil {
		log.Fatalf("Failed to read transactions: %v", err)
	}
	clientMap := map[string]int{
		"S1": 0, "S2": 1, "S3": 2, "S4": 3, "S5": 4,
	}
	currentSet := 1
	rowNo := 0

	// Iterate through transactions and handle terminal options after each iteration
	for rowNo < len(csvRows) {
		log.Println("processing Set ", currentSet)
		// Map transactions to clients based on the sender (just an example)
		log.Println(rowNo, csvRows[rowNo].Set, currentSet, csvRows[rowNo].LiveServers)

		MakeRoundwiseConnection(csvRows[rowNo].LiveServers)
		err := ValidatePeerConnections(csvRows[rowNo].LiveServers)
		if err != nil {
			log.Printf("Peer servers are down with err:%v", err)
			log.Println("Client can't go through the Transaction ", rowNo, csvRows[rowNo].Txn)
			rowNo += 1
			continue
		}
		for rowNo < len(csvRows) && csvRows[rowNo].Set == currentSet {

			if clientIndex, ok := clientMap[csvRows[rowNo].Txn.Sender]; ok {
				log.Println("Client sends Transaction ", rowNo, csvRows[rowNo].Txn)
				clients[clientIndex].SendTransaction(int64(rowNo), csvRows[rowNo].Txn)
			} else {
				// Handle unknown sender
				log.Printf("Unknown sender: %s\n", csvRows[rowNo].Txn.Sender)
			}
			PaxosServerMap[csvRows[rowNo].Txn.Sender].PrintBalance("")

			rowNo += 1

		}
		// log.Println("New i to", i)
		// Interactive session after processing each batch
		val := 0
		for val == 0 {
			val = RunInteractiveSession()
		}
		// CloseRoundwiseConnections()
		currentSet += val

	}
	CloseRoundwiseConnections()

	log.Println("All transactions processed.")
}
