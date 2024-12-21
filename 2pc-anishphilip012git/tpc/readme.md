# README for Fault-Tolerant Distributed Transaction Processing System

This project implements a **fault-tolerant distributed transaction processing system** for a banking application. It utilizes a **distributed consensus architecture** built with the Paxos and Two-Phase Commit (2PC) protocols, supporting both **intra-shard** and **cross-shard transactions**. The system is implemented in Go, featuring robust crash recovery, performance monitoring, and configurable cluster setups for enhanced scalability.

---

## Features

- **Modified Paxos Protocol**: Achieves consensus for intra-shard transactions within a cluster.
- **Two-Phase Commit Protocol**: Coordinates cross-shard transactions between clusters.
- **gRPC Communication**: Facilitates communication between servers using Protocol Buffers (proto).
- **Interactive Command Menu**: Provides options for transaction processing and monitoring via text-based input.
- **Performance Metrics**: Measures system throughput and latency.
- **Configurable Clusters**: Supports user-defined configurations for the number of clusters and servers per cluster.

---

## Installation and Setup

1. Clone the repository:
   ```bash
   git clone git@github.com/F24-CSE535/2pc-<YourGithubUsername>.git
   cd 2pc-<YourGithubUsername>
   ```

2. Install dependencies:
   ```bash
   go get -a -v
   ```

3. Build and run the program:
   ```bash
   go run .
   ```

4. The program logs outputs to both the console and a file named `output_<timestamp>.log`.

---

## Interactive Command Menu

The program features an interactive menu with commands accessible either by their **name** or **serial number**:

```plaintext
1. PrintDB <server> or all
2. Performance <server> or all
3. NextIteration
4. BalanceMap <client>
5. PrintLog <server> or all
```

For example:
- Use `NextIteration` or `3` to proceed to the next test case.
- Similarly, `PrintDB all` or `1 all` displays the database state for all servers.

---

## CSV Input File

The system reads transactions from a `.csv` file, with each row representing a test case. The file includes the following columns:

1. **Set Number**: Identifies the test case.
2. **Transactions**: List of transactions in the format `(Sender, Receiver, Amount)`.
3. **Live Servers**: List of active servers for the test case.
4. **Contact Servers**: Servers attempting to become leaders.

Example:
```csv
Set Number, Transactions, Live Servers, Contact Servers
1, "(21,700,2)", [S1,S2,S4,S6,S8,S9], [S1,S4,S8]
2, "(702,1301,2)", [S1,S2,S3,S5,S6,S8,S9], [S3,S6,S8]
```

---

## Command Help

### 1. **PrintDB `<server>` or `all`**  
- **Description**: Displays the current state of the database for the specified server(s).  
- **Usage**: 
  ```bash
  PrintDB S1
  PrintDB all
  ```
  or  
  ```bash
  1 S1
  1 all
  ```
- **Purpose**: Validates the state of the database and ensures data consistency.

---

### 2. **Performance `<server>` or `all`**  
- **Description**: Reports throughput and latency for the specified server(s).  
- **Usage**: 
  ```bash
  Performance S2
  Performance all
  ```
  or  
  ```bash
  2 S2
  2 all
  ```
- **Purpose**: Measures and evaluates system performance.

---

### 3. **NextIteration**  
- **Description**: Proceeds to the next test case or transaction set.  
- **Usage**: 
  ```bash
  NextIteration
  3
  ```
- **Purpose**: Facilitates testing or simulating multiple rounds of operations.

---

### 4. **BalanceMap `<client>`**  
- **Description**: Displays the balance of a specific client across servers.  
- **Usage**: 
  ```bash
  BalanceMap 123
  ```
  or  
  ```bash
  4 123
  ```
- **Purpose**: Verifies account balances across the distributed system.

---

### 5. **PrintLog `<server>` or `all`**  
- **Description**: Displays transaction logs for the specified server(s).  
- **Usage**: 
  ```bash
  PrintLog S3
  PrintLog all
  ```
  or  
  ```bash
  5 S3
  5 all
  ```
- **Purpose**: Tracks transactions and debugging events.

---

## System Architecture

- **Clusters and Shards**:  
  - Data is divided into 3 shards (`D1`, `D2`, `D3`) and replicated across 9 servers organized into 3 clusters:
    - Cluster 1 (C1) handles `D1` (IDs 1–1000).
    - Cluster 2 (C2) handles `D2` (IDs 1001–2000).
    - Cluster 3 (C3) handles `D3` (IDs 2001–3000).

- **Consensus Protocols**:  
  - **Modified Paxos Protocol**: For intra-shard transactions within clusters.
  - **Two-Phase Commit Protocol**: For cross-shard transactions between clusters.

- **Fault Tolerance**: Each shard is replicated across all servers in its cluster to ensure no data loss in case of server crashes.

---

## Implemented Features

1. **Distributed Consensus**: Achieves fault-tolerant consensus using Paxos and 2PC protocols.
2. **gRPC Communication**: Inter-server communication implemented with Protocol Buffers.
3. **Configurable Clusters**: Users can define the number of clusters and servers per cluster, enabling scalable system configurations.
4. **Performance Metrics**: Detailed throughput and latency measurements for transaction processing.
5. **Crash Recovery**: State synchronization for failed servers to maintain consistency.

---

## Sample Logs and Testing

1. **Sample Logs**:  
   - **SampleOutput.log**: Demonstrates system functionality, including consensus, transaction processing, and performance metrics.
   - **CrashSimulation.log**: Logs showing system behavior during a simulated server crash.

2. **Crash Testing**:  
   Simulate server crashes to test fault tolerance and recovery:
   ```bash
   CrashServer S4
   ```

---

## Bonus Features

### **Configurable Clusters**
- Users can configure the number of clusters and servers per cluster. This allows for system scalability based on the application's requirements.
- Example:
  - Initially: 3 clusters with 3 servers each.
  - Updated configuration: 4 clusters with 5 servers each.

---

## References

1. "Paxos Made Simple" by Leslie Lamport  
2. [gRPC-Go Documentation](https://github.com/grpc/grpc-go)  
3. ChatGPT (for structuring and debugging help)  

