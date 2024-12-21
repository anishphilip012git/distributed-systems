package server

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"sync"
	. "tpcbyz/security"
)

const (
	PRIMARY = "Primary"
	BKP     = "Backup"
)

var (
	ClientKeys KeyPair
)

var (
	mu            sync.Mutex
	PeerAddresses map[string]string
	PeerIdMap     map[string]string
	AllServers    []string
	LiveSvrMap    = make(map[string]bool)
	ByzSvrMap     = make(map[string]bool)
	// ServerRoles    map[string]string
	ServerClusters map[string]int          // Maps server ID to cluster ID
	ClusterMapping map[int]*ClusterDetails // Maps cluster ID to details about the cluster
	ShardMap       []int
)

// ClusterDetails holds information about a cluster
type ClusterDetails struct {
	Leader  string   // Leader server ID for the cluster
	Servers []string // List of servers in the cluster
}

// ServerKeys holds the key pairs of the servers.
var PrivateKeys map[string]*ecdsa.PrivateKey // Maps server ID to its KeyPair
var PublicKeys map[string][]byte

func InitByzSvrMap(byzServer []string) {
	for serverId, _ := range LiveSvrMap {
		ByzSvrMap[serverId] = false
	}
	for _, svr := range byzServer {
		ByzSvrMap[svr] = true
	}
	log.Println("ByzSvrMap", ByzSvrMap)
}

// distributeServers distributes servers across clusters
func distributeServers(numServers int, numClusters int) map[int]int {
	serverToCluster := make(map[int]int)
	serversPerCluster := numServers / numClusters
	extraServers := numServers % numClusters

	serverIndex := 1
	for cluster := 1; cluster <= numClusters; cluster++ {
		count := serversPerCluster
		if extraServers > 0 {
			count++
			extraServers--
		}

		for i := 0; i < count; i++ {
			if serverIndex <= numServers {
				serverToCluster[serverIndex] = cluster
				serverIndex++
			}
		}
	}

	return serverToCluster
}

// GenerateServerConfig initializes the server configuration
func GenerateServerConfig(totalServers int, numClusters int, f int) {
	mu.Lock()
	defer mu.Unlock()
	F = f
	// Clear existing maps and slices
	PeerAddresses = make(map[string]string)
	PeerIdMap = make(map[string]string)
	AllServers = make([]string, 0, totalServers)
	LiveSvrMap = make(map[string]bool)
	// ServerRoles = make(map[string]string)
	ServerClusters = make(map[string]int)
	ClusterMapping = make(map[int]*ClusterDetails)

	// Get server-to-cluster mapping
	serverToCluster := distributeServers(totalServers, numClusters)

	for serverCount := 1; serverCount <= totalServers; serverCount++ {
		serverID := fmt.Sprintf("S%d", serverCount)
		address := fmt.Sprintf("localhost:500%02d", serverCount)

		// Determine cluster ID
		clusterID := serverToCluster[serverCount]

		// Determine role: primary for the first server in a cluster, backup for others
		// role := BKP
		if _, exists := ClusterMapping[clusterID]; !exists {
			ClusterMapping[clusterID] = &ClusterDetails{
				Leader:  serverID,
				Servers: []string{},
			}
			// role = PRIMARY
		}

		// Add the server to the cluster
		ClusterMapping[clusterID].Servers = append(ClusterMapping[clusterID].Servers, serverID)

		// Populate the global variables
		PeerAddresses[serverID] = address
		PeerIdMap[address] = serverID
		AllServers = append(AllServers, serverID)
		LiveSvrMap[serverID] = true
		// ServerRoles[serverID] = role
		ServerClusters[serverID] = clusterID
	}

	// Print the configurations
	fmt.Println("PeerAddresses:", PeerAddresses)
	fmt.Println("PeerIdMap:", PeerIdMap)
	fmt.Println("AllServers:", AllServers)
	fmt.Println("LiveSvrMap:", LiveSvrMap)
	// fmt.Println("ServerRoles:", ServerRoles)
	fmt.Println("ServerClusters:", ServerClusters)

	GenerateKeyPairsForAllServers()
	// Print Cluster Mapping
	for clusterID, details := range ClusterMapping {
		fmt.Printf("Cluster %d: Leader: %s, Servers: %v\n", clusterID, details.Leader, details.Servers)
	}
}

// Example usage
func GenerateKeyPairsForAllServers() {
	// List of server IDs
	// serverIDs := []string{"server1", "server2", "server3"}

	// Shared variable to store key pairs
	PublicKeys = map[string][]byte{}
	PrivateKeys = map[string]*ecdsa.PrivateKey{}

	// Generate keys
	for _, serverID := range AllServers {
		privateKey, pubKeyPEM, err := GeneratePubPvtKeyPair(serverID)
		if err != nil {
			log.Printf("Error in Key genration for server %s : %v", serverID, err)
		}
		// Store the key pair in the shared variable
		PrivateKeys[serverID] = privateKey // Store the private key
		PublicKeys[serverID] = pubKeyPEM
	}
}

func UpdateContactServers(contactServers []string) {
	for i, val := range contactServers {
		ClusterMapping[i+1].Leader = val
	}
	log.Println("UpdateContactServers", contactServers)
}

// func main() {
// 	totalServers := 9
// 	numClusters := 3
// 	GenerateServerConfig(totalServers, numClusters)
// }
