# Outcome: style-sweep

## Built

- `selfhost/src`, `selfhost/tests/src` and `std/src` follow the step-comment style: comments on a function moved into short comments on the steps they explain, or were dropped when the name and body already said it. About 1300 function comment blocks went; about 500 facts moved into bodies.
- D-number citations are gone from code comments. The rule each comment stated is kept.
- `<noun>Of` functions renamed, with every caller:

| Old name | New name | File |
| --- | --- | --- |
| `queryTypeOf` | `queryType` | `selfhost/src/query.nra` |
| `checkValueCallOf` | `checkCallThroughValue` | `selfhost/src/closures.nra` (callers in calls, exprs) |
| `keyOf` | `buildKey` | `selfhost/src/types.nra` |
| `keysOf` | `joinKeys` | `selfhost/src/types.nra` |
| `keyOfFunction` | `buildFunctionKey` | `selfhost/src/types.nra` |
| `enumOf` | `readEnumIndex` | `selfhost/src/match.nra` |
| `prefixOf` | `readPrefix` | `selfhost/src/stdsized.nra` |
| `keyOf` | `readFunctionName` | `selfhost/src/stdsized.nra` |
| `rowOfType` | `findTypeRow` | `selfhost/src/stdsized.nra` |
| `collectUsesOf` | `collectListUses` | `selfhost/src/stmts.nra` |
| `typeOf` | `readNodeType` | `selfhost/src/emitstate.nra` (exported, 20 importers) |
| `refOf` | `readNodeRef` | `selfhost/src/emitstate.nra` (exported) |
| `localOf` | `readLocal` | `selfhost/src/emitstate.nra` (exported) |
| `charOfByte` | `findByteChar` | `selfhost/src/emitstate.nra` |
| `endOfName` | `findNameEnd` | `selfhost/src/querysites.nra` |
| `stepOf` | `readStep` | `std/src/coroutine.nra` (private) |
| `lineStartOf` | `findLineStart` | `selfhost/tests/src/codes.nra` |

- The E0507 row of the code table in `selfhost/src/diagnostics.nra` now lists both messages `traits.nra` prints: the short one ("... block", for a user trait) first, the one naming `derives` (built-in derivable traits) last. `narrac explain` prints the last row of a code, so its output for E0507 is unchanged.

## Counts

| Check | Before | After |
| --- | --- | --- |
| Go-form grep `^\s*// [a-z][A-Za-z]* (is\|returns\|reports\|checks) ` | 574 | 0 |
| D-number grep `D[0-9]{2}` | 88 lines | 10 lines, none a citation in a comment (see kept) |
| Function names ending in `Of` | 15 (plus 4 with `Of` inside) | 2, kept on purpose |
| Comment blocks directly above a function (script count) | 1323 blocks, 2546 lines | 25 blocks, 33 lines, kept on purpose |

## Kept on purpose

- `emitIndexOf` (`selfhost/src/stdstring.nra`): emits `Str.indexOf`; the name is the std function's, not a `<noun>Of`.
- `lastIndexOf` (`selfhost/tests/src/guide.nra`): mirrors the JS method of the same name.
- D-number grep hits left: hex literals (`0xD800` in `stdchar.nra`, `0x94D049BB...` in `std/src/random.nra`) and text inside generated string constants (`runtime*.nra` copies of `selfhost/runtime/`, including the `(D82)` comment in `selfhost/runtime/runtime.h`, which is outside this node's paths and is copied into every program's C; `stdtext.nra` copy of random.nra's hex).
- Function comments left (25 blocks):
  - `std/src/json.nra` (14), `std/src/shape.nra` (3), `std/src/coroutine.nra` (1): std/src positions appear in the emitted C of programs (bounds-check and panic messages such as `"std/src/json.nra:L:C"`), so removing a line there changes every sample that imports the module. These were edited in place only (D-numbers and Go-form wording), keeping each line where it was.
  - `selfhost/src/check.nra` (7): comments on the method signatures of the `Walker` trait (which module implements each). A trait method has no body to move them into.

## Deviations

- std/src function comments are kept (see above), to keep the emitted C of samples unchanged.
- `std/src/coroutine.nra`'s `stepOf` rename changes the emitted C of the coroutine samples by that function's C name only (`f_..._stepOf_<hash>` to `f_..._readStep_<hash>`). No other sample C changed.
- `selfhost/tests/src/generated.nra` lost the comment it wrote above `stdSourceFor` in the generated `stdtext.nra`; `stdtext.nra` is regenerated.

## Stubs left

- None.

## Gotchas

- `narrac fmt` with no paths formats every `.nra` file under the current directory, including roadmap evidence files and `bench/`. Pass paths.
- Do not run `narrac fmt` on `std/src`: the sources suite does not cover it, and moving lines there changes the positions in every program's C.
- The seed is not regenerated (the batch does that once). `bootstrap.sh` takes the stage 3 path and notes the seed is older.

## Proposed learnings

- Put a fact about a function on the step it explains, not above the function | function comments drift into restating the name, and the sweep removed about 1300 of them | selfhost/src, selfhost/tests/src | yes (a sources check could reject a comment block directly above `function`, with trait signatures exempt)
- Do not cite D-numbers in code comments; state the rule | the citation sends the reader to DECISIONS.md for what one sentence can say | selfhost/, std/ | yes (grep)

## Verification

- Before every group commit: the code-only diff filter (`git diff | grep '^[-+]' | grep -v '^[-+]\s*//'`) showed nothing but renames and formatting; `narrac fmt --check` on the group's files.
- The compiler's own C: the stage 2 C from the base tree and from the branch (with the E0507 row taken out) are identical after mapping the renames and ignoring source positions and literal lengths, except the embedded std text (`STD_COROUTINE`, `STD_JSON`, `STD_SHAPE`), whose only changes are the std comments and the `readStep` rename.
- Sample C: `narrac emit` of all 111 top-level samples in `testdata/` with the base compiler and the branch compiler: byte-identical, except the `readStep` C name in the two coroutine samples.
- `sh selfhost/bootstrap.sh`: fixed point holds through stage 3 (stage 2 and stage 3 C identical, 179455 lines); the seed is noted as older than the source, as expected before the batch seed update.
- `sh scripts/test.sh`: PASS (8 shards of 700 cases, 0 failed, 311 s), with the compiler built from the seed.

## Size

- Estimated L, actual L.
