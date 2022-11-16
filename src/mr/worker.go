package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
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

// Worker periodically asks the coordinator for work, sleeping with
// time.Sleep() between each request.
// For different types of tasks:
// (1) Map: Perform the map function on each K/V pair, store them in
//          NReduce intermediate files, finally atomically rename them
//          to `mr-Y-X`.
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {
	for {
		ok, reply := GetTask()
		if ok {
			log.Printf("Worker: Start %v[%v]\n", reply.TaskType, reply.TaskID)
			switch reply.TaskType {
			case MAP:
				performMap(reply.TaskID, reply.NReduceTasks, reply.Filename, mapf)
			case REDUCE:
				log.Printf("Unimplemented: performReduce")
				os.Exit(0)
			case DONE:
				log.Printf("pid[%v]: done\n", os.Getpid())
				os.Exit(0)
			default:
				panic("Unknown type of task")
			}
		} else {
			os.Exit(-1)
		}
		ok, _ = FinishTask(reply.TaskType, reply.TaskID)
		if !ok {
			os.Exit(0)
		}
		time.Sleep(time.Second)
	}
}

/*=== RPC request ===*/
func GetTask() (bool, *GetTaskReply) {
	args := GetTaskArgs{os.Getpid()}
	reply := GetTaskReply{}
	ok := call("Coordinator.HandleGetTask", &args, &reply)
	return ok, &reply
}

func FinishTask(taskType TaskType, taskID int) (bool, *FinishTaskReply) {
	args := FinishTaskArgs{taskType, taskID}
	reply := FinishTaskReply{}
	ok := call("Coordinator.HandleFinishTask", &args, &reply)
	return ok, &reply
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
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

/*=== Map task ===*/
// Map: Perform the map function on each K/V pair, store them in
//      NReduce intermediate files, finally atomically rename them
//      to `mr-Y-X`.
func performMap(taskID, nReduce int, filename string, mapf func(string, string) []KeyValue) {
	// 0. read raw input K/V from file
	intermediate := []KeyValue{}
	fp, err := os.Open(filename)
	defer fp.Close()
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := ioutil.ReadAll(fp)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}

	// 1. perform map function on each K/V pair
	kva := mapf(filename, string(content))
	intermediate = append(intermediate, kva...)

	// 2. create nReduce i-files, store K/V pairs in its bucket (use JSON format)
	pattern := "/tmp/mr/mr-$1-$2"
	for i := 0; i < nReduce; i++ {
		file, err := ioutil.TempFile("/tmp/mr", "tmp-map")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(file.Name())
		enc := json.NewEncoder(file)
		for _, kv := range kva { // for convenience
			if ihash(kv.Key)%nReduce == i {
				enc.Encode(&kv)
			}
		}
		// 3. rename to `mr-Y-X`, even if another worker cover this, it is still safe and atomic
		newName := strings.Replace(pattern, "$1", strconv.Itoa(taskID), 1)
		newName = strings.Replace(newName, "$2", strconv.Itoa(i), 1)
		err = os.Rename(file.Name(), newName)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		log.Printf("Map: Done process of %v\n", newName)
	}
}
