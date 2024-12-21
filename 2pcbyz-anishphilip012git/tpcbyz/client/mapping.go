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
