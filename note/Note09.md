# ZooKeeper

> [ZooKeeper](https://pdos.csail.mit.edu/6.824/papers/zookeeper.pdf)
>
> [Notion: ZooKeeper](https://www.notion.so/huangfeiyu/ZooKeeper-c1e722d06cf44221b7a6e7ce557e0570)

### High Performance

ZooKeeper: change correctness definition

* Linearizable writes
* FIFO client order

=> Read & Write

* Write: writes in client order
* Read
  * Prefix of the log => No back in time
  * Observe last write from same client

---

ZooKeeper uses `zxid` as the index of the log, read can only read from higher `zxid` values. Client can read from any followers. Client updates its `zxid` with writes response from leader.

### Rules help programming

Before read, you can certainly use `sync` to get *real linearizability*. For now, ignore it.

* Client must read newer value with `zxid`. `zxid` is the last write's that client observed.
* Client can watch/subscribe a data object. If the data object changes, client will receive a notification.

### Summary

Weaker consistency => well-designed APIs which can be used in lots of critical fields

High performance (because of weaker consistency)

