# MapReduce

> 6.824 [Lab 1](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html): MapReduce
> -- [Note on MapReduce](https://www.notion.so/huangfeiyu/MapReduce-1087e0a149d54e129a9adcf9c013e2ae)

## Prepare

MapReduce is a programming model that enables non-expert programmers works well
on distributed systems. It hides a lot of details: parallelism, data distribution,
error handling, load balancing. So that developers can only write `map` & `reduce`
functions, then run them on thousands of machines.

![MapReduce Overview](https://miro.medium.com/max/1400/1*g7loMfDE6uOq4wCxE5Mwug.png)

There only two kinds of processes: coordinator & worker
* one coordinator: distribute data/tasks, record status...
* map workers: coordinator gives Map tasks to workers until all Maps complete
    * write intermediate data to local disk (*mr-Y-X* in current directory)
    * split output into one file per Reduce task
* reduce workers: after all Maps have finished, coordinator hands out Reduces
    * fetch its intermediate output from [all] map workers
    * write a separate output file on GFS (in our case, write to *mr-out-X*)
* They communicate via RPC

* `map` (Select): $(k1, v1) \to list(k2, v2)$
* `reduce` (Group By): $(k2, list(v2)) \to list(v2)$
* (People says FP here means that `map`/`reduce` have no side effects, which
  acquire the implementation provides Atomicity)

> It is powerful because of streaming programming instead of functional programming.
> -- [MapReduce is not FP](https://jkff.medium.com/mapreduce-is-not-functional-programming-39109a4ba7b2)

---

[mrsequential.go](../src/main/mrsequential.go) code reading:
1. `loadPlugin`: load the application `map` & `reduce` functions from *.so* file<br/>
   (1) `map`: `func(filename, contents) -> []KeyValue`<br/>
   (2) `reduce`: `func(key, values) -> string`
2. Read each input filename, pass it to `map`, get the intermediate output
3. Sort the intermediate output by key
4. Call `reduce` on each distinct key, output the results to *mr-out-0*

---

Test your code:
```bash
# refresh word-count plugin
go build -race -buildmode=plugin ../mrapps/wc.go

# run the coordinator
rm mr-out*
go run -race mrcoordinator.go pg-*.txt

# run the worker in other window[s]
go run -race mrworker.go wc.so
```

## Implementation Notes

> Because I do not know how to start, so I take lecture 6 for inspiration.

* Coordinator
    * Store task state, I-file location for Reduce (no need to store worker state,
      every worker will ask for task initiative)
    * Assign tasks
* Worker
    * Ask for task to execute periodically
    * Perform tasks

General schedule of an MR:
1. Worker asks coordinator for task periodically: call `GetTask` to get a task
   from coordinator
2. Coordinator assigns task according to current phase, only until all Map tasks
   are done, coordinator will issue Reduce tasks. If everything is done, 
3. Worker gets reply(task) from coordinator, it will do stuff according to the
   task type:<br/>
   (1) `Map`: Perform the map function on each K/V pair, store them in NReduce
            intermediate files, finally atomically rename them to `mr-Y-X`.<br/>
   (2) `Reduce`: Read the intermediate files, collect the K/V pairs. Sort K/V
               pairs by key. Apply the reduce function for each distinct key.
               Create temp file to write, rename to the final name `mr-out-X`.<br/>
   (3) `Done`: just return.
4. Worker finishes up a task: call `FinishTask` to notify coordinator it's done
5. If all tasks are done, then entire system is done

---

Q: How does condition variable work in our lab?

Before answering the question, we need to know how an RPC works:
* Client (Worker)
    * Use `Dial()` creates a Socket connection to the server
    * Use `Call()` to ask the RPC library to perform the call
* Server (Coordinator)
    * Declare an object with methods as RPC handlers
    * Use `Register()` to register the object with the RPC library
    * Start listening Socket connections
    * Read each request, creates a new **goroutine** for this request, do stuff

In our case, for every `GetTask` RPC, creates a goroutine for it:
* If there is a task to be issued, just return and kill goroutine
* Otherwise, we need to wait if not all Map tasks have done. The goroutine now
  holds the lock, but it needs to wait until all Map tasks done. So, the
  goroutine releases the lock and sleeps (gives up CPU) itself until waken-up.
  * When a task finishes, the `FinishTask` handler will broadcast to wake all
    waiting goroutines up, tell them to check whether it is possible to issue a
    task then kill itself
  * But if all `FinishTask` RPCs arrived **after** `GetTask` RPC handler walks
    through it & **before** it gets waiting, the broadcast will be wasted. So,
    we lost some waken-up. => Need a goroutine periodically wake the goroutines
    of coordinator up.
