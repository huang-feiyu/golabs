package labutil

import (
	"fmt"
)

func Assert(cond bool, msg ...interface{}) {
	if !cond {
		fmt.Printf("Assert failure: %v\n", msg...)
		panic(msg)
	}
}
