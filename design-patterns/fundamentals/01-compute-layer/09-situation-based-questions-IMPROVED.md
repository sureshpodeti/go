# 100 Situation-Based Questions for Software Architects

## Overview

This document contains 100 real-world scenario-based questions with **detailed explanations** covering:
- CPU-bound workloads
- Memory-bound workloads
- I/O-bound workloads
- Scaling strategies
- Go-specific issues
- Debugging techniques
- Production troubleshooting

**Each question includes:**
1. **Problem Definition** - What's wrong and why
2. **Root Cause Analysis** - Technical explanation with definitions
3. **Solution Explanation** - How to fix it (in words)
4. **Code Examples** - Implementation with comments
5. **Metrics & Results** - Before/after comparison
6. **Key Takeaways** - Lessons learned

---

## Table of Contents

1. [CPU-Bound Scenarios](#cpu-bound-scenarios)
2. [Memory-Bound Scenarios](#memory-bound-scenarios)
3. [I/O-Bound Scenarios](#io-bound-scenarios)
4. [Scaling Scenarios](#scaling-scenarios)
5. [Go-Specific Issues](#go-specific-issues)
6. [Debugging & Monitoring](#debugging-monitoring)
7. [Data Structures](#data-structures)

---

## CPU-Bound Scenarios

### Q1: High CPU Usage in Image Processing Service

**Situation:**
Your Go-based image processing service is experiencing 95% CPU usage. It processes 1000 images/minute, resizing them from 4K to multiple resolutions (thumbnail, small, medium, large, original). Response time has increased from 200ms to 2 seconds.

**Problem Definition:**

The service creates **too many goroutines** competing for CPU resources. For every image, it spawns 5 goroutines (one per size), resulting in 5000+ concurrent goroutines when processing 1000 images. This causes:
- **Context switching overhead**: The OS scheduler constantly switches between goroutines
- **CPU thrashing**: More time spent switching than doing actual work
- **Cache misses**: Each context switch invalidates CPU caches

**Root Cause Analysis:**

**Goroutine explosion** occurs when you create more goroutines than your CPU can handle efficiently. While goroutines are lightweight (~2KB stack), having thousands of them competing for 8 CPU cores creates contention.

**Key concept**: For **CPU-bound tasks** (like image processing), the optimal number of goroutines = number of CPU cores. Creating more goroutines doesn't increase throughput—it decreases it due to context switching.

**Why this happens:**
```
1000 images × 5 sizes = 5000 goroutines
8 CPU cores trying to run 5000 goroutines
= 625 goroutines per core
= Constant context switching
= High CPU usage but low actual work done
```

**Solution Explanation:**

Use the **Worker Pool Pattern** to limit concurrent goroutines to match CPU cores:

1. **Create a fixed pool of workers** (one per CPU core)
2. **Queue incoming jobs** in a buffered channel
3. **Workers pull jobs** from the queue and process them sequentially
4. **No goroutine explosion** - only 8 goroutines running

This ensures:
- Each CPU core runs one goroutine continuously
- No context switching overhead
- Better CPU cache utilization
- Predictable resource usage

**Code Implementation:**

```go
// ❌ PROBLEM: Too many goroutines competing for CPU
func processImageBad(img image.Image) []image.Image {
    var wg sync.WaitGroup
    results := make([]image.Image, 5)
    
    // Creates 5 goroutines per image = 5000 goroutines total!
    for i, size := range sizes {
        wg.Add(1)
        go func(idx int, s Size) {
            defer wg.Done()
            results[idx] = resize(img, s) // CPU-intensive work
        }(i, size)
    }
    wg.Wait()
    return results
}

// ✅ SOLUTION: Worker pool pattern with limited goroutines
type ImageProcessor struct {
    jobs    chan ImageJob
    workers int
}

type ImageJob struct {
    Image      image.Image
    ResultChan chan []image.Image
}

func NewImageProcessor(workers int) *ImageProcessor {
    p := &ImageProcessor{
        jobs:    make(chan ImageJob, 100), // Buffered channel for queuing
        workers: workers,
    }
    
    // Create exactly 'workers' number of goroutines (typically = CPU cores)
    for i := 0; i < workers; i++ {
        go p.worker()
    }
    return p
}

func (p *ImageProcessor) worker() {
    // Each worker processes jobs sequentially from the queue
    for job := range p.jobs {
        // Process all sizes sequentially in ONE goroutine
        results := make([]image.Image, len(sizes))
        for i, size := range sizes {
            results[i] = resize(job.Image, size)
        }
        job.ResultChan <- results
    }
}

func (p *ImageProcessor) Process(img image.Image) []image.Image {
    resultChan := make(chan []image.Image, 1)
    p.jobs <- ImageJob{
        Image:      img,
        ResultChan: resultChan,
    }
    return <-resultChan
}

func main() {
    // Use runtime.NumCPU() to match hardware
    processor := NewImageProcessor(runtime.NumCPU()) // 8 workers for 8 cores
    
    // Now only 8 goroutines doing actual work
    // Each goroutine stays on its CPU core
}
```

**Metrics & Results:**

```
Before (Bad):
├─ Goroutines: 5000+
├─ CPU Usage: 95% (but mostly context switching)
├─ Actual throughput: 500 images/min
├─ Latency: 2000ms per image
└─ Context switches: 50,000/sec

After (Good):
├─ Goroutines: 8 (one per core)
├─ CPU Usage: 85% (actual work)
├─ Actual throughput: 1000 images/min
├─ Latency: 300ms per image
└─ Context switches: 1,000/sec
```

**Key Takeaways:**

1. **CPU-bound tasks**: Limit goroutines to `runtime.NumCPU()`
2. **Worker pool pattern**: Essential for controlling concurrency
3. **Buffered channels**: Queue work without blocking producers
4. **Measure goroutines**: Use `runtime.NumGoroutine()` to detect explosion
5. **Profile CPU**: Use `go tool pprof` to find context switching overhead

---

### Q2: Worker Pool with Backpressure

**Situation:**
Your job processing system has producers submitting jobs faster than workers can process them. Memory usage grows from 500MB to 8GB over 2 hours, eventually causing OOM (Out of Memory) crashes.

**Problem Definition:**

**Backpressure** is missing. When producers create jobs faster than consumers can process them, jobs accumulate in memory indefinitely. This is called **unbounded queue growth**.

**What is Backpressure?**
Backpressure is a mechanism to **slow down or reject producers** when consumers can't keep up. Without it, the system accepts unlimited work, leading to memory exhaustion.

**Root Cause Analysis:**

The problem occurs with **unbuffered or unbounded channels**:

```go
// ❌ Unbounded queue - jobs accumulate forever
jobs := make(chan Job) // Unbuffered channel

// Producer keeps adding jobs
for job := range incomingJobs {
    jobs <- job // Blocks if no worker available, but doesn't reject
}
```

**Why this causes memory issues:**

1. **Producer rate > Consumer rate**: 1000 jobs/sec produced, 500 jobs/sec processed
2. **Queue grows**: 500 jobs/sec accumulate in memory
3. **Memory leak**: After 2 hours = 500 × 7200 = 3.6 million jobs in memory
4. **OOM crash**: System runs out of memory

**Key concept**: **Bounded queues** with **timeout-based rejection** implement backpressure.

**Solution Explanation:**

Implement backpressure using:

1. **Bounded channel** (fixed size queue): `make(chan Job, queueSize)`
2. **Timeout on submission**: Reject jobs if queue is full for too long
3. **Error handling**: Return error to producer so it can retry or drop
4. **Monitoring**: Track queue depth and rejection rate

This ensures:
- **Memory is bounded**: Queue can't grow beyond `queueSize`
- **Producers are notified**: They know when system is overloaded
- **Graceful degradation**: System stays stable under load

**Code Implementation:**

```go
// ❌ PROBLEM: No backpressure - unbounded queue growth
type BadWorkerPool struct {
    jobs chan Job // Unbuffered - can grow indefinitely in memory
}

func (wp *BadWorkerPool) Submit(job Job) {
    wp.jobs <- job // Blocks forever if workers are slow
    // No way to reject or timeout
    // Memory keeps growing
}

// ✅ SOLUTION: Bounded queue with backpressure
type WorkerPool struct {
    workers   int
    jobs      chan Job      // Bounded channel - fixed size
    results   chan Result   // Bounded results channel
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
}

type Job interface {
    Execute() Result
}

type Result struct {
    Data  interface{}
    Error error
}

func NewWorkerPool(workers, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    wp := &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, queueSize),    // BOUNDED: max queueSize jobs
        results: make(chan Result, queueSize), // BOUNDED: max queueSize results
        ctx:     ctx,
        cancel:  cancel,
    }
    
    // Start fixed number of workers
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
            // Process job
            result := job.Execute()
            
            // Send result with context cancellation support
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

// Submit with backpressure - returns error if queue is full
func (wp *WorkerPool) Submit(job Job) error {
    select {
    case wp.jobs <- job:
        // Successfully queued
        return nil
        
    case <-time.After(time.Second):
        // BACKPRESSURE: Queue full for 1 second - reject job
        return errors.New("queue full, backpressure applied")
        
    case <-wp.ctx.Done():
        // Pool is shutting down
        return errors.New("pool shutting down")
    }
}

func (wp *WorkerPool) Results() <-chan Result {
    return wp.results
}

func (wp *WorkerPool) Shutdown() {
    close(wp.jobs)    // No more jobs accepted
    wp.cancel()       // Signal workers to stop
    wp.wg.Wait()      // Wait for workers to finish
    close(wp.results) // Close results channel
}

// Usage example
func main() {
    // Create pool: 10 workers, queue size 100
    pool := NewWorkerPool(10, 100)
    
    // Producer with backpressure handling
    for job := range incomingJobs {
        err := pool.Submit(job)
        if err != nil {
            // Backpressure applied - handle gracefully
            log.Printf("Job rejected: %v", err)
            // Options:
            // 1. Retry later
            // 2. Drop job
            // 3. Store in database for later processing
            // 4. Return 503 Service Unavailable to client
        }
    }
    
    pool.Shutdown()
}
```

**Metrics & Results:**

```
Before (No Backpressure):
├─ Queue size: Unbounded (grows to millions)
├─ Memory usage: 500MB → 8GB → OOM crash
├─ Job acceptance rate: 100% (accepts everything)
├─ System stability: Crashes after 2 hours
└─ Producer feedback: None (doesn't know system is overloaded)

After (With Backpressure):
├─ Queue size: Bounded (max 100 jobs)
├─ Memory usage: Stable at 500MB
├─ Job acceptance rate: 95% (rejects 5% during peak)
├─ System stability: Runs indefinitely
└─ Producer feedback: Errors when overloaded (can retry/drop)
```

**Key Takeaways:**

1. **Backpressure definition**: Mechanism to slow down/reject producers when consumers can't keep up
2. **Bounded channels**: Use `make(chan T, size)` to limit queue growth
3. **Timeout-based rejection**: Use `select` with `time.After()` to implement backpressure
4. **Error handling**: Return errors to producers so they can handle overload
5. **Monitoring**: Track queue depth (`len(jobs)`) and rejection rate
6. **Graceful degradation**: Better to reject some requests than crash the entire system

---

