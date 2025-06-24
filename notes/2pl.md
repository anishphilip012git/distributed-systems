# 🔐 Two-Phase Locking (2PL) – Concurrency Control

Two-Phase Locking (2PL) is a **pessimistic concurrency control protocol** that ensures **conflict serializability** in database transactions.

---

## 🎯 Goal

Guarantee **serializable schedules** during concurrent transaction execution by controlling how and when locks are acquired and released.

---

## 🧱 Core Idea

The execution of a transaction is split into two distinct phases:

1. **Growing Phase**:
   - The transaction **acquires** all required locks (shared or exclusive).
   - It can **upgrade** locks.
   - **No lock release** is allowed.

2. **Shrinking Phase**:
   - The transaction **releases** locks.
   - **No new locks** may be acquired.

> 📌 Once a lock is released, no more locks can be acquired.

---

## 🔐 Lock Types

- **Shared Lock (S-lock)** – for reading.
- **Exclusive Lock (X-lock)** – for writing.
- Lock Manager maintains:
  - Which transactions hold which locks.
  - Which transactions are waiting.

---

## ✅ Serializability Guarantee

- 2PL ensures that the resulting schedule is **conflict serializable**.
- It guarantees that the **precedence (dependency) graph** is **acyclic**.

---

## ⚠️ Limitations

- May allow:
  - **Cascading aborts** (if a transaction reads data from an uncommitted transaction that aborts).
  - **Deadlocks** (cycle of transactions waiting for each other's locks).

---

## 💪 Variants

### 1. **Strict 2PL (Rigorous 2PL)**
- Locks (both S and X) are **held until commit/abort**.
- Prevents **cascading aborts**.
- Stronger isolation.

### 2. **Conservative 2PL**
- All locks are acquired **at the beginning** of the transaction.
- Deadlock-free but less concurrent.

---

## 🔄 Deadlock Handling

- **Detection**: Build a **wait-for graph**, detect cycles, abort a transaction.
- **Prevention**: Use timestamp-based rules:
  - **Wait-die** or **wound-wait** strategies.
  - Abort younger transactions on conflict.

---

## 📘 Example

```text
T1: lock(A); lock(B); read(A); write(B); unlock(A); unlock(B)
T2: lock(B); lock(A); read(B); write(A); unlock(B); unlock(A)
````

Without 2PL: could interleave inconsistently.
With 2PL: locks ensure serializable schedule.

---

## 📌 Summary

| Feature          | 2PL                       |
| ---------------- | ------------------------- |
| Guarantee        | Conflict serializability  |
| Deadlock risk    | Yes                       |
| Cascading aborts | Yes                       |
| Strict 2PL       | Prevents cascading aborts |
| Use case         | Relational DBs, OLTP      |

---

## 🧠 Interview Tip

**"How do you guarantee serializability?"**
→ “By using Two-Phase Locking — acquire locks in the growing phase, release in shrinking. If strict 2PL is used, it also prevents cascading aborts.”


