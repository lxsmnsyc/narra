const base = 1
function run(): number {
  const base = 10
  function add(x: number): number { return x + base }
  return add(1)
}
console.log(run())
