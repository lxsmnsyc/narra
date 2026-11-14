package main

import "fmt"

// Eight scalar accumulators, summed in the order of Simd.sum in dot.nra.
func dot(left, right []float32, passes int) float32 {
	var s [8]float32
	for pass := 0; pass < passes; pass++ {
		for i := 0; i+8 <= len(left) && i+8 <= len(right); i += 8 {
			for k := 0; k < 8; k++ {
				// The conversion rounds the product, so it is not fused into the add.
				s[k] += float32(left[i+k] * right[i+k])
			}
		}
	}
	return ((s[0] + s[4]) + (s[2] + s[6])) + ((s[1] + s[5]) + (s[3] + s[7]))
}

func main() {
	const count = 1048576
	left := make([]float32, count)
	right := make([]float32, count)
	for i := 0; i < count; i++ {
		left[i] = float32(i%16-8) * 0.125
		right[i] = float32(i%7-3) * 0.25
	}
	result := dot(left, right, 200)
	fmt.Println(int64(result * 32))
}
