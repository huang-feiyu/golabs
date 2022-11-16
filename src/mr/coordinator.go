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
	lock  sync.Mutex // lock to protect internal data (better to use an RWMutex)
	cond  sync.Cond  // conditional variable to wait for all Map tasks done

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
	log.Printf("pid[%v]: GetTask\n", args.Pid)
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
				log.Printf("coordinator: Issue %v[%v]\n", MAP, i)
				c.mapTasksIssued[i] = time.Now()
				reply.TaskType = MAP
				reply.TaskID = i
				reply.NReduceTasks = c.nReduceTasks
				reply.Filename = c.files[i]
				reply.NMapTasks = c.nMapTasks
				return nil
			}
			cnt++
		}
		if cnt == 0 {
			break // all Map done
		} else {
			c.cond.Wait()
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
	log.Printf("coordinator: task %v[%v] done\n", args.TaskType, args.TaskID)
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
	c.cond.Broadcast()

	if c.isDone {
		return os.ErrClosed // just a placeholder to tell worker, everything is done
	}
	return nil
}

// IssueTask to reduce duplicate code
func (c *Coordinator) issueTask(taskType TaskType, taskID int, reply *GetTaskReply) {
	log.Printf("coordinator: Issue %v[%v]\n", taskType, taskID)
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
	ret := false

	// Your code here.

	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// initialize maintained data
	c.files = files

	c.nMapTasks = len(files)
	c.mapTasksIssued = make([]time.Time, len(files))
	c.mapTasksFinished = make([]bool, len(files))

	c.nReduceTasks = nReduce
	c.reduceTasksIssued = make([]time.Time, nReduce)
	c.reduceTasksFinished = make([]bool, nReduce)

	c.server()
	return &c
}
