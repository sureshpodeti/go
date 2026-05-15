# Reliability, Security & API Design — 15 Situation-Based Questions
## Software Architect Interview Preparation

---

## Overview

This document covers the **production engineering** side of software architecture — the topics that separate architects who can design systems from those who can run them reliably at scale.

**Topics Covered:**
- SLO/SLA/SLI and Error Budgets
- Chaos Engineering
- Disaster Recovery (RTO/RPO)
- Zero-Downtime Database Migrations
- Blue-Green, Canary, and Rolling Deployments
- Feature Flags
- OAuth2 / JWT / API Security
- Zero-Trust Networking
- Secrets Management
- DDoS Protection and WAF
- API Design (REST vs GraphQL vs gRPC)
- API Versioning
- Idempotency
- Cost Optimization at Scale
- Graceful Degradation

---

## Table of Contents

1. [Q1 — SLO, SLA, SLI and Error Budgets](#r1)
2. [Q2 — Chaos Engineering in Production](#r2)
3. [Q3 — Disaster Recovery: RTO and RPO in Practice](#r3)
4. [Q4 — Zero-Downtime Database Migrations](#r4)
5. [Q5 — Deployment Strategies: Blue-Green, Canary, Rolling](#r5)
6. [Q6 — Feature Flags for Safe Rollouts](#r6)
7. [Q7 — OAuth2, JWT, and API Security at Scale](#r7)
8. [Q8 — Zero-Trust Networking and mTLS](#r8)
9. [Q9 — Secrets Management (Vault, AWS Secrets Manager)](#r9)
10. [Q10 — DDoS Protection and WAF Architecture](#r10)
11. [Q11 — REST vs GraphQL vs gRPC — When to Use Each](#r11)
12. [Q12 — API Versioning Strategies](#r12)
13. [Q13 — Idempotency in Distributed Systems](#r13)
14. [Q14 — Cloud Cost Optimization at Scale](#r14)
15. [Q15 — Graceful Degradation Patterns](#r15)

---

## Q1: SLO, SLA, SLI and Error Budgets {#r1}

**Situation:**
Your team is constantly firefighting. Every week there's an incident. Engineers are burned out. Management wants "five nines" (99.999% uptime) but the team has no formal reliability targets. Incidents are handled reactively. There's no agreement on what "reliable enough" means. You need to establish a reliability framework.

**The Three Definitions:**

**SLI (Service Level Indicator)** — A metric that measures service behavior.
```
Examples:
  - Request success rate: (successful requests / total requests) × 100
  - Latency: P99 response time
  - Error rate: (5xx responses / total responses) × 100
  - Availability: (uptime minutes / total minutes) × 100
  - Throughput: requests per second
```

**SLO (Service Level Objective)** — Your internal target for an SLI.
```
Examples:
  - 99.9% of requests succeed (success rate SLO)
  - P99 latency < 200ms
  - Error rate < 0.1%
  - Availability > 99.9%

SLO is your internal goal. You set it. You own it.
```

**SLA (Service Level Agreement)** — A contractual commitment to customers.
```
SLA is typically more lenient than SLO:
  - SLO: 99.9% availability (internal target)
  - SLA: 99.5% availability (customer contract)
  
The gap (0.4%) is your buffer. If you breach SLO, you have time to fix
before breaching SLA and triggering penalties.
```

---

### Error Budget

**Error budget = 100% - SLO**

```
SLO: 99.9% availability
Error budget: 0.1% = 43.8 minutes/month of allowed downtime

SLO: 99.99% availability
Error budget: 0.01% = 4.38 minutes/month

SLO: 99.999% availability (5 nines)
Error budget: 0.001% = 26 seconds/month
```

**Why Error Budgets Change Everything:**

Without error budgets: "We need 100% uptime!" → Impossible → Engineers burned out → No deployments → System stagnates.

With error budgets: "We have 43 minutes/month to spend on risk." → Enables rational decisions:
- Deploy new features? Costs some error budget (risk of bugs)
- Run chaos experiments? Costs some error budget (intentional failures)
- If budget is exhausted: freeze deployments until next month

```
Error Budget Policy:
  Budget remaining > 50%: Deploy freely, run experiments
  Budget remaining 10-50%: Deploy with extra caution, no experiments
  Budget remaining < 10%: Freeze non-critical deployments
  Budget exhausted: Only critical fixes, post-mortem required
```

---

### Choosing the Right SLO

**Don't set SLOs too high:**
- 99.999% = 26 seconds/month downtime
- Requires massive investment in redundancy
- Any deployment is terrifying
- Engineers cannot sleep

**Don't set SLOs too low:**
- 99% = 7.3 hours/month downtime
- Users notice and complain
- Damages business

**Right approach — ask users:**
- What latency makes users abandon? (Usually >3 seconds)
- What error rate do users notice? (Usually >1%)
- What downtime is acceptable? (Depends on business)

**Typical SLOs by service type:**
```
Consumer-facing API: 99.9% availability, P99 < 500ms
Internal API: 99.5% availability, P99 < 1000ms
Batch jobs: 99% success rate, complete within 2x expected time
Data pipelines: 99.5% freshness (data < 1 hour old)
```

---

### Measuring SLIs

```go
// Prometheus metrics for SLI measurement
var (
    requestTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "http_requests_total"},
        []string{"method", "path", "status_code"},
    )
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
        },
        []string{"method", "path"},
    )
)

// SLI queries in Prometheus:
// Success rate (last 30 days):
// sum(rate(http_requests_total{status_code!~"5.."}[30d]))
// / sum(rate(http_requests_total[30d]))

// P99 latency:
// histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

// Error budget remaining:
// 1 - (sum(rate(http_requests_total{status_code=~"5.."}[30d]))
//      / sum(rate(http_requests_total[30d])))
// / 0.001  -- 0.001 = 0.1% error budget
```

---

### Key Takeaways

1. **SLO is your reliability contract with yourself** — set it based on user needs, not aspirations.
2. **Error budget makes reliability a shared responsibility** — product wants features (spends budget), SRE wants stability (saves budget). Both have skin in the game.
3. **"Five nines" is almost never the right answer** — 26 seconds/month is extremely expensive to achieve and maintain. Most services need 99.9% (43 min/month).
4. **Measure SLIs from the user's perspective** — not server uptime, but successful requests from the client's point of view.
5. **Toil is the enemy of reliability** — manual, repetitive work that doesn't improve the system. Automate it or eliminate it.

**Interview Follow-up Questions:**
- "How do you handle a situation where the error budget is exhausted mid-month?"
- "What is the difference between availability and reliability?"
- "How do you set SLOs for a new service with no historical data?"
- "What is toil in SRE and how do you measure it?"

---

## Q2: Chaos Engineering in Production {#r2}

**Situation:**
Your team claims the system is resilient. You have circuit breakers, retries, and redundancy. But you've never actually tested what happens when a database goes down, a network partition occurs, or a service crashes. Last month, a disk failure on one Kafka broker caused a 4-hour outage because the failover procedure had never been tested and the runbook was outdated. You need to build confidence in your system's resilience through controlled experiments.

**What is Chaos Engineering:**

Chaos Engineering is the practice of **intentionally injecting failures** into a production (or production-like) system to discover weaknesses before they cause unplanned outages.

The key insight: **You will have failures. The question is whether you discover them during a controlled experiment or during a 3 AM incident.**

**Chaos Engineering Principles (Netflix's approach):**
1. Define "steady state" — what does normal look like? (metrics, error rate, latency)
2. Hypothesize that steady state will continue in both control and experimental groups
3. Introduce variables that reflect real-world events (server crash, network latency, disk full)
4. Try to disprove the hypothesis — if steady state holds, confidence increases
5. If steady state breaks, you found a weakness — fix it before a real incident

---

### Failure Injection Categories

**Infrastructure failures:**
```
- Kill a random pod/container (Chaos Monkey)
- Kill an entire availability zone
- Fill disk to 95%
- Exhaust file descriptors
- CPU stress (100% CPU on one node)
- Memory pressure (OOM killer)
```

**Network failures:**
```
- Add 200ms latency to all requests to Service X
- Drop 10% of packets between Service A and Service B
- Partition: Service A cannot reach Service B at all
- DNS failure: Service cannot resolve downstream service
- Bandwidth throttling: limit to 1Mbps
```

**Application failures:**
```
- Kill the primary database, verify replica promotion
- Kill the leader in a Raft cluster
- Corrupt a message in Kafka
- Inject 500 errors from a downstream service
- Slow down a dependency (add 5 second delay)
```

---

### Chaos Engineering Maturity Model

```
Level 1 — Test in staging:
  Run chaos experiments in staging environment
  Low risk, low confidence (staging ≠ production)
  Good starting point

Level 2 — Test in production (off-peak):
  Run experiments at 2 AM on Sunday
  Higher confidence, lower blast radius
  Most teams should be here

Level 3 — Test in production (business hours):
  Run experiments during normal traffic
  Highest confidence
  Requires mature monitoring and rollback

Level 4 — Continuous chaos:
  Chaos runs automatically as part of CI/CD
  Netflix's Chaos Monkey runs continuously
  Only for very mature organizations
```

---

### Implementation: Chaos Toolkit

```yaml
# chaos-experiment.yaml — Kill a pod and verify recovery
version: "1.0.0"
title: "Kill Order Service pod and verify recovery"
description: "Verify that killing one Order Service pod does not affect availability"

steady-state-hypothesis:
  title: "Order Service is healthy"
  probes:
    - type: probe
      name: "order-service-responds"
      provider:
        type: http
        url: "https://api.example.com/health"
        timeout: 3
      tolerance: 200  # HTTP 200

method:
  - type: action
    name: "kill-order-service-pod"
    provider:
      type: process
      path: kubectl
      arguments: "delete pod -l app=order-service --field-selector=status.phase=Running -n production --wait=false"
    pauses:
      after: 30  # Wait 30 seconds after killing pod

rollbacks:
  - type: action
    name: "ensure-minimum-replicas"
    provider:
      type: process
      path: kubectl
      arguments: "scale deployment order-service --replicas=3 -n production"
```

---

### GameDay: Structured Chaos

A GameDay is a scheduled event where the team intentionally breaks things and practices recovery:

```
GameDay Agenda (4 hours):
  9:00 AM: Brief — what are we testing today?
  9:15 AM: Establish baseline metrics
  9:30 AM: Experiment 1 — Kill primary database
           Observe: Does replica promote? How long? Any errors?
  10:00 AM: Debrief — what happened, what surprised us?
  10:15 AM: Experiment 2 — Network partition between services
  11:00 AM: Debrief
  11:15 AM: Experiment 3 — Exhaust connection pool
  12:00 PM: Final debrief — action items, runbook updates

Output:
  - List of weaknesses discovered
  - Updated runbooks
  - Action items with owners and deadlines
  - Increased team confidence
```

---

### Key Takeaways

1. **Untested resilience is not resilience** — circuit breakers and retries only work if they've been tested under real failure conditions.
2. **Start in staging, graduate to production** — don't start chaos experiments in production on day one.
3. **Always have a rollback plan** — every chaos experiment must have a defined rollback procedure.
4. **Monitor during experiments** — you need real-time visibility into what's happening. Chaos without observability is just breaking things.
5. **Fix what you find** — chaos engineering is only valuable if discovered weaknesses are fixed. Track action items.
6. **Chaos Monkey is not chaos engineering** — randomly killing pods is a starting point, not a complete program.

**Interview Follow-up Questions:**
- "How do you get organizational buy-in for chaos engineering?"
- "What is the difference between chaos engineering and load testing?"
- "How do you ensure chaos experiments don't cause real customer impact?"
- "What tools do you use for chaos engineering? (Chaos Monkey, Gremlin, Chaos Toolkit, Litmus)"

---

## Q3: Disaster Recovery — RTO and RPO in Practice {#r3}

**Situation:**
Your company's primary data center in US-East suffers a catastrophic failure (power outage + cooling failure). All servers are down. You have a DR site in US-West with daily database backups. It takes your team 6 hours to restore service. During those 6 hours, you lose all transactions from the last 24 hours (since the last backup). Customers are furious. The CEO asks: "How do we prevent this?" You need to design a proper DR strategy.

**Two Critical Metrics:**

**RTO (Recovery Time Objective)** — How long can the business be down?
```
"We can tolerate 4 hours of downtime before significant business impact"
RTO = 4 hours

This drives: How fast must your recovery process be?
  RTO = 4 hours → You need automated failover or very fast manual process
  RTO = 1 hour → You need hot standby with automated failover
  RTO = 5 minutes → You need active-active multi-region
  RTO = 0 → You need active-active with no single point of failure
```

**RPO (Recovery Point Objective)** — How much data loss is acceptable?
```
"We can lose at most 1 hour of transactions"
RPO = 1 hour

This drives: How frequently must you replicate/backup data?
  RPO = 24 hours → Daily backups sufficient
  RPO = 1 hour → Hourly backups or continuous replication
  RPO = 5 minutes → Near-real-time replication
  RPO = 0 → Synchronous replication (no data loss ever)
```

---

### DR Tiers

```
Tier 1 — Backup and Restore (RTO: hours, RPO: hours)
  - Daily backups to S3
  - Restore from backup when disaster occurs
  - Cheapest, slowest
  - Acceptable for: dev/test, non-critical systems

Tier 2 — Pilot Light (RTO: 30-60 min, RPO: minutes)
  - Minimal infrastructure always running in DR site
  - Database replication running continuously
  - On disaster: scale up DR infrastructure, point DNS to DR
  - Medium cost
  - Acceptable for: internal tools, non-customer-facing

Tier 3 — Warm Standby (RTO: 5-15 min, RPO: seconds)
  - Scaled-down but fully functional system in DR site
  - Continuous database replication
  - On disaster: scale up DR site, failover DNS
  - Higher cost
  - Acceptable for: most production systems

Tier 4 — Active-Active (RTO: seconds, RPO: 0)
  - Both sites handle traffic simultaneously
  - No failover needed — traffic automatically reroutes
  - Highest cost
  - Required for: payments, healthcare, critical infrastructure
```

---

### Implementing Warm Standby

```
Primary (US-East):
  - 10 API servers (full capacity)
  - PostgreSQL primary
  - Redis cluster (3 nodes)
  - Kafka cluster (5 brokers)

DR (US-West):
  - 2 API servers (scaled down, can scale to 10 in 5 min)
  - PostgreSQL replica (streaming replication, lag < 1 second)
  - Redis replica (async replication)
  - Kafka MirrorMaker (replicates topics from US-East)

Failover procedure (automated):
  1. Health check detects US-East failure (3 consecutive failures × 10 sec = 30 sec)
  2. Route 53 health check triggers DNS failover to US-West
  3. DNS TTL: 60 seconds (all clients switch within 60 sec)
  4. US-West PostgreSQL replica promoted to primary (30 sec)
  5. US-West API servers scale from 2 to 10 (3 min via Auto Scaling)
  6. Total RTO: ~5 minutes

Data loss (RPO):
  - PostgreSQL replication lag: typically < 1 second
  - In-flight transactions at time of failure: lost (< 1 second)
  - RPO: ~1 second
```

---

### DR Testing

**The most important rule: Test your DR plan regularly.**

```
DR Test Schedule:
  Monthly: Verify backups are restorable (restore to test environment)
  Quarterly: Full DR failover drill (fail over to DR site, run for 1 hour)
  Annually: Full disaster simulation (pretend primary is gone, run on DR for 1 day)

DR Test Checklist:
  □ Can you restore the database from backup?
  □ How long does it take?
  □ Is the restored data correct?
  □ Can the application connect to the DR database?
  □ Are all services functional in DR?
  □ Is DNS failover working?
  □ Are monitoring and alerting working in DR?
  □ Do runbooks reflect current reality?
```

---

### Key Takeaways

1. **RTO and RPO are business decisions, not technical ones** — ask the business: "How much downtime and data loss can you afford?" Then design to meet those requirements.
2. **Untested DR plans fail** — the worst time to discover your DR plan doesn't work is during an actual disaster. Test quarterly.
3. **RPO = 0 requires synchronous replication** — which adds latency to every write. There's always a trade-off between RPO and write performance.
4. **DNS TTL is your failover speed** — set TTL to 60 seconds for critical services. Low TTL means faster failover but more DNS queries.
5. **Automate failover** — manual failover takes 30-60 minutes. Automated failover takes 1-5 minutes. For RTO < 15 minutes, automation is required.

**Interview Follow-up Questions:**
- "What is the difference between RTO and RPO? Give an example of each."
- "How do you test a DR plan without causing a real outage?"
- "What is the trade-off between RPO=0 (synchronous replication) and write latency?"
- "How would you design a DR strategy for a system with RTO=5 minutes and RPO=0?"

---

## Q4: Zero-Downtime Database Migrations {#r4}

**Situation:**
You need to add a `NOT NULL` column to the `orders` table which has 500 million rows. The naive approach (`ALTER TABLE orders ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending'`) will lock the entire table for 2-4 hours while PostgreSQL rewrites every row. During this time, no reads or writes can occur. You cannot afford 4 hours of downtime. You need to add this column with zero downtime.

**Why Naive Migrations Fail at Scale:**

```
ALTER TABLE orders ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending';

What PostgreSQL does:
  1. Acquires ACCESS EXCLUSIVE lock (blocks ALL reads and writes)
  2. Rewrites entire table (500M rows × 200 bytes = 100GB)
  3. Rebuilds all indexes
  4. Releases lock

Time: 100GB / 500MB/sec disk = 200 seconds minimum
With index rebuilds: 2-4 hours
During this time: ALL queries wait → application appears down
```

---

### The Expand-Contract Pattern (Safe Migration)

Break every migration into 3 phases deployed separately:

**Phase 1 — Expand (Add nullable column, deploy code that writes to both)**
```sql
-- Step 1: Add column as NULLABLE (instant, no table rewrite)
ALTER TABLE orders ADD COLUMN status VARCHAR(20);
-- This is instant — PostgreSQL just updates the schema, no row rewrite

-- Deploy application code that:
--   Writes to BOTH old and new column
--   Reads from OLD column (new column may be NULL for old rows)
```

**Phase 2 — Backfill (Fill in existing rows, in batches)**
```sql
-- Backfill in small batches to avoid locking
-- Run this as a background job, not a migration
DO $$
DECLARE
    batch_size INT := 10000;
    last_id BIGINT := 0;
    max_id BIGINT;
BEGIN
    SELECT MAX(id) INTO max_id FROM orders;
    
    WHILE last_id < max_id LOOP
        UPDATE orders
        SET status = 'completed'  -- or derive from existing data
        WHERE id > last_id
          AND id <= last_id + batch_size
          AND status IS NULL;
        
        last_id := last_id + batch_size;
        PERFORM pg_sleep(0.1);  -- 100ms pause between batches
    END LOOP;
END $$;

-- This runs for hours but:
-- Each batch takes <100ms (small lock window)
-- Application continues running normally
-- No downtime
```

**Phase 3 — Contract (Add NOT NULL constraint, switch reads to new column)**
```sql
-- After backfill is complete (verify: SELECT COUNT(*) FROM orders WHERE status IS NULL = 0)

-- Add NOT NULL constraint (fast — PostgreSQL validates without table rewrite if using NOT VALID)
ALTER TABLE orders ALTER COLUMN status SET NOT NULL;

-- Deploy application code that:
--   Reads from NEW column
--   Writes to NEW column only

-- Later: drop old column if applicable
```

---

### Adding an Index Without Downtime

```sql
-- WRONG: Locks table during index build
CREATE INDEX idx_orders_status ON orders(status);

-- RIGHT: Build index concurrently (no lock, but takes longer)
CREATE INDEX CONCURRENTLY idx_orders_status ON orders(status);
-- Takes 2-3x longer but never locks the table
-- Application continues running normally
```

---

### Renaming a Column Safely

```sql
-- WRONG: Rename directly (breaks application immediately)
ALTER TABLE orders RENAME COLUMN user_id TO customer_id;

-- RIGHT: Expand-Contract
-- Phase 1: Add new column
ALTER TABLE orders ADD COLUMN customer_id BIGINT;

-- Phase 2: Backfill
UPDATE orders SET customer_id = user_id WHERE customer_id IS NULL;

-- Phase 3: Deploy code that reads customer_id, writes both
-- Phase 4: Deploy code that only uses customer_id
-- Phase 5: Drop old column
ALTER TABLE orders DROP COLUMN user_id;
```

---

### Key Takeaways

1. **Never run long-running migrations on production tables** — any migration that touches every row will lock the table.
2. **Expand-Contract is the universal pattern** — add nullable → backfill → add constraint. Three separate deployments.
3. **`CREATE INDEX CONCURRENTLY`** — always use this for large tables. Never `CREATE INDEX` without CONCURRENTLY.
4. **Batch your backfills** — never `UPDATE orders SET status = 'x'` without a WHERE clause limiting rows. Always batch with sleep between batches.
5. **Test migrations on a copy of production data** — staging with 1000 rows doesn't reveal migration performance issues. Test with production-scale data.
6. **Schema migrations are code** — version control them, review them, test them.

**Interview Follow-up Questions:**
- "How do you handle a migration that requires changing a column's data type?"
- "What is the Expand-Contract pattern and when do you use it?"
- "How do you roll back a database migration that has already been applied?"
- "How do you handle migrations in a microservices environment where multiple services share a database?"

---

## Q5: Deployment Strategies — Blue-Green, Canary, Rolling {#r5}

**Situation:**
Your team deploys once a month because deployments are risky and require 2-hour maintenance windows. A bad deployment last quarter caused a 3-hour outage affecting all users. You need to move to continuous deployment (multiple times per day) with zero downtime and the ability to instantly roll back if something goes wrong.

**Three Deployment Strategies:**

---

### Blue-Green Deployment

Maintain two identical production environments. Only one is live at a time.

```
Blue (current live):  v1.0 — receiving 100% of traffic
Green (new version):  v2.0 — deployed, tested, ready

Switch:
  Load balancer: route 100% traffic from Blue → Green
  Takes: seconds (DNS change or LB config update)
  
Rollback:
  Load balancer: route 100% traffic back to Blue
  Takes: seconds
  
Pros:
  ✅ Instant rollback (switch back to Blue)
  ✅ Zero downtime
  ✅ Full testing of Green before switch
  ✅ Blue stays warm as rollback target

Cons:
  ❌ 2x infrastructure cost (two full environments)
  ❌ Database migrations must be backward compatible
     (both Blue and Green may run simultaneously during switch)
  ❌ Stateful services are complex (sessions, connections)
```

---

### Canary Deployment

Gradually shift traffic from old version to new version.

```
Stage 1: v2.0 receives 1% of traffic, v1.0 receives 99%
  Monitor for 30 minutes: error rate, latency, business metrics
  
Stage 2: v2.0 receives 10% of traffic
  Monitor for 1 hour
  
Stage 3: v2.0 receives 50% of traffic
  Monitor for 2 hours
  
Stage 4: v2.0 receives 100% of traffic
  Decommission v1.0

Rollback at any stage:
  Route 100% back to v1.0 (instant)

Pros:
  ✅ Real production traffic tests new version
  ✅ Blast radius limited (only 1% affected initially)
  ✅ Gradual confidence building
  ✅ Can detect issues that only appear under real load

Cons:
  ❌ Slower rollout (hours vs seconds)
  ❌ Two versions running simultaneously (API compatibility)
  ❌ More complex monitoring (must compare v1 vs v2 metrics)
```

**Canary with Kubernetes:**
```yaml
# v1.0: 9 replicas (90% traffic)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-v1
spec:
  replicas: 9
  selector:
    matchLabels:
      app: api
      version: v1

---
# v2.0: 1 replica (10% traffic)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-v2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: api
      version: v2

# Service selects both (traffic split by replica count)
apiVersion: v1
kind: Service
spec:
  selector:
    app: api  # Selects both v1 and v2 pods
```

---

### Rolling Deployment

Replace instances one at a time (or in small batches).

```
Start: 10 pods running v1.0
Step 1: Kill 1 v1.0 pod, start 1 v2.0 pod → 9×v1 + 1×v2
Step 2: Kill 1 v1.0 pod, start 1 v2.0 pod → 8×v1 + 2×v2
...
Step 10: All v2.0 → 10×v2

Kubernetes default strategy:
  maxSurge: 1 (can have 1 extra pod during rollout)
  maxUnavailable: 0 (never reduce below desired count)

Rollback:
  kubectl rollout undo deployment/api
  Kubernetes rolls back to previous version using same rolling strategy

Pros:
  ✅ No extra infrastructure cost
  ✅ Built into Kubernetes
  ✅ Gradual rollout

Cons:
  ❌ Rollback is slow (same rolling process in reverse)
  ❌ Two versions running simultaneously
  ❌ No traffic control (cannot send 1% to new version)
```

---

### When to Use Which

| Strategy | Use When | Avoid When |
|---|---|---|
| Blue-Green | Need instant rollback, have budget for 2x infra | Database migrations are complex |
| Canary | Want real traffic validation, risk-averse | Need fast rollout |
| Rolling | Cost-sensitive, Kubernetes default | Need instant rollback |

**Best practice: Canary for most deployments, Blue-Green for major releases.**

---

### Key Takeaways

1. **Zero-downtime deployment requires at least 2 versions running simultaneously** — design your API to be backward compatible during transitions.
2. **Database migrations must precede application deployments** — deploy schema changes first (backward compatible), then deploy new application code.
3. **Canary is the safest strategy** — real traffic, limited blast radius, gradual confidence.
4. **Rollback must be faster than the incident** — if rollback takes 30 minutes, you've already had a 30-minute outage.
5. **Monitor the right metrics during deployment** — error rate, latency, business metrics (orders placed, payments processed). Not just CPU/memory.

**Interview Follow-up Questions:**
- "How do you handle database migrations with blue-green deployments?"
- "What metrics do you monitor during a canary deployment to decide whether to proceed?"
- "How do you implement canary deployments with Istio/service mesh?"
- "What is a feature flag and how does it differ from a canary deployment?"

---

## Q6: Feature Flags for Safe Rollouts {#r6}

**Situation:**
Your team built a new checkout flow that took 3 months to develop. You want to release it to users but are afraid of bugs. You also want to A/B test it against the old flow. And you need the ability to instantly disable it if something goes wrong — without a deployment. Feature flags solve all three problems.

**What are Feature Flags:**

Feature flags (also called feature toggles or feature switches) are configuration values that control which code paths execute. They decouple **deployment** (code goes to production) from **release** (users see the feature).

```go
// Without feature flags: deploy = release
func handleCheckout(w http.ResponseWriter, r *http.Request) {
    // New checkout flow — deployed to all users immediately
    newCheckoutFlow(w, r)
}

// With feature flags: deploy ≠ release
func handleCheckout(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r)
    
    if featureFlags.IsEnabled("new-checkout-flow", userID) {
        newCheckoutFlow(w, r)  // Only for enabled users
    } else {
        oldCheckoutFlow(w, r)  // Everyone else
    }
}
```

---

### Feature Flag Types

**1. Release Toggles** — Control rollout of new features
```
new-checkout-flow: enabled for 10% of users
→ Gradually increase to 100% as confidence grows
→ Disable instantly if bugs found (no deployment needed)
```

**2. Experiment Toggles** — A/B testing
```
checkout-button-color: 
  Group A (50%): blue button
  Group B (50%): green button
→ Measure conversion rate per group
→ Ship winner to 100%
```

**3. Ops Toggles** — Emergency kill switches
```
enable-recommendations: true
→ If recommendations service is down: set to false
→ Checkout still works, just without recommendations
→ No deployment needed
```

**4. Permission Toggles** — Role-based features
```
advanced-analytics: enabled for enterprise tier only
beta-features: enabled for beta users only
```

---

### Implementation

```go
type FeatureFlagService struct {
    store  FlagStore  // Redis or database
    cache  *sync.Map  // Local cache for performance
}

type FlagConfig struct {
    Name        string
    Enabled     bool
    RolloutPct  int       // 0-100: percentage of users
    AllowList   []string  // Specific user IDs always enabled
    DenyList    []string  // Specific user IDs always disabled
}

func (f *FeatureFlagService) IsEnabled(flagName string, userID string) bool {
    config := f.getConfig(flagName)
    
    // Check deny list first
    if contains(config.DenyList, userID) {
        return false
    }
    
    // Check allow list
    if contains(config.AllowList, userID) {
        return true
    }
    
    // Check if globally disabled
    if !config.Enabled {
        return false
    }
    
    // Percentage rollout: deterministic based on user ID
    // Same user always gets same result (consistent experience)
    hash := fnv32(userID + flagName)
    return (hash % 100) < uint32(config.RolloutPct)
}

// Deterministic hash: same user always in same bucket
func fnv32(s string) uint32 {
    h := fnv.New32a()
    h.Write([]byte(s))
    return h.Sum32()
}
```

---

### Feature Flag Lifecycle

```
1. Create flag (disabled by default)
   new-checkout-flow: { enabled: false, rollout: 0% }

2. Enable for internal users (dogfooding)
   new-checkout-flow: { enabled: true, allowList: [employee_ids], rollout: 0% }

3. Enable for beta users
   new-checkout-flow: { enabled: true, rollout: 5% }

4. Gradual rollout
   new-checkout-flow: { enabled: true, rollout: 10% }
   new-checkout-flow: { enabled: true, rollout: 25% }
   new-checkout-flow: { enabled: true, rollout: 50% }
   new-checkout-flow: { enabled: true, rollout: 100% }

5. Remove flag (clean up technical debt)
   Delete flag, remove if/else from code
   Ship clean code with new flow only
```

**Flag cleanup is critical** — flags that are never removed become technical debt. Set a removal date when creating each flag.

---

### Key Takeaways

1. **Decouple deployment from release** — code can be in production for weeks before users see it. This reduces deployment risk dramatically.
2. **Kill switches are the most valuable flags** — the ability to disable a feature in 30 seconds without a deployment is worth the complexity.
3. **Percentage rollout must be deterministic** — use `hash(userID + flagName) % 100` so the same user always gets the same experience.
4. **Clean up flags** — every flag is technical debt. Set a removal date. Flags that live forever create unmaintainable code.
5. **Feature flags are not a substitute for testing** — they reduce blast radius but don't replace unit tests, integration tests, and canary deployments.

**Interview Follow-up Questions:**
- "How do you handle feature flags in a microservices environment where multiple services need to check the same flag?"
- "What is the difference between a feature flag and a configuration value?"
- "How do you prevent flag proliferation (too many flags making code unreadable)?"
- "What tools do you use for feature flag management? (LaunchDarkly, Unleash, Flagsmith)"

---

## Q7: OAuth2, JWT, and API Security at Scale {#r7}

**Situation:**
Your platform has a public API used by 10,000 third-party developers. You need to authenticate API requests, authorize access to specific resources, support multiple authentication methods (API keys, OAuth2 for user-delegated access), and handle token revocation. You also need to protect against common attacks: credential stuffing, token theft, replay attacks.

**Authentication vs Authorization:**
```
Authentication: Who are you? (Identity)
  - API key: "I am developer app X"
  - OAuth2 token: "I am user Y, acting through app X"
  - JWT: "I am user Y with these claims"

Authorization: What can you do? (Permissions)
  - Scopes: "This token can read orders but not write"
  - RBAC: "This user has the 'admin' role"
  - ABAC: "This user can access resources they own"
```

---

### OAuth2 Flow (Authorization Code)

Used when a third-party app needs to act on behalf of a user:

```
1. User clicks "Connect with MyApp" on third-party site
2. Third-party redirects to your authorization server:
   GET /oauth/authorize?
     client_id=app123&
     redirect_uri=https://thirdparty.com/callback&
     scope=read:orders write:profile&
     state=random_csrf_token&
     response_type=code

3. User logs in and approves permissions
4. Your server redirects back with authorization code:
   GET https://thirdparty.com/callback?code=AUTH_CODE&state=random_csrf_token

5. Third-party exchanges code for tokens (server-to-server):
   POST /oauth/token
   { grant_type: "authorization_code", code: AUTH_CODE, client_secret: SECRET }
   
   Response: {
     access_token: "eyJhbGc...",  // Short-lived (1 hour)
     refresh_token: "dGhpcyBp...", // Long-lived (30 days)
     expires_in: 3600,
     scope: "read:orders write:profile"
   }

6. Third-party uses access_token for API calls:
   GET /api/orders
   Authorization: Bearer eyJhbGc...

7. When access_token expires, use refresh_token to get new one:
   POST /oauth/token
   { grant_type: "refresh_token", refresh_token: "dGhpcyBp..." }
```

---

### JWT Structure and Validation

```
JWT = base64url(header) + "." + base64url(payload) + "." + signature

Header: { "alg": "RS256", "typ": "JWT" }
Payload: {
  "sub": "user_123",           // Subject (user ID)
  "iss": "https://auth.example.com",  // Issuer
  "aud": "https://api.example.com",   // Audience
  "exp": 1699000000,           // Expiry (Unix timestamp)
  "iat": 1698996400,           // Issued at
  "scope": "read:orders",      // Permissions
  "jti": "unique-token-id"     // JWT ID (for revocation)
}
Signature: RS256(base64url(header) + "." + base64url(payload), private_key)
```

**Validation steps (every API request):**
```go
func validateJWT(tokenString string) (*Claims, error) {
    // 1. Parse and verify signature
    token, err := jwt.ParseWithClaims(tokenString, &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            // Verify algorithm is what we expect
            if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return publicKey, nil  // RS256 public key
        })
    
    claims := token.Claims.(*Claims)
    
    // 2. Verify expiry
    if claims.ExpiresAt.Before(time.Now()) {
        return nil, ErrTokenExpired
    }
    
    // 3. Verify issuer
    if claims.Issuer != "https://auth.example.com" {
        return nil, ErrInvalidIssuer
    }
    
    // 4. Verify audience
    if !claims.VerifyAudience("https://api.example.com", true) {
        return nil, ErrInvalidAudience
    }
    
    // 5. Check revocation list (if needed)
    if tokenRevocationCache.IsRevoked(claims.ID) {
        return nil, ErrTokenRevoked
    }
    
    return claims, nil
}
```

---

### Token Revocation Problem

JWT is stateless — the server doesn't store tokens. This means you cannot "delete" a token. If a token is stolen, it's valid until expiry.

**Solutions:**

**1. Short expiry + refresh tokens**
```
Access token: 15 minutes expiry
Refresh token: 30 days expiry, stored in DB

If access token is stolen: attacker has 15 minutes
If refresh token is stolen: revoke it in DB immediately
```

**2. Token revocation list (Redis)**
```go
// On logout or token compromise:
redis.SETEX("revoked_token:"+jti, tokenTTL, "1")

// On every request:
if redis.EXISTS("revoked_token:" + claims.JTI) {
    return ErrTokenRevoked
}
// Adds ~0.5ms per request (Redis lookup)
```

**3. Opaque tokens (reference tokens)**
```
Instead of JWT, issue a random string: "tok_abc123xyz"
Store token → user mapping in Redis
On every request: look up token in Redis

Pros: Instant revocation (delete from Redis)
Cons: Every request hits Redis (stateful)
```

---

### Key Takeaways

1. **Use short-lived access tokens (15-60 min) + long-lived refresh tokens** — limits damage from token theft.
2. **RS256 over HS256** — asymmetric signing means only your auth server can issue tokens, but any service can verify them (using public key).
3. **Always validate `iss`, `aud`, `exp`** — many JWT vulnerabilities come from skipping these checks.
4. **Never put sensitive data in JWT payload** — it's base64 encoded, not encrypted. Anyone can decode it.
5. **Scope-based authorization** — tokens should carry minimal permissions. `read:orders` cannot write orders.
6. **Rotate secrets regularly** — API keys and signing keys should be rotated. Use key IDs (`kid`) in JWT header to support multiple active keys during rotation.

**Interview Follow-up Questions:**
- "What is the difference between OAuth2 and OpenID Connect (OIDC)?"
- "How do you implement token refresh without requiring the user to log in again?"
- "What is PKCE and why is it needed for mobile/SPA OAuth2 flows?"
- "How do you handle JWT key rotation without downtime?"

---

## Q8: Zero-Trust Networking and mTLS {#r8}

**Situation:**
Your microservices run inside a VPC. The security assumption is "trust everything inside the network perimeter." A developer accidentally deploys a misconfigured service that has no authentication. Because it's inside the VPC, any other service can call it freely. An attacker who compromises one internal service can now call any other service without restriction. You need to move to a zero-trust model where every service-to-service call is authenticated and authorized.

**What is Zero-Trust:**

Zero-trust is the security model: **"Never trust, always verify."** It rejects the idea of a trusted internal network. Every request — even from inside the VPC — must be authenticated and authorized.

```
Old model (perimeter security):
  Outside VPC: Untrusted (firewall blocks)
  Inside VPC:  Trusted (anything goes)
  
  Problem: One compromised service = attacker has free access to everything

Zero-trust model:
  Outside VPC: Untrusted
  Inside VPC:  Also untrusted by default
  Every service call: Must present valid identity + be authorized
```

---

### mTLS (Mutual TLS)

Standard TLS: Client verifies server's certificate. Server doesn't verify client.
mTLS: **Both** client and server verify each other's certificates.

```
Standard TLS:
  Client → Server: "Show me your certificate"
  Server → Client: "Here's my cert (signed by trusted CA)"
  Client: Verifies cert → Establishes encrypted connection

mTLS:
  Client → Server: "Show me your certificate"
  Server → Client: "Here's my cert. Show me yours."
  Client → Server: "Here's my cert (signed by internal CA)"
  Both: Verify each other's certs → Establish encrypted connection

Result:
  - Encryption: All traffic encrypted (no eavesdropping)
  - Authentication: Both sides prove identity via certificate
  - No passwords or API keys needed between services
```

---

### Service Mesh Implementation (Istio)

Implementing mTLS manually in every service is impractical. A service mesh (Istio/Linkerd) handles it automatically via sidecar proxies:

```yaml
# Istio: Enable mTLS for entire namespace
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: production
spec:
  mtls:
    mode: STRICT  # Reject any non-mTLS traffic

---
# Authorization policy: Order Service can only be called by API Gateway
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: order-service-policy
  namespace: production
spec:
  selector:
    matchLabels:
      app: order-service
  rules:
  - from:
    - source:
        principals:
          - "cluster.local/ns/production/sa/api-gateway"
          # Only API Gateway's service account can call Order Service
  - to:
    - operation:
        methods: ["GET", "POST"]
        paths: ["/api/orders*"]
```

**What Istio does automatically:**
- Issues certificates to every pod (via SPIFFE/SPIRE)
- Rotates certificates every 24 hours
- Enforces mTLS between all services
- Logs all service-to-service calls
- Applies authorization policies

---

### SPIFFE: Service Identity

SPIFFE (Secure Production Identity Framework for Everyone) provides a standard way to identify services:

```
SPIFFE ID format: spiffe://{trust-domain}/{path}
Example: spiffe://example.com/ns/production/sa/order-service

This ID is embedded in the service's X.509 certificate.
When Order Service calls Payment Service:
  - Order Service presents cert with SPIFFE ID
  - Payment Service verifies: "Is this caller allowed?"
  - Authorization policy: "Only order-service can call /api/payments"
```

---

### Network Policies (Kubernetes)

Even with mTLS, add network-level restrictions as defense in depth:

```yaml
# Only allow Order Service to receive traffic from API Gateway and Payment Service
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: order-service-ingress
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: order-service
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: api-gateway
    - podSelector:
        matchLabels:
          app: payment-service
    ports:
    - protocol: TCP
      port: 8080
```

---

### Key Takeaways

1. **Perimeter security is dead** — cloud environments, remote work, and microservices make the "trusted internal network" assumption invalid.
2. **mTLS provides mutual authentication** — both caller and callee prove their identity. No passwords, no API keys between services.
3. **Service mesh makes zero-trust practical** — implementing mTLS in every service manually is infeasible. Istio/Linkerd handles it transparently.
4. **Least privilege for service-to-service calls** — Order Service should only be able to call the specific endpoints it needs, not every endpoint of every service.
5. **Defense in depth** — mTLS + network policies + authorization policies. Multiple layers, each independent.

**Interview Follow-up Questions:**
- "What is the difference between mTLS and regular TLS?"
- "How does a service mesh differ from an API gateway?"
- "What is SPIFFE/SPIRE and how does it relate to zero-trust?"
- "How do you handle certificate rotation without downtime?"

---

## Q9: Secrets Management {#r9}

**Situation:**
A security audit reveals that your application has database passwords, API keys, and TLS certificates hardcoded in environment variables set in Kubernetes manifests — which are stored in Git. A developer accidentally pushed a manifest with a production database password to a public GitHub repo. The password was exposed for 3 hours before anyone noticed. You need a proper secrets management strategy.

**The Problem with Common Approaches:**

```
❌ Hardcoded in code:
   const dbPassword = "super_secret_123"
   → In version control forever, even after deletion

❌ Environment variables in Kubernetes manifests:
   env:
   - name: DB_PASSWORD
     value: "super_secret_123"
   → Manifest stored in Git → exposed

❌ Kubernetes Secrets (base64 only):
   apiVersion: v1
   kind: Secret
   data:
     password: c3VwZXJfc2VjcmV0XzEyMw==  # base64, not encrypted!
   → Anyone with kubectl access can decode instantly
   → Stored unencrypted in etcd by default

✅ Proper secrets management:
   - Secrets stored in dedicated vault (HashiCorp Vault, AWS Secrets Manager)
   - Encrypted at rest and in transit
   - Access controlled by policy (only specific services can read specific secrets)
   - Automatic rotation
   - Audit log of every access
```

---

### HashiCorp Vault

Vault is the industry standard for secrets management:

```
Vault capabilities:
  - Store secrets (key-value, database credentials, TLS certs)
  - Dynamic secrets: generate short-lived DB credentials on demand
  - Secret rotation: automatically rotate passwords
  - Audit log: every read/write logged
  - Fine-grained policies: service A can read secret X, not secret Y
  - Multiple auth methods: Kubernetes SA, AWS IAM, LDAP

Dynamic database credentials (most powerful feature):
  Instead of: "DB password is 'abc123', never changes"
  Vault does: "Generate a new DB user with 1-hour TTL for this service"
  
  Service requests credentials:
    vault read database/creds/order-service
    → { username: "v-order-svc-abc123", password: "A1b2C3d4", ttl: "1h" }
  
  After 1 hour: Vault revokes the user from PostgreSQL
  If credentials are stolen: they expire in max 1 hour
  No long-lived passwords anywhere
```

---

### Kubernetes + Vault Integration

```yaml
# Vault Agent Sidecar: injects secrets into pod as files
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-service
spec:
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/role: "order-service"
        vault.hashicorp.com/agent-inject-secret-db: "database/creds/order-service"
        vault.hashicorp.com/agent-inject-template-db: |
          {{- with secret "database/creds/order-service" -}}
          DB_USER={{ .Data.username }}
          DB_PASS={{ .Data.password }}
          {{- end }}
    spec:
      serviceAccountName: order-service  # Vault authenticates via K8s SA
      containers:
      - name: order-service
        # Secret available at /vault/secrets/db as a file
        # Application reads file, not env var
        # Vault agent refreshes file before TTL expires
```

**Application reads secret from file:**
```go
func getDatabaseCredentials() (string, string) {
    // Read from file injected by Vault agent
    data, _ := os.ReadFile("/vault/secrets/db")
    // Parse DB_USER and DB_PASS from file
    // Vault agent refreshes this file before TTL expires
    // Application re-reads on each connection (or watches file)
    return parseCredentials(data)
}
```

---

### AWS Secrets Manager

For AWS-native environments:

```go
import "github.com/aws/aws-sdk-go-v2/service/secretsmanager"

func getSecret(secretName string) (string, error) {
    client := secretsmanager.NewFromConfig(cfg)
    
    result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    })
    
    return *result.SecretString, err
}

// Cache secrets in memory with TTL (avoid calling Secrets Manager on every request)
var secretCache = ttlcache.New(5 * time.Minute)

func getCachedSecret(name string) string {
    if val, ok := secretCache.Get(name); ok {
        return val.(string)
    }
    secret, _ := getSecret(name)
    secretCache.Set(name, secret, 5*time.Minute)
    return secret
}
```

**Automatic rotation:**
```
AWS Secrets Manager can automatically rotate:
  - RDS passwords (built-in Lambda rotation function)
  - API keys (custom Lambda)
  - TLS certificates (via ACM)

Rotation process:
  1. Generate new password
  2. Update in database
  3. Update in Secrets Manager
  4. Application fetches new password on next cache miss
  5. Old password still works for 1 hour (grace period)
  6. Old password revoked
```

---

### Secret Scanning in CI/CD

Prevent secrets from entering Git:

```yaml
# .github/workflows/secret-scan.yml
name: Secret Scanning
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
      with:
        fetch-depth: 0  # Full history for scanning
    - name: Scan for secrets
      uses: trufflesecurity/trufflehog@main
      with:
        path: ./
        base: ${{ github.event.repository.default_branch }}
        # Fails CI if any secrets detected
```

---

### Key Takeaways

1. **Secrets in Git = secrets compromised** — treat any secret that has ever been in Git as compromised. Rotate immediately.
2. **Dynamic secrets are the gold standard** — short-lived, auto-rotated credentials limit the blast radius of any compromise.
3. **Never use environment variables for secrets in production** — they appear in process listings, crash dumps, and logs.
4. **Audit every secret access** — who read what secret, when. This is essential for incident response.
5. **Rotate secrets regularly** — even if not compromised. Rotation limits the window of exposure.
6. **Least privilege for secret access** — Order Service should only be able to read its own database credentials, not Payment Service's.

**Interview Follow-up Questions:**
- "How do you handle secret rotation without application downtime?"
- "What is the difference between HashiCorp Vault and AWS Secrets Manager?"
- "How do you prevent secrets from being logged accidentally?"
- "What do you do when a secret is accidentally committed to Git?"

---

## Q10: DDoS Protection and WAF Architecture {#r10}

**Situation:**
Your e-commerce platform is hit by a DDoS attack during Black Friday. 50,000 bots are sending 2 million requests/second to your checkout endpoint. Your origin servers are overwhelmed. Legitimate customers cannot complete purchases. You're losing $10,000/minute. You need a multi-layer DDoS protection strategy.

**DDoS Attack Types:**

```
Layer 3/4 (Network/Transport) — Volumetric attacks:
  - UDP flood: Send massive UDP packets to exhaust bandwidth
  - SYN flood: Send TCP SYN packets without completing handshake
  - ICMP flood: Ping flood
  Mitigation: Upstream scrubbing (ISP/CDN level), anycast routing

Layer 7 (Application) — Sophisticated attacks:
  - HTTP flood: Legitimate-looking HTTP requests at massive scale
  - Slowloris: Open many connections, send data very slowly
  - API abuse: Target expensive endpoints (search, checkout)
  Mitigation: WAF, rate limiting, bot detection, CAPTCHA
```

---

### Defense-in-Depth Architecture

```
Attack traffic
     │
     ▼
┌─────────────────────────────────────────────────────────┐
│  Layer 1: Anycast Network + Upstream Scrubbing          │
│  (Cloudflare, AWS Shield Advanced, Akamai)              │
│  - Absorbs volumetric attacks (Tbps capacity)           │
│  - Anycast: attack traffic distributed across PoPs      │
│  - Scrubbing: filters malicious traffic before CDN      │
└──────────────────────────┬──────────────────────────────┘
                           │ Clean traffic only
                           ▼
┌─────────────────────────────────────────────────────────┐
│  Layer 2: CDN Edge (Cloudflare, CloudFront)             │
│  - Serves cached content (no origin hit)                │
│  - Rate limiting per IP (1000 req/min)                  │
│  - Bot detection (JS challenge, fingerprinting)         │
│  - Geo-blocking (block countries you don't serve)       │
└──────────────────────────┬──────────────────────────────┘
                           │ Legitimate traffic only
                           ▼
┌─────────────────────────────────────────────────────────┐
│  Layer 3: WAF (Web Application Firewall)                │
│  - OWASP Top 10 protection (SQLi, XSS, CSRF)            │
│  - Custom rules (block known bad IPs, user agents)      │
│  - Rate limiting per user/API key                       │
│  - Request size limits                                  │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│  Layer 4: API Gateway                                   │
│  - Authentication (reject unauthenticated requests)     │
│  - Rate limiting (per authenticated user)               │
│  - Request validation                                   │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
                    Origin Servers
```

---

### WAF Rules

```
# Cloudflare WAF rule examples

# Block requests with no User-Agent (bots often omit this)
(not http.user_agent contains "Mozilla" and
 not http.user_agent contains "curl" and
 not http.user_agent contains "python")
→ Block

# Rate limit checkout endpoint
(http.request.uri.path eq "/api/checkout")
→ Rate limit: 10 requests per minute per IP

# Block known bad IPs (threat intelligence feed)
(ip.src in $threat_intelligence_list)
→ Block

# Challenge suspicious traffic (JS challenge)
(cf.threat_score gt 10)
→ JS Challenge (bots fail, humans pass)

# Block SQL injection attempts
(http.request.uri.query contains "' OR 1=1" or
 http.request.body contains "UNION SELECT")
→ Block + Log
```

---

### Bot Detection

```
Signals that indicate bot traffic:
  - No JavaScript execution (headless browsers)
  - Missing browser fingerprint (canvas, WebGL)
  - Abnormal request timing (too fast, too regular)
  - No mouse movement / keyboard events
  - Known bot IP ranges (datacenter IPs, Tor exit nodes)
  - Missing cookies (bots often don't store cookies)
  - Abnormal Accept-Language, Accept-Encoding headers

Progressive challenges:
  Suspicious score 0-10: Allow (normal)
  Suspicious score 10-30: JS challenge (invisible to humans)
  Suspicious score 30-60: CAPTCHA
  Suspicious score 60+: Block
```

---

### Emergency Response Playbook

```
DDoS detected (traffic spike + error rate spike):

Immediate (0-5 minutes):
  1. Enable "Under Attack Mode" in Cloudflare
     → All visitors get JS challenge
     → Bots blocked, humans pass after 5 seconds
  2. Enable rate limiting: 100 req/min per IP globally
  3. Block top attacking IPs (from access logs)

Short-term (5-30 minutes):
  4. Analyze attack pattern (which endpoints, which IPs, which countries)
  5. Add targeted WAF rules
  6. Enable geo-blocking if attack is from specific regions
  7. Scale up origin servers (auto-scaling)

Medium-term (30+ minutes):
  8. Contact upstream provider (Cloudflare, AWS Shield) for L3/L4 mitigation
  9. Implement CAPTCHA on targeted endpoints
  10. Analyze and block attack signatures
```

---

### Key Takeaways

1. **DDoS protection must be at the edge** — your origin servers cannot absorb a 2M req/sec attack. CDN/anycast absorbs it before it reaches you.
2. **Layer 7 attacks are harder than Layer 3/4** — volumetric attacks are easy to detect (just big). Application-layer attacks look like legitimate traffic.
3. **Rate limiting is your first application-layer defense** — limit requests per IP, per user, per endpoint.
4. **Bot detection is probabilistic** — no perfect solution. Use multiple signals (JS challenge, fingerprinting, behavioral analysis).
5. **Have a playbook before the attack** — during an attack is the wrong time to figure out how to enable Cloudflare's attack mode.
6. **AWS Shield Advanced vs Standard** — Standard is free, protects against common L3/L4 attacks. Advanced ($3K/month) adds L7 protection, 24/7 DDoS response team, cost protection.

**Interview Follow-up Questions:**
- "What is the difference between a DDoS attack and a DoS attack?"
- "How does anycast routing help with DDoS mitigation?"
- "What is the difference between a WAF and a firewall?"
- "How do you protect against Slowloris attacks?"

---

## Q11: REST vs GraphQL vs gRPC — When to Use Each {#r11}

**Situation:**
You are designing the API layer for a new platform. The mobile team complains that REST APIs return too much data (over-fetching) and require too many round trips (under-fetching). The internal microservices team wants high-performance binary communication. The third-party developer ecosystem needs a stable, well-documented API. You need to choose the right API style for each use case.

---

### REST (Representational State Transfer)

**What it is:** HTTP-based API using standard methods (GET, POST, PUT, DELETE) and URLs as resource identifiers.

```
GET    /api/users/123          → Get user
POST   /api/users              → Create user
PUT    /api/users/123          → Update user
DELETE /api/users/123          → Delete user
GET    /api/users/123/orders   → Get user's orders
```

**Strengths:**
- Universal: every language, framework, tool supports HTTP
- Cacheable: GET responses can be cached by CDN, browser
- Simple: easy to understand, debug with curl
- Stateless: each request is independent
- Well-understood: REST is the default for public APIs

**Weaknesses:**
- Over-fetching: endpoint returns all fields, client needs 3
- Under-fetching: need user + orders + addresses = 3 round trips
- No real-time: polling or WebSocket needed for live updates
- Versioning: breaking changes require new version (/v2/)

**Use REST when:**
- Public API for third-party developers
- Simple CRUD operations
- Caching is important (CDN-cacheable responses)
- Team is familiar with REST
- Interoperability is critical

---

### GraphQL

**What it is:** Query language where clients specify exactly what data they need. Single endpoint, flexible queries.

```graphql
# Client specifies exactly what it needs — no over/under-fetching
query {
  user(id: "123") {
    name
    email
    orders(last: 5) {
      id
      total
      status
      items {
        productName
        quantity
      }
    }
  }
}

# Response contains exactly what was requested — nothing more
```

**Strengths:**
- No over-fetching: get exactly the fields you need
- No under-fetching: get related data in one request
- Strongly typed schema: self-documenting, enables tooling
- Introspection: clients can query the schema itself
- Subscriptions: real-time updates via WebSocket

**Weaknesses:**
- Complexity: harder to implement, debug, cache
- N+1 problem: naive resolvers cause N+1 queries (need DataLoader)
- No HTTP caching: POST requests are not cached by CDN
- Rate limiting: harder (one endpoint, variable cost queries)
- Learning curve: teams need to learn GraphQL

**Use GraphQL when:**
- Mobile clients with bandwidth constraints
- Complex, nested data requirements
- Multiple clients with different data needs (mobile vs web vs TV)
- Rapid product iteration (add fields without versioning)
- BFF (Backend for Frontend) pattern

---

### gRPC (Google Remote Procedure Call)

**What it is:** Binary protocol using Protocol Buffers (protobuf) for serialization. HTTP/2 transport. Strongly typed contracts.

```protobuf
// Define service in .proto file
service OrderService {
  rpc GetOrder (GetOrderRequest) returns (Order);
  rpc CreateOrder (CreateOrderRequest) returns (Order);
  rpc StreamOrders (StreamOrdersRequest) returns (stream Order);
}

message Order {
  string id = 1;
  string user_id = 2;
  float total = 3;
  OrderStatus status = 4;
  repeated OrderItem items = 5;
}
```

**Strengths:**
- Performance: binary serialization (10x smaller than JSON), HTTP/2 multiplexing
- Strongly typed: compile-time type checking, auto-generated client code
- Streaming: client streaming, server streaming, bidirectional streaming
- Code generation: generate clients in any language from .proto file
- Deadline propagation: timeouts propagate across service calls

**Weaknesses:**
- Not human-readable: binary format, hard to debug with curl
- Browser support: limited (gRPC-Web needed for browsers)
- Learning curve: protobuf syntax, code generation
- Not cacheable: binary POST requests
- Ecosystem: fewer tools than REST

**Use gRPC when:**
- Internal microservice-to-microservice communication
- High-performance requirements (low latency, high throughput)
- Polyglot environment (services in Go, Java, Python — all use same .proto)
- Streaming data (real-time updates, large file transfers)
- Strong typing is critical

---

### Comparison Table

| Aspect | REST | GraphQL | gRPC |
|---|---|---|---|
| Protocol | HTTP/1.1 | HTTP/1.1 | HTTP/2 |
| Format | JSON/XML | JSON | Protobuf (binary) |
| Typing | Weak (OpenAPI) | Strong (schema) | Strong (protobuf) |
| Caching | Excellent (CDN) | Poor (POST) | None |
| Performance | Good | Good | Excellent |
| Browser support | Native | Native | Via gRPC-Web |
| Streaming | No (WebSocket) | Yes (subscriptions) | Yes (native) |
| Learning curve | Low | Medium | Medium |
| Best for | Public APIs | Mobile/BFF | Internal services |

---

### Real-World Architecture: Use All Three

```
External (third-party developers):
  REST API → /api/v1/...
  - Stable, versioned, well-documented
  - CDN-cacheable GET responses
  - OpenAPI/Swagger documentation

Mobile/Web clients:
  GraphQL API → /graphql
  - Flexible queries, no over/under-fetching
  - Subscriptions for real-time features
  - BFF pattern (one GraphQL layer per client type)

Internal microservices:
  gRPC → service:50051
  - High performance, low latency
  - Strongly typed contracts
  - Streaming for real-time data
  - Auto-generated clients in any language
```

---

### Key Takeaways

1. **REST for public APIs** — universal support, cacheable, well-understood by third-party developers.
2. **GraphQL for client-driven data fetching** — eliminates over/under-fetching, ideal for mobile and complex UIs.
3. **gRPC for internal services** — 10x performance over REST/JSON, streaming, strong typing.
4. **You don't have to choose one** — most large platforms use all three for different purposes.
5. **GraphQL N+1 is a real problem** — always use DataLoader pattern with GraphQL to batch database queries.
6. **gRPC requires HTTP/2** — ensure your load balancers and proxies support HTTP/2 (not all do by default).

**Interview Follow-up Questions:**
- "How do you handle API versioning in GraphQL?"
- "What is the N+1 problem in GraphQL and how do you solve it?"
- "How do you implement authentication in gRPC?"
- "What is the BFF (Backend for Frontend) pattern and when do you use it?"

---

## Q12: API Versioning Strategies {#r12}

**Situation:**
Your public API has 5,000 third-party integrations. You need to change the response format of `GET /api/orders` — the `status` field currently returns `"PENDING"` but needs to return `"pending"` (lowercase) to match a new standard. If you change it directly, all 5,000 integrations break. You need a versioning strategy that lets you evolve the API without breaking existing clients.

**Why API Versioning is Hard:**

```
The API contract problem:
  - You publish an API
  - 5,000 developers build integrations
  - You need to change the API
  - You cannot break existing integrations
  - You cannot maintain old behavior forever
  
The tension: Evolution vs Stability
```

---

### Strategy 1: URL Path Versioning (Most Common)

```
/api/v1/orders  → Old behavior (status: "PENDING")
/api/v2/orders  → New behavior (status: "pending")

Pros:
  ✅ Explicit, visible in URL
  ✅ Easy to route at load balancer level
  ✅ Easy to document separately
  ✅ Can run v1 and v2 simultaneously

Cons:
  ❌ URL is not "pure REST" (version is not a resource)
  ❌ Clients must update URLs to migrate
  ❌ Code duplication between versions

Implementation:
  /api/v1/ → v1 handler (old behavior)
  /api/v2/ → v2 handler (new behavior)
  
  Deprecation timeline:
    v2 released → v1 deprecated (12-month notice)
    After 12 months → v1 returns 410 Gone
```

---

### Strategy 2: Header Versioning

```
GET /api/orders
Accept: application/vnd.myapi.v2+json

Or:
GET /api/orders
API-Version: 2

Pros:
  ✅ Clean URLs (no version in path)
  ✅ Follows HTTP content negotiation standard

Cons:
  ❌ Less visible (version hidden in header)
  ❌ Harder to test in browser
  ❌ CDN caching requires Vary header
  ❌ Less common — developers expect URL versioning
```

---

### Strategy 3: Query Parameter Versioning

```
GET /api/orders?version=2

Pros:
  ✅ Optional (can default to latest)
  ✅ Easy to test in browser

Cons:
  ❌ Pollutes query string
  ❌ Easy to forget
  ❌ Not RESTful
```

---

### Additive Changes (Non-Breaking — No Version Needed)

Not all changes require a new version. Additive changes are backward compatible:

```
✅ Non-breaking (no version bump needed):
  - Add new optional field to response
  - Add new optional query parameter
  - Add new endpoint
  - Add new enum value (if clients ignore unknown values)
  - Change error message text (not error code)

❌ Breaking (requires new version):
  - Remove a field from response
  - Rename a field
  - Change field type (string → integer)
  - Change field semantics (status: "PENDING" → "pending")
  - Remove an endpoint
  - Change required parameters
  - Change authentication method
```

---

### Sunset Policy

```
Version lifecycle:
  GA (Generally Available): Fully supported
  Deprecated: Still works, but scheduled for removal
  Sunset: Removed, returns 410 Gone

Headers to communicate deprecation:
  Deprecation: Sat, 01 Jan 2025 00:00:00 GMT
  Sunset: Sat, 01 Jan 2026 00:00:00 GMT
  Link: <https://docs.example.com/migration/v1-to-v2>; rel="deprecation"

Deprecation timeline:
  Minimum 12 months notice for breaking changes
  Email all registered developers
  Show deprecation warnings in API responses
  Track which clients are still using deprecated version
```

---

### Versioning in Practice (Go)

```go
// Router setup with versioned handlers
func setupRoutes(r *mux.Router) {
    v1 := r.PathPrefix("/api/v1").Subrouter()
    v1.HandleFunc("/orders", v1OrdersHandler).Methods("GET")
    
    v2 := r.PathPrefix("/api/v2").Subrouter()
    v2.HandleFunc("/orders", v2OrdersHandler).Methods("GET")
}

// v1 handler: old behavior
func v1OrdersHandler(w http.ResponseWriter, r *http.Request) {
    orders := getOrders()
    
    // Add deprecation headers
    w.Header().Set("Deprecation", "Sat, 01 Jan 2025 00:00:00 GMT")
    w.Header().Set("Sunset", "Sat, 01 Jan 2026 00:00:00 GMT")
    w.Header().Set("Link", `<https://docs.example.com/v2>; rel="successor-version"`)
    
    // Old response format
    json.NewEncoder(w).Encode(map[string]interface{}{
        "orders": transformToV1Format(orders),  // status: "PENDING"
    })
}

// v2 handler: new behavior
func v2OrdersHandler(w http.ResponseWriter, r *http.Request) {
    orders := getOrders()
    json.NewEncoder(w).Encode(map[string]interface{}{
        "orders": transformToV2Format(orders),  // status: "pending"
    })
}
```

---

### Key Takeaways

1. **URL path versioning is the industry standard** — `/api/v1/`, `/api/v2/`. It's explicit, visible, and easy to route.
2. **Additive changes don't need versioning** — add fields, add endpoints, add optional parameters. Only breaking changes need a new version.
3. **Give 12+ months deprecation notice** — developers need time to migrate. Shorter notice = angry developers = bad reputation.
4. **Track version usage** — log which version each API key uses. Don't sunset a version that 1,000 clients still use.
5. **Design for extensibility** — use `additionalProperties: true` in JSON Schema, ignore unknown fields in clients. This makes additive changes truly non-breaking.

**Interview Follow-up Questions:**
- "How do you handle versioning in GraphQL (which has no URL versioning)?"
- "What is the difference between a breaking change and a non-breaking change?"
- "How do you migrate clients from v1 to v2 without forcing them?"
- "How do you version internal APIs (microservice-to-microservice)?"

---

## Q13: Idempotency in Distributed Systems {#r13}

**Situation:**
A customer clicks "Pay Now" on your checkout page. The payment request is sent to your API. The network times out after 30 seconds — the client doesn't know if the payment succeeded or failed. The client retries. Now the payment is charged twice. The customer calls support furious. You need to make your payment API idempotent — retrying the same request must never charge twice.

**What is Idempotency:**

An operation is **idempotent** if performing it multiple times has the same effect as performing it once.

```
Idempotent:
  GET /api/orders/123     → Always returns same order (no side effects)
  DELETE /api/orders/123  → First call deletes, subsequent calls return 404 (same end state)
  PUT /api/orders/123 {status: "shipped"}  → Sets status to shipped, regardless of how many times called

NOT idempotent:
  POST /api/payments {amount: 100}  → Each call creates a new charge!
  POST /api/emails/send             → Each call sends another email!
```

---

### Idempotency Keys

The standard solution: client generates a unique key for each logical operation. Server uses this key to deduplicate.

```
Client generates UUID for this payment attempt:
  idempotency_key = "550e8400-e29b-41d4-a716-446655440000"

First attempt:
  POST /api/payments
  Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
  { amount: 100, card: "tok_visa" }
  
  Server: Key not seen before → Process payment → Store result → Return 200

Network timeout — client doesn't know if it succeeded.

Retry (same key):
  POST /api/payments
  Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
  { amount: 100, card: "tok_visa" }
  
  Server: Key already seen → Return stored result → Return 200 (same response)
  → No duplicate charge!
```

---

### Implementation

```go
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
    idempotencyKey := r.Header.Get("Idempotency-Key")
    if idempotencyKey == "" {
        http.Error(w, "Idempotency-Key header required", 400)
        return
    }
    
    // Check if we've seen this key before
    cacheKey := "idempotency:" + idempotencyKey
    
    // Try to get cached response
    if cached, err := redis.Get(ctx, cacheKey).Bytes(); err == nil {
        // Key seen before — return cached response
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Idempotent-Replayed", "true")
        w.Write(cached)
        return
    }
    
    // Key not seen — process the payment
    // Use distributed lock to prevent concurrent processing of same key
    lock, err := redislock.Obtain(ctx, redis, "lock:"+idempotencyKey, 30*time.Second, nil)
    if err != nil {
        // Another request is processing this key right now
        http.Error(w, "Concurrent request with same idempotency key", 409)
        return
    }
    defer lock.Release(ctx)
    
    // Double-check after acquiring lock (another goroutine may have processed it)
    if cached, err := redis.Get(ctx, cacheKey).Bytes(); err == nil {
        w.Write(cached)
        return
    }
    
    // Process payment
    var req PaymentRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    result, err := h.paymentService.Charge(ctx, req)
    
    // Build response
    var response []byte
    if err != nil {
        response, _ = json.Marshal(map[string]string{"error": err.Error()})
        w.WriteHeader(400)
    } else {
        response, _ = json.Marshal(result)
        w.WriteHeader(200)
    }
    
    // Cache the response (store for 24 hours)
    // Store even errors — retrying a failed payment should return same error
    redis.Set(ctx, cacheKey, response, 24*time.Hour)
    
    w.Header().Set("Content-Type", "application/json")
    w.Write(response)
}
```

---

### Database-Level Idempotency

For operations that must survive Redis failures:

```sql
-- Store idempotency keys in database
CREATE TABLE idempotency_keys (
    key         VARCHAR(255) PRIMARY KEY,
    response    JSONB NOT NULL,
    status_code INT NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    expires_at  TIMESTAMP DEFAULT NOW() + INTERVAL '24 hours'
);

-- In payment processing (same transaction as payment):
BEGIN;
  -- Insert idempotency key (fails if duplicate)
  INSERT INTO idempotency_keys (key, response, status_code)
  VALUES ($1, $2, $3)
  ON CONFLICT (key) DO NOTHING;
  
  -- Check if we inserted (new) or it already existed (duplicate)
  GET DIAGNOSTICS rows_affected = ROW_COUNT;
  
  IF rows_affected = 0 THEN
    -- Duplicate: return stored response
    SELECT response, status_code FROM idempotency_keys WHERE key = $1;
    ROLLBACK;
    RETURN;
  END IF;
  
  -- New request: process payment
  INSERT INTO payments ...;
COMMIT;
```

---

### Key Takeaways

1. **Any non-idempotent operation that can be retried needs idempotency keys** — payments, email sends, order creation, any POST that creates a resource.
2. **Client generates the key, server enforces uniqueness** — UUID v4 is the standard. Key should be unique per logical operation, not per HTTP request.
3. **Store idempotency results durably** — Redis is fast but volatile. For payments, store in the database in the same transaction.
4. **Idempotency key TTL** — 24 hours is standard. After that, the same key can be reused (new operation).
5. **Return the same response for duplicate requests** — including the same HTTP status code. A retried failed payment should return the same error, not process again.
6. **Stripe's approach** — Stripe requires `Idempotency-Key` header for all POST requests. This is the industry standard for payment APIs.

**Interview Follow-up Questions:**
- "How do you handle idempotency for operations that span multiple services?"
- "What is the difference between idempotency and at-least-once delivery?"
- "How do you implement idempotency for Kafka consumers?"
- "What happens if two requests with the same idempotency key arrive simultaneously?"

---

## Q14: Cloud Cost Optimization at Scale {#r14}

**Situation:**
Your AWS bill is $800,000/month and growing 30% month-over-month. The CTO asks you to reduce it by 40% without degrading performance. You have 500 EC2 instances, 50TB of S3 data, 20 RDS databases, and significant data transfer costs. You have no tagging strategy, so you don't know which team or service is responsible for which cost.

**The Hidden Cost Killers:**

```
Typical AWS cost breakdown for a $800K/month bill:
  EC2 compute:        $320K (40%) ← Biggest opportunity
  RDS databases:      $160K (20%)
  Data transfer:      $120K (15%) ← Often overlooked
  S3 storage:         $80K  (10%)
  ElastiCache/Redis:  $40K  (5%)
  Other services:     $80K  (10%)
```

---

### EC2 Cost Optimization

**1. Right-sizing (Biggest Impact)**

```
Problem: Developers provision large instances "just in case"
  m5.4xlarge (16 vCPU, 64GB RAM) at $0.768/hr = $550/month
  Actual usage: 15% CPU, 20% memory

Right-sized:
  m5.xlarge (4 vCPU, 16GB RAM) at $0.192/hr = $138/month
  Savings: $412/month per instance × 100 instances = $41,200/month

Tools:
  AWS Compute Optimizer: Analyzes CloudWatch metrics, recommends right size
  CloudWatch: CPU, memory, network utilization over 14 days
  
Rule of thumb: If average CPU < 20% and memory < 40%, downsize.
```

**2. Reserved Instances / Savings Plans**

```
On-Demand pricing: $0.192/hr for m5.xlarge
1-year Reserved (no upfront): $0.122/hr → 36% savings
3-year Reserved (partial upfront): $0.076/hr → 60% savings
Compute Savings Plan (flexible): 40-66% savings

Strategy:
  Baseline load (always running): Reserved Instances (1-3 year)
  Variable load: On-Demand or Spot
  
  Example: 100 instances always running
  On-Demand: 100 × $0.192 × 8760 hrs = $168,192/year
  1-yr Reserved: 100 × $0.122 × 8760 hrs = $106,872/year
  Savings: $61,320/year
```

**3. Spot Instances (Up to 90% Savings)**

```
Spot instances use spare AWS capacity at 60-90% discount.
Risk: AWS can reclaim with 2-minute notice.

Good for:
  ✅ Batch processing (video transcoding, ML training)
  ✅ Stateless web servers (behind load balancer)
  ✅ Kafka consumers (can restart from last offset)
  ✅ CI/CD workers

Bad for:
  ❌ Databases (stateful, cannot be interrupted)
  ❌ Long-running jobs without checkpointing
  ❌ Services requiring guaranteed availability

Spot savings example:
  m5.4xlarge On-Demand: $0.768/hr
  m5.4xlarge Spot: $0.115/hr (85% savings)
  100 instances × $0.653 savings × 8760 hrs = $572,028/year saved
```

---

### Data Transfer Cost Optimization

**The hidden killer — data transfer costs:**

```
AWS data transfer pricing:
  Within same AZ: FREE
  Between AZs (same region): $0.01/GB each way = $0.02/GB
  Internet egress: $0.09/GB (first 10TB/month)

Example: 100TB/month cross-AZ traffic
  100TB × $0.02/GB = $2,048/month = $24,576/year

Fixes:
  1. Deploy services in same AZ when possible
  2. Use VPC endpoints for S3/DynamoDB (free, no internet egress)
  3. Use CloudFront for egress (cheaper than direct S3 egress)
  4. Compress data before transfer
  5. Use S3 Transfer Acceleration only when needed
```

---

### Database Cost Optimization

```
RDS cost reduction strategies:

1. Aurora Serverless v2 for variable workloads:
   - Scales from 0.5 to 128 ACUs
   - Pay per ACU-hour (not for idle capacity)
   - Good for: dev/test, variable production workloads

2. Read replicas only when needed:
   - Each replica = same cost as primary
   - Use ElastiCache instead of read replicas for cacheable data
   - Remove replicas in non-production environments

3. Storage optimization:
   - RDS charges for provisioned storage, not used storage
   - Audit and reduce over-provisioned storage
   - Use S3 for cold data (100x cheaper than RDS storage)

4. Multi-AZ only for production:
   - Multi-AZ = 2x cost (standby replica)
   - Dev/test: Single-AZ (accept downtime risk)
   - Production: Multi-AZ (required for HA)
```

---

### S3 Storage Tiering

```
S3 storage classes (price per GB/month):
  Standard:           $0.023  ← Hot data (accessed frequently)
  Standard-IA:        $0.0125 ← Infrequent access (>30 days old)
  Glacier Instant:    $0.004  ← Archive (>90 days old, instant retrieval)
  Glacier Flexible:   $0.0036 ← Archive (>90 days old, 3-5 hour retrieval)
  Glacier Deep:       $0.00099 ← Long-term archive (>180 days, 12 hour retrieval)

S3 Intelligent-Tiering: Automatically moves objects between tiers
  Cost: $0.0025/1000 objects monitoring fee
  Savings: 40-68% for data with unknown access patterns

Lifecycle policy example:
  0-30 days:   Standard
  30-90 days:  Standard-IA (45% cheaper)
  90-365 days: Glacier Instant (83% cheaper)
  365+ days:   Glacier Deep Archive (96% cheaper)
```

---

### Tagging Strategy (Foundation for Cost Visibility)

```
Without tags: "We spend $800K/month on AWS" (useless)
With tags: "Team Checkout spends $120K/month, 40% on idle dev instances" (actionable)

Mandatory tags:
  Environment: production | staging | development
  Team: checkout | payments | platform | data
  Service: order-api | payment-service | user-service
  CostCenter: engineering | data-science | infrastructure

Enforce via AWS Config rule:
  Rule: required-tags
  Action: Alert if resource missing required tags
  
Cost allocation:
  AWS Cost Explorer → Group by tag → See cost per team/service
  Set budget alerts per team: "Alert when Team Checkout exceeds $150K/month"
```

---

### Key Takeaways

1. **Right-sizing is the biggest opportunity** — most teams over-provision by 2-4x. Use AWS Compute Optimizer to find candidates.
2. **Reserved Instances for baseline, Spot for variable** — commit to 1-year reserved for always-on workloads, use Spot for batch/stateless.
3. **Data transfer costs are invisible until they're huge** — audit cross-AZ and internet egress. Use VPC endpoints for AWS services.
4. **S3 lifecycle policies are free money** — data older than 90 days in Standard tier is waste. Automate tiering.
5. **Tagging is the foundation** — you cannot optimize what you cannot measure. Enforce tags from day one.
6. **FinOps is a discipline** — assign cost ownership to teams. Engineers who see their team's AWS bill make better architectural decisions.

**Interview Follow-up Questions:**
- "How do you handle Spot Instance interruptions gracefully?"
- "What is the difference between Reserved Instances and Savings Plans?"
- "How do you implement a FinOps culture in an engineering organization?"
- "What are the hidden costs in AWS that teams often miss?"

---

## Q15: Graceful Degradation Patterns {#r15}

**Situation:**
Your e-commerce platform has a Recommendations Service that suggests products. It calls an ML model that takes 500ms to respond. During a traffic spike, the Recommendations Service becomes slow. Your product pages wait 500ms for recommendations before rendering. Page load time goes from 200ms to 700ms. Conversion rate drops 20%. You need the product page to remain fast even when recommendations are unavailable.

**What is Graceful Degradation:**

Graceful degradation means the system continues to function — at reduced capability — when a component fails or is slow. The opposite of graceful degradation is **catastrophic failure**: one component's failure takes down the entire system.

```
Without graceful degradation:
  Product page → Recommendations Service (500ms or timeout)
  If slow: User waits 500ms
  If down: User gets error page
  
With graceful degradation:
  Product page → Recommendations Service (100ms timeout)
  If slow/down: Show "Popular Products" (cached fallback)
  User experience: Slightly less personalized, but fast and functional
```

---

### Pattern 1: Timeout + Fallback

```go
func getRecommendations(userID string, productID string) []Product {
    // Set aggressive timeout — don't let slow service hurt page load
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    recommendations, err := recommendationService.Get(ctx, userID, productID)
    
    if err != nil || ctx.Err() != nil {
        // Fallback: return popular products (cached, always fast)
        log.Printf("Recommendations unavailable, using fallback: %v", err)
        return getPopularProducts(productID)  // From Redis cache, <1ms
    }
    
    return recommendations
}

func getPopularProducts(productID string) []Product {
    // Pre-computed popular products, cached in Redis
    // Updated every hour by a background job
    cached, _ := redis.Get(ctx, "popular:"+productID).Bytes()
    var products []Product
    json.Unmarshal(cached, &products)
    return products
}
```

---

### Pattern 2: Stale Cache Fallback

Serve stale data rather than no data:

```go
type CacheWithFallback struct {
    redis *redis.Client
}

func (c *CacheWithFallback) GetWithFallback(key string, fetchFn func() (interface{}, error)) (interface{}, error) {
    // Try to get fresh data
    ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
    defer cancel()
    
    freshData, err := fetchFn()
    if err == nil {
        // Success: update cache and return fresh data
        data, _ := json.Marshal(freshData)
        c.redis.Set(ctx, key, data, 5*time.Minute)
        c.redis.Set(ctx, "stale:"+key, data, 24*time.Hour)  // Keep stale copy longer
        return freshData, nil
    }
    
    // Fetch failed: try stale cache
    staleData, staleErr := c.redis.Get(ctx, "stale:"+key).Bytes()
    if staleErr == nil {
        log.Printf("Serving stale data for key %s", key)
        var result interface{}
        json.Unmarshal(staleData, &result)
        return result, nil  // Return stale data, no error
    }
    
    // No stale data either: return error
    return nil, fmt.Errorf("service unavailable and no cached data")
}
```

---

### Pattern 3: Feature Shedding Under Load

Automatically disable non-critical features when the system is under stress:

```go
type LoadShedder struct {
    currentLoad float64  // 0.0 to 1.0
    mu          sync.RWMutex
}

func (ls *LoadShedder) ShouldEnable(feature string) bool {
    ls.mu.RLock()
    load := ls.currentLoad
    ls.mu.RUnlock()
    
    // Feature priority: higher threshold = disabled sooner under load
    thresholds := map[string]float64{
        "recommendations":    0.7,  // Disable when load > 70%
        "personalization":    0.8,  // Disable when load > 80%
        "related_products":   0.85, // Disable when load > 85%
        "recently_viewed":    0.9,  // Disable when load > 90%
        "core_checkout":      1.1,  // Never disable (threshold > 1.0)
    }
    
    threshold, exists := thresholds[feature]
    if !exists {
        return true  // Unknown features: enable by default
    }
    
    return load < threshold
}

// Update load every 10 seconds from metrics
func (ls *LoadShedder) UpdateLoad() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        cpuLoad := getCPUUtilization()
        queueDepth := getQueueDepth()
        errorRate := getErrorRate()
        
        // Composite load score
        load := (cpuLoad*0.4 + queueDepth*0.3 + errorRate*0.3)
        
        ls.mu.Lock()
        ls.currentLoad = load
        ls.mu.Unlock()
    }
}

// Usage in handler:
func productPageHandler(w http.ResponseWriter, r *http.Request) {
    product := getProduct(r)  // Always fetch (critical)
    
    var recommendations []Product
    if loadShedder.ShouldEnable("recommendations") {
        recommendations = getRecommendations(product.ID)
    } else {
        recommendations = []Product{}  // Empty, not an error
    }
    
    renderPage(w, product, recommendations)
}
```

---

### Pattern 4: Partial Response

Return what you have, indicate what's missing:

```json
// Full response (all services healthy):
{
  "product": { "id": "123", "name": "Widget", "price": 29.99 },
  "recommendations": [...],
  "reviews": [...],
  "inventory": { "in_stock": true, "quantity": 42 }
}

// Partial response (recommendations service down):
{
  "product": { "id": "123", "name": "Widget", "price": 29.99 },
  "recommendations": null,
  "reviews": [...],
  "inventory": { "in_stock": true, "quantity": 42 },
  "_partial": true,
  "_unavailable": ["recommendations"]
}
```

```go
type PageResponse struct {
    Product         *Product   `json:"product"`
    Recommendations []Product  `json:"recommendations"`
    Reviews         []Review   `json:"reviews"`
    Inventory       *Inventory `json:"inventory"`
    Partial         bool       `json:"_partial,omitempty"`
    Unavailable     []string   `json:"_unavailable,omitempty"`
}

func buildPageResponse(productID string) PageResponse {
    resp := PageResponse{}
    var unavailable []string
    
    // Fetch all components concurrently with individual timeouts
    var wg sync.WaitGroup
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
        defer cancel()
        product, err := productService.Get(ctx, productID)
        if err != nil {
            unavailable = append(unavailable, "product")
        } else {
            resp.Product = product
        }
    }()
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
        defer cancel()
        recs, err := recommendationService.Get(ctx, productID)
        if err != nil {
            unavailable = append(unavailable, "recommendations")
            resp.Recommendations = getPopularProducts(productID)  // Fallback
        } else {
            resp.Recommendations = recs
        }
    }()
    
    wg.Wait()
    
    if len(unavailable) > 0 {
        resp.Partial = true
        resp.Unavailable = unavailable
    }
    
    return resp
}
```

---

### Degradation Hierarchy

Define explicitly what degrades in what order:

```
Priority 1 (Never degrade — core business):
  - Product display
  - Add to cart
  - Checkout
  - Payment processing
  - Order confirmation

Priority 2 (Degrade gracefully — important but not critical):
  - Inventory count (show "In Stock" vs exact count)
  - Pricing (show cached price, flag as "may have changed")
  - Search (show popular results if search is slow)

Priority 3 (Disable under load — nice to have):
  - Personalized recommendations → Popular products
  - Recently viewed → Empty
  - Product reviews → "Reviews temporarily unavailable"
  - Related products → Empty

Priority 4 (Always disable first):
  - A/B test variants → Control group only
  - Analytics tracking → Drop events
  - Non-critical logging → Reduce verbosity
```

---

### Key Takeaways

1. **Define your degradation hierarchy before an incident** — know which features to disable first. Don't make these decisions during a crisis.
2. **Timeouts are mandatory for all external calls** — a service with no timeout will wait forever, blocking your threads.
3. **Fallbacks must be pre-computed** — a fallback that requires a database query is not a fallback. Cache popular products, trending items, default content.
4. **Partial responses are better than errors** — a product page with no recommendations is better than a 500 error.
5. **Load shedding protects the critical path** — automatically disable non-critical features when the system is under stress.
6. **Test your degraded state** — regularly verify that fallbacks work. A fallback that has never been tested will fail when you need it.

**Interview Follow-up Questions:**
- "How do you decide which features to degrade first?"
- "What is the difference between graceful degradation and fault tolerance?"
- "How do you test that your graceful degradation actually works?"
- "How do you communicate degraded state to users without alarming them?"

---

---

## Quick Reference: Reliability & Security Cheat Sheet

### SLO Targets by Service Type
```
Consumer API:      99.9% availability, P99 < 500ms
Internal API:      99.5% availability, P99 < 1000ms
Batch jobs:        99% success rate
Data pipelines:    99.5% freshness (< 1 hour old)
Payment systems:   99.99% availability, P99 < 200ms
```

### Deployment Strategy Decision Tree
```
Need instant rollback + have budget?  → Blue-Green
Need real traffic validation?         → Canary
Cost-sensitive, Kubernetes default?   → Rolling
New feature, want gradual exposure?   → Feature Flag
```

### API Style Decision Tree
```
Public API for third-party devs?      → REST
Mobile client, complex data needs?    → GraphQL
Internal microservice communication?  → gRPC
Real-time streaming?                  → gRPC or WebSocket
```

### Security Checklist
```
□ Secrets in Vault/Secrets Manager (not env vars or Git)
□ mTLS between all internal services
□ JWT: validate iss, aud, exp on every request
□ Short-lived access tokens (15-60 min) + refresh tokens
□ Rate limiting: per IP, per user, per endpoint
□ WAF in front of all public endpoints
□ DDoS protection at CDN edge (Cloudflare/Shield)
□ Secret scanning in CI/CD pipeline
□ Least privilege: service A cannot call service B's private endpoints
□ Audit logs for all secret access and privileged operations
```

### Reliability Checklist
```
□ SLOs defined for all production services
□ Error budgets tracked and reviewed monthly
□ Chaos experiments run quarterly
□ DR plan tested quarterly (not just documented)
□ All external calls have timeouts
□ Circuit breakers on all downstream dependencies
□ Graceful degradation for non-critical features
□ Idempotency keys on all non-idempotent POST endpoints
□ Zero-downtime deployment strategy in place
□ Runbooks for top 10 most common incidents
```

### Cost Optimization Quick Wins
```
1. Right-size EC2 instances (Compute Optimizer)
2. Reserved Instances for baseline load (36-60% savings)
3. Spot Instances for batch/stateless (60-90% savings)
4. S3 lifecycle policies (Standard → IA → Glacier)
5. VPC endpoints for S3/DynamoDB (eliminate data transfer cost)
6. Remove unused resources (orphaned EBS volumes, idle RDS)
7. Tag everything (visibility = optimization opportunity)
```

---

*Document covers: SLO/SLA/SLI, Chaos Engineering, Disaster Recovery, Zero-Downtime Migrations, Deployment Strategies, Feature Flags, OAuth2/JWT, Zero-Trust/mTLS, Secrets Management, DDoS/WAF, REST/GraphQL/gRPC, API Versioning, Idempotency, Cost Optimization, Graceful Degradation*
