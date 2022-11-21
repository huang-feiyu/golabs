# Primary/Backup replication

[TOC]

### Failures

Failures:

* **Fail-stop** failures: there is a failure of physical machine => stops computer
* logic bugs, configuration error, malicious
* maybe: earthquake

### 2 Challenges

* Has primary really failed? (cannot figure out network partition or machine fail)
  * split-brain system
* Keep primary/backup in sync
  * apply changes in order
  * non-deterministic
* Fail over (故障转移)
  * backup take over

### Two approaches

* Replicated state of the primary

```
     +-----+                   +-----+
C -> |  P  | - state changes -> |  B  |
     +-----+                   +-----+
```

Send state changes (large bandwidth).

* Replicated state-machine

```
     +-----+                 +-----+
C -> |  P  | - operations -> |  B  |
     +-----+                 +-----+
```

Send deterministic operations.

---

Level of operations to replicate

* Application-level operations: file append, write
* Machine-level operations: interrupts, instructions (completely transparent => use virtual machines based on supervisor)

### Case study: VM FT

> [VM-FT](https://www.cs.princeton.edu/courses/archive/fall16/cos418/papers/scales-vm.pdf): exploit-virtualization

* transparent replication appears to Client that Server is a single machine
* VMware product (single-core solution)

##### Overview

```
+-------+                      +-------+
| APPs  |                      | APPs  |
+-------+                      +-------+
| Linux |                      | Linux |
+-------+                      +-------+
+----------+   Logging channel +----------+
|  VM-FT   | <---------------> |  VM-FT   |
+----------+                   +----------+
| Hardware |                   | Hardware |
+----------+                   +----------+
    |                                |
-------------------storage server------------
```

Only primary receives from client, and send log entries to backup. Only after receives an ACK, primary will send results to client.

Storage server is shared-disk. And there is an atomic flag as a lock to become a primary **when failure comes**.

##### Design

> Logging channel: only non-deterministic operations

Goal: behave like a single machine

=> backup does exactly the same instructions. **But** there are divergence sources: **non-deterministic** (input, packet order, timer interrupts)

For multi-core => disallow, it is complex to handle

---

Interrupts: primary sends the log entry (instruction number, interrupt, data) to backup, supervisor will take the log entry and do following the log entry.

---

Non-deterministic instruction: primary sends log entry(instruction number, output) to backup.

##### Failure Handle

* **Output Requirement**: if the backup VM ever takes over after a failure of the primary, the backup VM will continue executing in a way that is entirely consistent with all outputs that the primary VM has sent to the external world
* **Output Rule**: the primary VM may not send an output to the external world, until the backup VM has received and ACK the log entry associated with the operation producing the output.

##### Performance

Decrease by 30% when there is a network.

