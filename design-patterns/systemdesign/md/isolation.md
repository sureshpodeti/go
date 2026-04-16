# Isolation in ACID

Isolation means: concurrent transactions behave as if they were running one after another, even though they're actually running in parallel. One transaction's uncommitted changes should not leak into another transaction's view of the data.

Real-life analogy: Think of isolation like exam halls. Every student (transaction) is solving the same paper (database), but they can't see each other's answer sheets. Even though 200 students are writing simultaneously, each one's work is isolated. The final answer sheets (committed state) are collected independently.

---

## Is Isolation at Row Level or Table Level?

Short answer: **Row level.** Modern databases isolate at the row (and sometimes even sub-row) level, not at the table level.

When two concurrent transactions touch the same row, the database ensures they don't step on each other. When they touch different rows in the same table, they can proceed in parallel without blocking each other.

```
Transaction A: UPDATE accounts SET balance = balance - 100 WHERE id = 1;
Transaction B: UPDATE accounts SET balance = balance + 200 WHERE id = 2;

These two run concurrently with ZERO conflict — different rows.

Transaction A: UPDATE accounts SET balance = balance - 100 WHERE id = 1;
Transaction C: UPDATE accounts SET balance = balance + 50  WHERE id = 1;

These two conflict — same row. The DB must isolate them.
```

Real-life analogy: Think of a library. Two people can check out different books at the same time (different rows — no conflict). But if two people want the same book, the librarian makes one wait until the other is done (same row — isolation kicks in).

---

## How Does the DB Achieve Isolation?

Databases use two primary mechanisms, often in combination:

### 1. Locking (Pessimistic Concurrency Control)

The database acquires locks on rows before reading or writing them.

```
Lock Types:
┌──────────────┬─────────────────────────────────────────────┐
│ Lock Type    │ Behavior                                    │
├──────────────┼─────────────────────────────────────────────┤
│ Shared (S)   │ Multiple transactions can hold it (reads)   │
│ Exclusive (X) │ Only one transaction can hold it (writes)  │
└──────────────┴─────────────────────────────────────────────┘

Compatibility Matrix:
              Requesting
              S       X
Held   S      ✅      ❌
       X      ❌      ❌

Example:
  Txn A: SELECT * FROM accounts WHERE id = 1;  → acquires S lock on row 1
  Txn B: SELECT * FROM accounts WHERE id = 1;  → acquires S lock on row 1 (compatible ✅)
  Txn C: UPDATE accounts SET balance = 0 WHERE id = 1; → wants X lock → BLOCKED until A and B release
```

The granularity of locks:

```
Granularity Spectrum:

  Table Lock ──────── Page Lock ──────── Row Lock ──────── Column Lock
  (coarse)                                (fine)            (very fine, rare)

  Coarse = less overhead, more contention (blocks more transactions)
  Fine   = more overhead, less contention (better concurrency)

Modern databases default to ROW-LEVEL locking.
PostgreSQL, MySQL (InnoDB), Oracle — all row-level by default.
```

### 2. MVCC (Multi-Version Concurrency Control) — The Modern Approach

Instead of blocking readers, the database keeps multiple versions of each row. Each transaction sees a snapshot of the data as it was when the transaction started.

```
Timeline:
  t=0: Row(id=1, balance=1000, version=1)

  t=1: Txn A starts (sees snapshot at t=1)
  t=2: Txn B starts (sees snapshot at t=2)

  t=3: Txn A: UPDATE balance = 900 WHERE id = 1
       → Creates new version: Row(id=1, balance=900, version=2) [uncommitted]
       → Old version (balance=1000, version=1) still exists

  t=4: Txn B: SELECT balance WHERE id = 1
       → Txn A hasn't committed yet
       → Txn B sees version=1 → balance=1000 ✅ (no dirty read)

  t=5: Txn A: COMMIT
       → version=2 (balance=900) is now the committed version

  t=6: Txn B: SELECT balance WHERE id = 1
       → Depending on isolation level, Txn B might see 900 or still 1000
```

Real-life analogy: MVCC is like Google Docs version history. When you open a document, you see a snapshot. Someone else can be editing it simultaneously, but you don't see their half-typed sentence. You see a consistent version. When they save (commit), the next time you refresh (new query), you see the updated version.

```
How MVCC stores versions (PostgreSQL style):

  Heap (main table storage):
  ┌─────┬─────────┬────────┬────────┬──────────┐
  │ id  │ balance │ xmin   │ xmax   │ visible? │
  ├─────┼─────────┼────────┼────────┼──────────┤
  │ 1   │ 1000    │ 100    │ 105    │ to txns < 105 │
  │ 1   │ 900     │ 105    │ ∞      │ to txns ≥ 105 │
  └─────┴─────────┴────────┴────────┴──────────┘

  xmin = transaction ID that created this version
  xmax = transaction ID that deleted/replaced this version
  
  Each transaction checks: "Is this version visible to me based on my snapshot?"
```


---

## Snapshot Isolation vs MVCC — They're Not the Same Thing

A common confusion: people use "Snapshot Isolation" and "MVCC" interchangeably. They're related but fundamentally different.

```
┌──────────────────────┬──────────────────────────────────────────────────┐
│                      │ What is it?                                      │
├──────────────────────┼──────────────────────────────────────────────────┤
│ Snapshot Isolation    │ An isolation LEVEL — a behavioral guarantee.    │
│                      │ "Your transaction sees a consistent snapshot     │
│                      │  of the DB as of when it started."              │
├──────────────────────┼──────────────────────────────────────────────────┤
│ MVCC                 │ A concurrency control MECHANISM — an            │
│                      │ implementation technique. "We keep multiple      │
│                      │ versions of rows so readers and writers don't   │
│                      │ block each other."                              │
└──────────────────────┴──────────────────────────────────────────────────┘
```

Snapshot Isolation is the promise. MVCC is the most common way to deliver that promise.

But MVCC is more general — it's used to implement multiple isolation levels, not just Snapshot Isolation:

```
MVCC can implement:
  ├── Read Committed      → each STATEMENT gets a fresh snapshot
  ├── Repeatable Read     → entire TRANSACTION gets one snapshot
  ├── Snapshot Isolation   → transaction-level snapshot + first-committer-wins
  └── Serializable (SSI)  → snapshot + conflict detection at commit time

PostgreSQL uses MVCC for ALL its isolation levels.
The difference is how/when the snapshot is taken and what conflicts are checked.
```

And in theory, Snapshot Isolation could be implemented without MVCC (e.g., copy-on-write at the page level), though that's uncommon in practice.

```
The relationship:

  MVCC (mechanism) ──implements──▶ Snapshot Isolation (level)
                                   Read Committed
                                   Repeatable Read
                                   Serializable (SSI)

  Think of it like:
    MVCC = the engine
    Isolation level = the gear you're driving in
```

---

## Isolation Levels — From Weakest to Strongest

SQL standard defines four isolation levels. Think of them as a dial — you trade correctness for performance.

```
Weakest                                                    Strongest
   │                                                          │
   ▼                                                          ▼
Read Uncommitted → Read Committed → Repeatable Read → Serializable

More concurrency ◄──────────────────────────────► More correctness
Less overhead                                      More overhead
```

### Level 1: Read Uncommitted (Weakest)

A transaction can see another transaction's uncommitted (dirty) changes.

```
Txn A: BEGIN
Txn A: UPDATE accounts SET balance = 0 WHERE id = 1;  -- not committed yet

Txn B: BEGIN
Txn B: SELECT balance FROM accounts WHERE id = 1;
       → Returns 0 (dirty read! Txn A hasn't committed!)

Txn A: ROLLBACK  -- oops, Txn A aborts

Txn B now made a decision based on balance=0, which NEVER actually existed.
```

Real-life analogy: You're at an auction. Someone shouts "$500!" but then says "wait, never mind." But you already bid $501 based on their fake bid. You just got scammed by a dirty read.

### Level 2: Read Committed (Most Common Default)

A transaction only sees committed data. But if you read the same row twice, you might get different values (non-repeatable read).

```
Txn A: BEGIN
Txn A: SELECT balance FROM accounts WHERE id = 1;  → 1000

Txn B: BEGIN
Txn B: UPDATE accounts SET balance = 500 WHERE id = 1;
Txn B: COMMIT

Txn A: SELECT balance FROM accounts WHERE id = 1;  → 500 (changed!)
       Non-repeatable read: same query, different result within the same transaction.
```

Real-life analogy: You check the price of a flight — $200. You go grab your credit card, come back, and now it's $350. The price "committed" between your two reads. Frustrating, but at least you never saw a price that didn't actually exist.

### Level 3: Repeatable Read

Once you read a row, you'll always see the same value for that row within your transaction. But new rows inserted by other transactions might appear (phantom reads).

```
Txn A: BEGIN
Txn A: SELECT * FROM accounts WHERE balance > 500;
       → Returns: [id=1, balance=1000], [id=2, balance=700]

Txn B: BEGIN
Txn B: INSERT INTO accounts (id, balance) VALUES (3, 800);
Txn B: COMMIT

Txn A: SELECT * FROM accounts WHERE balance > 500;
       → Returns: [id=1, balance=1000], [id=2, balance=700], [id=3, balance=800]
       Phantom read: a NEW row appeared that matches the query!
       (Existing rows didn't change — that's guaranteed. But new ones snuck in.)
```

Real-life analogy: You count the people in a meeting room — 10 people. You look again — still the same 10 people (repeatable). But wait, someone new walked in through the back door. Now there are 11. The original 10 didn't change, but a phantom appeared.

Note: MySQL's InnoDB actually prevents phantom reads at Repeatable Read level using gap locks. PostgreSQL's Repeatable Read also prevents phantoms via its MVCC snapshot. So in practice, these two engines go beyond the SQL standard at this level.

### Level 4: Serializable (Strongest)

Transactions behave as if they executed one after another, in some serial order. No dirty reads, no non-repeatable reads, no phantoms. Full correctness.

```
Txn A: BEGIN
Txn A: SELECT SUM(balance) FROM accounts;  → 5000

Txn B: BEGIN
Txn B: INSERT INTO accounts (id, balance) VALUES (99, 200);
Txn B: COMMIT  → might be BLOCKED or Txn B might be aborted

Txn A: SELECT SUM(balance) FROM accounts;  → still 5000
       No phantoms. The world is frozen from Txn A's perspective.
Txn A: COMMIT
```

Real-life analogy: Serializable is like a single-lane bridge. Cars (transactions) cross one at a time. It's slow, but there are zero accidents (anomalies). In practice, databases are smarter — they let cars cross in parallel when they can prove there's no conflict, but force serialization when there is.

---

## Summary of Anomalies per Isolation Level

```
┌────────────────────┬──────────────┬────────────────────┬──────────────┐
│ Isolation Level    │ Dirty Read   │ Non-Repeatable Read│ Phantom Read │
├────────────────────┼──────────────┼────────────────────┼──────────────┤
│ Read Uncommitted   │ ✅ possible  │ ✅ possible        │ ✅ possible  │
│ Read Committed     │ ❌ prevented │ ✅ possible        │ ✅ possible  │
│ Repeatable Read    │ ❌ prevented │ ❌ prevented       │ ✅ possible* │
│ Serializable       │ ❌ prevented │ ❌ prevented       │ ❌ prevented │
└────────────────────┴──────────────┴────────────────────┴──────────────┘

* MySQL InnoDB and PostgreSQL prevent phantoms at Repeatable Read too (beyond SQL standard)
```

---

## What Default Isolation Level Do Modern Databases Use?

```
┌──────────────────────┬─────────────────────┬──────────────────────────────┐
│ Database             │ Default Level       │ Notes                        │
├──────────────────────┼─────────────────────┼──────────────────────────────┤
│ PostgreSQL           │ Read Committed      │ MVCC-based, very robust      │
│ MySQL (InnoDB)       │ Repeatable Read     │ Uses gap locks for phantoms  │
│ Oracle               │ Read Committed      │ MVCC (undo segments)         │
│ SQL Server           │ Read Committed      │ Lock-based by default        │
│ SQLite               │ Serializable        │ Single-writer model          │
│ CockroachDB          │ Serializable        │ Distributed, SSI-based       │
│ Google Spanner       │ Serializable +      │ External consistency via     │
│                      │ Linearizable        │ TrueTime (atomic clocks)     │
├──────────────────────┼─────────────────────┼──────────────────────────────┤
│ MongoDB              │ Read Uncommitted    │ Single-doc ops are atomic.   │
│                      │                     │ Multi-doc txns (4.0+) use    │
│                      │                     │ Snapshot Isolation.           │
├──────────────────────┼─────────────────────┼──────────────────────────────┤
│ Cassandra            │ No traditional      │ AP system (availability +    │
│                      │ isolation           │ partition tolerance). Eventual│
│                      │                     │ consistency by default. LWT  │
│                      │                     │ (Paxos) gives linearizable   │
│                      │                     │ consistency per-partition.    │
└──────────────────────┴─────────────────────┴──────────────────────────────┘

Most common default: Read Committed (PostgreSQL, Oracle, SQL Server)
MySQL is the outlier with Repeatable Read as default.
NoSQL databases like MongoDB and Cassandra don't map cleanly to SQL isolation levels — they have different tradeoff models entirely.
```


---

## Databases That Provide Serializable (Highest) Isolation

Yes, several databases offer serializable isolation, either as default or tunable:

### 1. PostgreSQL — Serializable Snapshot Isolation (SSI)

```sql
-- Per transaction:
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
SELECT * FROM accounts WHERE id = 1;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
COMMIT;  -- might fail with "could not serialize access" if conflict detected

-- Or globally:
ALTER SYSTEM SET default_transaction_isolation = 'serializable';
```

PostgreSQL uses SSI (Serializable Snapshot Isolation) — an optimistic approach. It lets transactions run concurrently on snapshots, then checks at commit time whether the execution was equivalent to some serial order. If not, it aborts one of the conflicting transactions.

```
How SSI works:

  Txn A reads X, writes Y (based on X)
  Txn B reads Y, writes X (based on Y)

  This is a "dangerous structure" (rw-antidependency cycle):
    A --reads--> X --written-by--> B
    B --reads--> Y --written-by--> A

  PostgreSQL detects this cycle and aborts one transaction.
  No locks held during execution — only validation at commit.
```

### 2. CockroachDB — Serializable by Default

CockroachDB is a distributed SQL database that defaults to serializable. It uses a combination of MVCC + timestamp ordering. You don't even have to ask for it — it's the only level available.

### 3. Google Spanner — Serializable + External Consistency

Spanner goes beyond serializable. It provides external consistency (also called strict serializability or linearizability + serializability). It uses TrueTime (GPS + atomic clocks) to assign globally meaningful timestamps.

```
Spanner's TrueTime:

  Normal DB:  "Transaction A committed before B" (local ordering)
  Spanner:    "Transaction A committed before B in REAL wall-clock time" (global ordering)

  This is possible because Spanner knows the actual time within a bounded uncertainty:
    TrueTime.now() → [earliest, latest]
    Spanner waits out the uncertainty window before committing.
```

### 4. SQLite — Serializable by Default

SQLite uses a single-writer model. Only one write transaction can execute at a time (readers don't block). This naturally gives serializable behavior.

### 5. MySQL (InnoDB) — Tunable

```sql
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
-- or globally:
SET GLOBAL TRANSACTION ISOLATION LEVEL SERIALIZABLE;
```

At serializable level, InnoDB converts all plain SELECTs to `SELECT ... LOCK IN SHARE MODE`, effectively acquiring shared locks on every read. This prevents phantoms and all anomalies but significantly reduces concurrency.

---

## What Happens with Weak Isolation? (The Horror Stories)

When isolation is too weak (e.g., Read Uncommitted or even Read Committed in some cases), you get real bugs:

### 1. Dirty Reads → Wrong Business Decisions

```
Scenario: E-commerce inventory system at Read Uncommitted

Txn A: UPDATE products SET stock = 0 WHERE id = 42;  -- preparing to restock
       (hasn't committed — still processing)

Txn B: SELECT stock FROM products WHERE id = 42;  → 0
       Website shows "OUT OF STOCK" to customers
       Customers leave, revenue lost

Txn A: UPDATE products SET stock = 500 WHERE id = 42;
Txn A: COMMIT
       Stock was never actually 0. But customers already left.
```

### 2. Lost Updates → Money Disappears

```
Scenario: Bank account with balance = 1000, Read Committed level

Txn A: reads balance → 1000
Txn B: reads balance → 1000

Txn A: writes balance = 1000 - 200 = 800
Txn A: COMMIT

Txn B: writes balance = 1000 + 500 = 1500  (based on stale read!)
Txn B: COMMIT

Final balance: 1500
Expected balance: 1000 - 200 + 500 = 1300

$200 just vanished. This is the classic "lost update" problem.
```

Real-life analogy: Two people share a Google Doc (without real-time sync). Both download it, both make edits, both upload. The second upload overwrites the first person's changes. Their work is lost.

### 3. Write Skew → Constraint Violations

```
Scenario: Hospital on-call system. Rule: at least 1 doctor must be on call.
Currently: Alice=on_call, Bob=on_call. Repeatable Read level.

Txn A (Alice): SELECT COUNT(*) FROM doctors WHERE on_call = true;  → 2
               "2 on call, safe for me to leave"
               UPDATE doctors SET on_call = false WHERE name = 'Alice';

Txn B (Bob):   SELECT COUNT(*) FROM doctors WHERE on_call = true;  → 2
               "2 on call, safe for me to leave"
               UPDATE doctors SET on_call = false WHERE name = 'Bob';

Both commit. Result: 0 doctors on call. Constraint violated.
Neither transaction saw the other's write because they read different rows.
This is "write skew" — only prevented at Serializable level.
```

### 4. Phantom Reads → Incorrect Aggregations

```
Scenario: Financial reporting at Repeatable Read (standard SQL, not MySQL/PG)

Txn A: SELECT SUM(amount) FROM transactions WHERE date = '2024-01-15';  → $10,000

Txn B: INSERT INTO transactions (amount, date) VALUES (5000, '2024-01-15');
Txn B: COMMIT

Txn A: SELECT COUNT(*) FROM transactions WHERE date = '2024-01-15';
       → includes the new row!

Txn A's SUM and COUNT are inconsistent within the same transaction.
The report is wrong.
```

---

## Isolation in Distributed Systems — Inspired by Database Isolation

In distributed systems, we face the exact same isolation problems, just across services instead of across rows. Here's a real scenario where we can directly apply database isolation thinking:

### Scenario: Distributed Booking System (Hotel + Flight)

```
The Problem:
  User wants to book a vacation package: Hotel + Flight
  Hotel Service and Flight Service are separate microservices with separate databases.

  User A: Books last hotel room + last flight seat
  User B: Books last hotel room + last flight seat (concurrently)

Without isolation:
  t=1: User A checks hotel → 1 room available ✅
  t=2: User B checks hotel → 1 room available ✅
  t=3: User A checks flight → 1 seat available ✅
  t=4: User B checks flight → 1 seat available ✅
  t=5: User A books hotel → success (room taken)
  t=6: User B books hotel → FAIL (no rooms)
  t=7: User A books flight → success
  t=8: User B books flight → FAIL
  
  User B wasted time. But worse:
  What if t=6 succeeded due to a race condition? Now we've overbooked.
```

### Solution: Apply Database Isolation Patterns

#### Pattern 1: Pessimistic Locking (like DB row locks)

```
Distributed Lock (using Redis/ZooKeeper):

  User A: ACQUIRE_LOCK("hotel:room:101")  → got it
  User A: ACQUIRE_LOCK("flight:seat:14A") → got it
  User A: Book both → success
  User A: RELEASE_LOCK("hotel:room:101")
  User A: RELEASE_LOCK("flight:seat:14A")

  User B: ACQUIRE_LOCK("hotel:room:101")  → BLOCKED until A releases
  User B: (after A releases) → room already booked → fail fast, no wasted work

This is exactly how databases do exclusive row locks.
```

```
Implementation sketch:

  func bookPackage(userID, hotelRoom, flightSeat) {
      lock1 := redis.Lock("hotel:" + hotelRoom, timeout=30s)
      lock2 := redis.Lock("flight:" + flightSeat, timeout=30s)
      
      defer lock1.Release()
      defer lock2.Release()
      
      // Now we have "isolation" — no other transaction can touch these resources
      if !hotelService.IsAvailable(hotelRoom) {
          return Error("room not available")
      }
      if !flightService.IsAvailable(flightSeat) {
          return Error("seat not available")
      }
      
      hotelService.Book(hotelRoom, userID)
      flightService.Book(flightSeat, userID)
      return Success
  }
```

#### Pattern 2: Optimistic Concurrency (like MVCC / SSI)

```
Instead of locking, use version checks:

  Hotel Room record: { id: 101, status: "available", version: 5 }

  User A reads: version=5
  User B reads: version=5

  User A books: UPDATE room SET status='booked' WHERE id=101 AND version=5
                → Success. version becomes 6.

  User B books: UPDATE room SET status='booked' WHERE id=101 AND version=5
                → FAILS. version is now 6, not 5. Conflict detected.
                → User B retries or gets "room no longer available."

This is exactly how MVCC works — read a snapshot, validate at commit time.
```

Real-life analogy: Two people trying to edit the same Wikipedia article. Both load version 5. First person saves — becomes version 6. Second person tries to save — Wikipedia says "edit conflict, someone changed the article since you loaded it." No data is lost, no corruption.

#### Pattern 3: Saga with Compensation (for eventual isolation)

```
When you can't hold locks across services (too slow, too risky):

  Step 1: Book hotel → success (tentative)
  Step 2: Book flight → FAIL
  Step 3: Compensate: Cancel hotel booking (undo step 1)

  This is like a database transaction rollback, but distributed.
  Each service provides a "compensating action" (the undo).

  Timeline:
    Hotel: available → tentatively_booked → cancelled (compensated)
    Flight: available → booking_failed

  The "tentatively_booked" state is like an uncommitted write in a database.
  Other users see it as "unavailable" (preventing dirty reads).
  If the saga fails, the compensation restores the original state.
```

### The Mapping: DB Isolation → Distributed System Isolation

```
┌─────────────────────────┬──────────────────────────────────────┐
│ Database Concept        │ Distributed System Equivalent        │
├─────────────────────────┼──────────────────────────────────────┤
│ Row-level lock          │ Distributed lock (Redis/ZooKeeper)   │
│ MVCC / versioning       │ Optimistic concurrency (version CAS) │
│ Transaction rollback    │ Saga compensation                    │
│ Serializable isolation  │ Distributed consensus (Raft/Paxos)   │
│ Snapshot isolation      │ Event sourcing (read from snapshot)   │
│ Dirty read prevention   │ "Tentative" status in saga pattern   │
│ Deadlock detection      │ Lock ordering / timeout-based release │
└─────────────────────────┴──────────────────────────────────────┘
```

---

## Key Takeaways

1. Isolation is row-level in modern databases, not table-level. Two transactions touching different rows don't interfere.

2. Databases achieve isolation via locking (pessimistic) or MVCC (optimistic, modern). Most use MVCC + locks together.

3. Four isolation levels exist: Read Uncommitted → Read Committed → Repeatable Read → Serializable. Each prevents more anomalies but costs more performance.

4. Most databases default to Read Committed (PostgreSQL, Oracle, SQL Server). MySQL defaults to Repeatable Read. CockroachDB and Spanner default to Serializable.

5. Weak isolation causes real bugs: dirty reads, lost updates, write skew, phantom reads. These aren't theoretical — they cause money loss, data corruption, and constraint violations.

6. Distributed systems face the same isolation challenges. The patterns (locking, versioning, compensation) are directly inspired by how databases solve isolation internally.

The fundamental insight: isolation is about controlling visibility. Who can see what, and when. Whether it's rows in a database or resources across microservices, the problem and the solutions are structurally identical.
