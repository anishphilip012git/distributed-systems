# Consistency in Distributed Systems

Mohammad Javad Amiri  
CSE 535: Distributed Systems

---

## Distributed Systems Review

- Ever-growing load  
- Failures are inevitable  
- Consistency is hard  

Goals:
- Consistency semantics  
- Scalability (via Sharding, Replication)  
- Fault tolerance  
- Protocols  

---

## Why Do We Need Consistency Guarantees

- In single-threaded systems: correctness is easy to define.
- In distributed systems:
  - Need to define ordering of operations.
  - Need to define when updates are visible.
  - Tradeoff between performance and correctness.
- **Consistency model** defines these rules.

---

## Consistency Model

- Describes how shared data should appear across distributed processes.
- Determines whether operations by one client are visible to another.
- **Consistency** means all nodes agree on the data.

---

## Consistency Models Spectrum

```

Strong                Weak
Strict > Linearizability > Sequential > Causal > Eventual

```

- Stronger = more guarantees, lower availability/performance.
- Weaker = more performance, risk of anomalies.
- Others: Session, Consistent Prefix

---

## Semantics of Consistency

- Controls:
  - Ordering of reads/writes.
  - Freshness/staleness of reads.
- Affects how replicas synchronize.

---

## Example: Replicated Distributed Shared Memory

- Clients perform writes and reads on shared variable `x`.
- Multiple replicas maintain copies.
- Challenge: what read values are valid?

---

## Strict Consistency

- All operations appear to execute in **real-time order**.
- Read always reflects the most recent write.
- Not practical in asynchronous systems.

---

## Linearizability

- Global order of operations that preserves real-time constraints.
- Read reflects the latest completed write.
- Examples:
  - Paxos, RAFT, PBFT
  - Single-threaded processors

---

## Sequential Consistency

- Operations appear in some global order **preserving program order** per process.
- Real-time constraints **not** enforced.
- Reads may be stale in real time but not logically.
- Appears like a single machine executing threads in some order.

---

## Linearizability vs Sequential Consistency

- Linearizability: Total order + real-time order.
- Sequential: Total order + per-process program order.
- Linearizability ⊂ Sequential Consistency

---

## Causal Consistency

- Operations respect causality:
  - If A → B, all replicas see A before B.
  - Concurrent ops may be seen in different orders.
- No total order required.
- Example: social media replies before tweets.

---

## Eventual Consistency

- All replicas **converge** eventually if no new updates.
- Reads can be stale.
- No guarantees on order or immediate visibility.
- Used in:
  - Amazon Dynamo
  - Git
  - File sync systems

---

## Sequential vs Eventual Consistency

- Sequential: pessimistic — control order during execution.
- Eventual: optimistic — reconcile later.
- Pros: fault-tolerance, parallelism.
- Cons: conflicts, anomalies.

---

## Session Consistency Models

1. **Read-your-writes**: Always see your own writes.
2. **Monotonic reads**: Once seen, never un-seen.
3. **Writes follow reads**: Causal propagation.
4. **Monotonic writes**: Writes ordered per session.

---

## Consistent Prefix Read

- All reads see a prefix of the write history.
- Ensures no out-of-order reads.
- E.g., sports score apps.

---

## CAP Theorem

A distributed system can satisfy **two** of the following at once:

- **Consistency** (Linearizability)
- **Availability** (non-error responses)
- **Partition Tolerance** (tolerate dropped/delayed messages)

---

## CAP Trade-off

- P is inevitable.
- Choices:
  - CA: not partition tolerant
  - CP: not highly available
  - AP: not strongly consistent

---

## Resources

- Herlihy & Wing: "Linearizability: A correctness condition for concurrent objects"  
  [ACM TOPLAS](https://dl.acm.org/doi/abs/10.1145/78969.78972)
```
