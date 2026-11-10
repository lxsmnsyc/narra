# Wave 19 report: batch-6

Batch-6 ran on `feat/batch-6` as one story: the rest of D94 (value semantics), the unbuilt parts of D92 (no hidden panics), D96 (`typeof`) and the spec's missing std functions. Children landed as direct commits after a worker check and a fresh review.

## Landed

"Compiler" is `narrac emit selfhost`, instructions retired, each compiler on its own source.

- type-of: `typeof NAME` in type positions, E0981 (cycle) and E0982 (type not known yet).
- proven-operands: divisors and shift counts need a proof, E0979 and E0980. Review fixes: sized constants that wrap to 0, and `u64` bounds at or above 2^63.
- unreachable-arms: E0983 for an arm no value can reach, including any arm after `_`.
- from-marker-followups: E0047 for `from` with no name, and the batch-5 follow-ups.
- shared-write-check: the shared-write rule (E0978) as a warning in `check`. Review fixes: for-mutable loops, element writes, the loop round cap, and an exponential loop walk (34.7G to 1.67G on `check selfhost`).
- std-spec-functions: the std functions SPEC lists, with Narra-source companions in `std/ext/`.
- std-no-panic: 11 std functions return `Option`, `Maps.update`, alloc charged only for heap enums. Temporary text helpers cost compiler +32% until the end of the batch.
- std-compound-sweep: the remaining compound std calls moved to the `Option` forms.
- shared-write-sweep and shared-write-sweep-rest: the compiler and test runner are clean of E0978 and E0976, and the `shared-results` switch is gone. `copy(...)` accepts enums with a tracked payload. A held-type cache keeps E0976 at about +3% (it was +66%).
- shared-write-error: E0978 is an error in every command and in the language server. The debug copy check covers heap enum payloads and runs for every program sample. Different literal indexes are separate places. Compiler +22.3% (+0.55G), above the +0.33G the brief expected.
- value-semantics-docs: SPEC section 6, D81 rule 5, D27 and the guide describe the rule and the `copy(...)` fix.
- End of batch: the text helpers call the Option builtins directly (compiler 3.02G to 2.44G), and the seed is regenerated.

Whole batch: compiler 1.74G to 2.47G instructions (+42%), peak memory 125 MB to 135 MB, `check selfhost` 1.24G to 1.93G. Most of it is the shared-write analysis now running in every command (+0.55G). The `from` check on the compiler's own code adds about +0.08G. The source grew 6.4% (53,459 to 56,903 lines). The inbox has a profiling item for `sharedwrites.nra`.

## Whole-branch review

No blockers. The review probed the interactions between children: proofs with `Option` results, `typeof` with shared writes, unreachable arms with the `Option` builtins, and the copy check with generators and enums. All behaved as SPEC says. One gap is new: elements that share one value (`Arrays.filled([0, 0], 3)`, `[row, row]`) are not tracked. It is now a known gap in SPEC section 6 and an inbox item. Two style nits went to the inbox.

## Process lessons

- No suite builds `bench/`, so proven-operands broke `bench/interp` unseen until the end-of-batch benchmark. Fixed with a checked `remainder` function; a suite is in the inbox.
- A probe that runs `narrac run` under a tool timeout orphans the program when it hangs: `narrac run` does not kill its child. 15 probes spun for 6 hours and kept the load average near 40. Cap probe programs with `perl -e 'alarm 20; exec @ARGV'` until the runtime forwards signals (inbox).
- A sweep sized from warning counts (198 E0978, 185 E0976) was too big for one node; splitting it into checker and rest nodes kept each review readable.
- Measure the cost of a check where it will run. shared-write-check measured `check` only; running the analysis in every command cost more than the brief expected.
