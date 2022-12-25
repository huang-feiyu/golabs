# Raft

> 6.824 [Lab 2](https://pdos.csail.mit.edu/6.824/labs/lab-raft.html):
> -- [Note on Raft](https://www.notion.so/huangfeiyu/Raft-13a975067399457d841546d699872a6f)

* [ ] Part A: Leader Election -- Implement Raft leader election and heartbeats
* [ ] Part B: Log Replication -- Implement appending new log entries
* [ ] Part C: Persistence -- Implement restoring persistent state
* [ ] Part D: Log Compaction -- Implement snapshot to reduce replay

## Prepare

* [Guide to Raft](https://thesquareplanet.com/blog/students-guide-to-raft/)
* [Raft Q&A](https://thesquareplanet.com/blog/raft-qa/)
* [Locking Advice](https://pdos.csail.mit.edu/6.824/labs/raft-locking.txt)
  * Rule 1: Every resource shared among goroutines should be locked to avoid
    accessing concurrently
  * Rule 2: Lock must be held when modifying as a whole
  * Rule 3: Lock must be held when reading
  * Rule 4: No lock while doing anything that might wait (use a new goroutine)
  * Rule 5: Be careful about assumptions across a drop and re-acquire of a lock
* [Structure Advice](https://pdos.csail.mit.edu/6.824/labs/raft-structure.txt)
  * Raft instance state: Use shared data and locks
  * Time-driven activities: Use a dedicated long-running goroutine
  * Management of the election timeout: Maintain a last time variable and use
    `time.Sleep()` to drive the periodic checks
  * `applyCh`: Use a separate long-running goroutine that sends committed log
    entries, use condition variable to notify apply channels
  * RPC should be sent in its own goroutine.
* [Debugging Advice](https://blog.josejg.com/debugging-pretty/)

![Diagram of Raft Interactions](https://user-images.githubusercontent.com/70138429/209463675-c60cec51-8fac-4166-97f9-583be4afd538.png)

---

In this lab, we will implement Raft, an understandable replicated state machine
protocol. A Raft instance has to deal with the arrival of external events
(Start() calls, AppendEntries and RequestVote RPCs, and RPC replies), and it has
to execute periodic tasks (elections and heart-beats).

TODO: Raft workflow & understanding on figure 2
