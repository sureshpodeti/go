# CPU Architecture & Performance

## CPU Fundamentals

### What is a CPU?

The Central Processing Unit (CPU) is the primary component that executes instructions from programs. Modern CPUs are incredibly complex, containing billions of transistors.

## CPU Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Modern CPU Architecture                   │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │              Control Unit (CU)                      │    │
│  │  • Instruction Decoder                              │    │
│  │  • Program Counter                                  │    │
│  │  • Instruction Register                             │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                   │
│                          ▼                                   │
│  ┌────────────────────────────────────────────────────┐    │
│  │       Arithmetic Logic Unit (ALU)                   │    │
│  │  • Integer Operations                               │    │
│  │  • Logical Operations                               │    │
│  │  • Comparison Operations                            │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                   │
│  ┌────────────────────────────────────────────────────┐    │
│  │       Floating Point Unit (FPU)                     │    │
│  │  • Float/Double Operations                          │    │
│  │  • SIMD Instructions                                │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                   │
│  ┌────────────────────────────────────────────────────┐    │
│  │              Registers                              │    │
│  │  • General Purpose (RAX, RBX, etc.)                 │    │
│  │  • Special Purpose (SP, IP, FLAGS)                  │    │
│  └────────────────────────────────────────────────────┘    │
│                          │                                   │
│  ┌────────────────────────────────────────────────────┐    │
│  │              Cache Hierarchy                        │    │
│  │  L1: 32-64 KB  (~1 ns)                             │    │
│  │  L2: 256-512 KB (~3 ns)                            │    │
│  │  L3: 8-64 MB   (~10 ns)                            │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## CPU Performance Factors

### 1. Clock Speed (Frequency)

Measured in GHz (Gigahertz) - billions of cycles per second.

```
Clock Cycle Visualization:

Time ──────────────────────────────────────►

     ┌─┐   ┌─┐   ┌─┐   ┌─┐   ┌─┐   ┌─┐
     │ │   │ │   │ │   │ │   │ │   │ │
─────┘ └───┘ └───┘ └───┘ └───┘ └───┘ └────

     Cycle 1  Cycle 2  Cycle 3  Cycle 4

3.0 GHz = 3,000,000,000 cycles per second
Each cycle = 0.33 nanoseconds
```

**Key Points:**
- Higher clock speed ≠ always better performance
- Power consumption increases with frequency
- Heat generation limits maximum speed
- Modern CPUs: 2.0-5.0 GHz typical

### 2. Cores and Threads

```
Single Core vs Multi-Core:

┌─────────────────┐         ┌─────────────────────────────┐
│  Single Core    │         │      Quad Core              │
│                 │         │                             │
│  ┌───────────┐  │         │  ┌─────┐ ┌─────┐          │
│  │   Core    │  │         │  │Core1│ │Core2│          │
│  │           │  │         │  └─────┘ └─────┘          │
│  └───────────┘  │         │  ┌─────┐ ┌─────┐          │
│                 │         │  │Core3│ │Core4│          │
└─────────────────┘         │  └─────┘ └─────┘          │
                            └─────────────────────────────┘

Hyperthreading (SMT - Simultaneous Multithreading):

┌─────────────────────────────────────────┐
│         Physical Core                    │
│                                          │
│  ┌──────────────┐  ┌──────────────┐    │
│  │  Thread 1    │  │  Thread 2    │    │
│  │  (Virtual)   │  │  (Virtual)   │    │
│  └──────────────┘  └──────────────┘    │
│           │              │              │
│           └──────┬───────┘              │
│                  ▼                      │
│         ┌────────────────┐              │
│         │  Execution     │              │
│         │  Resources     │              │
│         └────────────────┘              │
└─────────────────────────────────────────┘
```

**Core Types:**
- **Physical Cores**: Independent processing units
- **Logical Cores**: Virtual cores via hyperthreading
- **Performance Cores (P-cores)**: High performance, high power
- **Efficiency Cores (E-cores)**: Lower performance, lower power

### 3. Cache Memory

Fast memory close to CPU cores.

```
Cache Hierarchy:

┌──────────────────────────────────────────────────────┐
│                    CPU Package                        │
│                                                       │
│  ┌─────────────┐              ┌─────────────┐       │
│  │   Core 0    │              │   Core 1    │       │
│  │             │              │             │       │
│  │  ┌───────┐  │              │  ┌───────┐  │       │
│  │  │L1 Data│  │              │  │L1 Data│  │       │
│  │  │ 32 KB │  │              │  │ 32 KB │  │       │
│  │  └───────┘  │              │  └───────┘  │       │
│  │  ┌───────┐  │              │  ┌───────┐  │       │
│  │  │L1 Inst│  │              │  │L1 Inst│  │       │
│  │  │ 32 KB │  │              │  │ 32 KB │  │       │
│  │  └───────┘  │              │  └───────┘  │       │
│  │             │              │             │       │
│  │  ┌───────┐  │              │  ┌───────┐  │       │
│  │  │L2 Cache│  │              │  │L2 Cache│  │       │
│  │  │ 512 KB│  │              │  │ 512 KB│  │       │
│  │  └───────┘  │              │  └───────┘  │       │
│  └─────────────┘              └─────────────┘       │
│         │                            │               │
│         └────────────┬───────────────┘               │
│                      ▼                               │
│              ┌───────────────┐                       │
│              │  L3 Cache     │                       │
│              │  (Shared)     │                       │
│              │   16 MB       │                       │
│              └───────────────┘                       │
│                      │                               │
└──────────────────────┼───────────────────────────────┘
                       │
                       ▼
               ┌───────────────┐
               │  Main Memory  │
               │   (RAM)       │
               │   16-512 GB   │
               └───────────────┘
```

**Cache Characteristics:**

| Level | Size | Latency | Bandwidth | Scope |
|-------|------|---------|-----------|-------|
| L1 | 32-64 KB | ~1 ns | ~1 TB/s | Per core |
| L2 | 256-512 KB | ~3 ns | ~500 GB/s | Per core |
| L3 | 8-64 MB | ~10 ns | ~200 GB/s | Shared |
| RAM | 8-512 GB | ~100 ns | ~50 GB/s | System |

### 4. Instruction Pipeline

Modern CPUs use pipelining to execute multiple instructions simultaneously.

```
Instruction Pipeline (5-stage simplified):

Without Pipeline:
┌────────────────────────────────────────────────────┐
│ Inst1: Fetch→Decode→Execute→Memory→WriteBack      │
│                                                    │
│ Inst2:       Fetch→Decode→Execute→Memory→WriteBack│
└────────────────────────────────────────────────────┘
Time: 10 cycles for 2 instructions

With Pipeline:
┌────────────────────────────────────────────────────┐
│ Cycle: 1    2    3    4    5    6    7    8    9  │
│ Inst1: F    D    E    M    W                       │
│ Inst2:      F    D    E    M    W                  │
│ Inst3:           F    D    E    M    W             │
│ Inst4:                F    D    E    M    W        │
│ Inst5:                     F    D    E    M    W   │
└────────────────────────────────────────────────────┘
Time: 9 cycles for 5 instructions

F=Fetch, D=Decode, E=Execute, M=Memory, W=WriteBack
```

**Pipeline Hazards:**

1. **Structural Hazard**: Resource conflict
2. **Data Hazard**: Instruction depends on previous result
3. **Control Hazard**: Branch prediction failure

### 5. Branch Prediction

CPUs predict which way a branch will go to keep the pipeline full.

```
Branch Prediction:

Code:
if (x > 10) {
    // Path A
} else {
    // Path B
}

┌─────────────────────────────────────────┐
│      Branch Predictor                   │
│                                          │
│  History: A, A, A, A, A                 │
│  Prediction: A (90% confidence)         │
│                                          │
│  ┌────────────┐                         │
│  │ Predict A  │                         │
│  └─────┬──────┘                         │
│        │                                 │
│        ▼                                 │
│  ┌────────────┐      ┌────────────┐    │
│  │ Correct?   │─Yes─→│ Continue   │    │
│  └─────┬──────┘      └────────────┘    │
│        │                                 │
│        No                                │
│        │                                 │
│        ▼                                 │
│  ┌────────────┐                         │
│  │ Flush      │ ← Pipeline stall        │
│  │ Pipeline   │    (10-20 cycles lost)  │
│  └────────────┘                         │
└─────────────────────────────────────────┘
```

## CPU Architectures

### 1. x86/x64 (Intel, AMD)

**Characteristics:**
- CISC (Complex Instruction Set Computer)
- Variable-length instructions
- Rich instruction set
- Backward compatible

**Use Cases:**
- Servers
- Desktop computers
- High-performance computing

**Popular Processors:**
- Intel Xeon (servers)
- AMD EPYC (servers)
- Intel Core i7/i9 (desktop)
- AMD Ryzen (desktop)

### 2. ARM

**Characteristics:**
- RISC (Reduced Instruction Set Computer)
- Fixed-length instructions
- Power efficient
- Simpler design

**Use Cases:**
- Mobile devices
- IoT devices
- Apple Silicon (M1, M2, M3)
- AWS Graviton

**Advantages:**
- Lower power consumption
- Better performance per watt
- Scalable from tiny to large

### 3. RISC-V

**Characteristics:**
- Open-source ISA
- Modular design
- Royalty-free
- Growing ecosystem

**Use Cases:**
- Embedded systems
- Research
- Custom processors

## Performance Optimization

### 1. CPU Affinity

Binding processes to specific CPU cores.

```
CPU Affinity Example:

┌─────────────────────────────────────────┐
│         4-Core System                    │
│                                          │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐│
│  │Core 0│  │Core 1│  │Core 2│  │Core 3││
│  └───┬──┘  └───┬──┘  └───┬──┘  └───┬──┘│
│      │         │         │         │   │
│      ▼         ▼         ▼         ▼   │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐│
│  │Proc A│  │Proc B│  │Proc C│  │Proc D││
│  └──────┘  └──────┘  └──────┘  └──────┘│
└─────────────────────────────────────────┘

Benefits:
• Better cache utilization
• Reduced context switching
• Predictable performance
```

### 2. NUMA (Non-Uniform Memory Access)

In multi-socket systems, memory access time depends on location.

```
NUMA Architecture:

┌─────────────────────┐      ┌─────────────────────┐
│   Socket 0          │      │   Socket 1          │
│                     │      │                     │
│  ┌──────────────┐   │      │  ┌──────────────┐   │
│  │   CPU 0      │   │      │  │   CPU 1      │   │
│  │   8 cores    │   │      │  │   8 cores    │   │
│  └──────┬───────┘   │      │  └──────┬───────┘   │
│         │           │      │         │           │
│         ▼           │      │         ▼           │
│  ┌──────────────┐   │      │  ┌──────────────┐   │
│  │  Local RAM   │   │      │  │  Local RAM   │   │
│  │   64 GB      │   │      │  │   64 GB      │   │
│  └──────────────┘   │      │  └──────────────┘   │
└──────────┬──────────┘      └──────────┬──────────┘
           │                            │
           └────────────┬───────────────┘
                        │
                  Interconnect
                  (slower access)

Local Access:  ~100 ns
Remote Access: ~200 ns (2x slower!)
```

### 3. Cache Optimization

**Cache-Friendly Code:**

```go
// Bad: Cache-unfriendly (column-major access)
for col := 0; col < cols; col++ {
    for row := 0; row < rows; row++ {
        matrix[row][col] = value
    }
}

// Good: Cache-friendly (row-major access)
for row := 0; row < rows; row++ {
    for col := 0; col < cols; col++ {
        matrix[row][col] = value
    }
}
```

**Why?** Arrays are stored row-by-row in memory. Accessing sequentially loads entire cache lines.

```
Memory Layout:
[Row0Col0][Row0Col1][Row0Col2]...[Row1Col0][Row1Col1]...

Cache Line (64 bytes):
┌────────────────────────────────────────┐
│ [0,0] [0,1] [0,2] [0,3] [0,4] [0,5]... │
└────────────────────────────────────────┘

Row-major: Loads entire cache line, uses all data ✓
Column-major: Loads cache line, uses 1 element ✗
```

### 4. SIMD (Single Instruction Multiple Data)

Process multiple data elements with one instruction.

```
SIMD Operations:

Scalar (Traditional):
A[0] + B[0] = C[0]  ─┐
A[1] + B[1] = C[1]   ├─ 4 instructions
A[2] + B[2] = C[2]   │
A[3] + B[3] = C[3]  ─┘

SIMD (Vectorized):
┌────┬────┬────┬────┐   ┌────┬────┬────┬────┐
│A[0]│A[1]│A[2]│A[3]│ + │B[0]│B[1]│B[2]│B[3]│
└────┴────┴────┴────┘   └────┴────┴────┴────┘
         │                       │
         └───────────┬───────────┘
                     ▼
         ┌────┬────┬────┬────┐
         │C[0]│C[1]│C[2]│C[3]│  ← 1 instruction!
         └────┴────┴────┴────┘

4x speedup potential
```

**SIMD Instruction Sets:**
- SSE (Streaming SIMD Extensions)
- AVX (Advanced Vector Extensions)
- AVX-512
- ARM NEON

## CPU Selection for Different Workloads

### Workload Characteristics Matrix

```
┌─────────────────────────────────────────────────────────┐
│              Workload → CPU Mapping                     │
└─────────────────────────────────────────────────────────┘

Web Server (I/O Bound):
├─ Moderate core count (8-16 cores)
├─ High memory bandwidth
├─ Good single-thread performance
└─ Example: Intel Xeon Silver, AMD EPYC 7003

Database (Mixed):
├─ High core count (16-64 cores)
├─ Large cache (L3: 64-256 MB)
├─ High memory bandwidth
└─ Example: Intel Xeon Platinum, AMD EPYC 7004

Video Encoding (CPU Bound):
├─ Very high core count (32-128 cores)
├─ AVX-512 support
├─ High sustained frequency
└─ Example: AMD Threadripper, Intel Xeon W

Machine Learning (Compute Intensive):
├─ Specialized instructions (AVX-512, AMX)
├─ High memory bandwidth
├─ Or use GPU/TPU instead
└─ Example: Intel Xeon Scalable (Sapphire Rapids)

Microservices (Containerized):
├─ High core count for density
├─ Good multi-tenant performance
├─ Power efficiency
└─ Example: AWS Graviton, AMD EPYC
```

## Performance Monitoring

### Key Metrics to Monitor

```
CPU Performance Dashboard:

┌─────────────────────────────────────────┐
│ CPU Utilization                         │
│ ████████████████░░░░░░░░░░ 65%         │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Load Average (1m, 5m, 15m)              │
│ 2.5, 2.1, 1.8                           │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Context Switches/sec                    │
│ 15,234                                  │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ Cache Hit Rate                          │
│ L1: 95% | L2: 85% | L3: 70%            │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│ IPC (Instructions Per Cycle)            │
│ 2.1 (Good: >2.0)                        │
└─────────────────────────────────────────┘
```

### Linux Commands

```bash
# CPU information
lscpu
cat /proc/cpuinfo

# Real-time CPU usage
top
htop

# CPU statistics
mpstat -P ALL 1

# Cache information
lscpu -C

# NUMA topology
numactl --hardware

# Performance counters
perf stat -a sleep 5
```

## Summary

Key takeaways for architects:

1. **More cores ≠ better performance** - depends on workload parallelism
2. **Cache is critical** - design data structures for cache efficiency
3. **NUMA matters** - in multi-socket systems, memory locality is crucial
4. **Know your workload** - CPU-bound vs I/O-bound requires different optimization
5. **Modern CPUs are complex** - leverage hardware features (SIMD, branch prediction)

## Next Steps

Continue to [Memory Management](./03-memory-management.md) to understand how systems manage and optimize memory usage.
