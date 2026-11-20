package main

import "base:intrinsics"
import "core:fmt"

count_primes :: proc(n: int) -> int {
	marks := make([]int, n)
	defer delete(marks)
	for i in 0 ..< n {
		marks[i] = 1
	}
	marks[0] = 0
	marks[1] = 0
	for i := 2; i * i < n; i += 1 {
		if marks[i] == 1 {
			for j := i * i; j < n; j += i {
				marks[j] = 0
			}
		}
	}
	count := 0
	for k in 0 ..< n {
		count += marks[k]
	}
	return count
}

main :: proc() {
	n := 20000000
	n = intrinsics.volatile_load(&n)
	fmt.println(count_primes(n))
}
