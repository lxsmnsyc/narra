#[derive(Clone, Copy)]
struct Particle {
    x: i64,
    y: i64,
    vx: i64,
    vy: i64,
}

fn advance(p: Particle) -> Particle {
    let nx = p.x + p.vx;
    let ny = p.y + p.vy;
    let mut nvx = p.vx;
    let mut nvy = p.vy;
    if nx < 0 || nx > 10000 {
        nvx = -nvx;
    }
    if ny < 0 || ny > 10000 {
        nvy = -nvy;
    }
    Particle { x: nx, y: ny, vx: nvx, vy: nvy }
}

fn main() {
    let mut particles = Vec::new();
    for i in 0..1000i64 {
        particles.push(Particle { x: i * 7 % 10000, y: i * 13 % 10000, vx: i % 11 + 1, vy: i % 7 + 1 });
    }
    for _ in 0..std::hint::black_box(100000) {
        for i in 0..1000 {
            particles[i] = advance(particles[i]);
        }
    }
    let total: i64 = particles.iter().map(|p| p.x + p.y).sum();
    println!("{}", total);
}
