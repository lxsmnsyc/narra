package main

import (
	"fmt"
	"strconv"
)

func main() {
	s := ""
	for i := 0; i < 100000; i++ {
		s = s + strconv.Itoa(i%10)
	}
	fmt.Println(len(s))
}
