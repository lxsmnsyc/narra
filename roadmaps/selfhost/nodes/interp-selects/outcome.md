# Outcome: interp-selects

## Built

- A finding, no compiler change. The `interp` slowdown from D99 reproduces on `feat/batch-5` (+21% cycles). Two emitted forms remove it, and neither should be adopted.
- Why not: both forms turn the selects back into branches. Branches win only when the branch predictor learns the op order. The bench repeats a five-op cycle, which it learns perfectly. On random `Add`/`Sub` data the selects are 3.2x faster than branches. No emitted form wins on both, so no node is proposed. D99's cost line stands as written. The `pointer` form's effect on the compiler is unmeasured: the compiler probe only toggled early if-conversion.
- Cause: the for-of item is a by-value copy of the array element, so tag and payload are loaded together before the match. The payload is then already in a register. The AArch64 backend's early if-conversion turns the `Add` and `Sub` arms into `add`, `sub` and two `csel`. That puts three operations on the accumulator's dependency chain, where a predicted branch puts one.
- `-mllvm -aarch64-enable-early-ifcvt=false` removes the selects. The SimplifyCFG phi-folding thresholds do not. Before D99 the payload sat behind a pointer and was loaded inside each arm, so the arms stayed branches.
- Forms that avoid the selects (hand-edited C, `forms/make.py`):
  - `pointer`: the loop passes `&items[i]` and the parameter is `const Op*`. The emitted if-chain reads tag and payload through it, so nothing copies the value between the array and the arm. Rust's `bench/interp/interp.rs` does the same (`step(op: &Op)`).
  - `asm`: an empty `__asm__("" : "+r"(n))` after each payload binding. It blocks constant folding of payloads, so it is not a real candidate.
- Forms that keep the selects:
  - a `switch` with each payload read inside its case,
  - a match that reads through a pointer to a by-value parameter,
  - a pointer parameter whose match copies `*op` into the subject temp.
- Any copy of the whole value before the tag test brings the selects back.
- Files in `roadmaps/selfhost/nodes/interp-selects/forms/`:
  - `gen.sh REPO SRC`: writes `SRC/inline.c` and `SRC/heap.c`. `inline.c` is `narrac emit bench/interp/interp.nra` with narrac bootstrapped at 9f8ee5f1. `heap.c` is the same command with a compiler bootstrapped at 9501a3cc^, before option-inline. It differs only in the enum layout. The emitted C is not committed.
  - `make.py SRC OUT`: writes every form in four data orders (32 files).
  - `measure.sh SRC WORKDIR [ROUNDS]`: runs `make.py`, builds the forms with the `narrac build -release` flags and runs them interleaved.
  - `compiler.sh REPO WORKDIR [ROUNDS]`: the compiler probe below.
- Rebuild, from the repo root, with `<tmp>` outside the repo:

  ```sh
  sh roadmaps/selfhost/nodes/interp-selects/forms/gen.sh . <tmp>/src
  sh roadmaps/selfhost/nodes/interp-selects/forms/measure.sh <tmp>/src <tmp>/forms 11
  sh roadmaps/selfhost/nodes/interp-selects/forms/compiler.sh "$PWD" <tmp>/compiler 9
  ```

## Proof

Apple M2, macOS 27, Apple clang 21.0.0, `-O2 -DNARRA_RELEASE -std=c11`. No gcc is installed (`/usr/bin/gcc` is clang). Every number is a median of 11 interleaved runs (9 for the compiler probe) under the shared bench lock. The table is one run of `measure.sh` at load average 2.8 to 3.7. Earlier runs in `log.md` saw load 4 to 9. Its `cycle` column matched three earlier runs within 1%. Cycles and instructions come from `/usr/bin/time -l`. User time moves in 0.01 s steps, so it is shown only for the bench order.

- base: `feat/batch-5` 9f8ee5f1, `sh forms/measure.sh <src> <dir> 11`. Row `inline-cycle`: 0.18 s user, 0.626 G cycles, 1.964 G instructions. Row `heap-cycle` (pre-D99 layout): 0.15 s, 0.519 G cycles, 1.937 G instructions. That is +21% cycles with 1.4% more instructions, as option-inline reported.
- branch: not built (spike).

Data orders:

- `cycle`: the bench, a five-op cycle over 1000 ops, 150000 rounds.
- `random`: a splitmix64 hash picks each of the 1000 ops. The predictor still learns it.
- `wide`: the same hash over 1000000 ops, 150 rounds.
- `addsub`: 1000000 random `Add` and `Sub` ops, 150 rounds.

Cycles (G):

| form | cycle | random | wide | addsub |
| --- | --- | --- | --- | --- |
| inline (current, selects) | 0.626 | 0.625 | 2.582 | 0.519 |
| switch | 0.625 | 0.625 | 2.596 | 0.520 |
| match through a pointer only | 0.626 | 0.625 | 2.581 | 0.519 |
| pointer parameter only | 0.626 | 0.625 | 2.590 | 0.519 |
| heap (pre-D99) | 0.519 | 0.544 | 3.038 | 1.712 |
| pointer | 0.518 | 0.545 | 3.003 | 1.675 |
| asm barrier | 0.518 | 0.542 | 2.991 | 1.678 |
| asm volatile barrier | 0.549 | 0.568 | 3.335 | 1.928 |

- `pointer` brings `interp` back to 0.15 s and 0.518 G cycles, with 9% fewer instructions than today (1.784 G).
- On `addsub`, `pointer` is 3.2x slower than today (1.675 against 0.519 G cycles). On `wide` it is 16% slower (3.003 against 2.582).
- Other match-heavy code: `sh forms/compiler.sh <repo> <dir> 9` builds `seed/narrac.c` with and without early if-conversion. Both emit identical C for `selfhost`. `narrac emit selfhost` takes 0.789 G cycles (range 0.779 to 0.806) with selects and 0.785 G (range 0.780 to 0.794) without. The -0.5% is inside the spread. The compiler pays nothing measurable for selects, so a form that removes them would not speed it up. This probe does not measure the `pointer` form itself on the compiler.
- The other bench programs have no inline-enum matches. option-inline found their C byte-identical before and after D99.

## Deviations

- The brief asked to compare clang and gcc. No gcc is installed. The selects come from an LLVM AArch64 backend pass, so gcc and x86-64 may differ. Not measured.
- `bench/run.sh` was not run. The load average never reached 2, and the hand forms answer the question.
- `pointer` changes two things: the parameter convention for a 16-byte value (D50 passes by pointer only over 16 bytes) and the for-of item binding. Its effect on the whole compiler was not measured. Only the select part was, through the if-conversion probe.

## Stubs left

- None.

## Gotchas

- The M2 branch predictor learns a repeating op order. That includes the bench's `i % 5` and a Weyl hash such as `i * 0x9E3779B97F4A7C15 >> 40`. Use a mixing hash (splitmix64) to measure the unpredictable case.
- `/usr/bin/time` user time moves in 0.01 s steps. On a 0.15 s program, read `cycles elapsed`.
- A later node may pass borrowed values by pointer (for-of items, small parameters). `interp` then gets back to 0.15 s as a side effect, and `addsub`-like data gets 3x slower. Measure both orders before calling it a win.

## Proposed learnings

- Measure a branch-shape change on predictable and unpredictable data | interp's 20% loss from selects is a 3.2x gain on random Add/Sub data | bench | no
- Read cycles, not user time, for sub-second benchmarks on a loaded machine | user time steps by 0.01 s, and cycles held within 1% across runs | bench | no

## Verification

- `sh selfhost/bootstrap.sh <scratch>`: stage 1 and stage 2 emit identical C (164332 lines), and the seed is current.
- For each data order, every form printed the same output as `inline` (972838 for the bench order).
- `gen.sh` output is byte-identical to the C measured here, and `make.py` makes the same 32 files from it.
- `sh scripts/test.sh` not run: nothing outside `roadmaps/selfhost/nodes/interp-selects/` changed.

## Size

- Estimated S, actual S.
