# Quick Reference Guide - Compute Layer

## Essential Concepts at a Glance

### CPU Metrics

```
Clock Speed:     2.0-5.0 GHz (billions of cycles/sec)
Cores:           Physical processing units
Threads:         Virtual cores (hyperthreading)
Cache L1:        32-64 KB, ~1 ns
Cache L2:        256-512 KB, ~3 ns
Cache L3:        8-64 MB, ~10 ns
IPC:             Instructions per cycle (>2.0 is good)
```

### Memory Hierarchy

```
Registers:       <1 ns, ~100 bytes
L1 Cache:        ~1 ns, 32-64 KB
L2 Cache:        ~3 ns, 256-512 KB
L3 Cache:        ~10 ns, 8-64 MB
RAM:             ~100 ns, 8-512 GB
SSD:             ~100 μs, 256 GB-8 TB
HDD:             ~10 ms, 1-20 TB
```

### Storage Performance

```
HDD (7200 RPM):
• IOPS: 80-160
• Throughput: 100-200 MB/s
• Latency: 5-15 ms

SATA SSD:
• IOPS: 90K-100K
• Throughput: 500-600 MB/s
• Latency: 100-500 μs

NVMe SSD:
• IOPS: 500K-1M+
• Throughput: 3,000-7,000 MB/s
• Latency: 10-100 μs
```

### Process vs Thread

```
Process:
• Own address space
• Heavy context switch (1-10 μs)
• Strong isolation
• Use for: Different programs, security boundaries

Thread:
• Shared address space
• Light context switch (0.1-1 μs)
• Weak isolation
• Use for: Parallel tasks, shared state
```

### VM vs Container

```
Virtual Machine:
• Startup: Minutes
• Size: GBs
• Isolation: Strong
• Overhead: 5-15%
• Density: 10-50 per host

Container:
• Startup: Seconds
• Size: MBs
• Isolation: Moderate
• Overhead: 1-3%
• Density: 100-1000 per host
```

### AWS Instance Types

```
General Purpose:  t3, m6i, m7g
Compute Optimized: c6i, c7g
Memory Optimized:  r6i, r7g, x2iedn
Storage Optimized: i4i, d3
GPU:              p4d, g5
ARM (Graviton):   t4g, m7g, c7g (15% cheaper)
```

### Workload → Hardware Mapping

```
CPU-Bound:
• High core count (16-64+)
• High frequency (3.0+ GHz)
• Example: Video encoding, compilation

Memory-Bound:
• Large RAM (256 GB-2 TB)
• High bandwidth
• Example: In-memory databases, caching

I/O-Bound:
• Fast storage (NVMe SSD)
• High IOPS (500K+)
• Example: Databases, file servers

Network-Bound:
• High bandwidth (25-100 Gbps)
• Low latency NICs
• Example: Proxies, load balancers
```

### Scaling Strategies

```
Vertical (Scale Up):
• Increase resources per instance
• Pros: Simple, no code changes
• Cons: Limited, expensive, downtime

Horizontal (Scale Out):
• Add more instances
• Pros: Unlimited, HA, cost-effective
• Cons: Complex, requires stateless design
```

### Performance Optimization

```
CPU:
• Cache-friendly data structures
• SIMD instructions
• CPU affinity
• NUMA awareness

Memory:
• Huge pages (2 MB/1 GB)
• Memory pools
• Avoid false sharing
• Align to cache lines (64 bytes)

Storage:
• Caching (multiple layers)
• Read-ahead
• Write-back caching
• I/O scheduling

Network:
• Connection pooling
• Keep-alive
• Compression
• CDN
```

### Monitoring Commands

```bash
# CPU
lscpu
top / htop
mpstat -P ALL 1

# Memory
free -h
vmstat 1
cat /proc/meminfo

# Storage
iostat -x 1
iotop
df -h

# Network
iftop
netstat -s
ss -s

# Process
ps aux
pmap -x <pid>
strace -p <pid>

# Performance
perf stat -a sleep 5
perf top
```

### Capacity Planning Formula

```
Required Capacity = (Peak Load × (1 + Growth Rate)) / Target Utilization

Example:
• Peak Load: 10,000 req/s
• Growth Rate: 20% per year
• Target Utilization: 70%

Required = (10,000 × 1.2) / 0.7 = 17,143 req/s

If each instance handles 1,000 req/s:
Instances needed = 17,143 / 1,000 = 18 instances
```

### Cost Optimization

```
1. Right-sizing: Match instance to workload
2. Reserved Instances: 40-60% savings (1-3 year commit)
3. Spot Instances: 70-90% savings (interruptible)
4. Savings Plans: Up to 72% savings (flexible)
5. Graviton (ARM): 15-20% cheaper + better performance
```

### Common Pitfalls

```
❌ Over-provisioning: Wasting money on unused resources
❌ Under-provisioning: Poor performance, outages
❌ Ignoring NUMA: 2x slower memory access
❌ Cache-unfriendly code: 100x slower
❌ Synchronous I/O: Blocking threads
❌ No connection pooling: Overhead per request
❌ Large Docker images: Slow startup
❌ No monitoring: Flying blind
```

### Best Practices

```
✓ Profile before optimizing
✓ Monitor key metrics
✓ Plan for 20-30% headroom
✓ Design for horizontal scaling
✓ Use appropriate instance types
✓ Implement auto-scaling
✓ Cache aggressively
✓ Optimize critical path
✓ Test at scale
✓ Document decisions
```

### Interview Preparation Checklist

```
□ Understand CPU architecture (cores, cache, NUMA)
□ Know memory hierarchy and virtual memory
□ Explain storage types and performance characteristics
□ Describe process vs thread tradeoffs
□ Compare VMs vs containers
□ Design for different workload types
□ Calculate capacity requirements
□ Optimize for performance
□ Monitor and troubleshoot systems
□ Make cost-effective decisions
```

### Key Formulas

```
Throughput = IOPS × Block Size
Example: 100,000 IOPS × 4 KB = 390 MB/s

Amdahl's Law (Speedup):
Speedup = 1 / ((1 - P) + P/N)
P = Parallel portion
N = Number of processors

Little's Law:
L = λ × W
L = Average number in system
λ = Arrival rate
W = Average time in system

Cache Hit Rate:
Hit Rate = Cache Hits / Total Accesses
Example: 950 / 1000 = 95%
```

### Decision Trees

```
Choose Storage:
├─ Random access? → SSD
│  ├─ High IOPS? → NVMe
│  └─ Moderate? → SATA SSD
└─ Sequential? → HDD acceptable

Choose Scaling:
├─ Stateful? → Vertical first
├─ Stateless? → Horizontal
└─ Both? → Hybrid approach

Choose Isolation:
├─ Different OS? → VM
├─ Strong security? → VM
├─ Same kernel OK? → Container
└─ Microservices? → Container
```

---

## Quick Tips for Interviews

1. **Always ask about requirements first**
   - Scale (users, requests, data)
   - Latency requirements
   - Budget constraints

2. **Think about tradeoffs**
   - Performance vs Cost
   - Simplicity vs Scalability
   - Consistency vs Availability

3. **Use numbers**
   - Back-of-envelope calculations
   - Show you understand scale
   - Validate your design

4. **Draw diagrams**
   - Visual communication
   - Shows system thinking
   - Easier to discuss

5. **Consider operations**
   - Monitoring
   - Deployment
   - Disaster recovery

---

**Remember: There's no perfect solution, only appropriate tradeoffs for the given requirements!**
