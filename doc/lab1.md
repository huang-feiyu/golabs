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
    * Ask for task to execute

General schedule of an MR:
1. Worker asks coordinator for tasks: call `GetTask` to get a task from
   coordinator
2. Coordinator assigns task according to current phase, only until all Map tasks
   are done, coordinator will issue Reduce tasks
3. Worker gets reply(task) from coordinator, it will do stuff according to the
   task type:<br/>
   (1) `Map`: Perform the map function on each K/V pair, store them in an
       intermediate file, store them in NReduce intermediate files, finally
       atomically rename them to `mr-Y-X`<br/>
