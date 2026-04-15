# Scalable Sequencer — High-Level Design

## Key Questions Before Designing

### 1. Do we need numbers in sequence (i.e., strict ordering)?

It depends on the use case, and this is the single most important design decision:

- **Strictly ordered (1, 2, 3, 4…):** Required when the sequence itself carries meaning — like invoice numbers, ticket numbers, or database auto-increments. This is hard to scale because it introduces a serialization point. Every request must coordinate with a single source of truth.
- **Roughly ordered / k-sorted:** The IDs are time-sortable but not gap-free or perfectly sequential. Think Twitter Snowflake or ULIDs. You can tell which came first within a reasonable window, but there may be small out-of-order windows across nodes. This is much easier to scale.
- **Unordered unique:** UUIDs. No ordering guarantee at all. Trivially scalable but terrible for database index locality and human readability.

The architecture changes dramatically based on which guarantee you pick. Most real-world systems at scale choose "roughly ordered" because strict ordering creates a bottleneck.

### 2. How does it handle scale?

There are two fundamental strategies:

- **Pre-allocation / Range-based:** A central coordinator hands out ranges (e.g., Node A gets 1–10000, Node B gets 10001–20000). Each node then sequences locally without coordination. When a range is exhausted, the node fetches a new one. This is how Oracle sequences and Flickr's ticket servers work. The tradeoff: gaps are inevitable (if Node A crashes at 5000, IDs 5001–10000 are lost).

- **Composite ID / Bit-packing:** Embed the node identity into the ID itself. Snowflake does this — `[timestamp | node_id | local_counter]`. No central coordination needed for each ID, only at node registration time. Scales horizontally by adding nodes.

Both approaches avoid the "single writer" bottleneck. The key insight: you trade gap-free guarantees for horizontal scalability.

### 3. How does it ensure no duplicates?

Uniqueness is guaranteed structurally, not by checking a database:

- **Range-based:** Ranges never overlap. Node A owns 1–10000, Node B owns 10001–20000. Within a node, a simple atomic counter ensures uniqueness. Duplicates are impossible as long as the range allocator is correct (single writer to the range table, protected by a DB transaction or lock).

- **Composite ID:** The `node_id` bits guarantee cross-node uniqueness. The local counter (atomic increment) guarantees intra-node uniqueness. The timestamp adds an additional dimension. As long as no two nodes share the same `node_id`, duplicates are structurally impossible.

- **Clock skew (Snowflake-style):** If the clock goes backward, the node must either wait or refuse to issue IDs. This is a real operational concern — Snowflake handles it by throwing an exception until the clock catches up.

---

## High-Level Design: Scalable Sequencer

```
┌─────────────────────────────────────────────────────────┐
│                      Clients / Services                 │
│            (request next ID via gRPC / HTTP)            │
└──────────────┬──────────────────────┬───────────────────┘
               │                      │
         ┌─────▼─────┐         ┌──────▼────┐
         │ Sequencer  │         │ Sequencer │    ... N nodes
         │  Node A    │         │  Node B   │
         │            │         │           │
         │ node_id=1  │         │ node_id=2 │
         │ counter=0  │         │ counter=0 │
         │ range:     │         │ range:    │
         │ [1..10000] │         │ [10001..  │
         │            │         │  20000]   │
         └─────┬──────┘         └─────┬─────┘
               │  range exhausted?    │
               │  fetch new range     │
               ▼                      ▼
     ┌─────────────────────────────────────┐
     │         Range Allocator             │
     │   (single-writer, DB-backed)        │
     │                                     │
     │   table: sequences                  │
     │   ┌────────┬────────────────────┐   │
     │   │  name  │  next_range_start  │   │
     │   ├────────┼────────────────────┤   │
     │   │ orders │  20001             │   │
     │   │ users  │  50001             │   │
     │   └────────┴────────────────────┘   │
     │                                     │
     │   UPDATE sequences                  │
     │   SET next_range_start =            │
     │       next_range_start + range_size │
     │   WHERE name = ?                    │
     │   RETURNING old next_range_start    │
     │                                     │
     └─────────────────────────────────────┘

  ┌──────────────────────────────────────────┐
  │  Snowflake Mode (alternative path)       │
  │                                          │
  │  64-bit ID layout:                       │
  │  ┌──────────┬──────────┬────────────┐    │
  │  │ 41 bits  │ 10 bits  │  12 bits   │    │
  │  │timestamp │ node_id  │  counter   │    │
  │  └──────────┴──────────┴────────────┘    │
  │                                          │
  │  - 41 bits = ~69 years of milliseconds   │
  │  - 10 bits = 1024 nodes                  │
  │  - 12 bits = 4096 IDs/ms/node            │
  │  - Total: ~4M IDs/sec/node               │
  │                                          │
  │  Node registration via:                  │
  │  - ZooKeeper / etcd ephemeral nodes      │
  │  - or static config                      │
  └──────────────────────────────────────────┘
```

### Key design decisions

- Range size is tunable: small ranges (100) = fewer gaps but more DB round-trips. Large ranges (100K) = more gaps but nodes are independent longer.
- The range allocator is the only serialization point, and it's hit infrequently (once per range, not once per ID).
- For Snowflake mode, there's zero coordination at runtime — only at node startup for `node_id` assignment.
- Failover: if a node dies, its unused range is simply lost (acceptable gap). New node gets a fresh range.
- Persistence: the range allocator DB is the only state that matters. Sequencer nodes are stateless (ranges can be re-fetched on restart).

### Which mode to pick?

| Requirement | Range-based | Snowflake-style |
|---|---|---|
| Strictly sequential | Yes (within a node) | No |
| Gap-free | No | No |
| Human-readable | Yes (small numbers) | No (large 64-bit ints) |
| Throughput ceiling | Very high | Extremely high |
| Coordination needed | Per range refill | Only at boot |
| Clock dependency | None | Yes (critical) |

Neither approach gives you gap-free strict global ordering at scale. If you truly need that, you're looking at a single-writer bottleneck (one node, one counter, one lock) — which tops out around 10K–50K IDs/sec depending on your storage. That's fine for many use cases, but it won't scale horizontally.

The honest answer: pick the weakest ordering guarantee your business can tolerate, and you'll get the most scalable system.

---

## Snowflake and Integer Sequences

### Snowflake output is a 64-bit integer

The format `[timestamp | node_id | counter]` isn't a string or a composite key — it's bit-packing into a single `int64`. The pipes represent bit boundaries, not delimiters.

Here's how it works concretely:

```
Suppose:
  timestamp  = 1718000000000 (ms since custom epoch)
  node_id    = 5
  counter    = 42

Binary layout (64 bits total):
  [0][41 bits: timestamp][10 bits: node_id][12 bits: counter]

Bit shifting:
  id = (timestamp << 22) | (node_id << 12) | counter
  id = (1718000000000 << 22) | (5 << 12) | 42
  id = 7205759403792793642  ← this is a plain int64
```

So the output is just a number like `7205759403792793642`. It's a regular integer — you store it in a `BIGINT` column, pass it as a number in JSON, index it normally.

### The catch: it's not a *small* integer

Snowflake gives you integers, but they look like:

```
7205759403792793642
7205759403792793643
7205759403792797697   ← different node, same millisecond
```

Not like:

```
1
2
3
```

If you need small, human-friendly, sequential integers (like order #10042, ticket #3847), Snowflake is the wrong tool. It solves a different problem — globally unique, roughly time-sorted, no coordination.

### What to use for small sequential integers at scale

That's where the range-based approach is the right fit:

```
┌──────────────────────────────────────────────────┐
│              What you actually want               │
│                                                   │
│  order_id: 10001, 10002, 10003 ... (sequential)  │
│  ticket_id: 5001, 5002, 5003 ...  (sequential)   │
│                                                   │
│  ──► Range-based sequencer is your answer         │
│                                                   │
│  Node A gets range [10001..20000]                 │
│  Node B gets range [20001..30000]                 │
│                                                   │
│  Within each node: simple atomic increment        │
│  Across nodes: ranges never overlap = no dupes    │
│                                                   │
│  Tradeoff: IDs are sequential per-node,           │
│  but globally they interleave:                    │
│    10001, 20001, 10002, 20002, 10003 ...          │
│                                                   │
│  If you need strict global order:                 │
│    → single writer (bottleneck, ~10-50K IDs/sec)  │
│    → or accept interleaving                       │
└──────────────────────────────────────────────────┘
```

### Quick decision matrix

| You need | Use |
|---|---|
| Unique large int, time-sortable, massive scale | Snowflake (int64, but big numbers) |
| Small sequential int, some gaps OK | Range-based sequencer |
| Strictly ordered, gap-free, small int | Single-writer counter (limited scale) |
| Unique, no ordering needed | UUID (not an int at all) |

The bottom line: Snowflake does produce integers, just not the kind humans want to read on an invoice. If "integer sequence" means small, incrementing, human-readable numbers — go range-based and accept either interleaving across nodes or a single-writer bottleneck for strict ordering.
