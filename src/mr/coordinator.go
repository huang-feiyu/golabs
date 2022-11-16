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
	lock  sync.Mutex // lock to protect internal data (better to use an RWMutex)
	files []string   // for each file, assign a Map task

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

	// Issue Map tasks until all finished
	// TODO: Need a wait for all finished instead of issuing tasks
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
			panic("Unimplemented: Wait for all Map tasks done")
		}
	}

	panic("Unimplemented: issue Reduce task")

	return nil
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
