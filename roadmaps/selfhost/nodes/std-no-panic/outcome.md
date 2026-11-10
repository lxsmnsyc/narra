# Outcome: std-no-panic

## Built

- `Str.slice`, `Str.split`, `Str.repeat`, `Bytes.at`, `Bytes.slice`, `Float.truncate`, `Arrays.filled`, `Ranges.stepped`, `Math.clamp`, `Math.powInt` and `Random.nextInt` return `Option` and have no panic path. The table with each result type and its `None` case is in SPEC.md section 5.
- The runtime functions (`selfhost/runtime/runtime.h`, `std_string.c`, `std_bytes.c`) return a bool and store the result through a pointer; the emitters share `emitOptionOut` in `selfhost/src/stdemithelp.nra`. `std/math` and `std/random` are source modules (`std/src`).
- `Maps.update(mutable map, key, initial, change)` (`stdmap.nra`, `emitmaps.nra`). A compound assignment on a Map index is a compile error (E0226, `stmts.nra`); the emitter path is gone.
- Every caller updated: `std/src/json.nra`, the test runner (`selfhost/tests/src/textops.nra`), samples in `testdata/`, `bench/` (two programs), the guide (`docs/guide/02`, `05`, `06`, `08`, `10`), SPEC.md, DECISIONS.md (D92 "Built" line), compiler/README.md.
- Samples: `testdata/std_no_panic_strings.nra` and `testdata/std_no_panic_numbers.nra` cover the failure case of every function and `Maps.update`; `testdata/errors/map_compound_assign.nra` covers the compile error.

## Proof

Not a fix node.

## Deviations

- Edits outside the brief's paths: `bench/sieve/sieve.nra` and `bench/matmul/matmul.nra` (they call `Arrays.filled`), `testdata/**` goldens, `docs/guide/**` and `selfhost/tests/src/panics.nra` (removed four sample entries). All of them are listed as shared or follow from the brief's "update every caller".
- Four panic samples were deleted (`strings2_inside`, `strings2_range`, `strings2_split_empty`, `map_missing_key`) because the behavior they tested is gone.
- `F32.truncate` also returns `Option<int>`. `Str.sliceFrom`, `Str.charAt` and `Str.nextChar` still panic; SPEC and D92 list them as remaining.
- The brief says to keep compiler call sites on the old form. That cannot bootstrap: stage 2 is compiled by a compiler with the new tables. The call sites go through four helpers instead (see Stubs left).

## Stubs left

- `selfhost/src/textops.nra` holds `substring`, `splitString`, `repeatString` and `filledArray`. Their bodies avoid the changed builtins so the compiler builds with the seed's tables and with the new ones. 110 call sites in 41 files of `selfhost/src` use them (they replaced `Str.slice` 58, `Str.split` 28, `Str.repeat` 9, `Arrays.filled` 15). After the seed update, switch the four bodies to one call each and regenerate the seed again:
  - `substring`: `if (Str.slice(text, start, end) is Some(piece)) { return piece }` then `panic(...)`.
  - `splitString`: `Str.split(text, separator)`, same shape.
  - `repeatString`: `Str.repeat(text, count)`.
  - `filledArray`: `Arrays.filled(value, count)`.
  The helpers are temporary and cost compile time: `narrac emit selfhost` rose from a median 1.753G to 2.321G instructions (+32%, five pairs, measured in review), almost all in the per-character loops of `substring` and `splitString`. A tail slice and an empty slice now take a fast path (not remeasured). The second seed update turns the helpers into direct calls and removes the cost.
- The bootstrap takes the stage 3 path until the seed update (stage 1 and stage 2 C differ because the embedded runtime and std text changed).

## Gotchas

- `Some`, `Ok` and `Err` charge `alloc` only for an enum stored on the heap; `Option<int>` and the like need no `uses alloc`.
- `[]` has no type inside `Options.withDefault(x, [])`; pass a typed fallback.
- `Ranges.stepped` returns `Option<Range>`, so a `for` over a direct call needs an `is Some` first.
- The lsp samples embed `testdata/lsp/*/src/text.nra` in `session.txt`; a change to those sources must be repeated there.
- A test run killed with exit 137 or `Killed: 9` was memory pressure from the other workers, not a compiler fault; rerun.

## Proposed learnings

- A builtin whose result type changes needs a compatibility helper for the compiler's own call sites, because stage 2 is built with the new tables | the seed rule in `seed/README.md` and the brief assume the old call sites still compile | selfhost | no

## Verification

- `sh selfhost/bootstrap.sh <out>`: ok, stage 2 and stage 3 emit identical C (stage 3 path, seed older than source).
- `JOBS=3 sh scripts/test.sh`: PASS (6016 cases plus the bootstrap suite, 0 failed).
- Targeted suites on the built compiler, all passing: programs, panics, errors, dumps, leaks, packages, runner, fmt, fmt-stable, fmt-doc, trivia, query, lsp, sources, node-ids, edge, habits, manifest, std-tests, runner-options, emit, guide, spec, generated, codes.

## Size

- Estimated M, actual L. The bootstrap rule forced the helper layer, and about 40 samples, the runner and the guide needed edits.
