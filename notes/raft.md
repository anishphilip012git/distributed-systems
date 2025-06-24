# 🛠️ Raft Consensus Algorithm – 1-Page Summary

## 🎯 Goal
A consensus algorithm for **replicated state machines** that is **understandable**, **fault-tolerant**, and **strongly consistent** in distributed systems.

---

## 🧩 Core Concepts

### 🧠 Roles
- **Leader** – handles all client requests, coordinates log replication.
- **Follower** – passive, responds to leader/candidate RPCs.
- **Candidate** – initiates election on timeout.

### 📅 Terms
- Logical clock periods for elections and leadership.
- At most 1 leader per term.
- Triggers: election timeout or crash.

---

## 🔄 Leader Election
1. Follower times out → becomes **candidate**.
2. Increments term, votes for self, sends `RequestVote` RPCs.
3. Wins with **majority** vote → becomes leader.
4. Heartbeats (empty `AppendEntries`) maintain leadership.

📌 **Election safety**: only one winner per term via persistent vote record and majority overlap.

---

## 📦 Log Replication
- Clients send commands to **leader**.
- Leader appends to log, sends `AppendEntries` to followers.
- Entry is **committed** once stored on a **majority** and includes at least one entry from the current term.
- Leader notifies followers of committed entries.

---

## ✅ Safety Guarantees
- **Log Matching Property**: same index/term → same command.
- **Committed entries** must be present in **future leaders**.
- **Leaders never overwrite** committed entries.

---

## 🧪 Consistency & Recovery
- **AppendEntries includes** `<prevIndex, prevTerm>` for consistency check.
- Followers **reject mismatches** → leader decrements `nextIndex`, retries.
- Follower logs are repaired by deleting conflicts and appending missing entries.

---

## 🔁 Leader Change
- New leader starts with its log as “source of truth”.
- Followers sync with leader’s log.
- Old leaders are neutralized via **term check** on every RPC.

---

## 👨‍💻 Client Interaction
- Clients send command to leader → waits for commit + execution.
- For idempotency: each command tagged with unique ID (deduplicated on retry).

---

## ⚙️ Configuration Changes
- Handled via **joint consensus** (overlap old + new configs).
- Requires commitment from both majorities.
- Prevents split-brain or conflicting decisions.

---

## 🧱 Used In
Etcd, CockroachDB, MongoDB, Redpanda, TiDB, ScyllaDB, Neo4j, Kafka (KRaft), and more.

📚 [raft.github.io](https://raft.github.io) | [Ongaro & Ousterhout paper](https://www.usenix.org/system/files/conference/atc14/atc14-paper-ongaro.pdf)
