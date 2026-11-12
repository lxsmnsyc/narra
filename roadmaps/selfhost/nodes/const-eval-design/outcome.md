# Outcome: const-eval-design

## Built

- D105 in `DECISIONS.md`: `@const` functions run in a tree-walking evaluator inside the compiler. It answers the ten questions of the brief. The maintainer approved every recommended answer on 2026-11-12.
- `roadmaps/selfhost/nodes/const-functions/brief.md`: a full brief (L, `story: batch-11`, hard dependency on const-eval-design, codes E1050 to E1057, E1058 and E1059 spare).
- No compiler code. The prototype stays in `/private/tmp/claude-501/const-eval-design/probe/` (`walker.nra`, `crc_native.nra`, `gen_table.py`, `gen_c.py`) and is not committed.

## Proof

Measured on `feat/batch-11` 20f9a487, Apple M2, macOS 27, `-release` unless noted. Details in `log.md`.

- Today: `@const function square(...)` gives `expected a name, found "const"` (E0012); without the attribute the call is E0192.
- CRC-32 table (256 entries x 8 bits) in the prototype walker: 26,370 steps, the same table as the compiled function (`table[1] = 1996959894`, `table[255] = 755167117`), 44 instructions per step, about 80 times slower than native code. Against the chosen budgets: 0.26% of 10 M steps, about 2 KiB of 64 MiB, call depth 1 of 1000.
- 65,536 entries x 16 bits: 13,041,666 steps (over the 10 M budget), 512 KiB of values, 581 M instructions in the walker.
- C output of a table through today's D61 getter (one store per element): 256 entries build in 0.42 s; 4096 take 25 s and 16,384 take 126 s at `-O2`; 65,536 take 153 s in a debug build and did not finish in 4 minutes at `-O2`. A `static const` array with one `memcpy` builds 65,536 entries in 0.17 s. D105 item 7 uses that form for arrays of scalars.
- clang emits `fmadd` for `a * b + c` on arm64 at `-O2` and `-O0`; `#pragma STDC FP_CONTRACT OFF` turns it into `fmul` and `fadd`. D105 item 9 adds the pragma.
- Compiling and running C instead: 0.04 s for an empty C file, 0.42 s for a whole build of the 256-entry program.

## Deviations

- Question 7: the brief recommended turning the value back into a constant expression. D105 writes C from the value with a small writer instead, because a constant's body has no node ids or side-table rows for new AST nodes. It also writes arrays of scalars as `static const` data copied with `memcpy`, after the compile-time measurement above.
- Question 3: the allowed list leaves out the transcendental `Math` functions, whose results depend on the C library, the std functions that take a function, and user trait implementations. The first build covers six std modules (45 rows); the rest of the list is a follow-up node (suggested `const-functions-std`).
- Question 9: adds the FMA pragma to every emitted C file. It changes float results of existing programs in the last bit where clang contracted before.

## Stubs left

- None.

## Gotchas

- The same compile-time cost exists today for any large literal array constant (no `@const` needed): a 4096-element `const TABLE = [...]` takes 25 s to build at `-O2`. Inbox note for the orchestrator: the `static const` plus `memcpy` form could apply to literal scalar arrays in the D61 getter too, as its own small node.
- `testdata/effects_const_functions*` already exists and is unrelated (function values in constants). The new node's samples use the `const_functions` prefix and its error tests `const_eval_`, so the globs do not overlap.
- The evaluator runs on the compiler's C stack (about 20000 Narra frames). The 1000-call depth limit must be checked against the frames one interpreted call takes.

## Proposed learnings

- A large C function of element stores is slow to compile at `-O2` (growth faster than linear) | measured 25 s at 4096 stores, 126 s at 16,384 | selfhost/src emit modules | yes (an emit check on the C of a large table)

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/const-eval-design/out`: fixed point, seed current.
- Prototype runs as listed in Proof. No change under `selfhost/`, so `sh scripts/test.sh` was not needed for this node.

## Size

- Estimated S, actual S. The C compile-time runs added minutes of waiting, not work.
