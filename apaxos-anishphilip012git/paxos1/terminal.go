package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func PrintMenu() {
	log.Println("Choose an option:")
	log.Println("1. PrintBalance <client> or all")
	log.Println("2. PrintLog <server> or all")
	log.Println("3. PrintDB <server> or all")
	log.Println("4. Performance <server> or all")
	log.Println("5. NextIteration")
	log.Println("6. CrashServer <server>")
	log.Println("7. BalanceMap <server>")
}

func RunInteractiveSession() int {
	log.SetPrefix("")
	scanner := bufio.NewScanner(os.Stdin)
	servers := PaxosServerMap
	for {
		PrintMenu()
		log.Print("Enter option: ")
		scanner.Scan()
		input := scanner.Text()
		parts := strings.Split(input, " ")

		switch parts[0] {
		case "PrintBalance":
			if len(parts) != 2 {
				log.Println("Usage: PrintBalance <client>")
				continue
			}
			client := parts[1]
			for _, server := range servers {
				if server.serverID == client {
					server.PrintBalance(client)
				} else if client == "all" {
					server.PrintBalance(server.serverID)
				}
			}
		case "PrintLog":
			if len(parts) != 2 {
				log.Println("Usage: PrintLog <server>")
				continue
			}
			serverName := parts[1]
			for _, server := range servers {
				if server.serverID == serverName || serverName == "all" {
					server.PrintLog()
				}
			}
		case "PrintDB":
			if len(parts) != 2 {
				log.Println("Usage: PrintDB <server>")
				continue
			}
			serverName := parts[1]
			for _, server := range servers {
				if server.serverID == serverName || serverName == "all" {
					server.PrintDB()
				}
			}
		case "Performance":
			if len(parts) != 2 {
				log.Println("Usage: Performance <server>")
				continue
			}
			serverName := parts[1]
			for _, server := range servers {
				if server.serverID == serverName || serverName == "all" {
					server.Performance()
				}
			}
			GetTxnTimes()
		case "NextIteration":
			log.Println("Proceeding to next iteration...")
			return 1
		case "all":
			for _, server := range servers {
				server.PrintBalance(server.serverID)
			}
			for _, server := range servers {
				server.PrintDB()
			}
			for _, server := range servers {
				server.PrintLog()
			}
			for _, server := range servers {
				server.Performance()
			}
			GetTxnTimes()
			minBalances := map[string]int{}
			// val := ""
			for _, server := range servers {
				// val += fmt.Sprintf("Balance Map from server %s is %v \t", server.serverID, server.balanceMap)
				log.Printf("Balance Map from server %s is %v", server.serverID, server.balanceMap)
				for key, value := range server.balanceMap {
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
			log.Printf("Aggregated Balance: %v", minBalances)
		case "CrashServer":
			if len(parts) != 2 {
				log.Println("Usage: CrashServer <server>")
				continue
			}
			serverName := parts[1]
			for _, server := range servers {
				if server.serverID == serverName {
					PaxosServerWrapperMap[serverName].StopServer()
				}
			}
		case "BalanceMap":
			if len(parts) != 2 {
				log.Println("Usage: BalanceMap <server>")
				continue
			}
			serverName := parts[1]
			minBalances := map[string]int{}
			for _, server := range servers {
				if server.serverID == serverName || serverName == "all" {
					log.Printf("Balance Map from server %s is %v", server.serverID, server.balanceMap)
				}
				for key, value := range server.balanceMap {
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
			log.Printf("Aggregated Balance: %v", minBalances)

		default:
			log.Println("Invalid option")
		}
		return 0
	}

}
