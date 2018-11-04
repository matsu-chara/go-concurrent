package main

import (
	"fmt"
	"time"
)

var data int

func main() {
	go func() {
		data++
	}()

	time.Sleep(1 * (time.Second))
	if data == 0 {
		fmt.Printf("the value is %v.\n", data)
	}
}
