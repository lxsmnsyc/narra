// Eight scalar accumulators, summed in the order of Simd.sum in dot.nra.
fn dot(left: &[f32], right: &[f32], passes: usize) -> f32 {
    let mut s = [0.0f32; 8];
    for _ in 0..passes {
        let mut i = 0;
        while i + 8 <= left.len() && i + 8 <= right.len() {
            for k in 0..8 {
                s[k] += left[i + k] * right[i + k];
            }
            i += 8;
        }
    }
    ((s[0] + s[4]) + (s[2] + s[6])) + ((s[1] + s[5]) + (s[3] + s[7]))
}

fn main() {
    let count = 1048576;
    let mut left = Vec::with_capacity(count);
    let mut right = Vec::with_capacity(count);
    for i in 0..count as i64 {
        left.push((i % 16 - 8) as f32 * 0.125);
        right.push((i % 7 - 3) as f32 * 0.25);
    }
    let result = dot(&left, &right, std::hint::black_box(200));
    println!("{}", (result * 32.0) as i64);
}
