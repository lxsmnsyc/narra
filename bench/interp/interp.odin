package main

import "base:intrinsics"
import "core:fmt"

Add :: struct {
	n: int,
}
Sub :: struct {
	n: int,
}
Mul :: struct {
	n: int,
}
Mod :: struct {
	n: int,
}
Nop :: struct {}

Op :: union {
	Add,
	Sub,
	Mul,
	Mod,
	Nop,
}

step :: proc(op: Op, acc: int) -> int {
	switch o in op {
	case Add:
		return acc + o.n
	case Sub:
		return acc - o.n
	case Mul:
		return acc * o.n
	case Mod:
		return acc % o.n
	case Nop:
		return acc
	}
	return acc
}

main :: proc() {
	program: [dynamic]Op
	defer delete(program)
	for i in 0 ..< 1000 {
		switch i % 5 {
		case 0:
			append(&program, Add{i})
		case 1:
			append(&program, Mul{3})
		case 2:
			append(&program, Sub{7})
		case 3:
			append(&program, Mod{1000003})
		case:
			append(&program, Nop{})
		}
	}
	acc := 1
	rounds := 150000
	rounds = intrinsics.volatile_load(&rounds)
	for _ in 0 ..< rounds {
		for op in program {
			acc = step(op, acc)
		}
	}
	fmt.println(acc)
}
