fn multiply(a: &[i64], b: &[i64], n: usize) -> Vec<i64> {
    let mut c = vec![0i64; n * n];
    for i in 0..n {
        for j in 0..n {
            let mut sum = 0;
            for k in 0..n {
                sum += a[i * n + k] * b[k * n + j];
            }
            c[i * n + j] = sum;
        }
    }
    c
}

fn main() {
    let n = std::hint::black_box(600usize);
    let mut a = vec![0i64; n * n];
    let mut b = vec![0i64; n * n];
    for i in 0..n {
        for j in 0..n {
            a[i * n + j] = ((i + j) % 10) as i64;
            b[i * n + j] = ((i * j) % 10) as i64;
        }
    }
    let c = multiply(&a, &b, n);
    let total: i64 = c.iter().sum();
    println!("{}", total);
}
