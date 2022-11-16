package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

/*=== Data structures passing via RPC ===*/
// TaskType contains 3 types: MAP, REDUCE, DONE
type TaskType string

const (
	MAP    TaskType = "Map"
	REDUCE          = "Reduce"
	DONE            = "Done"
)

/*=== GetTask RPC ===*/
// GetTaskArgs is used for AskTask RPC (worker -> coordinator)
type GetTaskArgs struct {
	Pid int // for debugging
}

// GetTaskReply is used for AskTask RPC reply (coordinator -> worker)
type GetTaskReply struct {
	TaskType TaskType
	TaskID   int

	// For map
	NReduceTasks int // to know how many files to write, the X in `mr-Y-X`
	Filename     string

	// For reduce
	NMapTasks int // to know how many files to read, the Y in `mr-Y-X`
}

/*=== FinishTask RPC=== */
// FinishTaskArgs is used for FinishTask RPC (worker -> coordinator)
type FinishTaskArgs struct {
	TaskType TaskType
	TaskID   int
}

// FinishTaskReply is used for FinishTask RPC (worker -> coordinator)
type FinishTaskReply struct {
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
