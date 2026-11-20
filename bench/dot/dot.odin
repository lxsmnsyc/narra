package main

import "base:intrinsics"
import "core:fmt"

// Eight scalar accumulators, summed in the order of Simd.sum in dot.nra.
dot :: proc(left, right: []f32, passes: int) -> f32 {
	s: [8]f32
	for _ in 0 ..< passes {
		for i := 0; i + 8 <= len(left) && i + 8 <= len(right); i += 8 {
			for k in 0 ..< 8 {
				s[k] += left[i + k] * right[i + k]
			}
		}
	}
	return ((s[0] + s[4]) + (s[2] + s[6])) + ((s[1] + s[5]) + (s[3] + s[7]))
}

main :: proc() {
	count :: 1048576
	left := make([]f32, count)
	defer delete(left)
	right := make([]f32, count)
	defer delete(right)
	for i in 0 ..< count {
		left[i] = f32(i % 16 - 8) * 0.125
		right[i] = f32(i % 7 - 3) * 0.25
	}
	passes := 200
	passes = intrinsics.volatile_load(&passes)
	result := dot(left, right, passes)
	fmt.println(i64(result * 32))
}
