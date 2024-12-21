# README for Fault-Tolerant Byzantine Transaction Processing System

This project implements a **fault-tolerant distributed transaction processing system** for a banking application. The system utilizes a combination of **Linear Practical Byzantine Fault Tolerance (PBFT)** and **Two-Phase Commit (2PC)** protocols, supporting both **intra-shard** and **cross-shard transactions**. It is designed to operate in Byzantine environments, where up to one server per cluster may fail arbitrarily. The implementation ensures **data consistency**, **fault tolerance**, and **scalability** through replication and efficient consensus mechanisms.

---

## Features

- **Linear PBFT Protocol**: Used for intra-shard transactions to achieve Byzantine fault-tolerant consensus within a cluster.
- **Two-Phase Commit Protocol**: Coordinates cross-shard transactions between clusters, ensuring atomicity and consistency across shards.
- **gRPC Communication**: Facilitates inter-server communication using Protocol Buffers.
- **Byzantine Fault Tolerance**: Handles failures and ensures correctness even when some servers exhibit Byzantine behavior.
- **Interactive Command Menu**: Allows real-time transaction management and system monitoring via a text-based interface.
- **Performance Metrics**: Measures throughput and latency for comprehensive performance evaluation.
- **Configurable Shards and Clusters**: Supports dynamic configurations of data shards and clusters to accommodate varying scales.

---

## Architecture and Data Partitioning

- The system partitions a dataset of **3000 items** across **3 shards (D1, D2, D3)**:
  - **Cluster 1 (C1)**: Handles data items with IDs `1–1000`.
  - **Cluster 2 (C2)**: Handles data items with IDs `1001–2000`.
  - **Cluster 3 (C3)**: Handles data items with IDs `2001–3000`.
- Each shard is replicated across **4 servers per cluster** to ensure fault tolerance. The system can tolerate up to one Byzantine server per cluster.

---

## Transaction Types

1. **Intra-Shard Transactions**:
   - Handled by the Linear PBFT protocol within a single cluster.
   - Example: Transferring funds between accounts managed by the same shard.

2. **Cross-Shard Transactions**:
   - Coordinated using the Two-Phase Commit (2PC) protocol across multiple clusters.
   - Example: Transferring funds between accounts managed by different shards.

---

## Protocols Overview

### Linear PBFT Protocol (Intra-Shard)
- **Primary Leader**:
  - The client sends the transaction request `(Sender, Receiver, Amount)` to the primary server of the shard.
  - The leader checks:
    1. No locks on the involved data items.
    2. Sufficient balance in the sender's account.
  - Locks are acquired, and the leader initiates consensus by sending a **Pre-Prepare** message.
- **Backup Servers**:
  - Validate and acquire locks before proceeding.
  - Respond with **Prepare** and **Commit** messages.
- **Outcome**:
  - The transaction is executed, balances are updated, and locks are released.
  - Results are returned to the client.

### Two-Phase Commit Protocol (Cross-Shard)
- **Coordinator Cluster**:
  - Initiates Linear PBFT for the transaction's ordering and lock acquisition.
  - If consensus is achieved, sends a **Prepare** message to the participant cluster.
- **Participant Cluster**:
  - Validates the request and initiates Linear PBFT to achieve consensus.
  - Sends a **Prepared** or **Abort** response to the coordinator cluster.
- **Commit Phase**:
  - The coordinator determines the final outcome (commit/abort) based on participant responses.
  - Updates the datastore and releases locks.
  - Results are returned to the client.

---

## Installation and Setup

1. Clone the repository:
   ```bash
   git clone git@github.com:F24-CSE535/2pcbyz-<YourGithubUsername>.git
   cd 2pcbyz-<YourGithubUsername>
   ```

2. Install dependencies:
   ```bash
   go get -a -v
   ```

3. Build and run the program:
   ```bash
   go run .
   ```

4. Logs will be output to both the console and a timestamped log file, e.g., `output_<timestamp>.log`.

---

## Interactive Command Menu

Use the interactive command menu for real-time system interaction. Commands can be accessed by name or serial number:

```plaintext
1. PrintDB <server> or all
2. Performance <server> or all
3. NextIteration
4. BalanceMap <client>
5. PrintLog <server> or all
```

Here's an updated version of the README to include the **Contact Servers** column in the file format description:

---

## Input File Format

The system reads transactions from a CSV file, where each row represents a test case. The file includes the following columns:

1. **Set Number**: Identifies the test case.  
2. **Transactions**: A list of transactions in the format `(Sender, Receiver, Amount)`.  
3. **Live Servers**: Active servers for the test case.  
4. **Byzantine Servers**: Servers exhibiting Byzantine behavior during the test case.  
5. **Contact Servers**: Servers attempting to become leaders (e.g., initiators of PBFT or 2PC coordination).

### Example Format
```csv
Set Number, Transactions, Live Servers, Byzantine Servers, Contact Servers
1, "(21,700,2)", [S1,S2,S4,S6], [S9], [S1,S6]
2, "(702,1301,2)", [S1,S2,S3,S5,S6], [S8], [S3,S5]
```

### Explanation:
- **Set Number**: Sequential number identifying the transaction set.
- **Transactions**: List of `(Sender, Receiver, Amount)` tuples representing money transfers.
- **Live Servers**: Active servers participating in processing the transaction.
- **Byzantine Servers**: Servers simulating Byzantine faults for testing.
- **Contact Servers**: Servers that initiate consensus protocols (PBFT or 2PC coordination).

---

## Command Details

### PrintDB `<server>` or `all`
- Displays the current state of the database.
- Usage: `PrintDB S1` or `1 all`.

### Performance `<server>` or `all`
- Reports throughput and latency.
- Usage: `Performance all` or `2 S1`.

### NextIteration
- Moves to the next transaction set.
- Usage: `NextIteration` or `3`.

### BalanceMap `<client>`
- Displays the balance of a specific client.
- Usage: `BalanceMap 123` or `4 123`.

### PrintLog `<server>` or `all`
- Displays logs of committed transactions.
- Usage: `PrintLog all` or `5 S3`.

---

## Testing and Performance

1. **Throughput**: Transactions per second.
2. **Latency**: Average processing time per transaction.
3. **Fault Simulation**:
   - Simulate Byzantine failures: `in test case file`.

---

## References

1. "Paxos Made Simple" by Leslie Lamport  
2. "PBFT Explained" by Miguel Castro and Barbara Liskov  
3. [gRPC-Go Documentation](https://github.com/grpc/grpc-go)  

---

Let me know if you'd like further refinements!