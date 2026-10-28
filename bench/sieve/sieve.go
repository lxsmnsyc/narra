package main

import "fmt"

func countPrimes(n int) int {
	marks := make([]int, n)
	for i := range marks {
		marks[i] = 1
	}
	marks[0] = 0
	marks[1] = 0
	for i := 2; i*i < n; i++ {
		if marks[i] == 1 {
			for j := i * i; j < n; j += i {
				marks[j] = 0
			}
		}
	}
	count := 0
	for k := 0; k < n; k++ {
		count += marks[k]
	}
	return count
}

func main() {
	fmt.Println(countPrimes(20000000))
}
