# Outcome: std-design

## Built

- D118 in `DECISIONS.md`, "Proposed, waiting for the maintainer": the API of `std/iterator`, `std/arena`, `std/unicode`, `std/args`, `std/toml`, `std/regex` and `std/terminal`, which need runtime C (only terminal), the code split and seven open points with recommendations.
- SPEC section 3: a "Std modules for tools (D118). Not built yet." bullet per module. Section 4 "Strings" and section 6 "Cycles" point at D118; `Arena.get` became `Arenas.get` (the plural alias).
- `selfhost/src/diagnostics.nra` header: the range `E1300-E1399`.
- Full briefs: std-iterator (M), std-arena (S), std-unicode (L), std-args (M), std-toml (M), std-regex (L), std-terminal (M).
- Split: std-unicode into std-unicode and the new node std-unicode-normalize (M); std-regex into std-regex and the new node std-regex-const (M). The original briefs are the first child of each split.

## Deviations

- Codes: E1300-E1309 iterator, E1310-E1319 arena, E1320-E1324 unicode, E1325-E1329 unicode-normalize, E1330-E1339 args, E1340-E1349 toml (E1340 kept for the manifest move), E1350-E1359 regex, E1360-E1369 regex-const (E1360 used), E1370-E1379 terminal, E1380-E1399 free. Only E1360 is expected to be used in this batch.
- No node created for moving narrac onto `std/args` or `manifest.nra` onto `std/toml`: both need the batch's seed update first (std sources are embedded in narrac). Proposed nodes for the orchestrator: narrac-args (M, hard on std-args and the seed update; `-release` becomes `--release` in `scripts/test.sh` and `bench/run.sh`) and manifest-toml (S, hard on std-toml and the seed update; E1340, retire E0405, E0406, E0408).
- The guide has no standard library chapter. The briefs say the first build to land creates `docs/guide/12-standard-library.md`; the orchestrator may pick another number.

## Stubs left

- None.

## Gotchas

- Literal `Array<int>` constants of a few thousand items take about a minute in clang (D61 getter). Large tables must be string constants.
- SPEC section 5 names `Arrays.get`, which the batch-14 seed does not have.
- `import * as Regex` would clash with the type `Regex`; the alias is `Regexes`.
- D114 (tagged templates) is open in parallel. D118 ships no std tags; `std/terminal` style functions and `Regexes.checked` are what tags would call later.

## Proposed learnings

- Std tables go in string constants, not array literals | clang time on D61 getters grows past a minute at 6,000 items | std/src | yes (a sources check on array literal length)

## Verification

- Measurements and prototypes: `log.md`.
- `NARRAC=<seed-built narrac> sh scripts/test.sh --suite spec`: PASS, 62 cases. `--suite sources`: PASS, 1 case.

## Size

- Estimated S, actual S.
