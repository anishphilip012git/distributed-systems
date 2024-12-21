package data

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type CSVRow struct {
	Set              int
	Txn              Transaction
	LiveServers      []string
	ByzantineServers []string
	ContactServers   []string
}
type Transaction struct {
	Sender   int
	Receiver int
	Amount   int
}

func getTxn(txnString string) Transaction {
	record := strings.Split(strings.TrimSpace(strings.Trim(txnString, "()")), ",")
	// record := txnString.replace.replace(")", "").split(",")

	amount, err := strconv.Atoi(strings.TrimSpace(record[2]))
	if err != nil {
		log.Fatal("Error parsing csv for amount", err, record[2])
	}
	sender, err := strconv.Atoi(strings.TrimSpace(record[0]))
	if err != nil {
		log.Fatal("Error parsing csv for amount", err, record[2])
	}
	recvr, err := strconv.Atoi(strings.TrimSpace(record[1]))
	if err != nil {
		log.Fatal("Error parsing csv for amount", err, record[2])
	}
	tx := Transaction{
		Sender:   sender,
		Receiver: recvr,
		Amount:   amount,
	}
	return tx
}
func getLiveSvrs(ls string) []string {
	record := strings.Split(strings.TrimSpace(strings.Trim(ls, "[]")), ",")
	for i, row := range record {
		record[i] = strings.TrimSpace(row)
	}
	return record
}

// ReadCSV function with logic to handle missing values.
func ReadCSV(filename string) ([]CSVRow, error) {
	log.Println("Reading CSV...")
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read the first record to check for BOM.
	firstRecord, err := reader.Read()
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(firstRecord[0], "\ufeff") {
		firstRecord[0] = strings.TrimPrefix(firstRecord[0], "\ufeff")
	}

	var (
		csvRows            []CSVRow
		lastSet            int
		lastLiveServers    []string
		lastByzServers     []string
		lastContactServers []string
	)

	// Process the first record explicitly to initialize `lastSet` and `lastLiveServers`.
	set, err := strconv.Atoi(firstRecord[0])
	if err != nil {
		return nil, fmt.Errorf("error in parsing Set: %v", err)
	}
	// log.Println(firstRecord)
	lastSet = set
	lastLiveServers = getLiveSvrs(firstRecord[2])
	lastContactServers = getLiveSvrs(firstRecord[3])
	lastByzServers = getLiveSvrs(firstRecord[4])
	csvRow := CSVRow{
		Set:              lastSet,
		Txn:              getTxn(firstRecord[1]),
		LiveServers:      lastLiveServers,
		ByzantineServers: lastByzServers,
		ContactServers:   lastContactServers,
	}
	csvRows = append(csvRows, csvRow)

	// Process the remaining records.
	for {
		record, err := reader.Read()
		if err != nil {
			log.Println(err)
			break // Stop reading on EOF or error.
		}

		// Handle missing Set value.
		if record[0] != "" {
			set, err = strconv.Atoi(record[0])
			if err != nil {
				return nil, fmt.Errorf("error in parsing Set: %v", err)
			}
			lastSet = set
		}

		// Handle missing LiveServers value.
		if record[2] != "" {
			lastLiveServers = getLiveSvrs(record[2])
		}
		if record[3] != "" {
			lastContactServers = getLiveSvrs(record[3])
		}
		if record[4] != "" {
			lastByzServers = getLiveSvrs(record[4])
		}

		// Create a new CSVRow with the current or previous values.
		csvRow := CSVRow{
			Set:              lastSet,
			Txn:              getTxn(record[1]),
			LiveServers:      lastLiveServers,
			ByzantineServers: lastByzServers,
			ContactServers:   lastContactServers,
		}
		csvRows = append(csvRows, csvRow)
	}
	// log.Println("csv", csvRows)
	return csvRows, nil
}

// func main() {
// 	csvRows, err := ReadCSV("test.csv")
// 	if err != nil {
// 		log.Fatalf("Failed to read transactions: %v", err)
// 	}
// 	log.Println(len(csvRows))
// 	log.Println(csvRows)
// }
