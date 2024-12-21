package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"tpcbyz/client"
	"tpcbyz/data"
	"tpcbyz/db"
	"tpcbyz/server"
)

/*
PeerAddresses: map[S1:localhost:50001 S2:localhost:50002 S3:localhost:50003 S4:localhost:50004 S5:localhost:50005 S6:localhost:50006 S7:localhost:50007]
PeerIdMap: map[localhost:50001:S1 localhost:50002:S2 localhost:50003:S3 localhost:50004:S4 localhost:50005:S5 localhost:50006:S6 localhost:50007:S7]
AllServers: [S1 S2 S3 S4 S5 S6 S7]
LiveSvrMap: map[S1:true S2:true S3:true S4:true S5:true S6:true S7:true]
*/
func main() {
	// // Initialize serverMappings A, B, C, D, E

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

	totalServers := 12
	numClusters := 3
	F := 1
	server.GenerateServerConfig(totalServers, numClusters, F)

	noOfClients := 3000
	client.GenerateClientConfig(noOfClients, numClusters, F)

	// Output the generated configuration
	db.StartallDB(server.AllServers)
	// db.InitDBConnections(server.AllServers)

	//Paxos starting of individual servers
	server.StartAllServers(server.AllServers, noOfClients/numClusters)
	log.Println("MAIN :: TPC Server length :", len(server.TPCServerMap))

	server.MakePeerConnections(server.AllServers)

	// Read transactions from CSV
	csvRows, err := data.ReadCSV("data/test.csv")
	if err != nil {
		log.Fatalf("MAIN :: Failed to read transactions: %v", err)
	}

	currentSet := 1
	rowNo := 0
	// scanner := bufio.NewScanner(os.Stdin)
	// Iterate through transactions and handle terminal options after each iteration
	for rowNo < len(csvRows) {
		log.Println("MAIN :: processing Set ", currentSet)
		// Map transactions to clients based on the sender (just an example)
		log.Println("MAIN ::", rowNo, csvRows[rowNo].Set, currentSet, csvRows[rowNo].LiveServers)

		server.MakeRoundwiseConnection(csvRows[rowNo].LiveServers)
		server.ValidatePeerConnections(csvRows[rowNo].LiveServers)
		server.InitByzSvrMap(csvRows[rowNo].ByzantineServers)
		server.UpdateContactServers(csvRows[rowNo].ContactServers)
		log.Printf("MAIN :: --> %d  < %d , %d == %d ", rowNo, len(csvRows), csvRows[rowNo].Set, currentSet)
		for rowNo < len(csvRows) && csvRows[rowNo].Set == currentSet {
			// rowNoCopy := rowNo // Avoid potential issues with closure variables
			row := rowNo
			// go func(row int) {
			clientIndex := (csvRows[row].Txn.Sender)
			log.Printf("MAIN :: Client %d sends Transaction %d : %v", clientIndex, row+1, csvRows[row].Txn)
			client.AllClients.SendRequest(int64(row+1), csvRows[row].Txn)
			// }(rowNoCopy) // Pass rowNoCopy to avoid closure issues
			rowNo++
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
