# Performance Bottleneck Theory: CPU, Memory, and I/O Bounds

## Table of Contents
1. [Overview](#overview)
2. [CPU-Bound Systems](#cpu-bound-systems)
3. [Memory-Bound Systems](#memory-bound-systems)
4. [I/O-Bound Systems](#io-bound-systems)
5. [Mixed Bottlenecks](#mixed-bottlenecks)
6. [Systematic Diagnosis Framework](#systematic-diagnosis-framework)
7. [Tool Reference Guide](#tool-reference-guide)

---

## Overview

### What is a Performance Bottleneck?

A **performance bottleneck** is a component or resource in a system that limits overall performance. The system can only perform as fast as its slowest component allows.

**Analogy**: Think of a highway with multiple lanes that suddenly narrows to one lane. No matter how fast cars travel on the multi-lane section, they must slow down at the bottleneck.

### The Three Primary Bottleneck Types

```
┌─────────────────────────────────────────────────────────┐
│                    System Resources                      │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   CPU        │  │   Memory     │  │   I/O        │  │
│  │              │  │              │  │              │  │
│  │ Computation  │  │ Data Storage │  │ Disk/Network │  │
│  │ Processing   │  │ Caching      │  │ Operations   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                          │
└─────────────────────────────────────────────────────────┘

---

## CPU-Bound Systems

### What Does CPU-Bound Mean?

A system is **CPU-bound** when the bottleneck is the processor's ability to execute instructions. The CPU is working at or near 100% capacity, and adding more CPU power would directly improve performance. The program spends most of its time doing computation rather than waiting for external resources.

**Simple definition**: The CPU is the slowest part. It cannot process data fast enough.

**Real-world analogy**: A chef who can only chop vegetables at a fixed speed. No matter how many ingredients you bring, the chef (CPU) is the limiting factor.

### Symptoms of CPU-Bound Systems

#### Observable Symptoms

1. **High CPU utilization** — CPU usage consistently at 80-100% across all cores
2. **Slow response times** — Requests take longer than expected even with no I/O
3. **Linear degradation** — Performance degrades proportionally as load increases
4. **Low I/O wait** — Disk and network are mostly idle while CPU is maxed
5. **High load average** — System load average exceeds number of CPU cores
6. **Goroutine/thread starvation** — Other goroutines cannot get scheduled
7. **Latency spikes** — Periodic spikes when GC or other CPU tasks run
8. **No improvement from more connections** — Adding DB connections or network threads doesn't help
9. **Thermal throttling** — CPU overheats and reduces clock speed on physical machines
10. **Context switch overhead** — Too many goroutines competing for CPU time

#### Metrics That Indicate CPU-Bound

```
top / htop output:
  %Cpu(s): 98.5 us,  0.5 sy,  0.0 ni,  0.5 id,  0.0 wa

  us = user space CPU (your code)       → HIGH = CPU-bound in your code
  sy = kernel/system CPU                → HIGH = too many syscalls
  id = idle                             → LOW  = CPU is busy
  wa = I/O wait                         → LOW  = not I/O bound

Load average: 15.2, 14.8, 14.1  (on 8-core machine)
  → Load > num_cores means CPU is overloaded
```

### Possible Causes of CPU-Bound Problems

#### 1. Inefficient Algorithms (O(n²) or worse)

The most common cause. Using a nested loop where a hash map would work, or sorting repeatedly instead of once.

```go
// ❌ O(n²) — CPU-bound for large inputs
func findDuplicates(items []string) []string {
    var duplicates []string
    for i := 0; i < len(items); i++ {
        for j := i + 1; j < len(items); j++ {  // nested loop = O(n²)
            if items[i] == items[j] {
                duplicates = append(duplicates, items[i])
            }
        }
    }
    return duplicates
}

// ✅ O(n) — hash map lookup
func findDuplicates(items []string) []string {
    seen := make(map[string]bool, len(items))
    var duplicates []string
    for _, item := range items {
        if seen[item] {
            duplicates = append(duplicates, item)
        }
        seen[item] = true
    }
    return duplicates
}
```

#### 2. Excessive JSON Serialization / Deserialization

JSON encoding/decoding is CPU-intensive. Doing it in a hot path (every request) is a common mistake.

```go
// ❌ JSON marshal/unmarshal on every request
func handler(w http.ResponseWriter, r *http.Request) {
    var req Request
    json.NewDecoder(r.Body).Decode(&req)   // CPU: parse JSON
    result := process(req)
    json.NewEncoder(w).Encode(result)       // CPU: serialize JSON
}

// ✅ Use faster serializers or cache results
// Options: easyjson, jsoniter, protobuf, msgpack
import jsoniter "github.com/json-iterator/go"
var json = jsoniter.ConfigCompatibleWithStandardLibrary
```

#### 3. Garbage Collection Pressure

Go's GC runs on the CPU. Allocating millions of small objects forces frequent GC cycles, consuming CPU time.

```go
// ❌ Allocating in a loop — GC pressure
func processRequests(requests []Request) []Response {
    responses := []Response{}
    for _, req := range requests {
        result := &Response{Data: make([]byte, 1024)} // heap alloc every iteration
        responses = append(responses, *result)
    }
    return responses
}

// ✅ Reuse memory with sync.Pool
var responsePool = sync.Pool{
    New: func() interface{} {
        return &Response{Data: make([]byte, 1024)}
    },
}

func processRequests(requests []Request) []Response {
    responses := make([]Response, 0, len(requests))
    for _, req := range requests {
        result := responsePool.Get().(*Response)
        // use result...
        responses = append(responses, *result)
        responsePool.Put(result)
    }
    return responses
}
```

#### 4. Regex Compilation in Hot Path

Compiling a regex pattern is expensive. Doing it on every request is a CPU killer.

```go
// ❌ Compiling regex on every call
func validate(email string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
    return re.MatchString(email)
}

// ✅ Compile once at package level
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validate(email string) bool {
    return emailRegex.MatchString(email)
}
```

#### 5. Cryptographic Operations Without Hardware Acceleration

Hashing passwords (bcrypt), TLS handshakes, and encryption are CPU-intensive.

```go
// ❌ bcrypt cost too high for your hardware
hash, _ := bcrypt.GenerateFromPassword(password, 14) // cost=14 is very slow

// ✅ Tune cost to your hardware (benchmark first)
hash, _ := bcrypt.GenerateFromPassword(password, 10) // cost=10 is standard

// ✅ For high-throughput: use Argon2 with tuned parameters
// or offload to a dedicated auth service
```

#### 6. String Concatenation in Loops

In Go, strings are immutable. Concatenating with `+` in a loop creates a new string each time.

```go
// ❌ O(n²) string concatenation
func buildReport(lines []string) string {
    result := ""
    for _, line := range lines {
        result += line + "\n"  // new allocation every iteration
    }
    return result
}

// ✅ strings.Builder — O(n)
func buildReport(lines []string) string {
    var sb strings.Builder
    sb.Grow(len(lines) * 80) // pre-allocate estimated size
    for _, line := range lines {
        sb.WriteString(line)
        sb.WriteByte('\n')
    }
    return sb.String()
}
```

#### 7. Unnecessary Reflection

`reflect` package operations are 10-100x slower than direct operations.

```go
// ❌ Using reflection in hot path
func setField(obj interface{}, name string, value interface{}) {
    reflect.ValueOf(obj).Elem().FieldByName(name).Set(reflect.ValueOf(value))
}

// ✅ Use direct struct access or code generation
// Use tools like: go generate, protobuf, or direct field access
```

#### 8. Too Many Goroutines Competing for CPU

Spawning a goroutine per request without limits causes scheduler overhead.

```go
// ❌ Unbounded goroutines
for _, item := range millionItems {
    go process(item)  // 1,000,000 goroutines competing for CPU
}

// ✅ Worker pool with bounded concurrency
func processWithPool(items []Item, workers int) {
    jobs := make(chan Item, len(items))
    var wg sync.WaitGroup

    for w := 0; w < workers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for item := range jobs {
                process(item)
            }
        }()
    }

    for _, item := range items {
        jobs <- item
    }
    close(jobs)
    wg.Wait()
}
```

### How to Identify CPU-Bound Problems

#### Step 1: Check System-Level CPU Usage

```bash
# Real-time CPU usage per core
top -1          # Linux: show all CPUs
htop            # Interactive, color-coded

# macOS
top -o cpu      # Sort by CPU usage

# Check load average vs CPU count
uptime
# output: load average: 7.5, 7.2, 6.9
# If load > num_cores → CPU overloaded

nproc           # Linux: number of CPU cores
sysctl -n hw.ncpu  # macOS: number of CPU cores
```

#### Step 2: Profile CPU Usage in Go (pprof)

```go
// Add to your main.go or HTTP server
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    // ... rest of your app
}
```

```bash
# Capture 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Inside pprof interactive shell:
(pprof) top10          # top 10 CPU-consuming functions
(pprof) top10 -cum     # cumulative (includes callees)
(pprof) list funcName  # show source with CPU annotations
(pprof) web            # open flame graph in browser (requires graphviz)

# Generate flame graph directly
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

#### Step 3: Write CPU Benchmarks

```go
// benchmark_test.go
func BenchmarkFindDuplicates(b *testing.B) {
    items := generateItems(10000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        findDuplicates(items)
    }
}

// Run benchmark
go test -bench=. -benchmem -cpuprofile=cpu.prof ./...

// Analyze profile
go tool pprof cpu.prof
```

#### Step 4: Trace Goroutine Scheduling

```go
import "runtime/trace"

f, _ := os.Create("trace.out")
trace.Start(f)
defer trace.Stop()
// ... run your code
```

```bash
go tool trace trace.out
# Opens browser with:
# - Goroutine scheduling timeline
# - CPU utilization per core
# - GC events
# - Syscall events
```

#### Step 5: Check GC Pressure

```bash
# Run with GC stats
GODEBUG=gctrace=1 ./myapp

# Output example:
# gc 14 @2.345s 8%: 0.5+12+0.3 ms clock, 4+8/10/0+2 ms cpu, 45->48->24 MB, 50 MB goal, 8 P
#                ^                                                ^
#                8% of time in GC                                heap size
```

### How to Fix CPU-Bound Problems

#### Fix 1: Parallelize with GOMAXPROCS

```go
// Use all available CPU cores
import "runtime"

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    // Go 1.5+ sets this automatically, but explicit is clear
}
```

#### Fix 2: Use Worker Pools for CPU-Intensive Tasks

```go
type CPUWorkerPool struct {
    jobs    chan func()
    wg      sync.WaitGroup
    workers int
}

func NewCPUWorkerPool(workers int) *CPUWorkerPool {
    p := &CPUWorkerPool{
        jobs:    make(chan func(), workers*10),
        workers: workers,
    }
    for i := 0; i < workers; i++ {
        p.wg.Add(1)
        go func() {
            defer p.wg.Done()
            for job := range p.jobs {
                job()
            }
        }()
    }
    return p
}

func (p *CPUWorkerPool) Submit(job func()) {
    p.jobs <- job
}

func (p *CPUWorkerPool) Wait() {
    close(p.jobs)
    p.wg.Wait()
}

// Usage: one worker per CPU core for CPU-bound work
pool := NewCPUWorkerPool(runtime.NumCPU())
```

#### Fix 3: Reduce Allocations (Escape Analysis)

```go
// Check what escapes to heap
go build -gcflags="-m" ./...

// Output:
// ./main.go:15:12: &Response literal escapes to heap
// ./main.go:22:14: make([]byte, 1024) does not escape

// Use stack allocation where possible
// Avoid returning pointers to local variables in hot paths
// Use value types instead of pointer types for small structs
```

#### Fix 4: Cache Expensive Computations

```go
// Memoization for pure functions
type Cache struct {
    mu    sync.RWMutex
    store map[string]Result
}

func (c *Cache) Get(key string) (Result, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    r, ok := c.store[key]
    return r, ok
}

func (c *Cache) Set(key string, result Result) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.store[key] = result
}

func expensiveCompute(input string) Result {
    if r, ok := globalCache.Get(input); ok {
        return r  // cache hit: no CPU work
    }
    result := doHeavyComputation(input)
    globalCache.Set(input, result)
    return result
}
```

#### Fix 5: Use SIMD / Assembly for Hot Loops

For extreme cases, use Go assembly or cgo to leverage CPU vector instructions (SIMD).

```go
// Use optimized standard library functions that use SIMD internally
// bytes.Equal, bytes.IndexByte, strings.Contains use assembly

// For custom hot loops, consider:
// - github.com/klauspost/compress (SIMD compression)
// - github.com/minio/sha256-simd (SIMD hashing)
```

#### Fix 6: Tune GC

```go
// Increase GC target to reduce frequency (uses more memory)
// Default is 100 (GC when heap doubles)
os.Setenv("GOGC", "200")  // GC when heap grows 200%

// Or in code:
debug.SetGCPercent(200)

// Disable GC for batch jobs (re-enable after)
debug.SetGCPercent(-1)
defer debug.SetGCPercent(100)
```

### CPU-Bound: Tools Summary

| Tool | Purpose | Command |
|------|---------|---------|
| `top` / `htop` | Real-time CPU per core | `top -1` |
| `pprof` | CPU flame graph | `go tool pprof` |
| `go trace` | Goroutine scheduling | `go tool trace` |
| `perf` (Linux) | Hardware CPU counters | `perf stat ./app` |
| `instruments` (macOS) | CPU profiling GUI | Xcode Instruments |
| `go bench` | Micro-benchmarks | `go test -bench=.` |
| `GODEBUG=gctrace` | GC frequency/duration | env var |
| `go build -gcflags="-m"` | Escape analysis | build flag |


---

## Memory-Bound Systems

### What Does Memory-Bound Mean?

A system is **memory-bound** when the bottleneck is RAM — either there is not enough of it, it is being used inefficiently, or the CPU is spending most of its time waiting for data to be fetched from memory rather than computing.

There are two distinct sub-types:

1. **Memory capacity bound** — Running out of RAM, causing swapping or OOM kills
2. **Memory bandwidth bound** — CPU is fast but memory bus cannot deliver data fast enough (cache misses)

**Simple definition**: Either you don't have enough memory, or accessing memory is too slow.

**Real-world analogy**: A chef with a tiny counter (RAM). They constantly have to walk to the pantry (disk/swap) to get ingredients because the counter is full. The walking time dominates the cooking time.

### Symptoms of Memory-Bound Systems

#### Observable Symptoms

1. **High memory usage** — RSS/heap growing continuously or near system limit
2. **OOM kills** — Process killed by OS with "out of memory" error
3. **Swap usage** — System using swap space (disk as RAM — 1000x slower)
4. **Memory leaks** — Heap grows over time and never shrinks
5. **Frequent GC pauses** — GC runs constantly trying to free memory
6. **High GC CPU usage** — 20-40% of CPU time spent in garbage collection
7. **Cache miss rate** — CPU stalls waiting for data from RAM (L1/L2/L3 cache misses)
8. **Slow after running for hours** — Memory leak causes degradation over time
9. **OOM errors in logs** — `runtime: out of memory`, `fatal error: runtime: out of memory`
10. **Goroutine leak** — Number of goroutines grows unboundedly

#### Metrics That Indicate Memory-Bound

```
free -h output (Linux):
              total   used    free    shared  buff/cache  available
Mem:           16G    15.8G   200M    100M    500M        100M
Swap:           4G     3.9G   100M
→ Swap being used = memory-bound (severe)

Go runtime metrics:
  HeapAlloc:   14.2 GB   → current heap in use
  HeapSys:     16.0 GB   → total heap from OS
  HeapIdle:    200 MB    → heap not in use
  NumGC:       1847      → GC has run 1847 times (high)
  PauseTotalNs: 45s      → 45 seconds total in GC pauses
```

### Possible Causes of Memory-Bound Problems

#### 1. Memory Leaks — Goroutine Leaks

The most common Go memory leak. A goroutine is started but never exits, holding references to memory.

```go
// ❌ Goroutine leak — channel never closed, goroutine blocks forever
func startWorker(data chan int) {
    go func() {
        for v := range data {  // blocks here if data is never closed
            process(v)
        }
    }()
}

// Called 1000 times → 1000 goroutines stuck forever
for i := 0; i < 1000; i++ {
    ch := make(chan int)
    startWorker(ch)
    // ch is never closed!
}

// ✅ Always use context for cancellation
func startWorker(ctx context.Context, data chan int) {
    go func() {
        for {
            select {
            case v, ok := <-data:
                if !ok {
                    return  // channel closed, goroutine exits
                }
                process(v)
            case <-ctx.Done():
                return  // context cancelled, goroutine exits
            }
        }
    }()
}
```

#### 2. Memory Leaks — Slice Retaining Large Backing Array

Slicing a large slice keeps the entire backing array in memory.

```go
// ❌ Retains entire 1GB backing array
func getFirstTen(data []byte) []byte {
    return data[:10]  // small slice, but holds reference to all of data
}

// ✅ Copy to new slice — releases backing array
func getFirstTen(data []byte) []byte {
    result := make([]byte, 10)
    copy(result, data[:10])
    return result
}
```

#### 3. Memory Leaks — Map Never Shrinks

Go maps grow but never shrink. Deleting keys frees values but not the map's internal buckets.

```go
// ❌ Map grows to hold 1M entries, then entries deleted
// Map still holds memory for 1M buckets
cache := make(map[string][]byte)
for i := 0; i < 1_000_000; i++ {
    cache[fmt.Sprintf("key-%d", i)] = make([]byte, 1024)
}
for k := range cache {
    delete(cache, k)  // values freed, but map buckets remain
}
// cache still uses ~50MB of memory

// ✅ Replace map with new one after bulk delete
cache = make(map[string][]byte)  // old map GC'd

// ✅ Or use a cache library with TTL and eviction
// github.com/patrickmn/go-cache
// github.com/dgraph-io/ristretto
```

#### 4. Unbounded Caches

Caching without eviction causes memory to grow until OOM.

```go
// ❌ Unbounded cache — grows forever
var cache = make(map[string][]byte)

func getFromCache(key string) []byte {
    if v, ok := cache[key]; ok {
        return v
    }
    v := fetchFromDB(key)
    cache[key] = v  // never evicted!
    return v
}

// ✅ LRU cache with size limit
import lru "github.com/hashicorp/golang-lru"

var cache, _ = lru.New(10000)  // max 10,000 entries

func getFromCache(key string) []byte {
    if v, ok := cache.Get(key); ok {
        return v.([]byte)
    }
    v := fetchFromDB(key)
    cache.Add(key, v)  // automatically evicts LRU entry
    return v
}
```

#### 5. Large Object Allocation in Hot Path

Allocating large objects frequently causes GC pressure.

```go
// ❌ Allocating 1MB buffer on every request
func handleRequest(r *http.Request) {
    buf := make([]byte, 1024*1024)  // 1MB per request
    io.ReadFull(r.Body, buf)
    process(buf)
}

// ✅ Pool large buffers
var bufPool = sync.Pool{
    New: func() interface{} {
        buf := make([]byte, 1024*1024)
        return &buf
    },
}

func handleRequest(r *http.Request) {
    bufPtr := bufPool.Get().(*[]byte)
    buf := *bufPtr
    defer bufPool.Put(bufPtr)

    io.ReadFull(r.Body, buf)
    process(buf)
}
```

#### 6. String Interning Issues

Storing many duplicate strings wastes memory.

```go
// ❌ Storing 1M copies of the same status strings
type Event struct {
    Status string  // "active", "inactive", "pending" — repeated 1M times
}

// ✅ Use string interning or enums
type Status int
const (
    StatusActive Status = iota
    StatusInactive
    StatusPending
)

type Event struct {
    Status Status  // 8 bytes instead of 16+ bytes per string
}
```

#### 7. Holding References in Closures

Closures capture variables by reference, preventing GC.

```go
// ❌ Closure captures large slice, preventing GC
func processLargeData() func() string {
    data := make([]byte, 100*1024*1024)  // 100MB
    result := compute(data)
    return func() string {
        return result  // closure holds reference to data (100MB stays in memory)
    }
}

// ✅ Only capture what you need
func processLargeData() func() string {
    data := make([]byte, 100*1024*1024)
    result := compute(data)
    data = nil  // explicitly release large data
    return func() string {
        return result  // only result is captured
    }
}
```

### How to Identify Memory-Bound Problems

#### Step 1: Check System Memory Usage

```bash
# Linux
free -h                    # overall memory and swap
cat /proc/meminfo          # detailed memory breakdown
vmstat -s                  # virtual memory stats
cat /proc/<pid>/status     # per-process memory (VmRSS = resident set size)

# macOS
vm_stat                    # virtual memory statistics
top -o mem                 # sort processes by memory
activity monitor           # GUI

# Check if swap is being used (bad sign)
swapon --show              # Linux
sysctl vm.swapusage        # macOS
```

#### Step 2: Profile Memory in Go (pprof heap)

```go
// Add pprof endpoint
import _ "net/http/pprof"

go func() {
    http.ListenAndServe("localhost:6060", nil)
}()
```

```bash
# Capture heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Inside pprof:
(pprof) top10              # top memory consumers
(pprof) top10 -cum         # cumulative allocations
(pprof) list funcName      # source with memory annotations
(pprof) web                # flame graph

# Compare two heap profiles (find leaks)
go tool pprof -base heap1.prof heap2.prof
(pprof) top10              # shows what GREW between snapshots
```

#### Step 3: Check Goroutine Count

```bash
# Check number of goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Or in code
fmt.Println("Goroutines:", runtime.NumGoroutine())

# Goroutine profile (shows stack traces of all goroutines)
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

#### Step 4: Monitor Runtime Memory Stats

```go
func printMemStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    fmt.Printf("HeapAlloc:   %v MB\n", m.HeapAlloc/1024/1024)
    fmt.Printf("HeapSys:     %v MB\n", m.HeapSys/1024/1024)
    fmt.Printf("HeapObjects: %v\n",    m.HeapObjects)
    fmt.Printf("NumGC:       %v\n",    m.NumGC)
    fmt.Printf("PauseTotal:  %v ms\n", m.PauseTotalNs/1e6)
    fmt.Printf("Goroutines:  %v\n",    runtime.NumGoroutine())
}

// Call periodically
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        printMemStats()
    }
}()
```

#### Step 5: Detect Goroutine Leaks with goleak

```go
// In tests
import "go.uber.org/goleak"

func TestMyFunction(t *testing.T) {
    defer goleak.VerifyNone(t)  // fails if goroutines leaked
    myFunction()
}
```

#### Step 6: Enable GC Trace

```bash
GODEBUG=gctrace=1 ./myapp 2>&1 | head -50

# Output:
# gc 1 @0.012s 2%: 0.1+1.2+0.1 ms clock, 0.8+0.4/1.0/0+0.8 ms cpu, 4->4->2 MB, 5 MB goal, 8 P
#        ^      ^                                                       ^    ^    ^
#        time   % in GC                                                before->after->live heap
```

### How to Fix Memory-Bound Problems

#### Fix 1: Fix Goroutine Leaks with Context

```go
// Pattern: always pass context, always handle Done
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case job, ok := <-jobs:
            if !ok {
                return  // channel closed
            }
            process(job)
        case <-ctx.Done():
            return  // cancelled
        }
    }
}

// Always cancel contexts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()  // ALWAYS defer cancel
```

#### Fix 2: Use sync.Pool for Frequently Allocated Objects

```go
var pool = sync.Pool{
    New: func() interface{} {
        return &MyObject{
            Buffer: make([]byte, 4096),
        }
    },
}

func process() {
    obj := pool.Get().(*MyObject)
    defer pool.Put(obj)

    // reset state before use
    obj.reset()
    // use obj...
}
```

#### Fix 3: Implement Proper Cache Eviction

```go
// TTL-based cache
type TTLCache struct {
    mu    sync.RWMutex
    items map[string]cacheItem
}

type cacheItem struct {
    value     interface{}
    expiresAt time.Time
}

func (c *TTLCache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = cacheItem{value: value, expiresAt: time.Now().Add(ttl)}
}

func (c *TTLCache) cleanup() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        c.mu.Lock()
        for k, v := range c.items {
            if time.Now().After(v.expiresAt) {
                delete(c.items, k)
            }
        }
        c.mu.Unlock()
    }
}
```

#### Fix 4: Reduce Heap Allocations

```go
// Use value types instead of pointers for small structs
type Point struct{ X, Y float64 }  // 16 bytes on stack

// Pre-allocate slices with known capacity
items := make([]Item, 0, expectedCount)  // no re-allocation

// Use arrays instead of slices for fixed-size data
var buf [256]byte  // stack allocated, not heap

// Avoid interface{} boxing in hot paths
// interface{} causes heap allocation for non-pointer types
```

#### Fix 5: Set Memory Limits (Go 1.19+)

```go
import "runtime/debug"

// Set soft memory limit (Go 1.19+)
// GC will run more aggressively to stay under this limit
debug.SetMemoryLimit(512 * 1024 * 1024)  // 512MB limit
```

### Memory-Bound: Tools Summary

| Tool | Purpose | Command |
|------|---------|---------|
| `free -h` | System memory overview | `free -h` |
| `pprof heap` | Heap allocation profile | `go tool pprof .../heap` |
| `pprof goroutine` | Goroutine leak detection | `go tool pprof .../goroutine` |
| `goleak` | Goroutine leak in tests | `goleak.VerifyNone(t)` |
| `GODEBUG=gctrace` | GC frequency and pause | env var |
| `runtime.MemStats` | Programmatic memory stats | Go API |
| `valgrind` | Memory errors (cgo) | `valgrind ./app` |
| `heapster` | Kubernetes memory metrics | K8s tool |
| `debug.SetMemoryLimit` | Soft memory cap | Go 1.19+ API |


---

## I/O-Bound Systems

### What Does I/O-Bound Mean?

A system is **I/O-bound** when the bottleneck is input/output operations — reading from or writing to disk, network, databases, or other external systems. The CPU is mostly idle, waiting for data to arrive or be written.

There are two main types:
1. **Disk I/O bound** — Slow reads/writes to storage (HDD, SSD, NFS)
2. **Network I/O bound** — Slow network calls (HTTP APIs, databases, message queues)

**Simple definition**: The program spends most of its time waiting, not computing.

**Real-world analogy**: A chef waiting for ingredients to be delivered from a warehouse. The chef (CPU) is idle and ready, but cannot work until the delivery (I/O) arrives.

### Symptoms of I/O-Bound Systems

#### Observable Symptoms

1. **High I/O wait** — CPU `%wa` (iowait) is high (>20%)
2. **Low CPU utilization** — CPU is mostly idle while requests are slow
3. **High disk utilization** — Disk at 100% with long queue depth
4. **Slow database queries** — Queries taking seconds instead of milliseconds
5. **Network latency** — High RTT (round-trip time) to external services
6. **Goroutines blocked on I/O** — Most goroutines in `chan receive` or `syscall` state
7. **Throughput doesn't improve with more CPU** — Adding cores doesn't help
8. **Improves with caching** — Adding a cache dramatically improves performance
9. **Disk queue depth > 1** — Requests queuing up waiting for disk
10. **Connection pool exhaustion** — All DB connections in use, new requests wait

#### Metrics That Indicate I/O-Bound

```
top output:
  %Cpu(s):  2.1 us,  0.5 sy,  0.0 ni, 17.4 id, 79.8 wa
                                                  ^^^^
                                                  79.8% I/O wait = I/O bound

iostat -x output:
  Device  r/s   w/s   rkB/s  wkB/s  await  %util
  sda     0.0  1200   0.0    4800   180ms  100%
                                    ^^^^^  ^^^^
                                    180ms wait  100% busy = disk bottleneck

netstat / ss output:
  Recv-Q  Send-Q  Local Address  Foreign Address  State
  0       65535   0.0.0.0:8080   ...              LISTEN
          ^^^^^
          Send queue full = network I/O bound
```

### Possible Causes of I/O-Bound Problems

#### 1. Synchronous Blocking I/O in Serial

Making I/O calls one after another when they could be parallel.

```go
// ❌ Serial I/O — total time = sum of all calls
func getUserData(userID int) UserData {
    profile := fetchProfile(userID)    // 100ms
    orders  := fetchOrders(userID)     // 150ms
    reviews := fetchReviews(userID)    // 80ms
    // Total: 330ms
    return combine(profile, orders, reviews)
}

// ✅ Parallel I/O — total time = slowest call
func getUserData(userID int) UserData {
    var (
        profile Profile
        orders  []Order
        reviews []Review
        wg      sync.WaitGroup
        mu      sync.Mutex
    )

    wg.Add(3)
    go func() { defer wg.Done(); mu.Lock(); profile = fetchProfile(userID); mu.Unlock() }()
    go func() { defer wg.Done(); mu.Lock(); orders = fetchOrders(userID); mu.Unlock() }()
    go func() { defer wg.Done(); mu.Lock(); reviews = fetchReviews(userID); mu.Unlock() }()
    wg.Wait()
    // Total: 150ms (slowest call)
    return combine(profile, orders, reviews)
}
```

#### 2. No Connection Pooling

Opening a new database/HTTP connection for every request is expensive (TCP handshake + TLS = 50-200ms).

```go
// ❌ New connection per request
func queryDB(query string) []Row {
    db, _ := sql.Open("postgres", dsn)  // new connection every time!
    defer db.Close()
    rows, _ := db.Query(query)
    return scanRows(rows)
}

// ✅ Shared connection pool
var db *sql.DB

func init() {
    db, _ = sql.Open("postgres", dsn)
    db.SetMaxOpenConns(25)       // max connections
    db.SetMaxIdleConns(10)       // keep 10 idle connections warm
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(1 * time.Minute)
}

func queryDB(query string) []Row {
    rows, _ := db.Query(query)  // reuses pooled connection
    return scanRows(rows)
}
```

#### 3. Missing Database Indexes

A query without an index does a full table scan — reads every row from disk.

```sql
-- ❌ No index on user_id — full table scan
SELECT * FROM orders WHERE user_id = 12345;
-- Reads all 10M rows from disk: 30 seconds

-- ✅ Add index
CREATE INDEX idx_orders_user_id ON orders(user_id);
-- Index lookup: reads ~10 rows: 2ms
```

```go
// Detect slow queries in Go
db.SetConnMaxLifetime(time.Minute)

// Use query logging middleware
type loggingDB struct {
    db *sql.DB
}

func (l *loggingDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
    start := time.Now()
    rows, err := l.db.Query(query, args...)
    duration := time.Since(start)
    if duration > 100*time.Millisecond {
        log.Printf("SLOW QUERY (%v): %s", duration, query)
    }
    return rows, err
}
```

#### 4. Reading/Writing Files Without Buffering

Unbuffered I/O makes a syscall for every byte, causing massive overhead.

```go
// ❌ Unbuffered — one syscall per write
func writeLog(file *os.File, msg string) {
    file.WriteString(msg + "\n")  // syscall every time
}

// ✅ Buffered — batches writes, fewer syscalls
writer := bufio.NewWriterSize(file, 64*1024)  // 64KB buffer
func writeLog(msg string) {
    writer.WriteString(msg + "\n")  // writes to buffer
}
// Flush periodically or on close
writer.Flush()
```

#### 5. N+1 Query Problem

Loading a list of items, then making one DB query per item.

```go
// ❌ N+1: 1 query for users + 1 query per user = 1001 queries
users, _ := db.Query("SELECT id, name FROM users LIMIT 1000")
for _, user := range users {
    orders, _ := db.Query("SELECT * FROM orders WHERE user_id = ?", user.ID)
    user.Orders = orders
}
// 1001 round trips to DB × 5ms each = 5 seconds

// ✅ Single JOIN query
rows, _ := db.Query(`
    SELECT u.id, u.name, o.id, o.total
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    LIMIT 1000
`)
// 1 round trip = 50ms
```

#### 6. No Caching for Repeated Reads

Reading the same data from disk or DB repeatedly when it rarely changes.

```go
// ❌ DB query on every request for rarely-changing data
func getConfig(key string) string {
    var value string
    db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
    return value  // DB hit every time
}

// ✅ Cache with TTL
var configCache = &sync.Map{}

func getConfig(key string) string {
    if v, ok := configCache.Load(key); ok {
        return v.(string)
    }
    var value string
    db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
    configCache.Store(key, value)
    return value
}
```

#### 7. Large Payloads Without Streaming

Loading entire large files or responses into memory before processing.

```go
// ❌ Load entire file into memory
data, _ := os.ReadFile("huge_file.csv")  // 2GB into RAM
process(data)

// ✅ Stream line by line
file, _ := os.Open("huge_file.csv")
defer file.Close()
scanner := bufio.NewScanner(file)
for scanner.Scan() {
    processLine(scanner.Text())  // one line at a time
}
```

#### 8. Chatty Protocols — Too Many Small Requests

Making many small network requests instead of batching them.

```go
// ❌ One Redis call per item
for _, userID := range userIDs {
    score, _ := redisClient.ZScore(ctx, "leaderboard", userID).Result()
    scores[userID] = score
}
// 1000 round trips × 1ms = 1 second

// ✅ Batch with pipeline
pipe := redisClient.Pipeline()
cmds := make([]*redis.FloatCmd, len(userIDs))
for i, userID := range userIDs {
    cmds[i] = pipe.ZScore(ctx, "leaderboard", userID)
}
pipe.Exec(ctx)
// 1 round trip = 5ms
```

### How to Identify I/O-Bound Problems

#### Step 1: Check I/O Wait at System Level

```bash
# Linux: check iowait
top                    # look at %wa column
iostat -x 1            # disk I/O stats every second
iostat -x 1 | grep -v "^$"

# Key iostat columns:
# await  = average wait time per I/O request (ms)
# %util  = % of time device was busy
# r/s    = reads per second
# w/s    = writes per second

# macOS
iostat -w 1            # disk stats
fs_usage               # per-process file system calls

# Network I/O
netstat -s             # network statistics
ss -s                  # socket summary
iftop                  # real-time network usage per connection
```

#### Step 2: Profile I/O in Go (pprof block + mutex)

```bash
# Block profile — shows where goroutines are blocked (waiting for I/O, channels, mutexes)
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex profile — shows mutex contention
go tool pprof http://localhost:6060/debug/pprof/mutex

# Enable block profiling in code (has overhead, use in dev/staging)
runtime.SetBlockProfileRate(1)    // profile every blocking event
runtime.SetMutexProfileFraction(1) // profile every mutex contention
```

#### Step 3: Use Go Trace to See I/O Blocking

```go
f, _ := os.Create("trace.out")
trace.Start(f)
// ... run workload
trace.Stop()
f.Close()
```

```bash
go tool trace trace.out
# Look for:
# - Long syscall durations (disk reads/writes)
# - Goroutines in "network wait" state
# - Goroutines in "chan receive" state (waiting for data)
```

#### Step 4: Identify Slow Database Queries

```bash
# PostgreSQL: find slow queries
SELECT query, mean_exec_time, calls, total_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

# MySQL: slow query log
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 0.1;  -- log queries > 100ms

# PostgreSQL: EXPLAIN ANALYZE
EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 12345;
# Look for: Seq Scan (bad) vs Index Scan (good)
```

#### Step 5: Trace Network Calls

```bash
# tcpdump — capture network packets
tcpdump -i eth0 -w capture.pcap port 5432  # capture PostgreSQL traffic

# strace — trace syscalls (Linux)
strace -p <pid> -e trace=network,read,write 2>&1 | head -100

# dtruss — trace syscalls (macOS)
sudo dtruss -p <pid> 2>&1 | grep -E "read|write|connect"

# lsof — list open files and network connections
lsof -p <pid>
lsof -i :5432  # who is connected to PostgreSQL
```

#### Step 6: Check Connection Pool Health

```go
// Monitor DB pool stats
func monitorDBPool(db *sql.DB) {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        stats := db.Stats()
        log.Printf("DB Pool: open=%d idle=%d inUse=%d waitCount=%d waitDuration=%v",
            stats.OpenConnections,
            stats.Idle,
            stats.InUse,
            stats.WaitCount,
            stats.WaitDuration,
        )
        // Alert if WaitCount > 0 (requests waiting for connection)
    }
}
```

### How to Fix I/O-Bound Problems

#### Fix 1: Parallelize Independent I/O with errgroup

```go
import "golang.org/x/sync/errgroup"

func getUserData(ctx context.Context, userID int) (*UserData, error) {
    g, ctx := errgroup.WithContext(ctx)

    var profile *Profile
    var orders  []Order
    var reviews []Review

    g.Go(func() error {
        var err error
        profile, err = fetchProfile(ctx, userID)
        return err
    })
    g.Go(func() error {
        var err error
        orders, err = fetchOrders(ctx, userID)
        return err
    })
    g.Go(func() error {
        var err error
        reviews, err = fetchReviews(ctx, userID)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return combine(profile, orders, reviews), nil
}
```

#### Fix 2: Implement Multi-Level Caching

```go
// L1: in-process memory cache (fastest)
// L2: Redis (shared across instances)
// L3: Database (source of truth)

type MultiLevelCache struct {
    l1    *lru.Cache       // in-memory, ~1ms
    l2    *redis.Client    // Redis, ~1-5ms
    l3    *sql.DB          // Database, ~5-50ms
}

func (c *MultiLevelCache) Get(ctx context.Context, key string) ([]byte, error) {
    // L1 check
    if v, ok := c.l1.Get(key); ok {
        return v.([]byte), nil
    }

    // L2 check
    v, err := c.l2.Get(ctx, key).Bytes()
    if err == nil {
        c.l1.Add(key, v)  // populate L1
        return v, nil
    }

    // L3 fetch
    v, err = c.fetchFromDB(ctx, key)
    if err != nil {
        return nil, err
    }

    // Populate caches
    c.l2.Set(ctx, key, v, 5*time.Minute)
    c.l1.Add(key, v)
    return v, nil
}
```

#### Fix 3: Use Read Replicas for Read-Heavy Workloads

```go
type DBCluster struct {
    primary  *sql.DB
    replicas []*sql.DB
    counter  uint64
}

// Route reads to replicas, writes to primary
func (c *DBCluster) Read(query string, args ...interface{}) (*sql.Rows, error) {
    idx := atomic.AddUint64(&c.counter, 1) % uint64(len(c.replicas))
    return c.replicas[idx].Query(query, args...)
}

func (c *DBCluster) Write(query string, args ...interface{}) (sql.Result, error) {
    return c.primary.Exec(query, args...)
}
```

#### Fix 4: Implement Write Batching

```go
// Batch writes to reduce I/O operations
type WriteBatcher struct {
    mu       sync.Mutex
    pending  []WriteOp
    maxBatch int
    maxWait  time.Duration
    flush    func([]WriteOp) error
}

func (b *WriteBatcher) Write(op WriteOp) {
    b.mu.Lock()
    b.pending = append(b.pending, op)
    shouldFlush := len(b.pending) >= b.maxBatch
    b.mu.Unlock()

    if shouldFlush {
        b.Flush()
    }
}

func (b *WriteBatcher) run() {
    ticker := time.NewTicker(b.maxWait)
    for range ticker.C {
        b.Flush()
    }
}

func (b *WriteBatcher) Flush() {
    b.mu.Lock()
    if len(b.pending) == 0 {
        b.mu.Unlock()
        return
    }
    batch := b.pending
    b.pending = nil
    b.mu.Unlock()

    b.flush(batch)  // single I/O operation for entire batch
}
```

#### Fix 5: Use Async I/O for Non-Critical Writes

```go
// Fire-and-forget for non-critical writes (analytics, logs)
type AsyncWriter struct {
    queue chan WriteOp
}

func NewAsyncWriter(bufSize int) *AsyncWriter {
    w := &AsyncWriter{queue: make(chan WriteOp, bufSize)}
    go w.worker()
    return w
}

func (w *AsyncWriter) Write(op WriteOp) error {
    select {
    case w.queue <- op:
        return nil
    default:
        return errors.New("write queue full")  // backpressure
    }
}

func (w *AsyncWriter) worker() {
    for op := range w.queue {
        performWrite(op)  // actual I/O happens asynchronously
    }
}
```

### I/O-Bound: Tools Summary

| Tool | Purpose | Command |
|------|---------|---------|
| `iostat -x 1` | Disk I/O stats | `iostat -x 1` |
| `iotop` | Per-process disk I/O | `iotop -o` |
| `netstat` / `ss` | Network connections | `ss -s` |
| `iftop` | Real-time network usage | `iftop` |
| `tcpdump` | Capture network packets | `tcpdump -i eth0` |
| `strace` / `dtruss` | Syscall tracing | `strace -p <pid>` |
| `lsof` | Open files/connections | `lsof -p <pid>` |
| `pprof block` | Blocking profile | `go tool pprof .../block` |
| `go trace` | I/O wait visualization | `go tool trace` |
| `EXPLAIN ANALYZE` | DB query plan | SQL command |
| `pg_stat_statements` | Slow query tracking | PostgreSQL extension |
| `fio` | Disk I/O benchmarking | `fio --name=test` |


---

## Mixed Bottlenecks

### Why Mixed Bottlenecks Are Common

Real systems rarely have a single pure bottleneck. Most production systems exhibit combinations:

- **CPU + I/O**: Image processing service (CPU for encoding, I/O for reading/writing files)
- **Memory + I/O**: Database (memory for buffer pool, I/O for disk reads)
- **CPU + Memory**: Machine learning inference (CPU for computation, memory for model weights)
- **All three**: Video transcoding (CPU for encoding, memory for frame buffers, I/O for reading/writing video)

### How to Identify Mixed Bottlenecks

The key is to measure all three simultaneously and find which one is the primary constraint.

```bash
# Run all three monitors simultaneously in separate terminals

# Terminal 1: CPU
watch -n 1 "mpstat -P ALL 1 1"

# Terminal 2: Memory
watch -n 1 "free -h && cat /proc/meminfo | grep -E 'MemAvailable|SwapUsed'"

# Terminal 3: I/O
iostat -x 1

# Terminal 4: Your application
./myapp --load-test
```

### The Bottleneck Shift Phenomenon

Fixing one bottleneck often reveals the next one. This is normal and expected.

```
Before fix:  CPU 95%, Memory 40%, I/O 20%  → CPU is bottleneck
Fix CPU:     CPU 60%, Memory 40%, I/O 80%  → I/O is now bottleneck
Fix I/O:     CPU 60%, Memory 85%, I/O 40%  → Memory is now bottleneck
Fix Memory:  CPU 60%, Memory 50%, I/O 40%  → Balanced, no single bottleneck
```

This is called **bottleneck shifting** and is a sign of progress.

---

## Systematic Diagnosis Framework

### The Universal 5-Step Diagnosis Process

No matter what type of bottleneck you suspect, follow this process:

```
Step 1: OBSERVE    → What are the symptoms? (slow, OOM, high CPU?)
Step 2: MEASURE    → Quantify with metrics (CPU%, memory MB, latency ms)
Step 3: PROFILE    → Find the specific function/query/call causing it
Step 4: HYPOTHESIZE → Form a theory about the root cause
Step 5: FIX & VERIFY → Apply fix, measure again to confirm improvement
```

### Step 1: OBSERVE — Recognize the Symptoms

```
Symptom                          → Likely Bottleneck
─────────────────────────────────────────────────────
High CPU, low I/O wait           → CPU-bound
Low CPU, high I/O wait           → I/O-bound (disk)
Low CPU, slow network calls      → I/O-bound (network)
Memory growing over time         → Memory leak
OOM kills                        → Memory capacity
Slow after hours of running      → Memory leak or resource leak
Fast with 1 user, slow with 100  → Concurrency/contention
Slow queries in logs             → Database I/O or missing index
High GC pause times              → Memory allocation pressure
```

### Step 2: MEASURE — Quantify the Problem

Always establish a baseline before making changes. You cannot improve what you cannot measure.

```go
// Instrument your code with metrics from day one
type Metrics struct {
    RequestDuration  prometheus.Histogram
    RequestsTotal    prometheus.Counter
    ErrorsTotal      prometheus.Counter
    ActiveRequests   prometheus.Gauge
    DBQueryDuration  prometheus.Histogram
    CacheHitRate     prometheus.Gauge
}

func (m *Metrics) MeasureRequest(handler func() error) error {
    m.ActiveRequests.Inc()
    defer m.ActiveRequests.Dec()

    start := time.Now()
    err := handler()
    duration := time.Since(start)

    m.RequestDuration.Observe(duration.Seconds())
    m.RequestsTotal.Inc()
    if err != nil {
        m.ErrorsTotal.Inc()
    }
    return err
}
```

### Step 3: PROFILE — Find the Exact Cause

```bash
# CPU profiling
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/goroutine

# Block profiling (I/O waits, channel waits)
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/block

# Full execution trace
curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out
go tool trace trace.out
```

### Step 4: HYPOTHESIZE — Form a Theory

Ask these questions based on your profile data:

**For CPU:**
- Which function appears at the top of the CPU profile?
- Is it an algorithm issue (O(n²))? A serialization issue? GC?
- Can this work be parallelized?
- Can results be cached?

**For Memory:**
- Is heap growing over time (leak) or just large (capacity)?
- Which allocation site is responsible (heap profile)?
- Are goroutines leaking (goroutine profile)?
- Is GC running too frequently?

**For I/O:**
- Is it disk or network?
- Are calls serial that could be parallel?
- Is there a missing cache?
- Is there a missing index?
- Is connection pooling configured correctly?

### Step 5: FIX & VERIFY — Confirm the Improvement

```go
// Before fix: establish baseline
// Run benchmark or load test, record metrics

// Apply fix

// After fix: run same benchmark
// Compare: latency p50, p95, p99, throughput, error rate

// Example benchmark comparison
func BenchmarkBefore(b *testing.B) {
    for i := 0; i < b.N; i++ {
        oldImplementation()
    }
}

func BenchmarkAfter(b *testing.B) {
    for i := 0; i < b.N; i++ {
        newImplementation()
    }
}

// go test -bench=. -benchmem -count=5 | tee results.txt
// benchstat results.txt  (shows statistical comparison)
```

---

## Quick Reference: Symptoms → Diagnosis → Fix

### CPU-Bound Quick Reference

```
SYMPTOM                    CAUSE                        FIX
────────────────────────────────────────────────────────────────────────
CPU 100%, low iowait       Inefficient algorithm        Optimize O(n²) → O(n)
High GC CPU usage          Too many allocations         sync.Pool, reduce allocs
Slow JSON processing       Repeated marshal/unmarshal   Cache, faster library
Regex slow                 Compiling in hot path        Compile once at init
String concat slow         + operator in loop           strings.Builder
High context switch        Too many goroutines          Worker pool
Crypto operations slow     bcrypt/TLS overhead          Tune cost, hardware accel
```

### Memory-Bound Quick Reference

```
SYMPTOM                    CAUSE                        FIX
────────────────────────────────────────────────────────────────────────
Heap grows forever         Goroutine leak               Context + cancel
Heap grows forever         Map never shrinks            Replace map after bulk delete
Heap grows forever         Unbounded cache              LRU/TTL eviction
OOM kill                   Large allocations            sync.Pool, streaming
High GC frequency          Many small allocations       Pre-allocate, value types
Slow after hours           Memory leak                  pprof heap comparison
Goroutine count grows      Missing context cancel       Always defer cancel()
Swap usage                 Insufficient RAM             Increase RAM or reduce usage
```

### I/O-Bound Quick Reference

```
SYMPTOM                    CAUSE                        FIX
────────────────────────────────────────────────────────────────────────
High iowait                Unbuffered disk writes       bufio.Writer
Slow DB queries            Missing index                CREATE INDEX
Slow DB queries            N+1 queries                  JOIN or batch load
Slow DB queries            No connection pool           sql.DB with pool config
Slow API calls             Serial I/O                   Parallel with errgroup
Slow API calls             No caching                   Redis/in-memory cache
Network timeout            No timeout configured        context.WithTimeout
High DB connections        Connection leak              defer rows.Close()
Slow file processing       Loading entire file          Stream with bufio.Scanner
```

---

## Tool Reference Guide

### System-Level Tools

| Tool | OS | Purpose | Key Command |
|------|----|---------|-------------|
| `top` | All | CPU, memory overview | `top -1` (Linux) |
| `htop` | Linux/macOS | Interactive process viewer | `htop` |
| `vmstat` | Linux | Virtual memory stats | `vmstat 1` |
| `iostat` | Linux/macOS | Disk I/O stats | `iostat -x 1` |
| `iotop` | Linux | Per-process disk I/O | `iotop -o` |
| `netstat` | All | Network connections | `netstat -an` |
| `ss` | Linux | Socket statistics | `ss -s` |
| `iftop` | Linux/macOS | Network usage per connection | `iftop` |
| `lsof` | All | Open files and connections | `lsof -p <pid>` |
| `strace` | Linux | Syscall tracing | `strace -p <pid>` |
| `dtruss` | macOS | Syscall tracing | `sudo dtruss -p <pid>` |
| `perf` | Linux | Hardware performance counters | `perf stat ./app` |
| `tcpdump` | All | Network packet capture | `tcpdump -i eth0` |
| `free` | Linux | Memory overview | `free -h` |
| `uptime` | All | Load average | `uptime` |

### Go-Specific Tools

| Tool | Purpose | How to Use |
|------|---------|-----------|
| `pprof CPU` | CPU flame graph | `go tool pprof .../profile` |
| `pprof heap` | Memory allocation | `go tool pprof .../heap` |
| `pprof goroutine` | Goroutine stacks | `go tool pprof .../goroutine` |
| `pprof block` | Blocking events | `go tool pprof .../block` |
| `pprof mutex` | Mutex contention | `go tool pprof .../mutex` |
| `go trace` | Full execution trace | `go tool trace trace.out` |
| `go test -bench` | Micro-benchmarks | `go test -bench=. -benchmem` |
| `benchstat` | Benchmark comparison | `benchstat old.txt new.txt` |
| `goleak` | Goroutine leak detection | `goleak.VerifyNone(t)` |
| `GODEBUG=gctrace` | GC statistics | env var |
| `go build -gcflags="-m"` | Escape analysis | build flag |
| `runtime.MemStats` | Programmatic memory stats | Go API |

### Database Tools

| Tool | DB | Purpose | Command |
|------|----|---------|---------|
| `EXPLAIN ANALYZE` | PostgreSQL/MySQL | Query execution plan | SQL |
| `pg_stat_statements` | PostgreSQL | Slow query tracking | Extension |
| `slow_query_log` | MySQL | Log slow queries | Config |
| `pg_stat_activity` | PostgreSQL | Active connections | SQL |
| `SHOW PROCESSLIST` | MySQL | Active queries | SQL |
| `redis-cli monitor` | Redis | Real-time commands | CLI |
| `redis-cli info` | Redis | Memory and stats | CLI |

### Load Testing Tools

| Tool | Purpose | Command |
|------|---------|---------|
| `wrk` | HTTP load testing | `wrk -t4 -c100 -d30s http://...` |
| `ab` | Apache Bench | `ab -n 10000 -c 100 http://...` |
| `hey` | Go HTTP load tester | `hey -n 10000 -c 100 http://...` |
| `k6` | Scriptable load testing | `k6 run script.js` |
| `vegeta` | HTTP load testing | `echo "GET http://..." | vegeta attack` |

---

## Key Principles to Remember

### 1. Measure First, Optimize Second

Never guess at bottlenecks. Always profile before optimizing. The bottleneck is almost never where you think it is.

> "Premature optimization is the root of all evil." — Donald Knuth

### 2. Amdahl's Law — The Limit of Parallelism

Even if you parallelize perfectly, the serial portion of your code limits speedup.

```
Speedup = 1 / (S + (1-S)/N)

Where:
  S = fraction of code that is serial (cannot be parallelized)
  N = number of processors

Example: 20% serial code, 8 cores
  Speedup = 1 / (0.2 + 0.8/8) = 1 / (0.2 + 0.1) = 3.3x
  (not 8x as you might hope)
```

### 3. The 80/20 Rule of Performance

80% of performance problems come from 20% of the code. Find that 20% with profiling.

### 4. Bottleneck Shifting is Progress

When you fix one bottleneck and a new one appears, that is success. Keep iterating.

### 5. Caching is the Universal Accelerator

For I/O-bound systems, caching is almost always the highest-leverage fix. But caches introduce consistency challenges — always consider TTL and invalidation.

### 6. Concurrency is Not Parallelism

- **Concurrency**: Dealing with multiple things at once (Go goroutines handle I/O concurrently)
- **Parallelism**: Doing multiple things at once (multiple CPU cores computing simultaneously)
- I/O-bound: concurrency helps (more goroutines waiting in parallel)
- CPU-bound: parallelism helps (more cores computing in parallel)

### 7. The Three Knobs

Every performance problem can be addressed by tuning one of three knobs:
1. **Do less work** — cache, skip, batch, deduplicate
2. **Do work faster** — better algorithm, hardware, parallelism
3. **Do work later** — async, queue, defer

