package main

import (
	"fmt"
	"time"
)


func main() {
	start := time.Now()
	var count = 0
	for i := 0; i <= 10000000; i++ {
		count += i
		fmt.Println(count)
		time.Sleep(time.Millisecond * 10) 
	}
	fmt.Println(count, "elapsed", time.Since(start))
}
