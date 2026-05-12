# Memory Management

## Overview

Memory management is one of the most critical aspects of system performance and reliability. Understanding how memory works enables architects to design efficient, scalable systems.

## Memory Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                    Memory Hierarchy                          │
│                                                              │
│  Speed ▲                                    Cost per GB ▲   │
│        │                                                 │   │
│        │  ┌──────────────────────────┐                  │   │
│        │  │   CPU Registers          │                  │   │
│        │  │   ~100 bytes             │                  │   │
│        │  │   < 1 ns                 │                  │   │
│        │  └──────────────────────────┘                  │   │
│        │  ┌──────────────────────────┐                  │   │
│        │  │   L1 Cache               │                  │   │
│        │  │   32-64 KB per core      │                  │   │
│        │  │   ~1 ns                  │                  │   │
│        │  └──────────────────────────┘                  │   │
│        │  ┌──────────────────────────┐                  │   │
│        │  │   L2 Cache               │                  │   │
│        │  │   256-512 KB per core    │                  │   │
│        │  │   ~3 ns                  │                  │   │
│        │  └──────────────────────────┘                  │   │
│        │  ┌──────────────────────────┐                  │   │
│        │  │   L3 Cache               │                  │   │
│        │  │   8-64 MB (shared)       │                  │   │
│        │  │   ~10 ns                 │                  │   │
│        │  └──────────────────────────┘                  │   │
│        │  ┌──────────────────────────┐                  │   │
│        │  │   Main Memory (RAM)      │                  │   │
│        │  │   8-512 GB               │                  │   │
│        │  │   ~100 ns                │                  │   │
│        │  └──────────────────────────┘                  │   │
│        │  ┌──────────────────────────┐                  │   │
│        │  │   SSD Storage            │                  │   │
│        │  │   256 GB - 8 TB          │                  │   │
│        │  │   ~100 μs                │                  │   │
│        │  └──────────────────────────┘                  │   │
│        │  ┌──────────────────────────┐                  │   │
│        │  │   HDD Storage            │                  │   │
│        │  │   1-20 TB                │                  │   │
│        │  │   ~10 ms                 │                  │   │
│        │  └──────────────────────────┘                  │   │
│        ▼                                                 ▼   │
│    Capacity                                                  │
└─────────────────────────────────────────────────────────────┘
```

## Virtual Memory

### Concept

Virtual memory provides each process with its own isolated address space, abstracting physical memory.

```
Virtual Memory Architecture:

┌─────────────────────────────────────────────────────────┐
│                    Process View                          │
│                                                          │
│  Process A              Process B              Process C │
│  ┌──────────┐          ┌──────────┐          ┌──────────┐
│  │ Virtual  │          │ Virtual  │          │ Virtual  │
│  │ Address  │          │ Address  │          │ Address  │
│  │ Space    │          │ Space    │          │ Space    │
│  │ 0x0000   │          │ 0x0000   │          │ 0x0000   │
│  │   to     │          │   to     │          │   to     │
│  │ 0xFFFF   │          │ 0xFFFF   │          │ 0xFFFF   │
│  └────┬─────┘          └────┬─────┘          └────┬─────┘
│       │                     │                     │      │
│       └─────────────────────┼─────────────────────┘      │
│                             ▼                            │
│                    ┌─────────────────┐                   │
│                    │  MMU (Memory    │                   │
│                    │  Management     │                   │
│                    │  Unit)          │                   │
│                    └────────┬────────┘                   │
│                             │                            │
│                             ▼                            │
│                    ┌─────────────────┐                   │
│                    │  Page Table     │                   │
│                    │  Translation    │                   │
│                    └────────┬────────┘                   │
│                             │                            │
└─────────────────────────────┼────────────────────────────┘
                              ▼
                    ┌─────────────────┐
                    │  Physical RAM   │
                    │                 │
                    │  ┌───┬───┬───┐  │
                    │  │ A │ B │ C │  │
                    │  └───┴───┴───┘  │
                    └─────────────────┘
```

### Benefits of Virtual Memory

1. **Isolation**: Processes cannot access each other's memory
2. **Simplification**: Each process sees a simple, linear address space
3. **Flexibility**: Physical memory can be non-contiguous
4. **Overcommitment**: Total virtual memory > physical memory
5. **Sharing**: Multiple processes can share memory pages

## Paging

### Page Structure

Memory is divided into fixed-size blocks called pages.

```
Virtual and Physical Memory Pages:

Virtual Memory (Process View):
┌─────────────────────────────────────┐
│ Page 0 (4 KB)  │ 0x0000 - 0x0FFF   │
├─────────────────────────────────────┤
│ Page 1 (4 KB)  │ 0x1000 - 0x1FFF   │
├─────────────────────────────────────┤
│ Page 2 (4 KB)  │ 0x2000 - 0x2FFF   │
├─────────────────────────────────────┤
│ Page 3 (4 KB)  │ 0x3000 - 0x3FFF   │
└─────────────────────────────────────┘
         │
         ▼ (Page Table Translation)
         
Physical Memory (RAM):
┌─────────────────────────────────────┐
│ Frame 5        │ ← Page 0          │
├─────────────────────────────────────┤
│ Frame 2        │ ← Page 3          │
├─────────────────────────────────────┤
│ Frame 7        │ ← Page 1          │
├─────────────────────────────────────┤
│ Frame 1        │ ← Page 2          │
└─────────────────────────────────────┘

Common Page Sizes:
• 4 KB (standard)
• 2 MB (huge pages)
• 1 GB (gigantic pages)
```

### Page Table

Maps virtual addresses to physical addresses.

```
Page Table Structure:

Virtual Address (32-bit example):
┌──────────────────┬──────────────────┐
│  Page Number     │  Page Offset     │
│  (20 bits)       │  (12 bits)       │
└──────────────────┴──────────────────┘
        │                    │
        │                    └─► Offset within page (0-4095)
        │
        ▼
┌─────────────────────────────────────┐
│         Page Table                   │
├──────────┬──────────┬───────────────┤
│ VPN      │ PFN      │ Flags         │
├──────────┼──────────┼───────────────┤
│ 0x00000  │ 0x12345  │ R/W, Present  │
│ 0x00001  │ 0x67890  │ R/W, Present  │
│ 0x00002  │ ------   │ Not Present   │
│ 0x00003  │ 0xABCDE  │ R-only        │
└──────────┴──────────┴───────────────┘
        │
        ▼
Physical Address:
┌──────────────────┬──────────────────┐
│  Frame Number    │  Page Offset     │
│  (from PFN)      │  (same as above) │
└──────────────────┴──────────────────┘

Flags:
• Present: Page is in RAM
• Read/Write: Access permissions
• User/Supervisor: Privilege level
• Dirty: Page has been modified
• Accessed: Page has been read
```

### Multi-Level Page Tables

For large address spaces, use hierarchical page tables.

```
Two-Level Page Table:

Virtual Address (32-bit):
┌──────────┬──────────┬──────────────┐
│ P1 (10b) │ P2 (10b) │ Offset (12b) │
└──────────┴──────────┴──────────────┘
     │         │            │
     │         │            └─► Offset in page
     │         │
     │         └─► Index in 2nd level table
     │
     └─► Index in 1st level table

┌─────────────────────────────────────┐
│   Page Directory (Level 1)          │
│   ┌───┬───┬───┬───┬───┐            │
│   │ 0 │ 1 │ 2 │...│1023            │
│   └─┬─┴───┴───┴───┴───┘            │
│     │                               │
└─────┼───────────────────────────────┘
      │
      ▼
┌─────────────────────────────────────┐
│   Page Table (Level 2)              │
│   ┌───┬───┬───┬───┬───┐            │
│   │ 0 │ 1 │ 2 │...│1023            │
│   └─┬─┴───┴───┴───┴───┘            │
│     │                               │
└─────┼───────────────────────────────┘
      │
      ▼
┌─────────────────────────────────────┐
│   Physical Page Frame               │
└─────────────────────────────────────┘

Benefits:
• Sparse address spaces don't waste memory
• Only allocate page tables as needed
• 64-bit systems use 4-5 levels
```

## Translation Lookaside Buffer (TLB)

Cache for page table entries to speed up address translation.

```
TLB Operation:

┌─────────────────────────────────────────────────────┐
│              Address Translation                     │
│                                                      │
│  Virtual Address                                     │
│       │                                              │
│       ▼                                              │
│  ┌─────────┐                                        │
│  │   TLB   │  ← Fast cache of page table entries   │
│  └────┬────┘                                        │
│       │                                              │
│    Hit│  Miss                                        │
│   ┌───┴────┐                                        │
│   │        │                                        │
│   ▼        ▼                                        │
│  Fast   ┌──────────┐                               │
│  Path   │Page Table│  ← Slow memory access         │
│         │ Walk     │                               │
│         └────┬─────┘                               │
│              │                                      │
│              ▼                                      │
│         Update TLB                                  │
│              │                                      │
│              ▼                                      │
│      Physical Address                              │
└─────────────────────────────────────────────────────┘

TLB Hit:  ~1 ns
TLB Miss: ~100 ns (page table walk)

TLB Hit Rate: Typically 95-99%
```

## Memory Allocation

### Stack vs Heap

```
Process Memory Layout:

High Address (0xFFFFFFFF)
┌─────────────────────────────────────┐
│         Kernel Space                │
│         (Protected)                 │
├─────────────────────────────────────┤
│         Stack                       │
│         (grows downward)            │
│         • Local variables           │
│         • Function calls            │
│         • Return addresses          │
│              ▼                      │
│              │                      │
│              │                      │
│         (Free Space)                │
│              │                      │
│              │                      │
│              ▲                      │
│         Heap                        │
│         (grows upward)              │
│         • Dynamic allocation        │
│         • malloc/new                │
│         • Objects                   │
├─────────────────────────────────────┤
│         BSS Segment                 │
│         (Uninitialized data)        │
├─────────────────────────────────────┤
│         Data Segment                │
│         (Initialized data)          │
├─────────────────────────────────────┤
│         Text Segment                │
│         (Code)                      │
└─────────────────────────────────────┘
Low Address (0x00000000)
```

### Stack Memory

**Characteristics:**
- Fast allocation/deallocation (just move stack pointer)
- Automatic lifetime management
- Limited size (typically 1-8 MB)
- LIFO (Last In, First Out)
- Thread-local

```
Stack Frame Example:

Function Call: foo() calls bar(int x, int y)

┌─────────────────────────────────────┐
│  bar's Stack Frame                  │
├─────────────────────────────────────┤
│  Local variables                    │
│  int result                         │
├─────────────────────────────────────┤
│  Parameters                         │
│  int y                              │
│  int x                              │
├─────────────────────────────────────┤
│  Return address                     │
├─────────────────────────────────────┤
│  Saved frame pointer                │
├═════════════════════════════════════┤ ← Stack Pointer (SP)
│  foo's Stack Frame                  │
│  ...                                │
└─────────────────────────────────────┘

When bar() returns:
• Stack pointer moves back
• Memory automatically reclaimed
• No fragmentation
```

### Heap Memory

**Characteristics:**
- Slower allocation (need to find free block)
- Manual or GC-managed lifetime
- Large size (limited by virtual memory)
- Can fragment
- Shared across threads (with synchronization)

```
Heap Allocation Strategies:

1. First Fit:
┌───┬─────┬───┬─────────┬───┐
│ U │ F   │ U │ F       │ U │
└───┴─────┴───┴─────────┴───┘
     ▲
     └─ Allocate here (first free block)

2. Best Fit:
┌───┬─────┬───┬─────────┬───┐
│ U │ F   │ U │ F       │ U │
└───┴─────┴───┴─────────┴───┘
     ▲
     └─ Allocate here (smallest sufficient block)

3. Worst Fit:
┌───┬─────┬───┬─────────┬───┐
│ U │ F   │ U │ F       │ U │
└───┴─────┴───┴─────────┴───┘
                 ▲
                 └─ Allocate here (largest block)

U = Used, F = Free
```

### Memory Fragmentation

```
External Fragmentation:

Initial State:
┌─────┬─────┬─────┬─────┬─────┐
│  A  │  B  │  C  │  D  │  E  │
└─────┴─────┴─────┴─────┴─────┘

After freeing B and D:
┌─────┬─────┬─────┬─────┬─────┐
│  A  │ Free│  C  │ Free│  E  │
└─────┴─────┴─────┴─────┴─────┘

Problem: Can't allocate 2-block item
even though 2 blocks are free!

Solution: Compaction
┌─────┬─────┬─────┬─────────────┐
│  A  │  C  │  E  │    Free     │
└─────┴─────┴─────┴─────────────┘

Internal Fragmentation:

Request: 10 bytes
Allocated: 16 bytes (due to alignment)
┌──────────────────┐
│ Used (10 bytes)  │
├──────────────────┤
│ Wasted (6 bytes) │ ← Internal fragmentation
└──────────────────┘
```

## Memory Management Techniques

### 1. Garbage Collection

Automatic memory management.

```
Garbage Collection Approaches:

1. Reference Counting:
┌─────────┐
│ Object  │ ← RefCount = 2
└─────────┘
    ▲   ▲
    │   │
   Ptr1 Ptr2

When RefCount = 0 → Free

Problem: Circular references
A ──→ B
▲     │
└─────┘  (Both have RefCount > 0, but unreachable!)

2. Mark and Sweep:
┌─────────────────────────────────────┐
│ Phase 1: Mark (from roots)          │
│                                     │
│  Root → A → B → C                   │
│         ↓                           │
│         D                           │
│                                     │
│  E (unreachable)                    │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│ Phase 2: Sweep (free unmarked)     │
│                                     │
│  Keep: A, B, C, D                   │
│  Free: E                            │
└─────────────────────────────────────┘

3. Generational GC:
┌─────────────────────────────────────┐
│ Young Generation (frequent GC)      │
│ • Most objects die young            │
│ • Fast, small collections           │
├─────────────────────────────────────┤
│ Old Generation (infrequent GC)      │
│ • Long-lived objects                │
│ • Slower, larger collections        │
└─────────────────────────────────────┘
```

### 2. Memory Pools

Pre-allocate memory for specific object types.

```
Memory Pool:

┌─────────────────────────────────────┐
│         Object Pool                  │
│                                     │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐       │
│  │ Obj│ │ Obj│ │ Obj│ │ Obj│       │
│  └────┘ └────┘ └────┘ └────┘       │
│    ▲      ▲      ▲      ▲          │
│    │      │      │      │          │
│  Free   Free   Used   Used         │
└─────────────────────────────────────┘

Benefits:
• Fast allocation (O(1))
• No fragmentation
• Better cache locality
• Predictable performance

Use Cases:
• Network buffers
• Database connections
• Thread pools
```

### 3. Copy-on-Write (COW)

Delay copying until modification.

```
Copy-on-Write:

Initial State (after fork):
┌─────────────────────────────────────┐
│         Physical Memory              │
│  ┌────────────────────────────┐     │
│  │      Shared Page           │     │
│  │      (Read-Only)           │     │
│  └────────────────────────────┘     │
│         ▲              ▲             │
└─────────┼──────────────┼─────────────┘
          │              │
    ┌─────┴────┐   ┌────┴─────┐
    │ Parent   │   │  Child   │
    │ Process  │   │ Process  │
    └──────────┘   └──────────┘

After Child Writes:
┌─────────────────────────────────────┐
│         Physical Memory              │
│  ┌──────────────┐  ┌──────────────┐ │
│  │ Original Page│  │ Copied Page  │ │
│  │ (Read-Only)  │  │ (Read-Write) │ │
│  └──────────────┘  └──────────────┘ │
│         ▲                  ▲         │
└─────────┼──────────────────┼─────────┘
          │                  │
    ┌─────┴────┐       ┌────┴─────┐
    │ Parent   │       │  Child   │
    │ Process  │       │ Process  │
    └──────────┘       └──────────┘

Benefits:
• Fast process creation (fork)
• Memory efficient
• Used by: fork(), mmap(), containers
```

## Memory Performance Optimization

### 1. Huge Pages

Use larger page sizes to reduce TLB misses.

```
Standard Pages vs Huge Pages:

Standard (4 KB pages):
┌────────────────────────────────────┐
│ 1 GB Memory = 262,144 pages       │
│ TLB entries needed: 262,144       │
│ TLB size: ~1,024 entries          │
│ TLB miss rate: HIGH               │
└────────────────────────────────────┘

Huge Pages (2 MB pages):
┌────────────────────────────────────┐
│ 1 GB Memory = 512 pages           │
│ TLB entries needed: 512           │
│ TLB size: ~1,024 entries          │
│ TLB miss rate: LOW                │
└────────────────────────────────────┘

Performance Impact:
• 10-30% improvement for memory-intensive workloads
• Especially beneficial for databases, in-memory caches
```

### 2. NUMA Awareness

Optimize for Non-Uniform Memory Access.

```
NUMA System:

┌──────────────────┐      ┌──────────────────┐
│   Node 0         │      │   Node 1         │
│  ┌──────────┐    │      │  ┌──────────┐    │
│  │ CPU 0-7  │    │      │  │ CPU 8-15 │    │
│  └──────────┘    │      │  └──────────┘    │
│  ┌──────────┐    │      │  ┌──────────┐    │
│  │ 32 GB    │    │      │  │ 32 GB    │    │
│  │ Local    │    │      │  │ Local    │    │
│  └──────────┘    │      │  └──────────┘    │
└────────┬─────────┘      └─────────┬────────┘
         │                          │
         └──────────┬───────────────┘
                    │
              Interconnect
              (slower)

Access Latency:
• Local:  100 ns
• Remote: 200 ns (2x slower!)

Best Practices:
• Allocate memory on same node as CPU
• Pin processes to NUMA nodes
• Use numa_alloc_onnode()
```

### 3. Memory Alignment

Align data structures to cache line boundaries.

```
Cache Line Alignment:

Unaligned (False Sharing):
Cache Line 1 (64 bytes):
┌────────────────┬────────────────┐
│   Counter A    │   Counter B    │
│   (Thread 1)   │   (Thread 2)   │
└────────────────┴────────────────┘
         ▲                ▲
         └────────────────┘
    Both threads invalidate
    the same cache line!

Aligned (No False Sharing):
Cache Line 1:              Cache Line 2:
┌────────────────┐        ┌────────────────┐
│   Counter A    │        │   Counter B    │
│   (Thread 1)   │        │   (Thread 2)   │
└────────────────┘        └────────────────┘

Each thread has its own cache line!

Performance Impact:
• 10-100x improvement in multi-threaded scenarios
```

## Memory Monitoring

### Key Metrics

```
Memory Metrics Dashboard:

┌─────────────────────────────────────┐
│ Total Memory: 64 GB                 │
│ Used: 48 GB (75%)                   │
│ Free: 16 GB (25%)                   │
│ ████████████████████░░░░░░░         │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Swap Usage: 2 GB / 8 GB (25%)       │
│ ██████░░░░░░░░░░░░░░░░░░░░░         │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Page Faults/sec: 1,234              │
│ • Minor: 1,200 (in RAM)             │
│ • Major: 34 (from disk)             │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Cache Hit Rate: 95%                 │
└─────────────────────────────────────┘
```

### Linux Commands

```bash
# Memory information
free -h
cat /proc/meminfo

# Per-process memory
ps aux --sort=-%mem | head
pmap -x <pid>

# Memory statistics
vmstat 1

# NUMA information
numastat
numactl --hardware

# Huge pages
cat /proc/meminfo | grep Huge

# Memory pressure
cat /proc/pressure/memory
```

## Summary

Key takeaways for architects:

1. **Virtual memory provides isolation** - but adds translation overhead
2. **Cache locality is critical** - design data structures accordingly
3. **Memory allocation strategy matters** - stack vs heap vs pools
4. **NUMA awareness** - essential for multi-socket systems
5. **Monitor memory metrics** - page faults, swap usage, cache hit rates
6. **Huge pages** - significant performance boost for large memory workloads
7. **Alignment matters** - avoid false sharing in multi-threaded code

## Next Steps

Continue to [Storage Systems](./04-storage-systems.md) to understand persistent storage technologies and optimization.
