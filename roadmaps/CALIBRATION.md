# Calibration

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
