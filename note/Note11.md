# Distributed Transactions

[TOC]

### Transactions

Problem: cross-machine atomic operations

Primitives: Begin, Commit, Abort

ACID:

* Atomicity: all or nothing
* Consistency: looks correct
* Isolation: as if alone
* Duration: I will survive

Isolation level:

* Serializable: same outcome with serializable outcome

### 2-phase locking (2PL)

Pessimistic concurrency control

* txn acquires lock before using
* txn holds until commit

### 2-phase commit (2PC)

Coordinator (raft clust) + Participants

* phase 1: commit request
* phase 2: success (only when all agree) or failure

```
    协调者                                              参与者
                              QUERY TO COMMIT
                -------------------------------->
                              VOTE YES/NO           prepare*/abort*
                <-------------------------------
commit*/abort*                COMMIT/ROLLBACK
                -------------------------------->
                              ACKNOWLEDGMENT        commit*/abort*
                <--------------------------------
end
```

