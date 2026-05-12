# The Compute Layer (OS & Hardware)

## Overview

The Compute Layer forms the foundation of all software systems. As a software architect, understanding this layer is crucial for making informed decisions about performance, scalability, reliability, and cost optimization.

## Table of Contents

1. [Core Concepts](./01-core-concepts.md)
2. [CPU Architecture & Performance](./02-cpu-architecture.md)
3. [Memory Management](./03-memory-management.md)
4. [Storage Systems](./04-storage-systems.md)
5. [Operating Systems](./05-operating-systems.md)
6. [Virtualization & Containers](./06-virtualization-containers.md)
7. [Hardware Selection & Capacity Planning](./07-hardware-selection.md)
8. [Interview Questions & Answers](./08-interview-questions.md)

## Learning Path

```
┌─────────────────────────────────────────────────────────────┐
│                    COMPUTE LAYER MASTERY                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │   1. Hardware Fundamentals              │
        │   • CPU, Memory, Storage, I/O           │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │   2. Operating System Concepts          │
        │   • Process, Threads, Scheduling        │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │   3. Virtualization & Containers        │
        │   • VMs, Docker, Kubernetes             │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │   4. Performance & Optimization         │
        │   • Profiling, Tuning, Monitoring       │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │   5. Capacity Planning & Architecture   │
        │   • Sizing, Scaling, Cost Optimization  │
        └─────────────────────────────────────────┘
```

## Why This Matters for Software Architects

### Decision Impact Areas

1. **Performance**: Understanding CPU, memory, and I/O characteristics
2. **Scalability**: Knowing when to scale vertically vs horizontally
3. **Cost**: Optimizing resource utilization and cloud spending
4. **Reliability**: Designing for hardware failures and redundancy
5. **Security**: Understanding isolation, sandboxing, and attack surfaces

### Real-World Scenarios

- Choosing between EC2 instance types for different workloads
- Deciding when to use containers vs VMs vs bare metal
- Optimizing database performance through hardware understanding
- Designing systems that handle CPU-bound vs I/O-bound workloads
- Planning capacity for peak loads and growth

## Key Takeaways

> **"The best software architecture is one that understands and leverages the underlying hardware and OS capabilities effectively."**

- Hardware constraints shape software design
- OS abstractions enable portability but add overhead
- Modern architectures require understanding of distributed compute
- Performance optimization starts at the compute layer
- Cloud computing doesn't eliminate the need to understand hardware

## Next Steps

Start with [Core Concepts](./01-core-concepts.md) to build your foundation, then progress through each topic systematically.
