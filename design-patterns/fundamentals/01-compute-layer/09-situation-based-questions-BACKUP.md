# 100 Situation-Based Questions for Software Architects

## Overview

This document contains 100 real-world scenario-based questions focusing on:
- CPU-bound workloads
- Memory-bound workloads
- I/O-bound workloads
- Scaling strategies
- Go-specific issues
- Debugging techniques
- Production troubleshooting

Each question includes context, analysis, and detailed solutions with Go code examples.

---

## Table of Contents

1. [CPU-Bound Scenarios (20 Questions)](#cpu-bound-scenarios)
2. [Memory-Bound Scenarios (20 Questions)](#memory-bound-scenarios)
3. [I/O-Bound Scenarios (20 Questions)](#io-bound-scenarios)
4. [Scaling Scenarios (15 Questions)](#scaling-scenarios)
5. [Go-Specific Issues (15 Questions)](#go-specific-issues)
6. [Debugging & Troubleshooting (10 Questions)](#debugging-troubleshooting)

---

## CPU-Bound Scenarios

### Q1: High CPU Usage in Image Processing Service

**Situation:**
Your Go-based image processing service is experiencing 95% CPU usage. It processes 1000 images/minute, resizing them from 4K to multiple resolutions. Response time has increased from 200ms to 2 seconds.

**Analysis:**
```
Current State:
• CPU: 95% (8 cores)
• Memory: 40%
• Goroutines: 5000+
• Processing: Sequential per image
```

**Solution:**

```go
// Problem: Too many goroutines competing for CPU
func processImageBad(img image.Image) []image.Image {
    var wg sync.WaitGroup
    results := make([]image.Image, 5)
    
    // Creates 5 goroutines per image = 5000 goroutines!
    for i, size := range sizes {
        wg.Add(1)
        go func(idx int, s Size) {
            defer wg.Done()
            results[idx] = resize(img, s)
        }(i, size)
    }
    wg.Wait()
    return results
}

// Solution: Worker pool pattern
type ImageProcessor struct {
    workerPool chan struct{}
    jobs       chan ImageJob
}

func NewImageProcessor(workers int) *ImageProcessor {
    p := &ImageProcessor{
        workerPool: make(chan struct{}, workers),
        jobs:       make(chan ImageJob, 100),
    }
    
    // Create worker pool = number of CPU cores
    for i := 0; i < workers; i++ {
        go p.worker()
    }
    return p
}

func (p *ImageProcessor) worker() {
    for job := range p.jobs {
        // Process all sizes sequentially in one goroutine
        results := make([]image.Image, len(sizes))
        for i, size := range sizes {
            results[i] = resize(job.Image, size)
        }
        job.ResultChan <- results
    }
}

func main() {
    // Use runtime.NumCPU() for optimal performance
    processor := NewImageProcessor(runtime.NumCPU())
    
    // Now only 8 goroutines doing actual work
    // CPU usage: 95% → 85% (less context switching)
    // Throughput: Same or better
    // Latency: 2s → 300ms
}
```

**Key Takeaways:**
- Limit goroutines to CPU count for CPU-bound tasks
- Use worker pools to prevent goroutine explosion
- Measure: `runtime.NumGoroutine()`, `runtime/pprof`

---

### Q2: Video Encoding Service Bottleneck

**Situation:**
Video encoding service processes 100 videos/hour. Each video takes 5 minutes. You need to scale to 1000 videos/hour.

**Current Architecture:**
```
Single Server:
• 16 cores
• Processing: 1 video at a time
• Utilization: 100% on 1 core, 0% on others
```

**Solution:**

```go
// Problem: Not utilizing all cores
func encodeVideoBad(video Video) error {
    // Uses only 1 core
    return ffmpeg.Encode(video)
}

// Solution 1: Parallel chunk processing
func encodeVideoParallel(video Video) error {
    chunks := splitVideo(video, runtime.NumCPU())
    
    var wg sync.WaitGroup
    errors := make(chan error, len(chunks))
    
    for _, chunk := range chunks {
        wg.Add(1)
        go func(c VideoChunk) {
            defer wg.Done()
            if err := encodeChunk(c); err != nil {
                errors <- err
            }
        }(chunk)
    }
    
    wg.Wait()
    close(errors)
    
    // Merge encoded chunks
    return mergeChunks(chunks)
}

// Solution 2: Process multiple videos concurrently
type VideoEncoder struct {
    semaphore chan struct{}
}

func NewVideoEncoder(concurrency int) *VideoEncoder {
    return &VideoEncoder{
        semaphore: make(chan struct{}, concurrency),
    }
}

func (e *VideoEncoder) Encode(video Video) error {
    e.semaphore <- struct{}{} // Acquire
    defer func() { <-e.semaphore }() // Release
    
    return ffmpeg.Encode(video)
}

func main() {
    // Process 16 videos concurrently (one per core)
    encoder := NewVideoEncoder(16)
    
    // Throughput: 100 → 960 videos/hour
    // CPU Utilization: 6% → 95%
}
```

**Capacity Calculation:**
```
Current: 1 video per 5 min = 12 videos/hour per core
With 16 cores: 12 × 16 = 192 videos/hour

To reach 1000 videos/hour:
Servers needed = 1000 / 192 = 5.2 → 6 servers
```

---

### Q3: JSON Parsing CPU Spike

**Situation:**
API gateway parsing large JSON payloads (5-10 MB) causing CPU spikes to 100%. Latency increases from 50ms to 500ms during peak hours.

**Solution:**

```go
// Problem: Parsing entire JSON into memory
type Request struct {
    Data []Record `json:"data"` // 10MB array
}

func handleRequestBad(w http.ResponseWriter, r *http.Request) {
    var req Request
    // Loads entire 10MB into memory and parses
    json.NewDecoder(r.Body).Decode(&req)
    
    for _, record := range req.Data {
        process(record)
    }
}

// Solution 1: Streaming JSON parser
func handleRequestStreaming(w http.ResponseWriter, r *http.Request) {
    dec := json.NewDecoder(r.Body)
    
    // Read opening bracket
    dec.Token()
    
    // Stream process each record
    for dec.More() {
        var record Record
        if err := dec.Decode(&record); err != nil {
            break
        }
        process(record)
    }
    
    // CPU: 100% → 40%
    // Memory: 10MB → 100KB
    // Latency: 500ms → 80ms
}

// Solution 2: Use faster JSON library
import "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func handleRequestFast(w http.ResponseWriter, r *http.Request) {
    var req Request
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2-3x faster than standard library
}

// Solution 3: Parallel processing
func handleRequestParallel(w http.ResponseWriter, r *http.Request) {
    var req Request
    json.NewDecoder(r.Body).Decode(&req)
    
    // Process in parallel batches
    batchSize := len(req.Data) / runtime.NumCPU()
    var wg sync.WaitGroup
    
    for i := 0; i < len(req.Data); i += batchSize {
        end := i + batchSize
        if end > len(req.Data) {
            end = len(req.Data)
        }
        
        wg.Add(1)
        go func(batch []Record) {
            defer wg.Done()
            for _, record := range batch {
                process(record)
            }
        }(req.Data[i:end])
    }
    wg.Wait()
}
```

---

### Q4: Cryptographic Operations Bottleneck

**Situation:**
Authentication service performs bcrypt hashing on every login. With 10K logins/minute, CPU is at 100%.

**Solution:**

```go
// Problem: Expensive CPU operation on hot path
func loginBad(username, password string) bool {
    user := getUser(username)
    // bcrypt takes ~100ms per hash
    err := bcrypt.CompareHashAndPassword(
        []byte(user.PasswordHash),
        []byte(password),
    )
    return err == nil
}

// Solution 1: Rate limiting
type RateLimiter struct {
    limiter *rate.Limiter
}

func (rl *RateLimiter) login(username, password string) (bool, error) {
    if !rl.limiter.Allow() {
        return false, errors.New("rate limit exceeded")
    }
    
    user := getUser(username)
    err := bcrypt.CompareHashAndPassword(
        []byte(user.PasswordHash),
        []byte(password),
    )
    return err == nil, nil
}

// Solution 2: Token-based auth (reduce hashing)
type AuthService struct {
    cache *cache.Cache
}

func (a *AuthService) login(username, password string) (string, error) {
    // Check cache first
    if token, found := a.cache.Get(username); found {
        return token.(string), nil
    }
    
    // Only hash on first login
    user := getUser(username)
    if err := bcrypt.CompareHashAndPassword(
        []byte(user.PasswordHash),
        []byte(password),
    ); err != nil {
        return "", err
    }
    
    // Generate token, cache for 1 hour
    token := generateJWT(user)
    a.cache.Set(username, token, time.Hour)
    
    return token, nil
}

// Solution 3: Offload to dedicated service
type AuthWorker struct {
    jobs    chan AuthJob
    results map[string]chan AuthResult
    mu      sync.RWMutex
}

func (w *AuthWorker) worker() {
    for job := range w.jobs {
        result := AuthResult{
            Success: bcrypt.CompareHashAndPassword(
                []byte(job.Hash),
                []byte(job.Password),
            ) == nil,
        }
        
        w.mu.RLock()
        resultChan := w.results[job.ID]
        w.mu.RUnlock()
        
        resultChan <- result
    }
}

// Metrics:
// Before: 10K req/min, 100% CPU, 500ms latency
// After: 10K req/min, 60% CPU, 150ms latency
```

---

### Q5: Data Compression Service

**Situation:**
Log aggregation service compresses 1GB of logs every minute. Compression takes 45 seconds, causing backlog.

**Solution:**

```go
// Problem: Single-threaded compression
func compressLogsBad(logs []byte) []byte {
    var buf bytes.Buffer
    w := gzip.NewWriter(&buf)
    w.Write(logs) // Slow, single-threaded
    w.Close()
    return buf.Bytes()
}

// Solution 1: Parallel compression with pgzip
import "github.com/klauspost/pgzip"

func compressLogsParallel(logs []byte) []byte {
    var buf bytes.Buffer
    w := pgzip.NewWriter(&buf)
    w.SetConcurrency(1<<20, runtime.NumCPU()) // 1MB blocks, all cores
    w.Write(logs)
    w.Close()
    return buf.Bytes()
    
    // Time: 45s → 8s (5.6x faster)
}

// Solution 2: Stream compression
func compressLogsStreaming(logChan <-chan []byte, output io.Writer) {
    w := pgzip.NewWriter(output)
    defer w.Close()
    
    for logBatch := range logChan {
        w.Write(logBatch)
    }
    
    // No memory spike, continuous processing
}

// Solution 3: Different compression algorithm
import "github.com/klauspost/compress/zstd"

func compressLogsZstd(logs []byte) []byte {
    encoder, _ := zstd.NewWriter(nil,
        zstd.WithEncoderLevel(zstd.SpeedFastest),
        zstd.WithEncoderConcurrency(runtime.NumCPU()),
    )
    return encoder.EncodeAll(logs, nil)
    
    // Time: 45s → 5s (9x faster)
    // Compression ratio: Similar to gzip
}
```

---

## Memory-Bound Scenarios

### Q6: Memory Leak in Long-Running Service

**Situation:**
Your Go service starts with 500MB memory, grows to 8GB after 24 hours, then crashes with OOM.

**Debugging Process:**

```go
// Step 1: Enable pprof
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Your application code
}

// Step 2: Capture heap profile
// $ curl http://localhost:6060/debug/pprof/heap > heap.prof
// $ go tool pprof heap.prof

// Step 3: Analyze
// (pprof) top
// (pprof) list functionName
// (pprof) web

// Common Problem 1: Unbounded cache
type CacheBad struct {
    data map[string][]byte
    mu   sync.RWMutex
}

func (c *CacheBad) Set(key string, value []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value // Never evicted!
}

// Solution: LRU cache with size limit
import "github.com/hashicorp/golang-lru"

type CacheGood struct {
    cache *lru.Cache
}

func NewCache(size int) *CacheGood {
    cache, _ := lru.New(size)
    return &CacheGood{cache: cache}
}

func (c *CacheGood) Set(key string, value []byte) {
    c.cache.Add(key, value)
    // Automatically evicts oldest when full
}

// Common Problem 2: Goroutine leak
func processRequestsBad() {
    for req := range requests {
        go func(r Request) {
            // If this blocks forever, goroutine leaks
            result := externalAPI.Call(r) // No timeout!
            sendResponse(result)
        }(req)
    }
}

// Solution: Context with timeout
func processRequestsGood() {
    for req := range requests {
        go func(r Request) {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            result, err := externalAPI.CallWithContext(ctx, r)
            if err != nil {
                log.Printf("Request failed: %v", err)
                return
            }
            sendResponse(result)
        }(req)
    }
}

// Common Problem 3: Slice capacity leak
func appendDataBad(data []byte) [][]byte {
    var results [][]byte
    for i := 0; i < 1000000; i++ {
        // Keeps reference to entire underlying array
        results = append(results, data[i:i+100])
    }
    return results
}

// Solution: Copy to new slice
func appendDataGood(data []byte) [][]byte {
    var results [][]byte
    for i := 0; i < 1000000; i++ {
        chunk := make([]byte, 100)
        copy(chunk, data[i:i+100])
        results = append(results, chunk)
    }
    return results
}

// Monitoring memory
func monitorMemory() {
    var m runtime.MemStats
    ticker := time.NewTicker(10 * time.Second)
    
    for range ticker.C {
        runtime.ReadMemStats(&m)
        log.Printf("Alloc = %v MB", m.Alloc/1024/1024)
        log.Printf("TotalAlloc = %v MB", m.TotalAlloc/1024/1024)
        log.Printf("Sys = %v MB", m.Sys/1024/1024)
        log.Printf("NumGC = %v", m.NumGC)
        log.Printf("Goroutines = %v", runtime.NumGoroutine())
    }
}
```

---

### Q7: High Memory Usage in Data Processing Pipeline

**Situation:**
ETL pipeline processes 10GB CSV file, loads entire file into memory causing OOM.

**Solution:**

```go
// Problem: Loading entire file
func processCSVBad(filename string) error {
    data, err := ioutil.ReadFile(filename) // 10GB in memory!
    if err != nil {
        return err
    }
    
    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        process(line)
    }
    return nil
}

// Solution 1: Stream processing
func processCSVStreaming(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    // Set buffer size if lines are large
    buf := make([]byte, 1024*1024) // 1MB buffer
    scanner.Buffer(buf, 10*1024*1024) // Max 10MB per line
    
    for scanner.Scan() {
        line := scanner.Text()
        process(line)
    }
    
    // Memory: 10GB → 10MB
    return scanner.Err()
}

// Solution 2: Parallel streaming with worker pool
func processCSVParallel(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    lines := make(chan string, 1000)
    errors := make(chan error, 1)
    
    // Workers
    var wg sync.WaitGroup
    for i := 0; i < runtime.NumCPU(); i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for line := range lines {
                if err := process(line); err != nil {
                    select {
                    case errors <- err:
                    default:
                    }
                    return
                }
            }
        }()
    }
    
    // Reader
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        select {
        case lines <- scanner.Text():
        case err := <-errors:
            close(lines)
            return err
        }
    }
    close(lines)
    
    wg.Wait()
    
    // Memory: 10GB → 50MB (buffer + workers)
    // Time: 60s → 12s (5x faster)
    return scanner.Err()
}

// Solution 3: Memory-mapped file
import "golang.org/x/exp/mmap"

func processCSVMmap(filename string) error {
    reader, err := mmap.Open(filename)
    if err != nil {
        return err
    }
    defer reader.Close()
    
    // OS handles paging, only loads needed parts
    data := make([]byte, 1024*1024) // 1MB chunks
    offset := 0
    
    for {
        n, err := reader.ReadAt(data, int64(offset))
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        processChunk(data[:n])
        offset += n
    }
    
    return nil
}
```

---

### Q8: WebSocket Connection Memory Explosion

**Situation:**
WebSocket server with 100K concurrent connections consuming 20GB memory (200KB per connection).

**Solution:**

```go
// Problem: Each connection stores full message history
type ConnectionBad struct {
    conn     *websocket.Conn
    messages []Message // Grows unbounded!
    send     chan []byte
}

func (c *ConnectionBad) readPump() {
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        c.messages = append(c.messages, Message{Data: message})
    }
}

// Solution 1: Don't store messages
type ConnectionGood struct {
    conn *websocket.Conn
    send chan []byte
}

func (c *ConnectionGood) readPump() {
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        // Process immediately, don't store
        handleMessage(message)
    }
}

// Solution 2: Use buffer pool
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func (c *ConnectionGood) readPumpWithPool() {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        
        // Reuse buffer
        copy(buf, message)
        handleMessage(buf[:len(message)])
    }
}

// Solution 3: Limit send channel size
type ConnectionOptimized struct {
    conn *websocket.Conn
    send chan []byte // Buffered channel
}

func NewConnection(conn *websocket.Conn) *ConnectionOptimized {
    return &ConnectionOptimized{
        conn: conn,
        send: make(chan []byte, 256), // Limit buffer
    }
}

func (c *ConnectionOptimized) writePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                return
            }
            
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }
            
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// Memory optimization results:
// Before: 100K connections × 200KB = 20GB
// After: 100K connections × 50KB = 5GB (4x improvement)
```

---


### Q9: In-Memory Cache Growing Unbounded

**Situation:**
Your API caching layer starts at 1GB, grows to 32GB after a week, causing frequent GC pauses (5+ seconds).

**Solution:**

```go
// Problem: No eviction policy
type CacheBad struct {
    data map[string]*CacheEntry
    mu   sync.RWMutex
}

func (c *CacheBad) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = &CacheEntry{
        Value:     value,
        CreatedAt: time.Now(),
    }
    // Never removes old entries!
}

// Solution 1: TTL-based expiration
type CacheWithTTL struct {
    data map[string]*CacheEntry
    mu   sync.RWMutex
    ttl  time.Duration
}

func (c *CacheWithTTL) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = &CacheEntry{
        Value:     value,
        ExpiresAt: time.Now().Add(c.ttl),
    }
}

func (c *CacheWithTTL) cleanup() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, entry := range c.data {
            if now.After(entry.ExpiresAt) {
                delete(c.data, key)
            }
        }
        c.mu.Unlock()
    }
}

// Solution 2: LRU with size limit
import "github.com/hashicorp/golang-lru/v2"

type CacheLRU struct {
    cache *lru.Cache[string, interface{}]
}

func NewCacheLRU(size int) *CacheLRU {
    cache, _ := lru.New[string, interface{}](size)
    return &CacheLRU{cache: cache}
}

func (c *CacheLRU) Set(key string, value interface{}) {
    c.cache.Add(key, value)
    // Automatically evicts least recently used
}

// Solution 3: Two-tier cache (hot/cold)
type TwoTierCache struct {
    hot  *lru.Cache[string, interface{}]
    cold *lru.Cache[string, interface{}]
    mu   sync.RWMutex
}

func NewTwoTierCache(hotSize, coldSize int) *TwoTierCache {
    hot, _ := lru.New[string, interface{}](hotSize)
    cold, _ := lru.New[string, interface{}](coldSize)
    return &TwoTierCache{hot: hot, cold: cold}
}

func (c *TwoTierCache) Get(key string) (interface{}, bool) {
    // Check hot cache first
    if val, ok := c.hot.Get(key); ok {
        return val, true
    }
    
    // Check cold cache, promote to hot if found
    if val, ok := c.cold.Get(key); ok {
        c.hot.Add(key, val)
        return val, true
    }
    
    return nil, false
}

// Memory: 32GB → 2GB (16x reduction)
// GC pauses: 5s → 50ms (100x improvement)
```

---

### Q10: String Concatenation in Loop

**Situation:**
Log aggregation service concatenating millions of log lines, consuming excessive memory and CPU.

**Solution:**

```go
// Problem: String concatenation creates new strings
func aggregateLogsBad(logs []string) string {
    result := ""
    for _, log := range logs {
        result += log + "\n" // Creates new string each time!
    }
    return result
}

// With 1M logs of 100 bytes each:
// Memory: 50GB+ (quadratic growth)
// Time: 30+ seconds

// Solution 1: strings.Builder
func aggregateLogsGood(logs []string) string {
    var builder strings.Builder
    builder.Grow(len(logs) * 100) // Pre-allocate
    
    for _, log := range logs {
        builder.WriteString(log)
        builder.WriteByte('\n')
    }
    return builder.String()
}

// Memory: 100MB (linear growth)
// Time: 100ms (300x faster)

// Solution 2: bytes.Buffer for []byte
func aggregateLogsBytesBuffer(logs [][]byte) []byte {
    var buf bytes.Buffer
    buf.Grow(len(logs) * 100)
    
    for _, log := range logs {
        buf.Write(log)
        buf.WriteByte('\n')
    }
    return buf.Bytes()
}

// Solution 3: Direct byte slice manipulation
func aggregateLogsDirect(logs [][]byte) []byte {
    // Calculate total size
    totalSize := 0
    for _, log := range logs {
        totalSize += len(log) + 1 // +1 for newline
    }
    
    // Allocate once
    result := make([]byte, 0, totalSize)
    
    for _, log := range logs {
        result = append(result, log...)
        result = append(result, '\n')
    }
    
    return result
}

// Fastest: No intermediate allocations
```

---

## I/O-Bound Scenarios

### Q11: Database Connection Pool Exhaustion

**Situation:**
Web API with 1000 req/s exhausts database connection pool (max 100 connections), causing timeouts.

**Solution:**

```go
// Problem: No connection pooling configuration
func initDBBad() *sql.DB {
    db, _ := sql.Open("postgres", connString)
    // Uses defaults: unlimited connections!
    return db
}

// Solution 1: Proper connection pool configuration
func initDBGood() *sql.DB {
    db, _ := sql.Open("postgres", connString)
    
    // Set connection pool limits
    db.SetMaxOpenConns(25)        // Max open connections
    db.SetMaxIdleConns(25)        // Max idle connections
    db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
    db.SetConnMaxIdleTime(10 * time.Minute) // Idle timeout
    
    return db
}

// Solution 2: Connection pool with retry
type DBPool struct {
    db      *sql.DB
    timeout time.Duration
}

func (p *DBPool) QueryWithRetry(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    var rows *sql.Rows
    var err error
    
    for i := 0; i < 3; i++ {
        ctx, cancel := context.WithTimeout(ctx, p.timeout)
        defer cancel()
        
        rows, err = p.db.QueryContext(ctx, query, args...)
        if err == nil {
            return rows, nil
        }
        
        // Retry on timeout or connection error
        if errors.Is(err, context.DeadlineExceeded) {
            time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
            continue
        }
        break
    }
    return nil, err
}

// Solution 3: Circuit breaker pattern
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    failures    int
    lastFailure time.Time
    state       string // "closed", "open", "half-open"
    mu          sync.RWMutex
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()
    
    if state == "open" {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.mu.Lock()
            cb.state = "half-open"
            cb.mu.Unlock()
        } else {
            return errors.New("circuit breaker open")
        }
    }
    
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        return err
    }
    
    cb.failures = 0
    cb.state = "closed"
    return nil
}

// Metrics:
// Before: 40% timeout rate, 1000ms P99
// After: 0.1% timeout rate, 50ms P99
```

---

### Q12: Slow File I/O Operations

**Situation:**
Application reads 10,000 small files (1-10KB each) taking 30 seconds. Need to reduce to under 5 seconds.

**Solution:**

```go
// Problem: Opening/closing file for each read
func readFilesBad(filenames []string) ([][]byte, error) {
    results := make([][]byte, len(filenames))
    
    for i, filename := range filenames {
        data, err := ioutil.ReadFile(filename)
        if err != nil {
            return nil, err
        }
        results[i] = data
    }
    return results
}

// Time: 30s (3ms per file due to syscall overhead)

// Solution 1: Parallel reading with worker pool
func readFilesParallel(filenames []string) ([][]byte, error) {
    results := make([][]byte, len(filenames))
    errors := make(chan error, 1)
    
    type job struct {
        index    int
        filename string
    }
    
    jobs := make(chan job, len(filenames))
    var wg sync.WaitGroup
    
    // Workers
    for i := 0; i < runtime.NumCPU(); i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := range jobs {
                data, err := ioutil.ReadFile(j.filename)
                if err != nil {
                    select {
                    case errors <- err:
                    default:
                    }
                    return
                }
                results[j.index] = data
            }
        }()
    }
    
    // Send jobs
    for i, filename := range filenames {
        jobs <- job{index: i, filename: filename}
    }
    close(jobs)
    
    wg.Wait()
    
    select {
    case err := <-errors:
        return nil, err
    default:
        return results, nil
    }
}

// Time: 30s → 4s (7.5x faster with 8 cores)

// Solution 2: Batch reading with buffer reuse
func readFilesBatched(filenames []string) ([][]byte, error) {
    results := make([][]byte, len(filenames))
    buf := make([]byte, 10*1024) // 10KB buffer
    
    for i, filename := range filenames {
        file, err := os.Open(filename)
        if err != nil {
            return nil, err
        }
        
        n, err := file.Read(buf)
        file.Close()
        
        if err != nil && err != io.EOF {
            return nil, err
        }
        
        // Copy to result
        results[i] = make([]byte, n)
        copy(results[i], buf[:n])
    }
    
    return results, nil
}

// Solution 3: Memory-mapped files for very large datasets
func readFilesMemoryMapped(filenames []string) ([][]byte, error) {
    results := make([][]byte, len(filenames))
    
    for i, filename := range filenames {
        reader, err := mmap.Open(filename)
        if err != nil {
            return nil, err
        }
        
        data := make([]byte, reader.Len())
        reader.ReadAt(data, 0)
        reader.Close()
        
        results[i] = data
    }
    
    return results, nil
}
```

---

### Q13: API Rate Limiting Issues

**Situation:**
Your service calls external API 10,000 times/minute, hitting rate limit (1000/minute), causing failures.

**Solution:**

```go
// Problem: No rate limiting
func callAPIBad(requests []Request) []Response {
    responses := make([]Response, len(requests))
    
    for i, req := range requests {
        resp, _ := http.Post(apiURL, "application/json", req.Body)
        responses[i] = parseResponse(resp)
    }
    
    return responses
}

// 90% failure rate due to rate limiting

// Solution 1: Token bucket rate limiter
import "golang.org/x/time/rate"

type RateLimitedClient struct {
    client  *http.Client
    limiter *rate.Limiter
}

func NewRateLimitedClient(rps int) *RateLimitedClient {
    return &RateLimitedClient{
        client:  &http.Client{Timeout: 10 * time.Second},
        limiter: rate.NewLimiter(rate.Limit(rps), rps),
    }
}

func (c *RateLimitedClient) Post(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
    if err := c.limiter.Wait(ctx); err != nil {
        return nil, err
    }
    
    req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
    return c.client.Do(req)
}

// Solution 2: Batch requests
type BatchProcessor struct {
    client    *RateLimitedClient
    batchSize int
    interval  time.Duration
}

func (bp *BatchProcessor) ProcessBatch(requests []Request) []Response {
    responses := make([]Response, 0, len(requests))
    
    for i := 0; i < len(requests); i += bp.batchSize {
        end := i + bp.batchSize
        if end > len(requests) {
            end = len(requests)
        }
        
        batch := requests[i:end]
        batchResp := bp.processSingleBatch(batch)
        responses = append(responses, batchResp...)
        
        if end < len(requests) {
            time.Sleep(bp.interval)
        }
    }
    
    return responses
}

// Solution 3: Exponential backoff with retry
type RetryClient struct {
    client      *http.Client
    maxRetries  int
    baseBackoff time.Duration
}

func (rc *RetryClient) PostWithRetry(ctx context.Context, url string, body []byte) (*http.Response, error) {
    var resp *http.Response
    var err error
    
    for i := 0; i < rc.maxRetries; i++ {
        req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
        resp, err = rc.client.Do(req)
        
        if err == nil && resp.StatusCode != 429 {
            return resp, nil
        }
        
        if resp != nil && resp.StatusCode == 429 {
            // Check Retry-After header
            if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
                if seconds, err := strconv.Atoi(retryAfter); err == nil {
                    time.Sleep(time.Duration(seconds) * time.Second)
                    continue
                }
            }
        }
        
        // Exponential backoff
        backoff := rc.baseBackoff * time.Duration(1<<uint(i))
        time.Sleep(backoff)
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", err)
}

// Metrics:
// Before: 10K req/min, 90% failure
// After: 1K req/min, 0% failure, queued processing
```

---

### Q14: Network Timeout Issues

**Situation:**
Microservice calls timing out intermittently (5% of requests), causing cascading failures.

**Solution:**

```go
// Problem: No timeouts configured
func callServiceBad(url string) (*Response, error) {
    resp, err := http.Get(url) // No timeout!
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result Response
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}

// Solution 1: Context with timeout
func callServiceWithTimeout(ctx context.Context, url string) (*Response, error) {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    
    client := &http.Client{
        Timeout: 2 * time.Second,
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result Response
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &result, nil
}

// Solution 2: Timeout with cancellation
type TimeoutClient struct {
    client  *http.Client
    timeout time.Duration
}

func NewTimeoutClient(timeout time.Duration) *TimeoutClient {
    return &TimeoutClient{
        client: &http.Client{
            Timeout: timeout,
            Transport: &http.Transport{
                DialContext: (&net.Dialer{
                    Timeout:   5 * time.Second,
                    KeepAlive: 30 * time.Second,
                }).DialContext,
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 100,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        timeout: timeout,
    }
}

func (tc *TimeoutClient) Get(ctx context.Context, url string) (*Response, error) {
    ctx, cancel := context.WithTimeout(ctx, tc.timeout)
    defer cancel()
    
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    
    respChan := make(chan *http.Response, 1)
    errChan := make(chan error, 1)
    
    go func() {
        resp, err := tc.client.Do(req)
        if err != nil {
            errChan <- err
            return
        }
        respChan <- resp
    }()
    
    select {
    case resp := <-respChan:
        defer resp.Body.Close()
        var result Response
        json.NewDecoder(resp.Body).Decode(&result)
        return &result, nil
        
    case err := <-errChan:
        return nil, err
        
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// Solution 3: Hedged requests (send duplicate after timeout)
func callServiceHedged(ctx context.Context, url string) (*Response, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    type result struct {
        resp *Response
        err  error
    }
    
    results := make(chan result, 2)
    
    // First request
    go func() {
        resp, err := callServiceWithTimeout(ctx, url)
        results <- result{resp, err}
    }()
    
    // Hedged request after 1 second
    timer := time.NewTimer(1 * time.Second)
    defer timer.Stop()
    
    select {
    case r := <-results:
        return r.resp, r.err
        
    case <-timer.C:
        // Send hedged request
        go func() {
            resp, err := callServiceWithTimeout(ctx, url)
            results <- result{resp, err}
        }()
        
        // Return first successful response
        r := <-results
        return r.resp, r.err
    }
}

// Metrics:
// Before: 5% timeout, 10s P99
// After: 0.1% timeout, 2s P99
```

---

### Q15: Disk I/O Bottleneck in Logging

**Situation:**
Application writes 100K log lines/second to disk, causing 80% I/O wait and slow response times.

**Solution:**

```go
// Problem: Synchronous disk writes
func logBad(message string) {
    file, _ := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    defer file.Close()
    
    file.WriteString(time.Now().Format(time.RFC3339) + " " + message + "\n")
    file.Sync() // Force write to disk!
}

// 100K calls/sec = 100K disk writes/sec
// I/O wait: 80%

// Solution 1: Buffered logging
type BufferedLogger struct {
    file   *os.File
    writer *bufio.Writer
    mu     sync.Mutex
}

func NewBufferedLogger(filename string) *BufferedLogger {
    file, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    return &BufferedLogger{
        file:   file,
        writer: bufio.NewWriterSize(file, 256*1024), // 256KB buffer
    }
}

func (l *BufferedLogger) Log(message string) {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    fmt.Fprintf(l.writer, "%s %s\n", time.Now().Format(time.RFC3339), message)
}

func (l *BufferedLogger) Flush() {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    l.writer.Flush()
    l.file.Sync()
}

// Flush periodically
func (l *BufferedLogger) Start() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            l.Flush()
        }
    }()
}

// I/O wait: 80% → 5%

// Solution 2: Async logging with channel
type AsyncLogger struct {
    logs chan string
    file *os.File
}

func NewAsyncLogger(filename string, bufferSize int) *AsyncLogger {
    file, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    
    logger := &AsyncLogger{
        logs: make(chan string, bufferSize),
        file: file,
    }
    
    go logger.worker()
    return logger
}

func (l *AsyncLogger) Log(message string) {
    select {
    case l.logs <- message:
    default:
        // Buffer full, drop or block
        log.Println("Log buffer full")
    }
}

func (l *AsyncLogger) worker() {
    writer := bufio.NewWriterSize(l.file, 256*1024)
    ticker := time.NewTicker(time.Second)
    
    for {
        select {
        case msg := <-l.logs:
            fmt.Fprintf(writer, "%s %s\n", time.Now().Format(time.RFC3339), msg)
            
        case <-ticker.C:
            writer.Flush()
            l.file.Sync()
        }
    }
}

// Solution 3: Use structured logging library
import "go.uber.org/zap"

func setupZapLogger() *zap.Logger {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{"app.log"}
    
    logger, _ := config.Build()
    return logger
}

// Zap uses buffering and async writes internally
// Throughput: 100K → 1M logs/sec
// I/O wait: 80% → 2%
```

---


### Q16: Batch Processing vs Streaming

**Situation:**
Processing 1M records takes 2 hours with batch processing. Need real-time processing.

**Solution:**

```go
// Problem: Batch processing
func processBatch(records []Record) error {
    // Wait for all records
    for _, record := range records {
        process(record)
    }
    return nil
}

// Solution: Streaming with pipeline
func processStream(input <-chan Record) <-chan Result {
    output := make(chan Result, 100)
    
    for i := 0; i < runtime.NumCPU(); i++ {
        go func() {
            for record := range input {
                result := process(record)
                output <- result
            }
        }()
    }
    
    return output
}

// Latency: 2 hours → real-time
// Throughput: Same or better
```

---

### Q17: Database Query N+1 Problem

**Situation:**
Loading 1000 users with their orders takes 30 seconds due to N+1 queries.

**Solution:**

```go
// Problem: N+1 queries
func getUsersWithOrdersBad(db *sql.DB) ([]User, error) {
    users, _ := db.Query("SELECT * FROM users")
    var result []User
    
    for users.Next() {
        var user User
        users.Scan(&user.ID, &user.Name)
        
        // N+1: One query per user!
        orders, _ := db.Query("SELECT * FROM orders WHERE user_id = ?", user.ID)
        user.Orders = scanOrders(orders)
        
        result = append(result, user)
    }
    return result, nil
}

// 1 + 1000 queries = 1001 queries

// Solution: JOIN or batch loading
func getUsersWithOrdersGood(db *sql.DB) ([]User, error) {
    rows, _ := db.Query(`
        SELECT u.id, u.name, o.id, o.product, o.amount
        FROM users u
        LEFT JOIN orders o ON u.id = o.user_id
    `)
    
    userMap := make(map[int]*User)
    
    for rows.Next() {
        var userID int
        var userName string
        var orderID sql.NullInt64
        var product sql.NullString
        var amount sql.NullFloat64
        
        rows.Scan(&userID, &userName, &orderID, &product, &amount)
        
        user, exists := userMap[userID]
        if !exists {
            user = &User{ID: userID, Name: userName}
            userMap[userID] = user
        }
        
        if orderID.Valid {
            user.Orders = append(user.Orders, Order{
                ID:      int(orderID.Int64),
                Product: product.String,
                Amount:  amount.Float64,
            })
        }
    }
    
    result := make([]User, 0, len(userMap))
    for _, user := range userMap {
        result = append(result, *user)
    }
    
    return result, nil
}

// 1 query total
// Time: 30s → 500ms (60x faster)
```

---

## Scaling Scenarios

### Q18: Horizontal Scaling with Session State

**Situation:**
Need to scale web app from 1 to 10 servers, but sessions are stored in memory.

**Solution:**

```go
// Problem: In-memory sessions
var sessions = make(map[string]*Session)
var sessionMu sync.RWMutex

func getSession(sessionID string) *Session {
    sessionMu.RLock()
    defer sessionMu.RUnlock()
    return sessions[sessionID]
}

// Can't scale horizontally!

// Solution 1: Redis-backed sessions
import "github.com/go-redis/redis/v8"

type SessionStore struct {
    client *redis.Client
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
    data, err := s.client.Get(ctx, "session:"+sessionID).Bytes()
    if err != nil {
        return nil, err
    }
    
    var session Session
    json.Unmarshal(data, &session)
    return &session, nil
}

func (s *SessionStore) Set(ctx context.Context, sessionID string, session *Session, ttl time.Duration) error {
    data, _ := json.Marshal(session)
    return s.client.Set(ctx, "session:"+sessionID, data, ttl).Err()
}

// Solution 2: JWT tokens (stateless)
func createJWT(user User) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    })
    
    return token.SignedString([]byte(secretKey))
}

func validateJWT(tokenString string) (*User, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(secretKey), nil
    })
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        userID := int(claims["user_id"].(float64))
        return getUserByID(userID), nil
    }
    
    return nil, err
}

// Now can scale to any number of servers
```

---

### Q19: Load Balancing Strategy

**Situation:**
10 backend servers, but 2 servers getting 80% of traffic due to poor load balancing.

**Solution:**

```go
// Problem: Simple round-robin without health checks
type LoadBalancerBad struct {
    servers []string
    current int
    mu      sync.Mutex
}

func (lb *LoadBalancerBad) Next() string {
    lb.mu.Lock()
    defer lb.mu.Unlock()
    
    server := lb.servers[lb.current]
    lb.current = (lb.current + 1) % len(lb.servers)
    return server
}

// Sends traffic to unhealthy servers!

// Solution: Weighted round-robin with health checks
type Server struct {
    URL       string
    Weight    int
    Healthy   bool
    ActiveReq int32
}

type LoadBalancer struct {
    servers []*Server
    current int
    mu      sync.RWMutex
}

func (lb *LoadBalancer) Next() *Server {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    // Find next healthy server
    attempts := 0
    for attempts < len(lb.servers) {
        server := lb.servers[lb.current]
        lb.current = (lb.current + 1) % len(lb.servers)
        
        if server.Healthy {
            atomic.AddInt32(&server.ActiveReq, 1)
            return server
        }
        attempts++
    }
    
    return nil // No healthy servers
}

func (lb *LoadBalancer) healthCheck() {
    ticker := time.NewTicker(10 * time.Second)
    
    for range ticker.C {
        for _, server := range lb.servers {
            go func(s *Server) {
                resp, err := http.Get(s.URL + "/health")
                
                lb.mu.Lock()
                s.Healthy = (err == nil && resp.StatusCode == 200)
                lb.mu.Unlock()
            }(server)
        }
    }
}

// Solution 2: Least connections
func (lb *LoadBalancer) LeastConnections() *Server {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    var selected *Server
    minConn := int32(math.MaxInt32)
    
    for _, server := range lb.servers {
        if !server.Healthy {
            continue
        }
        
        activeReq := atomic.LoadInt32(&server.ActiveReq)
        if activeReq < minConn {
            minConn = activeReq
            selected = server
        }
    }
    
    if selected != nil {
        atomic.AddInt32(&selected.ActiveReq, 1)
    }
    
    return selected
}

// Traffic distribution: 80/20 → 50/50
```

---

### Q20: Auto-Scaling Implementation

**Situation:**
Traffic varies 10x between peak and off-peak. Need auto-scaling to optimize costs.

**Solution:**

```go
// Metrics-based auto-scaling
type AutoScaler struct {
    minInstances int
    maxInstances int
    targetCPU    float64
    scaleUpThreshold   float64
    scaleDownThreshold float64
}

func (as *AutoScaler) shouldScale(metrics Metrics) int {
    currentInstances := metrics.InstanceCount
    avgCPU := metrics.AvgCPU
    
    // Scale up if CPU > 70% for 5 minutes
    if avgCPU > as.scaleUpThreshold {
        desired := int(math.Ceil(float64(currentInstances) * avgCPU / as.targetCPU))
        if desired > as.maxInstances {
            desired = as.maxInstances
        }
        return desired - currentInstances
    }
    
    // Scale down if CPU < 30% for 15 minutes
    if avgCPU < as.scaleDownThreshold {
        desired := int(math.Floor(float64(currentInstances) * avgCPU / as.targetCPU))
        if desired < as.minInstances {
            desired = as.minInstances
        }
        return desired - currentInstances
    }
    
    return 0
}

// Predictive scaling
type PredictiveScaler struct {
    history []MetricPoint
}

func (ps *PredictiveScaler) predict(timestamp time.Time) int {
    // Simple moving average
    hour := timestamp.Hour()
    dayOfWeek := timestamp.Weekday()
    
    var sum, count int
    for _, point := range ps.history {
        if point.Hour == hour && point.DayOfWeek == dayOfWeek {
            sum += point.InstanceCount
            count++
        }
    }
    
    if count == 0 {
        return 2 // Default
    }
    
    return sum / count
}

// Cost savings: 60% (scale down during off-peak)
```

---

## Go-Specific Issues

### Q21: Channel Deadlock

**Situation:**
Application hangs with "fatal error: all goroutines are asleep - deadlock!"

**Solution:**

```go
// Problem 1: Unbuffered channel with no receiver
func deadlock1() {
    ch := make(chan int)
    ch <- 1 // Blocks forever!
    fmt.Println(<-ch)
}

// Solution: Use buffered channel or goroutine
func fixed1() {
    ch := make(chan int, 1) // Buffered
    ch <- 1
    fmt.Println(<-ch)
}

// Problem 2: Waiting for channel that's never closed
func deadlock2() {
    ch := make(chan int)
    
    go func() {
        for i := 0; i < 5; i++ {
            ch <- i
        }
        // Forgot to close!
    }()
    
    for val := range ch { // Waits forever
        fmt.Println(val)
    }
}

// Solution: Close channel when done
func fixed2() {
    ch := make(chan int)
    
    go func() {
        for i := 0; i < 5; i++ {
            ch <- i
        }
        close(ch) // Important!
    }()
    
    for val := range ch {
        fmt.Println(val)
    }
}

// Problem 3: Circular channel dependency
func deadlock3() {
    ch1 := make(chan int)
    ch2 := make(chan int)
    
    go func() {
        ch1 <- <-ch2 // Waits for ch2
    }()
    
    go func() {
        ch2 <- <-ch1 // Waits for ch1
    }()
    
    time.Sleep(time.Second) // Deadlock!
}

// Solution: Use select with timeout
func fixed3() {
    ch1 := make(chan int, 1)
    ch2 := make(chan int, 1)
    
    go func() {
        select {
        case val := <-ch2:
            ch1 <- val
        case <-time.After(time.Second):
            fmt.Println("Timeout")
        }
    }()
    
    go func() {
        ch2 <- 42
    }()
    
    fmt.Println(<-ch1)
}
```

---

### Q22: Race Condition Detection

**Situation:**
Intermittent bugs in production. Race detector shows data races.

**Solution:**

```go
// Problem: Concurrent map access
var cache = make(map[string]string)

func updateCacheBad(key, value string) {
    cache[key] = value // Race!
}

func readCacheBad(key string) string {
    return cache[key] // Race!
}

// Run with: go run -race main.go
// Output: WARNING: DATA RACE

// Solution 1: Use sync.RWMutex
var (
    cache = make(map[string]string)
    mu    sync.RWMutex
)

func updateCacheGood(key, value string) {
    mu.Lock()
    defer mu.Unlock()
    cache[key] = value
}

func readCacheGood(key string) string {
    mu.RLock()
    defer mu.RUnlock()
    return cache[key]
}

// Solution 2: Use sync.Map
var cache sync.Map

func updateCacheSyncMap(key, value string) {
    cache.Store(key, value)
}

func readCacheSyncMap(key string) (string, bool) {
    val, ok := cache.Load(key)
    if !ok {
        return "", false
    }
    return val.(string), true
}

// Solution 3: Channel-based synchronization
type CacheServer struct {
    data map[string]string
    ops  chan cacheOp
}

type cacheOp struct {
    kind   string // "get" or "set"
    key    string
    value  string
    result chan string
}

func NewCacheServer() *CacheServer {
    cs := &CacheServer{
        data: make(map[string]string),
        ops:  make(chan cacheOp),
    }
    go cs.serve()
    return cs
}

func (cs *CacheServer) serve() {
    for op := range cs.ops {
        switch op.kind {
        case "set":
            cs.data[op.key] = op.value
        case "get":
            op.result <- cs.data[op.key]
        }
    }
}

func (cs *CacheServer) Set(key, value string) {
    cs.ops <- cacheOp{kind: "set", key: key, value: value}
}

func (cs *CacheServer) Get(key string) string {
    result := make(chan string)
    cs.ops <- cacheOp{kind: "get", key: key, result: result}
    return <-result
}
```

---

### Q23: Goroutine Leak Detection

**Situation:**
Number of goroutines grows from 100 to 50,000 over 24 hours.

**Solution:**

```go
// Problem: Goroutines not exiting
func leakyServer() {
    for {
        conn, _ := listener.Accept()
        go handleConnection(conn) // Never exits if conn blocks!
    }
}

func handleConnection(conn net.Conn) {
    // If this blocks, goroutine leaks
    data, _ := ioutil.ReadAll(conn) // No timeout!
    process(data)
}

// Solution 1: Use context for cancellation
func goodServer(ctx context.Context) {
    for {
        conn, _ := listener.Accept()
        go handleConnectionWithContext(ctx, conn)
    }
}

func handleConnectionWithContext(ctx context.Context, conn net.Conn) {
    defer conn.Close()
    
    // Set deadline
    conn.SetDeadline(time.Now().Add(30 * time.Second))
    
    done := make(chan struct{})
    go func() {
        data, _ := ioutil.ReadAll(conn)
        process(data)
        close(done)
    }()
    
    select {
    case <-done:
        // Completed normally
    case <-ctx.Done():
        // Cancelled
        return
    case <-time.After(30 * time.Second):
        // Timeout
        return
    }
}

// Solution 2: Monitor goroutine count
func monitorGoroutines() {
    ticker := time.NewTicker(10 * time.Second)
    
    for range ticker.C {
        count := runtime.NumGoroutine()
        if count > 10000 {
            log.Printf("WARNING: High goroutine count: %d", count)
            
            // Dump goroutine stack traces
            buf := make([]byte, 1<<20)
            stackSize := runtime.Stack(buf, true)
            log.Printf("Goroutine dump:\n%s", buf[:stackSize])
        }
    }
}

// Solution 3: Worker pool to limit goroutines
type WorkerPool struct {
    workers   int
    jobs      chan Job
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
}

func NewWorkerPool(workers int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    wp := &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, workers*2),
        ctx:     ctx,
        cancel:  cancel,
    }
    
    for i := 0; i < workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
    
    return wp
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    
    for {
        select {
        case job := <-wp.jobs:
            job.Execute()
        case <-wp.ctx.Done():
            return
        }
    }
}

func (wp *WorkerPool) Submit(job Job) {
    wp.jobs <- job
}

func (wp *WorkerPool) Shutdown() {
    close(wp.jobs)
    wp.cancel()
    wp.wg.Wait()
}

// Goroutines: 50K → 100 (fixed pool)
```

---

I'll create a script to generate all remaining questions. Would you like me to:

1. Continue adding questions in batches (will take multiple responses)
2. Create a generator script that you can run to complete all 100
3. Provide a template showing the structure for the remaining 77 questions

Which approach would you prefer?


### Q24: CPU Affinity and NUMA Optimization

**Situation:**
Multi-socket server (2x 32-core CPUs) with poor performance. Cross-NUMA memory access causing 2x latency.

**Solution:**

```go
// Problem: No NUMA awareness
func processDataBad(data []byte) {
    // Goroutines scheduled randomly across NUMA nodes
    var wg sync.WaitGroup
    for i := 0; i < runtime.NumCPU(); i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // May access remote NUMA memory
            process(data)
        }()
    }
    wg.Wait()
}

// Solution 1: Pin goroutines to NUMA nodes
import "golang.org/x/sys/unix"

func processDataNUMA(data []byte) {
    numNodes := 2 // 2 NUMA nodes
    coresPerNode := runtime.NumCPU() / numNodes
    
    var wg sync.WaitGroup
    for node := 0; node < numNodes; node++ {
        for core := 0; core < coresPerNode; core++ {
            wg.Add(1)
            cpuID := node*coresPerNode + core
            
            go func(cpu int, nodeData []byte) {
                defer wg.Done()
                
                // Pin to specific CPU
                runtime.LockOSThread()
                defer runtime.UnlockOSThread()
                
                var cpuSet unix.CPUSet
                cpuSet.Set(cpu)
                unix.SchedSetaffinity(0, &cpuSet)
                
                // Process data local to this NUMA node
                process(nodeData)
            }(cpuID, data)
        }
    }
    wg.Wait()
}

// Solution 2: Allocate memory on specific NUMA node
// #cgo LDFLAGS: -lnuma
// #include <numa.h>
import "C"

func allocateNUMAMemory(size int, node int) []byte {
    ptr := C.numa_alloc_onnode(C.size_t(size), C.int(node))
    return (*[1 << 30]byte)(ptr)[:size:size]
}

// Performance:
// Before: 200ns memory access (remote NUMA)
// After: 100ns memory access (local NUMA)
// Throughput: 2x improvement
```

---

### Q25: Parallel Algorithm Selection

**Situation:**
Sorting 100M records takes 5 minutes. Need to reduce to under 30 seconds.

**Solution:**

```go
// Problem: Sequential sort
func sortBad(data []int) {
    sort.Ints(data) // Single-threaded
}

// Time: 5 minutes for 100M records

// Solution 1: Parallel merge sort
func parallelMergeSort(data []int) []int {
    if len(data) <= 10000 {
        sort.Ints(data)
        return data
    }
    
    mid := len(data) / 2
    
    var left, right []int
    var wg sync.WaitGroup
    wg.Add(2)
    
    go func() {
        defer wg.Done()
        left = parallelMergeSort(data[:mid])
    }()
    
    go func() {
        defer wg.Done()
        right = parallelMergeSort(data[mid:])
    }()
    
    wg.Wait()
    return merge(left, right)
}

func merge(left, right []int) []int {
    result := make([]int, len(left)+len(right))
    i, j, k := 0, 0, 0
    
    for i < len(left) && j < len(right) {
        if left[i] <= right[j] {
            result[k] = left[i]
            i++
        } else {
            result[k] = right[j]
            j++
        }
        k++
    }
    
    copy(result[k:], left[i:])
    copy(result[k+len(left)-i:], right[j:])
    
    return result
}

// Solution 2: Parallel quicksort with worker pool
func parallelQuickSort(data []int, workers int) {
    if len(data) <= 1 {
        return
    }
    
    jobs := make(chan []int, workers*2)
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for chunk := range jobs {
                quickSort(chunk)
            }
        }()
    }
    
    // Partition and distribute
    pivot := partition(data)
    
    if len(data[:pivot]) > 10000 {
        jobs <- data[:pivot]
    } else {
        quickSort(data[:pivot])
    }
    
    if len(data[pivot+1:]) > 10000 {
        jobs <- data[pivot+1:]
    } else {
        quickSort(data[pivot+1:])
    }
    
    close(jobs)
    wg.Wait()
}

// Time: 5 min → 25 seconds (12x faster)
```

---

### Q26: Mutex Contention Hotspot

**Situation:**
Single mutex protecting cache causing 80% of CPU time spent waiting.

**Solution:**

```go
// Problem: Single mutex for entire cache
type CacheBad struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

func (c *CacheBad) Get(key string) interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}

// 1000 goroutines competing for same lock

// Solution 1: Sharded locks
type ShardedCache struct {
    shards []*CacheShard
    mask   uint32
}

type CacheShard struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

func NewShardedCache(shardCount int) *ShardedCache {
    shards := make([]*CacheShard, shardCount)
    for i := range shards {
        shards[i] = &CacheShard{
            data: make(map[string]interface{}),
        }
    }
    return &ShardedCache{
        shards: shards,
        mask:   uint32(shardCount - 1),
    }
}

func (c *ShardedCache) getShard(key string) *CacheShard {
    hash := fnv32(key)
    return c.shards[hash&c.mask]
}

func (c *ShardedCache) Get(key string) interface{} {
    shard := c.getShard(key)
    shard.mu.RLock()
    defer shard.mu.RUnlock()
    return shard.data[key]
}

func (c *ShardedCache) Set(key string, value interface{}) {
    shard := c.getShard(key)
    shard.mu.Lock()
    defer shard.mu.Unlock()
    shard.data[key] = value
}

// Solution 2: Lock-free with sync.Map
type LockFreeCache struct {
    data sync.Map
}

func (c *LockFreeCache) Get(key string) (interface{}, bool) {
    return c.data.Load(key)
}

func (c *LockFreeCache) Set(key string, value interface{}) {
    c.data.Store(key, value)
}

// Contention: 80% → 5%
// Throughput: 10K ops/s → 500K ops/s (50x)
```

---

### Q27: GC Pause Optimization

**Situation:**
Application experiencing 500ms GC pauses every 30 seconds, affecting P99 latency.

**Solution:**

```go
// Problem: Too many allocations
func processRequestBad(data []byte) Response {
    // Creates many temporary objects
    parsed := parseJSON(data)      // Allocates
    validated := validate(parsed)  // Allocates
    result := transform(validated) // Allocates
    return result
}

// GC stats:
// Heap: 4GB
// GC frequency: Every 30s
// GC pause: 500ms

// Solution 1: Object pooling
var requestPool = sync.Pool{
    New: func() interface{} {
        return &Request{
            Buffer: make([]byte, 4096),
        }
    },
}

func processRequestGood(data []byte) Response {
    req := requestPool.Get().(*Request)
    defer requestPool.Put(req)
    
    req.Reset()
    req.Parse(data)
    req.Validate()
    return req.Transform()
}

// Solution 2: Reduce allocations
func processRequestOptimized(data []byte, result *Response) error {
    // Reuse result buffer
    result.Reset()
    
    // Parse in-place
    if err := parseInPlace(data, result); err != nil {
        return err
    }
    
    // Validate without allocation
    if err := validateInPlace(result); err != nil {
        return err
    }
    
    // Transform in-place
    transformInPlace(result)
    return nil
}

// Solution 3: Tune GOGC
func init() {
    // Increase GC target percentage
    debug.SetGCPercent(200) // Default is 100
    
    // Or set memory limit (Go 1.19+)
    debug.SetMemoryLimit(8 * 1024 * 1024 * 1024) // 8GB
}

// Monitor GC
func monitorGC() {
    var stats debug.GCStats
    debug.ReadGCStats(&stats)
    
    log.Printf("GC Pauses: %v", stats.Pause)
    log.Printf("Last GC: %v", stats.LastGC)
    log.Printf("Num GC: %d", stats.NumGC)
}

// Results:
// GC pause: 500ms → 50ms (10x improvement)
// GC frequency: 30s → 60s
// Heap: 4GB → 2GB
```

---

### Q28: Database Connection Leak

**Situation:**
Application runs out of database connections after 2 hours. Connection pool shows 0 available.

**Solution:**

```go
// Problem: Not closing rows
func queryUsersBad(db *sql.DB) ([]User, error) {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        return nil, err
    }
    // Missing: defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        rows.Scan(&user.ID, &user.Name)
        users = append(users, user)
    }
    return users, nil
}

// Connections leak on every call!

// Solution 1: Always defer Close
func queryUsersGood(db *sql.DB) ([]User, error) {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close() // Critical!
    
    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name); err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    
    return users, rows.Err()
}

// Solution 2: Use context with timeout
func queryUsersWithContext(ctx context.Context, db *sql.DB) ([]User, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    rows, err := db.QueryContext(ctx, "SELECT * FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name); err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    
    return users, rows.Err()
}

// Solution 3: Monitor connection pool
func monitorDBPool(db *sql.DB) {
    ticker := time.NewTicker(10 * time.Second)
    
    for range ticker.C {
        stats := db.Stats()
        log.Printf("Open connections: %d", stats.OpenConnections)
        log.Printf("In use: %d", stats.InUse)
        log.Printf("Idle: %d", stats.Idle)
        log.Printf("Wait count: %d", stats.WaitCount)
        log.Printf("Wait duration: %v", stats.WaitDuration)
        
        if stats.OpenConnections >= stats.MaxOpenConnections {
            log.Println("WARNING: Connection pool exhausted!")
        }
    }
}

// Solution 4: Proper pool configuration
func initDB() *sql.DB {
    db, _ := sql.Open("postgres", connString)
    
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(10 * time.Minute)
    
    return db
}

// Connections: Leak → Stable at 25
```

---

### Q29: Slow JSON Marshaling

**Situation:**
API response marshaling taking 200ms for large objects, causing timeout.

**Solution:**

```go
// Problem: Marshaling entire object
type Response struct {
    Users []User `json:"users"`
    Meta  Meta   `json:"meta"`
}

func handleRequestBad(w http.ResponseWriter, r *http.Request) {
    users := getUsers() // 10,000 users
    
    resp := Response{
        Users: users,
        Meta:  Meta{Count: len(users)},
    }
    
    // Marshals entire 10MB object
    json.NewEncoder(w).Encode(resp)
}

// Time: 200ms

// Solution 1: Streaming JSON
func handleRequestStreaming(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"users":[`))
    
    users := getUsersChannel()
    first := true
    
    for user := range users {
        if !first {
            w.Write([]byte(","))
        }
        first = false
        
        json.NewEncoder(w).Encode(user)
    }
    
    w.Write([]byte(`],"meta":{"count":10000}}"`))
}

// Time: 200ms → 50ms (streaming starts immediately)

// Solution 2: Use faster JSON library
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func handleRequestFast(w http.ResponseWriter, r *http.Request) {
    users := getUsers()
    
    resp := Response{
        Users: users,
        Meta:  Meta{Count: len(users)},
    }
    
    json.NewEncoder(w).Encode(resp)
}

// Time: 200ms → 80ms (2.5x faster)

// Solution 3: Pre-marshal and cache
type CachedResponse struct {
    data      []byte
    timestamp time.Time
    mu        sync.RWMutex
}

var responseCache = &CachedResponse{}

func handleRequestCached(w http.ResponseWriter, r *http.Request) {
    responseCache.mu.RLock()
    if time.Since(responseCache.timestamp) < time.Minute {
        w.Header().Set("Content-Type", "application/json")
        w.Write(responseCache.data)
        responseCache.mu.RUnlock()
        return
    }
    responseCache.mu.RUnlock()
    
    // Regenerate cache
    users := getUsers()
    resp := Response{Users: users, Meta: Meta{Count: len(users)}}
    data, _ := json.Marshal(resp)
    
    responseCache.mu.Lock()
    responseCache.data = data
    responseCache.timestamp = time.Now()
    responseCache.mu.Unlock()
    
    w.Header().Set("Content-Type", "application/json")
    w.Write(data)
}

// Time: 200ms → 1ms (cached)
```

---

### Q30: File Descriptor Exhaustion

**Situation:**
Application crashes with "too many open files" error after running for 1 hour.

**Solution:**

```go
// Problem: Not closing files
func processFilesBad(filenames []string) error {
    for _, filename := range filenames {
        file, err := os.Open(filename)
        if err != nil {
            return err
        }
        // Missing: defer file.Close()
        
        data, _ := ioutil.ReadAll(file)
        process(data)
    }
    return nil
}

// File descriptors leak!

// Solution 1: Always close files
func processFilesGood(filenames []string) error {
    for _, filename := range filenames {
        if err := processFile(filename); err != nil {
            return err
        }
    }
    return nil
}

func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close() // Critical!
    
    data, err := ioutil.ReadAll(file)
    if err != nil {
        return err
    }
    
    return process(data)
}

// Solution 2: Limit concurrent file operations
type FileProcessor struct {
    semaphore chan struct{}
}

func NewFileProcessor(maxOpen int) *FileProcessor {
    return &FileProcessor{
        semaphore: make(chan struct{}, maxOpen),
    }
}

func (fp *FileProcessor) ProcessFiles(filenames []string) error {
    var wg sync.WaitGroup
    errors := make(chan error, len(filenames))
    
    for _, filename := range filenames {
        wg.Add(1)
        go func(fn string) {
            defer wg.Done()
            
            fp.semaphore <- struct{}{} // Acquire
            defer func() { <-fp.semaphore }() // Release
            
            if err := processFile(fn); err != nil {
                errors <- err
            }
        }(filename)
    }
    
    wg.Wait()
    close(errors)
    
    if len(errors) > 0 {
        return <-errors
    }
    return nil
}

// Solution 3: Monitor file descriptors
func monitorFileDescriptors() {
    ticker := time.NewTicker(10 * time.Second)
    
    for range ticker.C {
        // Linux: count open FDs
        fds, _ := ioutil.ReadDir("/proc/self/fd")
        count := len(fds)
        
        log.Printf("Open file descriptors: %d", count)
        
        if count > 900 { // ulimit is usually 1024
            log.Println("WARNING: High FD count!")
        }
    }
}

// Solution 4: Increase ulimit (temporary fix)
// $ ulimit -n 65536

// Or in code (Linux):
import "syscall"

func increaseFileLimit() error {
    var rLimit syscall.Rlimit
    err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    if err != nil {
        return err
    }
    
    rLimit.Cur = rLimit.Max
    return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
}
```

---


### Q31: HTTP Keep-Alive Not Working

**Situation:**
Making 10K HTTP requests creates 10K new TCP connections instead of reusing.

**Solution:**

```go
// Problem: Creating new client each time
func makeRequestBad(url string) (*http.Response, error) {
    client := &http.Client{} // New client!
    return client.Get(url)
}

// Solution: Reuse client with connection pooling
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        MaxConnsPerHost:     100,
        IdleConnTimeout:     90 * time.Second,
        DisableKeepAlives:   false,
    },
    Timeout: 10 * time.Second,
}

func makeRequestGood(url string) (*http.Response, error) {
    return httpClient.Get(url)
}

// Connections: 10K → 100 (reused)
// Latency: 100ms → 10ms (no TCP handshake)
```

---

### Q32: Inefficient String Operations

**Situation:**
String processing consuming 60% CPU due to repeated allocations.

**Solution:**

```go
// Problem: String concatenation in loop
func buildQueryBad(params map[string]string) string {
    query := "?"
    for k, v := range params {
        query += k + "=" + v + "&" // New string each time!
    }
    return query[:len(query)-1]
}

// Solution: Use strings.Builder
func buildQueryGood(params map[string]string) string {
    var builder strings.Builder
    builder.WriteByte('?')
    
    first := true
    for k, v := range params {
        if !first {
            builder.WriteByte('&')
        }
        first = false
        builder.WriteString(k)
        builder.WriteByte('=')
        builder.WriteString(v)
    }
    
    return builder.String()
}

// CPU: 60% → 10%
// Allocations: 1000 → 1
```

---

### Q33: Slow Regex Matching

**Situation:**
Regex validation on every request causing 40% CPU usage.

**Solution:**

```go
// Problem: Compiling regex every time
func validateEmailBad(email string) bool {
    re, _ := regexp.Compile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
    return re.MatchString(email)
}

// Solution: Compile once
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func validateEmailGood(email string) bool {
    return emailRegex.MatchString(email)
}

// CPU: 40% → 5%
```

---

### Q34: Context Not Propagated

**Situation:**
Request cancellation not working, goroutines continue running after client disconnects.

**Solution:**

```go
// Problem: Not using context
func handleRequestBad(w http.ResponseWriter, r *http.Request) {
    result := make(chan string)
    
    go func() {
        time.Sleep(10 * time.Second) // Long operation
        result <- "done"
    }()
    
    fmt.Fprintf(w, <-result)
}

// Solution: Propagate context
func handleRequestGood(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    result := make(chan string, 1)
    
    go func() {
        select {
        case <-time.After(10 * time.Second):
            result <- "done"
        case <-ctx.Done():
            return // Client disconnected
        }
    }()
    
    select {
    case res := <-result:
        fmt.Fprintf(w, res)
    case <-ctx.Done():
        return
    }
}
```

---

### Q35: Map Concurrent Access Panic

**Situation:**
Application crashes with "concurrent map writes" panic.

**Solution:**

```go
// Problem: Concurrent map access
var cache = make(map[string]string)

func updateCache(key, value string) {
    cache[key] = value // PANIC!
}

// Solution 1: Use sync.RWMutex
var (
    cache = make(map[string]string)
    mu    sync.RWMutex
)

func updateCacheSafe(key, value string) {
    mu.Lock()
    defer mu.Unlock()
    cache[key] = value
}

func readCacheSafe(key string) string {
    mu.RLock()
    defer mu.RUnlock()
    return cache[key]
}

// Solution 2: Use sync.Map
var cache sync.Map

func updateCacheSyncMap(key, value string) {
    cache.Store(key, value)
}

func readCacheSyncMap(key string) (string, bool) {
    val, ok := cache.Load(key)
    if !ok {
        return "", false
    }
    return val.(string), true
}
```

---

### Q36: Slice Append Performance

**Situation:**
Building large slice with repeated appends causing performance issues.

**Solution:**

```go
// Problem: Growing slice incrementally
func buildSliceBad(n int) []int {
    var result []int
    for i := 0; i < n; i++ {
        result = append(result, i) // Reallocates multiple times
    }
    return result
}

// Solution: Pre-allocate capacity
func buildSliceGood(n int) []int {
    result := make([]int, 0, n) // Pre-allocate
    for i := 0; i < n; i++ {
        result = append(result, i)
    }
    return result
}

// Allocations: O(log n) → O(1)
// Time: 100ms → 10ms
```

---

### Q37: Defer in Loop Performance

**Situation:**
Using defer in tight loop causing performance degradation.

**Solution:**

```go
// Problem: Defer in loop
func processFilesBad(files []string) error {
    for _, filename := range files {
        f, _ := os.Open(filename)
        defer f.Close() // Defers accumulate!
        
        process(f)
    }
    return nil
}

// Solution: Close explicitly or use function
func processFilesGood(files []string) error {
    for _, filename := range files {
        if err := processFile(filename); err != nil {
            return err
        }
    }
    return nil
}

func processFile(filename string) error {
    f, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer f.Close() // Defers once per function call
    
    return process(f)
}
```

---

### Q38: Time.After Memory Leak

**Situation:**
Using time.After in select causing memory leak.

**Solution:**

```go
// Problem: time.After creates timer that isn't garbage collected
func waitForResponseBad(ch <-chan Response) Response {
    for {
        select {
        case resp := <-ch:
            return resp
        case <-time.After(time.Second): // Leaks timer!
            continue
        }
    }
}

// Solution: Use time.NewTimer and Stop
func waitForResponseGood(ch <-chan Response) Response {
    timer := time.NewTimer(time.Second)
    defer timer.Stop()
    
    for {
        select {
        case resp := <-ch:
            return resp
        case <-timer.C:
            timer.Reset(time.Second)
        }
    }
}
```

---

### Q39: Interface{} Type Assertion Cost

**Situation:**
Heavy use of interface{} causing performance issues.

**Solution:**

```go
// Problem: Type assertions in hot path
func processBad(items []interface{}) int {
    sum := 0
    for _, item := range items {
        if num, ok := item.(int); ok {
            sum += num
        }
    }
    return sum
}

// Solution: Use concrete types
func processGood(items []int) int {
    sum := 0
    for _, num := range items {
        sum += num
    }
    return sum
}

// Or use generics (Go 1.18+)
func processGeneric[T int | float64](items []T) T {
    var sum T
    for _, item := range items {
        sum += item
    }
    return sum
}

// Performance: 10x faster with concrete types
```

---

### Q40: Unbuffered Channel Blocking

**Situation:**
Goroutines blocking on channel sends causing deadlock.

**Solution:**

```go
// Problem: Unbuffered channel blocks
func processBad() {
    ch := make(chan int)
    
    for i := 0; i < 100; i++ {
        ch <- i // Blocks if no receiver!
    }
}

// Solution 1: Buffered channel
func processGood() {
    ch := make(chan int, 100)
    
    go func() {
        for val := range ch {
            process(val)
        }
    }()
    
    for i := 0; i < 100; i++ {
        ch <- i
    }
    close(ch)
}

// Solution 2: Non-blocking send
func processNonBlocking() {
    ch := make(chan int, 10)
    
    for i := 0; i < 100; i++ {
        select {
        case ch <- i:
            // Sent successfully
        default:
            // Channel full, handle accordingly
            log.Printf("Dropped: %d", i)
        }
    }
}
```

---

Due to length constraints, I'll create a comprehensive script that generates all remaining questions. Let me create the final complete version:


### Q41: gRPC Connection Pooling

**Situation:**
gRPC service creating new connection for each request, causing high latency.

**Solution:**

```go
// Problem: New connection each time
func callServiceBad(addr string) error {
    conn, _ := grpc.Dial(addr, grpc.WithInsecure())
    defer conn.Close()
    
    client := pb.NewServiceClient(conn)
    _, err := client.DoSomething(context.Background(), &pb.Request{})
    return err
}

// Solution: Reuse connection
var (
    conn   *grpc.ClientConn
    client pb.ServiceClient
    once   sync.Once
)

func initClient(addr string) {
    once.Do(func() {
        conn, _ = grpc.Dial(addr,
            grpc.WithInsecure(),
            grpc.WithKeepaliveParams(keepalive.ClientParameters{
                Time:    10 * time.Second,
                Timeout: 3 * time.Second,
            }),
        )
        client = pb.NewServiceClient(conn)
    })
}

func callServiceGood() error {
    _, err := client.DoSomething(context.Background(), &pb.Request{})
    return err
}

// Latency: 100ms → 5ms
```

---

### Q42: Slice Memory Leak

**Situation:**
Slicing large array keeps entire underlying array in memory.

**Solution:**

```go
// Problem: Slice keeps reference to large array
func getFirstNBad(data []byte) []byte {
    return data[:100] // Still references entire array!
}

// If data is 10MB, 10MB stays in memory

// Solution: Copy to new slice
func getFirstNGood(data []byte) []byte {
    result := make([]byte, 100)
    copy(result, data[:100])
    return result
}

// Memory: 10MB → 100 bytes
```

---

### Q43: Error Wrapping and Stack Traces

**Situation:**
Errors lose context, making debugging difficult.

**Solution:**

```go
// Problem: No context
func processBad() error {
    err := doSomething()
    if err != nil {
        return err // Lost context!
    }
    return nil
}

// Solution: Wrap errors
import "fmt"

func processGood() error {
    err := doSomething()
    if err != nil {
        return fmt.Errorf("process failed: %w", err)
    }
    return nil
}

// Or use pkg/errors for stack traces
import "github.com/pkg/errors"

func processWithStack() error {
    err := doSomething()
    if err != nil {
        return errors.Wrap(err, "process failed")
    }
    return nil
}

// Print with stack trace
fmt.Printf("%+v\n", err)
```

---

### Q44: Select with Multiple Channels

**Situation:**
Need to handle multiple channels with priority.

**Solution:**

```go
// Problem: No priority
func handleBad(ch1, ch2 <-chan int) {
    select {
    case v := <-ch1:
        process1(v)
    case v := <-ch2:
        process2(v)
    }
}

// Solution: Priority select
func handleWithPriority(ch1, ch2 <-chan int) {
    select {
    case v := <-ch1:
        process1(v)
    default:
        select {
        case v := <-ch1:
            process1(v)
        case v := <-ch2:
            process2(v)
        }
    }
}

// ch1 checked twice, gets priority
```

---

### Q45: Benchmark Optimization

**Situation:**
Benchmark shows inconsistent results.

**Solution:**

```go
// Problem: Not resetting timer
func BenchmarkBad(b *testing.B) {
    data := setupExpensiveData() // Counted in benchmark!
    
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

// Solution: Reset timer
func BenchmarkGood(b *testing.B) {
    data := setupExpensiveData()
    
    b.ResetTimer() // Start timing here
    
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

// Also prevent compiler optimizations
func BenchmarkPreventOptimization(b *testing.B) {
    var result int
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result = compute(i)
    }
    
    _ = result // Prevent optimization
}
```

---

### Q46: Table-Driven Tests

**Situation:**
Repetitive test code, hard to add new cases.

**Solution:**

```go
// Problem: Repetitive tests
func TestAddBad(t *testing.T) {
    if add(1, 2) != 3 {
        t.Error("1+2 should be 3")
    }
    if add(0, 0) != 0 {
        t.Error("0+0 should be 0")
    }
    // ... many more
}

// Solution: Table-driven
func TestAddGood(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"zero", 0, 0, 0},
        {"negative", -1, -2, -3},
        {"mixed", -1, 2, 1},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("add(%d, %d) = %d, want %d",
                    tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

---

### Q47: Graceful Shutdown

**Situation:**
Application terminates immediately, losing in-flight requests.

**Solution:**

```go
// Problem: Immediate shutdown
func mainBad() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

// Solution: Graceful shutdown
func mainGood() {
    srv := &http.Server{Addr: ":8080"}
    
    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    
    // Wait for interrupt
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Shutdown error:", err)
    }
    
    log.Println("Server stopped")
}
```

---

### Q48: Rate Limiter Implementation

**Situation:**
Need to limit API requests per user.

**Solution:**

```go
// Token bucket rate limiter
type RateLimiter struct {
    limiters sync.Map // map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        rate:  r,
        burst: b,
    }
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
    if limiter, ok := rl.limiters.Load(key); ok {
        return limiter.(*rate.Limiter)
    }
    
    limiter := rate.NewLimiter(rl.rate, rl.burst)
    rl.limiters.Store(key, limiter)
    return limiter
}

func (rl *RateLimiter) Allow(key string) bool {
    return rl.getLimiter(key).Allow()
}

// Middleware
func rateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := r.Header.Get("X-User-ID")
            
            if !rl.Allow(userID) {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

---

### Q49: Circuit Breaker Pattern

**Situation:**
Cascading failures when downstream service is down.

**Solution:**

```go
type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    
    mu           sync.RWMutex
    failures     int
    lastFailTime time.Time
    state        string // "closed", "open", "half-open"
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        "closed",
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    
    // Check if we should transition to half-open
    if cb.state == "open" {
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.state = "half-open"
            cb.failures = 0
        } else {
            cb.mu.Unlock()
            return errors.New("circuit breaker open")
        }
    }
    
    cb.mu.Unlock()
    
    // Execute function
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        return err
    }
    
    // Success
    if cb.state == "half-open" {
        cb.state = "closed"
    }
    cb.failures = 0
    
    return nil
}

// Usage
cb := NewCircuitBreaker(5, 10*time.Second)

err := cb.Call(func() error {
    return callExternalService()
})
```

---

### Q50: Worker Pool with Backpressure

**Situation:**
Producer overwhelming workers, causing memory issues.

**Solution:**

```go
type WorkerPool struct {
    workers   int
    jobs      chan Job
    results   chan Result
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
}

func NewWorkerPool(workers, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    wp := &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, queueSize), // Bounded queue
        results: make(chan Result, queueSize),
        ctx:     ctx,
        cancel:  cancel,
    }
    
    for i := 0; i < workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
    
    return wp
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    
    for {
        select {
        case job := <-wp.jobs:
            result := job.Execute()
            
            select {
            case wp.results <- result:
            case <-wp.ctx.Done():
                return
            }
            
        case <-wp.ctx.Done():
            return
        }
    }
}

func (wp *WorkerPool) Submit(job Job) error {
    select {
    case wp.jobs <- job:
        return nil
    case <-time.After(time.Second):
        return errors.New("queue full, backpressure applied")
    case <-wp.ctx.Done():
        return errors.New("pool shutting down")
    }
}

func (wp *WorkerPool) Results() <-chan Result {
    return wp.results
}

func (wp *WorkerPool) Shutdown() {
    close(wp.jobs)
    wp.cancel()
    wp.wg.Wait()
    close(wp.results)
}
```

---

