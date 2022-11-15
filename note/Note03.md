# GFS

> [The Google File System](https://static.googleusercontent.com/media/research.google.com/en//archive/gfs-sosp2003.pdf)

[TOC]

### Storage systems

> Building fault-tolerance storage system

* app=state less
* storage holds persistent state

Why hard?

* High performance => Shard data across servers
* many servers => constant faults
* fault tolerance -> replication
  * replication -> potential inconsistencies
  * strong consistency => lower performance

### Consistency

> [CSE 138 Consistency](https://github.com/huang-feiyu/CSE138-Notes/blob/main/Notes/Note10.md)

* Ideal consistency: Behave as if single system

---

How to get stronger consistency?

* Update all P+S or none
* ...

### GFS

##### Overview

* Performance = replication + FT consistency
* Successful system => 1000s machines
  * non-standard:<br/>(1) single Master<br/>(2) also has in-consistencies

gfs: file system for MapReduce

* Big: large data set
* Fast: automatic sharding
* Global: all apps see same fs
* Fault tolerant: automatic

##### Design

Size of a chunk is 64 MB.

![Arch](https://user-images.githubusercontent.com/70138429/201810367-14ae7466-4802-4d0e-b0cd-cc079779fec7.png)

* Master: states master maintains
  * file name -> array of chunk handles (stable storage [in log])
  * chunk handle ->
    * version # (stable storage)
    * list of chunk servers; primary, secondaries; lease time (volatile storage)
  * log(stable storage) + checkpoint(stable storage)

##### Read

1. Client sends filename+offset -> Master
2. Master -> Client: chunk handle + list of chunk servers + Version#
3. Client caches list  (less communication with Master)
4. Client reads from **closest** server
5. Server checks version#, if OK -> send data

##### Write

Write a file: append

1. Client -> Master: fn -> chunk handles
2. Master -> Client: handles -> list of servers (inc V#+P(lease)+S)
3. Primary checks V# & lease, if all valid -> write at specific offset (*at-least-offset*)

![Write](https://user-images.githubusercontent.com/70138429/201812426-504c8345-f5b3-43c2-b67e-902f2e89a5fa.png)

