# Hardware Selection & Capacity Planning

## Overview

Selecting the right hardware and planning capacity are critical architectural decisions that impact performance, cost, and scalability. This guide helps you make informed choices.

## Hardware Selection Framework

```
Hardware Selection Decision Tree:

Start
  │
  ▼
┌─────────────────────────────────────┐
│ What is your workload type?         │
└─────────────────────────────────────┘
  │
  ├─► CPU-Bound
  │   └─► High core count, high frequency
  │
  ├─► Memory-Bound
  │   └─► Large RAM, high bandwidth
  │
  ├─► I/O-Bound
  │   └─► Fast storage (NVMe SSD)
  │
  ├─► Network-Bound
  │   └─► High network bandwidth
  │
  └─► Mixed
      └─► Balanced configuration
```

## Workload Analysis

### 1. CPU-Bound Workloads

**Characteristics:**
- High CPU utilization (>80%)
- Low I/O wait time
- Examples: Video encoding, scientific computing, compilation

```
CPU-Bound Workload Profile:

┌─────────────────────────────────────┐
│ CPU Usage:  ████████████████ 95%   │
│ Memory:     ████░░░░░░░░░░░░ 30%   │
│ Disk I/O:   ██░░░░░░░░░░░░░░ 10%   │
│ Network:    █░░░░░░░░░░░░░░░  5%   │
└─────────────────────────────────────┘

Recommended Hardware:
┌─────────────────────────────────────┐
│ CPU:                                │
│ • High core count (16-64+ cores)    │
│ • High frequency (3.0+ GHz)         │
│ • Large cache (L3: 32-64 MB)        │
│ • AVX-512 support (if applicable)   │
│                                     │
│ Memory:                             │
│ • Moderate (32-128 GB)              │
│ • Standard speed (2666-3200 MHz)    │
│                                     │
│ Storage:                            │
│ • SATA SSD sufficient               │
│ • Moderate capacity                 │
│                                     │
│ Network:                            │
│ • 1-10 Gbps                         │
└─────────────────────────────────────┘

Example Processors:
• AMD EPYC 7763 (64 cores, 2.45 GHz)
• Intel Xeon Platinum 8380 (40 cores, 2.3 GHz)
• AMD Threadripper PRO 5995WX (64 cores, 2.7 GHz)
```

### 2. Memory-Bound Workloads

**Characteristics:**
- High memory usage
- Frequent cache misses
- Examples: In-memory databases, big data analytics, caching

```
Memory-Bound Workload Profile:

┌─────────────────────────────────────┐
│ CPU Usage:  ████░░░░░░░░░░░░ 30%   │
│ Memory:     ████████████████ 95%   │
│ Disk I/O:   ██░░░░░░░░░░░░░░ 10%   │
│ Network:    ███░░░░░░░░░░░░░ 20%   │
└─────────────────────────────────────┘

Recommended Hardware:
┌─────────────────────────────────────┐
│ CPU:                                │
│ • Moderate core count (8-32 cores)  │
│ • Large cache (L3: 64-256 MB)       │
│ • High memory bandwidth             │
│                                     │
│ Memory:                             │
│ • Very large (256 GB - 2 TB+)       │
│ • High speed (3200+ MHz)            │
│ • Many channels (8-12 channels)     │
│ • ECC recommended                   │
│                                     │
│ Storage:                            │
│ • NVMe SSD for swap (if needed)     │
│ • Large capacity                    │
│                                     │
│ Network:                            │
│ • 10-25 Gbps                        │
└─────────────────────────────────────┘

Memory Configuration Example:
┌─────────────────────────────────────┐
│ 2-Socket Server                     │
│                                     │
│ Socket 0: 512 GB (8x 64GB DIMMs)    │
│ Socket 1: 512 GB (8x 64GB DIMMs)    │
│                                     │
│ Total: 1 TB RAM                     │
│ Bandwidth: ~200 GB/s per socket     │
└─────────────────────────────────────┘

Example Systems:
• Dell PowerEdge R750 (up to 8 TB RAM)
• HPE ProLiant DL380 Gen10 Plus (up to 8 TB RAM)
• AWS r6i.32xlarge (1 TB RAM)
```

### 3. I/O-Bound Workloads

**Characteristics:**
- High disk I/O
- CPU waiting for I/O
- Examples: Databases, file servers, log processing

```
I/O-Bound Workload Profile:

┌─────────────────────────────────────┐
│ CPU Usage:  ███░░░░░░░░░░░░░ 25%   │
│ Memory:     ████░░░░░░░░░░░░ 30%   │
│ Disk I/O:   ████████████████ 95%   │
│ Network:    ████░░░░░░░░░░░░ 30%   │
└─────────────────────────────────────┘

Recommended Hardware:
┌─────────────────────────────────────┐
│ CPU:                                │
│ • Moderate core count (8-24 cores)  │
│ • Standard frequency                │
│                                     │
│ Memory:                             │
│ • Large for caching (64-256 GB)     │
│                                     │
│ Storage:                            │
│ • NVMe SSD (multiple drives)        │
│ • High IOPS (500K+)                 │
│ • RAID 10 for performance           │
│ • Consider Intel Optane             │
│                                     │
│ Network:                            │
│ • 10-25 Gbps                        │
└─────────────────────────────────────┘

Storage Configuration Example:
┌─────────────────────────────────────┐
│ Database Server                     │
│                                     │
│ OS: 2x 480GB SATA SSD (RAID 1)      │
│                                     │
│ Data: 4x 3.84TB NVMe SSD (RAID 10)  │
│ • Total: 7.68 TB usable             │
│ • IOPS: 2M+ random read             │
│ • Throughput: 20+ GB/s              │
│                                     │
│ Logs: 2x 960GB NVMe SSD (RAID 1)    │
└─────────────────────────────────────┘

Example Storage:
• Samsung PM1733 (NVMe, 6.4 GB/s)
• Intel Optane P5800X (NVMe, 1.5M IOPS)
• Micron 7450 MAX (NVMe, 1.5M IOPS)
```

### 4. Network-Bound Workloads

**Characteristics:**
- High network traffic
- Low CPU/disk usage
- Examples: Proxies, load balancers, CDN edge nodes

```
Network-Bound Workload Profile:

┌─────────────────────────────────────┐
│ CPU Usage:  ███░░░░░░░░░░░░░ 20%   │
│ Memory:     ██░░░░░░░░░░░░░░ 15%   │
│ Disk I/O:   █░░░░░░░░░░░░░░░  5%   │
│ Network:    ████████████████ 95%   │
└─────────────────────────────────────┘

Recommended Hardware:
┌─────────────────────────────────────┐
│ CPU:                                │
│ • Moderate core count (8-16 cores)  │
│ • Good single-thread performance    │
│                                     │
│ Memory:                             │
│ • Moderate (32-64 GB)               │
│                                     │
│ Storage:                            │
│ • SATA SSD sufficient               │
│                                     │
│ Network:                            │
│ • Multiple 25-100 Gbps NICs         │
│ • SR-IOV support                    │
│ • DPDK support                      │
│ • Low latency (<1 μs)               │
└─────────────────────────────────────┘

Network Configuration Example:
┌─────────────────────────────────────┐
│ Load Balancer / Proxy               │
│                                     │
│ 2x 100 Gbps NICs                    │
│ • Bonded for redundancy             │
│ • SR-IOV for VM/container isolation │
│                                     │
│ Network Offloading:                 │
│ • TCP/UDP checksum offload          │
│ • Large receive offload (LRO)       │
│ • Generic segmentation offload (GSO)│
└─────────────────────────────────────┘

Example NICs:
• Mellanox ConnectX-6 (100 Gbps)
• Intel E810 (100 Gbps)
• Broadcom NetXtreme-E (100 Gbps)
```

## Cloud Instance Selection

### AWS EC2 Instance Families

```
AWS EC2 Instance Types:

General Purpose (T, M):
┌─────────────────────────────────────┐
│ t3.medium                           │
│ • 2 vCPU, 4 GB RAM                  │
│ • Burstable CPU                     │
│ • $0.0416/hour                      │
│ Use: Web servers, dev/test          │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│ m6i.2xlarge                         │
│ • 8 vCPU, 32 GB RAM                 │
│ • Balanced compute/memory           │
│ • $0.384/hour                       │
│ Use: Application servers            │
└─────────────────────────────────────┘

Compute Optimized (C):
┌─────────────────────────────────────┐
│ c6i.4xlarge                         │
│ • 16 vCPU, 32 GB RAM                │
│ • High CPU performance              │
│ • $0.68/hour                        │
│ Use: Batch processing, HPC          │
└─────────────────────────────────────┘

Memory Optimized (R, X):
┌─────────────────────────────────────┐
│ r6i.4xlarge                         │
│ • 16 vCPU, 128 GB RAM               │
│ • High memory-to-CPU ratio          │
│ • $1.008/hour                       │
│ Use: Databases, caching             │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│ x2iedn.32xlarge                     │
│ • 128 vCPU, 4 TB RAM                │
│ • Extreme memory                    │
│ • $26.688/hour                      │
│ Use: SAP HANA, in-memory DBs        │
└─────────────────────────────────────┘

Storage Optimized (I, D):
┌─────────────────────────────────────┐
│ i4i.4xlarge                         │
│ • 16 vCPU, 128 GB RAM               │
│ • 3.75 TB NVMe SSD                  │
│ • $1.6416/hour                      │
│ Use: NoSQL databases, data warehouses│
└─────────────────────────────────────┘

Accelerated Computing (P, G):
┌─────────────────────────────────────┐
│ p4d.24xlarge                        │
│ • 96 vCPU, 1.1 TB RAM               │
│ • 8x NVIDIA A100 GPUs               │
│ • $32.77/hour                       │
│ Use: ML training, HPC               │
└─────────────────────────────────────┘

ARM-based (Graviton):
┌─────────────────────────────────────┐
│ m7g.2xlarge                         │
│ • 8 vCPU, 32 GB RAM                 │
│ • AWS Graviton3                     │
│ • $0.3264/hour (15% cheaper)        │
│ Use: General purpose, cost-optimized│
└─────────────────────────────────────┘
```

### Instance Selection Decision Matrix

```
Workload → Instance Type Mapping:

┌──────────────────┬─────────────────────┐
│ Workload         │ Recommended Type    │
├──────────────────┼─────────────────────┤
│ Web Server       │ t3, t4g (burstable) │
│ App Server       │ m6i, m7g            │
│ API Gateway      │ c6i, c7g            │
│ Database (OLTP)  │ r6i + i4i (storage) │
│ Database (OLAP)  │ r6i, x2iedn         │
│ Cache (Redis)    │ r6i, r7g            │
│ Queue (Kafka)    │ i4i, d3             │
│ Batch Processing │ c6i, c7g            │
│ ML Training      │ p4d, p5             │
│ ML Inference     │ g5, inf2            │
│ Video Encoding   │ c6i + GPU           │
│ File Server      │ d3, i4i             │
└──────────────────┴─────────────────────┘

Cost vs Performance:
┌─────────────────────────────────────┐
│                                     │
│ High │ p4d                          │
│  P   │      x2iedn                  │
│  e   │           i4i                │
│  r   │                r6i           │
│  f   │                    m6i       │
│  o   │                        c6i   │
│  r   │                          t3  │
│  m   │                              │
│ Low  └─────────────────────────────►│
│      Low                  High Cost │
└─────────────────────────────────────┘
```

## Capacity Planning

### Capacity Planning Process

```
Capacity Planning Workflow:

1. Baseline Measurement
   ┌─────────────────────────────────┐
   │ • Current resource usage        │
   │ • Peak usage patterns           │
   │ • Growth trends                 │
   └─────────────────────────────────┘
            │
            ▼
2. Workload Forecasting
   ┌─────────────────────────────────┐
   │ • Expected growth rate          │
   │ • Seasonal patterns             │
   │ • New features impact           │
   └─────────────────────────────────┘
            │
            ▼
3. Resource Modeling
   ┌─────────────────────────────────┐
   │ • CPU requirements              │
   │ • Memory requirements           │
   │ • Storage requirements          │
   │ • Network requirements          │
   └─────────────────────────────────┘
            │
            ▼
4. Headroom Planning
   ┌─────────────────────────────────┐
   │ • Safety margin (20-30%)        │
   │ • Burst capacity                │
   │ • Failure scenarios             │
   └─────────────────────────────────┘
            │
            ▼
5. Cost Optimization
   ┌─────────────────────────────────┐
   │ • Right-sizing                  │
   │ • Reserved instances            │
   │ • Spot instances                │
   └─────────────────────────────────┘
```

### Capacity Metrics

```
Key Capacity Metrics:

CPU:
┌─────────────────────────────────────┐
│ Current: 65% average, 85% peak      │
│ Target:  70% average, 90% peak      │
│ Headroom: 5% average, 5% peak       │
│                                     │
│ Action: Add 10% capacity            │
└─────────────────────────────────────┘

Memory:
┌─────────────────────────────────────┐
│ Current: 128 GB, 75% used           │
│ Growth:  10% per quarter            │
│ Target:  80% utilization            │
│                                     │
│ Forecast: Need 192 GB in 6 months   │
└─────────────────────────────────────┘

Storage:
┌─────────────────────────────────────┐
│ Current: 10 TB, 8 TB used (80%)     │
│ Growth:  500 GB per month           │
│ Target:  75% utilization            │
│                                     │
│ Forecast: Need 15 TB in 6 months    │
└─────────────────────────────────────┘

Network:
┌─────────────────────────────────────┐
│ Current: 10 Gbps, 6 Gbps peak       │
│ Growth:  15% per quarter            │
│ Target:  80% utilization            │
│                                     │
│ Forecast: Need 25 Gbps in 1 year    │
└─────────────────────────────────────┘
```

### Scaling Strategies

```
Vertical vs Horizontal Scaling:

Vertical Scaling (Scale Up):
┌──────────┐         ┌──────────┐
│ 4 cores  │   →     │ 8 cores  │
│ 16 GB    │         │ 32 GB    │
└──────────┘         └──────────┘

Pros:
• Simple (no code changes)
• No distributed system complexity
• Better for stateful applications

Cons:
• Limited by hardware
• Single point of failure
• Downtime for upgrades
• Expensive at scale

Horizontal Scaling (Scale Out):
┌──────────┐         ┌──────────┬──────────┐
│ 4 cores  │   →     │ 4 cores  │ 4 cores  │
│ 16 GB    │         │ 16 GB    │ 16 GB    │
└──────────┘         └──────────┴──────────┘

Pros:
• No hardware limits
• High availability
• Cost-effective
• Elastic scaling

Cons:
• Complex (load balancing, state management)
• Network overhead
• Eventual consistency
• Requires stateless design

Hybrid Approach:
┌─────────────────────────────────────┐
│ Scale vertically first              │
│ • Up to cost-effective limit        │
│ • Simpler operations                │
│                                     │
│ Then scale horizontally             │
│ • Beyond single-machine limits      │
│ • For high availability             │
└─────────────────────────────────────┘
```

### Auto-Scaling

```
Auto-Scaling Configuration:

Metric-Based Scaling:
┌─────────────────────────────────────┐
│ Scale Up Trigger:                   │
│ • CPU > 70% for 5 minutes           │
│ • Memory > 80% for 5 minutes        │
│ • Request queue > 100               │
│                                     │
│ Scale Down Trigger:                 │
│ • CPU < 30% for 15 minutes          │
│ • Memory < 40% for 15 minutes       │
│ • Request queue < 10                │
│                                     │
│ Limits:                             │
│ • Min instances: 2                  │
│ • Max instances: 20                 │
│ • Cooldown: 5 minutes               │
└─────────────────────────────────────┘

Schedule-Based Scaling:
┌─────────────────────────────────────┐
│ Monday-Friday:                      │
│ • 08:00-18:00: 10 instances         │
│ • 18:00-08:00: 3 instances          │
│                                     │
│ Saturday-Sunday:                    │
│ • All day: 2 instances              │
│                                     │
│ Special Events:                     │
│ • Black Friday: 50 instances        │
└─────────────────────────────────────┘

Predictive Scaling:
┌─────────────────────────────────────┐
│ ML-based forecasting:               │
│ • Historical patterns               │
│ • Seasonal trends                   │
│ • Pre-scale before load             │
│                                     │
│ Example:                            │
│ • Predict 2x load at 9 AM           │
│ • Scale up at 8:45 AM               │
│ • Avoid cold start delays           │
└─────────────────────────────────────┘
```

## Cost Optimization

### Cost Optimization Strategies

```
Cost Optimization Techniques:

1. Right-Sizing:
┌─────────────────────────────────────┐
│ Before: m5.4xlarge                  │
│ • 16 vCPU, 64 GB RAM                │
│ • Actual usage: 4 vCPU, 16 GB       │
│ • Cost: $0.768/hour                 │
│                                     │
│ After: m5.xlarge                    │
│ • 4 vCPU, 16 GB RAM                 │
│ • Cost: $0.192/hour                 │
│                                     │
│ Savings: 75% ($5,040/year)          │
└─────────────────────────────────────┘

2. Reserved Instances:
┌─────────────────────────────────────┐
│ On-Demand: $0.768/hour              │
│ 1-Year Reserved: $0.461/hour (40%)  │
│ 3-Year Reserved: $0.307/hour (60%)  │
│                                     │
│ For stable workloads                │
└─────────────────────────────────────┘

3. Spot Instances:
┌─────────────────────────────────────┐
│ On-Demand: $0.768/hour              │
│ Spot: $0.230/hour (70% savings)     │
│                                     │
│ For fault-tolerant workloads:       │
│ • Batch processing                  │
│ • Data analysis                     │
│ • CI/CD                             │
└─────────────────────────────────────┘

4. Savings Plans:
┌─────────────────────────────────────┐
│ Commit to $X/hour for 1-3 years     │
│ • Flexible across instance types    │
│ • Flexible across regions           │
│ • Up to 72% savings                 │
└─────────────────────────────────────┘

5. Graviton (ARM):
┌─────────────────────────────────────┐
│ m6i.2xlarge: $0.384/hour            │
│ m7g.2xlarge: $0.3264/hour           │
│                                     │
│ Savings: 15% + better performance   │
└─────────────────────────────────────┘
```

### Cost Monitoring

```
Cost Monitoring Dashboard:

┌─────────────────────────────────────┐
│ Monthly Cost: $15,234               │
│ Trend: ▲ 12% vs last month          │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Cost by Service:                    │
│ • Compute: $8,500 (56%)             │
│ • Storage: $3,200 (21%)             │
│ • Network: $2,100 (14%)             │
│ • Other: $1,434 (9%)                │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Top 5 Resources:                    │
│ 1. prod-db: $2,400/month            │
│ 2. prod-app: $1,800/month           │
│ 3. prod-cache: $1,200/month         │
│ 4. dev-cluster: $900/month          │
│ 5. test-env: $600/month             │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Optimization Opportunities:         │
│ • 15 underutilized instances        │
│ • Potential savings: $2,100/month   │
│ • 8 instances eligible for RI       │
│ • Potential savings: $1,500/month   │
└─────────────────────────────────────┘
```

## Performance Benchmarking

### Benchmarking Tools

```
Common Benchmarking Tools:

CPU:
• sysbench --test=cpu
• stress-ng
• SPEC CPU 2017

Memory:
• sysbench --test=memory
• STREAM benchmark
• mbw (memory bandwidth)

Storage:
• fio (Flexible I/O Tester)
• ioping
• dd

Network:
• iperf3
• netperf
• qperf

Application:
• Apache Bench (ab)
• wrk
• JMeter
```

### Benchmark Example

```
Storage Benchmark (fio):

Test: Random 4K Read
┌─────────────────────────────────────┐
│ SATA SSD:                           │
│ • IOPS: 95,000                      │
│ • Latency: 105 μs                   │
│ • Throughput: 371 MB/s              │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ NVMe SSD:                           │
│ • IOPS: 650,000                     │
│ • Latency: 15 μs                    │
│ • Throughput: 2,539 MB/s            │
└─────────────────────────────────────┘

Performance Gain: 6.8x IOPS, 7x faster
```

## Summary

Key takeaways for architects:

1. **Match hardware to workload**
   - Analyze workload characteristics
   - Choose appropriate instance types
   - Don't over-provision

2. **Plan for growth**
   - Monitor trends
   - Forecast capacity needs
   - Build in headroom (20-30%)

3. **Optimize costs**
   - Right-size instances
   - Use reserved/spot instances
   - Consider ARM (Graviton)

4. **Design for scaling**
   - Horizontal scaling preferred
   - Auto-scaling for elasticity
   - Stateless when possible

5. **Benchmark and validate**
   - Test before production
   - Monitor performance
   - Iterate and optimize

## Next Steps

Continue to [Interview Questions & Answers](./08-interview-questions.md) to test your knowledge and prepare for architect interviews.
