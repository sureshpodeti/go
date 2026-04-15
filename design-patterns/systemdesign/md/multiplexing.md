# Multiplexing

> Sending **multiple independent streams** of data over a **single shared connection**,
> instead of opening a separate connection for each one.

---

## The Core Idea

```
  ❌ WITHOUT Multiplexing (one connection per stream)
  ┌──────────┐                          ┌──────────┐
  │          │ ═══ Connection 1 ══════▶ │          │
  │          │ ═══ Connection 2 ══════▶ │          │
  │  Client  │ ═══ Connection 3 ══════▶ │  Server  │
  │          │ ═══ Connection 4 ══════▶ │          │
  │          │ ═══ Connection 5 ══════▶ │          │
  └──────────┘                          └──────────┘
        ⚠️ Expensive! Each connection = TCP handshake + memory + overhead


  ✅ WITH Multiplexing (all streams share one connection)
  ┌──────────┐                          ┌──────────┐
  │          │   ┌─ Stream A ─┐         │          │
  │          │   ├─ Stream B ─┤         │          │
  │  Client  │ ══╡  Stream C  ╞══════▶  │  Server  │
  │          │   ├─ Stream D ─┤         │          │
  │          │   └─ Stream E ─┘         │          │
  └──────────┘    One Connection        └──────────┘
        ✅ Efficient! Shared connection, independent streams
```

---

## 1. Multiplexing in HTTP/2

### The HTTP/1.1 Problem: Head-of-Line Blocking

```
  HTTP/1.1 — One request must complete before the next starts (per connection)

  Connection 1:  ██ style.css ██████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
  Connection 2:  ██ app.js ████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░
  Connection 3:  ██ logo.png ██████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
  Connection 4:  ░░░░░░░░░░░░░░░░░░░░██ data.json ██████████░░░░░░░░░░░░
  Connection 5:  ░░░░░░░░░░░░░░░░░░░░██ font.woff ████░░░░░░░░░░░░░░░░░░
  Connection 6:  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░██ icon.svg ██████░░░░░
                 ─────────────────────────────────────────────────────▶ time

  ⚠️ Browser opens 6 connections per domain (workaround, not a fix)
  ⚠️ Each connection = full TCP + TLS handshake
  ⚠️ Requests 4-6 must WAIT until a connection frees up
```

### The HTTP/2 Solution: Stream Multiplexing

```
  HTTP/2 — All requests/responses interleaved on ONE connection

  Single TCP Connection:
  ┌─────────────────────────────────────────────────────────────────┐
  │ ▓▓css▓▓ ░░js░░ ▒▒png▒▒ ░░js░░ ▓▓css▓▓ ██json██ ░░js░░ ▒▒png▒▒│
  └─────────────────────────────────────────────────────────────────┘
                                                                 ▶ time

  How it works — Frames with Stream IDs:
  ┌──────────┬──────────┬──────────┬──────────┬──────────┐
  │ Stream 1 │ Stream 2 │ Stream 3 │ Stream 1 │ Stream 2 │  ...
  │ css data │ js data  │ png data │ css data │ js data  │
  └──────────┴──────────┴──────────┴──────────┴──────────┘

  ✅ One connection, one TLS handshake
  ✅ No head-of-line blocking (at HTTP layer)
  ✅ Server responds in ANY order (whichever is ready first)
  ✅ Streams can have priorities
```

### HTTP/2 Frame Structure

```
  ┌─────────────────────────────────────┐
  │         HTTP/2 Frame                │
  ├──────────┬──────────┬───────────────┤
  │  Length   │   Type   │    Flags      │
  │ (24 bit) │ (8 bit)  │   (8 bit)     │
  ├──────────┴──────────┴───────────────┤
  │  Stream Identifier (31 bits)        │  ◄── This is the key!
  ├─────────────────────────────────────┤      Each stream gets a
  │                                     │      unique ID so frames
  │          Frame Payload              │      can be interleaved
  │                                     │      and reassembled.
  └─────────────────────────────────────┘
```

---

## 2. Multiplexing in Redis

### A) Client-Side: Pipelining

```
  ❌ Without Pipelining — Round trip per command

  Client                          Redis Server
    │                                  │
    │──── SET user:1 "Alice" ────────▶│
    │◀─── OK ──────────────────────────│  ⏱️ RTT
    │                                  │
    │──── GET user:2 ────────────────▶│
    │◀─── "Bob" ───────────────────────│  ⏱️ RTT
    │                                  │
    │──── INCR pageviews ────────────▶│
    │◀─── 42 ──────────────────────────│  ⏱️ RTT
    │                                  │
    Total: 3 round trips ❌


  ✅ With Pipelining — All commands sent at once

  Client                          Redis Server
    │                                  │
    │──── SET user:1 "Alice" ────────▶│
    │──── GET user:2 ────────────────▶│  📦 Batched!
    │──── INCR pageviews ────────────▶│
    │                                  │
    │◀─── OK ──────────────────────────│
    │◀─── "Bob" ───────────────────────│  📦 All responses
    │◀─── 42 ──────────────────────────│     back together
    │                                  │
    Total: 1 round trip ✅
```

### B) Server-Side: I/O Multiplexing (epoll / kqueue)

```
  Redis is SINGLE-THREADED but handles 1000s of clients. How?

  ┌─────────┐
  │Client A │──┐
  └─────────┘  │
  ┌─────────┐  │    ┌──────────────────────────────────┐
  │Client B │──┼───▶│  Event Loop (epoll / kqueue)     │
  └─────────┘  │    │                                  │
  ┌─────────┐  │    │  "Which sockets have data ready?" │
  │Client C │──┤    │                                  │
  └─────────┘  │    │   ┌─────────────────────────┐    │
  ┌─────────┐  │    │   │  Single Thread           │    │
  │Client D │──┘    │   │                         │    │
  └─────────┘       │   │  1. Read from ready fd  │    │
                    │   │  2. Process command      │    │
                    │   │  3. Write response       │    │
                    │   │  4. Next ready fd...     │    │
                    │   └─────────────────────────┘    │
                    └──────────────────────────────────┘

  The OS tells Redis: "Client A and Client C have data ready"
  Redis processes them one by one — fast enough to feel concurrent.
```

---

## Real-Life Analogy: The Restaurant Kitchen

```
  ❌ WITHOUT Multiplexing (dedicated chef per table)

  Table 1 ──▶ 👨‍🍳 Chef 1 ──▶ cooks full meal ──▶ serves
  Table 2 ──▶ 👨‍🍳 Chef 2 ──▶ cooks full meal ──▶ serves
  Table 3 ──▶ 👨‍🍳 Chef 3 ──▶ cooks full meal ──▶ serves
  Table 4 ──▶ 👨‍🍳 Chef 4 ──▶ cooks full meal ──▶ serves

  ⚠️ Expensive! Need one chef per table.


  ✅ WITH Multiplexing (one chef, interleaved work)

  ┌─────────────────────────────────────────────────────────┐
  │                    👨‍🍳 One Chef                          │
  │                                                         │
  │  ⏱️ Timeline:                                           │
  │                                                         │
  │  ▓▓ T1:Steak on grill ▓▓                               │
  │  ░░ T2:Chop salad ░░                                    │
  │  ▒▒ T3:Plate dessert ▒▒                                 │
  │  ▓▓ T1:Flip steak ▓▓                                    │
  │  ░░ T2:Dress salad ░░                                    │
  │  ▓▓ T1:Plate steak ▓▓                                   │
  │  ▒▒ T3:Pour coffee ▒▒                                    │
  │                                                         │
  │  ✅ One chef serves ALL tables by interleaving tasks     │
  │  ✅ While steak grills (I/O wait), chef does other work  │
  └─────────────────────────────────────────────────────────┘
```

---

## Another Analogy: Telecom

```
  Old Phone System (no multiplexing):
  ┌────────┐                         ┌────────┐
  │ Call 1 │════ Dedicated Wire 1 ══▶│        │
  │ Call 2 │════ Dedicated Wire 2 ══▶│ Switch │
  │ Call 3 │════ Dedicated Wire 3 ══▶│        │
  └────────┘                         └────────┘
  ⚠️ One physical wire per call = doesn't scale


  Modern System (frequency/time division multiplexing):
  ┌────────┐                         ┌────────┐
  │ Call 1 │─┐  ┌─ Freq 1 ─┐        │        │
  │ Call 2 │─┼──┤  Freq 2  ├─ 1 ───▶│ Switch │
  │ Call 3 │─┘  └─ Freq 3 ─┘ fiber  │        │
  └────────┘                         └────────┘
  ✅ Thousands of calls on ONE fiber
```

---

## Quick Comparison

```
  ┌──────────────┬─────────────────────┬──────────────────────┐
  │              │      HTTP/2         │       Redis          │
  ├──────────────┼─────────────────────┼──────────────────────┤
  │ What's       │ Multiple HTTP       │ Multiple commands    │
  │ multiplexed  │ request/responses   │ from many clients    │
  ├──────────────┼─────────────────────┼──────────────────────┤
  │ Over what    │ Single TCP          │ Single thread +      │
  │ shared       │ connection          │ event loop           │
  │ resource     │                     │                      │
  ├──────────────┼─────────────────────┼──────────────────────┤
  │ Mechanism    │ Stream IDs in       │ epoll/kqueue I/O     │
  │              │ binary frames       │ + pipelining         │
  ├──────────────┼─────────────────────┼──────────────────────┤
  │ Benefit      │ No HOL blocking,    │ Handle 100K+ clients │
  │              │ fewer connections   │ on one thread        │
  └──────────────┴─────────────────────┴──────────────────────┘
```
