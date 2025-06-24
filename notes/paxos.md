# 📜 Paxos Consensus Algorithm – Technical Summary

## 🧠 Core Idea
Paxos is a fault-tolerant protocol to achieve consensus (agree on a single value) in **asynchronous**, **fail-stop** distributed systems. It underpins **state machine replication**.

---

## 🔍 Paxos Participants
- **Proposer (Leader):** Proposes values to the group.
- **Acceptor:** Votes on proposed values. A majority of these are needed for agreement.
- **Learner:** Learns the chosen value (not critical for core protocol).

📝 In practice, nodes often play all three roles.

---

## 🔄 Protocol Phases

### 🔹 Phase 1: Prepare
1. **Proposer** selects a unique ballot number `n`.
2. Sends `prepare(n)` to a majority of acceptors.
3. **Acceptors** respond with:
   - A promise not to accept lower ballots.
   - Their most recently accepted ballot/value.

### 🔹 Phase 2: Accept
4. **Proposer** selects a value `v`:
   - If any acceptor returned an accepted value, reuse the one with the highest ballot.
   - Else, propose a new value.
5. Sends `accept(n, v)` to the majority.
6. **Acceptors** accept if `n` ≥ previously promised ballot.

### 🔹 Phase 3: Decision (Async)
7. Once a majority accepts `(n, v)`, the value is **chosen**.
8. Proposer sends `decide(v)` to all.

---

## 🛡️ Safety Properties
- **Only one value** is chosen.
- If a value is chosen, **all future proposals** will also choose that value.
- Majority quorum intersection guarantees safety.
- Acceptors **never accept older ballots** after promising a higher one.

---

## ⚙️ Ballot Numbers
- Lexicographically ordered `<ballot_num, proposer_id>`.
- Must be **monotonic and unique**.
- Used to compare proposal freshness.

---

## 💡 Key Variables
- `BallotNum`: highest promised ballot.
- `AcceptNum`: ballot number of last accepted proposal.
- `AcceptVal`: value of last accepted proposal.

---

## 🧷 Point of No Return
A value `v` is **chosen** once it is accepted by a **majority of acceptors**.
> Even if the proposer crashes, the value will survive future rounds via quorum intersection.

---

## 🔄 Leader Election and Failures
- Any node can try to become leader by issuing a higher ballot number.
- New leaders must recover the previous chosen value (by querying acceptors).
- Paxos **is safe but not always live** (FLP impossibility).
  - Solutions: randomized backoff or designate a stable leader (Multi-Paxos).

---

## 🚀 Optimizations

### ✅ Multi-Paxos
- Phase 1 run only during **leader change** (view change).
- Phase 2 repeated for each command (replication phase).
- Enables high-throughput consensus with low overhead.

### ✅ Flexible Paxos
- Separate quorum configurations for **leader election** and **replication**.
- Only **intersection between them** is needed — allows smaller replication quorums.

---

## 🏗️ Paxos in Real Systems
Used in:
- **Google Spanner, Bigtable (via Chubby)**
- **Amazon DynamoDB**
- **Apache Cassandra, Zookeeper (inspired variants)**
- **Hyperledger Fabric**
- **Microsoft Azure’s internal services**
- **CockroachDB, Etcd (use Raft, which simplifies Paxos)**

---

## 📚 Further Reading
- [Paxos Made Simple – Lamport (2001)](https://lamport.azurewebsites.net/pubs/paxos-simple.pdf)
- [Flexible Paxos – Howard et al. (2017)](https://subs.emis.de/LIPIcs/volltexte/2017/7094/pdf/LIPIcs-OPODIS-2016-25_.pdf)

