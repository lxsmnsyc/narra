package main

import "base:intrinsics"
import "core:fmt"

fib :: proc(n: int) -> int {
	if n < 2 {
		return n
	}
	return fib(n - 1) + fib(n - 2)
}

main :: proc() {
	n := 38
	n = intrinsics.volatile_load(&n)
	fmt.println(fib(n))
}
