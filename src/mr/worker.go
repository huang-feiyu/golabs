package mr

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	ok, reply := GetTask()
	if ok {
		// TODO: do stuff
		switch reply.TaskType {
		case DONE:
			log.Printf("pid[%v]: done\n", os.Getpid())
			os.Exit(0)
		default:
			panic("Unimplemented")
		}
	} else {
		os.Exit(0)
	}
}

/*=== RPC request ===*/
func GetTask() (bool, *GetTaskReply) {
	args := GetTaskArgs{os.Getpid()}
	reply := GetTaskReply{}

	ok := call("Coordinator.HandleGetTask", &args, &reply)

	return ok, &reply
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
