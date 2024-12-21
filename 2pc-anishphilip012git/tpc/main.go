package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"tpc/client"
	"tpc/data"
	"tpc/db"
	"tpc/server"
)

func main() {
	// Initialize file logging with timestamped file name
	currentTime := time.Now()
	timestamp := currentTime.Format("02012006_15_04")
	fileName := fmt.Sprintf("output_%s.log", timestamp)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	// Create a MultiWriter that writes to both stdout and the file
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	var totalServers, numClusters int

	// Function to prompt user and ensure non-zero positive integers
	reader := bufio.NewReader(os.Stdin)

	for {
		log.Println("Enter Total Servers (must be greater than 0):")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input) // Remove leading/trailing spaces
		value, err := strconv.Atoi(input)
		if err != nil || value <= 0 {
			log.Println("Invalid input. Please enter a positive integer.")
			continue
		}
		totalServers = value
		break
	}

	for {
		log.Println("Enter Number of Clusters (must be greater than 0):")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		value, err := strconv.Atoi(input)
		if err != nil || value <= 0 {
			log.Println("Invalid input. Please enter a positive integer.")
			continue
		}
		numClusters = value
		break
	}

	// Log the validated inputs
	log.Printf("Total Servers: %d, Number of Clusters: %d\n", totalServers, numClusters)

	// Example usage of server and client logic
	server.GenerateServerConfig(totalServers, numClusters)

	noOfClients := 3000
	client.GenerateClientConfig(noOfClients, totalServers, numClusters)

	// Output the generated configuration
	db.StartallDB(server.AllServers)
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
	for rowNo < len(csvRows) {
		log.Println("MAIN :: processing Set ", currentSet)
		server.MakeRoundwiseConnection(csvRows[rowNo].LiveServers)
		server.ValidatePeerConnections(csvRows[rowNo].LiveServers)
		server.UpdateContactServers(csvRows[rowNo].ContactServers)

		for rowNo < len(csvRows) && csvRows[rowNo].Set == currentSet {
			rowNoCopy := rowNo
			go func(row int) {
				clientIndex := csvRows[row].Txn.Sender
				log.Printf("MAIN :: Client %d sends Transaction %d : %v", clientIndex, row+1, csvRows[row].Txn)
				client.AllClients[clientIndex-1].SendTransaction(int64(row+1), csvRows[row].Txn)
			}(rowNoCopy)
			rowNo++
		}

		val := 0
		for val == 0 {
			val = RunInteractiveSession()
		}

		currentSet += val
	}

	server.CloseRoundwiseConnections()
	log.Println("All transactions processed.")
}
