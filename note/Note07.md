# Raft (2): log divergence, snapshots

> [Raft extended](https://pdos.csail.mit.edu/6.824/papers/raft-extended.pdf)
>
> [Notion: Raft](https://www.notion.so/huangfeiyu/Raft-13a975067399457d841546d699872a6f)

[TOC]

### Log divergence

Leader election rule:

* majority
* at-least up-to-date log

### Log catch up

For every node, leader maintains a `nextIndex` for it. When one follower gets behind, leader decrements the `nextIndex`.

For every node, leader also maintains a `matchIndex` for it. It is the lower bound.

---

When leader forces followers' log entries reaching a consensus point, it might erase some of them erase their log entries.

When leader commits in its own term, it also commits the previous terms' log entries.

---

Optimization: catch up quickly

Instead of just rejection, the rejection RPC can contain more information like term & index. So, the leader can jump back faster.

### Persistence

What happens after reboots?

* Strategy 1: re-join the raft cluster => replay log
* Strategy 2: from persistent state, voted for, log[], current term

### Snapshots

service recovery:

* Strategy 1: Replay log => recreate state
* Strategy 2: Snapshots (state stores all ops through 0 ~ specific committed log)

Snapshot can not only used for service recovery, but also used for compact log.

How state machine applies the snapshot? => Apply channel

---

Client communicates with server through clerk. Clerk is somehow a stub, which stores: leader and followers client thinks, and assign unique ID for each operation.

### Linearizability

> Act as a single machine => Strong consistency

Linearizability:

1. total order of operations
2. match real-time
3. read return results of the last write

