package main

import "fmt"

type Particle struct {
	x, y, vx, vy int
}

func move(p Particle) Particle {
	nx := p.x + p.vx
	ny := p.y + p.vy
	nvx := p.vx
	nvy := p.vy
	if nx < 0 || nx > 10000 {
		nvx = -nvx
	}
	if ny < 0 || ny > 10000 {
		nvy = -nvy
	}
	return Particle{nx, ny, nvx, nvy}
}

func main() {
	var particles []Particle
	for i := 0; i < 1000; i++ {
		particles = append(particles, Particle{i * 7 % 10000, i * 13 % 10000, i%11 + 1, i%7 + 1})
	}
	for round := 0; round < 100000; round++ {
		for i := 0; i < 1000; i++ {
			particles[i] = move(particles[i])
		}
	}
	total := 0
	for _, p := range particles {
		total += p.x + p.y
	}
	fmt.Println(total)
}
