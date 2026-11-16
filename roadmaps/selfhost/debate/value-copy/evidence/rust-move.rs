fn main() {
    let mut prices = vec![100, 200];
    let synonym = prices;          // moves: prices is no longer usable
    prices.push(300);
    println!("{}", synonym[0]);
}
