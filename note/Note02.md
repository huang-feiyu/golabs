# Threads & RPC

> [FAQ](https://pdos.csail.mit.edu/6.824/papers/tour-faq.txt)
>
> [Question: Web Crawler](https://go.dev/tour/concurrency/10)
>
> [RPC Package](https://pkg.go.dev/net/rpc)

[TOC]

### Why Go?

* good support for threads & rpc
* garbage collector
*  type safe
* simple
* compiled

### Thread

> Thread of execution

Go run: Create a OS process, go runtime system (with several go routines by `go`)

[light-weight] Thread in go => Go routine

* State: PC, stack, registers, **shared memory**
* Primitives
  * start/`go`: create thread
  * exit: generally exit implicitly
  * stop: get blocked, put it aside (take the state away)
  * resume: resume the stopped thread

##### Why threads?

Express concurrency

* I/O concurrency
* Multi-core parallelism
* Convenience

##### Thread challenges

challenges

* Race conditions
  * Avoid sharing (use channels)
  * Use locks (Go has a Race detector `-race`)
* Coordination (i.e., one must wait another)
  * Channels
  * Conditional variables
* Deadlock

##### Go: and challenges

* Approach #1: Channels (no sharing)
* Approach #2: locks + condition variables (shared memory)

### RPC

> Remote Procedure Call

Goal: RPC ≈ PC

Client calls the function defined in server, RPC makes sure passing parameters and return values (using network and stub procedure).

---

RPC semantics under failures

* at-least-once
* at-most-once (filtering duplicates)
* exactly-once (hard to do)

