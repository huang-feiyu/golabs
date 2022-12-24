# Chain Replication

> [Notion](https://www.notion.so/huangfeiyu/CR-e88d88a7d6614e129552e12f55c30d1d)
>
> [Paper](https://pdos.csail.mit.edu/6.824/papers/cr-osdi04.pdf): Chain Replication for Supporting High Throughput and Availability

[TOC]

### Approaches to RSMs

Replication State Machines

1. Run all ops through Raft/Paxos => Risk: size of checkpoint grows with storage size
2. Configuration service + P/B replication => common

### Chain Replication

> Influential design

Primary/Backup replication for approach 2

* Read ops involve only **one** server
* Simple recovery plan
* Linearizability

---

```
+-----------------------+
| Configuration Service |
+-----------------------+
     +----+
     | S1 | <- Head
     +----+
        |   (When applied it, send to next)
        V
     +----+
     | S2 |
     +----+
        |
        V
     +----+
     | S3 |  <- Tail
     +----+
```

* Write: Client sends write-req to **head**, chain updates to the **tail**. Tail commits and replies to client.
* Read: Clients sends read-req to **tail**, and **tail** replies.

### Crash

Configuration Service (Raft/Paxos/ZooKeeper) must be alive to do stuff.

* Header crashes: Update S2 to be head, no need to repair anything
* I-Node crashes: S1 points to S3, send request according to diff states
* Tail crashes: Update S2 to be tail, client now reads from S2

### Properties

* Advantages
  * Update and read RPCs split (读写分离)
  * Head sends update once
  * Read ops involve only Tail
  * Simple crash recovery
* Disadvantages
  * One failure requires reconfiguration

### Extension

* Splits object across many chains => Multiple tails

