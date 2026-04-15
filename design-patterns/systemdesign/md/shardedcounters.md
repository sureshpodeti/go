# Sharded Counters — High-Level Design

## The Core Problem

A single atomic counter becomes a bottleneck under high write throughput. Every increment contends on the same memory location (in-memory) or the same row/key (in a database). Sharded counters solve this by splitting one logical counter into N independent shards, spreading write contention across them.

## How It Works (Conceptual Model)

```
Logical Counter: "request_count" = 1,000,000

Physical Reality:
  Shard 0: 250,100
  Shard 1: 249,800
  Shard 2: 250,050
  Shard 3: 250,050
  ─────────────────
  Sum    = 1,000,000

Write path:  hash(request_id) % N → pick shard → atomic increment
Read path:   sum(all shards) → return aggregate
```

The tradeoff is simple: writes get faster (less contention), reads get slower (aggregation cost).

---

## Question 1: Where Are Sharded Counters Stored?

It depends on the durability and consistency requirements. There are three tiers:

### Tier A — In-Memory (Single Node)

```
┌─────────────────────────────────┐
│           Application Node       │
│                                  │
│  ┌────────┐ ┌────────┐          │
│  │Shard 0 │ │Shard 1 │  ...     │
│  │ int64  │ │ int64  │          │
│  └────────┘ └────────┘          │
│                                  │
│  Access: atomic CAS / mutex      │
│  Latency: ~10-50ns per increment │
└─────────────────────────────────┘
```

Use when: rate limiters, in-process metrics, ephemeral counters where losing data on crash is acceptable.

Examples: Go's `sync/atomic`, Java's `LongAdder` (which is literally a sharded counter internally).

### Tier B — External In-Memory Store (Redis, Memcached)

```
┌──────────┐     ┌──────────────────────────┐
│  App Node │────▶│  Redis                    │
│           │     │  counter:shard:0 = 250100 │
│  App Node │────▶│  counter:shard:1 = 249800 │
│           │     │  counter:shard:2 = 250050 │
└──────────┘     └──────────────────────────┘
                  Latency: ~0.1-1ms per increment
```

Use when: shared counters across multiple app nodes, rate limiting across a fleet, session counts. Redis `INCRBY` is atomic and fast.

### Tier C — Durable Database (DynamoDB, Spanner, Cassandra)

```
┌──────────┐     ┌─────────────────────────────────┐
│  App Node │────▶│  DynamoDB                        │
│           │     │  PK=counter#req_count, SK=shard0 │
│  App Node │────▶│  PK=counter#req_count, SK=shard1 │
│           │     │  ...                              │
└──────────┘     └─────────────────────────────────┘
                  Latency: ~1-10ms per increment
                  Uses: UpdateItem ADD (atomic)
```

Use when: counters that must survive restarts, likes/votes, inventory counts. Google's Firestore documentation literally recommends sharded counters as a pattern for high-throughput writes.

### Summary Table

| Storage         | Latency     | Durability | Shared Across Nodes | Typical Use              |
|-----------------|-------------|------------|---------------------|--------------------------|
| In-process mem  | ~10-50ns    | None       | No                  | Rate limiters, metrics   |
| Redis           | ~0.1-1ms   | Optional   | Yes                 | Distributed rate limits  |
| Database        | ~1-10ms    | Full       | Yes                 | Likes, votes, inventory  |

---

## Question 2: In-Memory Counters and Node Crashes

If shards live only in process memory, a crash means total data loss for that counter. Here's how you handle it:

### Strategy 1 — Accept the Loss (Ephemeral Counters)

For rate limiters, the counter resets on crash. When the node comes back, it starts from zero. This is often fine because:
- Rate limit windows are short (1s, 1min)
- A brief period of over-allowing after restart is acceptable
- The alternative (durable writes per request) is too expensive

### Strategy 2 — Periodic Snapshotting

```
┌─────────────────────────────────────────┐
│  Application Node                        │
│                                          │
│  Shards [s0, s1, s2, s3]                │
│       │                                  │
│       │ every 5 seconds                  │
│       ▼                                  │
│  Snapshot → Redis/Disk                   │
│  "counter:snapshot = {sum, timestamp}"   │
│                                          │
│  On restart:                             │
│  Load snapshot → distribute across shards│
└─────────────────────────────────────────┘
```

You lose at most one snapshot interval worth of data. This is the same idea as Redis's RDB persistence or a WAL checkpoint.

### Strategy 3 — Write-Ahead Log (WAL)

Every N increments (or every T milliseconds), batch-write the deltas to a durable log. On recovery, replay the log. This gives you near-zero data loss but adds write amplification.

```
Increment → update shard in memory
         → append to WAL buffer
         → flush WAL buffer every 100ms to disk/Redis

Recovery: replay WAL from last checkpoint
```

### Strategy 4 — Hybrid: In-Memory + Async Drain to External Store

This is the most common production pattern:

```
┌──────────────┐         ┌─────────┐
│  App Node     │  async  │  Redis   │
│  local shards │────────▶│  INCRBY  │
│  (fast path)  │  batch  │          │
└──────────────┘         └─────────┘

- Increments hit local memory (fast, no network)
- Background goroutine drains accumulated deltas to Redis every 100ms
- On crash: lose at most 100ms of increments
- On read: local_sum + redis_sum (or just redis if approximate is fine)
```

---

## Question 3: Distributed Sharded Counters — How Reads Work

When a single node can't handle the write load, you distribute shards across multiple nodes. Now reads require aggregation.

### Architecture

```
                    ┌─────────────────┐
                    │   Read Request   │
                    │  "get counter X" │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   Coordinator    │
                    │  (any app node)  │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │  Node A   │  │  Node B   │  │  Node C   │
        │ Shards    │  │ Shards    │  │ Shards    │
        │ 0,1,2     │  │ 3,4,5     │  │ 6,7,8     │
        │ sum=30100 │  │ sum=29950 │  │ sum=30000 │
        └──────────┘  └──────────┘  └──────────┘
                             │
                    Total = 90,050
```

### Read Strategies

#### 1. Scatter-Gather (Strong Consistency)

```
coordinator:
    results = parallel_fetch(all_nodes, "give me your shard sums")
    total = sum(results)
    return total
```

- Latency = max(individual node latencies) + aggregation overhead
- Consistent but slow under high read load
- Tail latency problem: one slow node delays the entire read

#### 2. Cached Aggregation (Eventual Consistency)

```
┌──────────────┐
│  Cache Layer  │  ← stores aggregated total
│  TTL = 1-5s   │
└──────┬───────┘
       │ cache miss
       ▼
  scatter-gather across nodes
```

Most reads hit the cache. You trade freshness for speed. For something like "total likes on a post," being 2 seconds stale is perfectly fine.

#### 3. Hierarchical Aggregation

```
Level 0 (shards):     s0  s1  s2  s3  s4  s5  s6  s7
                        \  /    \  /    \  /    \  /
Level 1 (local sums):  sum01  sum23  sum45  sum67
                          \    /        \    /
Level 2 (region sums):  sumAB          sumCD
                            \          /
Level 3 (global):        TOTAL
```

Each level aggregates periodically (push-based). Reads at any level give progressively more stale but cheaper results. This is how large-scale metrics systems (like those at social media companies) work.

#### 4. CRDTs (Conflict-Free Replicated Data Types)

A G-Counter (grow-only counter) is a CRDT where each node maintains its own count. The merge function is `max()` per node, and the value is `sum(all nodes)`.

```
Node A: {A: 500, B: 0,   C: 0  }
Node B: {A: 0,   B: 300, C: 0  }
Node C: {A: 0,   B: 0,   C: 200}

Merge:  {A: 500, B: 300, C: 200} → total = 1000
```

No coordination needed for writes. Reads are eventually consistent. This is what Riak and similar systems use internally.

---

## Additional Critical Questions You Should Be Thinking About

### 4. How Many Shards?

Too few → contention remains. Too many → read aggregation is expensive, memory overhead grows.

Rule of thumb:
- In-memory: `num_shards = num_CPU_cores * 2-4` (aligns with Go's runtime scheduling, Java's LongAdder does similar)
- Database: `num_shards = expected_writes_per_second / writes_per_shard_before_contention`
- Start with 8-16, measure, adjust. Some systems auto-scale shard count based on observed contention.

### 5. Hot Shards

If your hash function is poor, some shards get more traffic. Use a good hash (xxhash, murmur3) and monitor per-shard write rates. If using consistent hashing for distributed shards, virtual nodes help spread load.

### 6. Counter Decrement and Negative Values

Sharded counters work cleanly for increment-only (monotonic) counters. Decrements introduce complexity:
- A shard can go negative locally even if the global count is positive
- You need to decide: allow negative shards (simpler) or redistribute (complex)
- For inventory/stock, this is where you might need a different pattern (reservation-based)

### 7. Consistency vs. Availability During Partitions

In a distributed setup, during a network partition:
- Writes can continue to local shards (AP choice) → counter may over-count
- Or you reject writes if you can't confirm global state (CP choice) → availability drops

For rate limiting, over-counting (briefly allowing more requests) is usually better than blocking all requests.

### 8. Garbage Collection of Expired Shards

For time-windowed counters (rate limiters), old shards need cleanup. Use:
- TTL on Redis keys
- Background sweep for in-memory maps
- DynamoDB TTL attribute

Without this, memory/storage grows unbounded.

---

## Putting It All Together — Production Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Load Balancer                         │
└──────────┬──────────────┬──────────────┬───────────────┘
           ▼              ▼              ▼
    ┌─────────────┐┌─────────────┐┌─────────────┐
    │  App Node 1  ││  App Node 2  ││  App Node 3  │
    │              ││              ││              │
    │ Local Shards ││ Local Shards ││ Local Shards │
    │ [s0..s7]     ││ [s0..s7]     ││ [s0..s7]     │
    │              ││              ││              │
    │ Drain every  ││ Drain every  ││ Drain every  │
    │ 100ms to ──────────────────────────────────────┐
    └─────────────┘└─────────────┘└─────────────┘   │
                                                      ▼
                                              ┌──────────────┐
                                              │    Redis       │
                                              │  (durable      │
                                              │   aggregation) │
                                              └──────┬───────┘
                                                     │
                                              ┌──────▼───────┐
                                              │  Read Path    │
                                              │  Cache + Sum  │
                                              └──────────────┘
```

Write path: local atomic increment (~50ns) → async batch drain to Redis (~100ms batches)
Read path: Redis cached aggregate (or scatter-gather for strong consistency)
Crash recovery: lose at most one drain interval, Redis has the durable state
Scaling: add more app nodes, each with local shards, all draining to same Redis counter

This gives you sub-microsecond write latency at the app layer, durability within 100ms, and reads that are as fresh as your cache TTL allows.
