# Durability in ACID

Durability means: once a transaction is committed, its changes are permanent — even if the database crashes, the server loses power, or the OS panics one millisecond later. The data survives. Period.

This is the "D" in ACID, and it's arguably the most critical property. Atomicity, consistency, and isolation are all meaningless if committed data can vanish.

Real-life analogy: Think of durability like carving your name into wet concrete. Once the concrete dries (commit), your name is there permanently. Rain, wind, someone stepping on it — doesn't matter. It's etched in. Compare that to writing your name in sand (in-memory only) — one wave (crash) and it's gone.

---

## What Happens When the DB Receives a Write Operation?

When you execute something like `UPDATE accounts SET balance = 500 WHERE id = 1`, the database does NOT immediately write to the final data file on disk. Here's the actual sequence:

```
  Your Query                    Database Internals
  ──────────                    ──────────────────
  
  UPDATE accounts               ┌─────────────────────────────────────┐
  SET balance = 500             │  Step 1: BUFFER POOL (RAM)          │
  WHERE id = 1;                 │  Load the data page containing      │
                                │  row id=1 into memory (if not       │
                                │  already cached).                    │
                                │  Old value: balance = 1000           │
                                └──────────────┬──────────────────────┘
                                               │
                                               ▼
                                ┌─────────────────────────────────────┐
                                │  Step 2: WAL (Write-Ahead Log)      │
                                │  Write a log record to the WAL:     │
                                │  {TX-101, accounts, id=1,           │
                                │   old=1000, new=500}                │
                                │  FLUSH WAL TO DISK (fsync)          │
                                │  ← This is the critical step        │
                                └──────────────┬──────────────────────┘
                                               │
                                               ▼
                                ┌─────────────────────────────────────┐
                                │  Step 3: MODIFY IN MEMORY           │
                                │  Change the in-memory page:         │
                                │  balance = 500                      │
                                │  Mark page as "dirty"               │
                                └──────────────┬──────────────────────┘
                                               │
                                               ▼
                                ┌─────────────────────────────────────┐
                                │  Step 4: COMMIT                     │
                                │  Write COMMIT record to WAL         │
                                │  FLUSH WAL TO DISK (fsync)          │
                                │  Return "OK" to client              │
                                │  ← At this point, it's DURABLE      │
                                └──────────────┬──────────────────────┘
                                               │
                                               ▼
                                ┌─────────────────────────────────────┐
                                │  Step 5: LAZY FLUSH (later)         │
                                │  Background process (checkpointer)  │
                                │  eventually writes the dirty page   │
                                │  from memory to the actual data     │
                                │  file on disk.                      │
                                │  This can happen seconds, minutes,  │
                                │  or even hours later.               │
                                └─────────────────────────────────────┘
```

The key insight: the database says "committed" to you after Step 4, NOT after Step 5. The actual data file might still have the old value on disk. But the WAL on disk has the new value, and that's enough to guarantee durability.

```
  So to directly answer your question:

  ┌──────────────────────────────────────────────────────────────┐
  │                                                              │
  │  Q: Does it store to file, in-memory, or WAL?               │
  │                                                              │
  │  A: ALL THREE, in this order:                                │
  │                                                              │
  │  1. WAL file on disk    ← written FIRST (sequential write)  │
  │  2. Buffer pool in RAM  ← modified in memory (fast)         │
  │  3. Data files on disk  ← written LATER (lazy, background)  │
  │                                                              │
  │  Durability comes from #1 — the WAL flush.                  │
  │  #2 is for performance (serve reads from memory).            │
  │  #3 is the eventual "real" storage.                          │
  │                                                              │
  └──────────────────────────────────────────────────────────────┘
```

---

## Wait — Two Different "Writes to Disk"? Clarifying the Confusion

This is the part that trips everyone up. There are TWO completely different disk writes happening at different times, for different purposes. Let's walk through your exact example:

```
  Query: UPDATE accounts SET balance = 500 WHERE id = 1;

  Let's trace EVERY step, numbering the disk writes:
  ══════════════════════════════════════════════════════

  Step 1: Load page into buffer pool
  ┌──────────────────────────────────────────────────────────────┐
  │  The page containing row id=1 is read from the DATA FILE     │
  │  on disk into the buffer pool (RAM).                         │
  │  Engine sees: balance = 1000 (old value).                    │
  │                                                              │
  │  No disk write here — this is a READ.                        │
  └──────────────────────────────────────────────────────────────┘

  Step 2: Write WAL record to disk                    ← DISK WRITE #1
  ┌──────────────────────────────────────────────────────────────┐
  │  Engine constructs a WAL record:                             │
  │  { TX-100, accounts, id=1, old=1000, new=500 }              │
  │                                                              │
  │  This record is appended to the WAL FILE on disk.            │
  │  This is a SEQUENTIAL write to the WAL file.                 │
  └──────────────────────────────────────────────────────────────┘

  Step 3: Modify the page in memory (RAM)
  ┌──────────────────────────────────────────────────────────────┐
  │  The in-memory copy of the page in the buffer pool is        │
  │  modified: balance changes from 1000 → 500.                  │
  │  The page is marked as "dirty" (modified but not yet         │
  │  written to the data file).                                  │
  │                                                              │
  │  No disk write here — this is a RAM operation.               │
  └──────────────────────────────────────────────────────────────┘

  Step 4: COMMIT — flush WAL to disk with fsync       ← DISK WRITE #2
  ┌──────────────────────────────────────────────────────────────┐
  │  Engine writes a COMMIT record to the WAL:                   │
  │  { TX-100 COMMIT }                                           │
  │                                                              │
  │  Then calls fsync() on the WAL file.                         │
  │  This forces the OS to push the WAL data from the OS page    │
  │  cache all the way down to the physical disk platters/flash.  │
  │                                                              │
  │  AFTER fsync returns → client gets "COMMIT OK."              │
  │  At this point, the transaction is DURABLE.                  │
  │                                                              │
  │  ⚠️  NOTE: fsync is called on the WAL FILE, not the data    │
  │  file. The data file still has the old value (1000) on disk. │
  └──────────────────────────────────────────────────────────────┘

  ... time passes (seconds, minutes, maybe longer) ...

  Step 5: CHECKPOINT — flush dirty page to data file  ← DISK WRITE #3
  ┌──────────────────────────────────────────────────────────────┐
  │  A background process (the checkpointer) eventually runs.    │
  │  It takes the dirty page from the buffer pool and writes     │
  │  it DIRECTLY to the DATA FILE on disk. Then calls fsync()    │
  │  on the data file.                                           │
  │                                                              │
  │  Buffer Pool (RAM)  ──write()──>  Data File  ──fsync()──> Disk│
  │  (dirty page)          directly   (not WAL!)                 │
  │                                                              │
  │  NOW the data file has balance = 500.                        │
  │  The page is no longer dirty.                                │
  │                                                              │
  │  This write is to the DATA FILE, not the WAL file.           │
  │  This is a RANDOM write (the page goes to its specific       │
  │  location in the data file).                                 │
  │                                                              │
  │  "But doesn't this skip the WAL? Isn't that unsafe?"         │
  │                                                              │
  │  No — because the WAL already recorded this change back in   │
  │  Step 2, BEFORE the page was even modified in memory.        │
  │  The WAL's job is already done. This flush is just the       │
  │  data file catching up to what the WAL already knows.        │
  │                                                              │
  │  The full path of the data is:                               │
  │                                                              │
  │  ┌───────┐  Step 2  ┌───────┐  Step 3  ┌────────────┐       │
  │  │ Query │ ──────> │  WAL  │ ──────> │ Buffer Pool│       │
  │  └───────┘  write  │ (disk)│  modify │   (RAM)    │       │
  │             log    └───────┘  page   │ dirty page │       │
  │             first             in RAM  └─────┬──────┘       │
  │                                             │              │
  │                                    Step 5   │ (much later) │
  │                                  checkpoint │              │
  │                                             ▼              │
  │                                      ┌───────────┐         │
  │                                      │ Data File │         │
  │                                      │  (disk)   │         │
  │                                      └───────────┘         │
  │                                                              │
  │  The dirty page goes: WAL → RAM → Data File                  │
  │  It does NOT go: RAM → WAL → Data File                       │
  │  The WAL was ALREADY written. The page flush bypasses it.    │
  └──────────────────────────────────────────────────────────────┘
```

### The Two Writes Side by Side

```
  ┌──────────────────────────────────────────────────────────────────┐
  │                                                                  │
  │  DISK WRITE #1 + #2: WAL write + fsync (at COMMIT time)         │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  WHAT:    WAL record + COMMIT marker                      │  │
  │  │  WHERE:   WAL file (pg_wal/ or ib_logfile)                │  │
  │  │  WHEN:    During the transaction, at commit time           │  │
  │  │  HOW:     Sequential append + fsync                       │  │
  │  │  WHY:     This IS the durability guarantee                │  │
  │  │  SPEED:   Fast (sequential I/O)                           │  │
  │  │  BLOCKS:  Yes — client waits for this                     │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  DISK WRITE #3: Dirty page flush (at CHECKPOINT time)            │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  WHAT:    The actual modified data page (8KB/16KB)         │  │
  │  │  WHERE:   Data file (base/16384/16385 or .ibd file)       │  │
  │  │  WHEN:    Later — background process, not during TX        │  │
  │  │  HOW:     Random write to the page's location in the file │  │
  │  │  WHY:     So the data file catches up with the WAL,       │  │
  │  │           and old WAL segments can be recycled             │  │
  │  │  SPEED:   Slower (random I/O)                             │  │
  │  │  BLOCKS:  No — client doesn't wait for this               │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  KEY INSIGHT:                                                    │
  │  The dirty page flush (Write #3) is NOT part of the durability   │
  │  guarantee. It's a performance optimization. The WAL already     │
  │  has everything needed to reconstruct the data. The dirty page   │
  │  flush just keeps the data files up-to-date so that:             │
  │  1. Recovery is faster (less WAL to replay)                      │
  │  2. Old WAL segments can be freed                                │
  │  3. The buffer pool doesn't fill up with dirty pages             │
  │                                                                  │
  └──────────────────────────────────────────────────────────────────┘
```

### Your Specific Question Answered

```
  ┌──────────────────────────────────────────────────────────────────┐
  │                                                                  │
  │  Q: "The dirty page will be directly flushed to disk? Will it    │
  │      not be written first to WAL?"                               │
  │                                                                  │
  │  A: The dirty page IS already covered by the WAL.                │
  │                                                                  │
  │     The WAL record was written BEFORE the page was modified      │
  │     in memory (Step 2 happened before Step 3). So by the time    │
  │     the page becomes dirty, the WAL already has a record of      │
  │     what changed.                                                │
  │                                                                  │
  │     When the checkpointer flushes the dirty page to the data     │
  │     file (Step 5), it's writing the ACTUAL PAGE — the full       │
  │     8KB/16KB block with the row data. This is NOT a WAL          │
  │     operation. It's updating the data file directly.             │
  │                                                                  │
  │     Think of it this way:                                        │
  │     - WAL = "here's what CHANGED" (compact log record)           │
  │     - Data file = "here's the CURRENT STATE" (full page)         │
  │                                                                  │
  │     The WAL write happens during the transaction.                │
  │     The data file write happens later, in the background.        │
  │     Both are on disk, but they're different files with           │
  │     different purposes.                                          │
  │                                                                  │
  └──────────────────────────────────────────────────────────────────┘
```

### What If We Crash Between WAL Flush and Dirty Page Flush?

```
  State after COMMIT but before checkpoint:
  ┌──────────────────────────────────────────────────────────────┐
  │                                                              │
  │  WAL file on disk:     TX-100: balance 1000 → 500, COMMIT ✅│
  │  Data file on disk:    balance = 1000 (STALE!)               │
  │  Buffer pool (RAM):    balance = 500, dirty page (GONE 💀)   │
  │                                                              │
  │  💥 CRASH                                                    │
  │                                                              │
  │  On restart:                                                 │
  │  1. Recovery reads WAL → finds TX-100 is committed           │
  │  2. REDO: applies "balance 1000 → 500" to the data page      │
  │  3. Data file now has balance = 500 ✅                        │
  │                                                              │
  │  No data lost. The WAL had everything needed.                │
  │  The dirty page flush was never necessary for correctness —  │
  │  it's just an optimization.                                  │
  │                                                              │
  └──────────────────────────────────────────────────────────────┘
```

### Analogy: The Notebook and the Whiteboard

```
  ┌──────────────────────────────────────────────────────────────┐
  │                                                              │
  │  Imagine you're a teacher tracking student grades:            │
  │                                                              │
  │  NOTEBOOK (= WAL file):                                      │
  │  A permanent, ink-written log of every grade change.          │
  │  "Oct 15: Changed Alice's grade from B to A"                 │
  │  You write here FIRST, in ink. It's the official record.     │
  │                                                              │
  │  WHITEBOARD (= Buffer pool / RAM):                           │
  │  The current grades, easy to read and update.                │
  │  You update this right after writing in the notebook.         │
  │  Fast to look at, but gets erased if someone bumps it.       │
  │                                                              │
  │  GRADE BINDER (= Data file on disk):                         │
  │  The official printed grade sheets.                           │
  │  You update these at the end of the week (= checkpoint).     │
  │  Slow to update (need to find the right page, erase, rewrite)│
  │                                                              │
  │  If the whiteboard gets erased (crash):                      │
  │  → Open the notebook, replay changes → rebuild whiteboard    │
  │  → Then update the grade binder when convenient               │
  │                                                              │
  │  The notebook (WAL) is NEVER skipped.                        │
  │  The whiteboard (RAM) is fast but volatile.                  │
  │  The grade binder (data file) is updated lazily.             │
  │                                                              │
  │  The dirty page flush = updating the grade binder from       │
  │  the whiteboard. It doesn't go through the notebook again    │
  │  because the notebook already has the record.                │
  │                                                              │
  └──────────────────────────────────────────────────────────────┘
```

---

## On Disk: Where Does It Actually Store Data?

Databases store everything in regular files on the filesystem. There's no magic hardware or special OS structure. It's just files — but very carefully organized files.

### The File Layout (PostgreSQL as example)

```
  /var/lib/postgresql/16/main/          ← data directory (PGDATA)
  │
  ├── base/                             ← actual database files
  │   ├── 16384/                        ← database OID (one folder per database)
  │   │   ├── 16385                     ← table data file (heap file)
  │   │   ├── 16385.1                   ← overflow segment (if table > 1GB)
  │   │   ├── 16386                     ← another table
  │   │   ├── 16387                     ← index file (B+ Tree on disk)
  │   │   └── 16385_fsm                 ← free space map
  │   │   └── 16385_vm                  ← visibility map (for MVCC)
  │   └── ...
  │
  ├── pg_wal/                           ← WAL files (the write-ahead log)
  │   ├── 000000010000000000000001      ← WAL segment (16MB each)
  │   ├── 000000010000000000000002
  │   └── ...
  │
  ├── pg_xact/                          ← transaction commit status
  ├── pg_multixact/                     ← multi-transaction status
  ├── global/                           ← shared system catalogs
  └── postgresql.conf                   ← configuration
```

### How Data Files Are Structured Internally

Each table's data file is divided into fixed-size pages (typically 8KB in PostgreSQL, 16KB in MySQL InnoDB):

```
  Table Data File (e.g., "accounts" table)
  ═════════════════════════════════════════

  ┌──────────┬──────────┬──────────┬──────────┬──────────┐
  │  Page 0  │  Page 1  │  Page 2  │  Page 3  │   ...    │
  │  (8 KB)  │  (8 KB)  │  (8 KB)  │  (8 KB)  │          │
  └──────────┴──────────┴──────────┴──────────┴──────────┘

  Inside each page:
  ┌─────────────────────────────────────────────────────┐
  │  Page Header (24 bytes)                             │
  │  ┌─────────────────────────────────────────────┐    │
  │  │ LSN (Log Sequence Number) — links to WAL    │    │
  │  │ Checksum                                    │    │
  │  │ Free space pointers                         │    │
  │  └─────────────────────────────────────────────┘    │
  │                                                     │
  │  Item Pointers (array of offsets)                   │
  │  [ptr1] [ptr2] [ptr3] [ptr4] ...                    │
  │     │      │      │      │                          │
  │     ▼      ▼      ▼      ▼                          │
  │  ┌──────────────────────────────────────────────┐   │
  │  │         (free space in the middle)           │   │
  │  └──────────────────────────────────────────────┘   │
  │                                                     │
  │  Row Data (grows from bottom up):                   │
  │  ┌──────────────────────────────────────────────┐   │
  │  │ Row 4: {id=4, balance=300, xmin=..., xmax=..}│  │
  │  │ Row 3: {id=3, balance=700, ...}              │   │
  │  │ Row 2: {id=2, balance=200, ...}              │   │
  │  │ Row 1: {id=1, balance=500, ...}              │   │
  │  └──────────────────────────────────────────────┘   │
  └─────────────────────────────────────────────────────┘

  Item pointers grow top-down ↓
  Row data grows bottom-up ↑
  They meet in the middle when the page is full.
```

### MySQL InnoDB: Slightly Different

InnoDB uses a tablespace model:

```
  /var/lib/mysql/
  │
  ├── ibdata1                    ← system tablespace (shared)
  ├── ib_logfile0                ← redo log (InnoDB's WAL)
  ├── ib_logfile1                ← redo log (rotated)
  │
  └── mydb/
      ├── accounts.ibd           ← table data + indexes (file-per-table)
      └── orders.ibd             ← another table
```

InnoDB pages are 16KB (vs PostgreSQL's 8KB), and the table data is stored in a clustered index (B+ Tree ordered by primary key) — more on this in the index section below.

---

## How Does This Ensure Writes Are Durable?

The durability guarantee comes from a chain of mechanisms working together:

### 1. WAL + fsync = The Foundation

```
  The critical moment:

  ┌──────────────────────────────────────────────────────────┐
  │                                                          │
  │  Application: "COMMIT"                                   │
  │       │                                                  │
  │       ▼                                                  │
  │  Database: Write COMMIT record to WAL buffer             │
  │       │                                                  │
  │       ▼                                                  │
  │  Database: fsync() the WAL file                          │
  │       │                                                  │
  │       │  ┌─────────────────────────────────────────┐     │
  │       │  │  fsync() tells the OS:                  │     │
  │       │  │  "Flush this file from OS page cache    │     │
  │       │  │   to the PHYSICAL DISK PLATTERS. Now.   │     │
  │       │  │   Don't return until the bits are       │     │
  │       │  │   actually on the magnetic surface      │     │
  │       │  │   (or flash cells for SSD)."            │     │
  │       │  └─────────────────────────────────────────┘     │
  │       │                                                  │
  │       ▼                                                  │
  │  Database: Return "COMMIT OK" to application             │
  │                                                          │
  │  At this point, even if power goes out in the next       │
  │  nanosecond, the WAL record is on physical disk.         │
  │  On restart, recovery replays it. Data is safe.          │
  │                                                          │
  └──────────────────────────────────────────────────────────┘
```

Why fsync matters:

```
  WITHOUT fsync:
  ┌──────────┐    write()    ┌──────────────┐    ???     ┌──────────┐
  │ Database │ ────────────> │  OS Page     │ ────────> │  Disk    │
  │          │               │  Cache (RAM) │  maybe    │          │
  └──────────┘               └──────────────┘  later    └──────────┘

  The OS might cache the write in RAM and say "done!" to the database.
  If power goes out, the OS cache is lost. Data gone. 💀

  WITH fsync:
  ┌──────────┐    write()    ┌──────────────┐  fsync()  ┌──────────┐
  │ Database │ ────────────> │  OS Page     │ ────────> │  Disk    │
  │          │               │  Cache (RAM) │  forced   │  ✅ safe │
  └──────────┘               └──────────────┘  flush    └──────────┘

  fsync() blocks until the OS confirms the data is on physical media.
```

### 2. Why WAL Writes Are Fast (Sequential I/O)

You might wonder: "If we fsync on every commit, isn't that slow?" The answer is that WAL writes are sequential, which is fundamentally different from random writes:

```
  WAL writes (sequential):
  ┌────────────────────────────────────────────────────┐
  │  Disk head stays in one place, data streams in:    │
  │                                                    │
  │  [record1][record2][record3][record4][record5]...  │
  │  ──────────────────────────────────────────────>   │
  │  One continuous stream. Disk head barely moves.    │
  │                                                    │
  │  Speed: 200-400 MB/s (HDD), 1-3 GB/s (SSD)       │
  └────────────────────────────────────────────────────┘

  Data file writes (random):
  ┌────────────────────────────────────────────────────┐
  │  Disk head jumps around to update different pages:  │
  │                                                    │
  │  Page 47 ──jump──> Page 3 ──jump──> Page 891       │
  │  ──jump──> Page 12 ──jump──> Page 456              │
  │                                                    │
  │  Constant seeking. Very slow on HDD.               │
  │                                                    │
  │  Speed: 1-5 MB/s (HDD), 500 MB/s (SSD)            │
  └────────────────────────────────────────────────────┘

  This is why the database writes to the WAL first (fast sequential)
  and delays the data file writes (slow random) to a background process.
  You get durability AND performance.
```

### 3. Checkpoints — Bridging WAL and Data Files

Checkpoints periodically flush dirty pages from the buffer pool to the data files, so the WAL doesn't grow forever:

```
  Timeline:
  ═════════

  ──[writes]──[writes]──[CHECKPOINT]──[writes]──[writes]──[CHECKPOINT]──
                              │                                 │
                              ▼                                 ▼
                    Flush all dirty pages           Flush all dirty pages
                    to data files on disk.          to data files on disk.
                    WAL before this point           WAL before this point
                    can be recycled.                can be recycled.

  After a checkpoint:
  ┌──────────────────────────────────────────────────────────┐
  │  Data files on disk are up-to-date (as of checkpoint).   │
  │  WAL segments before the checkpoint are no longer needed │
  │  for recovery — they can be deleted or recycled.         │
  │                                                          │
  │  On crash, recovery only needs to replay WAL records     │
  │  AFTER the last checkpoint. Much faster recovery.        │
  └──────────────────────────────────────────────────────────┘
```

### 4. The Full Durability Chain

```
  ┌─────────────────────────────────────────────────────────────────┐
  │                    DURABILITY GUARANTEE CHAIN                   │
  │                                                                 │
  │  Layer 1: WAL + fsync                                           │
  │  → Every committed transaction's changes are on physical disk   │
  │    in the WAL before "OK" is returned to the client.            │
  │                                                                 │
  │  Layer 2: Checkpoints                                           │
  │  → Periodically, dirty pages are flushed to data files.         │
  │    This keeps recovery time bounded.                            │
  │                                                                 │
  │  Layer 3: Crash Recovery (ARIES-style)                          │
  │  → On restart, replay WAL from last checkpoint.                 │
  │    REDO committed transactions, UNDO uncommitted ones.          │
  │    Database returns to a consistent, durable state.             │
  │                                                                 │
  │  Layer 4: Page Checksums                                        │
  │  → Each page has a checksum. If a page is corrupted             │
  │    (partial write, bit rot), the DB detects it and              │
  │    can recover from WAL or backup.                              │
  │                                                                 │
  │  Layer 5: Replication (optional but common)                     │
  │  → Synchronous replication: WAL records are sent to a           │
  │    standby server and fsync'd there too. Even if the            │
  │    primary's disk catches fire, the standby has the data.       │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘
```

---

## What Happens If the Disk Crashes or DB Crashes?

### Scenario 1: Database Process Crashes (but disk is fine)

```
  ┌──────────────────────────────────────────────────────────────┐
  │  DB process dies (OOM kill, segfault, kill -9)               │
  │                                                              │
  │  What's lost:                                                │
  │  → Buffer pool (RAM) — all dirty pages in memory are gone    │
  │  → Any in-flight transactions that hadn't committed          │
  │                                                              │
  │  What survives:                                              │
  │  → WAL on disk — intact (it was fsync'd)                     │
  │  → Data files on disk — intact (might be slightly stale)     │
  │                                                              │
  │  Recovery:                                                   │
  │  1. Restart the database process                             │
  │  2. Recovery manager reads WAL from last checkpoint           │
  │  3. REDO: replay committed transactions to data files        │
  │  4. UNDO: rollback uncommitted transactions                  │
  │  5. Database is back to a consistent state                   │
  │  6. Accept new connections                                   │
  │                                                              │
  │  Data loss: ZERO for committed transactions.                 │
  │  Uncommitted transactions are rolled back (by design).       │
  └──────────────────────────────────────────────────────────────┘
```

### Scenario 2: Power Failure (everything in RAM is gone)

```
  ┌──────────────────────────────────────────────────────────────┐
  │  Power goes out. Server shuts down instantly.                │
  │                                                              │
  │  What's lost:                                                │
  │  → Everything in RAM (buffer pool, OS page cache)            │
  │  → Any writes that were in the OS page cache but not         │
  │    fsync'd to disk                                           │
  │                                                              │
  │  What survives:                                              │
  │  → WAL on disk — committed records were fsync'd ✅           │
  │  → Data files on disk — whatever was flushed before power    │
  │    loss (might be stale, but WAL covers the gap)             │
  │                                                              │
  │  Recovery: Same as Scenario 1.                               │
  │  WAL has everything needed to bring data files up to date.   │
  │                                                              │
  │  Data loss: ZERO for committed transactions.                 │
  │                                                              │
  │  ⚠️  CAVEAT: This assumes the disk's write cache is          │
  │  properly configured. Some disks have a volatile write       │
  │  cache that lies about fsync (says "done" but data is        │
  │  still in the disk's RAM cache). Enterprise SSDs have        │
  │  capacitors to flush their cache on power loss.              │
  │  Consumer SSDs... not always.                                │
  └──────────────────────────────────────────────────────────────┘
```

### Scenario 3: Disk Failure (disk is physically dead)

```
  ┌──────────────────────────────────────────────────────────────┐
  │  Disk dies. All data on it is gone.                          │
  │                                                              │
  │  This is the ONE scenario where the single-node durability   │
  │  guarantee breaks. If the disk is dead, WAL and data files   │
  │  are both gone.                                              │
  │                                                              │
  │  Protection mechanisms:                                      │
  │                                                              │
  │  1. RAID (hardware level)                                    │
  │     RAID 1: Mirror — data written to 2 disks simultaneously  │
  │     RAID 10: Striped mirrors — performance + redundancy      │
  │     One disk dies → other disk has the data                  │
  │                                                              │
  │  2. Replication (database level)                              │
  │     ┌──────────┐  WAL stream  ┌──────────┐                  │
  │     │ Primary  │ ──────────> │ Standby  │                   │
  │     │ (disk 1) │             │ (disk 2) │                   │
  │     └──────────┘             └──────────┘                   │
  │                                                              │
  │     Synchronous replication: standby confirms WAL is         │
  │     fsync'd before primary returns "COMMIT OK."              │
  │     → Zero data loss even if primary disk explodes.          │
  │                                                              │
  │     Asynchronous replication: standby might be slightly      │
  │     behind. You could lose the last few seconds of commits.  │
  │                                                              │
  │  3. Backups (last resort)                                    │
  │     pg_basebackup + WAL archiving → point-in-time recovery  │
  │     You lose data since the last backup/WAL archive.         │
  │                                                              │
  └──────────────────────────────────────────────────────────────┘
```

### Scenario 4: Partial Write (Torn Page)

```
  ┌──────────────────────────────────────────────────────────────┐
  │  Power fails WHILE a page is being written to disk.          │
  │  The page is half-old, half-new. Corrupted.                  │
  │                                                              │
  │  This is called a "torn page" or "partial write."            │
  │                                                              │
  │  How databases handle it:                                    │
  │                                                              │
  │  PostgreSQL: Full-page writes                                │
  │  → After each checkpoint, the FIRST time a page is modified, │
  │    the ENTIRE page image is written to the WAL.              │
  │  → On recovery, if a data page is torn, the full page image  │
  │    from the WAL replaces it. Then redo logs are applied.     │
  │                                                              │
  │  MySQL InnoDB: Doublewrite buffer                            │
  │  → Before writing dirty pages to their final location,       │
  │    InnoDB writes them to a "doublewrite buffer" area first.  │
  │  → If a page write is torn, InnoDB recovers the clean copy   │
  │    from the doublewrite buffer.                              │
  │                                                              │
  │  ┌─────────────────────────────────────────────────┐         │
  │  │  PostgreSQL approach:                           │         │
  │  │                                                 │         │
  │  │  WAL: [...][FULL PAGE IMAGE of page 47][...]    │         │
  │  │                                                 │         │
  │  │  If page 47 on disk is torn:                    │         │
  │  │  → Copy full page image from WAL → page 47      │         │
  │  │  → Then replay any subsequent WAL records        │         │
  │  │  → Page 47 is now correct ✅                    │         │
  │  └─────────────────────────────────────────────────┘         │
  │                                                              │
  │  ┌─────────────────────────────────────────────────┐         │
  │  │  InnoDB approach:                               │         │
  │  │                                                 │         │
  │  │  Step 1: Write page to doublewrite buffer       │         │
  │  │          (sequential area on disk) + fsync      │         │
  │  │  Step 2: Write page to its actual location      │         │
  │  │                                                 │         │
  │  │  If Step 2 is torn:                             │         │
  │  │  → Doublewrite buffer has the clean copy        │         │
  │  │  → Copy it to the actual location               │         │
  │  │  → Then apply redo log                          │         │
  │  └─────────────────────────────────────────────────┘         │
  │                                                              │
  └──────────────────────────────────────────────────────────────┘
```

---

## WAL Lifecycle — Does the WAL Keep the Data or Delete It?

Short answer: the WAL keeps the data until it's no longer needed for recovery, then it's recycled/deleted. The data is NOT removed immediately after flushing dirty pages to disk. There's a specific lifecycle.

### The WAL Retention Rule

```
  THE RULE:
  ═════════
  WAL records are kept until a CHECKPOINT confirms that all the dirty
  pages those records describe have been safely flushed to the data
  files on disk.

  Only AFTER a checkpoint can the WAL segments before that checkpoint
  be safely removed.
```

### The Full WAL Lifecycle

```
  Timeline of a WAL record's life:
  ═════════════════════════════════

  ┌──────────────────────────────────────────────────────────────────┐
  │                                                                  │
  │  Phase 1: BIRTH — WAL record is created                          │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  Transaction commits.                                     │  │
  │  │  WAL record written: "TX-101: A.balance 1000 → 500"       │  │
  │  │  fsync'd to disk.                                         │  │
  │  │  Client gets "COMMIT OK."                                 │  │
  │  │                                                           │  │
  │  │  At this point:                                           │  │
  │  │  - WAL on disk: has the record ✅                         │  │
  │  │  - Data file on disk: still has old value (1000) ⚠️       │  │
  │  │  - Buffer pool (RAM): has new value (500), page is dirty  │  │
  │  │                                                           │  │
  │  │  The WAL record is ESSENTIAL. If we crash now, the data   │  │
  │  │  file has stale data. WAL is the only place with the      │  │
  │  │  committed value.                                         │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                          │                                       │
  │                          ▼                                       │
  │  Phase 2: ALIVE — WAL record exists, dirty page not yet flushed  │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  Time passes. More transactions happen. More WAL records   │  │
  │  │  accumulate. The dirty page for A.balance=500 is still     │  │
  │  │  sitting in the buffer pool, waiting to be flushed.        │  │
  │  │                                                           │  │
  │  │  The WAL record is STILL ESSENTIAL. Data file still stale. │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                          │                                       │
  │                          ▼                                       │
  │  Phase 3: CHECKPOINT — dirty pages flushed to data files         │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  The checkpointer runs (periodically or when WAL grows     │  │
  │  │  too large).                                               │  │
  │  │                                                           │  │
  │  │  1. Flush ALL dirty pages from buffer pool to data files   │  │
  │  │     → A.balance=500 is now written to the data file ✅     │  │
  │  │  2. Write a CHECKPOINT record to the WAL:                  │  │
  │  │     "Checkpoint at LSN 12345 — all pages up to this        │  │
  │  │      point are safely on disk."                            │  │
  │  │  3. fsync the data files.                                  │  │
  │  │                                                           │  │
  │  │  At this point:                                           │  │
  │  │  - WAL on disk: still has the record                      │  │
  │  │  - Data file on disk: NOW has the new value (500) ✅       │  │
  │  │  - The WAL record is now REDUNDANT for recovery.           │  │
  │  │    (Data file already has the correct value.)              │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                          │                                       │
  │                          ▼                                       │
  │  Phase 4: DEATH — WAL segment recycled/deleted                   │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  The database knows: "Everything before checkpoint at      │  │
  │  │  LSN 12345 is safely in the data files."                   │  │
  │  │                                                           │  │
  │  │  WAL segments containing only records before LSN 12345     │  │
  │  │  are now safe to remove.                                   │  │
  │  │                                                           │  │
  │  │  PostgreSQL: recycles the WAL segment files (renames them  │  │
  │  │  for reuse, or deletes if too many).                       │  │
  │  │                                                           │  │
  │  │  MySQL InnoDB: redo log files are circular — the "tail"    │  │
  │  │  advances past the old records, effectively overwriting    │  │
  │  │  them.                                                     │  │
  │  │                                                           │  │
  │  │  The WAL record is GONE. But that's fine — the data file   │  │
  │  │  has the value. Recovery doesn't need it anymore.          │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  └──────────────────────────────────────────────────────────────────┘
```

### Visual Timeline

```
  WAL Segment File Timeline:
  ══════════════════════════

  WAL segments:  [seg 001] [seg 002] [seg 003] [seg 004] [seg 005] [seg 006]
                  │          │          │          │          │          │
  Contents:      TX-95      TX-98      TX-101     TX-105     TX-110     TX-115
                 TX-96      TX-99      TX-102     TX-106     TX-111     TX-116
                 TX-97      TX-100     TX-103     TX-107     TX-112
                                       TX-104     TX-108
                                                  TX-109

  Checkpoint happens at seg 004 (LSN in TX-108 range):
  "All dirty pages up to TX-108 are flushed to data files."

                  [seg 001] [seg 002] [seg 003] [seg 004] [seg 005] [seg 006]
                  ├─────────────────────────────┤├─────────────────────────────┤
                  │                              │                             │
                  ▼                              ▼                             │
              SAFE TO DELETE              STILL NEEDED                         │
              (data files have            (might have dirty pages              │
               these changes)              not yet flushed)                    │

  PostgreSQL action:
  → seg 001, 002, 003 are recycled (renamed to seg 007, 008, 009 for reuse)
  → seg 004+ are kept

  After recycling:
  [seg 004] [seg 005] [seg 006] [seg 007 (empty)] [seg 008 (empty)] [seg 009 (empty)]
```

### PostgreSQL vs MySQL: How They Handle WAL Cleanup

```
  ┌──────────────────────────────────────────────────────────────────┐
  │  PostgreSQL WAL Management                                       │
  │  ─────────────────────────                                       │
  │                                                                  │
  │  WAL files live in: pg_wal/ directory                            │
  │  Each segment: 16MB by default                                   │
  │                                                                  │
  │  Cleanup strategy:                                               │
  │  - After checkpoint, old segments are RECYCLED (renamed)         │
  │  - If too many segments accumulate, excess are DELETED           │
  │  - min_wal_size / max_wal_size control how many are kept         │
  │  - If archiving is enabled (WAL archiving for backups/replicas), │
  │    segments are NOT deleted until archived                       │
  │                                                                  │
  │  pg_wal/ directory:                                              │
  │  BEFORE checkpoint:  000000010000000000000001  (old, has data)   │
  │                      000000010000000000000002  (old, has data)   │
  │                      000000010000000000000003  (current)         │
  │                                                                  │
  │  AFTER checkpoint:   000000010000000000000003  (current)         │
  │                      000000010000000000000004  (recycled from 1) │
  │                      000000010000000000000005  (recycled from 2) │
  └──────────────────────────────────────────────────────────────────┘

  ┌──────────────────────────────────────────────────────────────────┐
  │  MySQL InnoDB Redo Log Management                                │
  │  ────────────────────────────────                                │
  │                                                                  │
  │  Redo log files: ib_logfile0, ib_logfile1 (circular)             │
  │  Total size: innodb_log_file_size × innodb_log_files_in_group    │
  │                                                                  │
  │  Cleanup strategy:                                               │
  │  - Redo log is CIRCULAR — it wraps around                        │
  │  - A "tail" pointer tracks the oldest needed record              │
  │  - A "head" pointer tracks where new records are written         │
  │  - After checkpoint, tail advances, freeing space                │
  │  - Old records are simply overwritten when head wraps around     │
  │                                                                  │
  │  ┌──────────────────────────────────────────────┐                │
  │  │         Circular Redo Log                     │                │
  │  │                                               │                │
  │  │              ┌─────────┐                      │                │
  │  │         ╱────│ current │────╲                  │                │
  │  │       ╱      │ write   │      ╲               │                │
  │  │     ╱        │ (head)  │        ╲             │                │
  │  │   ┌──────┐   └─────────┘   ┌──────┐          │                │
  │  │   │ free │                  │ active│          │                │
  │  │   │space │                  │records│          │                │
  │  │   └──────┘   ┌─────────┐   └──────┘          │                │
  │  │     ╲        │ oldest  │        ╱             │                │
  │  │       ╲      │ needed  │      ╱               │                │
  │  │         ╲────│ (tail)  │────╱                  │                │
  │  │              └─────────┘                      │                │
  │  │                                               │                │
  │  │  After checkpoint: tail moves forward,         │                │
  │  │  freeing space behind it for new writes.       │                │
  │  └──────────────────────────────────────────────┘                │
  └──────────────────────────────────────────────────────────────────┘
```

### Special Case: WAL Archiving (for replication and backups)

```
  ┌──────────────────────────────────────────────────────────────────┐
  │  Even after a checkpoint, WAL segments might be KEPT if:         │
  │                                                                  │
  │  1. Streaming Replication is active                              │
  │     → Standby server needs to replay WAL to stay in sync        │
  │     → WAL segments are kept until ALL standbys have received     │
  │       them (controlled by wal_keep_size in PostgreSQL)           │
  │                                                                  │
  │  2. WAL Archiving is enabled (for point-in-time recovery)        │
  │     → WAL segments are copied to an archive location before      │
  │       being recycled                                             │
  │     → archive_command = 'cp %p /backup/wal/%f'                   │
  │     → Segment is NOT recycled until archive_command succeeds     │
  │                                                                  │
  │  3. Logical Replication slots                                    │
  │     → Replication slots prevent WAL cleanup until the consumer   │
  │       has processed the records                                  │
  │     → ⚠️ A stale/abandoned replication slot can cause WAL to     │
  │       grow unbounded and fill the disk — a common production     │
  │       incident!                                                  │
  │                                                                  │
  │  Timeline with archiving:                                        │
  │                                                                  │
  │  [seg 001] ──archive──> /backup/wal/seg001 ──> then recycle      │
  │  [seg 002] ──archive──> /backup/wal/seg002 ──> then recycle      │
  │  [seg 003] ──still active, not archived yet──> KEEP              │
  │                                                                  │
  └──────────────────────────────────────────────────────────────────┘
```

### TL;DR — The Answer to Your Question

```
  ┌──────────────────────────────────────────────────────────────────┐
  │                                                                  │
  │  Q: After dirty pages are flushed to data files, does the WAL    │
  │     still contain the data or is it removed?                     │
  │                                                                  │
  │  A: The WAL STILL CONTAINS the data after the flush.             │
  │     It is NOT immediately removed.                               │
  │                                                                  │
  │     It's only removed/recycled AFTER a checkpoint confirms       │
  │     that all dirty pages covered by those WAL records have       │
  │     been safely written to the data files on disk.               │
  │                                                                  │
  │     And even then, it might be kept longer if:                   │
  │     - A standby replica still needs it                           │
  │     - WAL archiving hasn't copied it yet                         │
  │     - A replication slot is holding it                           │
  │                                                                  │
  │     The lifecycle is:                                            │
  │     WAL written → dirty page flushed → checkpoint → WAL recycled │
  │                                                                  │
  │     NOT:                                                         │
  │     WAL written → dirty page flushed → WAL immediately deleted   │
  │                                                                  │
  └──────────────────────────────────────────────────────────────────┘
```

---

## Real-Life Analogy: The Notary's Office

Think of database durability like a notary's office handling property transfers:

```
  ┌──────────────────────────────────────────────────────────────┐
  │                    THE NOTARY ANALOGY                        │
  │                                                              │
  │  You want to transfer property ownership from Alice to Bob.  │
  │                                                              │
  │  Step 1: LOGBOOK (= WAL)                                    │
  │  The notary writes the transfer in the official logbook      │
  │  with permanent ink, BEFORE updating the property registry.  │
  │  The logbook is kept in a fireproof safe.                    │
  │                                                              │
  │  Step 2: REGISTRY UPDATE (= Data file update)               │
  │  Later, a clerk updates the property registry (the big       │
  │  book of who owns what). This might happen hours later.      │
  │                                                              │
  │  Step 3: CONFIRMATION (= COMMIT OK)                          │
  │  The notary tells you "it's done" after the logbook entry,  │
  │  NOT after the registry update.                              │
  │                                                              │
  │  If the office burns down (crash):                           │
  │  → The logbook survives (fireproof safe = fsync to disk)     │
  │  → The registry can be rebuilt from the logbook              │
  │  → No property transfer is lost                              │
  │                                                              │
  │  If the clerk was halfway through updating the registry      │
  │  when the fire started (torn page):                          │
  │  → Doesn't matter. The logbook has the complete record.      │
  │  → New clerk rebuilds the registry from the logbook.         │
  │                                                              │
  │  Mapping:                                                    │
  │  ┌────────────────────┬──────────────────────────┐           │
  │  │ Notary's Office    │ Database                 │           │
  │  ├────────────────────┼──────────────────────────┤           │
  │  │ Logbook            │ WAL (Write-Ahead Log)    │           │
  │  │ Fireproof safe     │ fsync to disk            │           │
  │  │ Property registry  │ Data files (heap/B+Tree) │           │
  │  │ Clerk updating     │ Background checkpointer  │           │
  │  │ "It's done"        │ COMMIT OK to client      │           │
  │  │ Rebuild from log   │ Crash recovery (REDO)    │           │
  │  │ Second office copy │ Replication (standby)    │           │
  │  └────────────────────┴──────────────────────────┘           │
  │                                                              │
  └──────────────────────────────────────────────────────────────┘
```

---


## Durability of Indexes — Where and How Are They Stored?

Indexes are not some separate magical structure floating in the ether. They are stored on disk in the same way as table data — as files made up of pages. And they go through the exact same WAL-based durability pipeline.

### What Data Structure Are Indexes?

The dominant structure is the B+ Tree. Let's look at how it lives on disk:

```
  B+ Tree Index on "accounts.email" (on disk)
  ═════════════════════════════════════════════

  The B+ Tree is stored as a FILE on disk, made up of PAGES (8KB or 16KB each).
  Each page is a node in the tree.

  File: base/16384/16387  (PostgreSQL assigns a file per index)

  ┌─────────────────────────────────────────────────────────────┐
  │                     Root Page (Page 0)                      │
  │  ┌─────────────────────────────────────────────────────┐    │
  │  │  [jane@example.com]                                 │    │
  │  │   /                    \                            │    │
  │  │  ptr→Page 1           ptr→Page 2                    │    │
  │  └─────────────────────────────────────────────────────┘    │
  └─────────────────────────────────────────────────────────────┘

  ┌──────────────────────────┐  ┌──────────────────────────────┐
  │  Internal Page (Page 1)  │  │  Internal Page (Page 2)      │
  │  [alice@] [bob@] [dave@] │  │  [jane@] [mike@] [zara@]    │
  │   /    |     |      \    │  │   /    |     |       \       │
  │  P3   P4    P5      P6  │  │  P7   P8    P9      P10     │
  └──────────────────────────┘  └──────────────────────────────┘

  ┌──────────────────────────┐  ┌──────────────────────────────┐
  │  Leaf Page (Page 3)      │  │  Leaf Page (Page 4)          │
  │  ┌──────────┬──────────┐ │  │  ┌──────────┬──────────┐    │
  │  │ alice@.. │ anna@..  │ │  │  │ bob@..   │ carol@.. │    │
  │  │ →row ptr │ →row ptr │ │  │  │ →row ptr │ →row ptr │    │
  │  └──────────┴──────────┘ │  │  └──────────┴──────────┘    │
  │  next→Page 4             │  │  next→Page 5                │
  └──────────────────────────┘  └──────────────────────────────┘

  Key properties:
  - Each page is a node in the B+ Tree
  - Internal pages hold keys + pointers to child pages
  - Leaf pages hold keys + pointers to actual table rows (heap tuples)
  - Leaf pages are linked (next→) for efficient range scans
  - The whole thing is a regular file on disk
```

### How Index Writes Are Made Durable

Index modifications go through the exact same WAL pipeline as table data:

```
  INSERT INTO accounts (id, email) VALUES (5, 'eve@example.com');

  This triggers TWO sets of changes:
  ┌──────────────────────────────────────────────────────────────┐
  │                                                              │
  │  1. Table (heap) change:                                     │
  │     → Insert new row into a data page                        │
  │     → WAL record: "Insert row {id=5, email=eve@..} on       │
  │       page 12, offset 3 of table 16385"                      │
  │                                                              │
  │  2. Index change (for each index on the table):              │
  │     → Insert new key into B+ Tree                            │
  │     → WAL record: "Insert key 'eve@example.com' → (page 12, │
  │       offset 3) into index page 5 of index 16387"            │
  │                                                              │
  │  Both WAL records are written BEFORE the actual pages are    │
  │  modified on disk. Both are covered by the same COMMIT.      │
  │                                                              │
  └──────────────────────────────────────────────────────────────┘

  On crash recovery:
  → WAL replay reconstructs BOTH the table page AND the index page.
  → The index is always consistent with the table after recovery.
```

### Index Types and Their On-Disk Structures

```
  ┌──────────────────┬────────────────────┬──────────────────────────────┐
  │ Index Type       │ Data Structure     │ On-Disk Storage              │
  ├──────────────────┼────────────────────┼──────────────────────────────┤
  │ B-Tree / B+Tree  │ Balanced tree      │ File of 8/16KB pages, each  │
  │ (default, most   │ with sorted keys   │ page is a tree node.        │
  │  common)         │                    │ Leaf pages linked for scans. │
  │                  │                    │                              │
  │ Hash Index       │ Hash table with    │ File of pages organized as   │
  │                  │ buckets            │ hash buckets. Each bucket    │
  │                  │                    │ is one or more pages.        │
  │                  │                    │                              │
  │ GiST             │ Generalized search │ File of pages, tree-like.    │
  │ (PostgreSQL)     │ tree               │ Used for geometric, full-    │
  │                  │                    │ text, range queries.         │
  │                  │                    │                              │
  │ GIN              │ Inverted index     │ File of pages. Posting lists │
  │ (PostgreSQL)     │ (like search       │ stored in tree + overflow    │
  │                  │  engines)          │ pages.                       │
  │                  │                    │                              │
  │ BRIN             │ Block range index  │ Very small file. Stores      │
  │ (PostgreSQL)     │ (min/max per block │ min/max summaries per range  │
  │                  │  range)            │ of table pages.              │
  │                  │                    │                              │
  │ LSM Tree         │ Log-structured     │ Multiple sorted files        │
  │ (RocksDB,        │ merge tree         │ (SSTables) at different      │
  │  Cassandra,      │                    │ levels. Compacted            │
  │  LevelDB)        │                    │ periodically.                │
  └──────────────────┴────────────────────┴──────────────────────────────┘
```

### MySQL InnoDB: Clustered Index (Special Case)

In InnoDB, the table data IS the index. The primary key index is a B+ Tree, and the leaf nodes contain the actual row data (not just pointers):

```
  InnoDB Clustered Index (Primary Key = id)
  ══════════════════════════════════════════

  ┌─────────────────────────────────────────────────────────┐
  │                    Root Page                             │
  │              [id=50]                                     │
  │             /        \                                   │
  │         Page 1      Page 2                               │
  └─────────────────────────────────────────────────────────┘

  ┌─────────────────────────────┐  ┌────────────────────────────────┐
  │  Leaf Page 1                │  │  Leaf Page 2                   │
  │  ┌────────────────────────┐ │  │  ┌──────────────────────────┐  │
  │  │ id=1 │ Alice │ $1000  │ │  │  │ id=50 │ Jane │ $3000    │  │
  │  │ id=2 │ Bob   │ $500   │ │  │  │ id=51 │ Mike │ $200     │  │
  │  │ id=3 │ Carol │ $700   │ │  │  │ id=52 │ Zara │ $1500    │  │
  │  └────────────────────────┘ │  │  └──────────────────────────┘  │
  │  next→Leaf Page 2           │  │  next→null                     │
  └─────────────────────────────┘  └────────────────────────────────┘

  Notice: The leaf pages contain the FULL ROW DATA, not just pointers.
  The table IS the B+ Tree. There's no separate heap file.

  Secondary indexes (e.g., on email) store:
    key = email value
    value = primary key (id), NOT a row pointer

  So a secondary index lookup does:
    1. Search secondary B+ Tree → find primary key
    2. Search clustered (primary) B+ Tree → find actual row
    (This is called a "bookmark lookup" or "index lookup")
```

### PostgreSQL: Heap + Separate Index Files

```
  PostgreSQL stores table data and indexes separately:

  ┌──────────────────────────────────────────────────────────┐
  │                                                          │
  │  Table file (heap): base/16384/16385                     │
  │  ┌──────────┬──────────┬──────────┐                      │
  │  │  Page 0  │  Page 1  │  Page 2  │  ← rows stored here │
  │  └──────────┴──────────┴──────────┘                      │
  │                                                          │
  │  Index file (B+Tree): base/16384/16387                   │
  │  ┌──────────┬──────────┬──────────┐                      │
  │  │  Page 0  │  Page 1  │  Page 2  │  ← keys + TIDs      │
  │  └──────────┴──────────┴──────────┘                      │
  │                                                          │
  │  TID = (page number, item offset) in the heap file       │
  │  e.g., (1, 3) means "page 1, 3rd item" in the heap      │
  │                                                          │
  │  Index lookup:                                           │
  │  1. Search B+ Tree → find TID (1, 3)                     │
  │  2. Go to heap page 1, item 3 → read the full row        │
  │                                                          │
  └──────────────────────────────────────────────────────────┘
```

### Index Durability During Crash Recovery

```
  ┌──────────────────────────────────────────────────────────────┐
  │  SCENARIO: Crash during an INSERT that modifies both         │
  │  the heap page and the index page.                           │
  │                                                              │
  │  WAL contains:                                               │
  │  ┌────────────────────────────────────────────────────┐      │
  │  │  Record 1: TX-201 BEGIN                            │      │
  │  │  Record 2: TX-201 → heap page 12: insert row       │      │
  │  │  Record 3: TX-201 → index page 5: insert key       │      │
  │  │  Record 4: TX-201 COMMIT                           │      │
  │  └────────────────────────────────────────────────────┘      │
  │                                                              │
  │  Case A: Crash after Record 4 (committed)                    │
  │  → Recovery REDOs both heap and index changes                │
  │  → Table and index are consistent ✅                         │
  │                                                              │
  │  Case B: Crash after Record 2 but before Record 4            │
  │  → No COMMIT record found                                   │
  │  → Recovery UNDOs the heap change                            │
  │  → If index change was partially applied, it's undone too    │
  │  → Table and index are consistent ✅                         │
  │                                                              │
  │  Case C: Crash after Record 3 but before Record 4            │
  │  → Same as Case B — both heap and index changes undone       │
  │  → Table and index are consistent ✅                         │
  │                                                              │
  │  The WAL treats table and index modifications as part of     │
  │  the SAME transaction. They're atomic together.              │
  └──────────────────────────────────────────────────────────────┘
```

---

## How Does the Database Provide Such Good Durability? — Summary

```
  ┌──────────────────────────────────────────────────────────────────┐
  │                                                                  │
  │  DURABILITY = WAL + fsync + Recovery + Redundancy                │
  │                                                                  │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  1. WRITE-AHEAD LOGGING                                   │  │
  │  │     Every change is logged before it's applied.            │  │
  │  │     The log is the source of truth, not the data files.    │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  2. fsync ON COMMIT                                       │  │
  │  │     The WAL is forced to physical disk before "OK" is      │  │
  │  │     returned. No lying about persistence.                  │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  3. SEQUENTIAL I/O FOR WAL                                │  │
  │  │     WAL writes are append-only (sequential), making them   │  │
  │  │     fast even with fsync. Data file writes are deferred.   │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  4. CRASH RECOVERY (ARIES)                                │  │
  │  │     On restart: Analysis → Redo → Undo.                    │  │
  │  │     Committed data is replayed. Uncommitted is rolled back.│  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  5. TORN PAGE PROTECTION                                  │  │
  │  │     Full-page writes (PG) or doublewrite buffer (InnoDB)   │  │
  │  │     protect against partial writes during crashes.         │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  6. CHECKSUMS                                             │  │
  │  │     Every page has a checksum to detect corruption.        │  │
  │  │     Bit rot, bad sectors, firmware bugs — all caught.      │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  ┌────────────────────────────────────────────────────────────┐  │
  │  │  7. REPLICATION                                           │  │
  │  │     Synchronous replicas ensure data survives even total   │  │
  │  │     hardware failure of the primary.                       │  │
  │  └────────────────────────────────────────────────────────────┘  │
  │                                                                  │
  │  Together, these layers make it extraordinarily difficult to     │
  │  lose committed data. You'd need simultaneous failure of         │
  │  primary disk + replica disk + all backups. At that point,       │
  │  you have bigger problems than database durability.              │
  │                                                                  │
  └──────────────────────────────────────────────────────────────────┘
```

---

## LSM Trees — A Different Durability Approach (Cassandra, RocksDB, LevelDB)

Not all databases use B+ Trees. LSM Tree-based databases (Cassandra, RocksDB, LevelDB, ScyllaDB) take a different approach to durability:

```
  LSM Tree Write Path:
  ════════════════════

  ┌──────────┐    ①    ┌──────────────┐    ②    ┌──────────────┐
  │  Write   │──────>│  WAL / CommitLog│──────>│  MemTable    │
  │ (INSERT) │ write │  (on disk,     │ then  │  (in-memory  │
  └──────────┘  log  │   sequential)  │ write │   sorted     │
                first └──────────────┘  to    │   structure) │
                                        RAM   └──────┬───────┘
                                                     │
                                          when MemTable is full
                                                     │
                                                     ▼
                                              ┌──────────────┐
                                              │  SSTable     │
                                              │  (Sorted     │
                                              │   String     │
                                              │   Table)     │
                                              │  on disk     │
                                              └──────┬───────┘
                                                     │
                                              compaction merges
                                              SSTables over time
                                                     │
                                                     ▼
                                              ┌──────────────┐
                                              │  Level 0     │
                                              │  Level 1     │
                                              │  Level 2     │
                                              │  ...         │
                                              └──────────────┘

  Durability in LSM:
  - Step ① (WAL/CommitLog) provides durability — same principle as B+Tree DBs
  - MemTable is in RAM — volatile, but WAL can rebuild it on crash
  - SSTables on disk are immutable — once written, never modified
  - Compaction merges SSTables but never modifies existing ones
    (writes new files, then deletes old ones atomically)
```

```
  LSM vs B+Tree Durability Comparison:
  ┌────────────────────┬──────────────────────┬──────────────────────┐
  │ Aspect             │ B+Tree (PG, MySQL)   │ LSM (Cassandra, etc) │
  ├────────────────────┼──────────────────────┼──────────────────────┤
  │ WAL/CommitLog      │ Yes (WAL)            │ Yes (CommitLog)      │
  │ In-memory buffer   │ Buffer Pool (pages)  │ MemTable (sorted)    │
  │ On-disk data       │ Heap + B+Tree files  │ SSTables (immutable) │
  │ Write pattern      │ Random (update in    │ Sequential (append   │
  │                    │ place)               │ new, merge later)    │
  │ Torn page risk     │ Yes (needs FPW or    │ No (SSTables are     │
  │                    │ doublewrite)         │ immutable)           │
  │ Recovery source    │ WAL → redo/undo      │ CommitLog → rebuild  │
  │                    │                      │ MemTable             │
  └────────────────────┴──────────────────────┴──────────────────────┘
```

---

## The Durability Spectrum — Tunable Tradeoffs

Most databases let you tune durability vs performance:

```
  ┌──────────────────────────────────────────────────────────────────┐
  │                    DURABILITY DIAL                               │
  │                                                                  │
  │  Maximum Durability                    Maximum Performance       │
  │  ◄──────────────────────────────────────────────────────────►    │
  │                                                                  │
  │  fsync every commit     fsync every N commits    no fsync        │
  │  + sync replication     + async replication      + no replication│
  │                                                                  │
  │  PostgreSQL:                                                     │
  │    synchronous_commit = on        (default, full durability)     │
  │    synchronous_commit = off       (WAL written but not fsync'd   │
  │                                    on every commit — risk of     │
  │                                    losing last ~600ms of commits │
  │                                    on crash, but 2-3x faster)   │
  │                                                                  │
  │  MySQL InnoDB:                                                   │
  │    innodb_flush_log_at_trx_commit = 1  (fsync every commit) ✅   │
  │    innodb_flush_log_at_trx_commit = 2  (flush to OS cache only)  │
  │    innodb_flush_log_at_trx_commit = 0  (flush every second)      │
  │                                                                  │
  │  Redis:                                                          │
  │    appendfsync always     (fsync every write — slow but durable) │
  │    appendfsync everysec   (fsync every second — default)         │
  │    appendfsync no         (let OS decide — fast but risky)       │
  │                                                                  │
  └──────────────────────────────────────────────────────────────────┘
```

---

## Key Takeaways

1. Durability = "committed data survives crashes." The WAL + fsync is the mechanism that makes this possible. The database writes to the WAL first (sequential, fast), fsyncs it to physical disk, and only then tells you "committed."

2. Data lives in regular files on disk — heap files for table data, separate files for indexes (PostgreSQL) or a single clustered B+ Tree file (InnoDB). Pages are the unit of I/O (8KB or 16KB).

3. Indexes are stored as files of pages, just like table data. B+ Tree indexes have internal pages (routing) and leaf pages (keys + row pointers). They go through the same WAL pipeline and are recovered atomically with table data.

4. On crash, the WAL is the recovery source. REDO replays committed changes, UNDO rolls back uncommitted ones. Both table and index pages are recovered together.

5. Disk failure is the one scenario where single-node durability breaks. RAID, replication, and backups are the defense layers. Synchronous replication gives zero data loss even on total primary failure.

6. Torn pages (partial writes) are handled by full-page writes (PostgreSQL) or doublewrite buffers (InnoDB). The WAL or doublewrite area provides a clean copy to restore from.

7. The durability guarantee is tunable. You can trade some durability for performance (async commits, relaxed fsync). Know what you're giving up.

The fundamental insight: the WAL is the single source of truth for durability. Data files are just a materialized cache of what the WAL describes. If the data files are lost or corrupted, the WAL can rebuild them. That's why "write-ahead" logging is the backbone of every serious database engine.