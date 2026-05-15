# Data Architecture, Real-Time Systems & Consensus — 15 Questions
## Software Architect Interview Preparation

---

## Overview

This document covers the **data and real-time** side of distributed systems architecture — the topics that come up when designing data-intensive applications, real-time features, and systems that require coordination.

**Topics Covered:**
- Data Lakes, Warehouses, and Lakehouses
- ETL/ELT Pipelines at Scale
- Stream Processing (Flink, Spark Streaming)
- Schema Evolution and Migrations
- Search Architecture (Elasticsearch)
- Real-Time Features (WebSockets, SSE, Long Polling)
- Real-Time Leaderboards and Presence Systems
- Collaborative Editing (CRDTs)
- Leader Election and Consensus (Raft/Paxos)
- Distributed Locks
- Time Series Databases
- Data Partitioning Strategies
- Change Data Capture (CDC)
- Vector Databases and Semantic Search
- Event Sourcing

---

## Table of Contents

1. [Q1 — Data Lake vs Data Warehouse vs Lakehouse](#d1)
2. [Q2 — ETL/ELT Pipelines at Scale](#d2)
3. [Q3 — Stream Processing with Apache Flink](#d3)
4. [Q4 — Schema Evolution Without Downtime](#d4)
5. [Q5 — Search Architecture with Elasticsearch](#d5)
6. [Q6 — WebSockets vs SSE vs Long Polling](#d6)
7. [Q7 — Real-Time Leaderboard Design](#d7)
8. [Q8 — Presence System (Online/Offline Status)](#d8)
9. [Q9 — Collaborative Editing and CRDTs](#d9)
10. [Q10 — Leader Election and Raft Consensus](#d10)
11. [Q11 — Distributed Locks in Practice](#d11)
12. [Q12 — Time Series Databases](#d12)
13. [Q13 — Change Data Capture (CDC)](#d13)
14. [Q14 — Vector Databases and Semantic Search](#d14)
15. [Q15 — Event Sourcing Pattern](#d15)

---

## Q1: Data Lake vs Data Warehouse vs Lakehouse {#d1}

**Situation:**
Your company has data in PostgreSQL (transactions), Kafka (events), S3 (logs), and Salesforce (CRM). The data science team wants to run ML models. The analytics team wants dashboards. The finance team wants monthly reports. Each team is building their own data pipelines, creating data silos and inconsistencies. The CEO sees different revenue numbers from different teams. You need a unified data architecture.

**The Three Paradigms:**

---

### Data Warehouse (Structured, Fast Queries)

A data warehouse stores **structured, processed data** optimized for analytical queries (OLAP — Online Analytical Processing).

```
Characteristics:
  - Schema-on-write: data must conform to schema before loading
  - Structured data only (tables, columns, types)
  - Highly optimized for SQL queries (columnar storage)
  - Expensive storage (SSD, in-memory)
  - Fast query performance (seconds to minutes)

Examples: Snowflake, BigQuery, Redshift, Azure Synapse

Use when:
  - Business intelligence and dashboards
  - Known query patterns
  - Data quality is critical
  - Finance, sales, operations reporting

ETL process:
  Source → Extract → Transform (clean, validate) → Load → Warehouse
  Data is transformed BEFORE loading (schema-on-write)
```

---

### Data Lake (Raw, Flexible, Cheap)

A data lake stores **raw data in any format** — structured, semi-structured, unstructured.

```
Characteristics:
  - Schema-on-read: apply schema when querying, not when storing
  - Any format: JSON, CSV, Parquet, images, video, logs
  - Cheap storage (S3, HDFS, Azure Data Lake)
  - Slow queries (must scan raw files)
  - Flexible: store everything, decide later what to do with it

Examples: S3 + Athena, HDFS + Hive, Azure Data Lake

Use when:
  - ML training data (needs raw, unprocessed data)
  - Unknown future use cases ("store everything")
  - Log archival
  - Data exploration

ELT process:
  Source → Extract → Load (raw) → Transform (when querying)
  Data is loaded raw, transformed at query time (schema-on-read)
```

---

### Data Lakehouse (Best of Both)

A lakehouse combines the cheap storage of a data lake with the query performance and ACID transactions of a data warehouse.

```
Key technologies:
  - Delta Lake (Databricks): ACID transactions on S3/HDFS
  - Apache Iceberg: Table format with time travel, schema evolution
  - Apache Hudi: Upserts and incremental processing on data lakes

Characteristics:
  - Cheap storage (S3) + fast queries (columnar format + caching)
  - ACID transactions (no more "data swamp")
  - Schema enforcement + schema evolution
  - Time travel (query data as of any point in time)
  - Unified: one copy of data for both ML and BI

Architecture:
  Raw Zone (S3):     Raw data as-is (JSON, CSV, logs)
  Bronze Layer:      Ingested, minimal cleaning (Parquet format)
  Silver Layer:      Cleaned, validated, joined (Delta Lake tables)
  Gold Layer:        Aggregated, business-ready (for dashboards/reports)
```

---

### Unified Architecture

```
                    ┌──────────────────────────────────────────┐
                    │           Data Sources                    │
                    │  PostgreSQL | Kafka | S3 | Salesforce     │
                    └──────────────────┬───────────────────────┘
                                       │ CDC / Kafka / API
                                       ▼
                    ┌──────────────────────────────────────────┐
                    │         Ingestion Layer                   │
                    │  Apache Kafka + Kafka Connect             │
                    │  (streams all changes in real-time)       │
                    └──────────────────┬───────────────────────┘
                                       │
                    ┌──────────────────▼───────────────────────┐
                    │         Data Lakehouse (S3 + Delta Lake)  │
                    │  Bronze: raw ingested data                │
                    │  Silver: cleaned, joined, validated       │
                    │  Gold: aggregated, business metrics       │
                    └──────┬───────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
    ┌─────▼──────┐  ┌──────▼──────┐  ┌─────▼──────┐
    │  Snowflake  │  │  Spark/     │  │  Feature   │
    │  (BI/       │  │  Databricks │  │  Store     │
    │  Dashboards)│  │  (ML)       │  │  (ML Svc)  │
    └─────────────┘  └─────────────┘  └────────────┘
```

---

### Key Takeaways

1. **Data warehouse for known queries, data lake for unknown** — if you know what questions you'll ask, use a warehouse. If you're exploring, use a lake.
2. **Lakehouse is the modern default** — Delta Lake/Iceberg gives you cheap storage + fast queries + ACID. Most new architectures start here.
3. **Bronze/Silver/Gold layers enforce data quality** — raw data is never modified. Each layer adds quality. Gold layer is the single source of truth.
4. **One copy of data, multiple consumers** — ML, BI, and operations all read from the same lakehouse. No more data silos.
5. **Schema-on-read enables flexibility** — store raw data now, decide schema later. Useful when requirements are unclear.

**Interview Follow-up Questions:**
- "What is the difference between OLTP and OLAP?"
- "What is columnar storage and why is it faster for analytical queries?"
- "What is Delta Lake and what problems does it solve?"
- "How do you handle slowly changing dimensions (SCD) in a data warehouse?"

---

## Q2: ETL/ELT Pipelines at Scale {#d2}

**Situation:**
Your data team runs nightly ETL jobs that take 8 hours to complete. By the time dashboards are updated, the data is 12-20 hours old. Business users are making decisions on stale data. Additionally, the ETL jobs frequently fail due to schema changes in source systems, requiring manual intervention. You need near-real-time data with automated schema handling.

**ETL vs ELT:**

```
ETL (Extract → Transform → Load):
  - Transform data BEFORE loading into warehouse
  - Transformation happens in a separate compute layer
  - Data warehouse receives clean, structured data
  - Traditional approach (Informatica, SSIS)
  - Good when: warehouse compute is expensive, data quality is critical

ELT (Extract → Load → Transform):
  - Load raw data FIRST, transform inside the warehouse
  - Leverage warehouse's massive compute (BigQuery, Snowflake)
  - Raw data preserved (can re-transform if logic changes)
  - Modern approach (dbt + Snowflake/BigQuery)
  - Good when: warehouse compute is cheap, flexibility needed
```

---

### Modern ELT Stack

```
Layer 1 — Ingestion (Fivetran, Airbyte, Kafka Connect):
  - Connects to 200+ sources (PostgreSQL, Salesforce, Stripe, etc.)
  - Handles schema changes automatically
  - Incremental sync (only new/changed records)
  - Loads raw data to warehouse/lake

Layer 2 — Transformation (dbt — data build tool):
  - SQL-based transformations (version controlled in Git)
  - Dependency graph (knows which models depend on which)
  - Testing (assert column is not null, assert unique)
  - Documentation (auto-generated data catalog)
  - Incremental models (only process new data)

Layer 3 — Orchestration (Apache Airflow, Prefect, Dagster):
  - Schedule and monitor pipeline runs
  - Handle dependencies between jobs
  - Retry failed tasks
  - Alert on failures

Layer 4 — Serving (Snowflake, BigQuery, Redshift):
  - Optimized for analytical queries
  - Materialized views for fast dashboard queries
  - Row-level security for data access control
```

---

### dbt Model Example

```sql
-- models/silver/orders_enriched.sql
-- dbt model: joins orders with users and products

{{ config(
    materialized='incremental',  -- Only process new records
    unique_key='order_id',
    on_schema_change='sync_all_columns'  -- Auto-handle schema changes
) }}

SELECT
    o.order_id,
    o.created_at,
    o.total_amount,
    o.status,
    u.user_id,
    u.email,
    u.country,
    u.tier,
    COUNT(oi.product_id) as item_count,
    SUM(oi.quantity) as total_items
FROM {{ ref('bronze_orders') }} o
JOIN {{ ref('bronze_users') }} u ON o.user_id = u.user_id
JOIN {{ ref('bronze_order_items') }} oi ON o.order_id = oi.order_id

{% if is_incremental() %}
    -- Only process orders created since last run
    WHERE o.created_at > (SELECT MAX(created_at) FROM {{ this }})
{% endif %}

GROUP BY 1, 2, 3, 4, 5, 6, 7, 8

-- dbt test: orders_enriched.yml
-- - name: order_id
--   tests:
--     - unique
--     - not_null
-- - name: total_amount
--   tests:
--     - not_null
--     - dbt_utils.accepted_range:
--         min_value: 0
```

---

### Near-Real-Time with Kafka + Flink

For data freshness < 5 minutes, batch ELT is not enough. Use streaming:

```
PostgreSQL → Debezium (CDC) → Kafka → Flink → Data Warehouse
                                              → Real-time dashboard

Latency: 10-30 seconds (vs 8 hours for batch)
```

---

### Key Takeaways

1. **ELT over ETL for modern cloud warehouses** — BigQuery and Snowflake have massive compute. Transform inside the warehouse using dbt.
2. **dbt is the standard for SQL transformations** — version control, testing, documentation, incremental processing. Every data team should use it.
3. **Incremental models are essential at scale** — processing only new data reduces runtime from hours to minutes.
4. **Schema changes are inevitable** — use tools that handle them automatically (Fivetran, Airbyte) rather than brittle custom code.
5. **For < 5 minute freshness, use streaming** — batch ELT cannot achieve near-real-time. Kafka + Flink is the standard streaming stack.

**Interview Follow-up Questions:**
- "What is the difference between a fact table and a dimension table?"
- "What is dbt and how does it differ from traditional ETL tools?"
- "How do you handle late-arriving data in a streaming pipeline?"
- "What is data lineage and why does it matter?"

---

## Q3: Stream Processing with Apache Flink {#d3}

**Situation:**
Your fraud detection system currently runs as a batch job every 15 minutes. It analyzes transaction patterns and flags suspicious activity. By the time fraud is detected, the customer has already made 10 fraudulent transactions. You need real-time fraud detection — flag suspicious transactions within 1 second of occurrence.

**Batch vs Stream Processing:**

```
Batch Processing (current):
  - Process data in chunks (every 15 minutes)
  - High latency (15 min to detect fraud)
  - Simple to implement
  - Good for: reports, ETL, ML training

Stream Processing (needed):
  - Process each event as it arrives
  - Low latency (< 1 second)
  - More complex (state management, fault tolerance)
  - Good for: fraud detection, real-time recommendations, monitoring
```

---

### Apache Flink Core Concepts

**1. Streams and Transformations**
```java
// Flink DataStream API
DataStream<Transaction> transactions = env
    .addSource(new FlinkKafkaConsumer<>("transactions", schema, props));

DataStream<FraudAlert> alerts = transactions
    .keyBy(Transaction::getUserId)          // Group by user
    .window(TumblingEventTimeWindows.of(Time.minutes(5)))  // 5-min window
    .aggregate(new FraudDetectionAggregator())  // Detect patterns
    .filter(alert -> alert.getRiskScore() > 0.8);  // High risk only

alerts.addSink(new FlinkKafkaProducer<>("fraud-alerts", schema, props));
```

**2. Stateful Processing (Key to Fraud Detection)**
```java
// Stateful function: tracks user's transaction history
public class FraudDetector extends KeyedProcessFunction<String, Transaction, FraudAlert> {
    
    // State: last 10 transactions per user (persisted in RocksDB)
    private ListState<Transaction> recentTransactions;
    
    @Override
    public void processElement(Transaction tx, Context ctx, Collector<FraudAlert> out) {
        List<Transaction> history = Lists.newArrayList(recentTransactions.get());
        
        // Fraud rule 1: 3+ transactions in different countries in 10 minutes
        long recentForeignTx = history.stream()
            .filter(t -> !t.getCountry().equals(tx.getCountry()))
            .filter(t -> tx.getTimestamp() - t.getTimestamp() < 600_000)  // 10 min
            .count();
        
        if (recentForeignTx >= 2) {
            out.collect(new FraudAlert(tx.getUserId(), "MULTI_COUNTRY", 0.95));
        }
        
        // Fraud rule 2: Transaction amount > 3x user's average
        double avgAmount = history.stream().mapToDouble(Transaction::getAmount).average().orElse(0);
        if (tx.getAmount() > avgAmount * 3 && avgAmount > 0) {
            out.collect(new FraudAlert(tx.getUserId(), "AMOUNT_SPIKE", 0.85));
        }
        
        // Update state
        history.add(tx);
        if (history.size() > 10) history.remove(0);  // Keep last 10
        recentTransactions.update(history);
    }
}
```

**3. Windowing**
```
Tumbling Window: Fixed, non-overlapping windows
  [0-5min] [5-10min] [10-15min]
  Use for: Aggregations per time period (transactions per 5 min)

Sliding Window: Overlapping windows
  [0-5min] [1-6min] [2-7min]
  Use for: Moving averages, rolling counts

Session Window: Gap-based windows
  [activity...gap > 30min...activity]
  Use for: User session analysis

Global Window: All events in one window (manual trigger)
  Use for: Custom windowing logic
```

**4. Event Time vs Processing Time**
```
Processing time: When Flink processes the event
  - Simple, no out-of-order handling needed
  - Inaccurate if events arrive late

Event time: When the event actually occurred (timestamp in event)
  - Accurate even with late-arriving events
  - Requires watermarks to handle late data

Watermark: "I've seen all events up to timestamp T"
  - Flink uses watermarks to know when a window is complete
  - Late events (after watermark): handled by allowed lateness or side output

// Allow events up to 5 seconds late
WatermarkStrategy.forBoundedOutOfOrderness(Duration.ofSeconds(5))
```

---

### Flink Architecture

```
                    ┌──────────────────────────────────────────┐
                    │         Kafka (transaction events)        │
                    └──────────────────┬───────────────────────┘
                                       │
                    ┌──────────────────▼───────────────────────┐
                    │         Flink Job Manager                 │
                    │  (coordinates, schedules, checkpoints)    │
                    └──────────────────┬───────────────────────┘
                                       │
              ┌────────────────────────┼────────────────────┐
              │                        │                    │
    ┌─────────▼──────┐       ┌─────────▼──────┐   ┌────────▼───────┐
    │  Task Manager 1 │       │  Task Manager 2 │   │  Task Manager 3│
    │  (processes     │       │  (processes     │   │  (processes    │
    │   partitions    │       │   partitions    │   │   partitions   │
    │   0-3)          │       │   4-7)          │   │   8-11)        │
    │  State: RocksDB │       │  State: RocksDB │   │  State: RocksDB│
    └─────────────────┘       └─────────────────┘   └────────────────┘
                                       │
                    ┌──────────────────▼───────────────────────┐
                    │         Checkpoints (S3)                  │
                    │  Periodic snapshots of all state          │
                    │  On failure: restore from last checkpoint │
                    └──────────────────────────────────────────┘
```

---

### Key Takeaways

1. **Flink is the standard for stateful stream processing** — Kafka Streams is simpler but limited. Flink handles complex stateful operations at scale.
2. **State is the key differentiator** — Flink can maintain per-user state (transaction history) across millions of users in RocksDB.
3. **Event time over processing time** — always use event timestamps for accurate windowing. Processing time gives wrong results with late data.
4. **Checkpointing enables fault tolerance** — Flink periodically snapshots state to S3. On failure, restores from last checkpoint with exactly-once semantics.
5. **Watermarks handle late data** — define how long to wait for late events before closing a window.

**Interview Follow-up Questions:**
- "What is the difference between Flink and Spark Streaming?"
- "How does Flink achieve exactly-once processing semantics?"
- "What is a watermark in stream processing and why is it needed?"
- "How do you handle state that grows unboundedly in a Flink job?"

---

## Q4: Schema Evolution Without Downtime {#d4}

**Situation:**
Your Kafka topics carry JSON messages. The Order Service produces order events. Five downstream consumers (Analytics, Fraud, Shipping, Notifications, Warehouse) all consume these events. You need to add a new field `discount_code` to order events. Some consumers need it, others don't. You also need to rename `user_id` to `customer_id` across all services. How do you evolve schemas without breaking consumers or requiring coordinated deployments?

**The Schema Evolution Problem:**

```
Producer publishes: { "order_id": "123", "user_id": "456", "total": 99.99 }
Consumer A expects: { "order_id": "123", "user_id": "456", "total": 99.99 }

Producer changes to: { "order_id": "123", "customer_id": "456", "total": 99.99 }
Consumer A still expects "user_id" → BREAKS

With 5 consumers, coordinating simultaneous deployment is nearly impossible.
```

---

### Schema Registry (Confluent Schema Registry)

A schema registry stores and enforces schemas for Kafka topics. Producers and consumers register schemas. The registry enforces compatibility rules.

```
Compatibility modes:
  BACKWARD: New schema can read data written by old schema
    → Add optional fields, remove fields with defaults
    → Consumers can be upgraded before producers
    
  FORWARD: Old schema can read data written by new schema
    → Add fields with defaults, remove optional fields
    → Producers can be upgraded before consumers
    
  FULL: Both backward and forward compatible
    → Most restrictive, safest
    
  NONE: No compatibility checking
    → Dangerous, only for development
```

---

### Avro Schema Evolution

Avro is the standard format for Kafka with schema registry:

```json
// Version 1: Original schema
{
  "type": "record",
  "name": "OrderEvent",
  "namespace": "com.example",
  "fields": [
    {"name": "order_id", "type": "string"},
    {"name": "user_id", "type": "string"},
    {"name": "total", "type": "double"}
  ]
}

// Version 2: Add optional field (BACKWARD compatible)
{
  "type": "record",
  "name": "OrderEvent",
  "fields": [
    {"name": "order_id", "type": "string"},
    {"name": "user_id", "type": "string"},
    {"name": "total", "type": "double"},
    {"name": "discount_code", "type": ["null", "string"], "default": null}
    // ↑ Optional (union with null) + default = backward compatible
    // Old consumers: ignore this field
    // New consumers: use it if present
  ]
}
```

**Renaming a field (breaking change — requires migration):**
```json
// Cannot rename user_id → customer_id directly (breaking)
// Solution: Add alias, deprecate old field

// Version 3: Add customer_id as alias for user_id
{
  "fields": [
    {"name": "customer_id", "type": "string", "aliases": ["user_id"]},
    // Avro aliases: old consumers reading "user_id" get this field
    // New consumers use "customer_id"
    {"name": "total", "type": "double"},
    {"name": "discount_code", "type": ["null", "string"], "default": null}
  ]
}
```

---

### Migration Strategy for Breaking Changes

```
Step 1: Add new field alongside old (dual-write period)
  Producer writes both: user_id AND customer_id
  Consumers: can use either

Step 2: Migrate consumers to use new field
  Deploy each consumer to use customer_id
  No coordination needed (both fields available)

Step 3: Remove old field (after all consumers migrated)
  Producer stops writing user_id
  Old consumers: must be updated before this step

Timeline: 2-4 weeks for safe migration
```

---

### Key Takeaways

1. **Schema registry is mandatory for production Kafka** — without it, schema changes break consumers silently.
2. **Additive changes are safe** — add optional fields with defaults. Never remove or rename fields directly.
3. **Avro/Protobuf over JSON** — JSON has no schema enforcement. Avro and Protobuf enforce compatibility at the registry level.
4. **Dual-write for breaking changes** — write both old and new field simultaneously during migration window.
5. **Consumer-driven contract testing** — each consumer defines what it needs. Producer must satisfy all consumers.

**Interview Follow-up Questions:**
- "What is the difference between Avro, Protobuf, and JSON Schema?"
- "What is consumer-driven contract testing (Pact)?"
- "How do you handle schema evolution in a REST API vs a Kafka topic?"
- "What is the difference between backward and forward compatibility?"

---

## Q5: Search Architecture with Elasticsearch {#d5}

**Situation:**
Your e-commerce platform has 10 million products. Users search by keyword, filter by category/price/brand, sort by relevance/price/rating, and expect results in <100ms. PostgreSQL full-text search (`tsvector`) works for 100K products but is too slow at 10M. You need a dedicated search solution.

**Why Elasticsearch:**

```
PostgreSQL full-text search limitations:
  - Sequential scan for complex queries
  - No relevance ranking (BM25)
  - No faceted search (aggregations)
  - No fuzzy matching ("iphone" matches "iPhone")
  - Slow at 10M+ documents

Elasticsearch advantages:
  - Inverted index: O(1) term lookup
  - BM25 relevance scoring (industry standard)
  - Aggregations: faceted search (count by category, price ranges)
  - Fuzzy matching, synonyms, stemming
  - Horizontal scale: add shards for more capacity
  - Near-real-time: indexed documents searchable in <1 second
```

---

### Index Design

```json
// Product index mapping
PUT /products
{
  "settings": {
    "number_of_shards": 5,      // Distribute across 5 nodes
    "number_of_replicas": 1,    // 1 replica per shard (HA)
    "analysis": {
      "analyzer": {
        "product_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "stop", "snowball", "synonym_filter"]
        }
      },
      "filter": {
        "synonym_filter": {
          "type": "synonym",
          "synonyms": ["tv, television", "laptop, notebook", "phone, mobile, smartphone"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "product_id":   {"type": "keyword"},          // Exact match only
      "name":         {"type": "text", "analyzer": "product_analyzer"},
      "description":  {"type": "text", "analyzer": "product_analyzer"},
      "brand":        {"type": "keyword"},           // Exact match + aggregation
      "category":     {"type": "keyword"},           // Exact match + aggregation
      "price":        {"type": "float"},             // Range queries
      "rating":       {"type": "float"},
      "in_stock":     {"type": "boolean"},
      "tags":         {"type": "keyword"},
      "created_at":   {"type": "date"},
      "name_suggest": {"type": "completion"}         // Autocomplete
    }
  }
}
```

---

### Search Query

```json
// Search: "wireless headphones" under $200, in stock, sorted by rating
POST /products/_search
{
  "query": {
    "bool": {
      "must": [
        {
          "multi_match": {
            "query": "wireless headphones",
            "fields": ["name^3", "description", "tags"],  // name weighted 3x
            "fuzziness": "AUTO"  // Handle typos
          }
        }
      ],
      "filter": [
        {"term": {"in_stock": true}},
        {"range": {"price": {"lte": 200}}},
        {"term": {"category": "Electronics"}}
      ]
    }
  },
  "sort": [
    {"_score": "desc"},    // Relevance first
    {"rating": "desc"}     // Then by rating
  ],
  "aggs": {
    "brands": {
      "terms": {"field": "brand", "size": 10}  // Facet: count by brand
    },
    "price_ranges": {
      "range": {
        "field": "price",
        "ranges": [
          {"to": 50}, {"from": 50, "to": 100},
          {"from": 100, "to": 200}, {"from": 200}
        ]
      }
    }
  },
  "from": 0, "size": 20  // Pagination
}
```

---

### Keeping Elasticsearch in Sync

```
Strategy: CDC (Change Data Capture) via Kafka

PostgreSQL → Debezium → Kafka → Elasticsearch Sink Connector → Elasticsearch

On product update:
  1. PostgreSQL row updated
  2. Debezium captures WAL change → publishes to Kafka
  3. Kafka Connect Elasticsearch Sink → indexes document
  4. Latency: 1-5 seconds (near-real-time)

Alternative: Dual-write from application
  On product update:
    db.UpdateProduct(product)
    es.IndexProduct(product)
  Risk: If ES write fails, data is inconsistent
  Mitigation: Retry queue, periodic reconciliation job
```

---

### Key Takeaways

1. **Elasticsearch for search, PostgreSQL for transactions** — never use Elasticsearch as your primary database. It's eventually consistent and not ACID.
2. **Inverted index is why ES is fast** — term → list of documents. O(1) lookup regardless of dataset size.
3. **Shard sizing matters** — aim for 10-50GB per shard. Too small = overhead. Too large = slow queries.
4. **Use `keyword` for filtering/aggregation, `text` for full-text search** — `keyword` is exact match, `text` is analyzed (tokenized, stemmed).
5. **Synonyms and stemming improve recall** — "television" matches "TV", "running" matches "run". Configure analyzers carefully.

**Interview Follow-up Questions:**
- "What is an inverted index and how does it enable fast full-text search?"
- "How do you handle Elasticsearch index updates without downtime (zero-downtime reindexing)?"
- "What is the difference between a query and a filter in Elasticsearch?"
- "How do you implement autocomplete/typeahead with Elasticsearch?"

---

## Q6: WebSockets vs SSE vs Long Polling {#d6}

**Situation:**
You are building a live sports score app. Scores update every few seconds. 5 million users watch games simultaneously. You need to push score updates to all connected clients in real-time. Evaluate the three real-time communication options and choose the right one.

**Three Approaches:**

---

### Long Polling

Client makes a request. Server holds it open until data is available (or timeout), then responds. Client immediately makes another request.

```
Client: GET /scores?since=1699000000
Server: [holds connection open for up to 30 seconds]
Server: [score changes] → responds with new score
Client: immediately sends next request: GET /scores?since=1699000010
...repeat

Pros:
  ✅ Works everywhere (standard HTTP)
  ✅ No special infrastructure
  ✅ Firewall/proxy friendly

Cons:
  ❌ High overhead: new HTTP connection every update
  ❌ Latency: up to 30 seconds if no updates
  ❌ Server holds many open connections
  ❌ Not truly real-time

Use when: Simple, infrequent updates, broad compatibility needed
```

---

### Server-Sent Events (SSE)

Server pushes events to client over a persistent HTTP connection. One-directional (server → client only).

```
Client: GET /scores/stream
Server: Content-Type: text/event-stream
        [keeps connection open]
        data: {"team":"Lakers","score":98}\n\n
        data: {"team":"Warriors","score":102}\n\n
        ...

Client-side:
  const es = new EventSource('/scores/stream');
  es.onmessage = (event) => {
    const score = JSON.parse(event.data);
    updateScoreboard(score);
  };
  // Auto-reconnects if connection drops

Pros:
  ✅ Simple: standard HTTP, works through proxies
  ✅ Auto-reconnect built into browser
  ✅ HTTP/2 multiplexing (many SSE streams over one connection)
  ✅ Lower overhead than WebSocket for one-way data

Cons:
  ❌ One-directional only (server → client)
  ❌ No binary data (text only)
  ❌ Limited to 6 connections per domain in HTTP/1.1 (not an issue with HTTP/2)

Use when: Server pushes updates, client doesn't need to send data
  → Live scores, news feeds, notifications, stock prices
```

---

### WebSocket

Full-duplex, persistent TCP connection. Both client and server can send messages at any time.

```
Handshake (HTTP Upgrade):
  Client: GET /ws HTTP/1.1
          Upgrade: websocket
          Connection: Upgrade
  Server: HTTP/1.1 101 Switching Protocols
          Upgrade: websocket

After handshake: raw TCP frames (not HTTP)
  Client → Server: {"action": "subscribe", "game": "lakers-warriors"}
  Server → Client: {"type": "score", "team": "Lakers", "score": 98}
  Server → Client: {"type": "score", "team": "Warriors", "score": 102}
  Client → Server: {"action": "chat", "message": "Go Lakers!"}

Pros:
  ✅ Full-duplex: both sides can send anytime
  ✅ Low overhead: no HTTP headers after handshake
  ✅ Binary support
  ✅ Lowest latency

Cons:
  ❌ More complex: need WebSocket server, not just HTTP
  ❌ Stateful: load balancer needs sticky sessions or pub/sub routing
  ❌ Proxy/firewall issues (some block WebSocket upgrades)
  ❌ No auto-reconnect (must implement manually)

Use when: Bidirectional real-time communication
  → Chat, multiplayer games, collaborative editing, trading platforms
```

---

### Scaling WebSockets

The hard part: 5 million concurrent WebSocket connections.

```
Problem: WebSocket connections are stateful (tied to one server)
  User A connected to Server 1
  Score update → must reach Server 1 to push to User A
  
Solution: Pub/Sub routing via Redis

Architecture:
  Score Update Service → Redis Pub/Sub channel "game:lakers-warriors"
  
  Server 1 (100K connections) → subscribed to Redis channel
  Server 2 (100K connections) → subscribed to Redis channel
  Server 3 (100K connections) → subscribed to Redis channel
  
  When score updates:
    Score Service → PUBLISH "game:lakers-warriors" score_update
    All servers receive → push to their connected clients watching that game

Capacity:
  Each server: 100K WebSocket connections (Go handles this easily)
  50 servers × 100K = 5M concurrent connections
  Redis pub/sub: handles millions of messages/sec
```

---

### Decision Matrix

| Feature | Long Polling | SSE | WebSocket |
|---|---|---|---|
| Direction | Bidirectional | Server→Client | Bidirectional |
| Latency | High (seconds) | Low (ms) | Lowest (ms) |
| Overhead | High | Low | Lowest |
| Complexity | Low | Low | Medium |
| Proxy support | Excellent | Good | Variable |
| Auto-reconnect | Manual | Built-in | Manual |
| Best for | Simple updates | Live feeds | Chat/games |

**For live sports scores: SSE** — server pushes updates, clients don't send data, simpler than WebSocket.

---

### Key Takeaways

1. **SSE for server-push, WebSocket for bidirectional** — most "real-time" features only need server-push. SSE is simpler and sufficient.
2. **WebSocket requires pub/sub for horizontal scaling** — connections are stateful. Use Redis pub/sub to route messages across servers.
3. **HTTP/2 makes SSE more efficient** — multiple SSE streams multiplexed over one TCP connection.
4. **Long polling is a last resort** — use only when WebSocket/SSE are blocked by infrastructure constraints.
5. **Connection limits matter** — each WebSocket/SSE connection holds a file descriptor. Configure `ulimit` and OS settings for high connection counts.

**Interview Follow-up Questions:**
- "How do you handle WebSocket reconnection with exponential backoff?"
- "How does Socket.IO differ from raw WebSockets?"
- "How do you implement WebSocket authentication?"
- "What happens to WebSocket connections during a server deployment?"

---

## Q7: Real-Time Leaderboard Design {#d7}

**Situation:**
Your gaming platform has 10 million players. You need a global leaderboard showing the top 100 players by score, updated in real-time as scores change. Players also want to see their own rank (e.g., "You are #45,231 globally"). The leaderboard must update within 1 second of a score change.

**The Core Data Structure: Redis Sorted Set**

Redis Sorted Sets are the perfect data structure for leaderboards:
- Each member has a score (float)
- Members are always sorted by score
- O(log N) for add/update
- O(log N + K) for range queries (get top K)
- O(log N) for rank lookup

```
ZADD leaderboard 9850 "player:alice"   -- Add/update score
ZADD leaderboard 9200 "player:bob"
ZADD leaderboard 10100 "player:carol"

ZREVRANGE leaderboard 0 99 WITHSCORES  -- Top 100 players
→ [("player:carol", 10100), ("player:alice", 9850), ("player:bob", 9200)]

ZREVRANK leaderboard "player:alice"    -- Alice's rank (0-indexed)
→ 1  (rank #2, 0-indexed)

ZSCORE leaderboard "player:alice"      -- Alice's score
→ 9850

ZCARD leaderboard                      -- Total players
→ 3
```

---

### Architecture

```
Score Update Flow:
  Player completes level → Game Server
  Game Server → ZADD leaderboard {score} {player_id}  (Redis, O(log N))
  Game Server → Publish "leaderboard:updated" to Redis Pub/Sub
  
  WebSocket servers subscribed to "leaderboard:updated"
  → Push top 100 to all connected clients watching leaderboard

Rank Query Flow:
  Player requests their rank
  → ZREVRANK leaderboard {player_id}  (O(log N), ~0.1ms)
  → Return rank + score

Top 100 Query:
  → ZREVRANGE leaderboard 0 99 WITHSCORES  (O(log N + 100), ~0.5ms)
  → Cache result for 1 second (avoid hammering Redis on every page load)
```

---

### Segmented Leaderboards

Global leaderboard with 10M players is useful but users care more about their friends and region:

```
Multiple leaderboards in Redis:
  leaderboard:global          -- All 10M players
  leaderboard:region:us       -- US players only
  leaderboard:region:eu       -- EU players only
  leaderboard:friends:{user}  -- User's friends only

Friends leaderboard:
  When user loads leaderboard:
    friend_ids = db.GetFriends(user_id)
    scores = redis.ZMSCORE("leaderboard:global", friend_ids)
    // Build friends leaderboard from global scores
    // No separate sorted set needed for friends
```

---

### Time-Based Leaderboards (Weekly/Monthly)

```
Weekly leaderboard: reset every Monday
  Key: leaderboard:weekly:{year}:{week}
  TTL: 14 days (auto-expire after 2 weeks)

Score update:
  ZADD leaderboard:global score player_id
  ZADD leaderboard:weekly:2024:45 score player_id  -- Current week
  ZADD leaderboard:monthly:2024:11 score player_id  -- Current month

Weekly reset:
  New week starts → new key automatically
  Old key expires via TTL
  No explicit reset needed
```

---

### Key Takeaways

1. **Redis Sorted Set is the perfect leaderboard data structure** — O(log N) updates, O(log N) rank queries, built-in sorting.
2. **Segment leaderboards** — global, regional, friends. Users engage more with leaderboards where they can see themselves.
3. **Time-based leaderboards via key naming** — `leaderboard:weekly:2024:45`. New week = new key. TTL handles cleanup.
4. **Cache the top 100** — the top 100 is read by everyone. Cache for 1-5 seconds to avoid hammering Redis.
5. **Rank is 0-indexed in Redis** — `ZREVRANK` returns 0 for #1. Add 1 for display.

**Interview Follow-up Questions:**
- "How do you handle ties in a leaderboard (two players with the same score)?"
- "How would you design a leaderboard for 1 billion players?"
- "How do you implement a 'nearby players' leaderboard (show players ranked near you)?"
- "What are the memory implications of storing 10M players in a Redis Sorted Set?"

---

## Q8: Presence System (Online/Offline Status) {#d8}

**Situation:**
Your collaboration tool (like Slack) needs to show which users are online. 500K users are active simultaneously. Status must update within 5 seconds of a user going offline. The system must handle users on multiple devices (laptop + phone = online). You need to show "last seen" for offline users.

**The Core Challenge:**

```
How do you know a user is offline?
  - They don't explicitly disconnect (browser tab closed, network drop)
  - You must infer offline from absence of heartbeats
  
Heartbeat approach:
  Client sends heartbeat every 15 seconds
  Server: if no heartbeat for 30 seconds → user is offline
  
  TTL in Redis: SETEX "presence:{user_id}" 30 "online"
  Heartbeat: EXPIRE "presence:{user_id}" 30  (refresh TTL)
  No heartbeat: TTL expires → key deleted → user offline
```

---

### Multi-Device Presence

```
User has laptop + phone. Both are online.
User closes laptop. Phone still open.
User should still appear online.

Solution: Track per-device, aggregate to user level

Per-device presence:
  SETEX "presence:{user_id}:{device_id}" 30 "online"

User online check:
  KEYS "presence:{user_id}:*"  -- Any device online?
  → If any key exists: user is online
  → If no keys: user is offline

Better (avoid KEYS command — O(N)):
  Use a set to track active devices:
  SADD "devices:{user_id}" device_id
  SETEX "device:{device_id}" 30 "online"
  
  On heartbeat:
    EXPIRE "device:{device_id}" 30
  
  On TTL expiry (device offline):
    Use Redis keyspace notifications to detect expiry
    → Remove device_id from "devices:{user_id}" set
    → If set is empty: user is offline
```

---

### Last Seen

```
When user goes offline (TTL expires):
  Store last seen timestamp:
  HSET "user:{user_id}" "last_seen" {timestamp}
  
Display:
  If online: "Online"
  If offline: "Last seen {relative_time}"
    e.g., "Last seen 5 minutes ago"
         "Last seen yesterday at 3:45 PM"
```

---

### Scaling to 500K Concurrent Users

```
Redis memory:
  500K users × 30 bytes per key = 15MB (trivial)
  
Heartbeat load:
  500K users × 1 heartbeat/15sec = 33,333 Redis ops/sec
  Redis handles 100K+ ops/sec → no problem

Presence change notifications:
  When user goes online/offline → notify their contacts
  
  User A goes online:
    Get A's contacts (from DB or Redis set)
    Publish "presence:online" event to each contact's channel
    
  Contacts subscribed to their channel via WebSocket
    → Receive notification → Update UI
```

---

### Key Takeaways

1. **Heartbeat + TTL is the standard presence pattern** — client sends heartbeat every 15s, server TTL is 30s. No heartbeat = offline.
2. **Per-device tracking for multi-device users** — aggregate device presence to user presence.
3. **Redis keyspace notifications for TTL expiry** — detect when a device goes offline without polling.
4. **Last seen is a privacy feature** — many apps let users hide last seen (WhatsApp). Design for this from the start.
5. **Presence at scale is eventually consistent** — a user may appear online for up to 30 seconds after going offline. This is acceptable.

**Interview Follow-up Questions:**
- "How do you handle presence for users in different time zones?"
- "How would you implement 'typing indicator' (user is typing)?"
- "How do you scale presence notifications to millions of users?"
- "What are the privacy implications of presence systems?"

---

## Q9: Collaborative Editing and CRDTs {#d9}

**Situation:**
You are building a Google Docs-like collaborative editor. Multiple users edit the same document simultaneously. User A types "Hello" at position 5. User B simultaneously deletes the character at position 5. Both changes must be applied correctly without conflicts. The document must converge to the same state on all clients, even with network delays.

**The Conflict Problem:**

```
Initial state: "Hello World"
User A (offline for 2 seconds): inserts "Beautiful " at position 6
  → "Hello Beautiful World"
User B (offline for 2 seconds): deletes "World" (positions 6-10)
  → "Hello "

Both changes arrive at server simultaneously.
What is the correct merged result?
  → "Hello Beautiful " (A's insert + B's delete, both applied)
  
Naive approach (last-write-wins): One change is lost. Wrong.
Operational Transformation (OT): Complex, hard to implement correctly.
CRDT: Mathematically guaranteed convergence.
```

---

### CRDTs (Conflict-free Replicated Data Types)

A CRDT is a data structure that can be updated concurrently by multiple nodes and always converges to the same state, regardless of the order operations are applied.

**Key property: Commutativity + Associativity + Idempotency**
- Operations can be applied in any order → same result
- Operations can be applied multiple times → same result
- No coordination needed between nodes

**CRDT for Text Editing: RGA (Replicated Growable Array)**

Each character has a unique ID (timestamp + node ID). Characters are never truly deleted — they're marked as "tombstoned" (invisible but still in the structure).

```
Initial: [H(1,A), e(2,A), l(3,A), l(4,A), o(5,A)]
         "Hello"

User A inserts " Beautiful" after position 5:
  New chars: [(6,A)=' ', (7,A)='B', (8,A)='e', ...]
  Each char has unique ID: (timestamp, node_id)

User B deletes 'o' (char with ID (5,A)):
  Mark (5,A) as tombstone: (5,A, deleted=true)

Merge:
  Apply both operations in any order → same result
  [(1,A)H, (2,A)e, (3,A)l, (4,A)l, (5,A,deleted)o, (6,A) , (7,A)B, ...]
  Visible: "Hell Beautiful"
  
  Both users see: "Hell Beautiful" (correct merge)
```

---

### Practical Implementation: Yjs

Yjs is the most popular CRDT library for collaborative editing:

```javascript
// Server (Node.js)
const Y = require('yjs')
const { WebsocketProvider } = require('y-websocket')

const doc = new Y.Doc()
const text = doc.getText('content')

// Client A
text.insert(0, 'Hello')

// Client B (simultaneously, offline)
text.insert(5, ' World')

// When both sync: "Hello World" (correct merge, no conflicts)

// Observe changes
text.observe(event => {
  console.log('Document changed:', text.toString())
})
```

---

### Key Takeaways

1. **CRDTs guarantee convergence without coordination** — no central server needed to resolve conflicts. Mathematical proof of correctness.
2. **Tombstones are the key insight** — never delete, just mark as deleted. This preserves the position information needed for correct merging.
3. **Yjs is the production-ready choice** — used by Notion, Linear, and many others. Don't implement CRDTs from scratch.
4. **OT vs CRDT** — Google Docs uses OT (Operational Transformation). CRDTs are newer, simpler to reason about, and work offline-first.
5. **CRDTs enable offline-first apps** — users can edit offline, sync when reconnected. Changes always merge correctly.

**Interview Follow-up Questions:**
- "What is the difference between CRDTs and Operational Transformation?"
- "What is a tombstone in a CRDT and why is it needed?"
- "How do you handle cursor positions in a collaborative editor when text is inserted/deleted?"
- "What are the memory implications of tombstones in a long-lived document?"

---

## Q10: Leader Election and Raft Consensus {#d10}

**Situation:**
Your distributed database has 3 nodes. All 3 nodes can accept writes. Without coordination, two nodes might accept conflicting writes simultaneously. You need exactly one node to be the "leader" that accepts writes at any time. If the leader fails, a new leader must be elected automatically within 5 seconds.

**Why Leader Election is Hard:**

```
The split-brain problem:
  3 nodes: A, B, C
  Network partition: A cannot reach B and C
  
  A thinks: "B and C are down, I'm the only one alive, I'll be leader"
  B thinks: "A is down, C is alive, I'll be leader"
  
  Result: Two leaders simultaneously → conflicting writes → data corruption
  
  Solution: Quorum — a leader needs votes from majority (2 of 3 nodes)
  A alone cannot get majority → cannot become leader
  B + C = majority → B becomes leader
```

---

### Raft Consensus Algorithm

Raft is the most understandable consensus algorithm (designed to be easier than Paxos):

**Three roles:**
- **Leader**: Handles all writes, sends heartbeats to followers
- **Follower**: Accepts writes from leader, votes in elections
- **Candidate**: Requesting votes to become leader

**Leader Election:**
```
Normal operation:
  Leader sends heartbeat every 150ms to all followers
  Followers reset their election timeout (150-300ms random)

Leader fails:
  Followers stop receiving heartbeats
  First follower whose timeout expires → becomes Candidate
  Candidate increments term, votes for itself, requests votes from others
  
  If Candidate gets majority (2 of 3): becomes Leader
  If another Candidate wins: becomes Follower
  If tie: wait for random timeout, try again

Why random timeout?
  Prevents all followers from becoming candidates simultaneously
  One node will timeout first → wins election quickly
```

**Log Replication:**
```
Client writes to Leader:
  1. Leader appends to its log (uncommitted)
  2. Leader sends log entry to all followers
  3. Followers append to their logs, respond "OK"
  4. Leader receives majority acknowledgment (2 of 3)
  5. Leader commits entry (applies to state machine)
  6. Leader responds to client: "Success"
  7. Leader notifies followers: "Commit entry N"
  8. Followers commit entry

If leader fails after step 4 but before step 6:
  New leader is elected
  New leader has the committed entry (it was on majority of nodes)
  New leader commits it and responds to client
  → No data loss
```

---

### etcd: Raft in Production

etcd is the most widely used Raft implementation (used by Kubernetes):

```go
// etcd client: leader election
import clientv3 "go.etcd.io/etcd/client/v3"
import "go.etcd.io/etcd/client/v3/concurrency"

func runWithLeaderElection(client *clientv3.Client) {
    session, _ := concurrency.NewSession(client, concurrency.WithTTL(10))
    defer session.Close()
    
    election := concurrency.NewElection(session, "/my-service/leader")
    
    // Campaign: try to become leader (blocks until elected)
    if err := election.Campaign(ctx, "node-1"); err != nil {
        log.Fatal(err)
    }
    
    log.Println("I am the leader!")
    
    // Do leader work
    go doLeaderWork()
    
    // Watch for leadership loss
    <-session.Done()
    log.Println("Lost leadership (session expired)")
}
```

---

### Key Takeaways

1. **Quorum prevents split-brain** — a leader needs votes from majority (N/2 + 1). With 3 nodes, need 2 votes. A partitioned single node cannot become leader.
2. **Raft is Paxos made understandable** — same guarantees, much clearer algorithm. Use etcd or Consul (both use Raft) rather than implementing yourself.
3. **Odd number of nodes** — 3 nodes tolerates 1 failure. 5 nodes tolerates 2 failures. Even numbers waste a node (3 and 4 both tolerate 1 failure).
4. **Leader election adds latency** — during election (up to 5 seconds), writes are rejected. Design clients to retry.
5. **etcd is the production choice** — Kubernetes uses etcd for all cluster state. Battle-tested, well-documented.

**Interview Follow-up Questions:**
- "What is the difference between Raft and Paxos?"
- "How many nodes do you need to tolerate N failures?"
- "What happens to in-flight writes when a leader fails?"
- "How does ZooKeeper's ZAB protocol differ from Raft?"

---

## Q11: Distributed Locks in Practice {#d11}

**Situation:**
Your inventory service runs on 10 servers. When a customer places an order, you need to atomically check and decrement inventory. Without locking, two servers might both check inventory (100 units), both see "available", both decrement, resulting in -1 inventory (overselling). You need a distributed lock that works across all 10 servers.

**Why Local Locks Don't Work:**

```
Server 1: synchronized(inventoryLock) { check → decrement }
Server 2: synchronized(inventoryLock) { check → decrement }

Server 1's lock is local to Server 1's JVM/process.
Server 2 has its own lock — they don't coordinate.
Both can execute simultaneously → race condition.
```

---

### Redis Distributed Lock (Redlock)

```go
// Simple Redis lock (single node)
func acquireLock(key string, ttl time.Duration) (bool, error) {
    // SET key value NX PX milliseconds
    // NX: only set if not exists (atomic)
    // PX: expire in milliseconds
    lockValue := uuid.New().String()  // Unique value to identify our lock
    
    result, err := redis.SetNX(ctx, "lock:"+key, lockValue, ttl).Result()
    return result, err
}

func releaseLock(key string, lockValue string) error {
    // Must verify we own the lock before releasing
    // Use Lua script for atomic check-and-delete
    script := `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("DEL", KEYS[1])
        else
            return 0
        end
    `
    return redis.Eval(ctx, script, []string{"lock:" + key}, lockValue).Err()
}

// Usage
func decrementInventory(productID string, quantity int) error {
    lockValue := uuid.New().String()
    
    // Acquire lock with 5 second TTL
    acquired, err := acquireLock(productID, 5*time.Second)
    if !acquired {
        return errors.New("could not acquire lock, try again")
    }
    defer releaseLock(productID, lockValue)
    
    // Critical section: check and decrement
    inventory := db.GetInventory(productID)
    if inventory < quantity {
        return errors.New("insufficient inventory")
    }
    db.DecrementInventory(productID, quantity)
    return nil
}
```

**Why unique lock value matters:**
```
Without unique value:
  Server A acquires lock, starts work
  Server A's work takes longer than TTL → lock expires
  Server B acquires lock, starts work
  Server A finishes, releases lock → releases Server B's lock!
  Server C acquires lock → now A, B, C all in critical section

With unique value:
  Server A tries to release: "Is lock value mine?" → No (it's B's) → Don't release
  Server B's lock stays intact
```

---

### Redlock (Multi-Node Redis Lock)

For higher reliability, acquire lock on majority of Redis nodes:

```
5 Redis nodes (independent, no replication)

Acquire lock:
  Try to acquire on all 5 nodes simultaneously
  If acquired on 3+ (majority): lock is held
  If acquired on < 3: release all, retry

Release lock:
  Release on all 5 nodes

Why this is safer:
  If 1 Redis node fails: still have 4 nodes, can get majority (3)
  If 2 nodes fail: still have 3 nodes, can get majority (3)
  If 3 nodes fail: cannot get majority → lock acquisition fails safely
```

---

### Database-Level Locking (Simpler Alternative)

For many use cases, database-level locking is simpler and sufficient:

```sql
-- PostgreSQL advisory lock (application-level, not row-level)
SELECT pg_try_advisory_lock(hashtext('inventory:product-123'));
-- Returns true if lock acquired, false if already locked

-- Do work...

SELECT pg_advisory_unlock(hashtext('inventory:product-123'));

-- Or: SELECT FOR UPDATE (row-level lock)
BEGIN;
SELECT inventory FROM products WHERE product_id = 123 FOR UPDATE;
-- Other transactions block here until this transaction commits
UPDATE products SET inventory = inventory - 1 WHERE product_id = 123;
COMMIT;
```

---

### Key Takeaways

1. **Unique lock value prevents accidental release** — always use a UUID as the lock value. Verify ownership before releasing.
2. **TTL is your safety net** — if the lock holder crashes, TTL ensures the lock is eventually released. Set TTL to 2-3x expected operation time.
3. **Database locks for simple cases** — `SELECT FOR UPDATE` is simpler than Redis locks and sufficient for most inventory/balance operations.
4. **Redlock for high availability** — single Redis node is a SPOF for locking. Redlock across 5 nodes tolerates 2 failures.
5. **Locks are a last resort** — prefer optimistic locking (version numbers) or database transactions over distributed locks when possible.

**Interview Follow-up Questions:**
- "What is the difference between optimistic locking and pessimistic locking?"
- "What are the failure modes of Redlock? (Martin Kleppmann's critique)"
- "How do you handle lock contention at high throughput?"
- "When would you use a database advisory lock vs a Redis lock?"

---

## Q12: Time Series Databases {#d12}

**Situation:**
Your IoT platform collects metrics from 1 million sensors. Each sensor sends a reading every 10 seconds: temperature, humidity, pressure. That's 100M data points per day. You need to: store all readings for 2 years, query "average temperature for sensor X over the last 24 hours", detect anomalies in real-time, and dashboard 1000 metrics simultaneously. PostgreSQL is struggling with 50B rows.

**Why Regular Databases Struggle with Time Series:**

```
Time series data characteristics:
  - Append-only: new data always has a newer timestamp
  - High write volume: 100M inserts/day
  - Time-range queries: "last 24 hours", "last week"
  - Aggregations: avg, min, max, percentile over time windows
  - Retention: keep raw data for 30 days, aggregated for 2 years
  - Compression: sequential timestamps compress extremely well

PostgreSQL problems:
  - B-tree index on timestamp: good for point queries, slow for range scans
  - No built-in time-based partitioning (need manual partition management)
  - No automatic downsampling (keep raw data forever or delete)
  - No columnar compression for time series
```

---

### InfluxDB / TimescaleDB

**TimescaleDB** (PostgreSQL extension — best for SQL users):
```sql
-- Create hypertable (automatically partitioned by time)
CREATE TABLE sensor_readings (
    time        TIMESTAMPTZ NOT NULL,
    sensor_id   TEXT NOT NULL,
    temperature DOUBLE PRECISION,
    humidity    DOUBLE PRECISION,
    pressure    DOUBLE PRECISION
);

SELECT create_hypertable('sensor_readings', 'time', chunk_time_interval => INTERVAL '1 day');
-- TimescaleDB automatically creates daily partitions
-- Each partition is a separate file → fast time-range queries

-- Continuous aggregate: pre-compute hourly averages
CREATE MATERIALIZED VIEW hourly_avg
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS hour,
    sensor_id,
    AVG(temperature) AS avg_temp,
    MIN(temperature) AS min_temp,
    MAX(temperature) AS max_temp
FROM sensor_readings
GROUP BY hour, sensor_id;

-- Retention policy: keep raw data 30 days, aggregates 2 years
SELECT add_retention_policy('sensor_readings', INTERVAL '30 days');
SELECT add_retention_policy('hourly_avg', INTERVAL '2 years');

-- Query: average temperature for sensor X, last 24 hours
SELECT time_bucket('1 hour', time) AS hour, AVG(temperature)
FROM sensor_readings
WHERE sensor_id = 'sensor-123'
  AND time > NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour;
-- Fast: only scans today's partition (1 day of data, not 2 years)
```

**InfluxDB** (purpose-built, InfluxQL/Flux query language):
```
Write: 1M points/second on single node
Storage: 10x compression vs PostgreSQL
Retention: automatic downsampling and deletion
Cardinality: handles millions of unique tag combinations
```

---

### Key Takeaways

1. **Time-based partitioning is the key optimization** — queries for "last 24 hours" only scan today's partition, not 2 years of data.
2. **Continuous aggregates eliminate query-time computation** — pre-compute hourly/daily averages. Dashboard queries hit aggregates, not raw data.
3. **Retention policies are essential** — raw data grows forever without them. Keep raw data short-term, aggregates long-term.
4. **TimescaleDB for SQL teams** — full PostgreSQL compatibility, familiar tooling. InfluxDB for pure time series with higher write throughput.
5. **Compression is massive** — time series data compresses 10-100x because timestamps and values are sequential/similar.

**Interview Follow-up Questions:**
- "What is a hypertable in TimescaleDB?"
- "How does InfluxDB's TSM storage engine differ from B-tree storage?"
- "What is downsampling and why is it important for long-term storage?"
- "How would you design a time series database from scratch?"

---

## Q13: Change Data Capture (CDC) {#d13}

**Situation:**
You need to keep Elasticsearch in sync with PostgreSQL. You also need to populate a Redis cache when database records change. And you need to publish events to Kafka whenever an order is created or updated. Currently, you have three separate background jobs polling the database every minute. This creates 3-minute delays, high database load from polling, and missed updates when jobs fail.

**What is CDC:**

Change Data Capture reads the database's **transaction log** (WAL in PostgreSQL, binlog in MySQL) to capture every insert, update, and delete in real-time. Instead of polling, you stream changes as they happen.

```
Without CDC (polling):
  Job runs every 60 seconds
  SELECT * FROM orders WHERE updated_at > last_run_time
  Problems:
    - 60 second delay
    - Misses deletes (deleted rows have no updated_at)
    - High DB load (full table scan every minute)
    - Race conditions (updates between polls)

With CDC (WAL streaming):
  PostgreSQL WAL → Debezium → Kafka
  Every INSERT/UPDATE/DELETE captured in real-time
  Latency: < 1 second
  No polling, no missed events, no DB load
```

---

### Debezium: CDC for PostgreSQL

```yaml
# Debezium PostgreSQL connector config
{
  "name": "postgres-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "debezium",
    "database.password": "secret",
    "database.dbname": "production",
    "database.server.name": "prod",
    "table.include.list": "public.orders,public.products",
    "plugin.name": "pgoutput",  // PostgreSQL logical replication
    "slot.name": "debezium_slot"
  }
}
```

**Kafka messages produced by Debezium:**
```json
// INSERT event
{
  "op": "c",  // c=create, u=update, d=delete, r=read(snapshot)
  "ts_ms": 1699000000000,
  "before": null,
  "after": {
    "order_id": "123",
    "user_id": "456",
    "total": 99.99,
    "status": "pending",
    "created_at": "2024-11-03T10:00:00Z"
  }
}

// UPDATE event
{
  "op": "u",
  "before": {"order_id": "123", "status": "pending"},
  "after": {"order_id": "123", "status": "shipped"}
}

// DELETE event
{
  "op": "d",
  "before": {"order_id": "123", ...},
  "after": null
}
```

---

### CDC Consumers

```go
// Elasticsearch sync consumer
func syncToElasticsearch(event DebeziumEvent) {
    switch event.Op {
    case "c", "u":  // Create or update
        es.Index("orders", event.After.OrderID, event.After)
    case "d":  // Delete
        es.Delete("orders", event.Before.OrderID)
    }
}

// Redis cache invalidation consumer
func invalidateCache(event DebeziumEvent) {
    if event.Op == "u" || event.Op == "d" {
        redis.Del("order:" + event.Before.OrderID)
    }
}

// Downstream event publisher
func publishBusinessEvent(event DebeziumEvent) {
    if event.Op == "c" {
        kafka.Publish("order.created", event.After)
    } else if event.Op == "u" && event.Before.Status != event.After.Status {
        kafka.Publish("order.status_changed", event.After)
    }
}
```

---

### Key Takeaways

1. **CDC eliminates polling** — real-time changes, no database load from polling, no missed deletes.
2. **Debezium is the standard** — supports PostgreSQL, MySQL, MongoDB, SQL Server. Kafka Connect integration.
3. **WAL is the source of truth** — every committed transaction is captured, in order, exactly once.
4. **Outbox pattern + CDC** — write events to an outbox table in the same transaction as business data. CDC picks up outbox events and publishes to Kafka. Guaranteed delivery.
5. **Replication slot management** — Debezium uses a PostgreSQL replication slot. If Debezium is down, the slot accumulates WAL. Monitor slot lag to prevent disk exhaustion.

**Interview Follow-up Questions:**
- "What is a PostgreSQL replication slot and what are its risks?"
- "How does CDC differ from the outbox pattern?"
- "How do you handle schema changes in CDC (adding a column to a captured table)?"
- "What is the initial snapshot in Debezium and when is it needed?"

---

## Q14: Vector Databases and Semantic Search {#d14}

**Situation:**
Your customer support platform has 1 million support tickets. Users search for similar past tickets to find solutions. Keyword search misses semantically similar tickets: "my laptop won't turn on" doesn't match "computer fails to boot" even though they mean the same thing. You need semantic search that understands meaning, not just keywords.

**Keyword Search vs Semantic Search:**

```
Keyword search (Elasticsearch):
  Query: "laptop won't turn on"
  Matches: documents containing "laptop", "won't", "turn", "on"
  Misses: "computer fails to boot", "PC not starting", "notebook dead"

Semantic search (Vector DB):
  Query: "laptop won't turn on"
  → Convert to vector: [0.23, -0.45, 0.12, ..., 0.67]  (1536 dimensions)
  → Find vectors closest to query vector (cosine similarity)
  → Returns: "computer fails to boot" (similar meaning, similar vector)
             "PC not starting" (similar meaning)
             "notebook dead" (similar meaning)
```

---

### How Vector Search Works

```
1. Embedding model (e.g., OpenAI text-embedding-ada-002):
   Text → Dense vector (1536 floats)
   
   "laptop won't turn on" → [0.23, -0.45, 0.12, ..., 0.67]
   "computer fails to boot" → [0.21, -0.43, 0.14, ..., 0.65]
   "pizza recipe" → [-0.89, 0.34, -0.56, ..., 0.12]
   
   Similar meaning → similar vectors (close in vector space)
   Different meaning → different vectors (far in vector space)

2. Vector database stores vectors + metadata:
   { id: "ticket-123", vector: [...], text: "laptop won't turn on", category: "hardware" }

3. Similarity search:
   Query vector → find K nearest neighbors (KNN)
   Algorithm: HNSW (Hierarchical Navigable Small World) — approximate KNN
   Latency: < 10ms for 1M vectors
```

---

### Pinecone / Weaviate / pgvector

```python
# pgvector: PostgreSQL extension for vector search
# Best for: existing PostgreSQL users, < 1M vectors

CREATE EXTENSION vector;

CREATE TABLE support_tickets (
    id          BIGSERIAL PRIMARY KEY,
    text        TEXT,
    embedding   vector(1536),  -- OpenAI ada-002 dimension
    category    TEXT,
    created_at  TIMESTAMP
);

CREATE INDEX ON support_tickets USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);  -- IVFFlat index for approximate KNN

-- Semantic search: find 5 most similar tickets
SELECT id, text, 1 - (embedding <=> query_embedding) AS similarity
FROM support_tickets
ORDER BY embedding <=> query_embedding  -- <=> = cosine distance
LIMIT 5;
```

```python
# Generate embedding with OpenAI
import openai

def get_embedding(text):
    response = openai.Embedding.create(
        input=text,
        model="text-embedding-ada-002"
    )
    return response['data'][0]['embedding']

# Index a ticket
embedding = get_embedding("laptop won't turn on")
db.execute("INSERT INTO support_tickets (text, embedding) VALUES ($1, $2)",
           "laptop won't turn on", embedding)

# Search
query_embedding = get_embedding("computer fails to boot")
results = db.execute("""
    SELECT id, text, 1 - (embedding <=> $1) AS similarity
    FROM support_tickets
    ORDER BY embedding <=> $1
    LIMIT 5
""", query_embedding)
```

---

### Key Takeaways

1. **Semantic search understands meaning, not keywords** — essential for support tickets, product search, document retrieval.
2. **Embeddings are the foundation** — convert text to vectors using a pre-trained model (OpenAI, Sentence Transformers). Similar text → similar vectors.
3. **pgvector for small scale, Pinecone/Weaviate for large** — pgvector works well up to ~1M vectors. Dedicated vector DBs scale to billions.
4. **Hybrid search is best** — combine keyword search (BM25) with semantic search (vector). Keyword for exact matches, semantic for meaning.
5. **Embedding model choice matters** — OpenAI ada-002 is the standard. Domain-specific models (medical, legal) outperform general models for specialized content.

**Interview Follow-up Questions:**
- "What is the difference between sparse vectors (BM25) and dense vectors (embeddings)?"
- "How do you handle multilingual semantic search?"
- "What is RAG (Retrieval Augmented Generation) and how does it use vector search?"
- "How do you evaluate the quality of a semantic search system?"

---

## Q15: Event Sourcing Pattern {#d15}

**Situation:**
Your banking application stores account balances. A bug in your code incorrectly calculated interest for 10,000 accounts last month. You need to recalculate the correct balances. With a traditional database (storing current state only), you cannot replay history — the incorrect balances are the only record. You need an architecture where you can replay all transactions to reconstruct any past state.

**Traditional State Storage vs Event Sourcing:**

```
Traditional (store current state):
  accounts table: { account_id, balance, updated_at }
  
  After 1000 transactions: balance = $5,432.10
  
  Problem: History is lost. Cannot answer:
    - What was the balance on March 15?
    - Which transaction caused the balance to go negative?
    - Replay transactions with corrected interest calculation

Event Sourcing (store events):
  events table: { event_id, account_id, type, amount, timestamp }
  
  Events:
    { type: "deposit", amount: 1000, timestamp: Jan 1 }
    { type: "withdrawal", amount: 200, timestamp: Jan 5 }
    { type: "interest", amount: 32.10, timestamp: Feb 1 }
    ...
  
  Current balance = replay all events = $5,432.10
  
  Can answer:
    - Balance on March 15: replay events up to March 15
    - Fix interest bug: replay with corrected interest calculation
    - Full audit trail: every change is recorded
```

---

### Event Sourcing Implementation

```go
// Event types
type AccountEvent struct {
    EventID    string    `json:"event_id"`
    AccountID  string    `json:"account_id"`
    EventType  string    `json:"event_type"`  // "deposit", "withdrawal", "interest"
    Amount     float64   `json:"amount"`
    Timestamp  time.Time `json:"timestamp"`
    Version    int       `json:"version"`     // Optimistic concurrency
}

// Append event (never update or delete)
func appendEvent(event AccountEvent) error {
    _, err := db.Exec(`
        INSERT INTO account_events (event_id, account_id, event_type, amount, timestamp, version)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, event.EventID, event.AccountID, event.EventType, event.Amount, event.Timestamp, event.Version)
    return err
}

// Reconstruct current state by replaying events
func getAccountBalance(accountID string) (float64, error) {
    events, err := db.Query(`
        SELECT event_type, amount FROM account_events
        WHERE account_id = $1
        ORDER BY version ASC
    `, accountID)
    
    balance := 0.0
    for _, event := range events {
        switch event.EventType {
        case "deposit":
            balance += event.Amount
        case "withdrawal":
            balance -= event.Amount
        case "interest":
            balance += event.Amount
        case "fee":
            balance -= event.Amount
        }
    }
    return balance, nil
}

// Replay with corrected interest calculation (bug fix)
func replayWithCorrectedInterest(accountID string) (float64, error) {
    events, _ := db.Query(`SELECT * FROM account_events WHERE account_id = $1 ORDER BY version`, accountID)
    
    balance := 0.0
    for _, event := range events {
        switch event.EventType {
        case "deposit":
            balance += event.Amount
        case "withdrawal":
            balance -= event.Amount
        case "interest":
            // Apply corrected interest calculation
            balance += calculateCorrectInterest(balance, event.Timestamp)
        }
    }
    return balance, nil
}
```

---

### Snapshots (Performance Optimization)

Replaying 10 years of events for every balance query is slow. Use snapshots:

```go
// Snapshot: store current state periodically
type AccountSnapshot struct {
    AccountID string
    Balance   float64
    Version   int       // Last event version included in snapshot
    CreatedAt time.Time
}

// Get balance: start from latest snapshot, replay only newer events
func getBalanceWithSnapshot(accountID string) (float64, error) {
    // Get latest snapshot
    snapshot, _ := db.QueryRow(`
        SELECT balance, version FROM account_snapshots
        WHERE account_id = $1
        ORDER BY version DESC LIMIT 1
    `, accountID)
    
    // Replay only events after snapshot
    events, _ := db.Query(`
        SELECT event_type, amount FROM account_events
        WHERE account_id = $1 AND version > $2
        ORDER BY version ASC
    `, accountID, snapshot.Version)
    
    balance := snapshot.Balance
    for _, event := range events {
        // Apply events...
    }
    return balance, nil
}
```

---

### Key Takeaways

1. **Event sourcing is an audit log that is also your database** — every change is recorded. You can reconstruct any past state.
2. **Events are immutable** — never update or delete events. Corrections are new events ("reversal", "adjustment").
3. **Snapshots prevent performance degradation** — take snapshots every 100-1000 events. Replay only from last snapshot.
4. **Event sourcing + CQRS is a natural pair** — write model: append events. Read model: pre-computed projections from events.
5. **Not for everything** — event sourcing adds complexity. Use it when audit trail, time travel, or event replay are genuine requirements (finance, healthcare, legal).

**Interview Follow-up Questions:**
- "What is the difference between event sourcing and event-driven architecture?"
- "How do you handle schema evolution in event sourcing (changing event structure)?"
- "What is a projection in event sourcing?"
- "How do you implement eventual consistency between the event store and read models?"

---

---

## Quick Reference: Data & Real-Time Cheat Sheet

### Storage Technology Selection

```
Relational data + ACID transactions:     PostgreSQL
Write-heavy time series:                 TimescaleDB / InfluxDB
Document store (flexible schema):        MongoDB
Wide-column (high write, time-series):   Cassandra
Key-value cache:                         Redis
Full-text + faceted search:              Elasticsearch
Semantic / vector search:                pgvector / Pinecone
Graph data (relationships):              Neo4j
Data warehouse (analytics):              Snowflake / BigQuery
Object storage (files, media):           S3
```

### Real-Time Communication Decision Tree

```
Server pushes data, client reads only?   → SSE
Bidirectional (chat, games)?             → WebSocket
Simple, infrequent updates?              → Long Polling
High-performance internal streaming?     → gRPC streaming
```

### Data Pipeline Patterns

```
Batch ETL (hours delay):     Airflow + dbt + Snowflake
Near-real-time (< 5 min):    Kafka + Flink + Data Warehouse
Real-time (< 1 sec):         Kafka + Flink stateful processing
DB sync to search/cache:     CDC (Debezium) → Kafka → consumers
```

### Consensus & Coordination

```
Leader election:             etcd (Raft) or ZooKeeper (ZAB)
Distributed lock:            Redis SET NX + Lua, or DB advisory lock
Distributed counter:         Redis INCR (atomic)
Distributed queue:           Kafka or Redis Streams
```

### Key Numbers

```
Redis Sorted Set:
  ZADD:      O(log N)
  ZREVRANGE: O(log N + K)
  ZREVRANK:  O(log N)
  1M members: ~100MB RAM, sub-millisecond queries

Elasticsearch:
  Index size: ~10x raw data size
  Shard size: 10-50GB optimal
  Search latency: 10-100ms for complex queries

Vector Search (pgvector):
  1M vectors × 1536 dims = ~6GB RAM
  KNN search: 10-50ms
  Exact KNN: O(N), Approximate (HNSW): O(log N)

Flink:
  Throughput: 1M+ events/sec per node
  State backend: RocksDB (disk-backed, handles TB of state)
  Checkpoint interval: 1-10 minutes (recovery point)
```

### Architect Interview: Data Questions to Expect

```
"How would you design a real-time analytics dashboard?"
→ Kafka → Flink → TimescaleDB → WebSocket push to browser

"How do you keep a search index in sync with your database?"
→ CDC (Debezium) → Kafka → Elasticsearch Sink Connector

"How would you implement a feature like Google Docs?"
→ CRDTs (Yjs) + WebSocket + event sourcing for history

"How do you design a system that can replay historical data?"
→ Event sourcing: store events, not state. Replay to reconstruct.

"How do you search 10M products by meaning, not keywords?"
→ Embeddings → pgvector/Pinecone → semantic similarity search
```

---

*Document covers: Data Lake/Warehouse/Lakehouse, ETL/ELT, Stream Processing (Flink), Schema Evolution, Elasticsearch, WebSockets/SSE, Leaderboards, Presence Systems, CRDTs, Raft Consensus, Distributed Locks, Time Series DBs, CDC, Vector Databases, Event Sourcing*

*Total across all 4 files: 60 questions + 10 system designs = complete software architect interview preparation*
