package main

import "fmt"

func main() {
	// Fixed-size array: a value. b := a copies.
	a := [2]int{100, 200}
	b := a
	a[0] = 9
	fmt.Println("array:", a, b)

	// Slice (what Go code uses): a header pointing at a shared backing array.
	s := []int{100, 200}
	t := s
	s[0] = 9
	fmt.Println("slice write:", s, t)

	// append with no spare capacity: s gets a new backing array, t keeps the old one.
	s = append(s, 300)
	s[1] = 7
	fmt.Println("after append, no capacity:", s, t)

	// append with spare capacity: both still share.
	u := make([]int, 2, 10)
	v := u
	u = append(u, 300)
	u[0] = 5
	fmt.Println("after append, spare capacity:", u, v)

	// Struct holding an array field copies; map is a reference.
	m := map[string]int{"a": 1}
	n := m
	m["a"] = 2
	fmt.Println("map:", m, n)
}
