package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	files []string   // for each file, assign a Map task
	lock  sync.Mutex // lock to protect internal data
	cond  *sync.Cond // conditional variable to wait for all Map tasks done

	// intermediate file `mr-Y-X`
	nMapTasks    int // Y
	nReduceTasks int // X

	mapTasksIssued      []time.Time
	mapTasksFinished    []bool
	reduceTasksIssued   []time.Time
	reduceTasksFinished []bool

	isDone bool
}

/*=== RPC handlers ===*/
// HandleGetTask assigns one of the remaining tasks to the asking worker
func (c *Coordinator) HandleGetTask(args *GetTaskArgs, reply *GetTaskReply) error {
	log.Printf("Coordinator: pid[%v]: GetTask\n", args.Pid)
	c.lock.Lock()
	defer c.lock.Unlock()

	// 1. Issue Map tasks until no running Map task
	for {
		cnt := 0 // # of issued but un-finished tasks
		for i, done := range c.mapTasksFinished {
			// if the task is done
			if done {
				continue
			}
			// if not finish, assign an unissued task to the worker
			if c.mapTasksIssued[i].IsZero() || time.Since(c.mapTasksIssued[i]) > 10*time.Second {
				c.issueTask(MAP, i, reply)
				return nil
			}
			cnt++
		}
		if cnt == 0 {
			break // all Map done
		} else {
			log.Println("Coordinator: Wait when not all Map done")
			c.cond.Wait() // just block here, waiting for waken up to check again
		}
	}

	// 2. Issue Reduce tasks until everything is done
	for {
		cnt := 0 // # of issued but un-finished tasks
		for i, done := range c.reduceTasksFinished {
			// if the task is done
			if done {
				continue
			}
			// if not finish, assign an unissued task to the worker
			if c.reduceTasksIssued[i].IsZero() || time.Since(c.reduceTasksIssued[i]) > 10*time.Second {
				c.issueTask(REDUCE, i, reply)
				return nil
			}
			cnt++
		}
		if cnt == 0 {
			break // all Reduce done
		} else {
			log.Println("Coordinator: Wait when not all Reduce done")
			c.cond.Wait()
		}
	}
	// 3. All is done
	c.isDone = true
	c.issueTask(DONE, 0, reply)
	return nil
}

// HandleFinishTask gets the done message from a worker whose has done its task
func (c *Coordinator) HandleFinishTask(args *FinishTaskArgs, reply *FinishTaskReply) error {
	log.Printf("Coordinator: task %v[%v] done & Broadcast\n", args.TaskType, args.TaskID)
	c.lock.Lock()
	defer c.lock.Unlock()

	taskID := args.TaskID
	switch args.TaskType {
	case MAP:
		c.mapTasksFinished[taskID] = true
	case REDUCE:
		c.reduceTasksFinished[taskID] = true
	case DONE:
		panic("Unreachable: finish done")
	}
	c.cond.Broadcast() // notify all stopping at cond.Wait()

	if c.isDone {
		return os.ErrClosed // just a placeholder to tell worker, everything is done
	}
	return nil
}

// IssueTask to reduce duplicate code
func (c *Coordinator) issueTask(taskType TaskType, taskID int, reply *GetTaskReply) {
	log.Printf("Coordinator: Issue %v[%v]\n", taskType, taskID)
	reply.TaskType = taskType
	reply.TaskID = taskID
	reply.NReduceTasks = c.nReduceTasks
	reply.NMapTasks = c.nMapTasks

	switch taskType {
	case MAP:
		c.mapTasksIssued[taskID] = time.Now()
		reply.Filename = c.files[taskID]
	case REDUCE:
		c.reduceTasksIssued[taskID] = time.Now()
	}
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	// Q: wait for some workers?
	// A: No, workers can assume that when they cannot receive valid reply,
	//    the whole work is done
	c.lock.Lock()
	defer c.lock.Unlock()
	res := c.isDone

	return res
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.cond = sync.NewCond(&c.lock)

	// initialize maintained data
	c.files = files

	c.nMapTasks = len(files)
	c.mapTasksIssued = make([]time.Time, len(files))
	c.mapTasksFinished = make([]bool, len(files))

	c.nReduceTasks = nReduce
	c.reduceTasksIssued = make([]time.Time, nReduce)
	c.reduceTasksFinished = make([]bool, nReduce)

	go func() {
		// a goroutine periodically wakes coordinator up,
		// to avoid lost broadcast
		for {
			c.cond.L.Lock()
			log.Printf("Coordinator: Broadcast periodically")
			c.cond.Broadcast()
			c.cond.L.Unlock()
			time.Sleep(time.Second)
		}
	}()

	c.server()
	return &c
}
