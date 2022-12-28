# Lab 2A_2B Q&A

[TOC]

### Debugging

* Run test case
* Log all messages

### Structure

A coroutine for applyCh.

Raft lock serializes all threads and RPC handlers atomic.

No locks across RPCs.

![Structure](https://user-images.githubusercontent.com/70138429/209463675-c60cec51-8fac-4166-97f9-583be4afd538.png)
