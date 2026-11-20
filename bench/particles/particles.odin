package main

import "base:intrinsics"
import "core:fmt"

Particle :: struct {
	x, y, vx, vy: int,
}

advance :: proc(p: Particle) -> Particle {
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

main :: proc() {
	particles: [dynamic]Particle
	defer delete(particles)
	for i in 0 ..< 1000 {
		append(&particles, Particle{i * 7 % 10000, i * 13 % 10000, i % 11 + 1, i % 7 + 1})
	}
	rounds := 100000
	rounds = intrinsics.volatile_load(&rounds)
	for _ in 0 ..< rounds {
		for i in 0 ..< 1000 {
			particles[i] = advance(particles[i])
		}
	}
	total := 0
	for p in particles {
		total += p.x + p.y
	}
	fmt.println(total)
}
