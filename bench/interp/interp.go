package main

import "fmt"

type Op interface{ isOp() }
type Add struct{ n int }
type Sub struct{ n int }
type Mul struct{ n int }
type Mod struct{ n int }
type Nop struct{}

func (Add) isOp() {}
func (Sub) isOp() {}
func (Mul) isOp() {}
func (Mod) isOp() {}
func (Nop) isOp() {}

func step(op Op, acc int) int {
	switch o := op.(type) {
	case Add:
		return acc + o.n
	case Sub:
		return acc - o.n
	case Mul:
		return acc * o.n
	case Mod:
		return acc % o.n
	}
	return acc
}

func main() {
	var program []Op
	for i := 0; i < 1000; i++ {
		switch i % 5 {
		case 0:
			program = append(program, Add{i})
		case 1:
			program = append(program, Mul{3})
		case 2:
			program = append(program, Sub{7})
		case 3:
			program = append(program, Mod{1000003})
		default:
			program = append(program, Nop{})
		}
	}
	acc := 1
	for round := 0; round < 150000; round++ {
		for _, op := range program {
			acc = step(op, acc)
		}
	}
	fmt.Println(acc)
}
