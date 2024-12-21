# Linear PBFT Distributed System with gRPC Communication

This project implements a linearized version of the Practical Byzantine Fault Tolerance (PBFT) protocol in Go, using gRPC for inter-server communication. The system aims to achieve secure consensus across distributed servers, ensuring reliability through logging, checkpointing, threshold signatures, and a file-based datastore using BoltDB.

## Features

- **Linear PBFT Consensus**: Implements a linearized PBFT protocol with threshold signatures for secure consensus and state consistency.
- **Checkpointing Mechanism**: Utilizes a checkpointing mechanism to manage memory and ensure all replicas are up-to-date.
- **gRPC Communication**: Inter-server communication is facilitated using Protocol Buffers and gRPC.
- **Logging**: Logs are stored in both console output and a time-stamped log file (`output_<timestamp>.log`).
- **Datastore**: BoltDB, a file-based key-value store, is used for persistent data storage.

## Installation and Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/pbft1.git
   cd pbft1
   ```

2. Install dependencies:
   ```bash
   go get -a -v
   ```

3. Run the project:
   ```bash
   go run .
   ```

   Logs will be generated in both the console and a file named `output_<timestamp>.log` in `ddmmyyyy_hh_mm` format.

## CSV Input File

The project reads a `test.csv` file containing transaction data by default. You can update the CSV file path in `main.go` if needed.

## Running the System

Upon starting, a menu displays various options:

```plaintext
1. PrintBalance <client> or all
2. PrintLog <server> or all
3. PrintDB <server> or all
4. Performance <server> or all
5. NextIteration
6. PrintView <server>
7. BalanceMap <server>
```

### General Option: `all`
Commands like `PrintBalance`, `PrintLog`, `PrintDB`, `Performance`, `PrintView`,and `BalanceMap` accept `all` to perform the action on all servers.

### Example Commands

- Print the balance of all clients:
  ```bash
  PrintBalance all
  ```
- Retrieve logs for a specific server:
  ```bash
  PrintLog S1
  ```
- View database entries of all servers:
  ```bash
  PrintDB all
  ```

## Command Help

### **1. PrintBalance <client> or all**  
Displays the current balance for a specified client or for all clients.

### **2. PrintLog <server> or all**  
Shows the transaction logs for a specific server or all servers for debugging.

### **3. PrintDB <server> or all**  
Prints the state of the database for a specific server or all servers, verifying data consistency.

### **4. Performance <server> or all**  
Reports performance metrics, such as RPC call counts and transaction processing times, for specified servers.

### **5. NextIteration**  
Moves the system to the next processing cycle, triggering new transactions or events.

### **6. BalanceMap <server>**  
Displays the balance map of the specified server, showing all clients' balances on that server.

## Implemented Features

### Linearized PBFT Protocol
The system uses a linear PBFT protocol with threshold signatures to ensure consensus and handle faulty or malicious servers.

### Checkpointing Mechanism
A checkpointing mechanism is used to efficiently manage memory and bring all replicas up-to-date. 
- **Garbage Collection**: Periodically removes data of completed consensus instances to free up storage.
- **State Synchronization**: Restores in-dark replicas (e.g., those with unreliable network connections or affected by leader faults) to the latest system state. Checkpointing occurs after a fixed number of requests (e.g., every 100 requests) in a decentralized manner, without relying on a leader.

### Threshold Signature Scheme (TSS)
The protocol uses threshold signatures for authorization, ensuring that a minimum number of key-share holders (e.g., 2f + 1 out of 3f + 1) is required to authorize transactions.
- **Improved Efficiency**: Threshold signatures reduce message size, as a single combined signature represents the entire group, rather than collecting individual signatures.
  
### gRPC Communication
The system uses Protocol Buffers for efficient and reliable communication between servers.

### Bonus Feature 1: BoltDB Datastore
BoltDB is used as a file-based key-value store for persistent data storage.

### Bonus Feature 2: Distributed Balance Map
Maintains a balance map across all servers, which can be verified by comparing outputs from `BalanceMap` and `PrintBalance` commands.

## Sample Output Files

- **SampleOutput1.log**: Demonstrates the system’s complete functionality, including consensus and transaction processing.
- **SampleOutput2.log**: Illustrates the `all` command functionality across iterations.

## Testing

- **Checkpoint Testing**: Trigger checkpointing by processing a batch of transactions (e.g., 100 requests), which initiates garbage collection and synchronizes any out-of-date replicas.

### References
1. ChatGPT for guidance on code structure and debugging.
2. **PBFT research papers**
3. [gRPC-Go](https://github.com/grpc/grpc-go)
4. [Protocol Buffers Documentation](https://protobuf.dev/)
