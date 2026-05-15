# System Design from Scratch — 10 Complete Walkthroughs
## Software Architect Interview Preparation

---

## Overview

This document contains **10 complete system design walkthroughs** for the most common interview questions. Each design follows the standard interview framework:

1. **Requirements Clarification** — Functional and non-functional requirements
2. **Capacity Estimation** — Traffic, storage, bandwidth calculations
3. **High-Level Design** — Core components and data flow
4. **Deep Dive** — Detailed architecture for critical components
5. **Trade-offs** — Design decisions and alternatives
6. **Bottlenecks & Solutions** — Scaling challenges and mitigations

**The 10 Systems:**
1. URL Shortener (bit.ly)
2. Twitter / X
3. Instagram
4. WhatsApp / Messaging System
5. YouTube / Video Streaming
6. Uber / Ride-Sharing
7. Netflix Recommendation System
8. Distributed Cache (Redis/Memcached)
9. Rate Limiter
10. Web Crawler

---

## Interview Framework (Use This for Every Design)

### Step 1: Requirements (5 minutes)
Ask clarifying questions. Don't assume.

**Functional Requirements:**
- What features must the system support?
- What are the core user flows?

**Non-Functional Requirements:**
- Scale: How many users? Requests per second?
- Latency: What's acceptable response time?
- Availability: 99.9%? 99.99%?
- Consistency: Strong or eventual?

### Step 2: Capacity Estimation (5 minutes)
Do back-of-the-envelope math. Show your thinking.

```
Users: 100M daily active users (DAU)
Requests: 100M users × 10 requests/day = 1B requests/day
QPS: 1B / 86,400 seconds ≈ 11,500 req/sec
Peak QPS: 11,500 × 3 (peak factor) = 35,000 req/sec
Storage: 100M users × 1KB/user × 365 days = 36.5TB/year
Bandwidth: 35,000 req/sec × 10KB/request = 350MB/sec
```

### Step 3: High-Level Design (10 minutes)
Draw boxes and arrows. Start simple.

```
Client → Load Balancer → API Servers → Database
                                     → Cache
```

### Step 4: Deep Dive (20 minutes)
Pick 2-3 components to go deep. Interviewer will guide.

Common deep dives:
- Database schema and indexing
- Caching strategy
- Sharding/partitioning
- Replication and consistency
- Message queues for async processing

### Step 5: Trade-offs (5 minutes)
Discuss alternatives and why you chose what you chose.

---

## Table of Contents

1. [Design 1: URL Shortener](#design1)
2. [Design 2: Twitter / X](#design2)
3. [Design 3: Instagram](#design3)
4. [Design 4: WhatsApp / Messaging](#design4)
5. [Design 5: YouTube / Video Streaming](#design5)
6. [Design 6: Uber / Ride-Sharing](#design6)
7. [Design 7: Netflix Recommendations](#design7)
8. [Design 8: Distributed Cache](#design8)
9. [Design 9: Rate Limiter](#design9)
10. [Design 10: Web Crawler](#design10)

---

## Design 1: URL Shortener (bit.ly) {#design1}

### Requirements Clarification

**Functional Requirements:**
- Given a long URL, generate a short URL (e.g., `bit.ly/abc123`)
- Redirect short URL to original long URL
- Custom aliases (optional): user can choose `bit.ly/my-brand`
- Expiry: URLs can expire after a set time
- Analytics: track click count, referrer, geography

**Non-Functional Requirements:**
- 100M URLs created per day
- 10:1 read/write ratio → 1B redirects per day
- Low latency: redirect in <10ms
- High availability: 99.99% (4 nines)
- URLs must be unique and not guessable

**Out of Scope:** User accounts, billing, team management

---

### Capacity Estimation

```
Write QPS:  100M / 86,400 = ~1,160 writes/sec
Read QPS:   1B / 86,400   = ~11,600 reads/sec (10x writes)
Peak reads: 11,600 × 3    = ~35,000 reads/sec

Storage per URL:
  - Short URL: 7 chars = 7 bytes
  - Long URL: avg 200 chars = 200 bytes
  - Metadata (created_at, expiry, user_id): 50 bytes
  - Total per record: ~257 bytes

Storage for 5 years:
  100M/day × 365 × 5 = 182.5B URLs
  182.5B × 257 bytes = ~47TB

Bandwidth:
  Read: 35,000 req/sec × 257 bytes = ~9MB/sec (tiny — mostly redirects)
  Write: 1,160 req/sec × 257 bytes = ~300KB/sec
```

---

### Core Algorithm: Short URL Generation

**Option 1 — MD5/SHA256 Hash (Don't Use)**
```
hash = MD5(long_url) → take first 7 chars
Problem: Collisions. Two different URLs can produce same hash prefix.
Problem: Same URL always produces same hash — cannot create multiple short URLs for same long URL.
```

**Option 2 — Base62 Encoding of Auto-Increment ID (Good)**
```
Database auto-increment ID: 1, 2, 3, ... 3,521,614,606
Base62 encode: 0-9, a-z, A-Z (62 characters)
ID 1        → "0000001"
ID 3.5B     → "5Feceb" (7 chars)
ID 3.5T     → "8M0kX5" (7 chars, handles trillions)

7 chars × 62 possibilities = 62^7 = 3.5 trillion unique URLs
At 100M/day: 3.5T / 100M = 35,000 days = 95 years of capacity
```

**Option 3 — Pre-generated Keys (Best for Scale)**
```
Key Generation Service (KGS):
- Pre-generates millions of random 7-char keys
- Stores in "unused_keys" table
- When URL is created: atomically move key from unused → used
- No collision risk, no coordination needed at write time
- KGS can be a separate service with its own DB
```

---

### High-Level Design

```
                    ┌─────────────────────────────────────────┐
                    │           API Gateway / LB               │
                    └──────────────┬──────────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
    ┌─────────▼──────┐   ┌─────────▼──────┐   ┌────────▼───────┐
    │  Write Service  │   │  Read Service   │   │ Analytics Svc  │
    │  (URL creation) │   │  (redirects)    │   │ (click tracking│
    └────────┬────────┘   └────────┬────────┘   └────────┬───────┘
             │                     │                      │
    ┌────────▼────────┐   ┌────────▼────────┐   ┌────────▼───────┐
    │  Key Generation │   │  Redis Cache    │   │  Kafka         │
    │  Service (KGS)  │   │  (hot URLs)     │   │  (async events)│
    └────────┬────────┘   └────────┬────────┘   └────────┬───────┘
             │                     │                      │
    ┌────────▼─────────────────────▼──────────────────────▼───────┐
    │                    PostgreSQL (Primary + Replicas)           │
    │  Table: urls (id, short_key, long_url, created_at, expiry)  │
    └──────────────────────────────────────────────────────────────┘
```

---

### Database Schema

```sql
CREATE TABLE urls (
    id          BIGSERIAL PRIMARY KEY,
    short_key   CHAR(7) NOT NULL UNIQUE,   -- "abc1234"
    long_url    TEXT NOT NULL,
    user_id     BIGINT,                    -- NULL for anonymous
    created_at  TIMESTAMP DEFAULT NOW(),
    expires_at  TIMESTAMP,                 -- NULL = never expires
    click_count BIGINT DEFAULT 0
);

CREATE INDEX idx_short_key ON urls(short_key);  -- Primary lookup
CREATE INDEX idx_user_id ON urls(user_id);       -- "My URLs" page
CREATE INDEX idx_expires_at ON urls(expires_at); -- Cleanup job

-- Analytics (separate table, high write volume)
CREATE TABLE clicks (
    id          BIGSERIAL PRIMARY KEY,
    short_key   CHAR(7) NOT NULL,
    clicked_at  TIMESTAMP DEFAULT NOW(),
    ip_address  INET,
    user_agent  TEXT,
    referrer    TEXT,
    country     CHAR(2)
);
-- Partition by month for efficient cleanup
```

---

### Read Path (Critical — 35,000 req/sec)

```
1. Request: GET /abc1234
2. Check Redis cache: key = "url:abc1234"
   ├── Cache HIT (99% of traffic): Return 301/302 redirect immediately
   └── Cache MISS (1%): Query PostgreSQL → Cache result → Return redirect

Redis cache entry:
  Key:   "url:abc1234"
  Value: "https://www.example.com/very/long/url"
  TTL:   24 hours (or until URL expires)

Cache hit rate target: 99%+ (Zipf distribution — top 1% of URLs = 99% of traffic)
```

**301 vs 302 Redirect:**
- `301 Moved Permanently`: Browser caches redirect. Subsequent requests go directly to long URL. Reduces server load but you lose click analytics.
- `302 Found` (Temporary): Browser always hits your server. You capture every click for analytics. Use this.

---

### Write Path

```
1. POST /api/shorten { long_url: "https://...", custom_alias: "my-brand" }
2. Validate URL (is it a real URL? is it malicious?)
3. Check if custom alias requested:
   ├── Yes: Check if alias available → Reserve it
   └── No: Get next key from KGS
4. Store in PostgreSQL
5. Cache in Redis (pre-warm)
6. Return short URL
```

---

### Handling Expiry

```go
// Cleanup job runs every hour
func cleanupExpiredURLs() {
    // Soft delete: mark as expired (don't delete — keep for analytics)
    db.Exec(`
        UPDATE urls SET deleted_at = NOW()
        WHERE expires_at < NOW() AND deleted_at IS NULL
    `)
    
    // Remove from Redis cache
    expiredKeys := db.Query(`
        SELECT short_key FROM urls
        WHERE expires_at < NOW() AND deleted_at IS NOT NULL
    `)
    for _, key := range expiredKeys {
        redis.Del("url:" + key)
    }
}
```

---

### Scaling Decisions

**Why separate Read and Write services?**
- Read: 35,000 req/sec, needs Redis cache, stateless, scale horizontally
- Write: 1,160 req/sec, needs KGS coordination, different scaling profile

**Why Redis for caching?**
- 35,000 reads/sec would destroy PostgreSQL without cache
- Redis handles 100,000+ ops/sec on a single node
- URL data is small (200 bytes) — cache millions of URLs in RAM

**Why KGS instead of hash?**
- No collision risk
- Pre-generated keys = no computation at write time
- KGS can be replicated for HA

---

### Trade-offs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Short key generation | KGS pre-generated | MD5 hash | No collisions, faster writes |
| Redirect type | 302 | 301 | Need analytics on every click |
| Cache | Redis | Memcached | Redis supports TTL, persistence |
| Analytics | Async (Kafka) | Synchronous | Don't slow down redirects |
| Storage | PostgreSQL | Cassandra | Relational queries needed (user's URLs) |

---

### Bottlenecks & Solutions

```
Bottleneck 1: Read QPS (35,000/sec)
Solution: Redis cache with 99%+ hit rate. Only 350 req/sec hit DB.

Bottleneck 2: KGS single point of failure
Solution: KGS has standby replica. Pre-load keys into Write Service memory (batch of 1000).

Bottleneck 3: Analytics writes (35,000 clicks/sec)
Solution: Async via Kafka. Click events buffered, batch-written to analytics DB.

Bottleneck 4: URL expiry cleanup
Solution: Background job + Redis TTL. Expired URLs return 404 immediately from cache.
```

---

### Interview Follow-up Questions
- "How would you prevent abuse (someone shortening malicious URLs)?"
- "How would you implement custom domains (mycompany.short/abc)?"
- "How would you scale the analytics to handle 35,000 writes/sec?"
- "What happens if the KGS goes down?"
- "How would you implement URL preview (show destination before redirecting)?"

---

### Key Takeaways

1. **Base62 encoding of auto-increment ID is the cleanest approach** — no collisions, predictable length, 62^7 = 3.5 trillion unique URLs.
2. **Pre-generated keys (KGS) eliminate write-time computation** — keys are ready before requests arrive. No hashing, no collision checking at write time.
3. **Redis cache absorbs 99%+ of read traffic** — Zipf distribution means top 1% of URLs get 99% of traffic. Cache them all.
4. **302 over 301 for analytics** — 301 is cached by browsers (you lose click data). 302 hits your server every time (you capture every click).
5. **Separate read and write services** — 35K reads/sec vs 1.2K writes/sec. Different scaling profiles, different caching needs.
6. **Analytics must be async** — never slow down a redirect (user experience) to record a click. Publish to Kafka, process asynchronously.

---

## Design 2: Twitter / X {#design2}

### Requirements Clarification

**Functional Requirements:**
- Post tweets (text ≤280 chars, images, videos)
- Follow/unfollow users
- Home timeline: see tweets from people you follow
- User profile timeline: see a user's tweets
- Like, retweet, reply
- Search tweets and users
- Notifications (likes, follows, mentions)

**Non-Functional Requirements:**
- 300M daily active users (DAU)
- 500M tweets posted per day
- Read-heavy: 100:1 read/write ratio
- Timeline load: <200ms
- High availability: 99.99%
- Eventual consistency acceptable for timelines

---

### Capacity Estimation

```
Write QPS:  500M tweets/day / 86,400 = ~5,800 tweets/sec
Read QPS:   5,800 × 100 = ~580,000 timeline reads/sec
Peak reads: 580,000 × 3 = ~1.74M reads/sec

Storage per tweet:
  - tweet_id: 8 bytes
  - user_id: 8 bytes
  - text: 280 bytes
  - created_at: 8 bytes
  - metadata: 20 bytes
  Total: ~324 bytes

Storage for 5 years:
  500M/day × 365 × 5 = 912.5B tweets
  912.5B × 324 bytes = ~296TB (tweets only)
  Media (images/video): ~10x = ~3PB total
```

---

### The Core Problem: Timeline Generation

This is the hardest part of Twitter's design. Two approaches:

**Approach 1 — Fan-out on Write (Push Model)**
When a user tweets, immediately push the tweet to all followers' timeline caches.

```
User A (100 followers) tweets:
→ Write tweet to DB
→ For each of 100 followers: append tweet_id to their timeline cache
→ Timeline read: just read from cache (fast!)

Problem: Celebrity with 10M followers tweets
→ 10M cache writes in real-time
→ "Thundering herd" on write
→ Lady Gaga's tweet takes 10 minutes to fan out
```

**Approach 2 — Fan-out on Read (Pull Model)**
When a user loads their timeline, query tweets from all followed users.

```
User loads timeline:
→ Get list of 500 followed users
→ Query each user's recent tweets
→ Merge and sort by time
→ Return top 20

Problem: 500 DB queries per timeline load
→ 580,000 timeline loads/sec × 500 queries = 290M queries/sec
→ Impossible
```

**Twitter's Actual Solution — Hybrid**
- Regular users (< 1M followers): Fan-out on write (push to followers' caches)
- Celebrities (> 1M followers): Fan-out on read (pull at read time, merge with cached timeline)

```
Timeline load for User B:
1. Read User B's pre-computed timeline cache (tweets from regular users)
2. For each celebrity User B follows: fetch their latest tweets
3. Merge celebrity tweets into timeline
4. Return sorted result

This limits fan-out write cost while keeping read latency low.
```

---

### High-Level Architecture

```
                         ┌──────────────────────────────┐
                         │      API Gateway / CDN        │
                         └──────────────┬───────────────┘
                                        │
          ┌─────────────────────────────┼──────────────────────────┐
          │                             │                          │
┌─────────▼──────┐           ┌──────────▼──────┐        ┌─────────▼──────┐
│  Tweet Service  │           │ Timeline Service │        │ Search Service │
│  (write tweets) │           │ (read timelines) │        │ (Elasticsearch)│
└────────┬────────┘           └──────────┬──────┘        └────────────────┘
         │                               │
         │ publish                       │ read
         ▼                               ▼
┌────────────────┐           ┌───────────────────────┐
│  Kafka         │           │  Redis Timeline Cache  │
│  (tweet events)│           │  user_id → [tweet_ids] │
└────────┬───────┘           └───────────────────────┘
         │
         ▼
┌────────────────────────────────────────────────────────┐
│              Fan-out Service (consumers)                │
│  For each tweet: push to followers' timeline caches    │
│  Skip celebrities (>1M followers) — pull at read time  │
└────────────────────────────────────────────────────────┘
         │
         ▼
┌────────────────────────────────────────────────────────┐
│                    Data Layer                           │
│  Tweets DB (Cassandra — write-heavy, time-series)      │
│  Users DB (PostgreSQL — relational, follows graph)     │
│  Media Store (S3 + CDN)                                │
└────────────────────────────────────────────────────────┘
```

---

### Database Design

**Tweets — Cassandra (Write-Heavy, Time-Series)**
```sql
-- Cassandra table (partition by user, cluster by time)
CREATE TABLE tweets (
    user_id    BIGINT,
    tweet_id   TIMEUUID,    -- UUID with embedded timestamp
    text       TEXT,
    media_urls LIST<TEXT>,
    like_count COUNTER,
    retweet_count COUNTER,
    PRIMARY KEY (user_id, tweet_id)
) WITH CLUSTERING ORDER BY (tweet_id DESC);

-- Fetch user's tweets: fast (single partition)
SELECT * FROM tweets WHERE user_id = 123 LIMIT 20;
```

**Why Cassandra for tweets?**
- Write-heavy (500M tweets/day = 5,800/sec)
- Time-series access pattern (recent tweets first)
- Horizontal scale by adding nodes
- No complex joins needed

**Users & Follows — PostgreSQL**
```sql
CREATE TABLE users (
    user_id     BIGSERIAL PRIMARY KEY,
    username    VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    bio         TEXT,
    follower_count INT DEFAULT 0,
    following_count INT DEFAULT 0,
    is_celebrity BOOLEAN DEFAULT FALSE  -- >1M followers
);

CREATE TABLE follows (
    follower_id BIGINT REFERENCES users(user_id),
    followee_id BIGINT REFERENCES users(user_id),
    created_at  TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id)
);

CREATE INDEX idx_followee ON follows(followee_id);  -- "who follows me"
CREATE INDEX idx_follower ON follows(follower_id);  -- "who do I follow"
```

---

### Timeline Cache (Redis)

```
Key:   "timeline:{user_id}"
Value: Sorted Set of tweet_ids, scored by timestamp
TTL:   48 hours (inactive users' caches expire)

Operations:
  ZADD timeline:123 1699000000 tweet_id_abc  -- Add tweet
  ZREVRANGE timeline:123 0 19               -- Get latest 20 tweets
  ZCARD timeline:123                        -- Count tweets in cache

Cache size per user: 800 tweet_ids × 8 bytes = 6.4KB
For 300M users: 300M × 6.4KB = 1.92TB Redis cluster
```

---

### Fan-out Service

```go
func fanOutTweet(tweet Tweet) {
    // Get all followers
    followers := db.Query(`
        SELECT follower_id, is_celebrity
        FROM follows f
        JOIN users u ON f.follower_id = u.user_id
        WHERE f.followee_id = ?
    `, tweet.UserID)
    
    // Check if author is celebrity
    author := db.GetUser(tweet.UserID)
    if author.IsCelebrity {
        // Don't fan out — followers will pull at read time
        return
    }
    
    // Fan out to all followers' timeline caches
    pipe := redis.Pipeline()
    for _, follower := range followers {
        key := fmt.Sprintf("timeline:%d", follower.ID)
        pipe.ZAdd(key, redis.Z{
            Score:  float64(tweet.CreatedAt.Unix()),
            Member: tweet.ID,
        })
        // Keep only latest 800 tweets per timeline
        pipe.ZRemRangeByRank(key, 0, -801)
    }
    pipe.Exec(ctx)
}
```

---

### Search

```
Elasticsearch index: tweets
  - tweet_id, user_id, text, created_at, hashtags, mentions
  - Full-text search on text field
  - Filter by date range, user, hashtag

Indexing: Kafka consumer reads tweet events → indexes to Elasticsearch
Search latency: <100ms for most queries
```

---

### Trade-offs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Timeline | Hybrid fan-out | Pure push or pull | Balance write cost vs read latency |
| Tweet storage | Cassandra | PostgreSQL | Write-heavy, time-series, no joins |
| Timeline cache | Redis Sorted Set | Memcached | Need sorted order by timestamp |
| Celebrity threshold | 1M followers | Any threshold | Balance fan-out cost vs read complexity |
| Search | Elasticsearch | PostgreSQL FTS | Scale, relevance ranking, facets |

---

### Bottlenecks & Solutions

```
Bottleneck 1: Celebrity tweet fan-out (10M followers)
Solution: Skip fan-out for celebrities. Pull at read time and merge.

Bottleneck 2: Timeline read (1.74M reads/sec)
Solution: Redis cache. 99%+ of timeline reads served from cache.

Bottleneck 3: Tweet writes (5,800/sec)
Solution: Cassandra handles 100K+ writes/sec per cluster.

Bottleneck 4: Follow graph queries
Solution: Cache follower lists in Redis. Invalidate on follow/unfollow.

Bottleneck 5: Media storage (images, videos)
Solution: S3 for storage, CloudFront CDN for delivery.
```

---

### Interview Follow-up Questions
- "How do you handle the 'Bieber problem' — a celebrity with 100M followers?"
- "How would you implement trending topics?"
- "How do you ensure tweet ordering is consistent across data centers?"
- "How would you implement the 'who to follow' recommendation?"
- "How do you handle tweet deletion — it must disappear from all timelines?"

---

### Key Takeaways

1. **Hybrid fan-out solves the celebrity problem** — push to regular followers' caches, pull celebrity tweets at read time. Neither pure push nor pure pull works alone.
2. **Cassandra for tweets, PostgreSQL for social graph** — tweets are write-heavy time-series (Cassandra). Follow relationships need relational queries (PostgreSQL).
3. **Redis Sorted Set for timeline cache** — score = timestamp, member = tweet_id. ZREVRANGE gives latest N tweets in O(log N + K).
4. **Timeline cache is the critical path** — 1.74M reads/sec must hit Redis, not Cassandra. 99%+ cache hit rate is the target.
5. **Fan-out is async** — tweet is written to DB first, then fan-out happens in background. User gets confirmation immediately, followers see it within seconds.
6. **Search is a separate system** — Elasticsearch indexed via Kafka consumer. Never query Cassandra for full-text search.

---

## Design 3: Instagram {#design3}

### Requirements Clarification

**Functional Requirements:**
- Upload photos and videos
- Follow/unfollow users
- News feed: see posts from followed users
- Like and comment on posts
- Stories (24-hour expiry)
- Explore/Discover page
- Direct messages (basic)

**Non-Functional Requirements:**
- 1B monthly active users, 500M DAU
- 100M photos uploaded per day
- 2B feed reads per day
- Photo upload: <2 seconds
- Feed load: <500ms
- High availability: 99.99%

---

### Capacity Estimation

```
Photo uploads: 100M/day = 1,160/sec
Feed reads: 2B/day = 23,150/sec
Peak feed reads: 23,150 × 3 = ~70,000/sec

Storage per photo:
  - Original: avg 3MB
  - Thumbnail (150×150): 20KB
  - Medium (640×640): 200KB
  - Large (1080×1080): 800KB
  Total per upload: ~4MB

Daily storage: 100M × 4MB = 400TB/day
5-year storage: 400TB × 365 × 5 = ~730PB (petabytes)
→ Need distributed object storage (S3)

Metadata per post:
  - post_id, user_id, caption, location, created_at: ~200 bytes
  - 100M posts/day × 200 bytes = 20GB/day metadata
```

---

### Photo Upload Pipeline

The upload path is critical — users abandon if it takes >2 seconds.

```
1. Client requests upload URL
2. Server generates pre-signed S3 URL (valid 10 minutes)
3. Client uploads directly to S3 (bypasses your servers!)
4. S3 triggers Lambda/event on upload complete
5. Image processing service:
   - Generate thumbnails (150px, 640px, 1080px)
   - Extract metadata (EXIF, dimensions)
   - Run content moderation (NSFW detection)
   - Store processed versions back to S3
6. Write metadata to database
7. Publish "post.created" event to Kafka
8. Fan-out service updates followers' feeds

Why direct-to-S3 upload?
- Your API servers don't handle large file uploads
- S3 handles multipart upload, retry, resumable uploads
- Scales to any upload volume without touching your servers
```

```
Upload Flow:
Client ──▶ API Server: "I want to upload a photo"
API Server ──▶ S3: Generate pre-signed URL
API Server ──▶ Client: { upload_url: "https://s3.amazonaws.com/...?signature=..." }
Client ──▶ S3 directly: PUT photo (3MB)
S3 ──▶ Lambda: "photo uploaded" event
Lambda ──▶ Image Processing Service: resize, moderate
Image Processing ──▶ S3: store thumbnails
Image Processing ──▶ DB: write post metadata
Image Processing ──▶ Kafka: publish "post.created"
```

---

### Feed Generation

Instagram uses **pre-computed feeds** (fan-out on write) for most users:

```
When User A posts:
1. Get User A's followers (from follow graph service)
2. For each follower: append post_id to their feed cache
3. Feed cache: Redis sorted set, scored by timestamp
4. Keep last 500 posts per user in cache

When User B loads feed:
1. Read from Redis: ZREVRANGE feed:user_b 0 19 (latest 20 posts)
2. Fetch post details from post cache/DB
3. Return to client

Feed cache structure:
  Key:   "feed:{user_id}"
  Value: Sorted Set { post_id: timestamp }
  Size:  500 posts × 8 bytes = 4KB per user
  Total: 500M users × 4KB = 2TB Redis cluster
```

---

### High-Level Architecture

```
                    ┌──────────────────────────────────────┐
                    │         CDN (CloudFront)              │
                    │   Serves photos, videos, thumbnails   │
                    └──────────────┬───────────────────────┘
                                   │
                    ┌──────────────▼───────────────────────┐
                    │         API Gateway / LB              │
                    └──┬──────────┬──────────┬─────────────┘
                       │          │          │
            ┌──────────▼──┐  ┌────▼────┐  ┌─▼──────────┐
            │ Upload Svc  │  │ Feed Svc│  │ User/Follow│
            │ (pre-signed │  │ (read)  │  │ Service    │
            │  S3 URLs)   │  └────┬────┘  └────────────┘
            └──────┬──────┘       │
                   │         ┌────▼────────────────┐
                   │         │  Redis Feed Cache    │
                   │         │  (pre-computed feeds)│
                   │         └─────────────────────┘
                   ▼
            ┌──────────────┐     ┌──────────────────────┐
            │  S3 (photos) │     │  PostgreSQL           │
            │  + CDN       │     │  (posts, users,       │
            └──────────────┘     │   follows, likes)     │
                                 └──────────────────────┘
                   │
                   ▼
            ┌──────────────┐
            │  Kafka       │──▶ Fan-out Service
            │  (events)    │──▶ Notification Service
            └──────────────┘──▶ Analytics Service
```

---

### Database Schema

```sql
-- Posts
CREATE TABLE posts (
    post_id     BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    caption     TEXT,
    location    POINT,
    created_at  TIMESTAMP DEFAULT NOW(),
    like_count  INT DEFAULT 0,
    comment_count INT DEFAULT 0
);

-- Post media (one post can have multiple photos/videos)
CREATE TABLE post_media (
    media_id    BIGSERIAL PRIMARY KEY,
    post_id     BIGINT REFERENCES posts(post_id),
    media_type  VARCHAR(10),  -- 'photo', 'video'
    s3_key      TEXT NOT NULL,  -- S3 object key
    width       INT,
    height      INT,
    duration_sec INT           -- for videos
);

-- Likes (high write volume — consider Cassandra)
CREATE TABLE likes (
    post_id     BIGINT,
    user_id     BIGINT,
    created_at  TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

-- Stories (24-hour TTL — use Redis or Cassandra with TTL)
CREATE TABLE stories (
    story_id    BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    s3_key      TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    expires_at  TIMESTAMP DEFAULT NOW() + INTERVAL '24 hours'
);
CREATE INDEX idx_stories_expiry ON stories(expires_at);
```

---

### Stories Architecture

Stories expire after 24 hours — this is a natural fit for Redis with TTL:

```
Upload story:
  redis.SETEX("story:{story_id}", 86400, story_data)  -- 24 hour TTL
  redis.ZADD("user_stories:{user_id}", timestamp, story_id)
  redis.EXPIRE("user_stories:{user_id}", 86400)

View stories:
  story_ids = redis.ZRANGE("user_stories:{user_id}", 0, -1)
  stories = redis.MGET(["story:{id}" for id in story_ids])

Auto-expiry: Redis TTL handles deletion automatically
```

---

### Trade-offs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Photo upload | Direct-to-S3 | Through API servers | API servers don't bottleneck on large files |
| Feed | Pre-computed (push) | Pull on read | 70K reads/sec needs pre-computation |
| Photo storage | S3 + CDN | Self-hosted | Petabyte scale, global delivery |
| Stories | Redis with TTL | PostgreSQL + cron | Natural TTL, fast reads |
| Likes | Cassandra | PostgreSQL | High write volume, simple access pattern |

---

### Interview Follow-up Questions
- "How do you handle the Explore/Discover page (content for users you don't follow)?"
- "How would you implement Instagram Reels (short videos)?"
- "How do you handle content moderation at scale?"
- "How would you implement 'seen by' for stories?"
- "How do you handle a user with 100M followers posting a photo?"

---

### Key Takeaways

1. **Direct-to-S3 upload is mandatory at scale** — never route large file uploads through your API servers. Pre-signed URLs let clients upload directly to S3.
2. **Pre-computed feeds (push model) for most users** — 70K feed reads/sec requires pre-computation. Pull model would require 70K × 500 DB queries/sec.
3. **Redis TTL is perfect for Stories** — 24-hour expiry maps directly to Redis TTL. No cron jobs, no cleanup queries needed.
4. **CDN for all media** — petabyte-scale photo/video storage requires S3 + CloudFront. Never serve media from your origin servers.
5. **Separate photo upload pipeline from feed** — upload is async (S3 event → Lambda → processing). Feed reads are synchronous. Different scaling requirements.
6. **Likes need special handling at scale** — 2B likes/day = 23K writes/sec. Use Cassandra or Redis counters, not PostgreSQL row updates.

---

## Design 4: WhatsApp / Messaging System {#design4}

### Requirements Clarification

**Functional Requirements:**
- 1:1 messaging (text, images, voice, video)
- Group messaging (up to 256 members)
- Message delivery receipts (sent ✓, delivered ✓✓, read ✓✓ blue)
- Online/offline presence
- End-to-end encryption
- Message history sync across devices

**Non-Functional Requirements:**
- 2B users, 100M DAU
- 100B messages per day
- Message delivery: <100ms (online), <1 second (offline)
- High availability: 99.999% (5 nines)
- Messages must never be lost
- End-to-end encrypted (server cannot read messages)

---

### Capacity Estimation

```
Messages: 100B/day = 1.16M messages/sec
Peak: 1.16M × 3 = ~3.5M messages/sec

Storage per message:
  - message_id: 16 bytes (UUID)
  - sender_id, receiver_id: 8 bytes each
  - content (encrypted): avg 100 bytes
  - timestamp: 8 bytes
  - status: 1 byte
  Total: ~141 bytes

Daily storage: 100B × 141 bytes = ~14TB/day
5-year storage: 14TB × 365 × 5 = ~25.5PB

Connections: 100M DAU, avg 30 min active = 100M concurrent WebSocket connections
```

---

### The Core Challenge: Real-Time Messaging

HTTP is request-response — the client must ask for new messages. For real-time messaging, you need a persistent connection where the server can push messages to the client.

**Options:**
1. **Short Polling**: Client polls every 1 second. Wasteful, high latency.
2. **Long Polling**: Client holds connection open until message arrives or timeout. Better but still HTTP overhead.
3. **WebSocket**: Persistent bidirectional TCP connection. Best for real-time. WhatsApp uses this.
4. **Server-Sent Events (SSE)**: Server pushes to client, client cannot push back. Good for notifications, not messaging.

**WhatsApp uses WebSocket** for real-time delivery.

---

### Message Flow

```
ONLINE DELIVERY (both users online):
Sender ──WebSocket──▶ Chat Server A
Chat Server A ──▶ Message Queue (Kafka)
Message Queue ──▶ Chat Server B (where receiver is connected)
Chat Server B ──WebSocket──▶ Receiver
Receiver ──▶ Chat Server B: "delivered" ack
Chat Server B ──▶ Kafka ──▶ Chat Server A
Chat Server A ──WebSocket──▶ Sender: ✓✓ (delivered)

OFFLINE DELIVERY (receiver offline):
Sender ──WebSocket──▶ Chat Server A
Chat Server A ──▶ Message stored in DB (Cassandra)
Chat Server A ──▶ Push notification to receiver's device (APNs/FCM)
Receiver comes online ──▶ Connects to Chat Server
Chat Server ──▶ Fetches undelivered messages from DB
Chat Server ──WebSocket──▶ Receiver: delivers messages
Receiver ──▶ "delivered" ack ──▶ Sender: ✓✓
```

---

### Architecture

```
                    ┌──────────────────────────────────────┐
                    │         Load Balancer (L4/L7)         │
                    │   Sticky sessions for WebSocket       │
                    └──────────────┬───────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
    ┌─────────▼──────┐   ┌─────────▼──────┐   ┌────────▼───────┐
    │  Chat Server 1  │   │  Chat Server 2  │   │  Chat Server N │
    │  (WebSocket)    │   │  (WebSocket)    │   │  (WebSocket)   │
    │  Holds 100K     │   │  Holds 100K     │   │  connections   │
    │  connections    │   │  connections    │   │  each          │
    └────────┬────────┘   └────────┬────────┘   └────────┬───────┘
             │                     │                      │
             └─────────────────────┼──────────────────────┘
                                   │
                    ┌──────────────▼───────────────────────┐
                    │              Kafka                    │
                    │  (message routing between servers)    │
                    └──────────────┬───────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
    ┌─────────▼──────┐   ┌─────────▼──────┐   ┌────────▼───────┐
    │  Cassandra      │   │  Redis          │   │  Push Notif.   │
    │  (messages)     │   │  (presence,     │   │  Service       │
    │                 │   │   sessions)     │   │  (APNs, FCM)   │
    └─────────────────┘   └─────────────────┘   └────────────────┘
```

---

### Presence System (Online/Offline)

```
User connects: 
  redis.SETEX("presence:{user_id}", 30, "online")  -- 30 second TTL
  redis.HSET("user_server:{user_id}", "server_id", "chat-server-42")

Heartbeat (every 15 seconds):
  redis.EXPIRE("presence:{user_id}", 30)  -- Refresh TTL

User disconnects or heartbeat stops:
  TTL expires → user appears offline automatically

Check if user is online:
  redis.EXISTS("presence:{user_id}")  -- O(1) lookup

Which server is user connected to?
  redis.HGET("user_server:{user_id}", "server_id")
  → "chat-server-42"
  → Route message to that server via Kafka topic "server-42"
```

---

### Message Storage (Cassandra)

```sql
-- Messages partitioned by conversation for fast retrieval
CREATE TABLE messages (
    conversation_id UUID,
    message_id      TIMEUUID,    -- UUID with embedded timestamp
    sender_id       BIGINT,
    content         BLOB,        -- Encrypted content
    message_type    VARCHAR(10), -- 'text', 'image', 'voice'
    status          TINYINT,     -- 0=sent, 1=delivered, 2=read
    PRIMARY KEY (conversation_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);

-- Fetch last 50 messages in a conversation:
SELECT * FROM messages
WHERE conversation_id = ? 
LIMIT 50;
-- Fast: single partition read
```

---

### Group Messaging

```
Group message delivery:
1. Sender sends message to group (group_id)
2. Chat server fetches group members (from Groups DB)
3. For each member:
   a. If online: route to their chat server via Kafka
   b. If offline: store in DB + send push notification
4. Delivery receipts: collected per-member
5. "Read by all" = all members have sent read receipt

Group size limit (256 members):
- Limits fan-out cost
- 256 Kafka messages per group message
- At 1M group messages/sec: 256M Kafka messages/sec (manageable)
```

---

### Trade-offs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Real-time protocol | WebSocket | Long polling | Persistent connection, lower overhead |
| Message storage | Cassandra | PostgreSQL | Write-heavy, time-series, no complex joins |
| Routing | Kafka | Direct server-to-server | Decoupled, handles server failures |
| Presence | Redis TTL | DB polling | Sub-millisecond, auto-expiry |
| Encryption | E2E (client-side) | Server-side | Server cannot read messages |

---

### Interview Follow-up Questions
- "How do you handle message ordering in a distributed system?"
- "How do you implement end-to-end encryption? What key exchange protocol?"
- "How do you handle 100M concurrent WebSocket connections?"
- "How do you sync message history when a user gets a new device?"
- "How do you implement 'delete for everyone' after a message is delivered?"

---

### Key Takeaways

1. **WebSocket for real-time messaging** — persistent bidirectional TCP connection. Lower overhead than HTTP polling, enables true push delivery.
2. **Redis presence + routing** — store which server each user is connected to. Route messages via Kafka to the correct server.
3. **Cassandra for message storage** — partition by conversation_id, cluster by message_id (TIMEUUID). Single-partition reads for conversation history.
4. **Offline delivery via push notifications** — when receiver is offline, store message in DB and send APNs/FCM push. Deliver when they reconnect.
5. **100M concurrent WebSocket connections** — each server handles ~100K connections. 1,000 servers for 100M users. Go handles this efficiently (goroutines are cheap).
6. **End-to-end encryption means server cannot read messages** — store encrypted blobs. Key exchange happens client-to-client (Signal Protocol).
7. **Group message fan-out is bounded** — 256 member limit keeps fan-out manageable. At 1M group messages/sec: 256M Kafka messages/sec.

---

## Design 5: YouTube / Video Streaming {#design5}

### Requirements Clarification

**Functional Requirements:**
- Upload videos (up to 4K, any length)
- Stream videos (adaptive bitrate)
- Search videos
- Like, comment, subscribe
- Recommendations
- Video analytics (views, watch time)

**Non-Functional Requirements:**
- 2B logged-in users/month, 500M DAU
- 500 hours of video uploaded per minute
- 1B hours of video watched per day
- Upload: process within 5 minutes of upload
- Streaming: start within 2 seconds, no buffering
- High availability: 99.99%

---

### Capacity Estimation

```
Video uploads: 500 hours/min = 30,000 hours/day
Storage per hour of video:
  - Raw 1080p: ~2GB/hour
  - After encoding (multiple resolutions): ~5GB/hour total
  Daily storage: 30,000 hours × 5GB = 150TB/day
  5-year storage: 150TB × 365 × 5 = ~274PB

Video views: 1B hours/day = 41.7M hours/hour = ~700K concurrent streams
Bandwidth: 700K streams × 5Mbps avg = 3.5Tbps (terabits per second!)
→ Must use CDN — impossible to serve from origin
```

---

### Video Upload and Processing Pipeline

This is the most complex part. Raw video must be transcoded into multiple formats and resolutions before it can be streamed.

```
Upload Pipeline:
1. User uploads raw video to S3 (direct upload, pre-signed URL)
2. S3 event triggers processing pipeline
3. Video Processing Service:
   a. Validate: check format, duration, file integrity
   b. Extract metadata: duration, resolution, codec, thumbnail
   c. Transcode to multiple resolutions:
      - 360p (mobile, low bandwidth)
      - 480p (standard)
      - 720p (HD)
      - 1080p (Full HD)
      - 1440p, 4K (if source is high quality)
   d. Generate thumbnails (every 10 seconds for scrubbing)
   e. Extract audio track (for captions)
   f. Run content moderation (copyright, NSFW)
4. Store processed files in S3
5. Update video metadata in DB
6. Invalidate CDN cache for this video
7. Notify user: "Your video is ready"

Transcoding is CPU-intensive:
  1 hour of 4K video → 4 hours of transcoding on 1 server
  Solution: Parallel transcoding — split video into 1-minute segments,
  transcode each segment in parallel, reassemble
  Result: 1 hour of 4K video → 15 minutes of transcoding (16x speedup)
```

---

### Adaptive Bitrate Streaming (ABR)

YouTube doesn't serve one video file. It serves small chunks and adapts quality based on network speed.

```
HLS (HTTP Live Streaming) or DASH:
- Video split into 2-10 second segments
- Each segment encoded at multiple bitrates
- Manifest file (.m3u8) lists all segments and their URLs

Client player:
1. Download manifest file
2. Start with medium quality (720p)
3. Measure download speed of each segment
4. If download is fast: switch to higher quality (1080p)
5. If download is slow: switch to lower quality (360p)
6. User sees: smooth playback, quality adjusts automatically

Segment storage in S3:
  video_id/
    manifest.m3u8
    360p/
      segment_001.ts
      segment_002.ts
      ...
    720p/
      segment_001.ts
      ...
    1080p/
      segment_001.ts
      ...
```

---

### Architecture

```
UPLOAD PATH:
User ──▶ API Server ──▶ Pre-signed S3 URL
User ──▶ S3 (direct upload)
S3 ──▶ SQS/Kafka ──▶ Video Processing Workers (EC2 fleet)
                      ├── Transcoding (FFmpeg)
                      ├── Thumbnail generation
                      ├── Content moderation
                      └── Metadata extraction
Processing Workers ──▶ S3 (processed segments)
Processing Workers ──▶ DB (video metadata)
Processing Workers ──▶ CDN (pre-warm popular videos)

STREAMING PATH:
User ──▶ CDN Edge (nearest PoP)
         ├── Cache HIT: Serve segment directly (99% of traffic)
         └── Cache MISS: Fetch from S3 origin, cache, serve
         
CDN PoPs: 200+ locations worldwide
Segment TTL: 1 year (content never changes once processed)
```

---

### Database Design

```sql
-- Videos metadata
CREATE TABLE videos (
    video_id        BIGSERIAL PRIMARY KEY,
    uploader_id     BIGINT NOT NULL,
    title           VARCHAR(100) NOT NULL,
    description     TEXT,
    duration_sec    INT,
    view_count      BIGINT DEFAULT 0,
    like_count      INT DEFAULT 0,
    status          VARCHAR(20) DEFAULT 'processing',
    -- 'processing', 'ready', 'rejected', 'deleted'
    s3_manifest_key TEXT,       -- Path to HLS manifest
    thumbnail_key   TEXT,       -- Path to thumbnail in S3
    created_at      TIMESTAMP DEFAULT NOW(),
    published_at    TIMESTAMP
);

-- View counts: high write volume, use approximate counting
-- Option 1: Redis INCR (fast, approximate)
-- Option 2: Kafka events → batch aggregation → DB update every minute
-- Option 3: HyperLogLog for unique viewers

-- Comments (Cassandra — high write volume, time-series)
CREATE TABLE comments (
    video_id    BIGINT,
    comment_id  TIMEUUID,
    user_id     BIGINT,
    text        TEXT,
    like_count  INT DEFAULT 0,
    PRIMARY KEY (video_id, comment_id)
) WITH CLUSTERING ORDER BY (comment_id DESC);
```

---

### View Count at Scale

```
Problem: 1B views/day = 11,600 view events/sec
         Cannot do UPDATE videos SET view_count = view_count + 1 for each view
         → Lock contention, DB bottleneck

Solution: Approximate counting with Redis + periodic flush

// On each view:
redis.INCR("views:{video_id}")

// Every 60 seconds, flush to DB:
for each video_id in redis:
    count = redis.GETDEL("views:{video_id}")
    db.Exec("UPDATE videos SET view_count = view_count + ? WHERE video_id = ?", count, video_id)

Result: DB gets 1 update per video per minute instead of 11,600/sec
```

---

### Trade-offs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Video storage | S3 + CDN | Self-hosted | Petabyte scale, global CDN |
| Streaming | HLS/DASH (ABR) | Single file | Adapts to network, no buffering |
| Transcoding | Parallel segments | Sequential | 16x faster processing |
| View counts | Redis + batch flush | Direct DB update | Avoid lock contention |
| Search | Elasticsearch | PostgreSQL FTS | Scale, relevance, facets |

---

### Interview Follow-up Questions
- "How do you handle copyright detection (Content ID system)?"
- "How would you implement video recommendations?"
- "How do you handle live streaming (different from recorded video)?"
- "How do you implement video chapters and timestamps?"
- "How do you handle a video going viral (10M views in 1 hour)?"

---

### Key Takeaways

1. **Direct-to-S3 upload + event-driven processing** — never route video uploads through your servers. S3 event triggers the processing pipeline.
2. **Parallel transcoding is the key optimization** — split video into 1-minute segments, transcode in parallel. 1 hour of 4K video in 15 minutes instead of 4 hours.
3. **HLS/DASH adaptive bitrate** — serve 2-10 second segments at multiple quality levels. Player adapts quality to network speed. No buffering.
4. **CDN absorbs 99.9% of streaming traffic** — 3.5Tbps bandwidth requirement is impossible from origin. CloudFront/Akamai with 200+ PoPs handles it.
5. **View counts via Redis + batch flush** — 11,600 view events/sec cannot hit PostgreSQL directly. Redis INCR + periodic flush to DB reduces DB writes by 100x.
6. **Content ID for copyright** — fingerprint every uploaded video, compare against rights-holder database. Block or monetize infringing content automatically.

---

## Design 6: Uber / Ride-Sharing {#design6}

### Requirements Clarification

**Functional Requirements:**
- Rider requests a ride (pickup + destination)
- Match rider with nearby driver
- Real-time location tracking (driver and rider)
- ETA calculation
- Surge pricing
- Trip history and receipts
- Driver and rider ratings

**Non-Functional Requirements:**
- 100M DAU (riders + drivers)
- 10M trips per day
- Driver location updates: every 4 seconds
- Match driver to rider: <1 minute
- Location accuracy: within 10 meters
- High availability: 99.99%

---

### Capacity Estimation

```
Active drivers: 1M at peak (10% of 10M drivers)
Location updates: 1M drivers × 1 update/4sec = 250,000 location updates/sec
Trip requests: 10M/day = 116/sec (peak: ~350/sec)

Location data per update:
  - driver_id: 8 bytes
  - latitude, longitude: 16 bytes
  - timestamp: 8 bytes
  - heading, speed: 8 bytes
  Total: ~40 bytes

Location update bandwidth: 250,000 × 40 bytes = 10MB/sec (manageable)

Geospatial index size:
  1M drivers × 40 bytes = 40MB (fits in RAM easily)
```

---

### The Core Problem: Geospatial Matching

Finding nearby drivers is a geospatial problem. You need to answer: "Find all drivers within 5km of this location."

**Naive approach:** `SELECT * FROM drivers WHERE distance(lat, lng, driver_lat, driver_lng) < 5km`
- Full table scan: O(N) for N drivers
- At 1M drivers: too slow

**Solution: Geohash**

Geohash divides the world into a grid of cells. Each cell has a string code. Nearby locations share a common prefix.

```
Location: (37.7749, -122.4194) → Geohash: "9q8yy"
Precision:
  1 char  = 5,000km × 5,000km cell
  4 chars = 39km × 20km cell
  6 chars = 1.2km × 0.6km cell  ← Good for Uber
  8 chars = 38m × 19m cell

Finding nearby drivers:
1. Compute rider's geohash (6 chars): "9q8yy4"
2. Find all 9 neighboring cells (3×3 grid around rider)
3. Query: SELECT * FROM drivers WHERE geohash LIKE "9q8yy4%" OR geohash LIKE "9q8yy5%" ...
4. Filter by exact distance
5. Sort by distance, return closest available drivers

Index: CREATE INDEX idx_geohash ON drivers(geohash);
Query: O(log N) with index, returns only nearby drivers
```

**Alternative: Redis GEO commands**
```
GEOADD drivers:available driver_id longitude latitude
GEORADIUS drivers:available longitude latitude 5 km ASC COUNT 10
```
Redis GEO uses geohash internally. Sub-millisecond for 1M drivers.

---

### Architecture

```
DRIVER LOCATION UPDATE (250,000/sec):
Driver App ──▶ Location Service ──▶ Redis GEO (real-time index)
                               ──▶ Kafka (location history)
                               ──▶ Cassandra (trip tracking, async)

RIDE REQUEST:
Rider App ──▶ Ride Request Service
              ├── Get rider location
              ├── Query Redis GEO: nearby available drivers
              ├── Calculate ETA for each (Google Maps API or internal)
              ├── Select best driver (closest, highest rating)
              ├── Send offer to driver (push notification + WebSocket)
              ├── Driver accepts → create trip
              └── Driver rejects → offer to next driver

TRIP IN PROGRESS:
Driver App ──▶ Location Service ──▶ Redis (driver location)
                                ──▶ Rider App (via WebSocket push)
                                    (rider sees driver moving on map)
```

---

### Driver State Machine

```
Driver states:
  OFFLINE → AVAILABLE → ON_TRIP → AVAILABLE

State stored in Redis:
  redis.HSET("driver:{id}", "status", "available", "lat", 37.77, "lng", -122.41)

Available drivers index:
  redis.GEOADD("drivers:available", lng, lat, driver_id)

When driver accepts trip:
  redis.ZREM("drivers:available", driver_id)  -- Remove from available pool
  redis.HSET("driver:{id}", "status", "on_trip", "trip_id", trip_id)

When trip ends:
  redis.GEOADD("drivers:available", new_lng, new_lat, driver_id)
  redis.HSET("driver:{id}", "status", "available")
```

---

### Surge Pricing

```
Surge pricing = demand / supply in a geohash cell

Every 30 seconds:
  For each geohash cell (6-char precision):
    demand = count of ride requests in last 5 minutes
    supply = count of available drivers in cell
    surge_multiplier = max(1.0, demand / supply × base_factor)

Store in Redis:
  redis.SETEX("surge:{geohash}", 60, surge_multiplier)

Show to rider before booking:
  surge = redis.GET("surge:{rider_geohash}")
  if surge > 1.5: show surge warning
```

---

### Trade-offs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Geospatial index | Redis GEO | PostGIS | Sub-millisecond, in-memory |
| Location updates | Kafka + Redis | Direct DB writes | 250K writes/sec, need real-time + history |
| Driver matching | Geohash grid | Quadtree | Simpler, Redis native support |
| Trip storage | Cassandra | PostgreSQL | Write-heavy, time-series |
| Real-time tracking | WebSocket | Polling | Low latency, push updates |

---

### Interview Follow-up Questions
- "How do you handle the matching algorithm when there are no nearby drivers?"
- "How would you implement carpooling (matching multiple riders going same direction)?"
- "How do you calculate ETA accurately considering traffic?"
- "How do you handle driver location spoofing (fraud)?"
- "How would you design the payment system for Uber?"

---

### Key Takeaways

1. **Redis GEO is the right tool for driver matching** — GEORADIUS returns nearby drivers in sub-millisecond. Backed by geohash internally.
2. **250K location updates/sec requires Kafka buffering** — write to Redis (real-time index) and Kafka (history/analytics) simultaneously. Never write 250K/sec directly to a database.
3. **Geohash enables efficient proximity queries** — nearby locations share a common prefix. Index on geohash prefix = fast range scan for nearby drivers.
4. **Driver state machine in Redis** — available/on_trip state transitions are atomic Redis operations. ZADD to add to available pool, ZREM to remove when trip starts.
5. **Surge pricing is a real-time computation** — demand/supply ratio per geohash cell, updated every 30 seconds, stored in Redis with 60-second TTL.
6. **Matching must be fast** — rider expects a driver within 30 seconds. The matching algorithm (find nearest available driver) must complete in <100ms.

---

## Design 7: Netflix Recommendation System {#design7}

### Requirements Clarification

**Functional Requirements:**
- Personalized homepage (rows of recommended content)
- "Because you watched X" recommendations
- Trending content (globally and by region)
- Similar titles
- Continue watching
- New releases for your taste

**Non-Functional Requirements:**
- 230M subscribers, 100M DAU
- Recommendations must update within 24 hours of new viewing
- Homepage load: <200ms
- Offline computation acceptable (not real-time)
- Recommendations must be explainable ("Because you watched...")

---

### Capacity Estimation

```
Users: 230M subscribers, 100M DAU
Homepage loads: 100M/day = ~1,160/sec (peak: ~3,500/sec)

Viewing events (input to recommendations):
  100M users × 2 hours/day × 1 event/min = 12B events/day
  12B / 86,400 = ~139,000 events/sec → Kafka handles this

Pre-computed recommendations storage:
  230M users × 200 recommendations × 50 bytes = ~2.3TB (Cassandra)

Trending content (Redis):
  ~200 regions × 100 trending items × 200 bytes = ~4MB (trivial)

ML model training:
  12B events/day → Spark batch job (nightly, 2-4 hours)
  Model size: ~10GB per algorithm
  Training compute: 100-node Spark cluster

Homepage assembly latency budget:
  Cassandra read: ~5ms
  Redis read: ~1ms
  Assembly + serialization: ~10ms
  Total: ~16ms (well under 200ms SLO)
```

---

### The Two Types of Recommendations

**1. Collaborative Filtering**
"Users similar to you also liked X."
- Find users with similar viewing history
- Recommend what they watched that you haven't
- Problem: Cold start (new users have no history)
- Algorithm: Matrix Factorization (SVD, ALS)

**2. Content-Based Filtering**
"This show is similar to what you've watched."
- Analyze content features: genre, actors, director, themes
- Match to your viewing preferences
- Works for new content (no viewing history needed)
- Algorithm: Cosine similarity on feature vectors

**Netflix uses both** — ensemble model combining multiple signals.

---

### Data Pipeline Architecture

```
DATA COLLECTION:
User watches video ──▶ Kafka (viewing events)
User rates/likes ──▶ Kafka (interaction events)
User searches ──▶ Kafka (search events)

BATCH PROCESSING (daily):
Kafka ──▶ Spark Streaming ──▶ Data Lake (S3/HDFS)
Data Lake ──▶ Spark ML Jobs (nightly):
  ├── Collaborative filtering model training
  ├── Content similarity computation
  ├── User preference vectors
  └── Trending content calculation

RESULTS STORAGE:
Spark ──▶ Cassandra (pre-computed recommendations per user)
Spark ──▶ Elasticsearch (content search index)
Spark ──▶ Redis (trending, new releases — fast reads)

SERVING:
User opens Netflix ──▶ Recommendation Service
                   ──▶ Cassandra: "recommendations:{user_id}"
                   ──▶ Redis: "trending:{region}"
                   ──▶ Assemble homepage rows
                   ──▶ Return in <200ms
```

---

### Pre-computed Recommendations

```
Cassandra schema:
CREATE TABLE recommendations (
    user_id     BIGINT,
    row_type    VARCHAR(50),  -- 'because_you_watched', 'trending', 'new_for_you'
    rank        INT,
    content_id  BIGINT,
    score       FLOAT,
    reason      TEXT,         -- "Because you watched Stranger Things"
    PRIMARY KEY (user_id, row_type, rank)
);

-- Pre-computed nightly for all 230M users
-- Each user: ~200 recommendations across 10 rows
-- Storage: 230M × 200 × 50 bytes = ~2.3TB

Homepage assembly (fast path):
  rows = cassandra.Query("SELECT * FROM recommendations WHERE user_id = ? LIMIT 200")
  trending = redis.GET("trending:{user_region}")
  return assemble_homepage(rows, trending)
  -- Total: ~20ms
```

---

### A/B Testing Recommendations

Netflix runs hundreds of A/B tests simultaneously:

```
User is assigned to experiment group at login:
  user_group = hash(user_id) % 100  -- 0-99

Recommendation service checks:
  if user_group < 10:  -- 10% of users
      use algorithm_v2
  else:
      use algorithm_v1

Metrics tracked per group:
  - Click-through rate (CTR)
  - Watch rate (clicked and watched >5 min)
  - Completion rate
  - Return rate (came back next day)

Winner deployed to 100% after statistical significance
```

---

### Trade-offs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Computation | Batch (nightly) | Real-time | Scale, cost, accuracy |
| Storage | Pre-computed in Cassandra | Compute on request | <200ms homepage requirement |
| Algorithm | Ensemble (CF + CB) | Single algorithm | Better accuracy, handles cold start |
| Trending | Redis (fast reads) | Cassandra | Sub-millisecond, small dataset |

---

### Interview Follow-up Questions
- "How do you handle the cold start problem for new users?"
- "How would you make recommendations real-time (update within minutes of watching)?"
- "How do you measure recommendation quality?"
- "How would you handle content that's only available in certain regions?"

---

### Key Takeaways

1. **Pre-compute everything** — at 100M DAU, computing recommendations on-demand is impossible. Nightly batch jobs pre-compute for all users.
2. **Batch ML is fine for recommendations** — users don't notice if recommendations are 24 hours stale. Real-time ML is expensive and rarely justified.
3. **Collaborative filtering + content-based = ensemble** — CF handles "users like you", CB handles new content and cold start. Neither alone is sufficient.
4. **A/B testing is how you improve** — never ship a new algorithm to 100% of users. Test on 1-10%, measure CTR and watch rate, then roll out the winner.
5. **Cold start requires special handling** — new users have no history. Use content-based filtering, ask for preferences at onboarding, or show popular content.
6. **Explainability drives engagement** — "Because you watched Stranger Things" gets more clicks than an unexplained recommendation. Store the reason with each recommendation.

---

## Design 8: Distributed Cache (Redis/Memcached) {#design8}

### Requirements Clarification

**Functional Requirements:**
- GET(key) → value
- SET(key, value, ttl)
- DELETE(key)
- Support for strings, lists, sets, sorted sets, hashes
- TTL-based expiry
- Eviction when memory is full

**Non-Functional Requirements:**
- Sub-millisecond latency (P99 < 1ms)
- 1M operations/sec throughput
- 99.99% availability
- Data can be lost on restart (cache, not primary store)
- Horizontal scaling

---

### Capacity Estimation

```
Operations: 1M ops/sec target
  GET: 80% = 800,000/sec
  SET: 15% = 150,000/sec
  DELETE: 5% = 50,000/sec

Memory per key-value pair:
  Key: avg 50 bytes
  Value: avg 200 bytes
  Redis overhead: ~100 bytes per key
  Total: ~350 bytes per entry

For 100M cached objects:
  100M × 350 bytes = 35GB RAM
  Single Redis node (64GB RAM): fits comfortably
  With 3-node cluster: ~12GB per node

Throughput per node:
  Single Redis node: ~100K-1M ops/sec (single-threaded)
  For 1M ops/sec: 1-3 nodes depending on operation mix

Network bandwidth:
  1M ops/sec × 250 bytes avg payload = 250MB/sec
  10GbE NIC: 1,250MB/sec → not a bottleneck

Cluster sizing for 1M ops/sec + 100M keys:
  3 master nodes × 100K ops/sec = 300K ops/sec (conservative)
  Scale to 10 nodes for 1M ops/sec with headroom
```

---

### Core Design Decisions

**1. In-Memory Storage**
All data in RAM. No disk I/O = sub-millisecond latency.
- Hash table for O(1) GET/SET
- Doubly linked list for LRU eviction order

**2. Single-Threaded Event Loop (Redis)**
Redis uses a single thread for all operations. No locks, no context switching.
- Throughput: 100K-1M ops/sec on single node
- Latency: consistent, no lock contention
- Limitation: Cannot use multiple CPU cores for single operations

**3. Eviction Policies**
When memory is full, which keys to evict?
- `noeviction`: Return error (don't evict)
- `allkeys-lru`: Evict least recently used key
- `volatile-lru`: Evict LRU key with TTL set
- `allkeys-random`: Evict random key
- `volatile-ttl`: Evict key with shortest TTL

**LRU Implementation:**
```
Hash table: key → (value, list_node)
Doubly linked list: most_recent ↔ ... ↔ least_recent

GET(key):
  node = hash_table[key]
  move node to front of list (most recently used)
  return node.value

SET(key, value):
  if memory full:
    evict = list.tail  -- least recently used
    delete hash_table[evict.key]
    list.remove(evict)
  node = new Node(key, value)
  hash_table[key] = node
  list.prepend(node)  -- most recently used
```

---

### Distributed Cache Architecture

Single Redis node handles ~1M ops/sec. For more scale: cluster.

**Redis Cluster (Sharding):**
```
16,384 hash slots distributed across nodes
key → CRC16(key) % 16384 → slot → node

3 master nodes:
  Node 1: slots 0-5460
  Node 2: slots 5461-10922
  Node 3: slots 10923-16383

Each master has 1-2 replicas for HA

Client library (redis-go, Jedis) handles routing:
  GET "user:123" → CRC16("user:123") % 16384 = 7823 → Node 2
```

**Consistent Hashing (Alternative):**
```
Nodes placed on a ring (0 to 2^32)
Key hashed to position on ring
Key assigned to next node clockwise

Adding a node: only ~1/N keys need to move
Removing a node: only that node's keys need to move
Virtual nodes: each physical node has 150 virtual nodes for even distribution
```

---

### Cache Patterns

**Cache-Aside (Lazy Loading):**
```go
func GetUser(id string) (User, error) {
    // 1. Check cache
    if val, err := redis.Get("user:" + id); err == nil {
        return deserialize(val), nil
    }
    // 2. Cache miss: load from DB
    user, err := db.GetUser(id)
    if err != nil { return User{}, err }
    // 3. Populate cache
    redis.Set("user:"+id, serialize(user), 1*time.Hour)
    return user, nil
}
```

**Write-Through:**
```go
func UpdateUser(user User) error {
    // Write to DB and cache simultaneously
    if err := db.UpdateUser(user); err != nil { return err }
    redis.Set("user:"+user.ID, serialize(user), 1*time.Hour)
    return nil
}
// Cache always consistent with DB
// Write latency slightly higher (two writes)
```

**Write-Behind (Write-Back):**
```go
// Write to cache immediately, async write to DB
func UpdateUser(user User) error {
    redis.Set("user:"+user.ID, serialize(user), 1*time.Hour)
    kafka.Publish("user.updated", user)  // Async DB write
    return nil  // Returns immediately
}
// Fastest writes, but risk of data loss if cache fails before DB write
```

---

### Trade-offs

| Pattern | Read Perf | Write Perf | Consistency | Data Loss Risk |
|---|---|---|---|---|
| Cache-Aside | Good (after warm) | Normal | Eventual | Low |
| Write-Through | Good | Slower (2 writes) | Strong | None |
| Write-Behind | Good | Fastest | Eventual | High |

---

### Interview Follow-up Questions
- "How does Redis handle persistence (RDB vs AOF)?"
- "What is the difference between Redis Cluster and Redis Sentinel?"
- "How do you handle cache invalidation in a distributed system?"
- "What is the difference between LRU and LFU eviction? When would you use each?"

---

### Key Takeaways

1. **In-memory = sub-millisecond** — all data in RAM, no disk I/O. This is the only way to achieve P99 < 1ms at scale.
2. **Redis single-threaded model eliminates lock contention** — one thread processes all commands sequentially. No deadlocks, no race conditions, predictable latency.
3. **LRU eviction is the safe default** — evict least recently used keys when memory is full. Use `allkeys-lru` for general caches, `volatile-lru` when some keys must never be evicted.
4. **Cache-Aside is the most common pattern** — application checks cache, falls back to DB on miss, populates cache. Simple, flexible, works with any DB.
5. **Write-Through for strong consistency** — write to DB and cache simultaneously. Slightly slower writes but cache is always consistent.
6. **Redis Cluster for horizontal scale** — 16,384 hash slots distributed across nodes. Client library handles routing transparently.
7. **Consistent hashing minimizes resharding** — when adding/removing nodes, only ~1/N keys need to move. Virtual nodes ensure even distribution.

---

## Design 9: Rate Limiter {#design9}

### Requirements Clarification

**Functional Requirements:**
- Limit requests per user/IP/API key
- Multiple rules: 100 req/min per user, 1000 req/min per IP
- Return 429 Too Many Requests with Retry-After header
- Different limits for different endpoints
- Whitelist/blacklist support

**Non-Functional Requirements:**
- Works across multiple servers (distributed)
- <1ms overhead per request
- 99.99% availability (rate limiter failure = allow all or deny all?)
- Accurate to within 0.1%

---

### Capacity Estimation

```
API traffic: 100M requests/sec (large API gateway)
Rate limit checks: 1 check per request = 100M checks/sec

Per rate limit check:
  - 1 Redis operation (INCR or Lua script)
  - Redis throughput: ~100K ops/sec per node
  - Nodes needed: 100M / 100K = 1,000 nodes (too many!)

Optimization: Local cache + Redis sync
  - Check local in-memory counter first (0.01ms, no network)
  - Sync to Redis every 100ms (batch)
  - Reduces Redis ops by 100x: 1M ops/sec → 10 Redis nodes

Storage per rate limit key:
  - Key: "ratelimit:{user_id}:{window}" = ~50 bytes
  - Value: counter = 8 bytes
  - TTL metadata: ~20 bytes
  - Total: ~80 bytes per key

For 10M active users × 2 windows (current + previous):
  10M × 2 × 80 bytes = 1.6GB (fits in single Redis node)

Redis nodes needed:
  Storage: 1 node (1.6GB)
  Throughput: 3-5 nodes (1M ops/sec after local caching)
  HA: 3 nodes minimum (1 primary + 2 replicas per shard)
```

---

### Algorithm Comparison

**Token Bucket:**
```
Bucket holds N tokens. Refills at R tokens/sec.
Each request consumes 1 token.
If bucket empty: reject request.

State: { tokens: float, last_refill: timestamp }

Allow burst up to N requests, then sustained rate of R/sec.
Best for: APIs that want to allow short bursts.
```

**Sliding Window Counter (Recommended):**
```
Approximate sliding window using two fixed windows:

current_window_count + (previous_window_count × overlap_percentage)

Example: Limit 100 req/min
  Current minute (40% elapsed): 30 requests
  Previous minute: 80 requests
  Overlap: 60% of previous minute is in our window
  
  Estimated count = 30 + (80 × 0.6) = 30 + 48 = 78
  78 < 100: Allow request

Error rate: ~0.003% (negligible for rate limiting)
Memory: O(1) — only 2 counters per key
```

---

### Distributed Implementation

```go
// Redis Lua script — atomic sliding window rate limiter
const rateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])  -- window in seconds
local now = tonumber(ARGV[3])

local current_window = math.floor(now / window)
local prev_window = current_window - 1

local current_key = key .. ":" .. current_window
local prev_key = key .. ":" .. prev_window

-- Get counts
local current_count = tonumber(redis.call('GET', current_key) or 0)
local prev_count = tonumber(redis.call('GET', prev_key) or 0)

-- Calculate overlap
local elapsed_in_window = now % window
local overlap = 1 - (elapsed_in_window / window)

-- Estimate total requests in sliding window
local estimated = current_count + (prev_count * overlap)

if estimated >= limit then
    -- Calculate retry after
    local retry_after = window - elapsed_in_window
    return {0, retry_after}  -- rejected
end

-- Increment current window counter
redis.call('INCR', current_key)
redis.call('EXPIRE', current_key, window * 2)

return {1, 0}  -- allowed
`

func IsAllowed(userID string, limit int, windowSec int) (bool, int) {
    key := "ratelimit:" + userID
    now := time.Now().Unix()
    
    result, err := redis.Eval(ctx, rateLimitScript,
        []string{key}, limit, windowSec, now).Slice()
    
    if err != nil {
        // Redis down: fail open (allow) or fail closed (deny)?
        // For most APIs: fail open (don't block users due to rate limiter failure)
        return true, 0
    }
    
    allowed := result[0].(int64) == 1
    retryAfter := int(result[1].(int64))
    return allowed, retryAfter
}
```

---

### Middleware Integration

```go
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Identify the client
        userID := getUserID(r)  // From JWT or API key
        if userID == "" {
            userID = getClientIP(r)  // Fall back to IP
        }
        
        // Check rate limit
        allowed, retryAfter := IsAllowed(userID, 100, 60)
        
        // Set rate limit headers (always, even on success)
        w.Header().Set("X-RateLimit-Limit", "100")
        w.Header().Set("X-RateLimit-Remaining", getRemainingCount(userID))
        w.Header().Set("X-RateLimit-Reset", getWindowResetTime())
        
        if !allowed {
            w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
            http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

### Multi-Tier Rate Limiting

```
Tier 1: IP-based (prevent DDoS)
  - 1000 req/min per IP
  - Checked at load balancer / API gateway

Tier 2: User-based (fair use)
  - Free tier: 100 req/min
  - Pro tier: 1000 req/min
  - Enterprise: 10000 req/min
  - Checked at application layer

Tier 3: Endpoint-based (protect expensive operations)
  - /api/search: 10 req/min (expensive)
  - /api/export: 5 req/hour (very expensive)
  - /api/profile: 100 req/min (cheap)
```

---

### Interview Follow-up Questions
- "How do you handle rate limiting across multiple data centers?"
- "What is the difference between rate limiting and throttling?"
- "How do you implement rate limiting without Redis (no external dependency)?"
- "How do you handle clients that retry aggressively after getting 429?"

---

### Trade-offs

| Algorithm | Accuracy | Memory | Burst Handling | Complexity |
|---|---|---|---|---|
| Fixed Window | Low (boundary burst) | O(1) | Poor | Simple |
| Sliding Window Log | Perfect | O(requests) | Good | Medium |
| Sliding Window Counter | ~99.997% | O(1) | Good | Medium |
| Token Bucket | Perfect | O(1) | Configurable | Medium |
| Leaky Bucket | Perfect | O(queue) | None (smoothed) | Medium |

| Scope | Pros | Cons |
|---|---|---|
| Per IP | Simple, no auth needed | Shared IPs (NAT) affect multiple users |
| Per user | Fair, per-account limits | Requires authentication |
| Per endpoint | Protects expensive ops | More rules to manage |
| Distributed (Redis) | Accurate across servers | Redis dependency, +0.5ms latency |
| Local only | Zero latency overhead | Bypassable by distributing requests |

---

### Key Takeaways

1. **Distributed rate limiting requires a shared store** — local counters are bypassable by distributing requests across servers. Redis is the standard shared counter.
2. **Sliding window counter is the best balance** — O(1) memory, ~0.003% error rate, no boundary burst problem. Use it over fixed window.
3. **Token bucket allows controlled bursting** — better UX than hard cutoffs. A user can burst 100 requests instantly, then is limited to the refill rate.
4. **Lua scripts for atomicity** — check-then-increment must be atomic. Redis Lua scripts execute atomically, preventing race conditions.
5. **Always return rate limit headers** — `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`. Clients need these to implement backoff.
6. **Layer rate limits** — per IP (DDoS protection) + per user (fair use) + per endpoint (protect expensive operations). Each layer serves a different purpose.
7. **Fail open vs fail closed** — if Redis is down, decide upfront: allow all requests (fail open, risk abuse) or reject all (fail closed, risk outage). Most APIs fail open.

---

## Design 10: Web Crawler {#design10}

### Requirements Clarification

**Functional Requirements:**
- Crawl the web starting from seed URLs
- Download and store web page content
- Extract links and add to crawl queue
- Respect robots.txt
- Handle duplicate URLs
- Re-crawl pages periodically (freshness)

**Non-Functional Requirements:**
- Crawl 1B pages per month
- Politeness: max 1 request/sec per domain
- Distributed: run on 100+ machines
- Handle failures gracefully (retry)
- Store 1B pages of content

---

### Capacity Estimation

```
Target: 1B pages/month = 385 pages/sec
Average page size: 100KB
Storage: 1B × 100KB = 100TB/month

With 100 crawler machines:
  Each machine: 385/100 = ~4 pages/sec
  Each machine: 4 × 100KB = 400KB/sec bandwidth (very manageable)

URL frontier size:
  Web has ~50B pages
  Queue: 50B URLs × 100 bytes = 5TB (need distributed queue)
```

---

### Architecture

```
                    ┌──────────────────────────────────────┐
                    │         URL Frontier (Priority Queue) │
                    │         Kafka or Redis Sorted Set     │
                    └──────────────┬───────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
    ┌─────────▼──────┐   ┌─────────▼──────┐   ┌────────▼───────┐
    │  Crawler 1      │   │  Crawler 2      │   │  Crawler N     │
    │  (fetch pages)  │   │  (fetch pages)  │   │  (fetch pages) │
    └────────┬────────┘   └────────┬────────┘   └────────┬───────┘
             │                     │                      │
             └─────────────────────┼──────────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
    ┌─────────▼──────┐   ┌─────────▼──────┐   ┌────────▼───────┐
    │  URL Extractor  │   │  Content Store  │   │  Seen URL      │
    │  (parse HTML,   │   │  (S3 + index)   │   │  Filter        │
    │   find links)   │   │                 │   │  (Bloom filter)│
    └────────┬────────┘   └─────────────────┘   └────────────────┘
             │
             ▼
    ┌─────────────────┐
    │  URL Normalizer  │
    │  + Deduplicator  │
    │  + Prioritizer   │
    └────────┬─────────┘
             │
             ▼
    Back to URL Frontier
```

---

### Deduplication: Bloom Filter

```
Problem: 50B URLs — cannot store all in memory to check duplicates
  50B × 100 bytes = 5TB (too large for RAM)

Solution: Bloom Filter
  - Probabilistic data structure
  - Space: 50B URLs → ~60GB (1.2 bytes/URL)
  - False positive rate: 0.1% (occasionally re-crawl a page — acceptable)
  - False negative rate: 0% (never miss a new URL)

How it works:
  - Array of M bits, initialized to 0
  - K hash functions
  - Add URL: set bits at positions hash1(url), hash2(url), ..., hashK(url)
  - Check URL: if ALL K positions are 1 → probably seen (may be false positive)
               if ANY position is 0 → definitely not seen

For 50B URLs with 0.1% false positive rate:
  M = 50B × 14.4 bits = 90GB (fits in distributed Redis)
  K = 10 hash functions
```

---

### Politeness: Domain Rate Limiting

```
Robots.txt compliance:
  - Fetch robots.txt for each domain before crawling
  - Cache robots.txt for 24 hours
  - Respect Crawl-delay directive
  - Respect Disallow rules

Per-domain rate limiting:
  redis.SETEX("crawl_delay:{domain}", delay_seconds, "1")
  
  Before fetching URL:
    if redis.EXISTS("crawl_delay:{domain}"):
        wait or skip
    else:
        fetch page
        redis.SETEX("crawl_delay:{domain}", 1, "1")  -- 1 req/sec default

URL Frontier partitioned by domain:
  Each crawler "owns" a set of domains
  Ensures politeness without coordination between crawlers
```

---

### URL Prioritization

```
Not all pages are equally important. Prioritize:
  1. High PageRank domains (news sites, Wikipedia)
  2. Recently updated pages (sitemap.xml lastmod)
  3. Pages with many inbound links
  4. Pages not crawled recently (freshness)

Priority queue:
  Redis Sorted Set: ZADD frontier priority url
  Higher score = higher priority
  
  Score calculation:
    score = domain_rank × 0.4
           + freshness_score × 0.3
           + inbound_links × 0.2
           + recrawl_urgency × 0.1
```

---

### Trade-offs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Deduplication | Bloom filter | Hash set in DB | Memory efficient, O(1) |
| URL queue | Kafka | Redis | Durable, partitioned by domain |
| Content storage | S3 | HDFS | Managed, cheap, scalable |
| Politeness | Per-domain Redis key | Sleep in crawler | Distributed, no coordination |
| Parsing | Distributed workers | In-crawler | Separate concerns, scale independently |

---

### Interview Follow-up Questions
- "How do you handle dynamic content (JavaScript-rendered pages)?"
- "How do you detect and handle crawler traps (infinite URL spaces)?"
- "How would you implement PageRank calculation on the crawled graph?"
- "How do you handle duplicate content (same page, different URLs)?"
- "How do you prioritize re-crawling — which pages need to be refreshed most often?"

---

### Key Takeaways

1. **Bloom filter for deduplication** — O(1) membership check, 1.2 bytes per URL. 50B URLs = ~60GB vs 5TB for a hash set. Accept 0.1% false positive rate.
2. **Politeness is mandatory** — crawling too fast gets your IPs banned. Respect `robots.txt` and `Crawl-delay`. One request per second per domain is the standard.
3. **Partition URL frontier by domain** — each crawler owns a set of domains. Ensures politeness without cross-crawler coordination.
4. **Priority queue for crawl order** — not all pages are equal. High-PageRank domains, recently updated pages, and pages with many inbound links should be crawled first.
5. **Parallel segment processing** — split large pages into segments, process in parallel. Scales horizontally by adding crawler machines.
6. **Handle crawler traps** — infinite URL spaces (calendars, session IDs, infinite scroll). Limit crawl depth, detect URL patterns, set max URLs per domain.
7. **Store raw content in S3** — cheap, durable, scalable. Index metadata (URL, title, links) in a database for fast lookups.

---

## System Design Cheat Sheet

### Back-of-Envelope Numbers (Memorize These)

```
Time:
  1 year ≈ 31.5M seconds ≈ 3 × 10^7 seconds
  1 day  = 86,400 seconds ≈ 10^5 seconds
  1 hour = 3,600 seconds

Traffic conversions:
  1M req/day  = ~12 req/sec
  10M req/day = ~116 req/sec
  1B req/day  = ~11,600 req/sec

Storage:
  1KB = 10^3 bytes
  1MB = 10^6 bytes
  1GB = 10^9 bytes
  1TB = 10^12 bytes
  1PB = 10^15 bytes

Typical sizes:
  Tweet: 300 bytes
  User record: 1KB
  Photo (compressed): 300KB
  Video (1 min, 720p): 50MB
  Web page: 100KB
```

### Technology Selection Guide

```
Use PostgreSQL when:
  - Complex queries, joins, transactions
  - Strong consistency required
  - Data fits on one server or small cluster
  - Relational data model

Use Cassandra when:
  - Write-heavy (>10K writes/sec)
  - Time-series data
  - No complex joins
  - Need linear horizontal scale

Use Redis when:
  - Caching (sub-millisecond reads)
  - Session storage
  - Rate limiting counters
  - Pub/sub messaging
  - Leaderboards (sorted sets)
  - Geospatial queries

Use Elasticsearch when:
  - Full-text search
  - Log aggregation
  - Complex filtering and facets
  - Analytics queries

Use Kafka when:
  - High-throughput event streaming
  - Fan-out to multiple consumers
  - Event sourcing
  - Decoupling services

Use S3 when:
  - Large binary files (images, videos, backups)
  - Petabyte-scale storage
  - Infrequent access patterns

Use CDN when:
  - Static assets (JS, CSS, images)
  - Video streaming
  - Global user base
  - High read traffic
```

### Common Patterns Summary

```
Fan-out on Write (Push):
  ✅ Fast reads (pre-computed)
  ❌ Expensive writes for users with many followers
  Use when: Read-heavy, followers < 1M

Fan-out on Read (Pull):
  ✅ Simple writes
  ❌ Slow reads (compute on demand)
  Use when: Write-heavy, or celebrity accounts

Hybrid:
  Regular users: push
  Celebrities: pull
  Use when: Mixed follower counts (Twitter, Instagram)

Pre-signed URLs:
  ✅ Offload large file uploads from your servers
  ✅ Scales to any upload volume
  Use when: File uploads (photos, videos, documents)

Event Sourcing:
  ✅ Complete audit trail
  ✅ Rebuild state from events
  ❌ Complex queries, eventual consistency
  Use when: Financial systems, audit requirements

CQRS:
  ✅ Optimize reads and writes independently
  ❌ Eventual consistency, complexity
  Use when: Read/write patterns are very different
```

### Interview Scoring Rubric

**What interviewers look for:**

| Area | Junior | Senior | Staff/Principal |
|---|---|---|---|
| Requirements | Asks some questions | Clarifies all ambiguities | Identifies hidden requirements |
| Estimation | Rough numbers | Accurate with reasoning | Identifies bottlenecks from numbers |
| Design | Basic components | Handles scale, failure | Trade-offs, alternatives, evolution |
| Deep dive | Describes components | Explains internals | Optimizes, handles edge cases |
| Communication | Explains what | Explains why | Drives conversation, asks good questions |

**Red flags:**
- Jumping to solution without requirements
- No capacity estimation
- Single point of failure in design
- No discussion of failure modes
- "I'd use microservices" without justification
- No trade-off discussion

**Green flags:**
- "Let me clarify the scale before designing"
- "This creates a bottleneck at X, here's how I'd solve it"
- "We could do A or B — A is simpler but B scales better"
- "What if this service goes down?"
- "Let me estimate the storage/bandwidth first"

---

*Document covers: URL Shortener, Twitter, Instagram, WhatsApp, YouTube, Uber, Netflix Recommendations, Distributed Cache, Rate Limiter, Web Crawler*

*Each design includes: Requirements → Estimation → Architecture → Database Schema → Key Algorithms → Trade-offs → Follow-up Questions*
