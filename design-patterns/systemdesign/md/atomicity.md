# Database Atomicity

## What is Atomicity?

Atomicity means a transaction is "all or nothing." If a transaction has 5 operations, either all 5 succeed and are committed, or none of them take effect. There's no in-between state.

```
Transaction: Transfer $500 from Account A → Account B
┌─────────────────────────────────────────────────┐
│  Op 1: READ balance of A          → $1000       │
│  Op 2: WRITE A.balance = $500                   │
│  Op 3: READ balance of B          → $200        │
│  Op 4: WRITE B.balance = $700                   │
│                                                 │
│  ✅ All succeed → COMMIT (changes are permanent)│
│  ❌ Any fails   → ABORT  (undo everything)      │
└─────────────────────────────────────────────────┘
```

If the system crashes after Op 2 but before Op 4, without atomicity you'd lose $500 into thin air. Atomicity prevents that.

## Real-Life Analogy: The Wedding Vows

Think of a wedding ceremony:

```
┌──────────────────────────────────────────────────────┐
│                    WEDDING CEREMONY                  │
│                                                      │
│   Priest: "Do you, A, take B?"                       │
│   A: "I do"  ✅                                      │
│                                                      │
│   Priest: "Do you, B, take A?"                       │
│   B: "I do"  ✅                                      │
│                                                      │
│   Priest: "Sign the register"  ✅                    │
│                                                      │
│   ALL steps done → 💍 MARRIED (committed)            │
│                                                      │
│   ─── BUT if B says "No" ───                        │
│                                                      │
│   B: "No"  ❌                                        │
│   → Wedding CANCELLED. A's "I do" doesn't count.    │
│   → Everyone goes home. No marriage happened.        │
│   → State reverts to: both are single (rollback)     │
└──────────────────────────────────────────────────────┘
```

The marriage only "commits" when every step completes. If any step fails, the whole thing is off — you don't end up half-married.

## How Does the Database Actually Ensure This?

The core mechanism is a **Write-Ahead Log (WAL)**, also called a redo/undo log or transaction log.

### The Write-Ahead Log (WAL)

The key rule: **before any change is written to the actual database pages on disk, a log record describing that change is written to the WAL first.**

```
                    THE WAL PROTOCOL
                    
  ┌──────────┐         ┌──────────────┐        ┌──────────────┐
  │  Client   │         │     WAL      │        │  Data Pages  │
  │ (your app)│         │  (on disk)   │        │  (on disk)   │
  └─────┬─────┘         └──────┬───────┘        └──────┬───────┘
        │                      │                       │
        │  BEGIN TX             │                       │
        │─────────────────────>│ Log: "TX-101 BEGIN"   │
        │                      │───────────────>       │
        │                      │                       │
        │  UPDATE A = 500      │                       │
        │─────────────────────>│ Log: "TX-101:         │
        │                      │  A: old=1000,new=500" │
        │                      │───────────────>       │
        │                      │                       │ (change A
        │                      │                       │  in memory)
        │  UPDATE B = 700      │                       │
        │─────────────────────>│ Log: "TX-101:         │
        │                      │  B: old=200,new=700"  │
        │                      │───────────────>       │
        │                      │                       │ (change B
        │                      │                       │  in memory)
        │  COMMIT              │                       │
        │─────────────────────>│ Log: "TX-101 COMMIT"  │
        │                      │───────────────>       │
        │                      │                       │ (flush to
        │  ✅ OK               │                       │  disk)
        │<─────────────────────│                       │
```

The WAL stores both the **old value** and the **new value** for every operation. This is critical.

### What Happens on Failure?

```
SCENARIO: Crash after Op 2, before Op 4
═══════════════════════════════════════════

WAL on disk contains:
┌─────────────────────────────────────────┐
│  Record 1: TX-101 BEGIN                 │
│  Record 2: TX-101 → A: 1000 → 500      │
│  Record 3: TX-101 → B: 200  → 700      │
│  ❌ NO "TX-101 COMMIT" record           │
└─────────────────────────────────────────┘

Database restarts → Recovery Manager kicks in:

  ┌─────────────────────────────────────────────────┐
  │           RECOVERY PROCESS                      │
  │                                                 │
  │  1. Scan the WAL from the end                   │
  │                                                 │
  │  2. Find TX-101: has BEGIN but NO COMMIT        │
  │     → This is an incomplete transaction         │
  │                                                 │
  │  3. UNDO phase: walk backwards through TX-101   │
  │     → Restore A from 500 back to 1000 (old val) │
  │     → Restore B from 700 back to 200  (old val) │
  │                                                 │
  │  4. Log: "TX-101 ABORT" → done                  │
  │                                                 │
  │  Result: Database is back to pre-transaction    │
  │          state. No money lost.                  │
  └─────────────────────────────────────────────────┘
```

And the happy path:

```
SCENARIO: Crash AFTER commit record is written
═══════════════════════════════════════════

WAL on disk contains:
┌─────────────────────────────────────────┐
│  Record 1: TX-101 BEGIN                 │
│  Record 2: TX-101 → A: 1000 → 500      │
│  Record 3: TX-101 → B: 200  → 700      │
│  Record 4: TX-101 COMMIT  ✅            │
└─────────────────────────────────────────┘

But maybe the actual data pages weren't fully flushed...

  ┌─────────────────────────────────────────────────┐
  │           RECOVERY PROCESS                      │
  │                                                 │
  │  1. Scan the WAL                                │
  │                                                 │
  │  2. Find TX-101: has BEGIN AND COMMIT           │
  │     → This is a committed transaction           │
  │                                                 │
  │  3. REDO phase: replay TX-101 forward           │
  │     → Ensure A = 500 on data pages              │
  │     → Ensure B = 700 on data pages              │
  │                                                 │
  │  Result: Committed data is guaranteed durable.  │
  └─────────────────────────────────────────────────┘
```

### The Full Recovery Model (ARIES-style)

Most production databases (PostgreSQL, MySQL/InnoDB, SQL Server) use a recovery algorithm inspired by ARIES. Here's the big picture:

```
┌─────────────────────────────────────────────────────────────┐
│                    CRASH RECOVERY                           │
│                                                             │
│  Phase 1: ANALYSIS                                         │
│  ┌───────────────────────────────────────────────────┐      │
│  │ Scan WAL to figure out:                           │      │
│  │  - Which transactions were active at crash time   │      │
│  │  - Which pages might be dirty (not flushed)       │      │
│  └───────────────────────────────────────────────────┘      │
│                          │                                  │
│                          ▼                                  │
│  Phase 2: REDO (replay history)                             │
│  ┌───────────────────────────────────────────────────┐      │
│  │ Replay ALL logged operations (committed or not)   │      │
│  │ to bring data pages to the exact state at crash   │      │
│  └───────────────────────────────────────────────────┘      │
│                          │                                  │
│                          ▼                                  │
│  Phase 3: UNDO (rollback losers)                            │
│  ┌───────────────────────────────────────────────────┐      │
│  │ For every transaction that did NOT commit:        │      │
│  │  → Walk backwards, restore old values             │      │
│  │  → Write "compensation log records" (CLRs)        │      │
│  │  → Mark transaction as ABORTED                    │      │
│  └───────────────────────────────────────────────────┘      │
│                                                             │
│  Result: Committed TXs are preserved, uncommitted are gone  │
└─────────────────────────────────────────────────────────────┘
```

### Why "Write-Ahead"?

```
  THE GOLDEN RULE:
  ════════════════

  Log record hits disk BEFORE the actual data change hits disk.
  
  ┌──────────┐    ①    ┌──────────┐    ②    ┌──────────┐
  │ Operation │──────>│   WAL    │──────>│ Data Page │
  │ (in mem)  │ write │ (on disk)│ then  │ (on disk) │
  └──────────┘  log   └──────────┘ write └──────────┘
                first              data

  If crash happens between ① and ②:
    → WAL has the record, recovery can redo it ✅
  
  If crash happens before ①:
    → Nothing was logged, nothing to redo, 
      old data is intact ✅
  
  If data hits disk BEFORE log (violating WAL):
    → Crash leaves changed data with no log to undo it 💀
       THIS IS WHY THE ORDER MATTERS
```

## Summary

The whole mechanism boils down to:

1. **Log every intended change** (with old + new values) to the WAL before touching actual data
2. **On commit**, write a COMMIT marker to the WAL and flush it to disk — this is the "point of no return"
3. **On crash recovery**, scan the WAL: redo committed transactions, undo uncommitted ones
4. The **COMMIT record** in the WAL is the single source of truth for whether a transaction happened or not

Your intuition was spot on — the database does keep a log of all operations in a transaction, and uses that log to decide whether to replay (redo) or revert (undo) after a failure. The WAL is the backbone of atomicity (and durability too).

---

## How Does WAL Know the Old Value?

The WAL doesn't "know" the old value on its own. The database engine reads the old value from the data page first, then constructs the log record with both old and new values before writing it to the WAL.

The actual sequence when you run `UPDATE accounts SET balance = 500 WHERE id = 'A'`:

```
Step 1: Engine reads the data page containing row A into memory (buffer pool)
        → It now sees: A.balance = 1000 (this is the old value)

Step 2: Engine constructs the WAL record:
        { tx: 101, table: accounts, row: A, col: balance, old: 1000, new: 500 }
        → old value came from Step 1
        → new value came from the UPDATE statement

Step 3: WAL record is written/flushed to disk

Step 4: The in-memory copy of the data page is modified (A.balance = 500)
        (the actual on-disk data page is updated lazily later)
```

So the flow is:

```
  ┌────────────────┐     read row     ┌────────────────────┐
  │  UPDATE A=500  │ ───────────────> │  Buffer Pool (RAM) │
  │  (your query)  │                  │  A.balance = 1000  │
  └────────┬───────┘                  └─────────┬──────────┘
           │                                    │
           │  engine now has:                   │
           │  old = 1000 (from buffer pool)     │
           │  new = 500  (from your query)      │
           │                                    │
           ▼                                    │
  ┌────────────────────────────────┐            │
  │  WAL Record:                   │            │
  │  TX-101, A, old=1000, new=500  │            │
  │  → written to disk FIRST       │            │
  └────────────────────────────────┘            │
                                                │
           then ──────────────────────────────> │
                                                │
                                    ┌───────────▼──────────┐
                                    │  Buffer Pool (RAM)   │
                                    │  A.balance = 500     │
                                    │  (page marked dirty) │
                                    └──────────────────────┘
```

The key insight: the database always reads before it writes. It has to — it needs to find the row, check constraints, evaluate WHERE clauses, etc. So by the time it's ready to modify anything, it already has the old value sitting right there in memory. Constructing the WAL record with both values is essentially free at that point.

The buffer pool (in-memory cache of data pages) is the central piece here. Data pages get loaded into the buffer pool on reads, and the engine works against that in-memory copy. The old value is never "looked up separately for the WAL" — it's just already there as part of normal query execution.

---

## Recovery: Crash vs Normal Runtime

### Two Different Situations

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  Situation 1: NORMAL OPERATION (no crash)                       │
│  → Transaction fails or app calls ROLLBACK                      │
│  → Recovery manager is NOT involved                             │
│  → The engine just reads the WAL for that TX and undoes it      │
│  → This happens in real-time, inline, while DB is running       │
│                                                                 │
│  Situation 2: CRASH RECOVERY (database restarts)                │
│  → Recovery manager kicks in ONCE at startup                    │
│  → Scans the entire WAL from last checkpoint                    │
│  → Fixes everything, then hands control to normal operations    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Situation 1: Normal Runtime (No Crash)

During normal operation, if a transaction needs to abort (constraint violation, deadlock, app calls ROLLBACK), the engine handles it immediately — no restart needed:

```
  Timeline (DB is running normally, multiple TXs active)
  ═══════════════════════════════════════════════════════

  TX-101: BEGIN → UPDATE A → UPDATE B → COMMIT ✅  (done, permanent)
  TX-102: BEGIN → UPDATE C → UPDATE D → ❌ constraint violation!
  TX-103: BEGIN → UPDATE E → ...still running...

  What happens to TX-102:
  ┌──────────────────────────────────────────────────┐
  │  Engine sees the error on TX-102                 │
  │                                                  │
  │  1. Reads WAL records for TX-102 backwards:      │
  │     → Found: D old=50, new=90                    │
  │     → Found: C old=100, new=200                  │
  │                                                  │
  │  2. Undoes them in reverse order:                 │
  │     → Restore D = 50                             │
  │     → Restore C = 100                            │
  │                                                  │
  │  3. Writes "TX-102 ABORT" to WAL                 │
  │                                                  │
  │  Done. No restart. Other TXs keep running.       │
  └──────────────────────────────────────────────────┘
```

This is NOT a continuous background process reading the WAL. It's triggered on-demand when a specific transaction needs to roll back.

### Situation 2: Crash Recovery (The Big One)

This only happens once — when the database starts up after an unclean shutdown. The recovery manager runs at startup, fixes everything, then gets out of the way.

At crash time, there could be dozens of transactions in various states:

```
  STATE AT CRASH TIME
  ════════════════════

  WAL on disk:
  ┌──────────────────────────────────────────────────────┐
  │  ...                                                 │
  │  CHECKPOINT (at time T0)                             │
  │  ...                                                 │
  │  TX-101 BEGIN                                        │
  │  TX-102 BEGIN                                        │
  │  TX-101 → A: 1000 → 500                             │
  │  TX-103 BEGIN                                        │
  │  TX-102 → B: 200 → 700                              │
  │  TX-101 → C: 300 → 100                              │
  │  TX-101 COMMIT  ✅                                   │
  │  TX-103 → D: 400 → 800                              │
  │  TX-102 → E: 50 → 90                                │
  │  TX-104 BEGIN                                        │
  │  TX-104 → F: 60 → 120                               │
  │  ════════════ 💥 CRASH HERE ════════════             │
  │                                                      │
  │  TX-101: has COMMIT    → winner ✅                   │
  │  TX-102: no COMMIT     → loser  ❌                   │
  │  TX-103: no COMMIT     → loser  ❌                   │
  │  TX-104: no COMMIT     → loser  ❌                   │
  └──────────────────────────────────────────────────────┘
```

Now the database restarts:

```
  RECOVERY AT STARTUP (runs ONCE)
  ════════════════════════════════

  Phase 1: ANALYSIS (scan forward from last checkpoint)
  ┌──────────────────────────────────────────────────┐
  │  Build two lists:                                │
  │                                                  │
  │  Winners (have COMMIT): [TX-101]                 │
  │  Losers  (no COMMIT):   [TX-102, TX-103, TX-104]│
  │                                                  │
  │  Also track which data pages might be dirty      │
  └──────────────────────────────────────────────────┘
                        │
                        ▼
  Phase 2: REDO (scan forward — replay EVERYTHING)
  ┌──────────────────────────────────────────────────┐
  │  Replay ALL operations, winners AND losers:      │
  │                                                  │
  │  → A: 1000 → 500  (TX-101)                      │
  │  → B: 200 → 700   (TX-102)                      │
  │  → C: 300 → 100   (TX-101)                      │
  │  → D: 400 → 800   (TX-103)                      │
  │  → E: 50 → 90     (TX-102)                      │
  │  → F: 60 → 120    (TX-104)                      │
  │                                                  │
  │  Why replay losers too? To bring data pages to   │
  │  the exact state they were in at crash time.     │
  │  We need that state to correctly undo them.      │
  └──────────────────────────────────────────────────┘
                        │
                        ▼
  Phase 3: UNDO (scan backward — rollback losers)
  ┌──────────────────────────────────────────────────┐
  │  Walk WAL backwards, undo only loser TXs:        │
  │                                                  │
  │  → F: 120 → 60   (undo TX-104)                  │
  │  → E: 90 → 50    (undo TX-102)                  │
  │  → D: 800 → 400  (undo TX-103)                  │
  │  → B: 700 → 200  (undo TX-102)                  │
  │                                                  │
  │  Write ABORT records for TX-102, TX-103, TX-104  │
  └──────────────────────────────────────────────────┘
                        │
                        ▼
  ┌──────────────────────────────────────────────────┐
  │  RECOVERY COMPLETE                               │
  │                                                  │
  │  Final state:                                    │
  │  A = 500, C = 100  (TX-101 committed, kept)      │
  │  B = 200           (TX-102 rolled back)          │
  │  D = 400           (TX-103 rolled back)          │
  │  E = 50            (TX-102 rolled back)          │
  │  F = 60            (TX-104 rolled back)          │
  │                                                  │
  │  DB now accepts new connections normally.         │
  └──────────────────────────────────────────────────┘
```

### What About Checkpoints?

Checkpoints prevent the WAL from growing forever and keep recovery fast:

```
  CHECKPOINTS (happen periodically during normal operation)
  ═════════════════════════════════════════════════════════

  ┌─────────────────────────────────────────────────────┐
  │  Checkpoint = "snapshot" of current state            │
  │                                                     │
  │  What it does:                                      │
  │  1. Flush all dirty pages from buffer pool to disk  │
  │  2. Write a CHECKPOINT record to WAL noting:        │
  │     → Which TXs are currently active                │
  │     → All dirty pages are now clean                 │
  │                                                     │
  │  Why it matters:                                    │
  │  → Recovery only needs to scan from the LAST        │
  │    checkpoint, not from the beginning of time       │
  │  → Old WAL segments before the checkpoint can       │
  │    be safely deleted                                │
  └─────────────────────────────────────────────────────┘

  WAL timeline:
  ──[old stuff]──[CHECKPOINT]──[new operations]──💥 CRASH
                      ▲
                      │
              Recovery starts HERE
              (ignores everything before)
```

### TL;DR

- During normal operation: rollback is handled inline, on-demand, per-transaction. No background scanner.
- After a crash: recovery runs once at startup, processes ALL in-flight transactions from the last checkpoint, then finishes.
- Checkpoints keep the WAL bounded so recovery doesn't take forever.
- The recovery manager is not a continuously running process — it's a startup procedure.

---

## Applying WAL to a Real Problem: Dual Write (Cassandra + Elasticsearch)

### The Problem

When doing dual writes — saving data to Cassandra first, then to Elasticsearch — if the Elasticsearch write fails, you need to revert Cassandra. Normally this is handled with application-level if/else checks. But we can apply the WAL pattern to make this more robust.

### The Approach

```
┌─────────────────────────────────────────────────────────────┐
│              APPLICATION-LEVEL WAL FOR DUAL WRITE           │
│                                                             │
│  1. Read old value from Cassandra                           │
│  2. Write WAL record: { old_value, new_value, status: PENDING }│
│  3. Write to Cassandra                                      │
│  4. Write to Elasticsearch                                  │
│  5a. Both succeed → WAL status = COMMITTED, delete WAL      │
│  5b. ES fails     → Read WAL, revert Cassandra using old_value│
│                     WAL status = ROLLED_BACK                │
│                                                             │
│  On app restart:                                            │
│    Scan WAL for any PENDING records → revert them           │
└─────────────────────────────────────────────────────────────┘
```

### Flow Diagrams

```
  Happy Path:
  ───────────

  App                    WAL File              Cassandra         Elasticsearch
   │                        │                      │                  │
   │  1. Read old value     │                      │                  │
   │───────────────────────────────────────────>   │                  │
   │  old = {name: "Alice"} │                      │                  │
   │<──────────────────────────────────────────    │                  │
   │                        │                      │                  │
   │  2. Write WAL record   │                      │                  │
   │──────────────────────> │                      │                  │
   │   {old: "Alice",       │                      │                  │
   │    new: "Bob",         │                      │                  │
   │    status: PENDING}    │                      │                  │
   │                        │                      │                  │
   │  3. Write Cassandra    │                      │                  │
   │───────────────────────────────────────────>   │                  │
   │                        │                      │                  │
   │  4. Write ES           │                      │                  │
   │──────────────────────────────────────────────────────────────>   │
   │                        │                      │                  │
   │  5. Mark COMMITTED     │                      │                  │
   │──────────────────────> │                      │                  │
   │  6. Delete WAL record  │                      │                  │
   │──────────────────────> │                      │                  │
   ✅ Done                  │                      │                  │


  Failure Path (ES fails):
  ────────────────────────

   │  4. Write ES           │                      │                  │
   │──────────────────────────────────────────────────────────────>   │
   │                        │                      │              ❌ FAIL
   │                        │                      │                  │
   │  5. Read WAL (old val) │                      │                  │
   │──────────────────────> │                      │                  │
   │  old = "Alice"         │                      │                  │
   │<────────────────────── │                      │                  │
   │                        │                      │                  │
   │  6. Revert Cassandra   │                      │                  │
   │───────────────────────────────────────────>   │                  │
   │   name = "Alice"       │                      │                  │
   │                        │                      │                  │
   │  7. Mark ROLLED_BACK   │                      │                  │
   │──────────────────────> │                      │                  │
   ✅ Reverted              │                      │                  │
```

### Go Implementation

#### WAL (wal.go)

```go
package dualwrite

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// WAL Status
type Status string

const (
	StatusPending    Status = "PENDING"
	StatusCommitted  Status = "COMMITTED"
	StatusRolledBack Status = "ROLLED_BACK"
)

// WALRecord holds the old and new state for a single operation.
type WALRecord struct {
	TxID      string          `json:"tx_id"`
	Table     string          `json:"table"`      // Cassandra table
	Key       string          `json:"key"`        // primary key of the row
	OldValue  json.RawMessage `json:"old_value"`  // state before the write
	NewValue  json.RawMessage `json:"new_value"`  // state after the write
	Status    Status          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

// WAL is a file-backed write-ahead log for dual-write atomicity.
type WAL struct {
	mu       sync.Mutex
	filePath string
	records  map[string]*WALRecord // txID -> record
}

// NewWAL creates or opens a WAL file. On open, it loads any existing
// records so you can recover PENDING transactions on restart.
func NewWAL(filePath string) (*WAL, error) {
	w := &WAL{
		filePath: filePath,
		records:  make(map[string]*WALRecord),
	}
	if err := w.load(); err != nil {
		return nil, fmt.Errorf("wal: load failed: %w", err)
	}
	return w, nil
}

// Begin writes a PENDING record to the WAL before any data store is touched.
func (w *WAL) Begin(txID, table, key string, oldValue, newValue any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	oldBytes, err := json.Marshal(oldValue)
	if err != nil {
		return fmt.Errorf("wal: marshal old value: %w", err)
	}
	newBytes, err := json.Marshal(newValue)
	if err != nil {
		return fmt.Errorf("wal: marshal new value: %w", err)
	}

	rec := &WALRecord{
		TxID:      txID,
		Table:     table,
		Key:       key,
		OldValue:  oldBytes,
		NewValue:  newBytes,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	w.records[txID] = rec
	return w.flush()
}

// Commit marks a transaction as committed and removes it from the WAL.
func (w *WAL) Commit(txID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	rec, ok := w.records[txID]
	if !ok {
		return fmt.Errorf("wal: tx %s not found", txID)
	}
	rec.Status = StatusCommitted
	delete(w.records, txID) // committed = safe to forget
	return w.flush()
}

// MarkRolledBack marks a transaction as rolled back.
func (w *WAL) MarkRolledBack(txID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	rec, ok := w.records[txID]
	if !ok {
		return fmt.Errorf("wal: tx %s not found", txID)
	}
	rec.Status = StatusRolledBack
	delete(w.records, txID)
	return w.flush()
}

// GetPendingRecords returns all PENDING records (used during recovery on restart).
func (w *WAL) GetPendingRecords() []*WALRecord {
	w.mu.Lock()
	defer w.mu.Unlock()

	var pending []*WALRecord
	for _, rec := range w.records {
		if rec.Status == StatusPending {
			pending = append(pending, rec)
		}
	}
	return pending
}

// GetRecord returns a specific WAL record by transaction ID.
func (w *WAL) GetRecord(txID string) (*WALRecord, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	rec, ok := w.records[txID]
	return rec, ok
}

// flush writes the current in-memory state to disk (the WAL file).
func (w *WAL) flush() error {
	data, err := json.MarshalIndent(w.records, "", "  ")
	if err != nil {
		return fmt.Errorf("wal: marshal: %w", err)
	}
	return os.WriteFile(w.filePath, data, 0644)
}

// load reads existing WAL records from disk into memory.
func (w *WAL) load() error {
	data, err := os.ReadFile(w.filePath)
	if os.IsNotExist(err) {
		return nil // no WAL file yet, fresh start
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &w.records)
}
```

#### Coordinator (coordinator.go)

```go
package dualwrite

// CassandraClient is the interface your Cassandra layer must implement.
type CassandraClient interface {
	Read(table, key string) (any, error)
	Write(table, key string, value any) error
}

// ElasticClient is the interface your Elasticsearch layer must implement.
type ElasticClient interface {
	Write(index, id string, value any) error
}

// Coordinator orchestrates atomic dual writes using the WAL.
type Coordinator struct {
	wal       *WAL
	cassandra CassandraClient
	elastic   ElasticClient
}

func NewCoordinator(wal *WAL, cass CassandraClient, es ElasticClient) *Coordinator {
	return &Coordinator{wal: wal, cassandra: cass, elastic: es}
}
```
