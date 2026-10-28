# bench

The same programs in Narra, Go and Rust, and a script that times them.

```sh
bench/run.sh            # all programs, 7 runs each
bench/run.sh -n 3 fib   # 3 runs, one program
```

`run.sh` builds each program in release mode into `bench/build/` (ignored by git), runs it, and prints the median wall-clock seconds per language. It also checks that every language prints the same answer. If `go`, `rustc` or a C compiler is missing, that language is skipped and shown as `-`.

Programs live in `bench/<name>/<name>.nra`, `.go` and `.rs`: `fib` (recursion), `sieve` and `matmul` (arrays), `strings` (string building), `interp` (enum and match), `particles` (structs in an array).

Run it on a quiet machine. Results are in `BASELINE.md`.

The Go and Rust programs are plain idiomatic code that mirror the Narra algorithm. Add a program by adding the three files and its name to the list in `run.sh`.
