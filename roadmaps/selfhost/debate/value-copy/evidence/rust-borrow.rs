fn main() {
    let mut prices = vec![100, 200];
    let synonym = &prices;
    prices.push(300);
    println!("{}", synonym[0]);
}
