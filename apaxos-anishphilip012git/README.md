# Paxos Distributed System with GRPC Communication

This project implements a variant of the Paxos consensus protocol in Go. It is designed to handle distributed transactions across multiple servers, where the servers reach consensus using Paxos if a client has insufficient funds for a transaction. The project supports logging, crash recovery, file-based datastore, and communication between servers using gRPC.

## Features

- **Leader Election and Consensus**: Paxos consensus ensures fault tolerance and consistency across servers.
- **gRPC Communication**: The system uses Protocol Buffers (proto) for gRPC-based communication between servers.
- **Logging**: Log output is directed to both the console and a file, with each run generating a file named `output_<timestamp>.log`.
- **Crash Recovery**: Simulates server crashes and handles recovery with distributed state synchronization.
- **Datastore**: A file-based key-value store (BadgerDB) is used to persist transactions efficiently.

## Installation and Setup

1. Clone the repository and navigate to the project directory.
   ```bash
   git clone https://github.com/your-repo/paxos.git
   cd paxos1
   ```

2. Install all dependencies:
   ```bash
   go get -a -v
   ```

3. Run the project:
   ```bash
   go run .
   ```

4. The program will log outputs to both the console and a file named `output_<timestamp>.log`. The file includes logs with timestamps in the format `ddmmyyyy_hh_mm`.

## CSV Input File

The project reads a CSV file named `test.csv` by default. This file contains transaction data for testing the distributed system. You can change the CSV file name in `main.go` if needed.

## Running the System

When you run the system, a menu appears with the following options:

```plaintext
2024/10/17 11:10:31 1. PrintBalance <client> or all
2024/10/17 11:10:31 2. PrintLog <server> or all
2024/10/17 11:10:31 3. PrintDB <server> or all
2024/10/17 11:10:31 4. Performance <server> or all
2024/10/17 11:10:31 5. NextIteration
2024/10/17 11:10:31 6. CrashServer <server>
2024/10/17 11:10:31 7. BalanceMap <server>

```
and a debug option `all` - which prints output of command 1,2,3,4,7 for all server

### General Option: `all`
- A general option `all` is available for commands like `PrintBalance`, `PrintLog`, `PrintDB`, `Performance` and `BalanceMap` . 
- For example, the following command prints the balance for all clients:
  ```bash
  PrintBalance all
  ```
  Similarly, you can print logs, the database state, and performance metrics for all servers with:
  ```bash
  PrintLog all
  PrintDB all
  Performance all
  ```

- **Client and Server options**: You can specify individual clients and servers with values `["S1", "S2", "S3", "S4", "S5"]`, or change these in the `AllServers` list in `main.go`.

### Example Commands
- To print the balance of a specific client:
  ```bash
  PrintBalance S1
  ```
  or for all clients:
  ```bash
  PrintBalance all
  ```

- To print logs from a specific server:
  ```bash
  PrintLog S1
  ```
  or for all servers:
  ```bash
  PrintLog all
  ```
## Command help

### **1. PrintBalance <client> or all**  
- **Description**:  
  This command prints the current balance of a specified client or **all clients** if `all` is provided.  
- **Usage**:  
  - `PrintBalance client1`: Prints the balance for `client1`.  
  - `PrintBalance all`: Prints the balances for all clients.  
- **Purpose**: Helps monitor the financial state of clients in the system.

---

### **2. PrintLog <server> or all**  
- **Description**:  
  Displays the transaction or event logs for a specified server or **all servers**.  
- **Usage**:  
  - `PrintLog server1`: Shows logs for `server1`.  
  - `PrintLog all`: Shows logs for all servers.  
- **Purpose**: Useful for **debugging** and tracking transactions handled by servers.

---

### **3. PrintDB <server> or all**  
- **Description**:  
  Displays the current state of the **database** for a specified server or **all servers**.  
- **Usage**:  
  - `PrintDB server1`: Prints the database records for `server1`.  
  - `PrintDB all`: Prints the records from all servers.  
- **Purpose**: Ensures that the state of the data is correct across servers.

---

### **4. Performance <server> or all**  
- **Description**:  
  This command reports the **performance metrics** (e.g., latency, throughput) of a given server or all servers.  
- **Usage**:  
  - `Performance server1`: Prints performance metrics for `server1`.  
  - `Performance all`: Prints metrics for all servers.  
  - I have calculated 2 metrics: 
      - RPC calls and time to process each.
      - Transaction processing time for each transaction and average time to process each transaction.

- **Purpose**: Helps in evaluating and improving the **performance** of servers.

---

### **5. NextIteration**  
- **Description**:  
  Moves the system to the **next iteration** of its processing cycle, potentially triggering new transactions or events.  
- **Usage**:  
  - `NextIteration`: Proceeds to the next iteration.  
- **Purpose**: Facilitates testing or simulating **multiple rounds** of transactions or operations.

---

### **6. CrashServer <server>**  
- **Description**:  
  Simulates the **crash** of a specific server to test the system’s **fault tolerance** and recovery mechanisms.  
- **Usage**:  
  - `CrashServer server1`: Crashes `server1`.  
- **Purpose**: Helps evaluate how well the system can handle server **failures**.

---

### **7. BalanceMap <server>**  
- **Description**:  
  Displays the **balance map** of the given server, showing the balances of all clients known to that server.  
- **Usage**:  
  - `BalanceMap server1`: Shows the balance map for `server1`.  
- **Purpose**: Provides visibility into how balances are distributed across different clients on a specific server.
- In this,  we retrieve the current balance of a specified client across different servers, taking into account both the datastore and local logs, all without requiring consensus among them.It effectively reads the client’s balance on each server and then prints the aggregated total.   


## Implemented Features

### Modified Paxos Protocol
- The system uses a modified Paxos protocol with a **catch-up mechanism** similar to Raft. It ensures servers can recover from failures and maintain consensus using `lastCommittedBallotNumber`.

### gRPC Communication with Proto
- The system uses gRPC for communication between servers, with Protocol Buffers (proto) files defining the communication protocol.

### Bonus Feature 1: Badger Datastore
- The project uses BadgerDB, a file-based key-value store in Go, to persist transactions and ensure durability.

### Bonus Feature 2: Crash Recovery
- The system handles crash failures, simulating server crashes and continuing with consensus where possible.
- Example: If server `S4` crashes after set 1, the system continues to process sets 2 and 3. However, from set 4 onwards, it skips processing due to the crash.


### Bonus Feature 3: Balance of all servers on all servers
- As required in bonus point 3 the system maintains balance of all servers on all servres, they are synced using commit logs.
however for the local servr the balances are synced immediately. This implies for all other servers we get updated state only after successful consensus.
- You can check it using `BalanceMap` option and this can be compared with `PrintBalance` option.
- You can check the run logs for it in `Bonus3 with all servers.log`.


## Sample Output Files

Two sample log files are provided:
- **SampleOutput1.log**: A complete run of the system, showing consensus, transaction processing, and log output.
- **SampleOutput2.log**: Another example log demonstrating system functionality In ths i run the entire test case , and use `all` command to print all paramters after each iteration.

## Testing

- **Crash Testing**: You can simulate server crashes by using the command `CrashServer <server>`. For example:
  ```bash
  CrashServer S4
  ```
  - **BONUS2 Crash failure and db usage.log** : In this file I hvae shown the output of one such simulation. You will see that set 2 and set 3 continue processing normally, while set 4 is skipped due to the crash. We can still obtain its log, and after that when we actually want to use S4 , I skip the transactions because S4 is crashed.

### References :
1. Chatgpt (Free version): for basic code setup , debugging and organisation help .
1. `Paxos Made Simple` paper.
1. [GRPC-go](https://github.com/grpc/grpc-go)
1. [Protobuf](https://protobuf.dev/)