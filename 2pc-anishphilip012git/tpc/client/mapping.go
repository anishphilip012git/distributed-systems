package client

func DistributeClients(clients int, numClusters int) []int {
	// Calculate the total number of clients per cluster
	clientsPerCluster := clients / numClusters
	extraClients := clients % numClusters

	// Create a slice to store the cluster assignment for each client (1-indexed)
	clientToCluster := make([]int, clients+1)
	clientIndex := 1 // Start from 1 for 1-indexing

	for cluster := 1; cluster <= numClusters; cluster++ {
		count := clientsPerCluster
		if extraClients > 0 {
			count++ // Add an extra client to this cluster
			extraClients--
		}

		for i := 0; i < count; i++ {
			if clientIndex <= clients {
				clientToCluster[clientIndex] = cluster
				clientIndex++
			}
		}
	}

	// Log the resulting array for debugging purposes
	return clientToCluster // Return the 1-indexed part
}

// func DistributeClients(clients int, numClusters int) []int {
// 	// Calculate the total number of clients per cluster
// 	clientsPerCluster := clients / numClusters
// 	extraClients := clients % numClusters

// 	// Create a slice to store the cluster assignment for each client
// 	clientToCluster := make([]int, clients)
// 	clientIndex := 0

// 	for cluster := 1; cluster <= numClusters; cluster++ {
// 		count := clientsPerCluster
// 		if extraClients > 0 {
// 			count++ // Add an extra client to this cluster
// 			extraClients--
// 		}

// 		for i := 0; i < count; i++ {
// 			if clientIndex < clients {
// 				clientToCluster[clientIndex] = cluster
// 				clientIndex++
// 			}
// 		}
// 	}

// 	// Log the resulting array for debugging purposes
// 	log.Println(clientToCluster)
// 	return clientToCluster
// }

// func DistributeClients(clients int, numClusters int) map[int]int {
// 	// Calculate the total number of clients per cluster
// 	clientsPerCluster := clients / numClusters
// 	extraClients := clients % numClusters

// 	// Create a mapping of client to cluster
// 	clientToCluster := make(map[int]int)
// 	clientIndex := 0

// 	for cluster := 1; cluster <= numClusters; cluster++ {
// 		count := clientsPerCluster
// 		if extraClients > 0 {
// 			count++ // Add an extra client to this cluster
// 			extraClients--
// 		}

// 		for i := 0; i < count; i++ {
// 			if clientIndex < clients {
// 				clientToCluster[clientIndex] = cluster
// 				clientIndex++
// 			}
// 		}
// 	}
// 	log.Println(clientToCluster)
// 	return clientToCluster
// }

// func main() {
// 	clients := []string{"A", "B", "C", "D", "E", "F"}
// 	numServers := 9

// 	numClusters := 3

// 	result := DistributeClients(clients, numServers, numClusters)
// 	fmt.Println(result)

// 	result2 := DistributeServers(numServers, numClusters)

//		fmt.Println(result2)
//	}
//
// Example to populate the server-to-cluster mapping
// func main() {
// 	// Example cluster mapping
// 	clusterMap := map[int]*ClusterInfo{
// 		1: {
// 			leader:  "S1",
// 			servers: []string{"S1", "S2", "S3"},
// 		},
// 		2: {
// 			leader:  "S4",
// 			servers: []string{"S4", "S5", "S6"},
// 		},
// 		3: {
// 			leader:  "S7",
// 			servers: []string{"S7", "S8", "S9"},
// 		},
// 	}

// 	// Create a client-side server mapping
// 	clientMapping := NewServerMapping("Client1", "localhost:4000", clusterMap)

// 	// Print client mapping for debug
// 	fmt.Printf("Client Mapping:\nName: %s\nAddress: %s\nCluster Map: %+v\n",
// 		clientMapping.name, clientMapping.addr, clientMapping.clusterMap)
// }
