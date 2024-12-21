package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"pbft1/client"
	"pbft1/data"
	"pbft1/db"
	"pbft1/server"
	"time"
)

/*
PeerAddresses: map[S1:localhost:50001 S2:localhost:50002 S3:localhost:50003 S4:localhost:50004 S5:localhost:50005 S6:localhost:50006 S7:localhost:50007]
PeerIdMap: map[localhost:50001:S1 localhost:50002:S2 localhost:50003:S3 localhost:50004:S4 localhost:50005:S5 localhost:50006:S6 localhost:50007:S7]
AllServers: [S1 S2 S3 S4 S5 S6 S7]
LiveSvrMap: map[S1:true S2:true S3:true S4:true S5:true S6:true S7:true]

*/
// Global variables
// var (
//
//	AllServers = []string{"S1", "S2", "S3", "S4", "S5", "S6", "S7"}                                                  // List of all servers
//	LiveSvrMap = map[string]bool{"S1": true, "S2": true, "S3": true, "S4": true, "S5": true, "S6": true, "S7": true} // Map of server status
//	// 	CrashSvrMap=map[string]bool{"S1": false, "S2": false, "S3": false, "S4": false, "S5": false}
//	mu sync.Mutex
//
// )

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

	n := 7 // You can set n to any desired number of servers
	f := 2

	server.GenerateServerConfig(n)
	client.GenerateClientConfig(10, f)

	// Output the generated configuration
	db.StartallDB(server.AllServers)
	// db.InitDBConnections(server.AllServers)

	//Paxos starting of individual servers
	server.StartAllServers(server.AllServers, 10)
	log.Println("PBFT Server length :", len(server.PBFTSvrMap))

	server.MakePeerConnections(server.AllServers)

	// Read transactions from CSV
	csvRows, err := data.ReadCSV("data/test.csv")
	if err != nil {
		log.Fatalf("Failed to read transactions: %v", err)
	}

	currentSet := 1
	rowNo := 0
	// scanner := bufio.NewScanner(os.Stdin)
	// Iterate through transactions and handle terminal options after each iteration
	for rowNo < len(csvRows) {
		log.Println("processing Set ", currentSet)
		// Map transactions to clients based on the sender (just an example)
		log.Println(rowNo, csvRows[rowNo].Set, currentSet, csvRows[rowNo].LiveServers)

		server.MakeRoundwiseConnection(csvRows[rowNo].LiveServers)
		err := server.ValidatePeerConnections(csvRows[rowNo].LiveServers)
		server.InitByzSvrMap(csvRows[rowNo].ByzantineServers)
		log.SetPrefix(fmt.Sprintf("Set:%d", currentSet))
		log.Println("ByzSvrMap", server.ByzSvrMap)
		if err != nil {
			log.Printf("Peer servers are down with err:%v", err)
			log.Println("Client can't go through the Transaction ", rowNo, csvRows[rowNo].Txn)
			rowNo += 1
			continue
		}
		// db.ClearAllDB(server.AllServers)
		server.ResetServerStats()
		log.Printf("%d  < %d , %d == %d ", rowNo, len(csvRows), csvRows[rowNo].Set, currentSet)
		for rowNo < len(csvRows) && csvRows[rowNo].Set == currentSet {
			rowNoCopy := rowNo // Avoid potential issues with closure variables
			cltIdx := 0

			go func(row int) {
				// if clientIndex, ok := client.ClientMap[csvRows[row].Txn.Sender]; ok {
				// 	log.Println("Client sends Transaction ", row+1, csvRows[row].Txn)
				client.AllClients[cltIdx].SendRequest(int64(row+1), csvRows[row].Txn)
				// } else {
				// 	// Handle unknown sender
				// 	log.Printf("Unknown sender: %s\n", csvRows[row].Txn.Sender)
				// }
			}(rowNoCopy) // Pass rowNoCopy to avoid closure issues
			rowNo++
			cltIdx++
			if cltIdx%10 == 0 {
				time.Sleep(5 * time.Second)
			}
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
	server.CloseRoundwiseConnections()

	log.Println("All transactions processed.")
}
