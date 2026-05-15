# Complete OS Understanding — Architect-Level Reference

> **Goal**: Deep understanding of how hardware, OS, and Go runtime interact.  
> Covers everything from transistors to goroutines. Architect-interview ready.

---

## Table of Contents

1. [Hardware Components Overview](#1-hardware-components-overview)
2. [Mechanical Empathy](#2-mechanical-empathy)
3. [CPU Deep Dive](#3-cpu-deep-dive)
4. [CPU Cache — L1, L2, L3](#4-cpu-cache--l1-l2-l3)
5. [Memory Layout](#5-memory-layout)
6. [CPU Memory Read/Write Operations](#6-cpu-memoryreadwrite-operations)
7. [The Kernel](#7-the-kernel)
8. [Processes and Threads](#8-processes-and-threads)
9. [Context Switching](#9-context-switching)
10. [User Mode → Kernel Mode Traps](#10-user-mode--kernel-mode-traps)
11. [Operating System and Virtual Address Space](#11-operating-system-and-virtual-address-space)
12. [Pages and Frames — Virtual Memory Mapping](#12-pages-and-frames--virtual-memory-mapping)
13. [Garbage Collection](#13-garbage-collection)
14. [Metrics and Units](#14-metrics-and-units)
15. [Go's Efficiency — Goroutines and the Runtime](#15-gos-efficiency--goroutines-and-the-runtime)
16. [Virtualisation vs Containerisation](#16-virtualisation-vs-containerisation)
17. [Architect-Level Q&A](#17-architect-level-qa)

---

## 1. Hardware Components Overview

Every computer is built from a small set of physical components. Understanding what each one does — and how fast it is — is the foundation of mechanical empathy.

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Computer System                              │
│                                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────────┐  │
│  │   CPU    │◄──►│   RAM    │    │   Disk   │    │ Network Card │  │
│  │(Processor│    │(Memory)  │    │(Storage) │    │   (NIC)      │  │
│  └────┬─────┘    └──────────┘    └──────────┘    └──────────────┘  │
│       │                                                             │
│  ┌────▼─────────────────────────────────────────────────────────┐  │
│  │                  System Bus / PCIe                           │  │
│  └────┬──────────────┬──────────────┬──────────────┬────────────┘  │
│       │              │              │              │               │
│  ┌────▼────┐   ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐         │
│  │Keyboard │   │ Monitor  │  │  Disk    │  │  NIC     │         │
│  │(Input)  │   │(Display) │  │Controller│  │Controller│         │
│  └─────────┘   └──────────┘  └──────────┘  └──────────┘         │
└─────────────────────────────────────────────────────────────────────┘
```

### Component Roles and Speed

| Component | Role | Typical Speed / Latency |
|-----------|------|------------------------|
| **CPU** | Executes instructions, performs computation | 3–5 GHz clock; ~0.3ns per cycle |
| **RAM** | Holds running programs and data | ~100ns access latency; 50–100 GB/s bandwidth |
| **Disk (SSD)** | Persistent storage | ~100µs latency; 500 MB/s – 7 GB/s |
| **Disk (HDD)** | Persistent storage (spinning) | ~10ms latency; 100–200 MB/s |
| **Network Card** | Sends/receives data over network | 1–100 Gbps; 1ms–100ms RTT |
| **Keyboard** | User input device | Interrupt-driven; human-scale latency |
| **Monitor** | Displays output via GPU/framebuffer | 60–240 Hz refresh; GPU-driven |

### The Speed Hierarchy (Critical for Architects)

```
CPU Registers     ~0.3 ns   ████ (fastest)
L1 Cache          ~1   ns   ████
L2 Cache          ~4   ns   ████
L3 Cache          ~10  ns   ████
RAM               ~100 ns   ████████
SSD (NVMe)        ~100 µs   ████████████████████
SSD (SATA)        ~500 µs   ████████████████████████
HDD               ~10  ms   ████████████████████████████████
Network (LAN)     ~1   ms   ████████████████████████
Network (WAN)     ~100 ms   ████████████████████████████████████████
```

> **Architect insight**: A cache miss that falls through to RAM is 100x slower than an L1 hit. A disk read is 100,000x slower than RAM. Design your systems around this hierarchy.

---

## 2. Mechanical Empathy

### What Is Mechanical Empathy?

The term comes from Formula 1 racing. Driver Jackie Stewart said the best drivers have **mechanical empathy** — they understand how the car works and drive in harmony with the machine rather than fighting it.

In software: **mechanical empathy means writing code that works with the hardware, not against it.**

A developer with mechanical empathy knows:
- That sequential memory access is faster than random access (CPU prefetcher)
- That a cache miss costs 100x more than a cache hit
- That a context switch has overhead
- That a system call crosses the user/kernel boundary (expensive)
- That false sharing kills multi-core performance

### Why It Matters for Architects

| Without Mechanical Empathy | With Mechanical Empathy |
|---------------------------|------------------------|
| Random memory access patterns | Sequential / cache-friendly access |
| Allocating millions of small objects | Pooling and reusing objects |
| Spawning OS threads per request | Using goroutines / async I/O |
| Ignoring cache line boundaries | Padding structs to avoid false sharing |
| Blocking on disk I/O in hot path | Async I/O, buffering, batching |

### Concrete Examples

**Example 1 — Cache-friendly iteration (Go)**
```go
// BAD: column-major access — cache unfriendly
// Each access jumps 1000 * 8 bytes = 8KB in memory
for col := 0; col < 1000; col++ {
    for row := 0; row < 1000; row++ {
        sum += matrix[row][col]  // jumps around in memory
    }
}

// GOOD: row-major access — cache friendly
// Each access is the next 8 bytes — prefetcher loves this
for row := 0; row < 1000; row++ {
    for col := 0; col < 1000; col++ {
        sum += matrix[row][col]  // sequential memory access
    }
}
// Row-major is ~5-10x faster on real hardware
```

**Example 2 — False sharing (Go)**
```go
// BAD: two goroutines write to fields on the same cache line
type Counters struct {
    A int64  // bytes 0-7
    B int64  // bytes 8-15  — same 64-byte cache line as A!
}
// When goroutine 1 writes A and goroutine 2 writes B simultaneously,
// the CPU must invalidate the other core's cache line → massive slowdown

// GOOD: pad to separate cache lines
type Counters struct {
    A   int64
    _   [56]byte  // padding to fill 64-byte cache line
    B   int64
    _   [56]byte
}
```

**Example 3 — System call batching**
```go
// BAD: one syscall per byte (write syscall is expensive)
for _, b := range data {
    os.Stdout.Write([]byte{b})  // syscall per byte!
}

// GOOD: buffer and flush — one syscall for many bytes
w := bufio.NewWriter(os.Stdout)
for _, b := range data {
    w.WriteByte(b)  // writes to userspace buffer
}
w.Flush()  // one syscall
```

---

## 3. CPU Deep Dive

### What Is a CPU?

The CPU (Central Processing Unit) is the brain of the computer. It executes instructions — billions per second. Every line of code you write eventually becomes machine instructions that the CPU executes one by one (or in parallel via pipelining and superscalar execution).

### The Fetch-Decode-Execute Cycle

Every CPU instruction goes through this cycle:

```
┌─────────────────────────────────────────────────────────────┐
│                   CPU Instruction Cycle                     │
│                                                             │
│  1. FETCH    Read next instruction from memory (via PC)     │
│       ↓                                                     │
│  2. DECODE   Figure out what the instruction means          │
│       ↓                                                     │
│  3. EXECUTE  Perform the operation (add, load, jump, etc.)  │
│       ↓                                                     │
│  4. WRITEBACK Store result back to register or memory       │
│       ↓                                                     │
│  (repeat — billions of times per second)                    │
└─────────────────────────────────────────────────────────────┘
```

### Logical Components of a CPU

```
┌──────────────────────────────────────────────────────────────┐
│                          CPU Die                             │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                  Control Unit (CU)                  │    │
│  │  Orchestrates fetch/decode/execute, manages timing  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌──────────────────────┐  ┌──────────────────────────┐     │
│  │  ALU (Arithmetic     │  │  FPU (Floating Point     │     │
│  │  Logic Unit)         │  │  Unit)                   │     │
│  │  +, -, *, /, AND,    │  │  float32/float64 math    │     │
│  │  OR, XOR, shifts     │  │                          │     │
│  └──────────────────────┘  └──────────────────────────┘     │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                    Registers                        │    │
│  │  PC  (Program Counter)  — address of next instr     │    │
│  │  SP  (Stack Pointer)    — top of current stack      │    │
│  │  BP  (Base Pointer)     — base of current frame     │    │
│  │  IR  (Instruction Reg)  — current instruction       │    │
│  │  RAX, RBX, RCX...       — general purpose (x86-64) │    │
│  │  FLAGS                  — zero, carry, overflow...  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Cache Hierarchy                        │    │
│  │  L1 (per core, ~32KB)  →  L2 (per core, ~256KB)    │    │
│  │  L3 (shared, ~8-32MB)                               │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  MMU (Memory Management Unit)                       │    │
│  │  Translates virtual addresses → physical addresses  │    │
│  │  Contains TLB (Translation Lookaside Buffer)        │    │
│  └─────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────┘
```

### How CPU Communicates with Other Components

#### CPU ↔ RAM

```
CPU                    Memory Controller         RAM
 │                           │                   │
 │  Virtual Address          │                   │
 │──────────────────────────►│                   │
 │                           │  Physical Address │
 │                           │──────────────────►│
 │                           │                   │
 │                           │◄── Data (64 bytes)─│
 │◄──────────────────────────│                   │
 │  Data in register         │                   │
```

- CPU puts a **virtual address** on the address bus
- MMU translates it to a **physical address** using the page table
- Memory controller fetches the data from RAM
- Data travels back on the **data bus** (64-bit wide = 8 bytes per transfer)
- Modern CPUs fetch a full **cache line (64 bytes)** at once, not just the requested bytes
- Latency: ~100ns (about 300 CPU cycles wasted waiting)

#### CPU ↔ Disk

The CPU does NOT talk to disk directly. The path is:

```
CPU (user code)
    │  system call (read/write)
    ▼
Kernel (VFS layer)
    │
    ▼
Block Device Driver
    │
    ▼
Disk Controller (via PCIe/SATA/NVMe)
    │
    ▼
Physical Disk (SSD/HDD)
    │
    ▼  DMA transfer (disk → RAM, bypassing CPU)
RAM (kernel buffer)
    │
    ▼  copy_to_user
User process memory
```

- **DMA (Direct Memory Access)**: The disk controller writes data directly into RAM without involving the CPU. When done, it sends an **interrupt** to the CPU.
- This frees the CPU to do other work while the disk transfer happens.

#### CPU ↔ Network Card (NIC)

```
CPU                    NIC                    Network
 │                      │                       │
 │  write to TX ring    │                       │
 │─────────────────────►│                       │
 │                      │──── packet ──────────►│
 │                      │                       │
 │                      │◄─── packet ───────────│
 │                      │  DMA to RX ring       │
 │                      │  (writes to RAM)      │
 │◄── interrupt ────────│                       │
 │  (packet arrived)    │                       │
```

- Sending: CPU writes packet descriptor to NIC's **TX ring buffer** (in RAM). NIC reads it via DMA and sends.
- Receiving: NIC writes incoming packet to **RX ring buffer** (in RAM) via DMA, then interrupts CPU.
- Modern NICs use **NAPI** (New API) — polling instead of interrupt per packet at high load.

#### CPU ↔ Keyboard

```
User presses key
    │
    ▼
Keyboard controller generates scan code
    │
    ▼
Sends IRQ (Interrupt Request) on IRQ1 line
    │
    ▼
CPU pauses current work, saves state
    │
    ▼
Jumps to keyboard interrupt handler (in kernel)
    │
    ▼
Reads scan code from I/O port 0x60
    │
    ▼
Translates to keycode, puts in input buffer
    │
    ▼
Resumes previous work
```

- Keyboard is **interrupt-driven** — the CPU doesn't poll the keyboard; the keyboard interrupts the CPU when a key is pressed.
- This is efficient: CPU does useful work 100% of the time and only handles keyboard events when they occur.

---

## 4. CPU Cache — L1, L2, L3

### Why Cache Exists

RAM is ~100ns away from the CPU. At 3GHz, the CPU executes one instruction every ~0.3ns. Without cache, the CPU would spend 99% of its time waiting for memory. Cache is fast SRAM (Static RAM) built directly on the CPU die.

### Cache Hierarchy

```
┌─────────────────────────────────────────────────────────────────┐
│                        CPU Core 0                               │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  L1 Cache (per core)                                    │   │
│  │  Size: 32–64 KB (split: 32KB data + 32KB instruction)   │   │
│  │  Latency: ~1 ns (4 cycles)                              │   │
│  │  Bandwidth: ~1 TB/s                                     │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  L2 Cache (per core)                                    │   │
│  │  Size: 256 KB – 1 MB                                    │   │
│  │  Latency: ~4 ns (12 cycles)                             │   │
│  │  Bandwidth: ~400 GB/s                                   │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  L3 Cache (shared across all cores)                             │
│  Size: 8–64 MB                                                  │
│  Latency: ~10–40 ns (30–100 cycles)                             │
│  Bandwidth: ~200 GB/s                                           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  RAM (DRAM)                                                     │
│  Size: 8 GB – 1 TB                                              │
│  Latency: ~100 ns (300 cycles)                                  │
│  Bandwidth: 50–100 GB/s                                         │
└─────────────────────────────────────────────────────────────────┘
```

### Latency Numbers Every Architect Must Know

| Level | Latency | Relative to L1 |
|-------|---------|----------------|
| L1 Cache | ~1 ns | 1x |
| L2 Cache | ~4 ns | 4x |
| L3 Cache | ~10–40 ns | 10–40x |
| RAM | ~100 ns | 100x |
| NVMe SSD | ~100 µs | 100,000x |
| SATA SSD | ~500 µs | 500,000x |
| HDD | ~10 ms | 10,000,000x |
| Network (LAN) | ~1 ms | 1,000,000x |

### Cache Hit vs Cache Miss

**Cache Hit** — data is found in cache:
```
CPU requests address 0x1234
    │
    ▼
Check L1 cache → FOUND ✓
    │
    ▼
Return data in ~1ns
CPU continues immediately
```

**Cache Miss** — data not in cache:
```
CPU requests address 0x1234
    │
    ▼
Check L1 cache → MISS ✗
    │
    ▼
Check L2 cache → MISS ✗
    │
    ▼
Check L3 cache → MISS ✗
    │
    ▼
Go to RAM → fetch 64-byte cache line (~100ns)
    │
    ▼
Load into L3, then L2, then L1
    │
    ▼
Return data to CPU
CPU was stalled for ~300 cycles
```

On a cache miss, the CPU **stalls** (waits). Modern CPUs use **out-of-order execution** to do other work while waiting, but a cache miss is still very expensive.

### Cache Line — The Unit of Transfer

Cache does not transfer individual bytes. It transfers **cache lines** of **64 bytes** at a time.

```
Memory:
Address: 0x00  0x08  0x10  0x18  0x20  0x28  0x30  0x38
         ├─────┼─────┼─────┼─────┼─────┼─────┼─────┤
         │  A  │  B  │  C  │  D  │  E  │  F  │  G  │  H  │
         └─────────────────────────────────────────────────┘
         ◄──────────── 64-byte cache line ────────────────►

If you access A, the CPU loads ALL 8 values (A through H) into cache.
Next access to B, C, D... → cache hit (already loaded).
```

This is why **sequential access** is fast and **random access** is slow.

### Is Cache Bounded or Unbounded?

**Bounded** — completely fixed by hardware. You cannot add more cache at runtime.

| Cache | Typical Size | Fixed? |
|-------|-------------|--------|
| L1 | 32–64 KB per core | Yes — soldered on die |
| L2 | 256 KB – 1 MB per core | Yes — soldered on die |
| L3 | 8–64 MB shared | Yes — soldered on die |

When the cache is full, the CPU **evicts** the least recently used (LRU) cache line to make room.

### Cache Invalidation — MESI Protocol

On multi-core CPUs, each core has its own L1/L2 cache. If Core 0 and Core 1 both cache the same memory address, and Core 0 modifies it, Core 1's copy is now **stale**. The **MESI protocol** handles this:

| State | Meaning |
|-------|---------|
| **M** (Modified) | This core has the only copy, and it's been modified (dirty) |
| **E** (Exclusive) | This core has the only copy, and it's clean |
| **S** (Shared) | Multiple cores have this cache line (all clean) |
| **I** (Invalid) | This cache line is stale — must re-fetch |

```
Core 0 writes to address X:
  1. Core 0 sends "invalidate" message to all other cores
  2. Other cores mark their copy of X as Invalid (I)
  3. Core 0's copy transitions to Modified (M)
  4. If Core 1 now reads X → cache miss → must fetch from Core 0 or RAM
```

**No TTL** — cache invalidation is not time-based. It's event-driven (write invalidates). There is no expiry timer in CPU cache.

### Write Policies

| Policy | How it works | Pros | Cons |
|--------|-------------|------|------|
| **Write-through** | Write to cache AND RAM simultaneously | RAM always up-to-date | Slower writes |
| **Write-back** | Write to cache only; flush to RAM on eviction | Faster writes | RAM may be stale |

Modern CPUs use **write-back** for performance.

### False Sharing — The Hidden Performance Killer

Two goroutines on different cores write to **different variables** that happen to be on the **same 64-byte cache line**. Each write invalidates the other core's cache line, causing constant cache misses.

```go
// BAD: A and B are on the same cache line
type Stats struct {
    Requests int64  // offset 0
    Errors   int64  // offset 8  ← same 64-byte line as Requests
}

var s Stats
go func() { atomic.AddInt64(&s.Requests, 1) }()  // Core 0
go func() { atomic.AddInt64(&s.Errors, 1) }()    // Core 1
// Core 0 and Core 1 fight over the same cache line → false sharing

// GOOD: pad to separate cache lines
type Stats struct {
    Requests int64
    _        [56]byte  // pad to 64 bytes
    Errors   int64
    _        [56]byte
}
```

---

## 5. Memory Layout

### What Happens When You Run a Program

```
1. You type: ./myapp
2. Shell calls execve() system call
3. Kernel:
   a. Reads ELF binary from disk
   b. Creates a new process (PCB, PID)
   c. Sets up virtual address space
   d. Maps Text segment (code) from binary
   e. Maps Data segment (initialized globals) from binary
   f. Allocates BSS segment (zeroed uninitialized globals)
   g. Allocates stack (initial size ~8MB on Linux)
   h. Sets up heap (initially empty, grows on demand)
   i. Loads dynamic linker (ld.so) if needed
   j. Sets PC (Program Counter) to entry point (_start → main)
4. CPU starts executing at main()
```

### Full Virtual Address Space Layout

```
High Address  0xFFFFFFFFFFFFFFFF
┌─────────────────────────────────────────────────────┐
│                  Kernel Space                       │
│  (OS kernel, drivers, kernel stacks)                │
│  User code CANNOT access this — segfault if tried   │
├─────────────────────────────────────────────────────┤  0xFFFF800000000000
│                                                     │
│                    Stack                            │
│  ┌─────────────────────────────────────────────┐   │
│  │  main() frame: argc, argv, local vars       │   │
│  │  ─────────────────────────────────────────  │   │
│  │  foo() frame: params, locals, return addr   │   │
│  │  ─────────────────────────────────────────  │   │
│  │  bar() frame: params, locals, return addr   │   │
│  └─────────────────────────────────────────────┘   │
│                      ↓ grows downward               │
│                                                     │
│              (free gap — unused space)              │
│                                                     │
│                      ↑ grows upward                 │
│                    Heap                             │
│  Dynamic allocations: new(T), make([]T,n), malloc   │
│                                                     │
├─────────────────────────────────────────────────────┤
│                  BSS Segment                        │
│  Uninitialized global/package-level vars            │
│  var x int  (no value)  → zeroed by OS at startup   │
├─────────────────────────────────────────────────────┤
│                  Data Segment                       │
│  Initialized global/package-level vars              │
│  var x int = 42  → stored in binary, loaded here    │
├─────────────────────────────────────────────────────┤
│                  Text Segment                       │
│  Compiled machine instructions (READ-ONLY)          │
│  Writing here → segmentation fault                  │
│  Shared between multiple instances of same program  │
└─────────────────────────────────────────────────────┘
Low Address   0x0000000000000000
```

### Where Does Code Live?

```go
package main

var globalInit = 42        // → Data segment (initialized global)
var globalZero int         // → BSS segment (uninitialized global)

func add(a, b int) int {   // → Text segment (compiled to machine code)
    result := a + b        // → Stack (local variable, lives in stack frame)
    return result
}

func main() {
    x := 10               // → Stack (local variable)
    y := 20               // → Stack (local variable)
    
    p := new(int)         // → Heap (pointer returned, escapes to heap)
    *p = 99
    
    s := make([]int, 100) // → Heap (slice backing array on heap)
    _ = s
    
    z := add(x, y)        // → Stack frame for add() pushed, then popped
    _ = z
}
```

### Stack Frames — How Function Calls Work

```
Before main() calls add(10, 20):

Stack (grows downward ↓):
┌─────────────────────────────────┐  ← high address
│         main() frame            │
│  x = 10                         │
│  y = 20                         │
│  return address (after call)    │
│  saved registers                │
└─────────────────────────────────┘  ← SP (stack pointer)

After call to add(10, 20):

┌─────────────────────────────────┐  ← high address
│         main() frame            │
│  x = 10                         │
│  y = 20                         │
│  return address                 │
├─────────────────────────────────┤
│         add() frame             │  ← new frame pushed
│  a = 10  (parameter)            │
│  b = 20  (parameter)            │
│  result = 30 (local var)        │
│  return address → back to main  │
└─────────────────────────────────┘  ← SP (stack pointer, moved down)

After add() returns:
  - add() frame is popped (SP moves back up)
  - return value (30) placed in register RAX
  - main() resumes from saved return address
```

### Stack vs Heap — When Does Each Get Used?

The Go compiler performs **escape analysis** at compile time to decide:

```go
// STACK — variable does not escape
func sum(a, b int) int {
    result := a + b   // result stays in this function → stack
    return result     // return VALUE, not pointer → stack is fine
}

// HEAP — variable escapes (pointer returned to caller)
func newCounter() *int {
    count := 0        // count must outlive this function
    return &count     // returning pointer → count escapes to heap
}

// HEAP — variable escapes (stored in interface)
func store(v interface{}) {
    cache = v         // v stored globally → escapes to heap
}

// Check escape analysis:
// go build -gcflags="-m" ./...
```

**Rule of thumb**:
- If a variable's lifetime is bounded by the function → **stack**
- If a variable outlives the function (pointer returned, stored globally, sent to channel) → **heap**

### What Happens on Stack Overflow?

```
Recursive function calls itself too many times:

main() → foo() → foo() → foo() → ... (10,000 deep)

Stack grows downward until it hits the heap or guard page.
OS detects the violation → sends SIGSEGV to process.
Program crashes: "runtime: goroutine stack exceeds 1000000000-byte limit"

Go is special: goroutine stacks START at 2KB and GROW dynamically
(copied to a larger location). Stack overflow only happens if you
exceed the maximum (default 1GB).
```

---

## 6. CPU Memory Read/Write Operations

### What Happens When CPU Reads from Memory

Every memory access goes through this pipeline:

```
CPU executes: MOV RAX, [0x7fff1234]  (load from virtual address 0x7fff1234)

Step 1: Check TLB (Translation Lookaside Buffer)
  ├── TLB HIT  → get physical address directly (~1ns extra)
  └── TLB MISS → walk page table in RAM (~100ns extra)
                  then cache the result in TLB

Step 2: Check L1 Cache
  ├── HIT  → return data (~1ns)
  └── MISS → check L2

Step 3: Check L2 Cache
  ├── HIT  → return data, load into L1 (~4ns)
  └── MISS → check L3

Step 4: Check L3 Cache
  ├── HIT  → return data, load into L2 and L1 (~10-40ns)
  └── MISS → go to RAM

Step 5: RAM Access
  → Fetch 64-byte cache line from physical RAM (~100ns)
  → Load into L3, L2, L1
  → Return requested bytes to CPU register

If page not in RAM (page fault):
  → OS loads page from disk (~100µs–10ms)
  → Then retry from Step 1
```

### What Happens When CPU Writes to Memory

```
CPU executes: MOV [0x7fff1234], RAX  (store to virtual address)

1. Translate virtual → physical (TLB / page table)
2. Write to L1 cache (write-back policy: mark cache line as dirty)
3. Cache line is NOT immediately written to RAM
4. When cache line is evicted (cache full), dirty line is written to RAM
5. MESI protocol ensures other cores see the updated value
```

### DMA — Direct Memory Access

Without DMA, every byte from disk/network would require CPU involvement:
```
Without DMA:
  Disk → CPU register → RAM  (CPU busy for every byte)

With DMA:
  Disk → DMA Controller → RAM  (CPU free to do other work)
  DMA Controller sends interrupt when done → CPU resumes
```

DMA is why disk and network I/O is non-blocking at the hardware level. The CPU sets up the transfer and moves on.

### Page Fault — When Memory Isn't in RAM

```
Process accesses virtual address 0xABCD1234

MMU checks page table:
  ├── Page present in RAM → normal access
  └── Page NOT present (present bit = 0) → PAGE FAULT

CPU raises page fault exception
  → Jumps to kernel's page fault handler

Kernel determines why:
  ├── Valid address, page swapped to disk
  │     → Read page from swap space into a free frame
  │     → Update page table (present bit = 1)
  │     → Resume process (retry the instruction)
  │
  ├── Valid address, first access (demand paging / copy-on-write)
  │     → Allocate new frame, zero it
  │     → Update page table
  │     → Resume process
  │
  └── Invalid address (null pointer, out of bounds)
        → Send SIGSEGV to process
        → Process crashes: "segmentation fault"
```

---

## 7. The Kernel

### What Is the Kernel?

The kernel is the **core of the operating system** — the software that runs with full hardware privileges and manages all resources. It is the only software that can directly talk to hardware.

```
┌─────────────────────────────────────────────────────────────┐
│                    User Space                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  Chrome  │  │  Go app  │  │  Python  │  │  MySQL   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│                                                             │
│  (Ring 3 — restricted, cannot access hardware directly)    │
├─────────────────────────────────────────────────────────────┤
│                  System Call Interface                      │
│  read(), write(), open(), fork(), mmap(), socket()...       │
├─────────────────────────────────────────────────────────────┤
│                    Kernel Space                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Process Manager  │  Memory Manager  │  Scheduler    │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  File System (VFS) │  Network Stack  │  IPC          │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  Device Drivers: disk, NIC, keyboard, GPU...         │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  (Ring 0 — full hardware access)                           │
├─────────────────────────────────────────────────────────────┤
│                    Hardware                                 │
│  CPU  │  RAM  │  Disk  │  NIC  │  Keyboard  │  GPU         │
└─────────────────────────────────────────────────────────────┘
```

### Kernel Responsibilities

| Responsibility | What It Does |
|---------------|-------------|
| **Process Management** | Create, schedule, terminate processes and threads |
| **Memory Management** | Virtual memory, page tables, swap, OOM killer |
| **File System** | VFS abstraction over ext4, APFS, NTFS, etc. |
| **Device Drivers** | Translate OS calls to hardware-specific commands |
| **Network Stack** | TCP/IP implementation, socket management |
| **Security** | Permissions, namespaces, capabilities, seccomp |
| **IPC** | Pipes, signals, shared memory, message queues |
| **Interrupt Handling** | Respond to hardware events (keyboard, NIC, timer) |

### Kernel Space vs User Space

| | User Space | Kernel Space |
|--|-----------|-------------|
| CPU ring | Ring 3 (restricted) | Ring 0 (privileged) |
| Hardware access | Not allowed | Full access |
| Memory access | Own virtual space only | All physical memory |
| Crash impact | Only that process dies | Entire system crashes (kernel panic) |
| How to enter | System call / interrupt | Always running or entered via trap |

---

## 8. Processes and Threads

### Program vs Process

```
Program (static):                    Process (dynamic):
┌─────────────────────┐              ┌─────────────────────────────┐
│  myapp binary       │   execve()   │  PID: 1234                  │
│  on disk            │ ──────────►  │  Virtual address space      │
│  (ELF file)         │              │  Open file descriptors      │
│  Code + data        │              │  CPU registers (state)      │
│  No state           │              │  Stack, Heap                │
│  No PID             │              │  Running, Sleeping, Zombie  │
└─────────────────────┘              └─────────────────────────────┘

YES — a process IS a running instance of a program.
Two instances of the same program = two separate processes,
each with their own PID, memory space, and state.
```

### Process Control Block (PCB)

The OS tracks every process using a PCB (also called task_struct in Linux):

```
PCB / task_struct:
┌─────────────────────────────────────────────────────┐
│  PID          — process ID (unique number)          │
│  PPID         — parent process ID                   │
│  State        — Running / Ready / Blocked / Zombie  │
│  PC           — program counter (next instruction)  │
│  Registers    — saved CPU register values           │
│  Stack ptr    — current stack pointer               │
│  Memory maps  — page table pointer, heap/stack info │
│  Open files   — file descriptor table               │
│  Signals      — pending signals, signal handlers    │
│  Priority     — scheduling priority / nice value    │
│  CPU time     — user time, kernel time used         │
│  Owner        — UID, GID (who owns this process)    │
└─────────────────────────────────────────────────────┘
```

### What Is a Thread?

A thread is a **lightweight unit of execution within a process**. All threads in a process share the same address space (code, heap, globals) but each has its own:
- Stack
- Program Counter
- Registers
- Thread ID (TID)

```
Process (PID 1234)
┌─────────────────────────────────────────────────────────────┐
│  Shared:                                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  Code    │  │  Heap    │  │  Globals │  │  Files   │   │
│  │ (Text)   │  │          │  │(Data/BSS)│  │  (FDs)   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│                                                             │
│  Per-thread (private):                                      │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │  Thread 1        │  │  Thread 2        │                │
│  │  Stack (1MB)     │  │  Stack (1MB)     │                │
│  │  PC, Registers   │  │  PC, Registers   │                │
│  │  TID: 1234       │  │  TID: 1235       │                │
│  └──────────────────┘  └──────────────────┘                │
└─────────────────────────────────────────────────────────────┘
```

### Process vs Thread Comparison

| Property | Process | Thread |
|----------|---------|--------|
| Memory | Own address space (isolated) | Shared address space |
| Creation cost | High (~1ms, fork+exec) | Low (~10µs) |
| Communication | IPC (pipes, sockets, shared mem) | Shared memory (direct) |
| Crash isolation | Crash doesn't affect others | Crash kills entire process |
| Context switch | Expensive (flush TLB, swap page tables) | Cheaper (same address space) |
| Overhead | ~1MB+ per process | ~1MB stack per thread |
| Parallelism | True parallelism (separate memory) | True parallelism (shared memory) |
| Synchronization | Not needed for memory | Needed (mutex, atomic, etc.) |

### Process States

```
                    fork()
                      │
                      ▼
                  ┌────────┐
                  │  New   │
                  └───┬────┘
                      │ admitted
                      ▼
  ┌──────────────►┌────────┐
  │  I/O complete │ Ready  │◄──────────────────┐
  │               └───┬────┘                   │
  │                   │ scheduled               │
  │                   ▼                         │
  │               ┌────────┐  preempted         │
  │               │Running │──────────────────► │
  │               └───┬────┘                   │
  │                   │ I/O wait / sleep        │
  │                   ▼                         │
  │               ┌────────┐                   │
  └───────────────│Blocked │                   │
                  └───┬────┘                   │
                      │ exit()                  │
                      ▼
                  ┌────────┐
                  │ Zombie │ (waiting for parent to read exit code)
                  └────────┘
```

---

## 9. Context Switching

### What Is Context Switching?

Context switching is the OS **saving the state of one process/thread and loading the state of another** so the CPU can run multiple tasks concurrently (even on a single core).

```
Time ──────────────────────────────────────────────────────────►

CPU:  [Process A][Process A][SWITCH][Process B][Process B][SWITCH][Process A]
                             ↑                             ↑
                        context switch               context switch
                        (~1-10µs)                    (~1-10µs)
```

### Steps in a Context Switch

```
1. Timer interrupt fires (or process blocks on I/O)
   │
   ▼
2. CPU switches to kernel mode (Ring 3 → Ring 0)
   │
   ▼
3. Kernel saves Process A's context into its PCB:
   - All CPU registers (RAX, RBX, RCX, RDX, RSI, RDI, RSP, RBP, RIP...)
   - Program Counter (where to resume)
   - Stack Pointer
   - CPU flags
   │
   ▼
4. Kernel runs scheduler to pick next process (Process B)
   │
   ▼
5. Kernel loads Process B's context from its PCB:
   - Restore all registers
   - Switch page tables (CR3 register on x86) → TLB flush!
   - Restore stack pointer
   │
   ▼
6. CPU switches back to user mode (Ring 0 → Ring 3)
   │
   ▼
7. Process B resumes from where it left off
```

**The expensive part**: switching page tables flushes the TLB. Process B's memory accesses will all be TLB misses initially → many trips to RAM to rebuild the TLB.

### CPU Scheduling Algorithms

#### 1. Round Robin (RR)
Each process gets a fixed time slice (quantum, e.g., 10ms). After the quantum expires, the next process runs.

```
Process: A  B  C  A  B  C  A  B  C
Time:    10 10 10 10 10 10 10 10 10 ms
```
- Simple, fair
- Good for interactive systems
- Bad for CPU-bound tasks (too many context switches)

#### 2. Priority Scheduling
Each process has a priority. Higher priority runs first.

```
Priority 1 (highest): System processes
Priority 2:           Interactive apps
Priority 3:           Background tasks
Priority 4 (lowest):  Batch jobs
```
- Risk: **starvation** — low priority processes may never run
- Fix: **aging** — gradually increase priority of waiting processes

#### 3. CFS — Completely Fair Scheduler (Linux default)

Linux uses CFS. Instead of fixed time slices, it tracks **virtual runtime** (vruntime) — how much CPU time each process has used. The process with the lowest vruntime runs next.

```
Red-Black Tree (sorted by vruntime):

        [P3: 100ms]
       /            [P1: 50ms]      [P5: 200ms]
   /
[P2: 30ms]  ← runs next (lowest vruntime)
```

- Always picks the "most starved" process
- Weighted by priority (nice value): lower nice = more CPU time
- O(log n) scheduling decisions

#### 4. MLFQ — Multi-Level Feedback Queue

Multiple queues with different priorities and time slices:

```
Queue 0 (highest priority, 8ms quantum):   Interactive tasks
Queue 1 (medium priority, 16ms quantum):   Mixed tasks
Queue 2 (lowest priority, 32ms quantum):   CPU-bound batch jobs

Rules:
- New process starts in Queue 0
- If it uses its full quantum → demoted to lower queue
- If it blocks before quantum expires → stays in same queue
- Periodically boost all processes back to Queue 0 (prevent starvation)
```

### Cost of Context Switching

| Component | Cost |
|-----------|------|
| Save/restore registers | ~200ns |
| TLB flush (page table switch) | ~1-10µs |
| Cache warming (cold caches after switch) | ~10-100µs |
| Total per context switch | ~1-10µs |

At 1000 context switches/second → 1-10ms/second wasted = 0.1-1% overhead.
At 100,000 context switches/second → significant overhead.

```go
// Check context switches in Go
// Too many goroutines → too many context switches
// Use GOMAXPROCS goroutines for CPU-bound work
runtime.GOMAXPROCS(runtime.NumCPU())
```

---

## 10. User Mode → Kernel Mode Traps

### CPU Privilege Rings

Modern CPUs have hardware-enforced privilege levels called **rings**:

```
┌─────────────────────────────────────────────────────────────┐
│                    Ring 0 — Kernel                          │
│  Full hardware access. Can execute any instruction.         │
│  Can read/write any memory. Can configure hardware.         │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Ring 1, 2 (rarely used)                │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │           Ring 3 — User Space               │   │   │
│  │  │  Restricted. Cannot access hardware.        │   │   │
│  │  │  Cannot access kernel memory.               │   │   │
│  │  │  Your Go/Java/Python code runs here.        │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### How Does User Code Enter the Kernel?

There are three mechanisms:

#### 1. System Calls (Intentional)

The program explicitly asks the kernel to do something:

```
User code calls: read(fd, buf, n)
    │
    ▼
C library (glibc/libc) sets up arguments in registers:
  RAX = syscall number (e.g., 0 = read on Linux x86-64)
  RDI = fd
  RSI = buf pointer
  RDX = n
    │
    ▼
Executes SYSCALL instruction (x86-64) or INT 0x80 (legacy x86)
    │
    ▼  ← CPU switches to Ring 0 here
Kernel's syscall handler runs:
  - Validates arguments
  - Performs the operation (reads from file descriptor)
  - Puts return value in RAX
    │
    ▼
SYSRET instruction → CPU switches back to Ring 3
    │
    ▼
User code continues with return value
```

**Cost**: ~100-1000ns per system call (mode switch + kernel work + mode switch back)

#### 2. Hardware Interrupts (Asynchronous)

Hardware signals the CPU that something needs attention:

```
Examples:
  - Timer interrupt (every 1-10ms) → triggers scheduler
  - Keyboard interrupt (key pressed) → read scan code
  - NIC interrupt (packet arrived) → process network data
  - Disk interrupt (I/O complete) → wake up waiting process

Flow:
  CPU is executing user code
      │
      ▼  (interrupt arrives on interrupt line)
  CPU finishes current instruction
      │
      ▼
  CPU saves current state (registers, PC, flags)
      │
      ▼  ← switches to Ring 0
  CPU looks up Interrupt Descriptor Table (IDT)
  Jumps to interrupt handler for that IRQ number
      │
      ▼
  Kernel handles the interrupt
      │
      ▼
  IRET instruction → restores saved state, returns to Ring 3
      │
      ▼
  User code continues (unaware the interrupt happened)
```

#### 3. Exceptions / Faults (Synchronous, caused by CPU)

The CPU itself detects an error and traps into the kernel:

```
Examples and what happens:

Page Fault (accessing unmapped memory):
  User code: MOV RAX, [0xDEADBEEF]
  → CPU: page not present → trap to kernel
  → Kernel: load page from disk, or send SIGSEGV

Divide by Zero:
  User code: DIV RCX  (where RCX = 0)
  → CPU: division by zero exception → trap to kernel
  → Kernel: send SIGFPE to process → process crashes

Illegal Instruction:
  User code: executes invalid opcode
  → CPU: invalid opcode exception → trap to kernel
  → Kernel: send SIGILL to process

Breakpoint (debugger):
  Debugger inserts INT3 instruction
  → CPU: breakpoint exception → trap to kernel
  → Kernel: notify debugger (ptrace)

System Call (INT 0x80 / SYSCALL):
  User code: SYSCALL instruction
  → CPU: software interrupt → trap to kernel
  → Kernel: execute requested service
```

### Real-World Examples of Kernel Traps in Go

```go
// 1. File I/O → read() syscall
data, err := os.ReadFile("config.json")
// Internally: open() syscall + read() syscall + close() syscall

// 2. Memory allocation → mmap() or brk() syscall
// When Go heap needs to grow, runtime calls mmap() to get more memory from OS
p := make([]byte, 1024*1024*100)  // 100MB → may trigger mmap() syscall

// 3. Network → socket(), connect(), send(), recv() syscalls
resp, err := http.Get("https://api.example.com")
// Internally: socket() + connect() + send() + recv() syscalls

// 4. Goroutine sleep → nanosleep() syscall
time.Sleep(100 * time.Millisecond)
// Internally: nanosleep() syscall → process blocked in kernel

// 5. Mutex → futex() syscall (when contended)
mu.Lock()
// If uncontended: pure userspace (atomic CAS, no syscall)
// If contended: futex() syscall to sleep until lock available

// 6. Page fault (transparent to Go code)
var m map[string]int  // nil map
m["key"] = 1          // nil pointer dereference → page fault → SIGSEGV → panic
```

---

## 11. Operating System and Virtual Address Space

### What Is an Operating System?

The OS is the software layer between hardware and applications. It provides:
1. **Abstraction** — hide hardware complexity (files instead of disk sectors)
2. **Resource management** — share CPU, RAM, disk fairly between processes
3. **Protection** — prevent processes from interfering with each other
4. **Services** — file system, networking, IPC, security

### How the OS Creates Virtual Address Space Per Process

When a new process is created (fork/exec), the OS:

```
1. Allocates a new Page Table (per-process data structure)
   
2. Maps the process's virtual address space:
   ┌─────────────────────────────────────────────────────┐
   │  Virtual Address  │  Physical Frame  │  Permissions │
   ├─────────────────────────────────────────────────────┤
   │  0x400000 (text)  │  Frame 1024      │  R-X (exec)  │
   │  0x600000 (data)  │  Frame 2048      │  RW- (read/write) │
   │  0x800000 (heap)  │  Frame 3072      │  RW-         │
   │  0x7fff0000(stack)│  Frame 4096      │  RW-         │
   └─────────────────────────────────────────────────────┘

3. Loads CR3 register with pointer to this process's page table
   (CR3 is the "page table base register" on x86)

4. Each process has its OWN page table → complete memory isolation
```

### Memory Isolation Between Processes

```
Process A (PID 100)          Process B (PID 200)
Virtual: 0x1000              Virtual: 0x1000
    │                            │
    ▼ (page table A)             ▼ (page table B)
Physical: Frame 500          Physical: Frame 800
    │                            │
    ▼                            ▼
RAM: [A's data here]         RAM: [B's data here]

Same virtual address → DIFFERENT physical memory.
Process A cannot read Process B's memory.
If A tries to access B's physical frame → page fault → SIGSEGV.
```

### Program vs Process — The Relationship

```
Program (binary file on disk):
  /usr/bin/nginx
  ├── ELF header
  ├── .text section (machine code)
  ├── .data section (initialized globals)
  ├── .bss section (uninitialized globals)
  └── .rodata section (string literals, constants)

Process (running instance):
  PID 1234 (nginx worker 1)
  ├── Virtual address space (unique)
  ├── Heap (unique, grows at runtime)
  ├── Stack (unique per thread)
  ├── Open file descriptors (unique)
  ├── CPU state (registers, PC)
  └── Shares TEXT segment with PID 1235 (nginx worker 2)
       (same code, different data — copy-on-write)

Multiple processes from same program:
  nginx master  PID 1000
  nginx worker  PID 1001  ─┐
  nginx worker  PID 1002  ─┤── All share same physical TEXT pages
  nginx worker  PID 1003  ─┘   Each has own heap/stack/data
```

### Virtual Memory Benefits

| Benefit | How It Works |
|---------|-------------|
| **Isolation** | Each process has own page table; can't access others' memory |
| **More memory than RAM** | Pages can be swapped to disk (swap space) |
| **Shared libraries** | libc.so mapped into many processes, one physical copy |
| **Memory-mapped files** | Files mapped directly into address space (mmap) |
| **Copy-on-write (fork)** | Parent and child share pages until one writes |
| **Guard pages** | Unmapped pages detect stack overflow |

---

## 12. Pages and Frames — Virtual Memory Mapping

### Pages and Frames

```
Virtual Memory (process's view):        Physical RAM (actual hardware):
┌──────────────────────────┐            ┌──────────────────────────┐
│  Page 0   (4KB)          │            │  Frame 0   (4KB)         │
│  Page 1   (4KB)          │            │  Frame 1   (4KB)         │
│  Page 2   (4KB)          │            │  Frame 2   (4KB)         │
│  Page 3   (4KB)          │            │  Frame 3   (4KB)         │
│  ...                     │            │  ...                     │
│  Page N   (4KB)          │            │  Frame M   (4KB)         │
└──────────────────────────┘            └──────────────────────────┘

Page = chunk of virtual address space (4KB)
Frame = chunk of physical RAM (4KB, same size as page)
Page Table = the mapping: Page N → Frame M
```

### How a Virtual Address Is Translated

On x86-64, a 64-bit virtual address is split into parts:

```
Virtual Address (48 bits used on x86-64):
┌──────────┬──────────┬──────────┬──────────┬──────────────┐
│  PML4    │  PDPT    │   PD     │   PT     │   Offset     │
│  index   │  index   │  index   │  index   │  (12 bits)   │
│  (9 bits)│  (9 bits)│  (9 bits)│  (9 bits)│              │
└──────────┴──────────┴──────────┴──────────┴──────────────┘

4-level page table walk:
  CR3 → PML4 table → PDPT table → PD table → PT table → Physical Frame
  
  Physical Address = Frame base address + Offset (12 bits = 4KB)
```

### The TLB — Hardware Cache for Page Translations

Walking 4 levels of page tables = 4 RAM accesses = ~400ns. This would be catastrophic for every memory access. The **TLB** caches recent translations:

```
CPU needs virtual address 0x7fff1234:

Check TLB:
  ├── TLB HIT  → physical address found in ~1ns → done
  └── TLB MISS → walk 4-level page table (~400ns)
                  → cache result in TLB
                  → future accesses to same page: TLB hit

TLB size: typically 64-1024 entries (very small!)
TLB is flushed on context switch (different process = different page table)
→ This is why context switches are expensive
```

### Page Fault Handling (Full Flow)

```
Process accesses virtual address VA:

1. MMU checks TLB → miss
2. MMU walks page table → present bit = 0 (page not in RAM)
3. CPU raises #PF (page fault exception)
4. CPU saves state, switches to Ring 0
5. Kernel's page fault handler runs:

   Is VA in process's valid memory regions (VMAs)?
   ├── NO  → invalid access → send SIGSEGV → process dies
   └── YES → valid access, page just not loaded

       Why is page missing?
       ├── First access (demand paging)
       │     → Allocate free frame
       │     → Zero-fill it (security: don't leak other process's data)
       │     → Update page table (present bit = 1)
       │     → Return to user code (retry the instruction)
       │
       ├── Page was swapped to disk
       │     → Find page in swap space
       │     → Read from disk into free frame (~100µs-10ms)
       │     → Update page table
       │     → Return to user code
       │
       └── Copy-on-write (after fork)
             → Allocate new frame
             → Copy parent's page content
             → Update child's page table to point to new frame
             → Return to user code
```

### How the OS Maintains the Mapping

```
Per-process data structures:

1. Page Table (hardware-walked by MMU):
   Virtual Page Number → Physical Frame Number + flags
   Flags: Present, Writable, User/Kernel, Accessed, Dirty, NX (no-execute)

2. VMA (Virtual Memory Areas) — Linux vm_area_struct:
   Describes regions of the address space:
   ┌─────────────────────────────────────────────────────┐
   │  VMA 1: 0x400000 - 0x401000  [text]  r-x  file-backed│
   │  VMA 2: 0x600000 - 0x601000  [data]  rw-  file-backed│
   │  VMA 3: 0x800000 - 0x900000  [heap]  rw-  anonymous  │
   │  VMA 4: 0x7fff0000-0x7ffff000[stack] rw-  anonymous  │
   └─────────────────────────────────────────────────────┘

3. Physical Frame Allocator (buddy system):
   Tracks which physical frames are free/used
   Allocates frames on page fault
   Reclaims frames when process exits
```

### Swap Space

When RAM is full, the OS can **evict** pages to disk (swap):

```
RAM full → OS picks victim page (LRU algorithm)
         → Writes victim page to swap partition/file on disk
         → Marks page table entry as "not present, in swap"
         → Frees the physical frame for another use

Later, if process accesses swapped-out page:
  → Page fault
  → OS reads page back from swap into a free frame
  → Updates page table
  → Process resumes

Swap is ~1000x slower than RAM.
Excessive swapping = "thrashing" = system appears frozen.
```

---

## 13. Garbage Collection

### What Is GC?

Garbage Collection is **automatic memory management** — the runtime automatically frees heap memory that is no longer reachable by the program. You don't call `free()`.

```
Without GC (C/C++):
  ptr = malloc(100);   // allocate
  use(ptr);
  free(ptr);           // YOU must free — forget this = memory leak
  use(ptr);            // use after free = undefined behavior / crash

With GC (Go, Java, Python):
  p := make([]byte, 100)  // allocate
  use(p)
  // p goes out of scope — GC will free it automatically
  // No use-after-free possible (GC won't free reachable memory)
```

### When Is GC Triggered in Go?

```
Go's GC triggers when:

1. Heap size doubles (GOGC=100 default):
   - If heap was 50MB after last GC
   - GC triggers when heap reaches 100MB
   - GOGC=200 → triggers at 150MB (less frequent, more memory used)
   - GOGC=50  → triggers at 75MB (more frequent, less memory used)

2. runtime.GC() called explicitly (rare, for benchmarks)

3. Memory limit exceeded (Go 1.19+):
   - debug.SetMemoryLimit(512 * 1024 * 1024)
   - GC runs aggressively to stay under limit

4. Two minutes since last GC (background trigger)
```

### Go's GC Algorithm — Tricolor Mark-and-Sweep

```
Three colors for heap objects:

WHITE = not yet visited (candidate for collection)
GREY  = reachable, but children not yet scanned
BLACK = reachable, children scanned (safe — will NOT be collected)

Phase 1: MARK (concurrent with your code)
  Start: all objects WHITE
  
  Root objects (globals, stack vars, registers) → mark GREY
  
  For each GREY object:
    Mark it BLACK
    Mark all objects it points to as GREY
  
  Repeat until no GREY objects remain.
  
  Result: BLACK = reachable, WHITE = garbage

Phase 2: SWEEP (concurrent)
  Walk heap, free all WHITE objects
  
Phase 3: STW (Stop The World) — very brief
  Two short STW pauses:
  1. Start of mark phase (~100µs): enable write barrier
  2. End of mark phase (~100µs): finalize
  
  Go's GC target: < 1ms STW pauses
```

### Stack vs Heap — Memory Management Perspective

```
STACK:
  ┌─────────────────────────────────────────────────────┐
  │  Allocation: O(1) — just decrement stack pointer    │
  │  Deallocation: O(1) — just increment stack pointer  │
  │  No GC needed — freed automatically on return       │
  │  No fragmentation                                   │
  │  Size: 2KB (goroutine) to 8MB (OS thread)           │
  │  Lifetime: tied to function scope                   │
  └─────────────────────────────────────────────────────┘

HEAP:
  ┌─────────────────────────────────────────────────────┐
  │  Allocation: slower — find free block, update lists │
  │  Deallocation: GC (Go/Java) or manual (C/C++)       │
  │  GC required — objects may be referenced anywhere   │
  │  Fragmentation possible over time                   │
  │  Size: limited by available RAM                     │
  │  Lifetime: until GC collects or explicit free       │
  └─────────────────────────────────────────────────────┘
```

### GC Tuning in Go

```go
import (
    "os"
    "runtime/debug"
)

// Increase GC target — less frequent GC, more memory used
// Default: 100 (GC when heap doubles)
os.Setenv("GOGC", "200")
// or:
debug.SetGCPercent(200)

// Set soft memory limit (Go 1.19+)
// GC will run more aggressively to stay under this
debug.SetMemoryLimit(512 * 1024 * 1024)  // 512MB

// Force GC (useful in benchmarks, not production)
runtime.GC()

// Disable GC for batch processing
debug.SetGCPercent(-1)
// ... do batch work ...
debug.SetGCPercent(100)  // re-enable

// Monitor GC
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("GC runs: %d, Pause total: %v ms
",
    m.NumGC, m.PauseTotalNs/1e6)
```

---

## 14. Metrics and Units

### CPU Metrics

| Metric | Unit | What It Measures | Good Range |
|--------|------|-----------------|------------|
| Clock speed | GHz (billion cycles/sec) | Raw CPU speed | 3–5 GHz typical |
| Cores | Count | Parallel execution units | 4–128 |
| Utilization | % (0–100 per core) | How busy the CPU is | <70% sustained |
| Load average | Dimensionless (1/5/15 min) | Runnable + waiting processes | < num_cores |
| Context switches | switches/sec | How often OS switches tasks | <10,000/sec |
| IPC | Instructions per cycle | CPU efficiency | 1–4 typical |
| Cache hit rate | % | L1/L2/L3 hit ratio | >95% L1 |
| CPU wait (iowait) | % | Time waiting for I/O | <5% |

```bash
# Linux tools
top                    # overall CPU %
mpstat -P ALL 1        # per-core utilization
perf stat ./myapp      # IPC, cache misses, branch mispredictions
vmstat 1               # context switches (cs column)
```

### Memory Metrics

| Metric | Unit | What It Measures |
|--------|------|-----------------|
| Total RAM | GB | Physical memory installed |
| RSS | MB/GB | Resident Set Size — actual RAM used by process |
| VSZ | MB/GB | Virtual Size — total virtual address space |
| Heap size | MB/GB | Dynamic allocation area |
| GC pause | ms/µs | Stop-the-world pause duration |
| Page faults | faults/sec | Minor (page in) or major (disk read) |
| Swap used | MB/GB | RAM overflow on disk (bad if high) |

```
Memory size units:
  1 KB  = 1,024 bytes
  1 MB  = 1,024 KB  = 1,048,576 bytes
  1 GB  = 1,024 MB  = 1,073,741,824 bytes
  1 TB  = 1,024 GB
```

### Disk Metrics

| Metric | Unit | What It Measures | Good Range |
|--------|------|-----------------|------------|
| IOPS | operations/sec | Random read/write operations | SSD: 100K-1M |
| Throughput | MB/s or GB/s | Sequential read/write speed | SSD: 500MB-7GB/s |
| Latency | ms or µs | Time for one I/O operation | NVMe: <100µs |
| Queue depth | count | Pending I/O requests | <32 |
| Utilization | % | How busy the disk is | <70% |

```bash
iostat -x 1            # IOPS, throughput, utilization, await
iotop                  # per-process disk I/O
```

### Network I/O Metrics

| Metric | Unit | What It Measures | Good Range |
|--------|------|-----------------|------------|
| Bandwidth | Mbps / Gbps | Data transfer rate | 1–100 Gbps |
| Latency | ms / µs | Round-trip time (RTT) | LAN: <1ms |
| Packets/sec | pps | Packet processing rate | 1M+ pps for 10G |
| Connections | count | Active TCP connections | Depends on app |
| Error rate | % | Dropped/errored packets | <0.01% |
| Retransmits | count/sec | TCP retransmissions | Near 0 |

```bash
iftop                  # per-connection bandwidth
netstat -s             # network statistics
ss -s                  # socket summary
ping host              # RTT latency
iperf3 -c host         # bandwidth test
```

---

## 15. Go's Efficiency — Goroutines and the Runtime

### The Problem with OS Threads

Traditional languages (Java, C++, Python threads) map 1 thread → 1 OS thread:

```
Java/C++ Thread Model:
  Thread 1 → OS Thread 1 → CPU Core
  Thread 2 → OS Thread 2 → CPU Core
  Thread 3 → OS Thread 3 → CPU Core
  ...
  Thread 10,000 → OS Thread 10,000 → ???

Problems:
  - Each OS thread: 1–8 MB stack (fixed)
  - 10,000 threads = 10–80 GB RAM just for stacks!
  - OS scheduler must manage 10,000 threads → overhead
  - Context switch between OS threads: ~1-10µs
  - Creating an OS thread: ~10ms
```

### Go's M:N Threading Model

Go uses a **M:N model**: M goroutines multiplexed onto N OS threads.

```
Go Runtime Scheduler:

Goroutines (G):          OS Threads (M):      CPU Cores (P):
  G1  G2  G3  G4           M1   M2              Core0  Core1
  G5  G6  G7  G8           M3   M4              Core2  Core3
  G9  G10 ...
  (millions possible)    (GOMAXPROCS threads)  (physical cores)

  G = Goroutine (user-space thread, 2KB stack)
  M = Machine (OS thread, managed by Go runtime)
  P = Processor (logical CPU, holds run queue of goroutines)
```

### The GMP Model in Detail

```
┌─────────────────────────────────────────────────────────────────┐
│                        Go Runtime                               │
│                                                                 │
│  P0 (Processor 0)              P1 (Processor 1)                │
│  ┌─────────────────────┐       ┌─────────────────────┐         │
│  │  Local Run Queue    │       │  Local Run Queue    │         │
│  │  [G3][G7][G12]      │       │  [G5][G9]           │         │
│  │         │           │       │         │           │         │
│  │         ▼           │       │         ▼           │         │
│  │  M1 (OS Thread)     │       │  M2 (OS Thread)     │         │
│  │  Currently: G1      │       │  Currently: G4      │         │
│  └─────────────────────┘       └─────────────────────┘         │
│                                                                 │
│  Global Run Queue: [G2][G6][G8][G10][G11]                      │
│                                                                 │
│  Idle M pool: [M3][M4]  (OS threads waiting for work)          │
└─────────────────────────────────────────────────────────────────┘
```

### Goroutine vs OS Thread

| Property | Goroutine | OS Thread |
|----------|-----------|-----------|
| Stack size | 2 KB (grows dynamically) | 1–8 MB (fixed) |
| Creation time | ~300ns | ~10ms |
| Context switch | ~100ns (userspace) | ~1-10µs (kernel) |
| Max practical count | Millions | Thousands |
| Scheduling | Go runtime (userspace) | OS kernel |
| Blocked on I/O | Goroutine parked, M reused | Thread blocked (wasted) |
| Memory per 10K | ~20 MB | ~10 GB |

### How Go Handles Blocking I/O

This is Go's killer feature. When a goroutine blocks on I/O, Go doesn't block the OS thread:

```
G1 calls: conn.Read(buf)  (network read, may block)

Go runtime:
  1. Registers the file descriptor with epoll/kqueue (OS async I/O)
  2. Parks G1 (moves it off the run queue)
  3. M1 picks up G2 from the run queue and runs it
  4. When data arrives (epoll event), G1 is put back on run queue
  5. M1 (or another M) picks up G1 and resumes it

Result: M1 is NEVER blocked. It always has work to do.
One OS thread can handle thousands of concurrent I/O operations.
```

### Work Stealing

When P0's local run queue is empty, it **steals** goroutines from P1:

```
P0 run queue: empty
P1 run queue: [G5][G6][G7][G8]

P0 steals half of P1's queue:
P0 run queue: [G7][G8]
P1 run queue: [G5][G6]

This keeps all CPU cores busy automatically.
No manual load balancing needed.
```

### Goroutine Scheduling — Cooperative + Preemptive

```
Go 1.13 and earlier: Cooperative only
  - Goroutine yields at: function calls, channel ops, syscalls, runtime.Gosched()
  - Problem: tight loops could starve other goroutines

Go 1.14+: Preemptive (signal-based)
  - Runtime sends SIGURG signal to OS thread every 10ms
  - Signal handler checks if goroutine has been running too long
  - If yes: preempt it (save state, put back on queue)
  - Tight loops can now be preempted
```

### Escape Analysis — Stack vs Heap in Go

```go
// Compiler decides at compile time where variables live

// STACK — does not escape
func multiply(a, b int) int {
    result := a * b    // result lives only in this function → stack
    return result      // return value, not pointer → stack is fine
}

// HEAP — escapes to heap
func newSlice(n int) []int {
    s := make([]int, n)  // s returned to caller → escapes to heap
    return s
}

// HEAP — escapes via interface
var sink interface{}
func escape(v int) {
    sink = v  // v stored in interface{} globally → escapes to heap
}

// Check with:
// go build -gcflags="-m -m" ./...
```

### Go vs Other Languages

| Feature | Go | Java | Python | Node.js |
|---------|-----|------|--------|---------|
| Concurrency model | Goroutines (M:N) | Threads (1:1) | GIL (1 thread) | Event loop (1 thread) |
| Parallelism | Yes (GOMAXPROCS) | Yes | No (GIL) | No (single thread) |
| Stack per unit | 2 KB (dynamic) | 512KB–1MB | ~8MB | N/A |
| GC | Concurrent, low-pause | Generational, configurable | Reference counting + cyclic | V8 generational |
| Blocking I/O | Non-blocking (netpoller) | Blocks thread | Blocks thread | Non-blocking (libuv) |
| Max concurrent units | Millions | ~10,000 | ~10,000 | ~10,000 (callbacks) |
| Memory per 10K units | ~20 MB | ~5–10 GB | ~5–10 GB | N/A |

### Practical Go Concurrency Patterns

```go
// Pattern 1: Worker pool (CPU-bound work)
func workerPool(jobs []Job, numWorkers int) []Result {
    jobCh := make(chan Job, len(jobs))
    resultCh := make(chan Result, len(jobs))
    
    // Start workers (one per CPU core for CPU-bound)
    for i := 0; i < numWorkers; i++ {
        go func() {
            for job := range jobCh {
                resultCh <- process(job)
            }
        }()
    }
    
    // Send jobs
    for _, job := range jobs {
        jobCh <- job
    }
    close(jobCh)
    
    // Collect results
    results := make([]Result, len(jobs))
    for i := range results {
        results[i] = <-resultCh
    }
    return results
}

// Pattern 2: Fan-out / Fan-in (I/O-bound work)
func fanOut(urls []string) []Response {
    ch := make(chan Response, len(urls))
    for _, url := range urls {
        go func(u string) {
            resp, _ := http.Get(u)  // goroutine parks while waiting
            ch <- Response{URL: u, Resp: resp}
        }(url)
    }
    results := make([]Response, len(urls))
    for i := range results {
        results[i] = <-ch
    }
    return results
}

// Pattern 3: Context cancellation
func withTimeout(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()  // ALWAYS defer cancel
    
    select {
    case result := <-doWork(ctx):
        return result
    case <-ctx.Done():
        return ctx.Err()  // deadline exceeded or cancelled
    }
}
```

---

## 16. Virtualisation vs Containerisation

### Virtualisation

Virtualisation creates a **complete virtual computer** in software. A **hypervisor** sits between hardware and VMs, emulating hardware for each VM.

```
┌─────────────────────────────────────────────────────────────────┐
│                      Physical Server                            │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │     VM 1     │  │     VM 2     │  │     VM 3     │         │
│  │  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │         │
│  │  │  App   │  │  │  │  App   │  │  │  │  App   │  │         │
│  │  ├────────┤  │  │  ├────────┤  │  │  ├────────┤  │         │
│  │  │  OS    │  │  │  │  OS    │  │  │  │  OS    │  │         │
│  │  │(Linux) │  │  │  │(Windows│  │  │  │(Ubuntu)│  │         │
│  │  ├────────┤  │  │  ├────────┤  │  │  ├────────┤  │         │
│  │  │Virtual │  │  │  │Virtual │  │  │  │Virtual │  │         │
│  │  │Hardware│  │  │  │Hardware│  │  │  │Hardware│  │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Hypervisor (VMware, KVM, Hyper-V)          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Physical Hardware                          │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

**Type 1 Hypervisor** (bare-metal): Runs directly on hardware. No host OS.
Examples: VMware ESXi, Microsoft Hyper-V, KVM (Linux kernel module)

**Type 2 Hypervisor** (hosted): Runs on top of a host OS.
Examples: VirtualBox, VMware Workstation, Parallels

### Containerisation (Docker)

Containers share the **host OS kernel**. They use Linux kernel features to isolate processes:

```
┌─────────────────────────────────────────────────────────────────┐
│                      Physical Server                            │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Container 1  │  │ Container 2  │  │ Container 3  │         │
│  │  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │         │
│  │  │  App   │  │  │  │  App   │  │  │  │  App   │  │         │
│  │  ├────────┤  │  │  ├────────┤  │  │  ├────────┤  │         │
│  │  │  Libs  │  │  │  │  Libs  │  │  │  │  Libs  │  │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         Container Runtime (Docker, containerd)          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Host OS Kernel (shared!)                   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Physical Hardware                          │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Docker Internals — Linux Kernel Features

Docker is not magic. It uses two Linux kernel features:

#### 1. Namespaces — Isolation

| Namespace | Isolates | What Container Sees |
|-----------|---------|---------------------|
| **PID** | Process IDs | Container's PID 1 is actually PID 5432 on host |
| **NET** | Network interfaces | Own eth0, IP address, routing table |
| **MNT** | File system mounts | Own root filesystem (from image) |
| **UTS** | Hostname | Own hostname |
| **IPC** | Inter-process communication | Own shared memory, semaphores |
| **USER** | User/group IDs | Root in container = non-root on host |
| **CGROUP** | cgroup hierarchy | Own resource limits view |

#### 2. cgroups — Resource Limits

cgroups (control groups) limit what resources a container can use:

```
Container resource limits:
  CPU:    max 2 cores (cpu.quota)
  Memory: max 512MB (memory.limit_in_bytes)
  Disk:   max 100 MB/s I/O (blkio.throttle)
  Network: max 100 Mbps (via tc/iptables)

Without cgroups: one container could use all CPU/RAM and starve others.
With cgroups: each container is bounded.
```

### VM vs Container Comparison

| Property | Virtual Machine | Container |
|----------|----------------|-----------|
| Isolation | Full (separate kernel) | Process-level (shared kernel) |
| OS | Each VM has own OS | Shares host kernel |
| Startup time | Minutes | Milliseconds |
| Size | GB (full OS image) | MB (just app + libs) |
| Overhead | 5–20% CPU/RAM overhead | <1% overhead |
| Security | Strong (hardware boundary) | Weaker (kernel escape possible) |
| Portability | Less (OS-specific) | High (same image anywhere) |
| Use case | Different OSes, strong isolation | Microservices, CI/CD, scaling |

### When to Use What

```
Use VMs when:
  - You need to run different operating systems (Windows + Linux)
  - Strong security isolation is required (multi-tenant, untrusted code)
  - Compliance requires full OS isolation
  - Running legacy applications that need specific OS versions

Use Containers when:
  - Microservices architecture
  - Fast scaling (start in milliseconds)
  - CI/CD pipelines
  - Consistent dev/prod environments
  - High density (many services on one host)

Use Both (common in cloud):
  - VMs for the base infrastructure (EC2 instances)
  - Containers (Kubernetes pods) running inside VMs
  - Best of both: VM isolation + container density
```

---

## 17. Architect-Level Q&A

---

### Q1: What is mechanical empathy and give a real-world example?

**A**: Mechanical empathy means writing code that works with the hardware's natural behavior rather than against it. 

Real example: A struct in Go with fields accessed by multiple goroutines. If two frequently-written fields share a 64-byte cache line, every write from one core invalidates the other core's cache line (false sharing), causing 10x slowdown. An engineer with mechanical empathy pads the struct to put each hot field on its own cache line.

---

### Q2: What happens end-to-end when you call `os.ReadFile("data.txt")` in Go?

**A**:
1. Go calls `open()` syscall → CPU traps to kernel (Ring 3 → Ring 0)
2. Kernel opens file, returns file descriptor
3. Go calls `read()` syscall with fd and buffer pointer
4. Kernel's VFS layer routes to the file system driver (ext4/APFS)
5. File system checks page cache — if page is cached, copy to user buffer (fast path)
6. If not cached: disk I/O request sent to block device driver
7. DMA transfer: disk controller writes data directly to kernel buffer in RAM
8. Disk controller sends interrupt to CPU when done
9. Kernel copies data from kernel buffer to user process buffer (copy_to_user)
10. `read()` returns, CPU switches back to Ring 3
11. Go runtime returns the data to your code

---

### Q3: What is the difference between a process and a thread?

**A**: A process is an isolated running instance of a program with its own virtual address space, PID, file descriptors, and resources. A thread is a lightweight execution unit within a process — threads share the process's address space (heap, globals, code) but each has its own stack, program counter, and registers.

Key difference: a crash in one process doesn't affect others. A crash in one thread kills the entire process. Threads communicate via shared memory (fast but needs synchronization). Processes communicate via IPC (pipes, sockets — slower but safer).

---

### Q4: How does the OS prevent one process from reading another process's memory?

**A**: Through virtual memory and page tables. Each process has its own page table, and the CPU's MMU uses the current process's page table for all address translations. Process A's virtual address 0x1000 maps to physical frame 500. Process B's virtual address 0x1000 maps to physical frame 800. They are completely different physical locations.

If Process A tries to access a virtual address that maps to Process B's physical frame — it can't, because Process A's page table simply doesn't have an entry for that physical frame. Any attempt to access an unmapped address causes a page fault → SIGSEGV → process A is killed.

---

### Q5: Explain a page fault end-to-end.

**A**: 
1. Process accesses virtual address VA
2. MMU checks TLB — miss
3. MMU walks page table — present bit = 0 (page not in RAM)
4. CPU raises #PF exception, switches to Ring 0
5. Kernel's page fault handler runs
6. Kernel checks if VA is in a valid VMA (virtual memory area)
   - If not: SIGSEGV (segfault)
   - If yes: determine why page is missing
7. If first access: allocate free frame, zero it, update page table
8. If swapped to disk: read from swap (~10ms), update page table
9. Return to user code, retry the faulting instruction
10. This time: TLB miss → page table walk → present bit = 1 → physical address found → data returned

---

### Q6: What is false sharing and how do you avoid it?

**A**: False sharing occurs when two CPU cores write to different variables that happen to reside on the same 64-byte cache line. Even though they're writing to different memory locations, the cache coherence protocol (MESI) treats the entire cache line as the unit of ownership. Each write from one core invalidates the other core's copy of the cache line, forcing a re-fetch from L3 or RAM.

Fix: pad structs so hot fields are on separate cache lines:
```go
type Counter struct {
    value int64
    _     [56]byte  // pad to 64 bytes
}
```

---

### Q7: Why is disk I/O so much slower than RAM?

**A**: 
- **HDD**: Mechanical arm must physically move to the right track (seek time ~5ms) and wait for the disk to rotate to the right sector (rotational latency ~4ms). Total: ~10ms per random access.
- **SSD**: No moving parts, but NAND flash has inherent read latency (~100µs) and write latency (~1ms). Erase cycles are slow.
- **RAM**: Electrons in capacitors/transistors. Access is purely electronic, no mechanical movement. ~100ns.

The ratio: RAM is 1,000x faster than SSD, 100,000x faster than HDD.

---

### Q8: What is DMA and why does it exist?

**A**: DMA (Direct Memory Access) allows hardware devices (disk, NIC, GPU) to transfer data directly to/from RAM without involving the CPU for each byte. Without DMA, the CPU would have to execute a loop reading each byte from the device and writing it to RAM — wasting CPU cycles on data movement.

With DMA: CPU sets up the transfer (source, destination, size), then continues doing useful work. The DMA controller handles the transfer. When done, it sends an interrupt to the CPU. This is why disk and network I/O can be non-blocking — the CPU is free while data moves.

---

### Q9: How does context switching affect performance?

**A**: Each context switch costs ~1-10µs (save registers, switch page tables, flush TLB, load new registers). The TLB flush is the most expensive part — the new process's memory accesses will all be TLB misses initially, causing many RAM accesses to rebuild the TLB.

At 1,000 context switches/second: ~1-10ms/second overhead (negligible).
At 100,000 context switches/second: ~100ms-1s/second overhead (significant).

In Go: goroutine context switches are ~100ns (userspace, no TLB flush) vs OS thread switches at ~1-10µs. This is why Go can have millions of goroutines efficiently.

---

### Q10: What is the kernel and why does it need a separate protection ring?

**A**: The kernel is the core OS software that manages all hardware resources. It needs Ring 0 (privileged mode) because:
1. It must access hardware directly (write to device registers, configure MMU)
2. It must be able to read/write any physical memory (to manage page tables)
3. It must execute privileged instructions (LGDT, LIDT, HLT, etc.)

User code runs in Ring 3 (restricted) because:
1. A buggy user program should not be able to corrupt the kernel or other processes
2. User code should not be able to bypass security (read other processes' memory)
3. Hardware access must be mediated by the kernel (so it can enforce permissions)

The hardware enforces this: if Ring 3 code tries to execute a privileged instruction, the CPU raises a General Protection Fault → kernel kills the process.

---

### Q11: How does Go's GC differ from Java's GC?

**A**:

| | Go GC | Java GC (G1/ZGC) |
|--|-------|-----------------|
| Algorithm | Tricolor concurrent mark-sweep | Generational (young/old gen) |
| STW pauses | <1ms target | 1-100ms (G1), <1ms (ZGC) |
| Generations | No (non-generational) | Yes (young gen collected more often) |
| Tuning | GOGC, SetMemoryLimit | -Xmx, -Xms, GC algorithm flags |
| Write barrier | Yes (during mark phase) | Yes |
| Compaction | No (no memory compaction) | Yes (G1 compacts) |

Go's GC is simpler (no generations) but has very low pause times. Java's generational GC is more throughput-efficient for long-running apps with lots of short-lived objects.

---

### Q12: What is the TLB and what happens on a TLB miss?

**A**: The TLB (Translation Lookaside Buffer) is a small, fast hardware cache inside the CPU that stores recent virtual-to-physical address translations. Without it, every memory access would require walking the 4-level page table (4 RAM accesses = ~400ns).

On a TLB miss:
1. CPU's hardware page table walker reads CR3 (page table base)
2. Walks PML4 → PDPT → PD → PT (4 RAM accesses)
3. Gets physical frame number
4. Caches the translation in TLB
5. Completes the memory access

TLB miss cost: ~400ns (4 RAM accesses). TLB hit: ~1ns. TLB is flushed on context switch (different process = different page table), which is why context switches are expensive.

---

### Q13: How does Docker achieve isolation without a hypervisor?

**A**: Docker uses two Linux kernel features:

1. **Namespaces**: Each container gets its own view of system resources. PID namespace: container's processes have their own PID numbering (PID 1 in container = PID 5432 on host). NET namespace: own network interfaces and IP. MNT namespace: own root filesystem. UTS: own hostname.

2. **cgroups**: Limit how much of each resource a container can use. CPU quota, memory limit, disk I/O throttle, network bandwidth.

The key difference from VMs: all containers share the same Linux kernel. There's no hardware emulation. A container is just a process (or group of processes) with restricted visibility and resource limits. This is why containers start in milliseconds and have near-zero overhead.

---

### Q14: What is NUMA and why does it matter for performance?

**A**: NUMA (Non-Uniform Memory Access) is a memory architecture used in multi-socket servers. Each CPU socket has its own local RAM. Accessing local RAM is fast (~100ns). Accessing RAM attached to another socket (remote NUMA node) is slower (~200-300ns).

```
Socket 0:  CPU 0-15  ←→  RAM Bank 0 (local, fast)
                     ←→  RAM Bank 1 (remote, slow, via QPI/UPI)
Socket 1:  CPU 16-31 ←→  RAM Bank 1 (local, fast)
```

For architects: pin latency-sensitive processes to a single NUMA node. In Go, the OS scheduler tries to keep goroutines on the same NUMA node, but you can use `numactl` to pin processes. Kubernetes has NUMA-aware scheduling for high-performance workloads.

---

### Q15: How does the Go scheduler implement work stealing?

**A**: Each P (logical processor) has a local run queue of goroutines. When P0's queue is empty:
1. P0 checks the global run queue
2. If global queue is empty, P0 randomly picks another P (say P1)
3. P0 steals half of P1's local run queue
4. P0 runs the stolen goroutines

This keeps all CPU cores busy without a central scheduler bottleneck. The randomness prevents all idle Ps from stealing from the same busy P simultaneously.

---

### Q16: What is a system call and what is its overhead?

**A**: A system call is the mechanism by which user-space code requests a service from the kernel. Examples: `read()`, `write()`, `open()`, `socket()`, `mmap()`, `fork()`.

Overhead breakdown:
- Save user registers: ~10ns
- Switch to kernel mode (SYSCALL instruction): ~10ns
- Kernel validates arguments: ~10-100ns
- Kernel performs the operation: varies (0ns for getpid, ms for disk read)
- Switch back to user mode (SYSRET): ~10ns
- Restore user registers: ~10ns

Total overhead (excluding the actual work): ~100-1000ns per syscall.

This is why batching is important: one `write()` with 4KB is much better than 4096 `write()` calls with 1 byte each.

---

### Q17: Explain the difference between virtual memory and physical memory.

**A**: 
- **Physical memory**: The actual RAM chips. Finite (e.g., 16GB). Shared by all processes. Addressed by physical addresses (0 to 16GB-1).
- **Virtual memory**: Each process's private view of memory. Appears to be a large, contiguous address space (0 to 2^48 on x86-64). Not real — it's an abstraction.

The MMU translates virtual → physical on every memory access using the page table. Benefits:
- Isolation: each process has its own virtual space, can't access others
- More memory than RAM: pages can be swapped to disk
- Shared libraries: one physical copy mapped into many virtual spaces
- Simplified programming: every process thinks it has the full address space

---

### Q18: What is swap and when does the OS use it?

**A**: Swap is disk space used as overflow when RAM is full. The OS evicts (writes) cold pages from RAM to the swap partition/file, freeing physical frames for active use. When the evicted page is needed again, a page fault occurs and the OS reads it back from swap.

Swap is ~1000x slower than RAM. Heavy swap usage ("thrashing") makes the system appear frozen — the CPU spends most of its time handling page faults and waiting for disk I/O.

Signs of swap thrashing: high `%wa` in top, high `si`/`so` in vmstat, system unresponsive.

Fix: add more RAM, reduce memory usage, set memory limits on containers/processes.

---

### Q19: How does escape analysis work in Go?

**A**: The Go compiler performs escape analysis at compile time to determine whether a variable can live on the stack (fast, auto-freed) or must be allocated on the heap (GC-managed).

A variable "escapes" to the heap when:
- A pointer to it is returned from the function
- It's stored in a global variable or interface
- It's sent to a channel
- It's captured by a closure that outlives the function
- It's too large for the stack

```go
// Does NOT escape — stays on stack
func add(a, b int) int {
    result := a + b
    return result  // return value, not pointer
}

// ESCAPES to heap — pointer returned
func newUser() *User {
    u := User{Name: "Alice"}
    return &u  // pointer to local var → must escape to heap
}
```

Check with: `go build -gcflags="-m" ./...`

---

### Q20: Why can Go have millions of goroutines but Java can only have thousands of threads?

**A**: 
- **Java thread**: 1:1 with OS thread. Each OS thread has a fixed stack of 512KB–1MB. 10,000 threads = 5–10 GB RAM just for stacks. OS scheduler must manage 10,000 kernel threads.
- **Go goroutine**: M:N model. Each goroutine starts with a 2KB stack (grows dynamically). 1,000,000 goroutines = ~2 GB RAM for stacks. Go's userspace scheduler manages goroutines on top of a small number of OS threads (GOMAXPROCS, typically = num CPU cores).

Additionally, when a goroutine blocks on I/O, Go parks it and reuses the OS thread for another goroutine. In Java, a blocked thread holds its OS thread (wasted). This is why Go can handle 100,000 concurrent HTTP connections with GOMAXPROCS=8 OS threads, while Java would need 100,000 OS threads.

---

### Q21: What happens when you call `make([]int, 1000000)` in Go?

**A**:
1. Compiler's escape analysis determines if the slice escapes (usually yes, if returned or stored)
2. Go runtime's `mallocgc()` is called
3. Runtime checks its size class — large allocations (>32KB) go directly to the heap via `mmap()` syscall
4. For smaller allocations: runtime checks the P's local mcache (thread-local cache of memory spans)
5. If mcache has a free span of the right size class: allocate from it (no lock needed)
6. If not: get a span from the mcentral (needs lock)
7. If mcentral is empty: get pages from mheap (needs lock, may call mmap())
8. Zero the memory (Go guarantees zero-initialized memory)
9. Return pointer to the caller

The slice header (pointer + length + capacity) goes on the stack. The backing array goes on the heap.

---

### Q22: What is the relationship between GOMAXPROCS, goroutines, and CPU cores?

**A**: 
- **GOMAXPROCS**: Number of P (logical processors) in the Go scheduler. Defaults to `runtime.NumCPU()`. Controls how many goroutines can run in parallel.
- **Goroutines**: User-space threads. Can be millions. Scheduled by Go runtime onto Ps.
- **CPU cores**: Physical execution units. One goroutine runs on one core at a time.

```
GOMAXPROCS=4, 8 CPU cores, 1,000,000 goroutines:
  - 4 goroutines run in parallel (one per P)
  - 999,996 goroutines wait in run queues
  - 4 OS threads are active (one per P)
  - 4 CPU cores are used

GOMAXPROCS=8:
  - 8 goroutines run in parallel
  - Uses all 8 CPU cores
  - Better for CPU-bound work
```

For I/O-bound work: GOMAXPROCS doesn't matter much (goroutines spend most time parked waiting for I/O). For CPU-bound work: set GOMAXPROCS = num CPU cores.

---

### Q23: How does the OS scheduler decide which process to run next?

**A**: Linux uses CFS (Completely Fair Scheduler). It tracks `vruntime` (virtual runtime) for each runnable process — how much CPU time it has received, weighted by priority. The process with the lowest vruntime runs next.

CFS uses a red-black tree sorted by vruntime. The leftmost node (lowest vruntime) is always the next to run. This gives O(log n) scheduling decisions.

Priority (nice value -20 to +19) affects how fast vruntime accumulates: lower nice = vruntime grows slower = gets more CPU time.

Real-time processes (SCHED_FIFO, SCHED_RR) bypass CFS and always preempt normal processes.

---

### Q24: What is copy-on-write (COW) and how does fork() use it?

**A**: Copy-on-write is an optimization where two processes share the same physical pages until one of them writes to a page. Only then is a private copy made.

When `fork()` is called:
1. Child process is created with a copy of the parent's page table
2. Both parent and child page table entries point to the SAME physical frames
3. Both entries are marked read-only
4. When either process writes to a page: page fault → kernel allocates new frame, copies the page, updates the writing process's page table to point to the new frame, marks it writable
5. The other process still points to the original frame

This makes `fork()` very fast (no copying of potentially GBs of memory). Only pages that are actually written get copied. This is why `fork()` + `exec()` is efficient for spawning new processes.

---

---

## Quick Reference — Numbers Every Architect Must Know

```
Latency Cheat Sheet:
  L1 cache hit          ~1   ns
  L2 cache hit          ~4   ns
  L3 cache hit          ~10  ns
  RAM access            ~100 ns
  NVMe SSD              ~100 µs   (1,000x RAM)
  SATA SSD              ~500 µs
  HDD seek              ~10  ms   (100,000x RAM)
  Network (same DC)     ~500 µs
  Network (cross-DC)    ~10  ms
  Network (cross-ocean) ~150 ms

Size Cheat Sheet:
  L1 cache              32–64 KB per core
  L2 cache              256 KB – 1 MB per core
  L3 cache              8–64 MB shared
  Cache line            64 bytes
  Page size             4 KB (typical)
  Goroutine stack       2 KB (initial)
  OS thread stack       1–8 MB (fixed)
  TLB entries           64–1024

Throughput Cheat Sheet:
  L1 cache bandwidth    ~1 TB/s
  RAM bandwidth         50–100 GB/s
  NVMe SSD              3–7 GB/s
  SATA SSD              500 MB/s
  HDD                   100–200 MB/s
  10 GbE NIC            1.25 GB/s
  100 GbE NIC           12.5 GB/s

Cost Cheat Sheet:
  Context switch        ~1–10 µs
  System call           ~100–1000 ns
  Goroutine switch      ~100 ns
  Goroutine creation    ~300 ns
  OS thread creation    ~10 ms
  Process fork          ~1 ms
  Page fault (minor)    ~1 µs
  Page fault (major)    ~10 ms (disk read)
```

---

## Summary — Key Mental Models for Architects

1. **The memory hierarchy is everything.** Design data structures and access patterns to maximize cache hits. Sequential > random. Small > large. Reuse > allocate.

2. **Processes are isolated; threads share.** Use processes for fault isolation. Use threads/goroutines for concurrency within a service.

3. **The kernel is the gatekeeper.** Every hardware access goes through the kernel. System calls have overhead. Batch I/O operations.

4. **Virtual memory gives each process its own world.** The MMU + page tables make isolation and overcommit possible. Page faults are the cost of this abstraction.

5. **Go's goroutines are cheap because they're not OS threads.** 2KB stack, userspace scheduling, non-blocking I/O via netpoller. This is why Go scales to millions of concurrent connections.

6. **Containers are processes with restrictions.** Namespaces for isolation, cgroups for limits. No hypervisor overhead. VMs for strong isolation, containers for density.

7. **GC trades CPU for safety.** Go's GC is concurrent and low-pause. Reduce GC pressure by minimizing heap allocations (use stack, pools, value types).

8. **Mechanical empathy = performance.** Know your cache line size (64 bytes). Know your latency numbers. Write code that the hardware can execute efficiently.
