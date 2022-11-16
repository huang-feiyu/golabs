package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var TMP_DIR string = "/tmp/mr/"
var CUR_DIR string = "./"
var I_FILE_PT string = "mr-$1-$2" // $1 => Yth map, $2 => Xth reduce
var O_FILE_PT string = "mr-out-$2"

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

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
// (2) Reduce: Read the intermediate files, collect the K/V pairs. Sort K/V pairs
//             by key. Apply the reduce function for each distinct key. Create temp
//             file to write, rename to the final name `mr-out-X`.
// (3) Done: just return.
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {
	for {
		ok, reply := GetTask()
		if ok {
			log.Printf("Worker: Start %v[%v]\n", reply.TaskType, reply.TaskID)
			switch reply.TaskType {
			case MAP:
				performMap(reply.TaskID, reply.NReduceTasks, reply.Filename, mapf)
			case REDUCE:
				performReduce(reply.TaskID, reply.NMapTasks, reducef)
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
			log.Printf("pid[%v]: done\n", os.Getpid())
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
	for i := 0; i < nReduce; i++ {
		file, err := ioutil.TempFile(TMP_DIR, "tmp-map")
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
		newName := newName(taskID, i, TMP_DIR, I_FILE_PT)
		err = os.Rename(file.Name(), newName)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		log.Printf("Map: Done process of %v\n", newName)
	}
}

/*=== Reduce task ===*/
// Reduce: Read the intermediate files, collect the K/V pairs. Sort K/V pairs
//         by key. Apply the reduce function for each distinct key. Create temp
//         file to write, rename to the final name `mr-out-X`.
func performReduce(taskID, nMap int, reducef func(string, []string) string) {
	// 1. read intermediate input K/V from files
	kva := []KeyValue{}
	for i := 0; i < nMap; i++ {
		filename := newName(i, taskID, TMP_DIR, I_FILE_PT)
		fp, err := os.Open(filename)
		defer os.Remove(fp.Name()) // remove intermediate file
		if err != nil {
			log.Fatalf("cannot open %v", filename)
		}
		dec := json.NewDecoder(fp)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv)
		}
	}

	// 2. sort the intermediate K/V pairs
	sort.Sort(ByKey(kva))

	// 3. create temp file
	file, err := ioutil.TempFile(TMP_DIR, "tmp-reduce")
	if err != nil {
		log.Fatal(err)
	}

	// 4. Apply the reduce function
	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}
		output := reducef(kva[i].Key, values)
		fmt.Fprintf(file, "%v %v\n", kva[i].Key, output)
		i = j
	}

	// 5. Rename to final output name (Because of different file system, need to copy and rename)
	endName := newName(0, taskID, CUR_DIR, O_FILE_PT)
	moveFile(file.Name(), endName)
	log.Printf("Reduce: Done process of %v\n", endName)
}

/*=== Utils ===*/
func newName(Y, X int, dir string, pattern string) string {
	pattern = dir + pattern
	newName := strings.Replace(pattern, "$1", strconv.Itoa(Y), 1)
	newName = strings.Replace(newName, "$2", strconv.Itoa(X), 1)
	return newName
}

func moveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
