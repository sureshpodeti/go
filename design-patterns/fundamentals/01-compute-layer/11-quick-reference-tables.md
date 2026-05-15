# Performance Bottleneck Master Reference

> One file to diagnose and fix CPU, Memory, and I/O bottlenecks.
> Structure: What is it → What processes cause it → Symptoms → Metrics → Tools → Fix → Standard/Convention

---

## PART 1 — CPU-BOUND

### What is CPU-Bound?
The processor is the bottleneck. The program spends time **computing**, not waiting.
Adding more CPU cores or a faster CPU directly improves performance.

---

### 1A. Processes / Operations That Are CPU-Bound (and Why)

| Process / Operation | Why It Is CPU-Bound |
|---------------------|---------------------|
| **Sorting large datasets** | Comparison-based sort = O(n log n) comparisons, all CPU work |
| **JSON / XML marshal & unmarshal** | Reflection + string parsing on every byte = pure CPU |
| **Image / video encoding (JPEG, H.264)** | Pixel math, DCT transforms, compression = heavy computation |
| **Cryptography (bcrypt, AES, RSA, TLS handshake)** | Intentionally expensive math (bcrypt) or modular arithmetic (RSA) |
| **Regular expression matching** | NFA/DFA state machine evaluation on every character |
| **Data compression (gzip, zstd, snappy)** | Sliding window search, entropy coding = CPU intensive |
| **Hash computation (SHA256, MD5)** | Bitwise operations over entire input |
| **Machine learning inference** | Matrix multiplications over millions of weights |
| **Template rendering (HTML, text/template)** | Parsing + string building on every render |
| **String concatenation in loops** | Each `+` creates a new string copy = O(n²) total |
| **Garbage collection** | GC scans all live objects = CPU time proportional to heap size |
| **Reflection (`reflect` package)** | Runtime type inspection is 10-100x slower than direct access |
| **Base64 encode/decode** | Bit manipulation over entire payload |
| **Protobuf / msgpack serialization** | Field encoding/decoding per message |
| **Number parsing (strconv.Atoi, ParseFloat)** | Character-by-character conversion |
| **Recursive algorithms without memoization** | Recomputes same subproblems repeatedly |
| **Bloom filter / HyperLogLog operations** | Hash function calls per operation |
| **Rate limiter token refill (tight loop)** | Continuous time checks and atomic operations |

---

### 1B. Symptoms of CPU-Bound

| Symptom | What You See | Why It Happens |
|---------|-------------|----------------|
| High CPU `%us` | `top` shows 80-100% user CPU | Your code is computing non-stop |
| High load average | `uptime` shows load > num cores | More runnable goroutines than CPU slots |
| Low `%wa` (iowait) | Near 0% iowait | CPU is busy, not waiting for disk/network |
| Slow response even with no DB/network | Latency high but no external calls | Pure computation is the bottleneck |
| All CPU cores maxed | `htop` shows all bars full | Work is parallelized but still not enough CPU |
| GC pauses in logs | `GODEBUG=gctrace` shows frequent GC | Too many allocations → GC steals CPU |
| Goroutines stuck in `running` state | `pprof goroutine` shows many `running` | Goroutines competing for CPU time |
| Throughput plateaus despite more goroutines | Adding goroutines doesn't help | CPU is already saturated |
| Thermal throttling (physical machines) | CPU clock speed drops | CPU overheating, reduces frequency |

---

### 1C. Metrics That Confirm CPU-Bound

| Metric | Tool | CPU-Bound Value | How to Read |
|--------|------|----------------|-------------|
| `%us` (user CPU) | `top` / `htop` | > 80% | Time in your application code |
| `%sy` (system CPU) | `top` | > 20% | Too many syscalls (also CPU cost) |
| `%id` (idle) | `top` | < 10% | CPU has no free time |
| `%wa` (iowait) | `top` | < 5% | Not waiting for I/O (rules out I/O-bound) |
| Load average | `uptime` | > num_cores | e.g. load=15 on 8-core = overloaded |
| `HeapAlloc` growth | `runtime.MemStats` | Rapid growth | Allocations driving GC CPU usage |
| `NumGC` | `runtime.MemStats` | High count | GC running frequently = CPU cost |
| `PauseTotalNs` | `runtime.MemStats` | > 1% of runtime | GC pauses stealing CPU |
| CPU profile top function | `pprof` | Your function at top | Confirms which code is hot |
| ns/op in benchmark | `go test -bench` | High value | Quantifies CPU cost per operation |

```bash
# Read these in top:
# %Cpu(s): 94.2 us,  3.1 sy,  0.0 ni,  2.1 id,  0.0 wa
#           ^^^^                         ^^^       ^^^
#           94% in your code             2% idle   0% iowait = CPU-bound confirmed

# Load average check:
uptime
# 15:04:01 up 3 days, load average: 14.8, 13.2, 12.1
# On 8-core machine: 14.8 >> 8 = severely CPU-bound
```

---

### 1D. Per-Process: What Goes Wrong, How to Identify, How to Fix

| Process | What Goes Wrong | How to Identify | Tool | Fix | Convention / Standard |
|---------|----------------|-----------------|------|-----|----------------------|
| **Sorting** | O(n²) sort (bubble/insertion) instead of O(n log n) | `pprof` shows sort function at top | `go test -bench` | Use `sort.Slice` (introsort, O(n log n)) | Always use stdlib sort; only write custom sort for specific data properties |
| **JSON** | `json.Marshal/Unmarshal` called on every request | `pprof` shows `encoding/json` at top | `go test -bench -benchmem` | Use `jsoniter`, `easyjson`, or cache serialized result | Pre-serialize static responses; use protobuf for internal services |
| **Regex** | `regexp.MustCompile(...)` inside function body | `pprof` shows `regexp.Compile` | Code review | Compile once: `var re = regexp.MustCompile(...)` at package level | Always declare regex as package-level `var` |
| **Crypto / bcrypt** | Cost factor too high (14+) for your hardware | Benchmark shows > 500ms per hash | `go test -bench=BenchmarkBcrypt` | Tune cost to ~10; use Argon2 with tuned params | bcrypt cost=10 is industry standard; benchmark on your hardware |
| **String concat** | `result += str` in a loop | `pprof` shows `runtime.concatstrings` | `go test -bench -benchmem` | `strings.Builder` with `Grow(estimatedSize)` | Always use `strings.Builder` for 3+ concatenations |
| **Template rendering** | `template.Parse(...)` on every request | `pprof` shows `text/template.Parse` | Code review | Parse once at startup, `Execute` per request | Parse templates at `init()` or startup, never in handlers |
| **Reflection** | `reflect.ValueOf(...)` in hot path | `pprof` shows `reflect.*` | Code review | Use direct struct access or code generation | Avoid `reflect` in any path called > 1000x/sec |
| **GC pressure** | Millions of small short-lived allocations | `GODEBUG=gctrace=1` shows frequent GC | `pprof heap` | `sync.Pool`, pre-allocate, value types | Pool objects > 1KB that are allocated > 1000x/sec |
| **Compression** | Compressing every response including small ones | `pprof` shows `compress/gzip` at top | Benchmark with/without | Only compress responses > 1KB; use faster `zstd` or `snappy` | HTTP: compress > 1KB; internal: use `snappy` for speed |
| **Image processing** | Processing images synchronously in request handler | High CPU, slow HTTP responses | `pprof` CPU profile | Offload to worker pool; use `runtime.NumCPU()` workers | Never process images in HTTP handler; always use async worker pool |
| **Unbounded goroutines** | `go process(item)` in a loop with no limit | Goroutine count in millions | `runtime.NumGoroutine()` | Worker pool with `runtime.NumCPU()` workers | Max goroutines for CPU work = num CPU cores |

---

### 1E. CPU-Bound Debugging Workflow

```
1. Confirm CPU-bound:
   top → %us > 80%, %wa < 5%, load > num_cores

2. Find the hot function:
   go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
   → Look at flame graph top → that function is your bottleneck

3. Quantify with benchmark:
   go test -bench=BenchmarkHotFunction -benchmem -count=5

4. Apply fix

5. Re-benchmark and compare:
   benchstat before.txt after.txt
```

---

### 1F. CPU-Bound Fix Patterns & Conventions

| Pattern | Use When | Convention |
|---------|----------|-----------|
| Worker pool (N = NumCPU) | CPU-intensive parallel work | `workers := runtime.NumCPU()` |
| `sync.Pool` | Object allocated and freed frequently | Reset object state before `Put()` |
| `strings.Builder` | String building in loops | `sb.Grow(estimatedBytes)` before writing |
| Package-level `var` for regex | Regex used in any function | `var re = regexp.MustCompile(...)` |
| Memoization / result cache | Pure function with repeated inputs | Use `sync.Map` or `lru.Cache` with TTL |
| `GOGC=200` | GC too frequent, can afford more memory | Set in env or `debug.SetGCPercent(200)` |
| Faster serializer | JSON in hot path (> 10K req/s) | `jsoniter` drop-in; protobuf for internal |
| Pre-allocate slices | Known or estimated size | `make([]T, 0, capacity)` |
| Value types over pointers | Small structs (< 128 bytes) | Avoids heap allocation, stays on stack |
| Benchmark before optimizing | Always | `go test -bench=. -benchmem -count=5` |


---

## PART 2 — MEMORY-BOUND

### What is Memory-Bound?
RAM is the bottleneck. Either the program **runs out of memory** (capacity) or the CPU spends time **waiting for data from RAM** (bandwidth). In Go, this also includes GC overhead from excessive allocations.

---

### 2A. Processes / Operations That Are Memory-Bound (and Why)

| Process / Operation | Why It Is Memory-Bound |
|---------------------|------------------------|
| **Loading large files into RAM** | `os.ReadFile` on a 2GB file puts 2GB in heap |
| **In-memory caching without eviction** | Cache grows unboundedly until OOM |
| **Building large data structures** | Trees, graphs, tries with millions of nodes |
| **Batch processing large datasets** | Loading all records before processing |
| **Session storage (in-memory)** | Each user session holds state in RAM |
| **Image/video decoding** | Decoded frames are much larger than compressed (1080p = 6MB/frame) |
| **String interning issues** | Storing millions of duplicate strings |
| **Goroutine leaks** | Each goroutine holds ~2-8KB stack + any captured variables |
| **Unbounded channel buffers** | `make(chan T, 1_000_000)` allocates upfront |
| **ORM with eager loading** | Loads entire object graph into memory |
| **Log buffering** | Buffering logs in memory before flushing |
| **Connection state** | Each open connection holds read/write buffers |
| **Recursive algorithms with deep stacks** | Each stack frame holds local variables |
| **Map with many entries** | Go maps never shrink; buckets stay allocated |
| **Slice operations retaining backing array** | `data[:10]` keeps entire original array alive |

---

### 2B. Symptoms of Memory-Bound

| Symptom | What You See | Why It Happens |
|---------|-------------|----------------|
| Heap grows over time | `pprof heap` shows steady growth | Memory leak — objects not being freed |
| OOM kill | Process exits with `signal: killed` or `fatal error: runtime: out of memory` | Heap exceeded available RAM |
| Swap in use | `free -h` shows swap > 0 | RAM exhausted, OS using disk as RAM (1000x slower) |
| Frequent GC | `GODEBUG=gctrace` shows GC every few seconds | Too many allocations, GC can't keep up |
| High GC CPU | `pprof` shows `runtime.gcBgMarkWorker` at top | GC consuming significant CPU |
| Goroutine count grows | `runtime.NumGoroutine()` increases over time | Goroutine leak |
| Performance degrades over hours | App starts fast, slows down after hours | Memory leak causing GC to work harder |
| `HeapObjects` very high | `runtime.MemStats.HeapObjects` in millions | Too many live objects for GC to track |
| RSS grows but heap stable | OS-level memory grows, Go heap stable | Memory fragmentation or cgo leak |

---

### 2C. Metrics That Confirm Memory-Bound

| Metric | Tool | Memory-Bound Value | How to Read |
|--------|------|-------------------|-------------|
| `HeapAlloc` | `runtime.MemStats` | Growing over time | Current heap in use; leak if never decreases |
| `HeapSys` | `runtime.MemStats` | Near system RAM | Total heap obtained from OS |
| `HeapObjects` | `runtime.MemStats` | Millions | Number of live objects GC must track |
| `NumGC` | `runtime.MemStats` | Very high (1000s) | How many GC cycles have run |
| `PauseTotalNs` | `runtime.MemStats` | > 1s total | Total time stopped for GC |
| `NextGC` | `runtime.MemStats` | Low value | GC will trigger soon (heap nearly full) |
| Goroutine count | `runtime.NumGoroutine()` | Growing | Goroutine leak |
| RSS (Resident Set Size) | `ps`, `/proc/<pid>/status` | Near system RAM | Actual RAM used by process |
| Swap used | `free -h` | Any > 0 | Severe: OS swapping to disk |

```bash
# Check memory in Go:
curl http://localhost:6060/debug/pprof/heap > heap1.prof
# ... wait 5 minutes ...
curl http://localhost:6060/debug/pprof/heap > heap2.prof
go tool pprof -base heap1.prof heap2.prof
(pprof) top10   # shows what GREW = the leak

# Check goroutine count:
curl http://localhost:6060/debug/pprof/goroutine?debug=1 | head -5
# goroutine profile: total 45821   ← 45K goroutines = likely leak

# GC trace:
GODEBUG=gctrace=1 ./myapp 2>&1 | head -10
# gc 847 @120.3s 18%: ...   ← 18% of time in GC = memory pressure
```

---

### 2D. Per-Process: What Goes Wrong, How to Identify, How to Fix

| Process | What Goes Wrong | How to Identify | Tool | Fix | Convention / Standard |
|---------|----------------|-----------------|------|-----|----------------------|
| **Goroutines** | Started but never exit (no context, no channel close) | Goroutine count grows in `pprof goroutine` | `goleak.VerifyNone(t)` in tests | Always pass `context.Context`; `defer cancel()` | Every goroutine must have an exit condition |
| **HTTP handlers** | Goroutine per request leaks if request context not respected | Goroutine count grows under load | `pprof goroutine` | Respect `ctx.Done()` in all blocking operations | Use `context.WithTimeout` on all outbound calls |
| **In-memory cache** | No eviction policy; grows until OOM | `HeapAlloc` grows monotonically | `pprof heap` | LRU with max size or TTL expiry | Max cache size = 20-30% of available RAM |
| **Slices** | `data[:n]` keeps entire backing array alive | `pprof heap` shows large arrays | Code review | `copy()` to new slice when keeping small portion | Always copy when keeping < 10% of original slice |
| **Maps** | Bulk delete doesn't shrink map buckets | `HeapAlloc` stays high after deletes | `runtime.MemStats` | Replace map: `m = make(map[K]V)` | Rebuild map after deleting > 50% of entries |
| **Channels** | Unbuffered or over-buffered channels | `make(chan T, 1_000_000)` in code | Code review | Size buffer to expected burst, not max possible | Buffer = expected burst size, not unlimited |
| **Closures** | Capture large variables by reference | `pprof heap` shows unexpected retention | Code review | Set large vars to `nil` before returning closure | Only capture what the closure actually needs |
| **sync.Pool** | Not used for frequently allocated objects | `pprof heap` shows many same-type allocs | `go test -benchmem` | Add `sync.Pool` for objects > 1KB allocated > 1K/s | Always reset pooled objects before `Put()` |
| **Buffers** | New `make([]byte, N)` per request | `pprof heap` shows `[]byte` at top | `go test -benchmem` | `sync.Pool` for byte slices | Pool buffers > 4KB used in hot paths |
| **String storage** | Storing millions of duplicate strings | High `HeapObjects`, many small strings | `pprof heap` | Use `iota` enums or string interning | Use typed constants for repeated string values |
| **ORM / DB results** | Loading entire result set into memory | `HeapAlloc` spikes on DB queries | `pprof heap` | Stream with `rows.Next()`, process one at a time | Never `rows.Scan` into a slice of millions |
| **Log buffers** | Buffering logs in memory without flush | Memory grows, logs lost on crash | Code review | Flush periodically; use async logger with bounded buffer | Max log buffer = 10K entries or 10MB |

---

### 2E. Memory-Bound Debugging Workflow

```
1. Confirm memory-bound:
   free -h → is swap used? Is available RAM < 20%?
   GODEBUG=gctrace=1 → is GC running every few seconds?

2. Is it a leak or just large?
   Watch HeapAlloc over time:
   - Growing continuously = LEAK
   - Large but stable = CAPACITY issue

3. Find the leak source:
   Take two heap profiles 5 minutes apart:
   go tool pprof -base heap1.prof heap2.prof
   (pprof) top10  → shows what grew

4. Check for goroutine leak:
   curl .../goroutine?debug=1 | head -3
   → If count > 10K and growing = goroutine leak

5. Apply fix

6. Verify: heap should stabilize, goroutine count should plateau
```

---

### 2F. Memory-Bound Fix Patterns & Conventions

| Pattern | Use When | Convention |
|---------|----------|-----------|
| `context` + `defer cancel()` | Any goroutine that could block | Every goroutine must have exit path |
| `sync.Pool` | Objects > 1KB, allocated > 1K/sec | Reset state before `Put()`; don't store pointers to pooled objects |
| LRU cache (`lru.New(N)`) | Cache without eviction | Max size = 20-30% of available RAM |
| TTL cache | Data that becomes stale | TTL = data freshness requirement |
| `copy()` for sub-slices | Keeping small portion of large slice | Copy when keeping < 10% of original |
| Replace map after bulk delete | Map won't shrink | `m = make(map[K]V)` after > 50% delete |
| `debug.SetMemoryLimit(N)` | Hard cap on memory (Go 1.19+) | Set to 80% of container memory limit |
| Stream instead of load-all | Large files or DB result sets | `bufio.Scanner` / `rows.Next()` |
| Value types for small structs | Structs < 128 bytes in hot path | Avoids heap allocation |
| Pre-allocate with known capacity | Slice size known or estimable | `make([]T, 0, n)` |
| `goleak` in tests | Catch goroutine leaks early | `defer goleak.VerifyNone(t)` in every test |


---

## PART 3 — I/O-BOUND

### What is I/O-Bound?
The bottleneck is **waiting for external data** — disk reads/writes, network calls, database queries. The CPU is mostly idle. Adding more CPU does nothing; reducing wait time (caching, parallelism, batching) is the fix.

---

### 3A. Processes / Operations That Are I/O-Bound (and Why)

| Process / Operation | Why It Is I/O-Bound |
|---------------------|---------------------|
| **Database queries** | Network round-trip + disk read for each query |
| **HTTP API calls** | Network latency (1ms local, 50-200ms internet) per call |
| **File reads / writes** | Disk speed (HDD: 100MB/s, SSD: 500MB/s) vs RAM (50GB/s) |
| **Log writing** | Frequent small writes to disk |
| **Redis / Memcached calls** | Network round-trip per command (~0.5-1ms) |
| **Message queue (Kafka, RabbitMQ)** | Network + disk for produce/consume |
| **DNS resolution** | Network call to DNS server per hostname |
| **TLS handshake** | Multiple network round-trips to establish connection |
| **gRPC / REST calls to microservices** | Network latency per service hop |
| **Object storage (S3, GCS)** | Network + remote disk per object |
| **Email / SMS sending** | External API call per message |
| **Webhook delivery** | HTTP call per event |
| **Config loading from remote** | Network call to config service |
| **Health checks** | Network call per check interval |
| **Streaming large responses** | Sustained network bandwidth |

---

### 3B. Symptoms of I/O-Bound

| Symptom | What You See | Why It Happens |
|---------|-------------|----------------|
| High `%wa` (iowait) | `top` shows > 20% iowait | CPU waiting for disk I/O to complete |
| Low CPU, slow responses | CPU < 20% but requests are slow | CPU is idle, waiting for network/disk |
| Disk `%util` = 100% | `iostat` shows disk fully busy | Disk cannot handle request rate |
| High disk `await` | `iostat` shows await > 50ms | Requests queuing up at disk |
| DB connection pool exhausted | `db.Stats().WaitCount > 0` | All connections in use, new requests wait |
| Goroutines in `chan receive` | `pprof goroutine` shows many waiting | Goroutines blocked waiting for I/O result |
| Slow queries in DB logs | Queries taking > 100ms | Missing index or full table scan |
| Network queue full | `ss` shows large `Recv-Q` or `Send-Q` | Network buffer backed up |
| Timeout errors | `context deadline exceeded` | I/O took longer than timeout |
| High `await` in `iostat` | > 10ms for SSD, > 50ms for HDD | Disk queue depth growing |

---

### 3C. Metrics That Confirm I/O-Bound

| Metric | Tool | I/O-Bound Value | How to Read |
|--------|------|----------------|-------------|
| `%wa` (iowait) | `top` | > 20% | CPU idle waiting for disk |
| `%util` | `iostat -x` | > 80% | Disk is saturated |
| `await` (ms) | `iostat -x` | > 10ms SSD, > 50ms HDD | Average I/O wait time |
| `r/s`, `w/s` | `iostat -x` | Near device limit | Reads/writes per second |
| `WaitCount` | `db.Stats()` | > 0 | Requests waiting for DB connection |
| `WaitDuration` | `db.Stats()` | > 1ms avg | Time spent waiting for connection |
| Goroutines in syscall | `pprof goroutine` | Many in `syscall` | Blocked on disk/network |
| Block profile | `pprof block` | High cumulative time | Where goroutines spend time waiting |
| Network RTT | `ping`, `curl -w` | > 10ms internal | Network latency is significant |
| Query time | DB slow query log | > 100ms | Query needs optimization |

```bash
# Confirm I/O-bound with iostat:
iostat -x 1
# Device  r/s   w/s   rkB/s  wkB/s  await  %util
# nvme0n1 0.0  2400   0.0    9600   85ms   100%
#                                    ^^^^   ^^^^
#                                    85ms wait, 100% busy = I/O bottleneck

# Check DB pool:
# In Go code:
stats := db.Stats()
fmt.Printf("WaitCount: %d, WaitDuration: %v\n", stats.WaitCount, stats.WaitDuration)
# WaitCount: 847 = 847 requests had to wait for a connection

# Check goroutines blocked on I/O:
curl http://localhost:6060/debug/pprof/block > block.prof
go tool pprof block.prof
(pprof) top10   # shows where goroutines spend time blocked
```

---

### 3D. Per-Process: What Goes Wrong, How to Identify, How to Fix

| Process | What Goes Wrong | How to Identify | Tool | Fix | Convention / Standard |
|---------|----------------|-----------------|------|-----|----------------------|
| **Database queries** | No index on WHERE/JOIN columns | `EXPLAIN ANALYZE` shows `Seq Scan` | `pg_stat_statements`, `EXPLAIN ANALYZE` | `CREATE INDEX` on query columns | Index all foreign keys and frequently filtered columns |
| **Database queries** | N+1 problem (1 query per row) | Slow page load; DB shows 1000s of identical queries | DB slow query log, `pg_stat_statements` | Single `JOIN` or batch `WHERE id IN (...)` | ORM: use `Preload`/`Eager`; raw SQL: always JOIN |
| **Database queries** | No connection pool | New TCP connection per query | `lsof` shows many short-lived connections | `sql.DB` with `SetMaxOpenConns(25)` | Pool: max_open=25, max_idle=10, lifetime=5min |
| **Database queries** | Connection leak | Pool exhausted; `WaitCount` grows | `db.Stats()`, `pg_stat_activity` | `defer rows.Close()`, `defer stmt.Close()`, `defer tx.Rollback()` | Always defer Close/Rollback immediately after open |
| **HTTP API calls** | Serial calls (one after another) | Total latency = sum of all calls | `pprof block`, request traces | `errgroup` for parallel calls | Independent calls must always be parallel |
| **HTTP API calls** | No timeout configured | Requests hang indefinitely | Hanging goroutines in `pprof goroutine` | `http.Client{Timeout: 5s}` + `context.WithTimeout` | Always set both client timeout AND context timeout |
| **HTTP API calls** | No connection reuse | New TCP+TLS per request | `tcpdump` shows many SYN packets | Reuse `http.Client` (has built-in pool) | Never create `http.Client` per request; use package-level var |
| **File reads** | No buffering | One syscall per byte/line | `strace` shows millions of `read()` calls | `bufio.NewReader(f)` or `bufio.NewScanner(f)` | Always wrap file I/O in `bufio` |
| **File writes** | Unbuffered + `Sync()` on every write | High `%wa`, slow throughput | `iostat`, `strace` | `bufio.NewWriter` + flush periodically | Flush every 1s or every 1000 writes; `Sync()` only on shutdown |
| **File reads** | Loading entire file into memory | OOM on large files | `pprof heap` shows large `[]byte` | Stream with `bufio.Scanner` | Never `os.ReadFile` on files > 100MB |
| **Redis** | One command per item in loop | High latency, many round-trips | `redis-cli monitor`, request traces | Pipeline: `pipe := client.Pipeline()` | Batch all independent Redis commands in one pipeline |
| **Redis** | No connection pool | New TCP per command | `redis-cli info clients` shows many connections | Use `go-redis` client (has built-in pool) | Pool size = expected concurrent goroutines |
| **Log writing** | Synchronous write per log line | High `%wa`, slow request handling | `iostat`, `strace` | Async logger with buffered channel | Buffer 10K log entries; flush every 1s |
| **Microservice calls** | No circuit breaker | Slow service causes cascade failure | Latency spikes, timeout errors | Circuit breaker (e.g. `gobreaker`) | Open circuit after 5 consecutive failures; half-open after 30s |
| **Microservice calls** | No retry with backoff | Transient failures cause errors | Error rate spikes | Exponential backoff: 100ms, 200ms, 400ms, 800ms | Max 3 retries; only retry idempotent operations |
| **DNS** | Resolving same hostname repeatedly | Latency on every new connection | `strace` shows `getaddrinfo` calls | DNS caching in HTTP transport | Go's `net.DefaultResolver` caches; ensure `DialContext` reuse |
| **Object storage (S3)** | Downloading large objects fully | Slow, high memory | Request traces | Range requests for partial reads | Use `Range: bytes=0-1023` for partial reads |

---

### 3E. I/O-Bound Debugging Workflow

```
1. Confirm I/O-bound:
   top → CPU < 20%, %wa > 20%  (disk I/O)
   top → CPU < 20%, %wa < 5%   (network I/O — CPU idle but not disk wait)

2. Is it disk or network?
   Disk: iostat -x 1 → %util > 80%, await > 10ms
   Network: pprof block → goroutines blocked in net.(*netFD).Read

3. Find the specific bottleneck:
   Disk: strace -p <pid> -e trace=read,write → which files?
   Network: pprof block → which function is blocking?
   DB: pg_stat_statements → which query is slowest?

4. Apply fix (cache, parallelize, index, pool, buffer)

5. Verify:
   Disk: iostat shows %util drops
   Network: pprof block shows less blocking time
   DB: EXPLAIN ANALYZE shows Index Scan instead of Seq Scan
```

---

### 3F. I/O-Bound Fix Patterns & Conventions

| Pattern | Use When | Convention |
|---------|----------|-----------|
| `errgroup` parallel I/O | Multiple independent I/O calls | All independent calls must be parallel |
| Multi-level cache (L1→L2→L3) | Repeated reads of same data | L1=map(1ms), L2=Redis(5ms), L3=DB(50ms) |
| `bufio.NewWriterSize(f, 65536)` | Any file write | Always buffer file I/O; 64KB is standard |
| `bufio.NewScanner(f)` | Reading large files | Never load > 100MB into memory at once |
| `sql.DB` connection pool | Any database usage | max_open=25, max_idle=10, lifetime=5min |
| `defer rows.Close()` | Any `db.Query()` call | Immediately after `rows, err := db.Query(...)` |
| `http.Client` package-level var | Any HTTP calls | One client per service, never per request |
| `context.WithTimeout` | All outbound I/O | DB: 5s, internal API: 2s, external API: 10s |
| Redis pipeline | Multiple Redis commands | Batch all independent commands |
| `CREATE INDEX` | Slow DB queries | Index every WHERE, JOIN ON, ORDER BY column |
| Circuit breaker | Calls to external services | Open after 5 failures; half-open after 30s |
| Exponential backoff | Retrying failed calls | Base 100ms, max 3 retries, jitter ±10% |
| Async write (buffered channel) | Non-critical writes (logs, analytics) | Buffer = 10K items; drop on full (backpressure) |
| Read replica | Read-heavy DB workload | Route all SELECTs to replica |
| Streaming response | Large payloads | `json.NewEncoder(w).Encode()` not `json.Marshal` |


---

## PART 4 — MASTER COMPARISON TABLES

### 4A. Side-by-Side: What Each Bound Looks Like

| Dimension | CPU-Bound | Memory-Bound | I/O-Bound |
|-----------|-----------|--------------|-----------|
| **Bottleneck resource** | Processor cores | RAM / heap | Disk or network |
| **CPU `%us`** | HIGH (80-100%) | Medium (30-60%) | LOW (2-20%) |
| **CPU `%wa`** | Low (< 5%) | Low (< 5%) | HIGH (> 20%) disk |
| **CPU idle `%id`** | LOW (< 10%) | Medium | HIGH (> 60%) |
| **Memory usage** | Normal / stable | HIGH / growing | Normal / stable |
| **Swap usage** | None | HIGH (severe) | None |
| **Disk `%util`** | Low | Low | HIGH (> 80%) |
| **Goroutine state** | `running` | `GC wait` / `chan recv` | `syscall` / `chan recv` |
| **GC frequency** | High (if alloc-heavy) | VERY HIGH | Low |
| **Load average** | HIGH (> num cores) | Normal-High | Normal |
| **Fix direction** | Do less / faster computation | Use less / reuse memory | Wait less / cache / parallelize |
| **Adding CPU cores helps?** | ✅ YES | ❌ No | ❌ No |
| **Adding RAM helps?** | ❌ No | ✅ YES | Slightly (OS cache) |
| **Adding cache helps?** | Slightly | ❌ No | ✅ YES |
| **Primary profiling tool** | `pprof CPU` | `pprof heap` | `pprof block` + `iostat` |

---

### 4B. All Symptoms in One Table

| Symptom | CPU | Memory | I/O |
|---------|-----|--------|-----|
| `top` shows `%us` > 80% | ✅ | | |
| `top` shows `%wa` > 20% | | | ✅ disk |
| `top` shows `%id` > 60% with slow responses | | | ✅ network |
| Load average > num cores | ✅ | | |
| `free -h` shows swap in use | | ✅ | |
| `iostat` shows `%util` > 80% | | | ✅ |
| `iostat` shows `await` > 50ms | | | ✅ |
| Heap grows continuously | | ✅ leak | |
| OOM kill / `fatal error: out of memory` | | ✅ | |
| Goroutine count grows unboundedly | | ✅ | |
| `GODEBUG=gctrace` shows GC every few seconds | | ✅ | |
| GC consuming > 15% CPU | | ✅ | |
| Slow even with no DB/network calls | ✅ | | |
| Slow DB queries in logs | | | ✅ |
| `db.Stats().WaitCount > 0` | | | ✅ |
| `context deadline exceeded` errors | | | ✅ |
| Goroutines stuck in `syscall` state | | | ✅ |
| Performance degrades over hours | | ✅ | |
| Fast with 1 user, slow with 100 | ✅ | ✅ | ✅ |
| Adding goroutines doesn't improve throughput | ✅ | | ✅ |
| `pprof CPU` shows your function at top | ✅ | | |
| `pprof heap` shows growing allocations | | ✅ | |
| `pprof block` shows long wait times | | | ✅ |

---

### 4C. All Causes in One Table

| Cause | CPU | Memory | I/O |
|-------|-----|--------|-----|
| O(n²) algorithm | ✅ | | |
| JSON marshal/unmarshal in hot path | ✅ | ✅ allocs | |
| Regex compiled per call | ✅ | | |
| String concat with `+` in loop | ✅ | ✅ | |
| Unbounded goroutines | ✅ | ✅ | |
| GC pressure (many small allocs) | ✅ | ✅ | |
| Reflection in hot path | ✅ | | |
| Template parsed per request | ✅ | | |
| bcrypt cost too high | ✅ | | |
| Goroutine leak (never exits) | | ✅ | |
| Slice retains backing array | | ✅ | |
| Map never shrinks | | ✅ | |
| Unbounded cache | | ✅ | |
| Large buffer per request | | ✅ | |
| Closure captures large var | | ✅ | |
| Serial I/O calls | | | ✅ |
| No DB connection pool | | | ✅ |
| Missing DB index | | | ✅ |
| N+1 query problem | | | ✅ |
| No caching for repeated reads | | | ✅ |
| Unbuffered file writes | | | ✅ |
| Loading entire large file | | ✅ | ✅ |
| Many small network requests | | | ✅ |
| No timeout on outbound calls | | | ✅ |
| DB connection leak | | | ✅ |
| No circuit breaker | | | ✅ |
| No retry with backoff | | | ✅ |

---

### 4D. All Tools in One Table

| Tool | CPU | Memory | I/O | Command |
|------|-----|--------|-----|---------|
| `top` | ✅ `%us` | ✅ `%MEM` | ✅ `%wa` | `top -1` |
| `htop` | ✅ | ✅ | ✅ | `htop` |
| `vmstat` | ✅ | ✅ swap | ✅ | `vmstat 1` |
| `iostat -x 1` | | | ✅ disk | `iostat -x 1` |
| `iotop` | | | ✅ per-proc disk | `iotop -o` |
| `free -h` | | ✅ | | `free -h` |
| `iftop` | | | ✅ network | `iftop` |
| `ss` / `netstat` | | | ✅ connections | `ss -s` |
| `lsof` | | | ✅ open files | `lsof -p <pid>` |
| `strace` (Linux) | | | ✅ syscalls | `strace -p <pid>` |
| `dtruss` (macOS) | | | ✅ syscalls | `sudo dtruss -p <pid>` |
| `tcpdump` | | | ✅ packets | `tcpdump -i eth0` |
| `perf stat` | ✅ hw counters | | | `perf stat ./app` |
| `pprof CPU` | ✅ flame graph | | | `go tool pprof .../profile` |
| `pprof heap` | | ✅ allocations | | `go tool pprof .../heap` |
| `pprof goroutine` | ✅ | ✅ leak | ✅ blocked | `go tool pprof .../goroutine` |
| `pprof block` | | | ✅ wait time | `go tool pprof .../block` |
| `pprof mutex` | ✅ contention | | | `go tool pprof .../mutex` |
| `go tool trace` | ✅ scheduling | ✅ GC | ✅ syscalls | `go tool trace trace.out` |
| `go test -bench` | ✅ | ✅ `-benchmem` | | `go test -bench=. -benchmem` |
| `benchstat` | ✅ | ✅ | | `benchstat old.txt new.txt` |
| `goleak` | | ✅ goroutine leak | | `goleak.VerifyNone(t)` |
| `GODEBUG=gctrace` | ✅ GC CPU | ✅ GC freq | | env var |
| `gcflags="-m"` | ✅ escape | ✅ heap alloc | | build flag |
| `EXPLAIN ANALYZE` | | | ✅ DB query plan | SQL |
| `pg_stat_statements` | | | ✅ slow queries | Postgres |
| `pg_stat_activity` | | | ✅ connections | Postgres |
| `redis-cli info` | | | ✅ Redis stats | CLI |
| `wrk` / `hey` / `ab` | ✅ load test | ✅ | ✅ | `wrk -t4 -c100 -d30s` |

---

### 4E. All Fixes in One Table

| Fix | CPU | Memory | I/O | Code Pattern |
|-----|-----|--------|-----|-------------|
| Worker pool (N=NumCPU) | ✅ | | | `make(chan Job, N)` + N goroutines |
| `sync.Pool` | ✅ GC | ✅ allocs | | `pool.Get()` / `pool.Put()` |
| Better algorithm | ✅ | | | O(n) map lookup vs O(n²) loop |
| Package-level regex | ✅ | | | `var re = regexp.MustCompile(...)` |
| `strings.Builder` | ✅ | ✅ | | `sb.Grow(n); sb.WriteString(s)` |
| Memoization / cache | ✅ | | ✅ | `lru.New(maxSize)` |
| Tune `GOGC` | ✅ | | | `GOGC=200` or `debug.SetGCPercent(200)` |
| Faster serializer | ✅ | ✅ | | `jsoniter`, protobuf, msgpack |
| Pre-allocate slices | ✅ | ✅ | | `make([]T, 0, capacity)` |
| Value types | ✅ | ✅ | | `T` not `*T` for small structs |
| `context` + `defer cancel()` | | ✅ goroutine | ✅ timeout | Every goroutine + outbound call |
| LRU / TTL cache | | ✅ | ✅ | `lru.New(N)` or TTL map |
| `copy()` sub-slice | | ✅ | | `copy(dst, src[:n])` |
| Replace map | | ✅ | | `m = make(map[K]V)` |
| `debug.SetMemoryLimit` | | ✅ | | Go 1.19+, 80% of container limit |
| Stream large data | | ✅ | ✅ | `bufio.Scanner` / `rows.Next()` |
| `errgroup` parallel | | | ✅ | `g.Go(func() error {...})` |
| Multi-level cache | | | ✅ | L1 map → L2 Redis → L3 DB |
| `bufio.Writer` | | | ✅ disk | `bufio.NewWriterSize(f, 65536)` |
| DB connection pool | | | ✅ | `SetMaxOpenConns(25)` |
| `defer rows.Close()` | | | ✅ | Immediately after `db.Query()` |
| Package-level `http.Client` | | | ✅ | One client per service |
| `CREATE INDEX` | | | ✅ | Index WHERE/JOIN/ORDER BY cols |
| JOIN / batch query | | | ✅ | `WHERE id IN (...)` |
| Redis pipeline | | | ✅ | `pipe := client.Pipeline()` |
| Circuit breaker | | | ✅ | `gobreaker.NewCircuitBreaker(...)` |
| Exponential backoff | | | ✅ | Base 100ms, max 3 retries |
| Async write | | | ✅ | Buffered channel + worker |
| Read replica | | | ✅ | Route SELECTs to replica |

---

## PART 5 — STANDARDS & CONVENTIONS REFERENCE

### 5A. Go Performance Standards

| Area | Standard / Convention |
|------|----------------------|
| **GOMAXPROCS** | Default = NumCPU (Go 1.5+). Explicit only if you need to limit. |
| **Worker pool size (CPU work)** | `runtime.NumCPU()` workers |
| **Worker pool size (I/O work)** | 10-100x NumCPU (goroutines are cheap for I/O waiting) |
| **DB pool: max open** | 25 (general), tune based on DB server max_connections |
| **DB pool: max idle** | 10 (keep warm connections ready) |
| **DB pool: conn lifetime** | 5 minutes (prevents stale connections) |
| **DB pool: idle timeout** | 1 minute |
| **HTTP client timeout** | External API: 10s, Internal service: 2s, DB: 5s |
| **Context timeout** | Always set; same values as HTTP client timeout |
| **Buffer size (file I/O)** | 64KB (`bufio.NewWriterSize(f, 65536)`) |
| **Cache max size** | 20-30% of available RAM |
| **Cache TTL** | Match data freshness requirement (config: 5min, user data: 1min) |
| **Log buffer** | 10K entries or 10MB, flush every 1s |
| **Retry attempts** | Max 3 retries for idempotent operations |
| **Retry backoff** | Exponential: 100ms → 200ms → 400ms with ±10% jitter |
| **Circuit breaker threshold** | Open after 5 consecutive failures |
| **Circuit breaker reset** | Half-open after 30s |
| **bcrypt cost** | 10 (standard); benchmark on your hardware |
| **GOGC** | Default 100; increase to 200 if GC is > 10% of CPU |
| **Memory limit** | `debug.SetMemoryLimit(0.8 * containerMemory)` |
| **Goroutine stack** | Starts at 2KB, grows to 1MB default max |
| **Channel buffer size** | Expected burst size (not max possible) |

### 5B. Thresholds: Good / Warning / Critical

| Metric | Good ✅ | Warning ⚠️ | Critical 🔴 |
|--------|---------|-----------|------------|
| CPU utilization | < 70% | 70-85% | > 85% |
| Load avg / core count | < 0.7 | 0.7-1.5 | > 1.5 |
| Memory utilization | < 70% | 70-85% | > 85% |
| Swap usage | 0 | Any | > 10% of RAM |
| GC time % of CPU | < 5% | 5-15% | > 15% |
| GC pause (single) | < 1ms | 1-10ms | > 10ms |
| Goroutine count | < 10K | 10K-100K | > 100K |
| DB connection wait | 0 | > 0 | > 10ms avg |
| DB query time | < 10ms | 10-100ms | > 100ms |
| Disk `%util` | < 60% | 60-80% | > 80% |
| Disk `await` | < 5ms SSD | 5-20ms | > 20ms SSD |
| HTTP p50 latency | < 50ms | 50-200ms | > 200ms |
| HTTP p99 latency | < 200ms | 200ms-1s | > 1s |
| Error rate | < 0.1% | 0.1-1% | > 1% |
| Network RTT (internal) | < 1ms | 1-10ms | > 10ms |

---

## PART 6 — QUICK DIAGNOSIS DECISION TREE

```
Your app is slow / using too many resources
│
├─ Step 1: Run `top -1`
│   │
│   ├─ %us > 80%?
│   │   └─► CPU-BOUND
│   │       → pprof CPU: go tool pprof -http=:8080 .../profile?seconds=30
│   │       → Look at flame graph top function
│   │       → Is it: algorithm? JSON? regex? GC? goroutines?
│   │       → Fix: worker pool / better algo / sync.Pool / cache
│   │
│   ├─ %wa > 20%?
│   │   └─► I/O-BOUND (disk)
│   │       → iostat -x 1: which device is at 100%?
│   │       → strace -p <pid>: which files are being read/written?
│   │       → Fix: bufio / async writes / SSD / read cache
│   │
│   └─ CPU < 20%, %wa < 5%, but still slow?
│       │
│       ├─ Check goroutine count: curl .../goroutine?debug=1
│       │   ├─ Count growing?
│       │   │   └─► GOROUTINE LEAK (Memory-bound)
│       │   │       → goleak in tests / pprof goroutine
│       │   │       → Fix: context + cancel, close channels
│       │   │
│       │   └─ Count stable but high?
│       │       └─► NETWORK I/O-BOUND
│       │           → pprof block: where are goroutines waiting?
│       │           → Fix: parallel calls / cache / connection pool
│       │
│       └─ Check heap: pprof heap
│           ├─ Heap growing over time?
│           │   └─► MEMORY LEAK
│           │       → Compare two heap profiles (5 min apart)
│           │       → Fix: LRU cache / copy slices / replace maps
│           │
│           └─ Heap large but stable?
│               └─► MEMORY CAPACITY
│                   → Reduce allocations / sync.Pool / stream data
│                   → debug.SetMemoryLimit / increase RAM
```

---

## PART 7 — ONE-LINE DEFINITIONS

| Term | Definition |
|------|-----------|
| **CPU-bound** | Bottleneck is the processor; adding CPU cores helps |
| **Memory-bound** | Bottleneck is RAM; either not enough or too many allocations |
| **I/O-bound** | Bottleneck is disk or network; CPU is idle waiting |
| **iowait (`%wa`)** | % of time CPU is idle specifically waiting for disk I/O |
| **Load average** | Avg number of processes wanting CPU over last 1/5/15 minutes |
| **Heap** | Dynamically allocated memory managed by Go's garbage collector |
| **Stack** | Per-goroutine memory for local variables; fast, auto-managed |
| **GC pressure** | Frequent GC cycles caused by many short-lived heap allocations |
| **Goroutine leak** | Goroutine started but never exits; holds memory and a goroutine slot |
| **N+1 problem** | 1 query to fetch list + 1 query per item = N+1 total DB round-trips |
| **Connection pool** | Pre-created reusable connections to avoid TCP/TLS setup per request |
| **Backpressure** | Rejecting or slowing new work when the system is at capacity |
| **Escape analysis** | Compiler determines if a variable can live on stack or must go to heap |
| **Amdahl's Law** | Max parallel speedup = 1/(serial_fraction + parallel_fraction/N) |
| **Bottleneck shift** | Fixing one bottleneck reveals the next; expected and normal |
| **Cache miss** | Requested data not in cache; must fetch from slower backing store |
| **Flame graph** | Visualization of CPU profile; width = time spent in that function |
| **pprof** | Go's profiling framework; CPU, heap, goroutine, block, mutex profiles |
| **GOMAXPROCS** | Number of OS threads Go scheduler uses; defaults to CPU core count |
| **GOGC** | GC target: heap can grow X% before GC runs; default 100 (doubles) |
| **sync.Pool** | Thread-safe pool of reusable objects; reduces GC pressure |
| **errgroup** | `golang.org/x/sync/errgroup`; run goroutines in parallel, collect errors |
| **Circuit breaker** | Stops calling a failing service after N failures; retries after timeout |
| **Exponential backoff** | Retry delay doubles each attempt: 100ms, 200ms, 400ms... |
| **RSS** | Resident Set Size; actual physical RAM used by a process |
| **Swap** | Disk space used as overflow RAM; 1000x slower than RAM |
| **await (iostat)** | Average time (ms) an I/O request waits in disk queue + service time |
| **`%util` (iostat)** | % of time the disk device was busy handling requests |
| **Block profile** | pprof profile showing where goroutines spend time blocked/waiting |
| **Mutex profile** | pprof profile showing where goroutines wait to acquire a mutex lock |
| **Benchstat** | Tool to statistically compare two sets of Go benchmark results |
| **goleak** | Go library to detect goroutine leaks in tests |
| **LRU cache** | Least Recently Used cache; evicts oldest-accessed item when full |
| **TTL cache** | Time-To-Live cache; evicts items after a fixed duration |
| **Read replica** | DB copy that handles read queries; primary handles writes only |
| **Pipeline (Redis)** | Send multiple Redis commands in one network round-trip |
| **`bufio`** | Go package that adds buffering to I/O; reduces syscall count |
| **`strings.Builder`** | Efficient string building; avoids O(n²) allocations from `+` operator |
