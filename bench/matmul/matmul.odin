package main

import "base:intrinsics"
import "core:fmt"

multiply :: proc(a, b: []int, n: int) -> []int {
	c := make([]int, n * n)
	for i in 0 ..< n {
		for j in 0 ..< n {
			sum := 0
			for k in 0 ..< n {
				sum += a[i * n + k] * b[k * n + j]
			}
			c[i * n + j] = sum
		}
	}
	return c
}

main :: proc() {
	n := 600
	n = intrinsics.volatile_load(&n)
	a := make([]int, n * n)
	defer delete(a)
	b := make([]int, n * n)
	defer delete(b)
	for i in 0 ..< n {
		for j in 0 ..< n {
			a[i * n + j] = (i + j) % 10
			b[i * n + j] = (i * j) % 10
		}
	}
	c := multiply(a, b, n)
	defer delete(c)
	total := 0
	for x in c {
		total += x
	}
	fmt.println(total)
}
