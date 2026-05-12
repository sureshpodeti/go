# Interview Questions & Answers

## Overview

This document contains comprehensive interview questions and answers for software architects focusing on the Compute Layer. Questions are organized by difficulty level and topic.

## Question Categories

- **Foundational**: Basic concepts every architect should know
- **Intermediate**: Practical application and design decisions
- **Advanced**: Deep technical knowledge and complex scenarios
- **Scenario-Based**: Real-world problem-solving

---

## CPU Architecture & Performance

### Foundational Questions

**Q1: What is the difference between CPU cores and threads?**

**Answer:**
- **CPU Cores**: Physical processing units within a CPU. Each core can independently execute instructions.
- **Threads**: Virtual cores created through technologies like Intel Hyperthreading or AMD SMT (Simultaneous Multithreading).

```
Physical Core with Hyperthreading:
┌─────────────────────────────────┐
│      Physical Core              │
│  ┌──────────┐  ┌──────────┐    │
│  │ Thread 1 │  │ Thread 2 │    │
│  └──────────┘  └──────────┘    │
│         Shared Resources        │
└─────────────────────────────────┘
```

**Key Points:**
- 1 physical core with hyperthreading = 2 logical cores
- Threads share execution resources (ALU, FPU, cache)
- Performance gain: 20-30% (not 2x)
- Best for workloads with different resource needs

**Follow-up**: When would you disable hyperthreading?
- Latency-sensitive applications
- Security concerns (side-channel attacks)
- CPU-bound workloads with no I/O wait

---

**Q2: Explain the CPU cache hierarchy and its impact on performance.**

**Answer:**
```
Cache Hierarchy:
L1: 32-64 KB, ~1 ns, per-core
L2: 256-512 KB, ~3 ns, per-core
L3: 8-64 MB, ~10 ns, shared
RAM: 8-512 GB, ~100 ns, system-wide

Performance Impact:
L1 hit: 1 ns
L2 hit: 3 ns (3x slower)
L3 hit: 10 ns (10x slower)
RAM hit: 100 ns (100x slower)
```

**Architectural Implications:**
1. **Data Locality**: Keep frequently accessed data close together
2. **Cache-Friendly Algorithms**: Sequential access > random access
3. **Working Set Size**: Keep hot data under L3 cache size
4. **False Sharing**: Avoid in multi-threaded applications

**Example:**
```go
// Bad: Cache-unfriendly
for col := 0; col < cols; col++ {
    for row := 0; row < rows; row++ {
        matrix[row][col] = value  // Column-major access
    }
}

// Good: Cache-friendly
for row := 0; row < rows; row++ {
    for col := 0; col < cols; col++ {
        matrix[row][col] = value  // Row-major access
    }
}
```

---

### Intermediate Questions

**Q3: How would you design a system to handle both CPU-bound and I/O-bound workloads efficiently?**

**Answer:**

**Strategy: Workload Separation**

```
Architecture:
┌─────────────────────────────────────┐
│   Load Balancer                     │
└──────────┬──────────────────────────┘
           │
     ┌─────┴─────┐
     │           │
     ▼           ▼
┌─────────┐ ┌─────────┐
│CPU-Bound│ │I/O-Bound│
│ Workers │ │ Workers │
│         │ │         │
│• High   │ │• Moderate│
│  CPU    │ │  CPU    │
│• Low    │ │• High   │
│  count  │ │  count  │
│• Compute│ │• Async  │
│  optimized│ │  I/O    │
└─────────┘ └─────────┘
```

**Implementation Details:**

1. **CPU-Bound Pool:**
   - Fewer, more powerful cores
   - Thread pool size = CPU cores
   - Synchronous processing
   - Example: Video encoding, data compression

2. **I/O-Bound Pool:**
   - More instances with moderate CPU
   - Thread pool size = CPU cores × 2-4
   - Async/await patterns
   - Example: API calls, database queries

3. **Routing Logic:**
```python
def route_request(request):
    if request.type == "compute_intensive":
        return cpu_bound_pool.submit(request)
    else:
        return io_bound_pool.submit_async(request)
```

**Monitoring:**
- CPU utilization per pool
- Queue depth
- Response time
- Adjust pool sizes based on metrics

---

**Q4: Explain NUMA and its implications for application design.**

**Answer:**

**NUMA (Non-Uniform Memory Access):**

```
NUMA Architecture:
┌──────────────┐      ┌──────────────┐
│   Node 0     │      │   Node 1     │
│  CPU 0-15    │      │  CPU 16-31   │
│  128 GB RAM  │      │  128 GB RAM  │
└──────┬───────┘      └──────┬───────┘
       │                     │
       └────Interconnect─────┘
       
Local Access:  100 ns
Remote Access: 200 ns (2x slower!)
```

**Implications:**

1. **Memory Allocation:**
   - Allocate memory on same NUMA node as CPU
   - Use `numa_alloc_onnode()` or `numactl`

2. **Process Pinning:**
```bash
# Pin process to NUMA node 0
numactl --cpunodebind=0 --membind=0 ./myapp
```

3. **Application Design:**
   - Partition data by NUMA node
   - Avoid cross-node memory access
   - Use NUMA-aware data structures

4. **Database Optimization:**
```
PostgreSQL Example:
- Pin connections to NUMA nodes
- Partition buffer pool by node
- Use node-local storage
```

**Monitoring:**
```bash
numastat
numactl --hardware
```

**When NUMA Matters:**
- Multi-socket servers (2+ CPUs)
- Memory-intensive workloads
- High-performance databases
- Real-time systems

---

### Advanced Questions

**Q5: Design a high-performance computing cluster for scientific simulations. What hardware and software considerations would you make?**

**Answer:**

**Hardware Architecture:**

```
HPC Cluster Design:

Compute Nodes (100x):
┌─────────────────────────────────┐
│ 2x AMD EPYC 7763 (128 cores)    │
│ 512 GB RAM (DDR4-3200)          │
│ 2x 1.92TB NVMe SSD (local)      │
│ 2x 100 Gbps InfiniBand          │
└─────────────────────────────────┘

Storage Nodes (10x):
┌─────────────────────────────────┐
│ Parallel File System (Lustre)   │
│ 1 PB total capacity             │
│ 200 GB/s aggregate throughput   │
└─────────────────────────────────┘

Network:
┌─────────────────────────────────┐
│ InfiniBand HDR (200 Gbps)       │
│ Fat-tree topology               │
│ <1 μs latency                   │
└─────────────────────────────────┘
```

**Software Stack:**

1. **Job Scheduler:**
   - Slurm or PBS Pro
   - Resource allocation
   - Queue management

2. **MPI Implementation:**
   - OpenMPI or Intel MPI
   - RDMA over InfiniBand
   - Optimized collectives

3. **Compilers:**
   - Intel oneAPI (best performance)
   - GCC with optimization flags
   - Profile-guided optimization

4. **Libraries:**
   - Intel MKL (math)
   - FFTW (FFT)
   - HDF5 (I/O)

**Optimization Strategies:**

1. **CPU Optimization:**
```bash
# Compiler flags
-O3 -march=native -mtune=native -mavx512f
```

2. **Memory Optimization:**
   - NUMA-aware allocation
   - Huge pages (2MB/1GB)
   - Memory pinning

3. **I/O Optimization:**
   - Parallel I/O (MPI-IO)
   - Collective buffering
   - Stripe files across OSTs

4. **Network Optimization:**
   - RDMA for MPI
   - GPUDirect for GPU communication
   - Topology-aware placement

**Monitoring:**
- Ganglia or Prometheus
- Job profiling (Intel VTune)
- Network monitoring (InfiniBand counters)

---

## Memory Management

### Foundational Questions

**Q6: Explain virtual memory and its benefits.**

**Answer:**

**Virtual Memory Concept:**

```
Process View (Virtual):
┌─────────────────────┐
│ Process A: 0x0-0xFFFF│
│ Process B: 0x0-0xFFFF│
│ Process C: 0x0-0xFFFF│
└──────────┬──────────┘
           │ MMU Translation
           ▼
Physical Memory:
┌─────────────────────┐
│ [A][B][C][A][B]     │
└─────────────────────┘
```

**Benefits:**

1. **Isolation:**
   - Processes cannot access each other's memory
   - Security and stability

2. **Simplification:**
   - Each process sees linear address space
   - No need to manage physical addresses

3. **Flexibility:**
   - Physical memory can be fragmented
   - Processes can be larger than RAM

4. **Sharing:**
   - Shared libraries mapped once
   - Copy-on-write for fork()

5. **Overcommitment:**
   - Total virtual > physical memory
   - Swap to disk when needed

**Implementation:**
- Page tables for translation
- TLB for caching translations
- Page faults for demand paging

---

**Q7: What is the difference between stack and heap memory?**

**Answer:**

```
Memory Layout:
┌─────────────────────┐
│ Stack (grows down)  │ ← Fast, automatic
│         ▼           │
│                     │
│     (Free Space)    │
│                     │
│         ▲           │
│ Heap (grows up)     │ ← Slower, manual
├─────────────────────┤
│ Data Segment        │
├─────────────────────┤
│ Code Segment        │
└─────────────────────┘
```

**Comparison:**

| Aspect | Stack | Heap |
|--------|-------|------|
| **Allocation** | Automatic | Manual (malloc/new) |
| **Deallocation** | Automatic | Manual (free/delete) |
| **Speed** | Very fast (pointer bump) | Slower (find free block) |
| **Size** | Limited (1-8 MB) | Large (GB) |
| **Lifetime** | Function scope | Until freed |
| **Fragmentation** | None | Possible |
| **Thread Safety** | Per-thread | Shared (needs sync) |

**When to Use:**

**Stack:**
- Local variables
- Function parameters
- Small, short-lived data
- Known size at compile time

**Heap:**
- Dynamic size
- Long lifetime
- Large objects
- Shared across functions

**Example:**
```c
void example() {
    int stack_var = 10;           // Stack
    int* heap_var = malloc(sizeof(int));  // Heap
    *heap_var = 20;
    free(heap_var);               // Must free!
}  // stack_var automatically freed
```

---

### Intermediate Questions

**Q8: How would you diagnose and fix a memory leak in a production system?**

**Answer:**

**Diagnosis Process:**

1. **Identify the Leak:**
```bash
# Monitor memory growth
ps aux | grep myapp
top -p <pid>

# Check for OOM kills
dmesg | grep -i "out of memory"
```

2. **Heap Profiling:**
```bash
# Using Valgrind
valgrind --leak-check=full --show-leak-kinds=all ./myapp

# Using Go pprof
import _ "net/http/pprof"
go tool pprof http://localhost:6060/debug/pprof/heap

# Using jemalloc
export MALLOC_CONF="prof:true,prof_leak:true"
```

3. **Analyze Allocation Patterns:**
```
Heap Profile:
┌─────────────────────────────────────┐
│ Total: 2.5 GB                       │
│                                     │
│ Top Allocations:                    │
│ 1. UserCache: 1.2 GB (48%)          │
│ 2. ConnectionPool: 800 MB (32%)     │
│ 3. RequestBuffer: 300 MB (12%)      │
│ 4. Other: 200 MB (8%)               │
└─────────────────────────────────────┘
```

**Common Causes:**

1. **Unreleased Resources:**
```go
// Bad
func processFile(filename string) {
    file, _ := os.Open(filename)
    // Missing: defer file.Close()
    // Process file...
}

// Good
func processFile(filename string) {
    file, _ := os.Open(filename)
    defer file.Close()
    // Process file...
}
```

2. **Growing Caches:**
```go
// Bad: Unbounded cache
cache := make(map[string][]byte)
func getData(key string) []byte {
    if val, ok := cache[key]; ok {
        return val
    }
    val := fetchData(key)
    cache[key] = val  // Never evicted!
    return val
}

// Good: LRU cache with size limit
cache := lru.New(1000)  // Max 1000 entries
```

3. **Circular References:**
```python
# Bad
class Node:
    def __init__(self):
        self.parent = None
        self.children = []

# Creates circular reference
parent.children.append(child)
child.parent = parent  # Leak if not using weak references

# Good
import weakref
child.parent = weakref.ref(parent)
```

**Fix Strategies:**

1. **Immediate Mitigation:**
   - Restart affected instances
   - Reduce traffic
   - Scale horizontally

2. **Short-term Fix:**
   - Add memory limits
   - Implement cache eviction
   - Add monitoring/alerts

3. **Long-term Solution:**
   - Fix root cause
   - Add automated tests
   - Implement leak detection in CI/CD

**Prevention:**
- Code reviews
- Static analysis tools
- Memory profiling in testing
- Automated leak detection

---

### Advanced Questions

**Q9: Design a memory management strategy for a high-throughput, low-latency trading system.**

**Answer:**

**Requirements:**
- Latency: <10 μs (99th percentile)
- Throughput: 1M+ messages/second
- No GC pauses
- Predictable performance

**Memory Architecture:**

```
Memory Layout:
┌─────────────────────────────────────┐
│ Huge Pages (2 MB pages)             │
│ • Reduces TLB misses                │
│ • Locked in RAM (no swap)           │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Pre-allocated Object Pools          │
│ • Order objects                     │
│ • Market data objects               │
│ • Network buffers                   │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Ring Buffers (Lock-free)            │
│ • Inter-thread communication        │
│ • No dynamic allocation             │
└─────────────────────────────────────┘
```

**Implementation:**

1. **Object Pooling:**
```cpp
template<typename T>
class ObjectPool {
private:
    std::vector<T*> pool;
    std::atomic<size_t> next{0};
    
public:
    ObjectPool(size_t size) {
        pool.reserve(size);
        for (size_t i = 0; i < size; ++i) {
            pool.push_back(new T());
        }
    }
    
    T* acquire() {
        size_t idx = next.fetch_add(1) % pool.size();
        return pool[idx];
    }
    
    void release(T* obj) {
        obj->reset();  // Clear state
        // Object stays in pool
    }
};
```

2. **Memory Pinning:**
```cpp
// Pin memory to NUMA node
void* mem = numa_alloc_onnode(size, node_id);

// Lock pages in RAM
mlock(mem, size);

// Use huge pages
mmap(..., MAP_HUGETLB | MAP_LOCKED, ...);
```

3. **Cache-Line Alignment:**
```cpp
struct alignas(64) Order {  // 64-byte cache line
    uint64_t order_id;
    uint32_t price;
    uint32_t quantity;
    // ... pad to 64 bytes
};
```

4. **Lock-Free Data Structures:**
```cpp
// SPSC ring buffer
template<typename T, size_t Size>
class RingBuffer {
private:
    alignas(64) std::atomic<size_t> write_pos{0};
    alignas(64) std::atomic<size_t> read_pos{0};
    std::array<T, Size> buffer;
    
public:
    bool push(const T& item) {
        size_t write = write_pos.load(std::memory_order_relaxed);
        size_t next = (write + 1) % Size;
        if (next == read_pos.load(std::memory_order_acquire))
            return false;  // Full
        buffer[write] = item;
        write_pos.store(next, std::memory_order_release);
        return true;
    }
};
```

**Monitoring:**
```cpp
struct MemoryMetrics {
    uint64_t pool_hits;
    uint64_t pool_misses;
    uint64_t cache_line_bounces;
    uint64_t numa_remote_accesses;
};
```

**Key Principles:**
- Zero dynamic allocation in hot path
- NUMA-aware allocation
- Cache-line alignment
- Lock-free algorithms
- Huge pages for TLB efficiency

---

## Storage Systems

### Foundational Questions

**Q10: Explain the difference between IOPS and throughput.**

**Answer:**

**IOPS (Input/Output Operations Per Second):**
- Number of read/write operations per second
- Important for random access workloads
- Measured in operations/second

**Throughput (Bandwidth):**
- Amount of data transferred per second
- Important for sequential access workloads
- Measured in MB/s or GB/s

```
Example Comparison:

Scenario 1: Random 4KB reads
┌────────────────────────────────┐
│ IOPS: 100,000                  │
│ Throughput: 390 MB/s           │
│ (100,000 × 4KB)                │
└────────────────────────────────┘

Scenario 2: Sequential 1MB reads
┌────────────────────────────────┐
│ IOPS: 3,500                    │
│ Throughput: 3,500 MB/s         │
│ (3,500 × 1MB)                  │
└────────────────────────────────┘
```

**When Each Matters:**

**IOPS-Critical:**
- Databases (OLTP)
- Virtual machines
- Boot volumes
- Random access patterns

**Throughput-Critical:**
- Video streaming
- Big data analytics
- Backups
- Sequential access patterns

**Hardware Comparison:**
```
HDD (7200 RPM):
• IOPS: 80-160
• Throughput: 100-200 MB/s

SATA SSD:
• IOPS: 90,000-100,000
• Throughput: 500-600 MB/s

NVMe SSD:
• IOPS: 500,000-1,000,000
• Throughput: 3,000-7,000 MB/s
```

---



## Operating Systems

### Foundational Questions

**Q11: What is the difference between a process and a thread?**

**Answer:**

**Process:**
- Independent execution unit
- Own address space
- Own resources (file descriptors, memory)
- Heavy context switch
- Isolated from other processes

**Thread:**
- Lightweight execution unit within a process
- Shared address space
- Shared resources
- Light context switch
- Can communicate easily

```
Process vs Thread:

Process:
┌─────────────────────────┐
│ Process A               │
│ ┌─────────────────────┐ │
│ │ Code, Data, Heap    │ │
│ └─────────────────────┘ │
│ ┌─────┐ ┌─────┐        │
│ │Stack│ │Stack│        │
│ │Thr 1│ │Thr 2│        │
│ └─────┘ └─────┘        │
└─────────────────────────┘

Context Switch Cost:
Process: 1-10 μs (save/restore + TLB flush)
Thread:  0.1-1 μs (save/restore only)
```

**When to Use:**

**Processes:**
- Strong isolation needed
- Different programs
- Fault tolerance (one crash doesn't affect others)
- Security boundaries

**Threads:**
- Shared state
- Performance critical
- Same program, parallel tasks
- Lower overhead

---

**Q12: Explain the different CPU scheduling algorithms.**

**Answer:**

**1. First-Come, First-Served (FCFS):**
```
Processes: P1(24ms), P2(3ms), P3(3ms)
Timeline: [P1: 24ms][P2: 3ms][P3: 3ms]
Avg Wait: (0 + 24 + 27) / 3 = 17ms

Pros: Simple, fair
Cons: Convoy effect (short jobs wait for long jobs)
```

**2. Shortest Job First (SJF):**
```
Processes: P1(24ms), P2(3ms), P3(3ms)
Timeline: [P2: 3ms][P3: 3ms][P1: 24ms]
Avg Wait: (0 + 3 + 6) / 3 = 3ms

Pros: Optimal average waiting time
Cons: Need to know execution time, starvation
```

**3. Round Robin (RR):**
```
Processes: P1(24ms), P2(3ms), P3(3ms)
Time Quantum: 4ms
Timeline: [P1:4][P2:3][P3:3][P1:4][P1:4]...

Pros: Fair, good for interactive systems
Cons: Context switch overhead
```

**4. Priority Scheduling:**
```
Processes: P1(Pri=3), P2(Pri=1), P3(Pri=2)
Timeline: [P2][P3][P1]

Pros: Important tasks first
Cons: Starvation of low-priority tasks
Solution: Priority aging
```

**5. Multi-Level Feedback Queue:**
```
Queue 0 (Highest): Time quantum = 8ms
Queue 1 (Medium):  Time quantum = 16ms
Queue 2 (Lowest):  FCFS

New processes start in Queue 0
If not finished, move to Queue 1
If not finished, move to Queue 2

Pros: Favors short, interactive processes
Cons: Complex to implement
```

---

### Intermediate Questions

**Q13: How would you debug a deadlock in a production system?**

**Answer:**

**Detection Steps:**

1. **Identify Symptoms:**
```bash
# Check for hung processes
ps aux | grep D  # D = uninterruptible sleep

# Check system load
uptime
top

# Check for lock contention
# Java
jstack <pid> | grep -A 10 "BLOCKED"

# Go
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Linux
cat /proc/<pid>/stack
```

2. **Analyze Lock Dependencies:**
```
Deadlock Example:

Thread 1:          Thread 2:
lock(A)            lock(B)
  ...                ...
  lock(B) ← WAIT   lock(A) ← WAIT
```

3. **Use Debugging Tools:**
```bash
# GDB
gdb -p <pid>
(gdb) thread apply all bt

# strace
strace -p <pid>

# perf
perf record -p <pid> -g
perf report
```

**Prevention Strategies:**

1. **Lock Ordering:**
```go
// Always acquire locks in same order
func transfer(from, to *Account, amount int) {
    // Sort accounts by ID to ensure consistent order
    first, second := from, to
    if from.ID > to.ID {
        first, second = to, from
    }
    
    first.Lock()
    defer first.Unlock()
    
    second.Lock()
    defer second.Unlock()
    
    // Transfer logic
}
```

2. **Timeouts:**
```go
func acquireLocks(locks []*sync.Mutex) bool {
    timeout := time.After(5 * time.Second)
    
    for _, lock := range locks {
        select {
        case <-timeout:
            // Release acquired locks
            return false
        default:
            lock.Lock()
        }
    }
    return true
}
```

3. **Deadlock Detection:**
```python
class DeadlockDetector:
    def __init__(self):
        self.wait_graph = {}  # resource -> waiting threads
        
    def check_cycle(self):
        # Detect cycles in wait graph
        visited = set()
        rec_stack = set()
        
        for node in self.wait_graph:
            if self.has_cycle(node, visited, rec_stack):
                return True
        return False
```

**Resolution:**
- Restart affected processes
- Release locks manually (if possible)
- Implement automatic deadlock detection and recovery

---

### Advanced Questions

**Q14: Design a custom memory allocator for a specific workload.**

**Answer:**

**Scenario: Web Server with Many Small Allocations**

**Requirements:**
- Frequent allocations (10K-100K/sec)
- Small objects (16-256 bytes)
- Short lifetime (milliseconds)
- Low fragmentation

**Design:**

```cpp
class SlabAllocator {
private:
    struct Slab {
        void* memory;
        size_t object_size;
        size_t capacity;
        std::vector<void*> free_list;
        
        Slab(size_t obj_size, size_t count) 
            : object_size(obj_size), capacity(count) {
            // Allocate large chunk
            memory = mmap(nullptr, 
                         obj_size * count,
                         PROT_READ | PROT_WRITE,
                         MAP_PRIVATE | MAP_ANONYMOUS,
                         -1, 0);
            
            // Initialize free list
            for (size_t i = 0; i < count; ++i) {
                void* obj = static_cast<char*>(memory) + 
                           (i * obj_size);
                free_list.push_back(obj);
            }
        }
    };
    
    // Size classes: 16, 32, 64, 128, 256 bytes
    std::array<Slab*, 5> slabs;
    
public:
    SlabAllocator() {
        slabs[0] = new Slab(16, 10000);
        slabs[1] = new Slab(32, 10000);
        slabs[2] = new Slab(64, 5000);
        slabs[3] = new Slab(128, 2000);
        slabs[4] = new Slab(256, 1000);
    }
    
    void* allocate(size_t size) {
        // Find appropriate slab
        Slab* slab = find_slab(size);
        if (!slab || slab->free_list.empty()) {
            // Fallback to system allocator
            return malloc(size);
        }
        
        // Pop from free list
        void* obj = slab->free_list.back();
        slab->free_list.pop_back();
        return obj;
    }
    
    void deallocate(void* ptr, size_t size) {
        Slab* slab = find_slab(size);
        if (!slab || !belongs_to_slab(slab, ptr)) {
            free(ptr);
            return;
        }
        
        // Push to free list
        slab->free_list.push_back(ptr);
    }
};
```

**Benefits:**
- O(1) allocation/deallocation
- No fragmentation within size class
- Cache-friendly (objects close together)
- Thread-local slabs for scalability

**Monitoring:**
```cpp
struct AllocatorStats {
    uint64_t allocations;
    uint64_t deallocations;
    uint64_t slab_hits;
    uint64_t slab_misses;
    size_t memory_used;
    size_t memory_wasted;  // fragmentation
};
```

---

## Virtualization & Containers

### Foundational Questions

**Q15: What is the difference between containers and virtual machines?**

**Answer:**

```
Virtual Machines:
┌────────┬────────┐
│  App A │  App B │
├────────┼────────┤
│Guest OS│Guest OS│ ← Full OS per VM
│ (GB)   │ (GB)   │
├────────┴────────┤
│   Hypervisor    │
├─────────────────┤
│    Host OS      │
├─────────────────┤
│   Hardware      │
└─────────────────┘

Containers:
┌────────┬────────┐
│  App A │  App B │
├────────┼────────┤
│ Libs   │ Libs   │ ← Shared kernel
├────────┴────────┤
│  Container RT   │
├─────────────────┤
│    Host OS      │
├─────────────────┤
│   Hardware      │
└─────────────────┘
```

**Comparison:**

| Aspect | VMs | Containers |
|--------|-----|------------|
| **Startup** | Minutes | Seconds |
| **Size** | GBs | MBs |
| **Performance** | Overhead 5-15% | Near-native |
| **Isolation** | Strong (hardware) | Moderate (OS) |
| **Density** | 10-50 per host | 100-1000 per host |
| **OS** | Different kernels | Same kernel |
| **Use Case** | Strong isolation | Microservices |

**When to Use:**

**VMs:**
- Different OS needed
- Strong isolation required
- Legacy applications
- Compliance requirements

**Containers:**
- Microservices
- CI/CD
- Fast scaling
- Cloud-native apps

---

**Q16: Explain Docker image layers and their benefits.**

**Answer:**

**Layer Structure:**

```
Docker Image Layers:

┌─────────────────────────────────┐
│ Container Layer (R/W)           │ ← Changes
├─────────────────────────────────┤
│ Layer 4: COPY app.js            │ ← App code
├─────────────────────────────────┤
│ Layer 3: RUN npm install        │ ← Dependencies
├─────────────────────────────────┤
│ Layer 2: COPY package.json      │ ← Package file
├─────────────────────────────────┤
│ Layer 1: FROM node:18           │ ← Base image
└─────────────────────────────────┘

Each layer is read-only except container layer
```

**Benefits:**

1. **Sharing:**
```
Image A:                Image B:
┌─────────────┐        ┌─────────────┐
│ App A code  │        │ App B code  │
├─────────────┤        ├─────────────┤
│ node:18     │◄───────┤ node:18     │
└─────────────┘        └─────────────┘
         Shared layer (stored once)
```

2. **Caching:**
```dockerfile
# Dockerfile optimization
FROM node:18

# Layer 1: Dependencies (changes rarely)
COPY package*.json ./
RUN npm install

# Layer 2: Code (changes frequently)
COPY . .

# If only code changes, npm install is cached!
```

3. **Incremental Updates:**
```
Version 1.0:          Version 1.1:
┌─────────────┐      ┌─────────────┐
│ App v1.0    │      │ App v1.1    │ ← Only this layer
├─────────────┤      ├─────────────┤    needs to be
│ Dependencies│◄─────┤ Dependencies│    downloaded
├─────────────┤      ├─────────────┤
│ Base OS     │◄─────┤ Base OS     │
└─────────────┘      └─────────────┘
```

**Best Practices:**

1. **Order Matters:**
```dockerfile
# Bad: Code changes invalidate all layers
COPY . .
RUN npm install

# Good: Dependencies cached separately
COPY package*.json ./
RUN npm install
COPY . .
```

2. **Minimize Layers:**
```dockerfile
# Bad: Many layers
RUN apt-get update
RUN apt-get install -y package1
RUN apt-get install -y package2

# Good: Single layer
RUN apt-get update && \
    apt-get install -y package1 package2 && \
    rm -rf /var/lib/apt/lists/*
```

3. **Multi-Stage Builds:**
```dockerfile
# Build stage
FROM node:18 AS builder
COPY . .
RUN npm install && npm run build

# Production stage
FROM node:18-alpine
COPY --from=builder /app/dist ./dist
CMD ["node", "dist/index.js"]

# Final image only contains production artifacts
```

---

### Intermediate Questions

**Q17: How would you optimize container startup time?**

**Answer:**

**Optimization Strategies:**

1. **Image Size Reduction:**
```dockerfile
# Before: 1.2 GB
FROM ubuntu:20.04
RUN apt-get update && apt-get install -y python3

# After: 50 MB
FROM python:3.9-alpine
```

2. **Layer Caching:**
```dockerfile
# Optimize layer order
FROM node:18-alpine

# 1. System dependencies (rarely change)
RUN apk add --no-cache git

# 2. App dependencies (change occasionally)
COPY package*.json ./
RUN npm ci --only=production

# 3. App code (changes frequently)
COPY . .
```

3. **Parallel Pulls:**
```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    image: myapp:latest
    pull_policy: always
  
  db:
    image: postgres:14
    pull_policy: always

# Pulls images in parallel
```

4. **Image Pre-pulling:**
```bash
# Pre-pull on all nodes
kubectl create daemonset image-puller \
  --image=myapp:latest \
  --restart=Never

# Or use image pull secrets with caching
```

5. **Lazy Loading:**
```
Traditional:
┌─────────────────────────────────┐
│ 1. Pull entire image (30s)      │
│ 2. Extract layers (10s)         │
│ 3. Start container (1s)         │
│ Total: 41s                      │
└─────────────────────────────────┘

With Lazy Loading (e.g., Stargz):
┌─────────────────────────────────┐
│ 1. Pull manifest (1s)           │
│ 2. Start container (1s)         │
│ 3. Pull layers on-demand        │
│ Total: 2s to start!             │
└─────────────────────────────────┘
```

6. **Application Optimization:**
```go
// Reduce initialization time
func main() {
    // Bad: Sequential initialization
    initDatabase()      // 5s
    initCache()         // 3s
    initMessageQueue()  // 2s
    // Total: 10s
    
    // Good: Parallel initialization
    var wg sync.WaitGroup
    wg.Add(3)
    go func() { defer wg.Done(); initDatabase() }()
    go func() { defer wg.Done(); initCache() }()
    go func() { defer wg.Done(); initMessageQueue() }()
    wg.Wait()
    // Total: 5s (limited by slowest)
}
```

**Measurement:**
```bash
# Measure startup time
time docker run --rm myapp:latest

# Breakdown
docker history myapp:latest
docker inspect myapp:latest
```

**Results:**
```
Before Optimization:
• Image size: 1.2 GB
• Pull time: 30s
• Startup time: 45s

After Optimization:
• Image size: 150 MB
• Pull time: 5s
• Startup time: 8s

Improvement: 5.6x faster!
```

---

## Scenario-Based Questions

### Q18: System Design - High-Performance API Gateway

**Scenario:** Design an API gateway that handles 100K requests/second with <10ms latency.

**Answer:**

**Architecture:**

```
┌─────────────────────────────────────────┐
│         Load Balancer (L4)              │
│         (ECMP, Direct Server Return)    │
└────────────┬────────────────────────────┘
             │
     ┌───────┴────────┐
     │                │
     ▼                ▼
┌─────────┐      ┌─────────┐
│Gateway 1│      │Gateway 2│  ... (10+ instances)
└────┬────┘      └────┬────┘
     │                │
     └────────┬───────┘
              │
     ┌────────┴────────┐
     │                 │
     ▼                 ▼
┌─────────┐      ┌─────────┐
│Backend 1│      │Backend 2│
└─────────┘      └─────────┘
```

**Hardware Selection:**

```
Gateway Instances:
┌─────────────────────────────────┐
│ AWS c7g.2xlarge (Graviton3)     │
│ • 8 vCPU                        │
│ • 16 GB RAM                     │
│ • 12.5 Gbps network             │
│ • Cost: $0.29/hour              │
│                                 │
│ Why:                            │
│ • High single-thread perf       │
│ • Cost-effective                │
│ • Good network bandwidth        │
└─────────────────────────────────┘
```

**Software Stack:**

1. **Runtime:**
```
Language: Go or Rust
• Low latency
• Efficient concurrency
• Small memory footprint
```

2. **Networking:**
```go
// Use SO_REUSEPORT for load balancing
listener, _ := net.Listen("tcp", ":8080")
file, _ := listener.(*net.TCPListener).File()
syscall.SetsockoptInt(
    int(file.Fd()),
    syscall.SOL_SOCKET,
    syscall.SO_REUSEPORT,
    1,
)

// Multiple goroutines accept on same port
for i := 0; i < runtime.NumCPU(); i++ {
    go func() {
        for {
            conn, _ := listener.Accept()
            go handleConnection(conn)
        }
    }()
}
```

3. **Connection Pooling:**
```go
// Reuse connections to backends
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        1000,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 5 * time.Second,
}
```

4. **Caching:**
```go
// In-memory cache for hot data
cache := &sync.Map{}

func getFromCache(key string) ([]byte, bool) {
    if val, ok := cache.Load(key); ok {
        return val.([]byte), true
    }
    return nil, false
}
```

**Optimization Techniques:**

1. **Zero-Copy:**
```go
// Use sendfile for static content
file, _ := os.Open("response.json")
defer file.Close()

// Zero-copy to socket
conn.(*net.TCPConn).ReadFrom(file)
```

2. **Memory Pooling:**
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    
    // Use buffer
}
```

3. **CPU Pinning:**
```bash
# Pin process to specific cores
taskset -c 0-7 ./api-gateway
```

**Monitoring:**

```
Key Metrics:
┌─────────────────────────────────┐
│ Requests/sec: 98,543            │
│ P50 Latency: 3.2ms              │
│ P99 Latency: 8.7ms              │
│ P99.9 Latency: 15.3ms           │
│ Error Rate: 0.01%               │
│ CPU Usage: 65%                  │
│ Memory Usage: 8 GB              │
│ Network: 8.5 Gbps               │
└─────────────────────────────────┘
```

**Capacity Planning:**

```
Per Instance:
• Capacity: 10K req/s
• Headroom: 20%
• Effective: 8K req/s

For 100K req/s:
• Instances needed: 100K / 8K = 12.5
• Provision: 15 instances (20% buffer)
• Cost: 15 × $0.29 = $4.35/hour
```

---

## Summary

This comprehensive guide covers:

1. **CPU Architecture**: Cores, threads, cache, NUMA
2. **Memory Management**: Virtual memory, stack/heap, optimization
3. **Storage Systems**: IOPS, throughput, caching
4. **Operating Systems**: Processes, threads, scheduling
5. **Virtualization**: VMs vs containers, Docker, Kubernetes
6. **Capacity Planning**: Hardware selection, scaling strategies
7. **Real-World Scenarios**: High-performance system design

## Study Tips

1. **Understand Fundamentals**: Master the basics before advanced topics
2. **Practice Calculations**: Be comfortable with capacity planning math
3. **Draw Diagrams**: Visualize architectures during interviews
4. **Know Trade-offs**: Every decision has pros and cons
5. **Stay Current**: Cloud technologies evolve rapidly
6. **Hands-On Experience**: Build and measure real systems

## Additional Resources

- Linux Performance Tools: http://www.brendangregg.com/linuxperf.html
- AWS Architecture Center: https://aws.amazon.com/architecture/
- Google SRE Books: https://sre.google/books/
- High Performance Browser Networking: https://hpbn.co/
- Systems Performance by Brendan Gregg

---

**Good luck with your interviews!**
