# Rate Limiting

## What does a rate limiter limit?

Not just API requests. A rate limiter can be placed in front of anything that has a capacity constraint:

- API/HTTP requests (the most common use case)
- Database queries (protect your DB from being hammered)
- Message queue publish rates
- Login/authentication attempts (brute-force protection)
- Email/SMS sends (cost control + provider limits)
- File uploads or downloads
- Third-party API calls (respect their rate limits)
- CPU-intensive operations (like image processing, video encoding)

The core idea: anything that consumes a shared, finite resource can benefit from rate limiting.

## Why rate limit?

- Protect availability — prevent one noisy client from starving everyone else
- Prevent abuse — brute-force attacks, scraping, spam
- Cost control — cloud resources, third-party API calls, SMS/email all cost money
- Fair usage — ensure equitable access across users/tenants
- Stability — avoid cascading failures when load spikes hit
- Compliance — some services contractually require you to stay under certain thresholds

## Rate limiting algorithms

Here are the main ones, each with different tradeoffs:

### 1. Fixed Window Counter

Divide time into fixed windows (e.g., 1-minute intervals). Count requests per window. Reject when the count exceeds the limit.

```
Window: 12:00:00 - 12:00:59 → allow 100 requests
Window: 12:01:00 - 12:01:59 → counter resets
```

Pros: simple, low memory.
Cons: burst problem at window boundaries — a client can send 100 requests at 12:00:59 and 100 more at 12:01:00, effectively getting 200 in 2 seconds.

### 2. Sliding Window Log

Store the timestamp of every request. When a new request comes in, count how many timestamps fall within the last N seconds.

Pros: precise, no boundary burst issue.
Cons: memory-heavy — you're storing every timestamp.

### 3. Sliding Window Counter

A hybrid. Combine the current window's count with a weighted portion of the previous window's count based on how far into the current window you are.

```
effective_count = prev_window_count * overlap_percentage + current_window_count
```

Pros: good accuracy, low memory.
Cons: still an approximation, but a very good one.

### 4. Token Bucket

Imagine a bucket that holds tokens. Tokens are added at a fixed rate. Each request consumes a token. If the bucket is empty, the request is rejected.

```
bucket_capacity = 10
refill_rate = 1 token/second

Request arrives → token available? → allow and remove token
                  no token?       → reject
```

Pros: allows controlled bursts (up to bucket capacity), smooth rate over time.
Cons: slightly more state to manage.

This is what AWS API Gateway and many CDNs use.

### 5. Leaky Bucket

Requests enter a queue (the bucket). They're processed at a fixed rate, like water leaking from a hole. If the bucket is full, new requests are dropped.

Pros: perfectly smooth output rate.
Cons: no burst tolerance — even legitimate spikes get queued or dropped.

### 6. Concurrency Limiter (not time-based)

Instead of "X requests per second," limit "X requests in-flight at once." Once a request completes, a slot opens up.

Pros: adapts naturally to request duration.
Cons: doesn't control throughput directly.

## What happens when the limit is exceeded?

This is a design decision. There are several strategies:

| Strategy | Behavior |
|---|---|
| Reject (hard drop) | Return `429 Too Many Requests` immediately. Most common for APIs. |
| Queue / buffer | Hold excess requests and process them when capacity frees up. Good for background jobs. |
| Throttle / delay | Slow down the response instead of rejecting. The client gets a response, just later. |
| Degrade | Return a cheaper/cached response instead of doing the full computation. |
| Shed load selectively | Drop low-priority requests but keep high-priority ones. |

In practice, most public APIs reject with a 429 and include headers like:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 30
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1718400060
```

The request is not "killed" — the client gets a clear signal to back off.

## Where to enforce rate limiting

- At the edge / API gateway — before requests even hit your services (e.g., NGINX, AWS API Gateway, Cloudflare)
- In a reverse proxy or load balancer — centralized enforcement
- In application code — per-service or per-endpoint granularity
- At the database layer — connection pooling, query queuing
- Client-side — well-behaved clients self-throttle (exponential backoff)

Layering these is common. Edge handles the broad strokes, application code handles fine-grained per-user or per-endpoint limits.

## Distributed rate limiting

When you have multiple server instances, you need shared state. Options:

- Redis — the go-to. Atomic operations like `INCR` + `EXPIRE` make it natural for counters. Libraries like `redis-cell` implement token bucket natively.
- Memcached — similar, but less feature-rich.
- Sticky sessions — route the same client to the same server, so local state works. Fragile.
- Approximate local limits — each instance enforces `limit / num_instances`. Simple but imprecise.

## Core fundamentals

1. Rate limiting is about protecting resources, not punishing users.
2. The "right" algorithm depends on your traffic pattern — bursty traffic needs token bucket, smooth pipelines need leaky bucket.
3. Granularity matters — per-IP, per-user, per-API-key, per-endpoint, or global. Choose based on your threat model.
4. Always communicate limits to clients — use standard headers and clear error messages.
5. Rate limiting alone isn't security — it's one layer. Combine with authentication, authorization, and monitoring.
6. Test your limits under load — a rate limiter that breaks under pressure is worse than none at all.
