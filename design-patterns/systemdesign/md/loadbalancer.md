# Load Balancer

## What is a Load Balancer?

A load balancer is a component (hardware or software) that sits between clients and a pool of backend servers, distributing incoming network traffic across those servers. Think of it as a traffic cop for your infrastructure.

It ensures no single server gets overwhelmed while others sit idle.

## What Does "Load" Mean?

"Load" is not limited to reads or writes — it refers to any type of incoming request or connection:

- HTTP/HTTPS requests (reads, writes, uploads, downloads)
- TCP/UDP connections
- Database queries (read and write)
- WebSocket connections
- gRPC calls
- Streaming traffic

Essentially, any work that a server has to process counts as load. The balancer doesn't typically distinguish between read and write — it distributes requests based on its configured algorithm, regardless of the operation type. That said, some architectures do split read vs write traffic intentionally (e.g., database read replicas), but that's an architectural decision, not a load balancer limitation.

## Load Balancing Algorithms

| Algorithm | How It Works |
|---|---|
| Round Robin | Requests go to each server in sequence: A → B → C → A → B → C... Simple and stateless. |
| Weighted Round Robin | Same as round robin, but servers with higher weight get proportionally more requests. Useful when servers have different capacities. |
| Least Connections | Routes to the server with the fewest active connections. Good when request processing times vary. |
| Weighted Least Connections | Combines least connections with server weights. |
| IP Hash | Hashes the client IP to consistently route the same client to the same server. Provides sticky sessions without cookies. |
| Least Response Time | Routes to the server with the fastest response time and fewest active connections. |
| Random | Picks a server at random. Surprisingly effective at scale due to the power of random distribution. |
| Consistent Hashing | Maps both servers and requests onto a hash ring. Minimizes redistribution when servers are added/removed. Common in caches and distributed systems. |
| Resource-Based (Adaptive) | Queries servers for current load (CPU, memory) and routes to the least loaded. More overhead but more accurate. |

## Other Responsibilities Beyond Balancing

Load balancers do a lot more than just distribute traffic:

- SSL/TLS Termination — decrypt HTTPS at the balancer so backends deal with plain HTTP, offloading crypto overhead
- Health Checks — continuously probe backends and stop sending traffic to unhealthy ones
- Session Persistence (Sticky Sessions) — ensure a user's requests consistently go to the same backend when needed
- Rate Limiting — throttle excessive requests from specific clients
- DDoS Protection — absorb or filter malicious traffic before it reaches backends
- Compression — gzip/brotli responses before sending to clients
- Caching — serve cached responses for common requests without hitting backends
- Request Routing — route based on URL path, headers, cookies (Layer 7 routing)
- Connection Pooling — maintain persistent connections to backends, reducing connection overhead
- Logging and Monitoring — centralized access logs, metrics, and tracing
- Failover — automatic rerouting when a backend or even an entire data center goes down
- A/B Testing and Canary Deployments — route a percentage of traffic to new versions

## Well-Known Load Balancers

### Software

- NGINX — the workhorse of the internet, reverse proxy + load balancer
- HAProxy — high-performance TCP/HTTP load balancer, battle-tested
- Envoy — modern L7 proxy, backbone of service meshes (Istio)
- Traefik — cloud-native, auto-discovers services in Docker/Kubernetes
- Caddy — simple config, automatic HTTPS

### Cloud-Managed

- AWS Elastic Load Balancer (ALB, NLB, GLB)
- Google Cloud Load Balancing
- Azure Load Balancer / Application Gateway
- Cloudflare Load Balancing

### Hardware (legacy but still in use)

- F5 BIG-IP
- Citrix ADC (NetScaler)

## How DNS Load Balancing Works

DNS-based load balancing is one of the simplest forms. When a client resolves `example.com`, the DNS server returns multiple IP addresses:

```
example.com → 192.168.1.1, 192.168.1.2, 192.168.1.3
```

The DNS server rotates the order of IPs on each query (DNS round robin). The client typically connects to the first IP in the list.

More advanced DNS load balancing (like Route 53, Cloudflare DNS) can:

- Return different IPs based on the client's geographic location (geo-routing)
- Perform health checks and remove unhealthy IPs from responses
- Use weighted routing to send more traffic to certain IPs
- Use latency-based routing to direct users to the nearest/fastest server

Limitations of DNS load balancing:

- DNS responses are cached (TTL), so changes aren't instant
- No awareness of server load or connection count
- Clients may ignore IP ordering
- Granularity is coarse — you can't balance per-request, only per-resolution

## How CDN Load Balancing Works

CDNs like Cloudflare, Akamai, and CloudFront operate a global network of edge servers. Their load balancing is multi-layered:

1. **DNS-level routing** — when a user resolves a CDN-hosted domain, the CDN's DNS returns the IP of the nearest/best edge node using anycast or geo-DNS
2. **Anycast routing** — multiple edge servers share the same IP address. BGP routing naturally directs packets to the closest server in terms of network hops
3. **Edge-level balancing** — within a PoP (Point of Presence), traditional load balancing (round robin, least connections) distributes requests across servers in that location
4. **Origin shielding** — if the edge doesn't have cached content, the CDN picks an optimal path back to the origin server, often through an intermediate "shield" layer to reduce origin load
5. **Health-aware routing** — if an edge node or PoP goes down, traffic is automatically rerouted to the next closest healthy location

The combination of anycast + geo-DNS + health checks is what makes CDNs feel fast everywhere — users are always hitting a server that's geographically and network-topologically close to them.

---

In short: a load balancer is the unsung hero of any scalable system. It's not just about spreading requests around — it's about reliability, performance, security, and operational flexibility. Every production system of meaningful scale has load balancing at multiple layers.
