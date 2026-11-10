# Outcome: regions-design

## Built

- Decision D100 in `DECISIONS.md`, "Not built yet". It answers the five questions of the brief and recommends parking node regions (priority could) and building a size-class allocator first.
- Prototype under `roadmaps/selfhost/nodes/regions-design/prototype/`, no shipped code:
  - `variant.py` rewrites C emitted by narrac into measurement variants: `count`, `freelist`, `bump`, `nofree`, `nocount`, `sign`.
  - `measure.sh [runs]` builds narrac from the seed, builds every variant, checks each emits the same C as the base compiler, and prints instructions retired, user seconds and peak memory for `emit selfhost` and `fmt --check selfhost/src`, then the sign check on the six `bench/` programs. macOS only (`/usr/bin/time -l`).

## Result

- result: answered. Question 5: no phase justifies a region now.
- numbers (Apple M2, 8 GB, macOS 27, median of 7 runs, load average 4.7 at start and 3.4 at end; a repeat at load 19 gave the same instruction counts within 0.2%):

| variant | `emit selfhost` instructions | user s | peak | `fmt --check selfhost/src` instructions |
| --- | --- | --- | --- | --- |
| base | 9.27 G | 0.57 | 135 MB | 14.21 G |
| freelist | 5.45 G | 0.42 | 153 MB | 9.22 G |
| bump (counts kept) | 5.04 G | 0.42 | 807 to 911 MB | 9.47 G |
| nofree (counts kept, no drop walk) | 4.63 G | 0.36 | 898 MB | 9.15 G |
| nocount (no counts, no frees) | 3.75 G | 0.29 | 909 MB | 8.77 G |
| sign | 9.64 G | 0.58 | 135 MB | 14.26 G |

- counts per `emit selfhost`: 19.6 M allocations, 580 MB allocated, 510 K reallocs, 60.2 M retains, 70.3 M releases, 15232 copies on write.
- sign check on `bench/` (instructions, median of 5, load 2.9): `interp` +4.9%, `strings` +0.3%, `sieve` within 0.4% either way, `fib`, `matmul`, `particles` 0%. User time unchanged at 0.01 s resolution for all six.
- repeat run of `measure.sh 5` at load 1.9 to 3.3: every instruction count within 0.5% of the table.
- reproduce: commit `81638c3` (feat/batch-4), `sh roadmaps/selfhost/nodes/regions-design/prototype/measure.sh 7`.

## Proposed children

Recommendation: park node regions. If the maintainer wants it anyway, these are the children, in order. Diagnostic codes are not reserved yet.

1. **id** `fast-alloc` (not a regions child; proposed for batch-4, perf).
   - title: Size-class free lists in `narra_alloc`
   - goal: replace `malloc`/`free` for blocks up to 512 bytes with per-thread size-class free lists carved from 1 MiB chunks, as `variant.py freelist` does; `narra_realloc` moves small blocks.
   - size: S. dependencies: bench-compiler (soft, for the baseline).
   - acceptance: `emit selfhost` at most 6.0 G instructions (prototype 5.45 G against 9.27 G); peak memory within 20% of base; `sh scripts/test.sh` clean under ASan and the leak check (ASan builds keep plain `malloc`, so ASan still sees every block); seed regenerated.
2. **id** `region-runtime` (parked).
   - title: Region blocks, the routing flag and the sign check
   - goal: per-thread region stack, bump blocks, negative mark on region objects, sign check on array, map, box and enum retain and release, growth by copy, a store list for heap values held by region containers, generator resume outside the region.
   - size: M. dependencies: shared-write-check (hard: in-place writes on region values need D94), fast-alloc (soft).
   - acceptance: a C-level test allocates, grows and frees a region with the leak check at zero; `bench/` instructions within 5% of base.
3. **id** `region-copy-out` (parked).
   - title: Generated copy functions for values leaving a region
   - goal: one copy function per boxed type and closure environment, forwarding pointers for shared subtrees, heap objects retained not copied.
   - size: M. dependencies: region-runtime.
   - acceptance: samples copy strings, arrays of enums, maps and closures out of a region; a DAG with shared subtrees copies each once; ASan and leak check clean.
4. **id** `region-check` (parked).
   - title: `Region.run` in std/region and its two errors
   - goal: `Region.run(() => ...)` with no region parameter; errors for a write or `mutable` pass of a captured `let` or `mutable` parameter, and for a result type holding an identity type, `Generator`, `Coroutine` or `Async`; run-time copy on writes into identity objects outside the region; SPEC section 6 and D48 updated.
   - size: M. dependencies: region-copy-out, two reserved diagnostic codes.
   - acceptance: `testdata/regions*.nra` samples, `testdata/errors/regions_*.nra` for both errors, `narrac fmt` on one file in a region gives the same output.

## Deviations

- None. No compiler, runtime or test file changed. `DECISIONS.md` gains D100 only.

## Stubs left

- None.

## Gotchas

- `bench/BASELINE.md` user times disagree with the ones here: user time on this M2 swings 3x with load and core type. Compare instructions retired.
- The `nocount` variant emits the same C although it skips 15232 copies on write. That shows the compiler does not need those copies on its own source. It does not show it never needs them; only D94 can prove that.
- The free-list prototype keeps an 8-byte size header and never returns chunks to the system. It is a measurement, not the design for fast-alloc.

## Proposed learnings

- Measure allocator cost before reference counts in narrac | `malloc`/`free` is 41% of the compiler's instructions, counts at most 9% | selfhost/runtime | no

## Verification

- `sh roadmaps/selfhost/nodes/regions-design/prototype/measure.sh 7`: every variant emits the same C as the base compiler.
- No fast check: no compiler code changed (spike).

## Size

- Estimated M, actual M.
