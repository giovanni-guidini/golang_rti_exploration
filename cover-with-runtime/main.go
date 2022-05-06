package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
)

func getGID() uint64 {
	b := make([]byte, 256)
	b = b[:runtime.Stack(b, false)]
	// fmt.Printf("%s\n", string(b))
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func LeftFunction(channel chan int) {
	RecordLineCoverage(runtime.Caller(0))
	if _, file, line, ok := runtime.Caller(0); ok {
		RecordLineCoverage(runtime.Caller(0))
		fmt.Printf("left (gid %d): %s:%d\n", getGID(), file, line)
	}
	RecordLineCoverage(runtime.Caller(0))
	channel <- 0
}

func RightFunction(channel chan int) {
	RecordLineCoverage(runtime.Caller(0))
	if _, file, line, ok := runtime.Caller(0); ok {
		RecordLineCoverage(runtime.Caller(0))
		fmt.Printf("right (gid %d): %s:%d\n", getGID(), file, line)
	}
	RecordLineCoverage(runtime.Caller(0))
	channel <- 1
}

func RecordLineCoverage(pc uintptr, file string, line int, ok bool) {
	Seen[line] += 1
}

var Seen []int

func main() {
	Seen = make([]int, 50)
	channel := make(chan int)
	seen := [2]bool{false, false}

	go LeftFunction(channel)
	go RightFunction(channel)

	for seen[0] == false || seen[1] == false {
		n := <-channel
		seen[n] = true
	}

	fmt.Printf("Exiting.\nSeen = {\n")
	for i, v := range Seen {
		if v > 0 {
			fmt.Printf("[%d]: %d\n", i, v)
		}
	}
	fmt.Printf("}\n")
	fmt.Println("Entries from Seen not appearing are 0")
}
