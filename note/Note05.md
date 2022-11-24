# Raft (1): Leader Election & Log Replication

> [Raft extended](https://pdos.csail.mit.edu/6.824/papers/raft-extended.pdf)
>
> [Notion: Raft](https://www.notion.so/huangfeiyu/Raft-13a975067399457d841546d699872a6f)

[TOC]

### Majority Rule

Pattern: Single point of failure => 系统中一旦失效，就会让整个系统无法运行的组件

* MapReduce: Coordinator
* GFS: Master
* VM-FT: Storage server

=> Avoid Split-Brain

When encounter network partition, we have two servers as "Master".

To solve network partition: **majority rule**.

What is majority: Majority of all servers including failed servers

### How to use raft

Use raft to build Replicated State-Machine:

Consensus Module in our case is Raft.

![RSM with Raft](https://user-images.githubusercontent.com/70138429/203718066-58de140c-9c9a-40c4-8604-6473bb2992c6.png)

### Details

##### Why logs?

> logs are identical on all servers

* Retransmission
* Persistence
* Space tentative (没有被提交)

##### Log entry

Log entry = Instruction + Leader's term + Log index

* Elect leader
* Ensure logs identical

##### Election

Reason to elect: someone misses heartbeat

Challenge: split vote => No one can get the majority of votes

=> **Randomized election timeout**: followers set their election timer(150~300ms), everyone reset timeout, only when the timer goes off, it will start election. The interval between 2 timers may be enough to complete election.

* few heartbeats <= Election timeouts (减少选举可能性)
* Random value
* Short enough that downtime is short

##### Log diverge

![image](https://user-images.githubusercontent.com/70138429/203726244-175009f2-c119-4f05-b59b-591d0daae74d.png)

If leader is elected in a~f besides top:

* term 2 & 3 will be rejected
* term 4 will be accepted
* term 7 will be accepted if (d) becomes leader<br/>term 7 will be rejected if (c) becomes leader

Who can become a leader: a, c, d

