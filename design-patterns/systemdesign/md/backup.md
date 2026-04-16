# Database Backup Strategies

## Cold Backup (Offline)

- Take the DB offline, copy the data files
- Simple and consistent, but requires downtime
- Fine for dev/staging, not ideal for production

## Hot Backup Techniques (No Downtime)

### 1. Logical Backups

Tools like `pg_dump` (Postgres) or `mysqldump` (MySQL) read data through the DB engine while it's running.

- They acquire a consistent snapshot using MVCC or a brief global read lock
- Output is SQL statements or a portable format
- Good for smaller databases, but slow for large ones since they read every row
- Can increase load on the running DB

### 2. WAL / Log-Based Continuous Archiving

The most common production approach.

- The DB already writes every change to a Write-Ahead Log (WAL in Postgres, binlog in MySQL, redo log in Oracle)
- Take one base backup (a file-level copy while the DB runs), then continuously archive the WAL segments
- To restore: replay the base backup + WAL segments up to any point in time
- This gives **Point-in-Time Recovery (PITR)**
- Postgres does this with `pg_basebackup` + WAL archiving
- MySQL uses binary log replication for the same idea

### 3. Snapshot-Based (Filesystem / Storage Level)

- Use LVM snapshots, ZFS snapshots, or cloud volume snapshots (e.g., EBS snapshots)
- These are copy-on-write — the snapshot is near-instant, no downtime
- The DB keeps running; the snapshot captures a crash-consistent state
- Some DBs need a brief `FLUSH` or `fsync` to ensure the snapshot is clean, but it's milliseconds, not real downtime
- Cloud-native DBs lean heavily on this approach

### 4. Replica-Based Backup

- Set up a read replica (streaming replication, log shipping, etc.)
- Take the backup from the replica instead of the primary
- Zero impact on the primary — the replica can be paused or snapshotted freely
- Very common in production setups

### 5. Incremental / Differential Backups

Tools like `pgBackRest`, `Percona XtraBackup` (MySQL), or Oracle RMAN:

- First take a full base backup
- Then only back up changed pages/blocks since the last backup
- Much faster for large DBs
- XtraBackup copies InnoDB data files while the DB runs, using the redo log to make the copy consistent — no locking needed


## How Consistency Is Maintained Without Downtime

The core trick across all these techniques is the same idea:

```
snapshot/copy of data files  +  replay of changes that happened during the copy
                                (from WAL / redo log / binlog)
= consistent backup
```

MVCC helps too — the backup process can read a consistent view of the data at a specific point in time without blocking writers.

## Comparison

| Technique | Downtime | Speed | PITR | DB Load |
|---|---|---|---|---|
| Cold backup | Yes | Fast (file copy) | No | None |
| Logical (pg_dump) | No | Slow for large DBs | No | Medium |
| WAL/binlog archiving | No | Fast | Yes | Low |
| Storage snapshots | No | Near-instant | With WAL, yes | Very low |
| Replica-based | No | Depends | Yes | None on primary |
| Incremental (XtraBackup) | No | Fast | Yes | Low-Medium |

## Production Recommendation

For most production systems, the go-to is **WAL-based continuous archiving** combined with **periodic base backups**, often taken from a **replica**. This gives:

- Zero downtime
- Low impact on the primary
- Ability to restore to any point in time (PITR)
