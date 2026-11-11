# Log

- Built stage 2 with `sh selfhost/bootstrap.sh`; wrote `bench.nra` modeled on `leaks.nra`; formatted with `narrac fmt`.
- Suite on clean tree: 6 passed.
- Failing-benchmark demonstration: changed `return n` to `return "n"` in `bench/fib/fib.nra`; suite output `FAIL bench/fib` with `fib.nra:5:12: expected int, found string`; reverted with `git checkout`.
- `--suite sources` passed; `sh scripts/test.sh` passed (6567 cases, 213s).
