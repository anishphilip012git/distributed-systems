package main

import (
	"bufio"
	"log"
	"os"
	"pbft1/server"
	s "pbft1/server"
	"strconv"
	"strings"
)

func PrintMenu() {
	log.Println("Choose an option:")
	log.Println("1. PrinStatus <client> or all")
	log.Println("2. PrintLog <server> or all")
	log.Println("3. PrintDB <server> or all")
	log.Println("4. Performance <server> or all")
	log.Println("5. NextIteration")
	log.Println("6. PrintView <server>")
	log.Println("7. BalanceMap <server>")
}

func RunInteractiveSession() int {
	log.SetPrefix("")
	scanner := bufio.NewScanner(os.Stdin)
	servers := server.PBFTSvrMap
	for {
		PrintMenu()
		log.Print("Enter option: ")
		scanner.Scan()
		input := scanner.Text()
		parts := strings.Split(input, " ")

		switch parts[0] {
		case "PrinStatus", "1":
			if len(parts) != 2 {
				log.Println("Usage: PrinStatus <server>")
				continue
			}
			svr := parts[1]
			for _, server := range servers {
				if svr == "all" {
					server.PrinStatus(-1)
				} else {
					seqNo, _ := strconv.Atoi(svr)
					server.PrinStatus(int64(seqNo))
				}
			}
		case "PrintLog", "2":
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
		case "PrintDB", "3":
			if len(parts) != 2 {
				log.Println("Usage: PrintDB <server>")
				continue
			}
			serverName := parts[1]
			for serverID, _ := range servers {
				if serverID == serverName || serverName == "all" {
					s.PrintDB(serverID)
				}
			}
		case "Performance", "4":
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
			// GetTxnTimes()
		case "NextIteration", "5":
			log.Println("Proceeding to next iteration...")
			return 1
		case "PrintView", "6":
			if len(parts) != 2 {
				log.Println("Usage: Performance <server>")
				continue
			}
			serverName := parts[1]
			for serverID, server := range servers {
				if serverID == serverName || serverName == "all" {
					server.PrintView()
				}
			}
		case "BalanceMap", "7":
			if len(parts) != 2 {
				log.Println("Usage: BalanceMap <server>")
				continue
			}
			serverName := parts[1]
			minBalances := map[string]int{}
			for _, server := range servers {
				if server.ServerID == serverName || serverName == "all" {
					log.Printf("Balance Map from server %s is %v", server.ServerID, server.BalanceMap)
				}
				for key, value := range server.BalanceMap {
					if _, exists := minBalances[key]; !exists {
						minBalances[key] = value
					} else {
						// If the key exists, store the minimum value
						minBalances[key] = min(minBalances[key], value)
						//Here min would work because each server local log only reduces the value,
						// so we know we only need to use the minimum of the balances to get the updated balance
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
