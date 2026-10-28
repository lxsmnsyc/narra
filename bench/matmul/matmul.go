package main

import "fmt"

func multiply(a, b []int, n int) []int {
	c := make([]int, n*n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			sum := 0
			for k := 0; k < n; k++ {
				sum += a[i*n+k] * b[k*n+j]
			}
			c[i*n+j] = sum
		}
	}
	return c
}

func main() {
	n := 600
	a := make([]int, n*n)
	b := make([]int, n*n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			a[i*n+j] = (i + j) % 10
			b[i*n+j] = (i * j) % 10
		}
	}
	c := multiply(a, b, n)
	total := 0
	for _, x := range c {
		total += x
	}
	fmt.Println(total)
}
