#include <stdint.h>
#include <stdio.h>

typedef struct {
    int64_t x;
    int64_t y;
    int64_t vx;
    int64_t vy;
} Particle;

static Particle advance(Particle p) {
    int64_t nx = p.x + p.vx;
    int64_t ny = p.y + p.vy;
    int64_t nvx = p.vx;
    int64_t nvy = p.vy;
    if (nx < 0 || nx > 10000) {
        nvx = -nvx;
    }
    if (ny < 0 || ny > 10000) {
        nvy = -nvy;
    }
    return (Particle){nx, ny, nvx, nvy};
}

int main(void) {
    Particle particles[1000];
    for (int64_t i = 0; i < 1000; i++) {
        particles[i] = (Particle){i * 7 % 10000, i * 13 % 10000, i % 11 + 1, i % 7 + 1};
    }
    volatile int64_t rounds = 100000;
    for (int64_t round = 0; round < rounds; round++) {
        for (int k = 0; k < 1000; k++) {
            particles[k] = advance(particles[k]);
        }
    }
    int64_t total = 0;
    for (int k = 0; k < 1000; k++) {
        total += particles[k].x + particles[k].y;
    }
    printf("%lld\n", (long long)total);
    return 0;
}
