package main

import "base:intrinsics"
import "core:fmt"

main :: proc() {
	s := ""
	count := 100000
	count = intrinsics.volatile_load(&count)
	for i in 0 ..< count {
		next := fmt.aprintf("%s%d", s, i % 10)
		delete(s)
		s = next
	}
	fmt.println(len(s))
	delete(s)
}
