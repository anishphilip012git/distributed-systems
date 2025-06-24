# Concurrency Control in Distributed Databases

Mohammad Javad Amiri  
CSE 590: Distributed and Decentralized Data Management

---

## Implementing Transactions

- **Atomicity and Durability**:
  - All operations in a transaction either complete or none do.
  - Use **logging** to ensure changes can be undone or redone.

- **Isolation**:
  - Transactions observe a consistent view of the DB.
  - Achieved via **locking** mechanisms.

---

## Isolation

- Transactions should appear as if run alone.
- **Interleaving** by the DBMS must preserve correctness.
- Requires **concurrency control protocols**.

---

## Motivation Example

- T1: Transfers $100 from A to B.
- T2: Applies 6% interest to both accounts.
- Total must be $2120.
- Execution order affects the outcome, but result must match **some serial order**.

---

## Interleaving

- Done to improve throughput and responsiveness.
- Interleaving can cause **inconsistencies** without control.

---

## Serializability

- A **schedule** is a series of operations.
- A schedule is **serializable** if it is equivalent to some serial execution.
- Two types:
  - **Conflict Serializability** (used in practice)
  - **View Serializability** (theoretical)

---

## Conflicting Operations

- Conflict if:
  - Different transactions
  - Operate on same object
  - At least one is a **write**
- Types:
  - **Read-Write (RW)**
  - **Write-Read (WR)**
  - **Write-Write (WW)**

---

## Conflict Serializability

- Schedule is serializable if **conflict equivalent** to a serial one.
- Achieved by **reordering** non-conflicting operations.
- Use **dependency graphs** to check:
  - Nodes = Transactions
  - Edge T1 → T2 if T1’s operation conflicts and occurs before T2’s
  - Acyclic = Serializable

---

## Concurrency Control Protocols

1. **Pessimistic**: Assume conflict, prevent it.
   - **Two-Phase Locking (2PL)**:
     - Phase 1: Acquire all locks
     - Phase 2: Release locks
     - Guarantees serializability
2. **Optimistic**: Assume no conflict, check at commit.
   - Timestamp ordering
   - Validation-based
   - MVCC (e.g. Snapshot Isolation)

---

## Locking

- **S-LOCK**: Shared (read)
- **X-LOCK**: Exclusive (write)
- Lock manager tracks:
  - Held locks
  - Waiters
- Locks ensure correct interleaving.

---

## Problems with Basic Locking

- May still allow:
  - **Dirty reads**
  - **Cascading aborts**
- Fix: **Strict 2PL**:
  - Holds all locks until commit
  - Prevents dirty reads
  - Simplifies rollback

---

## Deadlocks

- Circular waits on locks.
- **Detection**:
  - Build wait-for graph
  - Detect cycles
  - Abort victim
- **Prevention**:
  - Priority-based aborts (e.g. older txn wins)

---

## Isolation Levels

| Level              | Dirty Read | Unrepeatable Read | Phantom |
|--------------------|------------|--------------------|---------|
| SERIALIZABLE        | No         | No                 | No      |
| REPEATABLE READ     | No         | No                 | Maybe   |
| READ COMMITTED      | No         | Maybe              | Maybe   |
| READ UNCOMMITTED    | Maybe      | Maybe              | Maybe   |

- Tradeoff: **Concurrency vs Correctness**

---

## Snapshot Isolation (MVCC)

- Transaction sees snapshot from start.
- Writers get new versions.
- Read-only transactions never block.
- If two transactions update the same item → conflict on commit.

---

## Timestamp Ordering

- Each txn gets timestamp.
- Ensure reads/writes happen in timestamp order.
- Abort if a txn reads/writes out of order.

---

## Optimistic Concurrency Control (OCC)

- Phases:
  1. **Read**: Track read/write sets
  2. **Validate**: Check conflicts
  3. **Write**: If valid, apply updates
- Abort and retry if conflicts found.

---

## Summary

- Goal: Ensure **correctness + concurrency**
- Techniques:
  - 2PL, MVCC, Timestamp Ordering, OCC
- Tools:
  - Locks, Dependency Graphs, Logs

---

## Resources

- Bernstein et al. *Concurrency Control and Recovery in DB Systems*  
  https://www.sigmod.org/publications/dblp/db/books/dbtext/bernstein87.html
