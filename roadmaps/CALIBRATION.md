# Calibration

Notes:

- selfhost M nodes took 178k-607k tokens in wave 4 (median about 390k).
- Nodes that touch the checker edit `check.go`, `check_call.go`, `modules_check.go`, `cnames.go`, `diagnostic_codes.go` and `compiler/README.md`, whatever the brief says. List them as shared up front.

| Roadmap | Node | Estimate | Actual | Review rounds | Note |
| --- | --- | --- | --- | --- | --- |
| selfhost | split-compiler | M | M | 2 | Round 1 found two changed error messages that no test covered. |
| selfhost | bench-harness | M | M | 2 | Round 1 found an all-skipped row that reported ok. |
| selfhost | modules | L | L | 2 | Round 1 found types compared by name only across modules, and a C name collision through `__`. |
| selfhost | traits | L | L | 2 | Built in parallel with modules, then rebased onto it. Rounds found a constructor lookup that depended on map order, an implementation name collision, and function types that never unified. |
| selfhost | closures | L | L | 2 | Round 1 found two use-after-free bugs, both proven by ASan: a stack environment escaping through a trait, and a callee freed during its own call. |
| selfhost | patterns | L | L+ | 3 | Grew mid-node when the user changed the tuple syntax to `#[...]`. Rounds found tuple C name collisions twice. |
| selfhost | poly-recursion | S | S | 1 | Found by a hang: on main the new error tests time out at 15 s. |
| selfhost | hashed-cnames | M | M | 2 | Touched about 20 files, and the generated C differed in definition order for 4 of 31 samples. Needed a 31-sample before/after comparison. |
| selfhost | box-recursive | M | M | 2 | Heap payloads changed every enum read. A bench claim in review did not reproduce, and `interp` is a real 2x slowdown, accepted until perceus. |
| selfhost | io-std | M | M | 1 | Needed a hotfix (PR #11): merged before rebasing onto box-recursive, and main broke. |
| selfhost | strings-std | M | M | 2 | 331k tokens. Review found `StringBuilder` nameable without its runtime emitted. Fixed a bug that made every `Mutable` builtin unusable. |
| selfhost | maps | M | M | 2 | 468k tokens. Review found the key Equatable/Hashable check skipped in signatures. Touched 17 files. |
| selfhost | effects | M | M | 2 | 490k tokens. Review found private higher-order functions wrongly strict. Six existing samples needed `uses`. |
| selfhost | json-diagnostics | M | M | 2 | 385k tokens. Review found `clone` dropping `StringLit.End`, which broke a fix in generics. Codes by format string kept hot files nearly untouched. |
| selfhost | selfhost-spike | M | M | 2 | 399k tokens. Gate. Review asked for doc fixes. 372 Go lexer lines became 545 Narra lines (1.45x). |
| selfhost | std-convert | M | M | - | 178k tokens. Spike guessed S; nine small functions and a runtime fix. |
| selfhost | opaque-distinct | M | M | - | 340k tokens. Edited 13 files outside its owned paths, a few lines each. |
| selfhost | top-level-const | M | M | 2 | 607k tokens, the most of the wave. Review found a constant/type name clash undetected. Lazy getters replaced static data. |
