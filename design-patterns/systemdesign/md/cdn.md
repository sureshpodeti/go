# CDN (Content Delivery Network) — Quick Revision Guide

## What is a CDN?

A geographically distributed network of proxy servers and data centers. Goal: serve content to users from the nearest physical location to reduce latency.

---

## What Kind of Data is Stored in a CDN?

CDNs cache and serve anything deliverable over HTTP(S):

- Static assets: HTML, CSS, JavaScript, fonts
- Images: JPEG, PNG, WebP, SVG, AVIF
- Video and audio files (full files or chunks)
- API responses (when configured for dynamic caching)
- Software downloads, PDFs, documents
- Live and on-demand streaming segments (HLS/DASH manifests and chunks)

---

## How Does a CDN Store Data? Is It a File System?

**No.** Not a traditional file system like ext4 or APFS. CDN edge servers use highly optimized storage layers:

- **In-memory caches** (RAM-based LRU caches) — fastest tier
- **SSD-backed disk caches** — for warm content
- Content is keyed by **URL + headers** (cache key), not by file path
- Think of it as a **distributed key-value store**: key = request URL, value = response body + headers

### Image Optimization

Modern CDNs (Cloudflare, CloudFront, Akamai) do on-the-fly transformations — resizing, format conversion (WebP/AVIF), quality adjustment — and cache each variant separately:

```
/image.jpg?w=800&fmt=webp  → cached variant 1
/image.jpg?w=400&fmt=avif  → cached variant 2
/image.jpg?w=200&fmt=jpeg  → cached variant 3
```

More efficient than storing every possible size upfront.

---

## How Files Are Stored In-Memory (Deep Dive)

There's no "file" in memory. The image/video is stored as a **raw byte array (buffer)** in RAM.

### What a Cache Entry Looks Like in RAM

```
┌─────────────────────────────────────────────┐
│  Cache Entry (in RAM)                       │
│                                             │
│  Key:   "GET /cat.jpg"                      │
│  Metadata:                                  │
│    Content-Type: image/jpeg                 │
│    Content-Length: 245760                    │
│    ETag: "abc123"                           │
│    Cache-Control: max-age=86400             │
│    Expires-At: 1713200000                   │
│    Last-Accessed: 1713113600                │
│                                             │
│  Body:  [0xFF 0xD8 0xFF 0xE0 ... 245760    │
│          bytes of raw JPEG data in a        │
│          contiguous memory buffer]          │
└─────────────────────────────────────────────┘
```

It's a key-value pair: metadata + a byte buffer. No file system involved.

### Memory Organization — Slab Allocator

Most caching systems use a **slab allocator** pattern (similar to memcached):

```
RAM Pool (e.g., 64 GB allocated to cache)
├── Slab Class 1:  items up to 1 KB
├── Slab Class 2:  items up to 4 KB
├── Slab Class 3:  items up to 16 KB
├── Slab Class 4:  items up to 64 KB
├── Slab Class 5:  items up to 256 KB   ← a 240KB image lands here
├── Slab Class 6:  items up to 1 MB
└── Slab Class 7:  items up to 4 MB
```

- RAM is pre-divided into fixed-size slabs to **avoid fragmentation**
- A 240KB JPEG goes into the "up to 256KB" slab class
- A **hash table** (in memory) maps cache keys → slab locations
- Lookup is **O(1)** — hash the URL, find the pointer, read the bytes

For large objects (video segments, 2-10MB), some systems use a **hybrid approach** — metadata + first chunk in RAM, rest on SSD.

### Eviction — When RAM Fills Up

RAM is finite. Eviction policies decide what gets removed:

- **LRU (Least Recently Used)** — most common. Longest-untouched item evicted first
- **LFU (Least Frequently Used)** — rarely accessed items evicted, even if recently touched
- **TTL-based** — items expire after their Cache-Control max-age

In practice, CDNs use a combination: hot content stays in RAM, warm content demoted to SSD, cold content evicted entirely.

### Cache Lookup Speed Comparison

```
Request comes in for /cat.jpg
        ↓
Hash lookup in memory index → O(1)
        ↓
   Found in RAM? ──Yes──→ Read bytes directly from RAM buffer
        ↓ No                (~100 nanoseconds)
   Found on SSD? ──Yes──→ Read from SSD, promote to RAM
        ↓ No                (~100 microseconds)
   Fetch from origin ────→ Cache in RAM + SSD
                            (~50-200 milliseconds)
```

**Speed difference:** RAM is ~1,000x faster than SSD, and ~1,000,000x faster than a network round-trip to origin.

### Why This Is So Fast

When a user requests an image:
1. Cache daemon receives the HTTP request
2. Hashes the URL → O(1) hash table lookup
3. Reads byte buffer directly from RAM — no disk I/O, no file system, no syscalls
4. Writes bytes straight to the network socket

No "opening a file." No file system. It's **pointer → bytes → socket**.

---

## CDN Streaming — HLS & DASH

**Yes, CDNs enable streaming.** This is exactly how YouTube, Netflix, and Twitch work.

### Two Main Protocols

| Protocol | Full Name | Notes |
|----------|-----------|-------|
| HLS | HTTP Live Streaming | Apple's protocol, most widely supported |
| DASH | Dynamic Adaptive Streaming over HTTP | Open standard |

### How Streaming Works

1. Origin server (or transcoding service) breaks video into **small segments** (2-10 seconds each)
2. A **manifest file** (`.m3u8` for HLS, `.mpd` for DASH) lists all segments and quality levels
3. Player fetches manifest, then requests segments one by one
4. Each segment is a regular HTTP request — perfect for CDN caching

```
video/
  manifest.m3u8          ← playlist file
  segment_001_720p.ts    ← 2-sec chunk at 720p
  segment_001_1080p.ts   ← same 2 sec at 1080p
  segment_002_720p.ts
  segment_002_1080p.ts
  ...
```

### Adaptive Bitrate Streaming (ABR)

Player monitors bandwidth and switches quality on the fly:
- Good connection → 1080p segments
- Bandwidth drops → switches to 720p or 480p
- Recovers → back to higher quality

### Live Streaming

Same approach — segments generated in real-time, pushed to CDN with very short TTLs. Edge nodes cache and serve them to thousands of concurrent viewers.

### Why CDNs Are Perfect for Streaming

Video segments are small (2-second HLS chunk at 1080p ≈ 1-4MB). Popular segments stay hot in RAM across edge nodes. A viral video's first few segments can be served from RAM millions of times without touching disk.

---

## How Does the CDN Get Data? (Origin Pull)

The most common model is **origin pull**:

```
User Request → CDN Edge Node
                  ↓
            Cache HIT? → Yes → Serve immediately (fast)
                  ↓ No
            Check regional/shield cache
                  ↓ No
            Fetch from Origin Server (S3, app server, etc.)
                  ↓
            Cache the response at edge (with TTL)
                  ↓
            Serve to user
```

### Cache Miss Flow

1. Edge node checks if a nearby **regional/shield cache** has it
2. If not, pulls from the **origin server**
3. Response cached at edge with a **TTL** (Time To Live) you configure
4. Subsequent requests from that region served from cache

### Push CDN (Less Common)

You proactively upload content to CDN nodes before anyone requests it. Useful for major launches or live events.

### Cache Invalidation

- **Purge by URL** — invalidate specific resource
- **Purge by tag/surrogate key** — group related content
- **Wildcard purge** — `/images/*`
- **Versioned URLs** — `style.v3.css` (most reliable, old URLs naturally expire)

---

## Expert-Level CDN Q&A

### Q: What is a cache key and how do you control it?
The cache key determines what makes a request "unique." Default = URL. You can include/exclude query params, headers, cookies. Misconfigured cache keys cause: serving stale content OR cache fragmentation (too many variants).

### Q: What is cache stampede (thundering herd)?
When a popular cached item expires, hundreds of simultaneous requests hit origin. Solutions:
- **Request coalescing** — CDN groups identical requests, makes one origin fetch
- **Stale-while-revalidate** — serve stale content while refreshing in background
- **Lock-based refresh** — only one request fetches, others wait

### Q: Push CDN vs Pull CDN?
| Type | Behavior | Best For |
|------|----------|----------|
| Pull | Fetches from origin on demand | Most use cases, simpler |
| Push | You upload content proactively | Large/predictable content, launches |

### Q: How does CDN handle dynamic/personalized content?
- **Edge computing** (Cloudflare Workers, Lambda@Edge) — run code at CDN nodes
- **Vary headers** — cache different versions per header value
- **Edge Side Includes (ESI)** — assemble pages from cached fragments

### Q: What is origin shielding?
An intermediate cache layer between edge nodes and origin. Instead of 200 edge nodes each hitting origin on a miss, they all go through **one shield node** — dramatically reduces origin load.

### Q: What is anycast and why do CDNs use it?
Multiple servers share the same IP address. Network routes users to the nearest server automatically based on topology. No complex DNS logic needed.

### Q: How do CDNs handle SSL/TLS?
CDNs **terminate TLS at the edge** (handshake is fast, close to user). Connection to origin can be separate TLS or plain HTTP over private backbone. Huge latency win since TLS handshakes involve multiple round trips.

### Q: What are the important CDN cache headers?

| Header | Purpose |
|--------|---------|
| `Cache-Control` | max-age, s-maxage, no-cache, no-store |
| `Vary` | Cache variants by header (Accept-Encoding, etc.) |
| `ETag` / `Last-Modified` | Conditional requests (revalidation) |
| `CDN-Cache-Control` | CDN-specific directives |
| `Surrogate-Key` | Tag-based purging |

### Q: How do CDNs protect against DDoS?
- Absorb traffic across hundreds of **PoPs** (Points of Presence)
- **Rate limiting** at the edge
- **WAF** (Web Application Firewall) rules
- **Bot detection** and challenge pages
- All before traffic reaches your origin

### Q: What is stale-while-revalidate?
A Cache-Control directive: `Cache-Control: max-age=3600, stale-while-revalidate=60`. After max-age expires, serve stale content for up to 60 more seconds while fetching fresh content in the background. Users never see a slow response.

### Q: How do multi-CDN setups work?
Large platforms use multiple CDN providers (Akamai + CloudFront + Fastly) with a **traffic manager** or **DNS-based routing** to:
- Failover if one CDN has issues
- Route to cheapest/fastest provider per region
- Avoid vendor lock-in

---

## CDN Architecture — Big Picture

```
                    ┌──────────────┐
                    │ Origin Server│ (S3, App Server, etc.)
                    └──────┬───────┘
                           │
                    ┌──────┴───────┐
                    │ Origin Shield│ (intermediate cache)
                    └──────┬───────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
   ┌──────┴──────┐ ┌──────┴──────┐ ┌──────┴──────┐
   │ Edge PoP    │ │ Edge PoP    │ │ Edge PoP    │
   │ (US-East)   │ │ (EU-West)   │ │ (Asia-Pac)  │
   └──────┬──────┘ └──────┬──────┘ └──────┬──────┘
          │                │                │
     US Users         EU Users        Asia Users
```

Each Edge PoP contains:
- RAM cache (hot content)
- SSD cache (warm content)
- Load balancers
- TLS termination
- Optional edge compute runtime
