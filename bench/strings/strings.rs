fn main() {
    let mut s = String::new();
    for i in 0..std::hint::black_box(100000i64) {
        s = format!("{}{}", s, i % 10);
    }
    println!("{}", s.len());
}
