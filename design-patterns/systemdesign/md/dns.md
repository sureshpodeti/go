# DNS & Networking — Quick Revision Guide

## CAP Theorem

A distributed system can only guarantee **2 out of 3** properties simultaneously:

- **Consistency** — every read gets the most recent write
- **Availability** — every request gets a non-error response
- **Partition Tolerance** — system works despite network partitions

Network partitions are unavoidable, so the real trade-off is **C vs A** during a partition:

| Type | Example | Behavior during partition |
|------|---------|--------------------------|
| CP | MongoDB, HBase, etcd | Sacrifices availability to maintain consistency |
| AP | Cassandra, DynamoDB, CouchDB | Sacrifices consistency to stay available |
| CA | Only theoretical | Requires no partitions (single node) |

**MongoDB (CP):** Uses replica set with single primary. During partition, minority side loses write availability. Majority side elects new primary and keeps going. Reads on secondaries possible with `readPreference: secondary` but may be stale.

**Cassandra (AP):** Masterless — every node accepts reads/writes. At `ONE` consistency, stays available during partitions but may serve stale data. At `QUORUM`, starts behaving like CP — requests fail if quorum unreachable. Trade-off is per-query, not system-wide.

**PACELC** extends CAP: also considers latency vs consistency when there's no partition.

---

## DNS — Is it Read-Heavy or Write-Heavy?

**Massively read-heavy.** Billions of devices resolve domains every second. Records change rarely (a few times a year). Read-to-write ratio is easily 100,000:1+. The entire architecture is optimized for reads.

---

## Where is Domain-to-IP Mapping Stored?

Hierarchical, distributed database — no single place holds all mappings.

```
Root Name Servers (13 logical clusters: a.root-servers.net → m.root-servers.net)
    ↓ "I don't know the IP, but ask the .com TLD server"
TLD Servers (.com, .org, .io — managed by registries like Verisign)
    ↓ "Ask ns1.example.com for that domain"
Authoritative Name Servers (where actual DNS records live)
    ↓ Returns the IP address
```

Records stored as **zone files**:
```
example.com.      300   IN  A    93.184.216.34
mail.example.com. 3600  IN  MX   10 mailserver.example.com.
```

The numbers (300, 3600) are **TTL** (Time To Live) in seconds — key to how reads scale.

---

## How DNS Supports Reads at Scale

Multiple layers of caching:

1. **Browser cache** — remembers recent lookups
2. **OS-level cache** — operating system DNS cache
3. **Resolver cache** (recursive resolver like 8.8.8.8, 1.1.1.1) — caches based on TTL. Biggest win. One user resolves google.com → next million users get cached answer.
4. **Anycast routing** — same IP advertised from hundreds of physical locations. Query goes to nearest one.
5. **Replication** — authoritative servers run multiple replicas across regions.

Most DNS queries never reach the authoritative server — answered from cache within milliseconds.

---

## DNS Resolution Methods

### Recursive Resolution (most common)
- Client asks recursive resolver (e.g., 8.8.8.8): "What's the IP for api.example.com?"
- Resolver does all the work: root → TLD → authoritative
- Client makes one request, gets complete answer

### Iterative Resolution
- Client asks root → root says "ask .com TLD"
- Client asks .com TLD → says "ask ns1.example.com"
- Client asks ns1.example.com → gets final answer
- Client does the legwork at each step

**In practice:** Your device uses recursive. The resolver itself uses iterative when talking to root/TLD/authoritative servers.

---

## DNS Record Types

| Type | Purpose |
|------|---------|
| A | Domain → IPv4 |
| AAAA | Domain → IPv6 |
| CNAME | Alias to another domain (can't coexist with other records at zone apex) |
| MX | Mail routing |
| NS | Delegates subdomain to specific name servers |
| TXT | Arbitrary text (SPF, DKIM, domain verification) |
| SRV | Host + port for services (service discovery) |
| SOA | Zone metadata (refresh intervals, TTL defaults) |

---

## DNS Failure Scenarios

### DDoS on DNS Providers
- 2016 Dyn attack took down Twitter, GitHub, Netflix (Mirai botnet)
- **Fix:** Multi-homing — use multiple DNS providers (Route 53 + Cloudflare)

### Cache Poisoning / Spoofing
- Attacker injects fake DNS responses into resolver cache
- **Fix:** DNSSEC (cryptographic signing), DNS over HTTPS (DoH), DNS over TLS (DoT)

### TTL Misconfiguration
- Too high → stale records during migration, traffic goes to old IP for hours
- Too low → every request hits authoritative server, increased latency
- **Fix:** TTL lowering strategy before migrations (see below)

### Authoritative Server Failure
- If authoritative servers go down and cached TTLs expire → domain unresolvable
- **Fix:** Run 2-3+ authoritative servers across different networks/providers

### Resolver Failure
- ISP's DNS goes down
- **Fix:** Configure fallback resolvers (8.8.8.8, 1.1.1.1). Run local caching resolver (unbound, dnsmasq)

### Network/Firewall Issues
- DNS uses UDP port 53 (TCP 53 for large responses/zone transfers). Blocked port 53 silently breaks resolution.
- **Fix:** Ensure port 53 open for both UDP and TCP

---

## TTL Migration Strategy — Step by Step

**Scenario:** Migrating `api.example.com` from `10.0.1.1` to `10.0.2.2`. Current TTL is 86400 (24 hours).

**Problem:** If you just change the IP, resolvers worldwide have cached the old IP for up to 24 hours. Split traffic, potential errors.

**Solution:**

```
Day 0 (normal):     TTL=86400  IP=10.0.1.1   (business as usual)
Day 1 (prep):       TTL=60     IP=10.0.1.1   (lower TTL, same IP — no user impact)
Day 2 (wait):       TTL=60     IP=10.0.1.1   (old 86400 caches fully expired)
Day 3 (migrate):    TTL=60     IP=10.0.2.2   (switch IP — propagates in ~60s)
Day 4 (stabilize):  TTL=60     IP=10.0.2.2   (monitor, confirm everything works)
Day 5 (harden):     TTL=3600   IP=10.0.2.2   (raise TTL back up)
```

**Why lower TTL 2 days before?** Old 86400 TTL is still cached. Need to wait for those caches to expire. After 24h, every resolver picks up the new 60s TTL. Extra day = safety buffer.

**Why not keep TTL low forever?**
- More DNS queries hitting authoritative servers (cost + load)
- Higher latency for users (fresh lookup every 60s)
- More exposure to DNS outages (low-TTL clients lose resolution faster when provider has a blip)

---

## Additional DNS Topics

### DNS Load Balancing
- Round-robin: return multiple IPs for same domain
- Weighted routing: 80% traffic to region A, 20% to region B
- Not a replacement for proper load balancer (no health awareness by default)

### DNS-Based Service Discovery
- Kubernetes uses CoreDNS: `payment-service.default.svc.cluster.local`
- Consul provides DNS-based discovery
- Trade-off: DNS caching can cause stale endpoints → some teams prefer client-side discovery

### Split-Horizon DNS
- Same domain resolves to different IPs based on query source
- Internal users → private IP (10.x.x.x), external users → public IP
- Common in enterprise and hybrid cloud

### DNS Propagation
- No push mechanism. "Propagation" = cache expiration across thousands of resolvers globally
- Why TTL planning matters before any infrastructure change

### GeoDNS / Latency-Based Routing
- Route users to nearest data center by geography or measured latency
- Route 53, Cloudflare support natively
- Foundation for multi-region active-active systems

### DNS over HTTPS (DoH) / DNS over TLS (DoT)
- Traditional DNS is plaintext UDP — anyone on network can see queries
- DoH/DoT encrypt DNS queries
- Trade-off: bypasses corporate security controls

### Negative Caching
- NXDOMAIN (domain doesn't exist) responses are also cached
- New subdomain might be "doesn't exist" in some resolvers for a while

---

## TCP — Three-Way Handshake

### Why Random Sequence Numbers?
The starting number is just a label — no "missing" bytes before it. Random to prevent TCP hijacking (if always started at 0, attackers could predict and inject fake packets).

### The Handshake

```
Client                          Server
  |                               |
  |---- SYN (seq=100) ----------->|  "I want to connect. My bytes start at 100"
  |                               |
  |<--- SYN-ACK (seq=300,ack=101)-|  "Got it. My bytes start at 300. Expecting your 101 next"
  |                               |  (ack=101 because SYN consumes one sequence number)
  |                               |
  |---- ACK (ack=301) ----------->|  "Got yours. We're synced."
  |                               |
  |    Connection established      |
```

### Data Flow Example

```
Client sends 10 bytes "ABCDEFGHIJ" split into segments:

Segment 1: seq=101, "ABCD"  (4 bytes, covers 101-104)
Segment 2: seq=105, "EFG"   (3 bytes, covers 105-107)
Segment 3: seq=108, "HIJ"   (3 bytes, covers 108-110)
```

---

## TCP — How Ordering Works

Server doesn't know total segment count. It only tracks: "What's the next byte I expect?"

Packets arrive out of order:
```
Arrives first:  seq=108, "HIJ"  → Server expects 101. Buffer this.
Arrives second: seq=101, "ABCD" → That's what I expected! Accept. Now expect 105.
Arrives third:  seq=105, "EFG"  → Accept. Now expect 108. Already buffered! Accept "HIJ" too.

Result: "ABCDEFGHIJ" — perfect order.
```

---

## TCP — How No-Missing-Packets Works

### Happy Path
```
Client                              Server
  |-- seq=101, "ABCD" -------------->|
  |<------------ ack=105 ------------|  "Got 101-104, send 105 next"
  |-- seq=105, "EFG" --------------->|
  |<------------ ack=108 ------------|  "Got 105-107, send 108 next"
  |-- seq=108, "HIJ" --------------->|
  |<------------ ack=111 ------------|  "Got 108-110, done"
```

### Packet Loss — Two Detection Mechanisms

**Mechanism 1 — Timeout (RTO):**
Client starts timer per segment. No ACK within timeout → assume lost → retransmit.

**Mechanism 2 — Fast Retransmit (3 duplicate ACKs):**
Faster than waiting for timeout.

```
Client                              Server
  |-- seq=101, "ABCD" -------------->|
  |<------------ ack=105 ------------|
  |-- seq=105, "EFG" ------X         |  LOST!
  |-- seq=108, "HIJ" --------------->|
  |<------------ ack=105 ------------|  Duplicate ACK #1 — "still need 105"
  |-- seq=111, "KLM" --------------->|
  |<------------ ack=105 ------------|  Duplicate ACK #2
  |-- seq=114, "NOP" --------------->|
  |<------------ ack=105 ------------|  Duplicate ACK #3
  |                                   |
  |  3 duplicate ACKs → retransmit!   |
  |                                   |
  |-- seq=105, "EFG" --------------->|  Retransmit!
  |<------------ ack=117 ------------|  "Got everything now"
```

Server had buffered 108, 111, 114. Once 105 arrives, fills the gap, ACKs everything at once.

---

## TCP — How Does the Server Know a Message is Complete?

**TCP doesn't know.** TCP is a continuous byte stream — it has no concept of "message boundaries."

The **application layer** decides:

| Method | Example |
|--------|---------|
| Content-Length header | `Content-Length: 10` → read exactly 10 bytes |
| Chunked transfer | Zero-length chunk `0\r\n` means done |
| Delimiter | Newline `\n` means end of message |
| Length prefix | `[10]ABCDEFGHIJ` — first bytes specify length |

**Connection termination** uses the **FIN flag**:
```
Client -- FIN, seq=111 ---------> Server   "I'm done sending"
Client <-------- ack=112 ------- Server   "Got your FIN"
Client <-------- FIN, seq=301 -- Server   "I'm done too"
Client -- ack=302 --------------> Server   "Connection closed"
```

---

## TCP vs UDP

| Aspect | TCP | UDP |
|--------|-----|-----|
| Connection | Handshake required | Connectionless (fire and forget) |
| Reliability | Guaranteed delivery + retransmission | No guarantee |
| Ordering | In-order delivery | No ordering |
| Flow control | Sliding window | None |
| Congestion control | Slow start, AIMD | None |
| Header size | 20-60 bytes | 8 bytes |
| Latency | Higher (handshake + acks) | Lower (no overhead) |
| Use cases | HTTP, SSH, FTP, SMTP, DB connections | DNS, video streaming, gaming, VoIP |

**Use TCP when:** data integrity matters, ordering matters, you need reliability without building it yourself.

**Use UDP when:** speed > completeness, some data loss is OK (live streaming), high-frequency small messages (DNS queries), or you want to build custom reliability on top.

**Modern trend — QUIC (HTTP/3):** Built on UDP but adds reliability, ordering, encryption at application layer. Why not just use TCP? TCP is in the OS kernel — can't easily change. UDP + custom logic = best of both worlds (connection migration, 0-RTT handshakes, independent stream multiplexing).
