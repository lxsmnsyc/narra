enum Op {
    Add(i64),
    Sub(i64),
    Mul(i64),
    Mod(i64),
    Nop,
}

fn step(op: &Op, acc: i64) -> i64 {
    match op {
        Op::Add(n) => acc + n,
        Op::Sub(n) => acc - n,
        Op::Mul(n) => acc * n,
        Op::Mod(n) => acc % n,
        Op::Nop => acc,
    }
}

fn main() {
    let mut program = Vec::new();
    for i in 0..1000i64 {
        program.push(match i % 5 {
            0 => Op::Add(i),
            1 => Op::Mul(3),
            2 => Op::Sub(7),
            3 => Op::Mod(1000003),
            _ => Op::Nop,
        });
    }
    let mut acc = 1i64;
    for _ in 0..std::hint::black_box(150000) {
        for op in &program {
            acc = step(op, acc);
        }
    }
    println!("{}", acc);
}
