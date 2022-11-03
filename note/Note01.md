# Introduction

> [MapReduce](https://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf)

[TOC]

### Distributed System

* Keywords
  * multiple
  * networked
  * cooperating
  * computers
* Whys
  * Connect physically [sharing] separated machines
  * Increase capacity through parallelism
  * Tolerate faults
  * Achieve security [via isolation]

### Historical context

1. Local area networks (1980s)<br/>DNS + email
2. Datacenters (1990s)<br/>Big web sites, Web search, shopping
3. Cloud computing (2000s)<br/>
4. Current state: active but difficult

### Challenges

* many concurrent parts => complexity
* must deal with *partial failure* => complexity
* tricky release the performance benefit

### Course structure

Why take 6.824?

* Interesting: hard problems but powerful solutions
* Used in real world
* Active area of research
* Hands on experience

---

Lectures: big ideas

Papers: case study

Labs: public tricky test cases<br/>(1) MapReduce<br/>(2) Replication using Raft<br/>(3) Replicated K/V service<br/>(4) Sharded K/V services

### Main topics

Infrastructure

* Storage: K/V, FS
* Computation: MapReduce
* Communication: RPC, more in 6.829 (Network systems)

Abstractions: we want our sys-dis work like a single one.

---

Main topics

* Fault tolerance
  * Availability: replication
  * Recoverability: logging/transactions, durable storage
* Consistency
  * does `get()` return the value of the **last** `put()`?
* Performance
  * Throughput
  * Tail latency
* Implementation

### MapReduce

* Context
  * Data centers in Google
  * Multi-hours
  * Terabytes data
  * Web Indexing
* Goal: easy for non-experts to write distributed applications
* Approach
  * map + reduce => **sequential code**
  * MR deals with distribution

---

**Abstract view**

Notes in [CSE138: L16 & L17](https://github.com/huang-feiyu/CSE138-Notes).


3 phases in MapReduce

- map phase
  - `map` function: input k-v pair -> **set** of intermediate k-v pairs
  - `<Doc1, [the, quick, fox]>` -> `{<the, Doc1>, <quick, Doc1>, <fox, Doc1>}`
- shuffle phase (**expensive**)
  - data (intermediate k-v) from map workers is sent to reduce workers, according to some data partitioning function, e.g. `hash(key) % R` (would be bad if R changes)
- reduce phase
  - `reduce` function: (key, set of values) -> set of output values
  - `<dog, {Doc1, Doc3}>` -> `<dog, [Doc1, Doc3]`
    `<dog, {1, 3, 2}>` -> `<dog, 6>`

---

MapReduce [paper](https://pdos.csail.mit.edu/6.824/papers/mapreduce.pdf)

---

Fault tolerance:

* Coordinator returns map & reduce<br/>When it doesn't receive the worker's feedback, then run task twice
  * in function deterministic, it is okay for both `map` & `reduce`. When re-run reduce, clean up the intermediate files
* Other failures
  * Can coordinator fail? No, need a whole re-run.
  * Slow workers (stragglers)? do backup task.

