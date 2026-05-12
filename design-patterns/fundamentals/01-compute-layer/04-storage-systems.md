# Storage Systems

## Overview

Storage systems provide persistent data storage. Understanding storage characteristics is crucial for designing performant, reliable, and cost-effective systems.

## Storage Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                    Storage Hierarchy                         │
│                                                              │
│  Speed ▲                                    Cost per GB ▼   │
│        │                                                     │
│        │  ┌──────────────────────────┐                      │
│        │  │   RAM (Volatile)         │                      │
│        │  │   100 ns latency         │                      │
│        │  │   $5-10 per GB           │                      │
│        │  └──────────────────────────┘                      │
│        │  ┌──────────────────────────┐                      │
│        │  │   NVMe SSD               │                      │
│        │  │   10-100 μs latency      │                      │
│        │  │   $0.10-0.30 per GB      │                      │
│        │  └──────────────────────────┘                      │
│        │  ┌──────────────────────────┐                      │
│        │  │   SATA SSD               │                      │
│        │  │   100-500 μs latency     │                      │
│        │  │   $0.08-0.15 per GB      │                      │
│        │  └──────────────────────────┘                      │
│        │  ┌──────────────────────────┐                      │
│        │  │   HDD (7200 RPM)         │                      │
│        │  │   5-10 ms latency        │                      │
│        │  │   $0.02-0.05 per GB      │                      │
│        │  └──────────────────────────┘                      │
│        │  ┌──────────────────────────┐                      │
│        │  │   Tape/Archive           │                      │
│        │  │   Seconds latency        │                      │
│        │  │   $0.002-0.01 per GB     │                      │
│        │  └──────────────────────────┘                      │
│        ▼                                                     │
│    Capacity                                                  │
└─────────────────────────────────────────────────────────────┘
```

## Storage Technologies

### 1. Hard Disk Drives (HDD)

Magnetic storage with rotating platters.

```
HDD Architecture:

┌─────────────────────────────────────────┐
│         Hard Disk Drive                  │
│                                          │
│     ┌─────────────────────┐             │
│     │   Read/Write Head   │             │
│     └──────────┬──────────┘             │
│                │                         │
│                ▼                         │
│        ┌───────────────┐                │
│        │               │                │
│        │   Platter     │ ← Rotating     │
│        │   (Magnetic)  │   5400-15000   │
│        │               │   RPM          │
│        └───────────────┘                │
│                                          │
│     ┌──────────────────────┐            │
│     │  Actuator Arm        │            │
│     └──────────────────────┘            │
│                                          │
│     ┌──────────────────────┐            │
│     │  Controller          │            │
│     └──────────────────────┘            │
└─────────────────────────────────────────┘

Access Time Components:
┌─────────────────────────────────────────┐
│ Seek Time:    4-10 ms (move head)       │
│ Rotational:   2-4 ms (wait for data)    │
│ Transfer:     0.1-1 ms (read data)      │
│ ─────────────────────────────────────   │
│ Total:        6-15 ms average           │
└─────────────────────────────────────────┘
```

**Characteristics:**
- **Capacity**: 1-20 TB per drive
- **Speed**: 100-200 MB/s sequential
- **IOPS**: 80-160 random IOPS
- **Latency**: 5-15 ms
- **Lifespan**: 3-5 years
- **Cost**: $0.02-0.05 per GB

**Best For:**
- Archival storage
- Cold data
- Sequential access workloads
- Cost-sensitive applications

### 2. Solid State Drives (SSD)

Flash-based storage with no moving parts.

```
SSD Architecture:

┌─────────────────────────────────────────┐
│         Solid State Drive                │
│                                          │
│  ┌────────────────────────────────┐     │
│  │      Controller                │     │
│  │  • Wear Leveling               │     │
│  │  • Garbage Collection          │     │
│  │  • Error Correction            │     │
│  └──────────┬─────────────────────┘     │
│             │                            │
│             ▼                            │
│  ┌──────────────────────────────┐       │
│  │      DRAM Cache              │       │
│  │      (Optional)              │       │
│  └──────────────────────────────┘       │
│             │                            │
│             ▼                            │
│  ┌────┬────┬────┬────┬────┬────┐       │
│  │Die │Die │Die │Die │Die │Die │       │
│  │ 0  │ 1  │ 2  │ 3  │ 4  │ 5  │       │
│  └────┴────┴────┴────┴────┴────┘       │
│                                          │
│  Each Die contains:                     │
│  ┌──────────────────────────────┐       │
│  │ Blocks (128-256 pages)       │       │
│  │  └─ Pages (4-16 KB)          │       │
│  └──────────────────────────────┘       │
└─────────────────────────────────────────┘

Flash Memory Operations:
┌─────────────────────────────────────────┐
│ Read:   Page level (4-16 KB)            │
│         Fast: 25-100 μs                 │
│                                          │
│ Write:  Page level (4-16 KB)            │
│         Slower: 200-1000 μs             │
│         Must be erased first!           │
│                                          │
│ Erase:  Block level (128-256 pages)     │
│         Very slow: 1-3 ms               │
└─────────────────────────────────────────┘
```

**SSD Types:**

#### SATA SSD
- **Interface**: SATA III (6 Gbps)
- **Speed**: 500-600 MB/s
- **IOPS**: 90K-100K
- **Latency**: 100-500 μs
- **Use**: General purpose, legacy systems

#### NVMe SSD
- **Interface**: PCIe (32 Gbps+)
- **Speed**: 3,000-7,000 MB/s
- **IOPS**: 500K-1M+
- **Latency**: 10-100 μs
- **Use**: High-performance databases, caching

```
SATA vs NVMe Performance:

Sequential Read:
SATA SSD:  ████████ 550 MB/s
NVMe SSD:  ████████████████████████████ 3,500 MB/s

Random Read IOPS:
SATA SSD:  ████████ 95K IOPS
NVMe SSD:  ████████████████████████████ 600K IOPS

Latency:
SATA SSD:  ████████ 100 μs
NVMe SSD:  ██ 20 μs
```

**SSD Challenges:**

1. **Write Amplification**
```
Write Amplification:

User writes 4 KB:
┌────┐
│ 4KB│
└────┘
    │
    ▼
SSD must:
1. Read entire block (256 KB)
2. Modify 4 KB
3. Erase block
4. Write entire block (256 KB)

Write Amplification Factor = 256 KB / 4 KB = 64x

Impact:
• Reduced performance
• Reduced lifespan
• Increased power consumption
```

2. **Wear Leveling**
```
Wear Leveling:

Without Wear Leveling:
┌────┬────┬────┬────┬────┐
│ A  │ B  │ C  │ D  │ E  │
└────┴────┴────┴────┴────┘
  ▲
  └─ Block A written 10,000 times (worn out)
     Other blocks barely used

With Wear Leveling:
┌────┬────┬────┬────┬────┐
│ A  │ B  │ C  │ D  │ E  │
└────┴────┴────┴────┴────┘
  ▲    ▲    ▲    ▲    ▲
  └────┴────┴────┴────┘
  All blocks written ~2,000 times (even wear)
```

### 3. NVMe (Non-Volatile Memory Express)

Modern protocol designed specifically for SSDs.

```
NVMe vs SATA Protocol:

SATA (Legacy):
┌──────────┐
│   CPU    │
└────┬─────┘
     │
     ▼
┌──────────┐
│   AHCI   │ ← Designed for HDDs
└────┬─────┘    (Single queue, 32 commands)
     │
     ▼
┌──────────┐
│   SATA   │
│   SSD    │
└──────────┘

NVMe (Modern):
┌──────────┐
│   CPU    │
└────┬─────┘
     │
     ▼
┌──────────┐
│   NVMe   │ ← Designed for SSDs
└────┬─────┘    (64K queues, 64K commands each)
     │
     ▼
┌──────────┐
│   PCIe   │
│   NVMe   │
│   SSD    │
└──────────┘

Benefits:
• Lower latency (fewer layers)
• Higher parallelism (more queues)
• Better CPU efficiency
```

## Storage Performance Metrics

### Key Metrics

```
Storage Performance Metrics:

1. Throughput (Bandwidth):
   ┌─────────────────────────────────┐
   │ Sequential Read:  3,500 MB/s    │
   │ Sequential Write: 3,000 MB/s    │
   └─────────────────────────────────┘
   
2. IOPS (Input/Output Operations Per Second):
   ┌─────────────────────────────────┐
   │ Random Read:  600,000 IOPS      │
   │ Random Write: 500,000 IOPS      │
   └─────────────────────────────────┘
   
3. Latency:
   ┌─────────────────────────────────┐
   │ Average: 20 μs                  │
   │ 99th %:  100 μs                 │
   │ 99.9th%: 500 μs                 │
   └─────────────────────────────────┘
   
4. Queue Depth:
   ┌─────────────────────────────────┐
   │ Number of outstanding I/O       │
   │ operations                      │
   │ Higher QD = Better throughput   │
   └─────────────────────────────────┘
```

### Sequential vs Random Access

```
Access Patterns:

Sequential Access:
┌───┬───┬───┬───┬───┬───┬───┬───┐
│ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │ 6 │ 7 │
└───┴───┴───┴───┴───┴───┴───┴───┘
  ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼
  Read blocks in order

Performance:
• HDD:  Good (100-200 MB/s)
• SSD:  Excellent (3,000-7,000 MB/s)

Random Access:
┌───┬───┬───┬───┬───┬───┬───┬───┐
│ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │ 6 │ 7 │
└───┴───┴───┴───┴───┴───┴───┴───┘
  ▼       ▼   ▼           ▼
  Read blocks in random order

Performance:
• HDD:  Poor (80-160 IOPS)
• SSD:  Excellent (500K-1M IOPS)

Ratio (Sequential / Random):
• HDD:  ~1,000x difference
• SSD:  ~5-10x difference
```

## RAID (Redundant Array of Independent Disks)

Combine multiple disks for performance and/or redundancy.

```
RAID Levels:

RAID 0 (Striping):
┌────────┬────────┐
│ Disk 0 │ Disk 1 │
├────────┼────────┤
│ Block0 │ Block1 │
│ Block2 │ Block3 │
│ Block4 │ Block5 │
└────────┴────────┘

• Capacity: 100% (2x)
• Performance: 2x read, 2x write
• Redundancy: None (any disk failure = data loss)
• Use: Temporary data, caching

RAID 1 (Mirroring):
┌────────┬────────┐
│ Disk 0 │ Disk 1 │
├────────┼────────┤
│ Block0 │ Block0 │ ← Same data
│ Block1 │ Block1 │
│ Block2 │ Block2 │
└────────┴────────┘

• Capacity: 50% (1x)
• Performance: 2x read, 1x write
• Redundancy: 1 disk can fail
• Use: Critical data, databases

RAID 5 (Striping + Parity):
┌────────┬────────┬────────┐
│ Disk 0 │ Disk 1 │ Disk 2 │
├────────┼────────┼────────┤
│ Block0 │ Block1 │ Parity │
│ Block2 │ Parity │ Block3 │
│ Parity │ Block4 │ Block5 │
└────────┴────────┴────────┘

• Capacity: (N-1)/N (67% for 3 disks)
• Performance: Good read, slower write
• Redundancy: 1 disk can fail
• Use: General purpose storage

RAID 6 (Striping + Double Parity):
┌────────┬────────┬────────┬────────┐
│ Disk 0 │ Disk 1 │ Disk 2 │ Disk 3 │
├────────┼────────┼────────┼────────┤
│ Block0 │ Block1 │ Parity │ Parity │
│ Block2 │ Parity │ Parity │ Block3 │
└────────┴────────┴────────┴────────┘

• Capacity: (N-2)/N (50% for 4 disks)
• Performance: Good read, slow write
• Redundancy: 2 disks can fail
• Use: Critical data, large arrays

RAID 10 (1+0, Mirrored Stripes):
┌────────┬────────┐  ┌────────┬────────┐
│ Disk 0 │ Disk 1 │  │ Disk 2 │ Disk 3 │
├────────┼────────┤  ├────────┼────────┤
│ Block0 │ Block0 │  │ Block1 │ Block1 │
│ Block2 │ Block2 │  │ Block3 │ Block3 │
└────────┴────────┘  └────────┴────────┘
      Mirror              Mirror
         └──────Stripe──────┘

• Capacity: 50%
• Performance: Excellent read/write
• Redundancy: 1 disk per mirror can fail
• Use: High-performance databases
```

## File Systems

### Common File Systems

```
File System Comparison:

┌──────────┬─────────┬──────────┬─────────────┐
│ FS       │ Max File│ Max Vol  │ Features    │
├──────────┼─────────┼──────────┼─────────────┤
│ ext4     │ 16 TB   │ 1 EB     │ Journaling  │
│          │         │          │ Stable      │
├──────────┼─────────┼──────────┼─────────────┤
│ XFS      │ 8 EB    │ 8 EB     │ High perf   │
│          │         │          │ Large files │
├──────────┼─────────┼──────────┼─────────────┤
│ Btrfs    │ 16 EB   │ 16 EB    │ Snapshots   │
│          │         │          │ Compression │
├──────────┼─────────┼──────────┼─────────────┤
│ ZFS      │ 16 EB   │ 256 ZB   │ Checksums   │
│          │         │          │ Snapshots   │
│          │         │          │ Compression │
├──────────┼─────────┼──────────┼─────────────┤
│ NTFS     │ 16 EB   │ 16 EB    │ Windows     │
│          │         │          │ ACLs        │
└──────────┴─────────┴──────────┴─────────────┘
```

### File System Structure

```
File System Layout:

┌─────────────────────────────────────────┐
│         Disk Partition                   │
├─────────────────────────────────────────┤
│ Boot Block                              │
├─────────────────────────────────────────┤
│ Superblock                              │
│ • File system metadata                  │
│ • Block size, total blocks              │
│ • Free block count                      │
├─────────────────────────────────────────┤
│ Inode Table                             │
│ • File metadata                         │
│ • Permissions, timestamps               │
│ • Block pointers                        │
├─────────────────────────────────────────┤
│ Data Blocks                             │
│ • Actual file content                   │
│ • Directory entries                     │
└─────────────────────────────────────────┘

Inode Structure:
┌─────────────────────────────────────────┐
│ Inode #12345                            │
├─────────────────────────────────────────┤
│ Mode: -rw-r--r--                        │
│ Owner: user:group                       │
│ Size: 1,048,576 bytes                   │
│ Created: 2026-01-15 10:30:00            │
│ Modified: 2026-05-11 14:20:00           │
├─────────────────────────────────────────┤
│ Direct Pointers (12):                   │
│ [Block 1000] [Block 1001] ...           │
├─────────────────────────────────────────┤
│ Indirect Pointer:                       │
│ → [Block 2000] → [Block 3001] ...       │
├─────────────────────────────────────────┤
│ Double Indirect Pointer:                │
│ → [Block 3000] → [Block 4000] → ...     │
└─────────────────────────────────────────┘
```

## Storage Optimization Techniques

### 1. Caching

```
Storage Caching Layers:

┌─────────────────────────────────────────┐
│         Application                      │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│    Application Cache (Redis, Memcached) │
│    Hit Rate: 90-95%                     │
└────────────────┬────────────────────────┘
                 │ Miss
                 ▼
┌─────────────────────────────────────────┐
│    OS Page Cache (RAM)                  │
│    Hit Rate: 80-90%                     │
└────────────────┬────────────────────────┘
                 │ Miss
                 ▼
┌─────────────────────────────────────────┐
│    SSD Cache (if present)               │
│    Hit Rate: 70-80%                     │
└────────────────┬────────────────────────┘
                 │ Miss
                 ▼
┌─────────────────────────────────────────┐
│    Primary Storage (HDD/SSD)            │
└─────────────────────────────────────────┘

Effective Hit Rate:
• L1 (App): 90% → 10% miss
• L2 (OS):  80% of 10% = 8% → 2% miss
• L3 (SSD): 70% of 2% = 1.4% → 0.6% miss
• Disk: 0.6% of requests

99.4% of requests served from cache!
```

### 2. Read-Ahead (Prefetching)

```
Read-Ahead Strategy:

Sequential Access Detected:
┌───┬───┬───┬───┬───┬───┬───┬───┐
│ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │ 6 │ 7 │
└───┴───┴───┴───┴───┴───┴───┴───┘
  ▲   ▲
  │   └─ Read block 1
  └───── Read block 0

OS predicts: blocks 2, 3, 4 will be needed
┌───┬───┬───┬───┬───┬───┬───┬───┐
│ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │ 6 │ 7 │
└───┴───┴───┴───┴───┴───┴───┴───┘
          ▲   ▲   ▲
          └───┴───┘
      Prefetch into cache

When app requests block 2:
• Already in cache!
• No disk I/O needed
```

### 3. Write-Back Caching

```
Write-Back vs Write-Through:

Write-Through:
App → Cache → Disk (wait) → Acknowledge
Latency: ~10 ms (disk latency)

Write-Back:
App → Cache → Acknowledge (immediate)
              ↓
         Disk (async)
Latency: ~100 ns (cache latency)

┌─────────────────────────────────────────┐
│         Write-Back Cache                 │
│                                          │
│  ┌────────────────────────────────┐     │
│  │ Dirty Pages (not yet on disk)  │     │
│  │ ┌────┬────┬────┬────┐          │     │
│  │ │ A  │ B  │ C  │ D  │          │     │
│  │ └────┴────┴────┴────┘          │     │
│  └────────────────────────────────┘     │
│              │                           │
│              ▼ (Periodic flush)          │
│  ┌────────────────────────────────┐     │
│  │         Disk                    │     │
│  └────────────────────────────────┘     │
└─────────────────────────────────────────┘

Risk: Data loss if power failure before flush
Mitigation: Battery-backed cache, UPS
```

### 4. I/O Scheduling

```
I/O Schedulers:

1. NOOP (No Operation):
   ┌───┬───┬───┬───┐
   │ A │ B │ C │ D │ → Process in order
   └───┴───┴───┴───┘
   Use: SSDs (no seek time)

2. Deadline:
   ┌───────────────────────────────┐
   │ Read Queue (deadline: 500ms)  │
   │ Write Queue (deadline: 5s)    │
   └───────────────────────────────┘
   Use: Real-time systems

3. CFQ (Completely Fair Queuing):
   ┌─────────┬─────────┬─────────┐
   │Process A│Process B│Process C│
   │  Queue  │  Queue  │  Queue  │
   └─────────┴─────────┴─────────┘
   Round-robin between processes
   Use: General purpose (HDD)

4. BFQ (Budget Fair Queuing):
   Similar to CFQ but better for SSDs
   Use: Modern Linux default
```

## Cloud Storage Types

```
Cloud Storage Tiers:

┌─────────────────────────────────────────┐
│ Hot/Premium Storage                     │
│ • SSD-backed                            │
│ • Low latency (<10ms)                   │
│ • High IOPS (>10K)                      │
│ • Cost: $$$                             │
│ Use: Databases, active data             │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Standard Storage                        │
│ • HDD or SSD                            │
│ • Medium latency (10-50ms)              │
│ • Medium IOPS (500-5K)                  │
│ • Cost: $$                              │
│ Use: File shares, general purpose       │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Cool/Infrequent Access                  │
│ • HDD-backed                            │
│ • Higher latency (50-100ms)             │
│ • Lower IOPS (<500)                     │
│ • Cost: $                               │
│ Use: Backups, archives (30+ days)       │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Archive/Glacier                         │
│ • Tape or cold storage                  │
│ • Very high latency (hours)             │
│ • Retrieval fees                        │
│ • Cost: ¢                               │
│ Use: Long-term archives (90+ days)      │
└─────────────────────────────────────────┘
```

## Storage Monitoring

### Key Metrics

```bash
# Disk I/O statistics
iostat -x 1

# Per-process I/O
iotop

# Disk usage
df -h
du -sh /path/*

# Inode usage
df -i

# SMART health
smartctl -a /dev/sda

# Block device info
lsblk
blkid
```

### Performance Metrics

```
Storage Performance Dashboard:

┌─────────────────────────────────────────┐
│ Disk Utilization: 65%                   │
│ ████████████████████░░░░░░░░░░░         │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ IOPS: 15,234                            │
│ Read:  10,123 (66%)                     │
│ Write:  5,111 (34%)                     │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Throughput: 450 MB/s                    │
│ Read:  300 MB/s                         │
│ Write: 150 MB/s                         │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Average Latency: 2.5 ms                 │
│ Read:  1.8 ms                           │
│ Write: 4.2 ms                           │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Queue Depth: 8                          │
└─────────────────────────────────────────┘
```

## Summary

Key takeaways for architects:

1. **Match storage to workload**
   - Sequential: HDD acceptable
   - Random: SSD required
   - High IOPS: NVMe essential

2. **Understand the tradeoffs**
   - Performance vs Cost
   - Capacity vs Speed
   - Durability vs Latency

3. **Leverage caching**
   - Multiple cache layers
   - Dramatically improves performance
   - Monitor hit rates

4. **Plan for redundancy**
   - RAID for local redundancy
   - Replication for geographic redundancy
   - Backups for disaster recovery

5. **Monitor storage health**
   - IOPS, throughput, latency
   - Disk utilization
   - SMART metrics

## Next Steps

Continue to [Operating Systems](./05-operating-systems.md) to understand OS concepts and their impact on system design.
