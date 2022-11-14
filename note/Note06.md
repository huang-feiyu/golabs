# Lab1 Q&A

> [Debugging pretty](https://blog.josejg.com/debugging-pretty/)
>
> [Debugging](https://pdos.csail.mit.edu/6.824/notes/debugging.pdf)
>
> [Lab1: MapReduce](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html)

[TOC]

### Solution demo

* Step 1 - `rpc.go`: APIs that coordinator communicates with workers

  * `GetTask` RPCs
  * `FinishedTask` RPCs
* Step 2 - `coordinator.go`: Handlers for RPCs
  * Fill up the status data structures
  * Handler for `GetTask`
  * Handler for `FinishedTask`
* Step 3 - `coordinator.go`: Send the RPCs
  * An infinite loop to perform `map`/`reduce` until `Done`
* Step 4 - `worker.go`: Handlers for managing the intermediate files
* Step 5 - `worker.go`: `PerfromMap`
  * Reads the file
  * Map them to keys
  * Create temp files
  * Write to intermediate files
* Step 6 - `worker.go`: `PerformReduce`
  * Get the intermediate files, collect the K/V pairs
  * Sort by the key
  * Create temp file to write
  * Rename to final reduce file
* Step 7 - `coordinator.go`: Handle `GetTask`
  * Do map tasks until no map left
  * Do reduce tasks, all done
* Step 8 - `coordinator.go`: Wait for a task issue
  * Can use a condition variable

### General Tips

* `Printf`s for debugging
  * Conditional `Printf`s
  * Formatting trick: color scheme for RPCs, columns, server IDs
  * Redirect output to file (command `&>` file_output.txt)
* Type Ctrl-\ to kill a program and dump all its goroutine stacks
* Defers pushed onto stack, executed in FIFO order when fxc returns

```go
mu.Lock()
defer mu.Unlock()
defer FxnToRunBeforeUnlock()
```

### Problems of MapReduce

MapReduce: Anything that takes local computed statistics and computes a global statistic over them

* When to shuffle (combiner)
  * Combining occurs at `map`, result written to intermediate file
  * Sorting occurs at reduce after *all* outputs from map are read by reduce
* Difference between using a pointer and a value for the argument object to an RPC call
  * Passing a reference can be cheaper
* Clean way to exit the workers and coordinator
  * Send an "Exit" RPC from coordinator to worker as final response to `GetTask` RPC
* Unexpected EOF errors come from where
  * workers calling socket when coordinator has closed

