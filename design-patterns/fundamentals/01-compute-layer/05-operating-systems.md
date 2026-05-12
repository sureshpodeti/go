# Operating Systems

## Overview

The Operating System (OS) is the software layer that manages hardware resources and provides services to applications. Understanding OS concepts is essential for designing efficient, scalable systems.

## OS Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Operating System Layers                   │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         User Space (Ring 3)                            │ │
│  │                                                        │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │ │
│  │  │   App 1  │  │   App 2  │  │   App 3  │           │ │
│  │  └──────────┘  └──────────┘  └──────────┘           │ │
│  │                                                        │ │
│  │  ┌────────────────────────────────────────────────┐   │ │
│  │  │         System Libraries (libc, etc.)         │   │ │
│  │  └────────────────────────────────────────────────┘   │ │
│  └────────────────────────────────────────────────────────┘ │
│                          │                                   │
│                          │ System Calls                      │
│                          ▼                                   │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Kernel Space (Ring 0)                          │ │
│  │                                                        │ │
│  │  ┌────────────────────────────────────────────────┐   │ │
│  │  │         System Call Interface                  │   │ │
│  │  └────────────────────────────────────────────────┘   │ │
│  │                                                        │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │ │
│  │  │ Process  │  │  Memory  │  │   File   │           │ │
│  │  │  Mgmt    │  │   Mgmt   │  │  System  │           │ │
│  │  └──────────┘  └──────────┘  └──────────┘           │ │
│  │                                                        │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │ │
│  │  │   I/O    │  │ Network  │  │ Security │           │ │
│  │  │  Mgmt    │  │  Stack   │  │          │           │ │
│  │  └──────────┘  └──────────┘  └──────────┘           │ │
│  │                                                        │ │
│  │  ┌────────────────────────────────────────────────┐   │ │
│  │  │         Device Drivers                         │   │ │
│  │  └────────────────────────────────────────────────┘   │ │
│  └────────────────────────────────────────────────────────┘ │
│                          │                                   │
│                          ▼                                   │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                    Hardware                            │ │
│  │         CPU, Memory, Storage, Network                  │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Process Management

### Process vs Thread

```
Process:
┌─────────────────────────────────────────┐
│         Process A                        │
│                                          │
│  ┌────────────────────────────────┐     │
│  │      Code Segment              │     │
│  └────────────────────────────────┘     │
│  ┌────────────────────────────────┐     │
│  │      Data Segment              │     │
│  └────────────────────────────────┘     │
│  ┌────────────────────────────────┐     │
│  │      Heap                      │     │
│  └────────────────────────────────┘     │
│  ┌────────────────────────────────┐     │
│  │      Stack                     │     │
│  └────────────────────────────────┘     │
│                                          │
│  • Own address space                    │
│  • Isolated from other processes        │
│  • Heavy context switch                 │
└─────────────────────────────────────────┘

Thread:
┌─────────────────────────────────────────┐
│         Process B                        │
│                                          │
│  ┌────────────────────────────────┐     │
│  │      Code (Shared)             │     │
│  └────────────────────────────────┘     │
│  ┌────────────────────────────────┐     │
│  │      Data (Shared)             │     │
│  └────────────────────────────────┘     │
│  ┌────────────────────────────────┐     │
│  │      Heap (Shared)             │     │
│  └────────────────────────────────┘     │
│                                          │
│  ┌──────────┐  ┌──────────┐            │
│  │ Thread 1 │  │ Thread 2 │            │
│  │  Stack   │  │  Stack   │            │
│  └──────────┘  └──────────┘            │
│                                          │
│  • Shared address space                 │
│  • Lightweight context switch           │
│  • Need synchronization                 │
└─────────────────────────────────────────┘
```

### Process States

```
Process State Diagram:

┌─────────┐
│   New   │  ← Process created
└────┬────┘
     │
     ▼
┌─────────┐
│  Ready  │ ◄──────────────┐
└────┬────┘                │
     │                     │
     │ Scheduled           │ Preempted/
     │                     │ Time slice expired
     ▼                     │
┌─────────┐                │
│ Running │ ───────────────┘
└────┬────┘
     │
     │ I/O or Event Wait
     ▼
┌─────────┐
│ Waiting │
└────┬────┘
     │
     │ I/O Complete
     │
     ▼
┌─────────┐
│  Ready  │
└────┬────┘
     │
     │ Exit
     ▼
┌─────────┐
│Terminated│
└─────────┘
```

### Process Control Block (PCB)

```
Process Control Block:

┌─────────────────────────────────────────┐
│         PCB for Process #1234           │
├─────────────────────────────────────────┤
│ Process ID: 1234                        │
│ Parent PID: 1000                        │
│ State: Running                          │
├─────────────────────────────────────────┤
│ CPU Registers:                          │
│ • Program Counter: 0x4000A120           │
│ • Stack Pointer: 0x7FFF1234             │
│ • General Registers: ...                │
├─────────────────────────────────────────┤
│ Memory Management:                      │
│ • Page Table Pointer                    │
│ • Memory Limits                         │
├─────────────────────────────────────────┤
│ Scheduling Info:                        │
│ • Priority: 20                          │
│ • CPU Time Used: 1.5s                   │
│ • Time Slice Remaining: 10ms            │
├─────────────────────────────────────────┤
│ I/O Status:                             │
│ • Open Files: [fd0, fd1, fd2]           │
│ • Network Sockets: [sock1]              │
├─────────────────────────────────────────┤
│ Accounting:                             │
│ • CPU Time: 1.5s                        │
│ • Memory Usage: 128 MB                  │
│ • I/O Operations: 1,234                 │
└─────────────────────────────────────────┘
```

## CPU Scheduling

### Scheduling Algorithms

```
1. First-Come, First-Served (FCFS):
┌────┬────┬────┬────┐
│ P1 │ P2 │ P3 │ P4 │
└────┴────┴────┴────┘
Simple but can cause convoy effect

2. Shortest Job First (SJF):
┌──┬────┬──────┬──────────┐
│P2│ P1 │  P3  │    P4    │
└──┴────┴──────┴──────────┘
Optimal average waiting time
Problem: Need to know execution time

3. Round Robin (RR):
Time Quantum = 10ms
┌────┬────┬────┬────┬────┬────┐
│ P1 │ P2 │ P3 │ P1 │ P2 │ P3 │
└────┴────┴────┴────┴────┴────┘
Fair, good for interactive systems

4. Priority Scheduling:
┌──────────┬────┬──┬────┐
│P1 (Pri=1)│P2=2│P3│P4=4│
└──────────┴────┴──┴────┘
Problem: Starvation of low-priority

5. Multi-Level Feedback Queue:
┌─────────────────────────────────┐
│ Queue 0 (Highest Priority)      │
│ Time Quantum: 8ms               │
├─────────────────────────────────┤
│ Queue 1 (Medium Priority)       │
│ Time Quantum: 16ms              │
├─────────────────────────────────┤
│ Queue 2 (Lowest Priority)       │
│ Time Quantum: FCFS              │
└─────────────────────────────────┘

New processes → Queue 0
If not finished → Move to Queue 1
If not finished → Move to Queue 2

Favors short, interactive processes
```

### Context Switching

```
Context Switch Process:

Time ──────────────────────────────────►

Process A Running:
┌──────────────────┐
│   Process A      │
│   Executing      │
└──────────────────┘
         │
         │ Interrupt (timer, I/O, etc.)
         ▼
┌──────────────────┐
│ Save Process A   │
│ • Registers      │
│ • Program Counter│
│ • Stack Pointer  │
│ • State → Ready  │
└──────────────────┘
         │
         ▼
┌──────────────────┐
│ Scheduler        │
│ Select Process B │
└──────────────────┘
         │
         ▼
┌──────────────────┐
│ Restore Process B│
│ • Registers      │
│ • Program Counter│
│ • Stack Pointer  │
│ • State → Running│
└──────────────────┘
         │
         ▼
┌──────────────────┐
│   Process B      │
│   Executing      │
└──────────────────┘

Context Switch Cost:
• Direct: 1-10 μs (save/restore)
• Indirect: Cache/TLB flush (100+ μs)
• Total: Can be significant!
```

## Inter-Process Communication (IPC)

```
IPC Mechanisms:

1. Pipes:
┌──────────┐    ┌──────┐    ┌──────────┐
│Process A │───►│ Pipe │───►│Process B │
└──────────┘    └──────┘    └──────────┘
• Unidirectional
• Parent-child processes
• Byte stream

2. Named Pipes (FIFO):
┌──────────┐    ┌──────────┐    ┌──────────┐
│Process A │───►│Named Pipe│◄───│Process B │
└──────────┘    │/tmp/fifo │    └──────────┘
                └──────────┘
• Bidirectional
• Unrelated processes
• File system entry

3. Message Queues:
┌──────────┐                    ┌──────────┐
│Process A │                    │Process B │
└────┬─────┘                    └────▲─────┘
     │                               │
     ▼                               │
┌────────────────────────────────────┴─────┐
│         Message Queue                     │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐            │
│  │Msg1│ │Msg2│ │Msg3│ │Msg4│            │
│  └────┘ └────┘ └────┘ └────┘            │
└──────────────────────────────────────────┘
• Structured messages
• Asynchronous
• Persistent

4. Shared Memory:
┌──────────┐    ┌──────────────┐    ┌──────────┐
│Process A │───►│Shared Memory │◄───│Process B │
└──────────┘    │   Region     │    └──────────┘
                └──────────────┘
• Fastest IPC
• Requires synchronization
• Direct memory access

5. Sockets:
┌──────────┐                    ┌──────────┐
│Process A │◄──────Socket──────►│Process B │
└──────────┘                    └──────────┘
• Network communication
• Local or remote
• TCP/UDP protocols

6. Signals:
┌──────────┐                    ┌──────────┐
│Process A │─────SIGTERM───────►│Process B │
└──────────┘                    └──────────┘
• Asynchronous notifications
• Limited data (signal number)
• Interrupt handling
```

## Synchronization

### Race Conditions

```
Race Condition Example:

Shared Variable: counter = 0

Thread 1:              Thread 2:
┌──────────────┐      ┌──────────────┐
│ Read counter │      │ Read counter │
│ (value = 0)  │      │ (value = 0)  │
│              │      │              │
│ Increment    │      │ Increment    │
│ (0 + 1 = 1)  │      │ (0 + 1 = 1)  │
│              │      │              │
│ Write counter│      │ Write counter│
│ (counter = 1)│      │ (counter = 1)│
└──────────────┘      └──────────────┘

Expected: counter = 2
Actual: counter = 1 (Lost update!)
```

### Synchronization Primitives

```
1. Mutex (Mutual Exclusion):
┌─────────────────────────────────────┐
│         Critical Section            │
│                                     │
│  Thread 1:                          │
│  mutex.lock()                       │
│  // Critical section                │
│  counter++                          │
│  mutex.unlock()                     │
│                                     │
│  Thread 2:                          │
│  mutex.lock() ← Blocks until unlock │
│  // Critical section                │
│  counter++                          │
│  mutex.unlock()                     │
└─────────────────────────────────────┘

2. Semaphore:
┌─────────────────────────────────────┐
│    Semaphore (count = 3)            │
│                                     │
│  Thread 1: sem.wait() → count = 2   │
│  Thread 2: sem.wait() → count = 1   │
│  Thread 3: sem.wait() → count = 0   │
│  Thread 4: sem.wait() → BLOCKS      │
│                                     │
│  Thread 1: sem.signal() → count = 1 │
│  Thread 4: UNBLOCKS                 │
└─────────────────────────────────────┘

3. Condition Variable:
┌─────────────────────────────────────┐
│    Producer-Consumer                │
│                                     │
│  Producer:                          │
│  mutex.lock()                       │
│  while (queue.full())               │
│    cond_not_full.wait(mutex)        │
│  queue.add(item)                    │
│  cond_not_empty.signal()            │
│  mutex.unlock()                     │
│                                     │
│  Consumer:                          │
│  mutex.lock()                       │
│  while (queue.empty())              │
│    cond_not_empty.wait(mutex)       │
│  item = queue.remove()              │
│  cond_not_full.signal()             │
│  mutex.unlock()                     │
└─────────────────────────────────────┘

4. Read-Write Lock:
┌─────────────────────────────────────┐
│    Read-Write Lock                  │
│                                     │
│  Multiple readers: OK               │
│  ┌────────┐ ┌────────┐ ┌────────┐  │
│  │Reader 1│ │Reader 2│ │Reader 3│  │
│  └────────┘ └────────┘ └────────┘  │
│         ▼        ▼        ▼         │
│      ┌──────────────────────┐       │
│      │    Shared Data       │       │
│      └──────────────────────┘       │
│                                     │
│  Writer: Exclusive                  │
│  ┌────────┐                         │
│  │Writer  │ ← Blocks all others     │
│  └────────┘                         │
│      ▼                              │
│  ┌──────────────────────┐           │
│  │    Shared Data       │           │
│  └──────────────────────┘           │
└─────────────────────────────────────┘
```

### Deadlock

```
Deadlock Example:

Resource A        Resource B
    ▲                 ▲
    │                 │
    │ Holds           │ Holds
    │                 │
┌───┴────┐        ┌───┴────┐
│Thread 1│        │Thread 2│
└───┬────┘        └───┬────┘
    │                 │
    │ Waits for       │ Waits for
    │                 │
    ▼                 ▼
Resource B        Resource A

Deadlock Conditions (all must be true):
1. Mutual Exclusion: Resources can't be shared
2. Hold and Wait: Hold resources while waiting
3. No Preemption: Can't forcibly take resources
4. Circular Wait: Circular chain of waiting

Prevention:
• Break one of the four conditions
• Resource ordering (always acquire in same order)
• Timeouts
• Deadlock detection and recovery
```

## Memory Management (OS Perspective)

### Virtual Memory Management

```
Virtual Memory System:

┌─────────────────────────────────────────┐
│         Virtual Address Space           │
│                                          │
│  Process 1:                             │
│  ┌────────────────────────────────┐     │
│  │ 0x0000 - 0xFFFF (4 GB)         │     │
│  └────────────────────────────────┘     │
│                │                         │
│                ▼ MMU Translation         │
│  ┌────────────────────────────────┐     │
│  │ Physical RAM (16 GB)           │     │
│  │ ┌────┬────┬────┬────┬────┐     │     │
│  │ │ P1 │ P2 │ P1 │ P3 │ P2 │     │     │
│  │ └────┴────┴────┴────┴────┘     │     │
│  └────────────────────────────────┘     │
│                │                         │
│                ▼ Page Fault              │
│  ┌────────────────────────────────┐     │
│  │ Swap Space (Disk)              │     │
│  │ ┌────┬────┬────┬────┐          │     │
│  │ │ P1 │ P2 │ P3 │ P4 │          │     │
│  │ └────┴────┴────┴────┘          │     │
│  └────────────────────────────────┘     │
└─────────────────────────────────────────┘
```

### Page Replacement Algorithms

```
Page Replacement Algorithms:

1. FIFO (First-In, First-Out):
Memory: [1][2][3]
Access: 4
Replace: 1 (oldest)
Memory: [4][2][3]

2. LRU (Least Recently Used):
Memory: [1][2][3]
Access: 2, 4
Replace: 1 (least recently used)
Memory: [2][4][3]

3. LFU (Least Frequently Used):
Memory: [1:5][2:3][3:8]
       (page:count)
Access: 4
Replace: 2 (lowest count)
Memory: [1:5][4:1][3:8]

4. Clock (Second Chance):
┌───┬───┬───┬───┐
│1:1│2:0│3:1│4:0│ ← Reference bit
└───┴───┴───┴───┘
     ▲
     └─ Clock hand

Replace: Find first page with bit=0
Set bit=0 as you scan
```

## File Systems (OS Perspective)

### File Operations

```
File System Operations:

┌─────────────────────────────────────────┐
│         File Descriptor Table           │
│                                          │
│  Process:                               │
│  ┌────┬──────────────────┐              │
│  │ FD │ File Pointer     │              │
│  ├────┼──────────────────┤              │
│  │ 0  │ stdin            │              │
│  │ 1  │ stdout           │              │
│  │ 2  │ stderr           │              │
│  │ 3  │ /var/log/app.log │              │
│  │ 4  │ /data/file.txt   │              │
│  └────┴──────────────────┘              │
│         │                                │
│         ▼                                │
│  ┌────────────────────────────────┐     │
│  │    Open File Table             │     │
│  │  • File offset                 │     │
│  │  • Access mode (r/w/rw)        │     │
│  │  • Reference count             │     │
│  └────────────────────────────────┘     │
│         │                                │
│         ▼                                │
│  ┌────────────────────────────────┐     │
│  │    Inode Table                 │     │
│  │  • File metadata               │     │
│  │  • Block pointers              │     │
│  └────────────────────────────────┘     │
└─────────────────────────────────────────┘
```

### I/O Buffering

```
I/O Buffering Strategies:

1. No Buffering:
App ──► Kernel ──► Disk (every write)
Slow but immediate

2. User-Level Buffering:
App ──► Buffer ──► Kernel ──► Disk
        (libc)
Fast for small writes

3. Kernel-Level Buffering:
App ──► Kernel Buffer ──► Disk
        (page cache)
Shared across processes

4. Double Buffering:
┌──────────┐
│ Buffer A │ ← App writes
└──────────┘
┌──────────┐
│ Buffer B │ ← Kernel flushes to disk
└──────────┘
Parallel I/O and computation
```

## System Calls

### Common System Calls

```
System Call Categories:

1. Process Control:
   • fork()    - Create child process
   • exec()    - Execute program
   • exit()    - Terminate process
   • wait()    - Wait for child
   • kill()    - Send signal

2. File Management:
   • open()    - Open file
   • read()    - Read from file
   • write()   - Write to file
   • close()   - Close file
   • lseek()   - Move file pointer

3. Device Management:
   • ioctl()   - Device control
   • read()    - Read from device
   • write()   - Write to device

4. Information:
   • getpid()  - Get process ID
   • alarm()   - Set alarm
   • sleep()   - Suspend process

5. Communication:
   • pipe()    - Create pipe
   • socket()  - Create socket
   • send()    - Send message
   • recv()    - Receive message

6. Memory:
   • mmap()    - Map memory
   • munmap()  - Unmap memory
   • brk()     - Change heap size
```

### System Call Overhead

```
System Call Execution:

User Mode:
┌──────────────────────────────────┐
│ Application Code                 │
│ result = read(fd, buf, size)     │
└────────────┬─────────────────────┘
             │
             │ Trap/Interrupt
             ▼
Kernel Mode:
┌──────────────────────────────────┐
│ 1. Save user context             │
│ 2. Validate parameters           │
│ 3. Execute system call           │
│ 4. Restore user context          │
│ 5. Return to user mode           │
└────────────┬─────────────────────┘
             │
             ▼
User Mode:
┌──────────────────────────────────┐
│ Application continues            │
└──────────────────────────────────┘

Cost: 100-1000 ns per system call
Minimize system calls for performance!
```

## OS Performance Tuning

### Key Parameters

```
Linux Kernel Parameters:

1. Process Limits:
   /proc/sys/kernel/pid_max
   • Maximum number of processes

2. File Descriptors:
   /proc/sys/fs/file-max
   • System-wide file descriptor limit
   
   ulimit -n
   • Per-process limit

3. Network:
   /proc/sys/net/core/somaxconn
   • Socket listen backlog
   
   /proc/sys/net/ipv4/tcp_max_syn_backlog
   • SYN queue size

4. Memory:
   /proc/sys/vm/swappiness
   • Swap aggressiveness (0-100)
   
   /proc/sys/vm/dirty_ratio
   • Dirty page threshold

5. Scheduler:
   /sys/kernel/debug/sched_features
   • Scheduler features
```

## Summary

Key takeaways for architects:

1. **Understand process vs thread tradeoffs**
   - Processes: Isolation, safety
   - Threads: Performance, shared state

2. **Choose appropriate IPC mechanism**
   - Shared memory: Fastest
   - Sockets: Most flexible
   - Message queues: Structured

3. **Avoid synchronization issues**
   - Race conditions
   - Deadlocks
   - Starvation

4. **Minimize system call overhead**
   - Batch operations
   - Use buffering
   - Async I/O

5. **Monitor OS metrics**
   - CPU scheduling
   - Memory usage
   - I/O wait time
   - Context switches

## Next Steps

Continue to [Virtualization & Containers](./06-virtualization-containers.md) to understand modern compute abstraction technologies.
