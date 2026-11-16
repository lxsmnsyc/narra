fn main() {
    // Fixed-size array is Copy: b = a copies.
    let mut a = [100, 200];
    let b = a;
    a[0] = 9;
    println!("array: {:?} {:?}", a, b);

    // Vec: let b = a moves. To keep both, clone (explicit copy) or borrow.
    let mut prices = vec![100, 200];
    let synonym = prices.clone();
    prices.push(300);
    println!("clone: {:?} {:?}", prices, synonym);

    // A shared borrow cannot coexist with a write.
    let mut p = vec![100, 200];
    let r = &p;
    println!("borrow: {:?}", r);
    p.push(300);
    println!("after the borrow ends: {:?}", p);
}
