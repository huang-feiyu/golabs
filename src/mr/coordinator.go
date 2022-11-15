package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type Coordinator struct {
	files []string // for each file, assign a Map task

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
// HandleGetTask assigns one of the remaining tasks to the asking work
func (c *Coordinator) HandleGetTask(args *GetTaskArgs, reply *GetTaskReply) error {
	log.Printf("pid[%v]: GetTask\n", args.Pid)

	// TODO: Assign tasks according to different phases
	// for now, just send a Done message to worker
	reply.TaskType = DONE
	reply.TaskID = 0
	reply.NReduceTasks = c.nReduceTasks
	reply.Filename = c.files[0]
	reply.NMapTasks = c.nMapTasks

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
