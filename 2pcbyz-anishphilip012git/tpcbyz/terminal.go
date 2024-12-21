package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"tpcbyz/db"
	"tpcbyz/server"
)

func PrintMenu() {
	log.Println("Choose an option:")
	log.Println("1. PrintDB <server> or all")
	log.Println("2. Performance <server> or all")
	log.Println("3. NextIteration")
	log.Println("4. BalanceMap <client>")
	log.Println("5. PrintLog <server> or all")
}

func RunInteractiveSession() int {
	log.SetPrefix("")
	scanner := bufio.NewScanner(os.Stdin)
	servers := server.TPCServerMap
	for {
		PrintMenu()
		log.Print("Enter option: ")
		scanner.Scan()
		input := scanner.Text()
		parts := strings.Split(input, " ")

		switch parts[0] {

		case "PrintLog", "5":
			if len(parts) != 2 {
				log.Println("Usage: PrintLog <server>")
				continue
			}
			serverName := parts[1]
			for serverID, server := range servers {
				if serverID == serverName || serverName == "all" {
					server.PrintLog()
				}
			}
		case "PrintDB", "1":
			if len(parts) != 2 {
				log.Println("Usage: PrintDB <server>")
				continue
			}
			serverName := parts[1]
			for serverID, s := range servers {
				if serverID == serverName || serverName == "all" {
					s.PrintDB()
				}
			}
		case "Performance", "2":
			if len(parts) != 2 {
				log.Println("Usage: Performance <server>")
				continue
			}
			serverName := parts[1]
			for serverID, server := range servers {
				if serverID == serverName || serverName == "all" {
					server.Performance()
				}
			}
			db.GetTxnTimes(db.DbConnections["S1"])

		case "NextIteration", "3":
			log.Println("Proceeding to next iteration...")
			return 1

		case "BalanceMap", "4":
			if len(parts) != 2 {
				log.Println("Usage: BalanceMap <client>")
				continue
			}
			arr := strings.Split(parts[1], ",")
			for _, val := range arr {
				clientVal, _ := strconv.Atoi(val)
				client := int64(clientVal)
				for _, svr := range servers {
					if server.ServerClusters[svr.ServerID] == server.ShardMap[client] {
						bal, _ := svr.BalanceMap.Read(client)
						log.Printf("Balance Map for Client %d from server %s is %v", client, svr.ServerID, bal)
					}
				}
			}
		// 	log.Printf("Aggregated Balance: %v", minBalances)

		default:
			log.Println("Invalid option")
		}
		return 0
	}

}
