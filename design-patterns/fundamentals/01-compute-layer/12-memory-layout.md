# Memory Layout in an Operating System

---

## 1. What is a Memory Address?

Memory (RAM) is a giant array of bytes. Every byte has a unique number called an **address**.

```
Address:  0    1    2    3    4    5  ...  65535
         ┌────┬────┬────┬────┬────┬────┬────┐
Memory:  │    │    │    │    │    │    │    │
         └────┴────┴────┴────┴────┴────┴────┘
          ↑                              ↑
       Low address                  High address
       (small number)               (large number)
```

- **Low address** = memory location with a small number (close to 0)
- **High address** = memory location with a large number (close to max)
- On a 64-bit system: low = `0x0000000000000000`, high = `0xFFFFFFFFFFFFFFFF`

---

## 2. The Full Memory Layout of a Process

Every process gets its own **virtual address space**. The OS divides it into segments:

```
High Address  0xFFFFFFFFFFFFFFFF
┌─────────────────────────────────┐
│         Kernel Space            │  ← OS kernel lives here
│   (not accessible to user code) │    your code cannot read/write here
├─────────────────────────────────┤
│             Stack               │  ← grows DOWNWARD ↓
│   local vars, function frames,  │    each function call pushes a frame
│   return addresses, registers   │    each return pops a frame
│               ↓                 │
│                                 │
│           (free gap)            │  ← unused space between stack and heap
│                                 │
│               ↑                 │
│             Heap                │  ← grows UPWARD ↑
│   dynamic allocations:          │    malloc (C), new/make (Go), new (Java)
│   new(T), make([]T, n)          │
├─────────────────────────────────┤
│          BSS Segment            │  ← uninitialized global/static vars
│   var x int  (no value given)   │    OS zeroes this before program starts
├─────────────────────────────────┤
│         Data Segment            │  ← initialized global/static vars
│   var x int = 42                │    exists for entire process lifetime
├─────────────────────────────────┤
│         Text Segment            │  ← compiled machine instructions (code)
│   (read-only)                   │    writing here = segfault
└─────────────────────────────────┘
Low Address   0x0000000000000000
```

---

## 3. Each Segment Explained

### Text Segment (Code)
- Contains compiled machine instructions
- **Read-only** — writing to it causes a segmentation fault
- Shared between multiple instances of the same program (two nginx processes share one copy of code in physical RAM)
- Loaded from the binary on disk at startup

### Data Segment
- Initialized global and static variables
- `var x int = 42` at package level lives here
- Exists for the **entire lifetime** of the process
- Stored in the binary on disk

### BSS Segment *(Block Started by Symbol)*
- Uninitialized global and static variables
- `var x int` (no value assigned) lives here
- OS **zeroes this out** before your program starts — so uninitialized vars are always 0/nil/false
- Takes **no space in the binary** on disk — just a size record saying "reserve N bytes"

### Heap
- Dynamic allocations at runtime
- `new(T)`, `make([]T, n)` in Go; `malloc()` in C; `new Object()` in Java
- Grows **upward** toward higher addresses
- Managed by the runtime (Go GC, Java GC) or manually (C/C++)
- Fragmentation can occur over time

### Stack
- One stack **per thread** (or per goroutine in Go)
- Grows **downward** toward lower addresses
- Holds: local variables, function arguments, return addresses, saved CPU registers
- Each function call pushes a **stack frame**; each return pops it
- Fixed size in most languages: typically **1–8 MB per thread**
- **Go is special**: goroutine stacks start at **2 KB** and grow dynamically

---

## 4. Why Stack Grows Down, Heap Grows Up

They start from opposite ends and grow toward each other — maximizing usable space:

```
Low  0x0000  ┌─────────────┐
             │    Text     │  fixed position
             ├─────────────┤
             │    Data     │  fixed position
             ├─────────────┤
             │    Heap     │  starts low
             │      ↑      │  grows toward HIGH →
             │             │
             │  (free gap) │  unused space in the middle
             │             │
             │      ↓      │  grows toward LOW ←
             │    Stack    │  starts high
High 0xFFFF  └─────────────┘
```

If they meet in the middle:
- Stack hits heap → **stack overflow**
- Heap hits stack → **out of memory**

---

## 5. Virtual Memory vs Physical Memory

This is the key insight. Each process thinks it owns the entire address space — but it doesn't.

```
Process A sees:          Process B sees:
┌──────────────┐         ┌──────────────┐
│ 0x0000-0xFFFF│         │ 0x0000-0xFFFF│
│  (full space)│         │  (full space)│
└──────┬───────┘         └──────┬───────┘
       │         MMU (Memory Management Unit — hardware chip)
       ▼
┌──────────────────────────────────────────┐
│           Physical RAM (actual)          │
│  A's page 1  →  frame 47                │
│  A's page 2  →  frame 12                │
│  B's page 1  →  frame 83                │
│  B's page 2  →  DISK (swapped out)      │
└──────────────────────────────────────────┘
```

- Each process gets its own **virtual address space** — isolated from other processes
- The **MMU** (hardware) translates virtual → physical addresses on every memory access
- The OS maintains a **page table** per process for this mapping
- If a page is not in RAM, the OS triggers a **page fault** and loads it from disk (swap)

---

## 6. Pages and Frames

Memory is divided into fixed-size chunks:

| Term | What It Is | Typical Size |
|------|-----------|-------------|
| **Page** | A chunk of virtual memory | 4 KB |
| **Frame** | A chunk of physical RAM | 4 KB (same as page) |
| **Page table** | OS data structure mapping pages → frames | Per process |
| **TLB** | Hardware cache of recent page translations | In CPU |

```
Virtual Address: 0x12345678
                 ├──────────────┤──────────┤
                 Page number        Offset
                 (index into        (which byte
                  page table)        within the 4KB page)
```

---

## 7. Stack vs Heap — Practical Comparison

| Property | Stack | Heap |
|----------|-------|------|
| Speed | **Extremely fast** (just move stack pointer) | Slower (allocator finds free block) |
| Size | Small (2 KB – 8 MB) | Large (limited by RAM) |
| Lifetime | Tied to function scope — auto freed on return | Until freed / garbage collected |
| Allocation | Automatic (compiler handles it) | Manual (`free`) or GC |
| Fragmentation | **None** | Yes, over time |
| Thread safety | Each thread/goroutine has its own | Shared — needs synchronization |
| Overflow | Stack overflow (too many nested calls) | OOM (too many allocations) |
| Where | High address area | Mid address area |

---

## 8. Go-Specific Memory Layout

Go adds its own layer on top of the OS layout:

```
OS Virtual Address Space
┌──────────────────────────────────────────┐
│  Go Runtime                              │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │  Go Heap (managed by GC)           │  │
│  │  ┌────────┐ ┌────────┐ ┌────────┐  │  │
│  │  │ span   │ │ span   │ │ span   │  │  │  ← heap divided into 8KB spans
│  │  │ (8KB)  │ │ (8KB)  │ │ (8KB)  │  │  │
│  │  └────────┘ └────────┘ └────────┘  │  │
│  └────────────────────────────────────┘  │
│                                          │
│  Goroutine 1 Stack  (starts at 2KB)      │
│  Goroutine 2 Stack  (starts at 2KB)      │
│  Goroutine N Stack  (starts at 2KB)      │
│                                          │
│  Global vars  (Data / BSS segments)      │
│  Code         (Text segment)             │
└──────────────────────────────────────────┘
```

### Key Go Differences from Other Languages

| Feature | Go | Java / C# | C / C++ |
|---------|-----|-----------|---------|
| Goroutine stack start | **2 KB** | 512 KB – 1 MB per thread | 1–8 MB per thread |
| Stack growth | **Dynamic** (copies to larger location) | Fixed | Fixed |
| Heap management | **GC** (tricolor mark-sweep) | GC | Manual (`malloc`/`free`) |
| Escape analysis | **Yes** (compiler decides stack vs heap) | Partial | No |
| Max goroutines | **Millions** (cheap stacks) | Thousands (expensive threads) | Thousands |

### Escape Analysis in Go

The Go compiler decides at **compile time** whether a variable goes on the stack (fast) or heap (GC managed):

```go
// Stack allocated — compiler knows it doesn't escape
func add(a, b int) int {
    result := a + b   // result stays on stack, freed when add() returns
    return result
}

// Heap allocated — escapes because caller holds a pointer
func newUser(name string) *User {
    u := User{Name: name}  // u escapes to heap — pointer returned to caller
    return &u
}
```

```bash
# See escape analysis decisions:
go build -gcflags="-m" ./...

# Output:
# ./main.go:8:2: moved to heap: u
# ./main.go:3:2: result does not escape
```

---

## 9. Common Memory Problems Explained by Layout

| Problem | Segment | Cause | Language |
|---------|---------|-------|----------|
| **Stack overflow** | Stack | Too many nested/recursive calls; stack hits its size limit | All |
| **Segmentation fault** | Text / protected | Writing to read-only memory or dereferencing nil pointer | C/C++, Go |
| **Memory leak** | Heap | Allocated but never freed; GC can't reach it | All |
| **OOM kill** | Heap | Heap grew beyond available RAM; OS kills process | All |
| **Swap thrashing** | Virtual → Disk | Pages evicted to disk, then needed again repeatedly | All |
| **Goroutine leak** | Stack (×N) | Each leaked goroutine holds its 2KB+ stack forever | Go |
| **Buffer overflow** | Stack or Heap | Writing past end of allocated buffer | C/C++ |
| **Use-after-free** | Heap | Accessing memory after it was freed | C/C++ |
| **Double free** | Heap | Calling `free()` twice on same pointer | C/C++ |
| **Heap fragmentation** | Heap | Many alloc/free cycles leave unusable gaps | All |
| **False sharing** | Cache (L1/L2) | Two goroutines write to different vars on same cache line | Go/C++ |

---

## 10. Address Space Layout on 64-bit Linux (Actual Numbers)

```
0xFFFFFFFFFFFFFFFF  ─── top of address space
        │
0xFFFF800000000000  ─── kernel space starts (top 128 TB)
        │               OS kernel, drivers, kernel stacks
        │               user code CANNOT access this
        │
0x00007FFFFFFFFFFF  ─── user space ends
        │
        │   User Space (128 TB available to your program)
        │
        │   Stack      → starts near 0x00007FFFFFFFFFFF, grows down
        │   ...
        │   Heap       → starts after BSS, grows up
        │   BSS        → after Data
        │   Data       → after Text
        │   Text       → starts around 0x0000000000400000
        │
0x0000000000000000  ─── bottom (NULL — never valid to access)
```

---

## 11. One-Line Definitions

| Term | Definition |
|------|-----------|
| **Address** | A number identifying a specific byte in memory |
| **Low address** | Memory location with a small number (close to 0) |
| **High address** | Memory location with a large number (close to max) |
| **Virtual memory** | Each process's private view of the address space (not real RAM) |
| **Physical memory** | Actual RAM chips; shared by all processes via the OS |
| **MMU** | Hardware chip that translates virtual addresses to physical addresses |
| **Page** | 4KB chunk of virtual memory |
| **Frame** | 4KB chunk of physical RAM |
| **Page fault** | CPU can't find a page in RAM; OS loads it from disk |
| **Page table** | OS data structure mapping virtual pages to physical frames |
| **TLB** | CPU cache of recent virtual→physical translations (very fast) |
| **Stack frame** | One function call's worth of data on the stack (locals, args, return addr) |
| **Stack overflow** | Stack grew too large and ran out of space |
| **Heap** | Region for dynamic allocations; managed by GC or manually |
| **BSS** | Segment for uninitialized globals; OS zeroes it at startup |
| **Text segment** | Read-only segment containing compiled machine instructions |
| **Escape analysis** | Compiler decides if a variable lives on stack or heap |
| **Swap** | Disk space used as overflow when RAM is full; 1000x slower than RAM |
| **Segfault** | Process tried to access memory it's not allowed to; OS kills it |
