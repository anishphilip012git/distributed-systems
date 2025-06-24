# 🔐 PBFT – Practical Byzantine Fault Tolerance

## 🎯 Goal
Enable state machine replication in **asynchronous** networks that tolerate **Byzantine faults** (arbitrary, possibly malicious behavior).

---

## ⚙️ System Model & Assumptions
- Total nodes: `n = 3f + 1` to tolerate `f` Byzantine failures.
- Communication: unreliable, unordered, but eventually delivered.
- Fault model: Byzantine — nodes can lie, collude, or go silent.
- All messages are **authenticated** via digital signatures or MACs.

---

## 🧱 Key Protocol Design
- **Three-phase protocol** to ensure agreement across honest replicas.
- **Quorum size: 2f + 1** (majority of honest nodes despite up to `f` bad ones).
- **Views**: Each view has a designated **primary** (leader), others are **backups**.
- **Primary = view_number % n**.

---

## 🔄 Normal Operation Phases

### 🧾 1. Pre-Prepare
- Client sends signed request `⟨REQUEST, op, timestamp, client_id⟩`.
- Primary assigns a sequence number and sends `⟨PRE-PREPARE, v, n, digest⟩ₛₚ`.

### 🛠️ 2. Prepare
- Backups validate and multicast `⟨PREPARE, v, n, digest, i⟩ₛᵢ`.
- Prepared if node has pre-prepare and **2f matching prepares**.

### 🔐 3. Commit
- On prepared state, send `⟨COMMIT, v, n, digest, i⟩ₛᵢ`.
- Committed when **2f + 1 commit messages** match.

### ✅ Execute & Reply
- Replica executes in sequence order.
- Sends signed `⟨REPLY, v, timestamp, client_id, replica_id, result⟩ₛᵢ`.
- Client accepts when **f + 1 matching replies** are received.

---

## 🛑 Safety & Liveness

### 🔒 Safety
- Uses quorum intersection of size `2f + 1` to guarantee one correct replica overlaps across quorums.
- All non-faulty replicas execute **same request in same order**.

### ⚡ Liveness
- Achieved via **view-change** if primary suspected to be faulty.
- New view initiated by `VIEW-CHANGE` messages and acknowledged by new primary via `NEW-VIEW`.

---

## 🔁 View Change Protocol

### Trigger
- Replica timer expires without progress → send `VIEW-CHANGE` to all.

### Messages
- `VIEW-CHANGE`: Contains latest checkpoint, prepared certs.
- `NEW-VIEW`: Sent by new primary with `2f + 1` view-change proofs.

### Recovery
- Backups re-prepare and re-commit using `O` set in new-view.

---

## ♻️ Garbage Collection
- Periodic checkpoints via `⟨CHECKPOINT, seq, digest, id⟩ₛᵢ`.
- If `2f + 1` replicas agree, prune older logs.
- Low/high water marks bound sequence space and prevent DoS by faulty primaries.

---

## ⚖️ Comparison with Paxos

| Feature             | Paxos                     | PBFT                           |
|---------------------|---------------------------|--------------------------------|
| Failure Model       | Crash faults              | Byzantine faults               |
| Nodes Needed        | 2f + 1                    | 3f + 1                         |
| Quorum Size         | f + 1                     | 2f + 1                         |
| Message Complexity  | O(n)                      | O(n²)                          |
| Leader Election     | Repeated or stable        | View-based, rotates on failure |
| Authentication      | Optional                  | Required (MACs/signatures)     |
| Use Case            | Reliable consensus        | Byzantine-safe systems (e.g., blockchain) |

---

## 🔐 Threshold Signatures (Optimization)

### ✅ What It Is
- A **(t, n) threshold signature** scheme allows `t` out of `n` participants to produce a valid signature that’s **as short and verifiable as a single signature**.
- Used to **aggregate multiple replica votes** (prepare/commit) into a **single compact proof**.

### 🧠 Impact
- **Reduces bandwidth**: replaces O(n²) messages with O(n) aggregate signatures.
- **Simplifies client verification**: fewer signatures to validate.
- **Popular schemes**: BLS signatures (Boneh–Lynn–Shacham), RSA-based threshold signatures.

### 🚧 Caveats
- Computationally more expensive.
- Requires secure Distributed Key Generation (DKG) at setup.

---

## 🛠️ Other Optimizations

### 1. **Batched Requests**
- Aggregate multiple client operations per pre-prepare → amortize communication.
- Improves throughput, especially for small ops.

### 2. **Speculative Execution**
- Execute in parallel while waiting for commit quorum (used in Zyzzyva).
- Speeds up latency but requires rollback on mis-speculation.

### 3. **Lazy Replication**
- Delay sending commit/prepare messages to tail replicas (if quorum already satisfied).
- Saves bandwidth with small impact on resilience.

### 4. **Checkpointing**
- Periodically commit state + discard old logs.
- Avoids log bloat; uses `CHECKPOINT` messages.

### 5. **Watermark Windows**
- Bounds on sequence numbers: [low, high] to prevent resource exhaustion from faulty primaries sending extreme values.

### 6. **View Change Batching**
- Batch requests in-flight across view changes; prevents re-execution.
- Helps minimize disruption on leader change.

---

## 🧪 Used In
- **Blockchain Platforms**: Hyperledger Fabric, Tendermint, Zilliqa
- **Secure Replication**: Zyzzyva, BFT-SMaRt, UpRight
- **Byzantine-tolerant storage and databases**

---

## 📚 References
- Castro, Miguel, and Barbara Liskov. "Practical Byzantine Fault Tolerance." (OSDI 1999)
- [PBFT (OSDI '99)](https://pmg.csail.mit.edu/papers/osdi99.pdf)
- [Zyzzyva (OSDI '07)](https://www.usenix.org/legacy/events/osdi07/tech/full_papers/kotla/kotla.pdf)
- [BFT-SMaRt Library](https://github.com/bft-smart/library)

