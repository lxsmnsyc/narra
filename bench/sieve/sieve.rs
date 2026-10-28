fn count_primes(n: usize) -> i64 {
    let mut marks = vec![1i64; n];
    marks[0] = 0;
    marks[1] = 0;
    let mut i = 2;
    while i * i < n {
        if marks[i] == 1 {
            let mut j = i * i;
            while j < n {
                marks[j] = 0;
                j += i;
            }
        }
        i += 1;
    }
    let mut count = 0;
    for k in 0..n {
        count += marks[k];
    }
    count
}

fn main() {
    println!("{}", count_primes(std::hint::black_box(20000000)));
}
