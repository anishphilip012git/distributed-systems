# 💾 Recovery in Distributed Databases

Mohammad Javad Amiri  
CSE 590: Distributed and Decentralized Data Management

---

## 🎯 Goals of Recovery

- Ensure **Atomicity** and **Durability** even in the presence of failures.
- Use logging and recovery protocols to:
  - Undo partial transactions (atomicity)
  - Redo committed but unflushed changes (durability)

---

## 🧰 Execution Model

- Disk blocks must be **loaded into memory** to be accessed.
- Changes are made in memory.
- Memory pages must be **flushed** to disk eventually.

---

## ⚠️ Failure Scenarios

1. **Crash after commit** – not all changes flushed ⇒ needs **redo**.
2. **Crash mid-transaction** – some changes flushed ⇒ needs **undo**.

---

## 🚫 Naïve Recovery Approaches

### Force
- Write all changes to disk at commit.
- Ensures durability.
- 🔻 Poor performance (random writes).

### No-Steal
- Don’t write dirty pages before commit.
- Ensures atomicity.
- 🔻 Requires lots of memory.

---

## ✅ Write-Ahead Logging (WAL)

- Log every change **before** writing to DB.
- Stored in **stable storage** (disk).
- WAL enables:
  - **Undo** (revert uncommitted changes)
  - **Redo** (replay committed changes)

---

## 📝 Log Format

- `⟨Ti, start⟩`
- `⟨Ti, X, old_val, new_val⟩`
- `⟨Ti, commit⟩` or `⟨Ti, abort⟩`

---

## 🔁 Undo/Redo Process

### Redo Phase
- Scan log **forward**.
- Reapply all operations from **committed transactions**.

### Undo Phase
- Scan log **backward**.
- Revert operations from **active (uncommitted) transactions**.

---

## 🔍 Example

Log:
```

\<T1,start>
\<T1,A,800,700>
\<T1,B,400,500>
\<T1,commit>

```

If crash happens:
- Before commit → **undo** A, B
- After commit → **redo** A, B

---

## 🧠 Active Transaction Tracking

- Maintain a set `U` of active txns at crash time.
- Redo all, then undo all in `U`.

---

## 🏁 Checkpointing

- Periodically write checkpoint info to log:
  - `⟨begin_checkpoint⟩`
  - Save active transaction list.
  - `⟨end_checkpoint⟩`

✅ Reduces recovery time.

---

## 🧍 Shadow Paging

- Maintain two page versions:
  - **Master**: committed state
  - **Shadow**: uncommitted changes

### Commit:
- Atomically switch root pointer to shadow pages.

### Pros:
- No redo needed.
- Easy rollback (discard shadow).

### Cons:
- High commit cost
- Fragmentation, GC needed
- Only one writer at a time

---

## 🔁 Summary Table

| Technique        | Ensures     | Tradeoffs                     |
|------------------|-------------|-------------------------------|
| Force            | Durability  | Low performance               |
| No-Steal         | Atomicity   | High memory use               |
| WAL              | A + D       | Requires stable storage logs  |
| Shadow Paging    | A + D       | Slow commits, limited concurrency |

---

## 📚 Resources

- Michael J. Franklin, "Concurrency Control and Recovery" (1997)  
  [Link](https://sands.kaust.edu.sa/classes/CS240/F24/papers/franklin97.pdf)

