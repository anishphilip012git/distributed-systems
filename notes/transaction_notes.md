
# Transactions in Distributed Systems

Mohammad Javad Amiri  
CSE 590: Distributed and Decentralized Data Management

---

## Database

A key component in any distributed application is a (distributed) database that maintains shared state.

### Goals:
- Protect the data in the database.
- Correct querying despite concurrency and failures.
- Handle:
  - **Failures** (partial computations, recovery after crash)
  - **Concurrency** (race conditions, synchronization)

**Core components:**
-  [Concurrency Control](./concurrency_control_notes.md)
- [Recovery](./recovery_notes.md)

---

## Transaction

A **transaction** is an abstraction that encapsulates a unit of work on a database.

- Guarantees **atomic** execution despite failures and **isolation** from concurrent transactions.
- Example: Transferring $10 from A to B:
  1. Read A
  2. Subtract from A
  3. Add to B
  4. Write both

---

## Strawman Solution

- Run transactions **one-by-one**.
- Copy DB before transaction, replace after commit.
- Safe, but poor concurrency and performance.

---

## Problem Statement

- Better: execute **concurrent transactions**.
- Risks:
  - Temporary inconsistency (OK)
  - **Permanent inconsistency** (bad!)
- Need **formal correctness criteria** to validate concurrent execution.

---

## Transaction API

```python
txID = Begin()
Commit(txID)  # All or nothing
Abort(txID)   # Cancel all effects
````

* Commit applies all changes.
* Abort erases effects.

---

## ACID Properties

1. **Atomicity**: All or nothing
2. **Consistency**: DB constraints respected
3. **Isolation**: Transactions don’t interfere
4. **Durability**: Changes persist after commit

---

## Example

TRANSFER(src, dst, x):

```text
1. Read src
2. If src >= x:
3. Subtract x from src
4. Write src
5. Read dst
6. Add x to dst
7. Write dst
```

PRINT\_SUM(acc1, acc2):

```text
1. Read acc1
2. Read acc2
3. Print sum
```

Interleaving without transactions → inconsistency.

---

## With Transactions

TRANSFER:

```text
txID = Begin()
...
if src > x:
  update src/dst
Commit(txID)
```

PRINT\_SUM:

```text
txID = Begin()
Read acc1
Read acc2
Print sum
Commit(txID)
```

---

## Implementing Transactions

### Atomicity & Durability

* **Logging**: log operations before applying (Write-Ahead Logging)

### Isolation

* **Locking**: hold locks on read/write objects until end of transaction

---

## Atomicity in Practice

Only two possible outcomes:

* **Commit**
* **Abort**

DB guarantees one of these happens atomically.

---

## Ensuring Atomicity

### Logging

* Undo/redo via log entries
* Used by almost all DBMSs

### Shadow Paging

* Copy-on-write model (used by CouchDB, LMDB)

---

## Consistency

* DB must model the **real world** and satisfy constraints.
* Transactions must maintain consistency if started in a consistent state.

---

## Durability

* Committed changes survive system crashes.
* Ensured via logs or shadow pages.
* No partial/torn writes.

---

## Isolation

* Users see transaction as if it’s running alone.
* DB interleaves transactions using **concurrency control protocols**.

---

## Resources

* Lampson & Sturgis (1979): Crash Recovery
  [https://web.cs.wpi.edu/\~cs502/cisco11/Papers/LampsonSturgis\_Crash%20recovery\_later.pdf](https://web.cs.wpi.edu/~cs502/cisco11/Papers/LampsonSturgis_Crash%20recovery_later.pdf)
* Bernstein et al. (1987): Concurrency Control and Recovery
  [https://www.sigmod.org/publications/dblp/db/books/dbtext/bernstein87.html](https://www.sigmod.org/publications/dblp/db/books/dbtext/bernstein87.html)

