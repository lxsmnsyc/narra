# checker-emitter-copies log

## Proof on feat/batch-10 (commit ececc85a)

Debug compiler: `cc -O2 -std=c11 -w -o narrac-debug seed/narrac.c -lm` (no `-DNARRA_RELEASE`).

- `NARRA_CHECK_COPIES=1 narrac-debug check testdata/arrays.nra`: exit 103, `array 3, map 1`.
- `NARRA_CHECK_COPIES=1 narrac-debug emit selfhost > base.c`: exit 0, no report. The exit is not clean:
  `exitWithoutRelease` calls `Process.exit`, which skips the copy report.
- `NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1 narrac-debug emit selfhost > base.c`: exit 103,
  `array 956, map 971`. Reproduced.

## Where the copies come from (base, traced)

Traced with a stage 1 compiler built from the base source: a patched `narra_copy_write` logs a
backtrace at each unexplained copy, `atos` names the frames. `emit selfhost`, 1927 copies:

| Count | Site | Pattern |
| --- | --- | --- |
| 904 map + 881 array | `sharedwrites.nra` `mergeInto`, from `walkLoop` and `walkStatement` | `into.pairs = other.pairs` on a dead `into`: the `mutable` parameter takes the caller's other state, which the caller keeps writing. A store into a `mutable` parameter of a value that shares another parameter: no rule sees it (E0976 checks only results). |
| 33 map + 30 array | `pushPair` from `addChained`, `keepArguments`, `bindShares` | follow-on writes to the same shared state |
| 22 + 12 map | `noteWrite`, `forgetRoot` (`Maps.remove` on `pairKeys`) | same |
| 8 array | `checkHit` | same |
| 31 array | `parser.nra` `parseTypeBody` (`>>` split, `parser.tokens[i] = ...`) | `tokens: lexed.value` in `parse`: `lexed` keeps the array until `parse` returns. Gap 1 (write after another name's last use). |
| 5 array | `resolve.nra` `resolveImpl` | |
| 1 array | `modules.nra` `visit` (`orderModules`) | |

`check testdata/arrays.nra`: 2 in `parseTypeBody`, 1 array + 1 map in `mergeInto`.

Note: `emit` ends through `exitWithoutRelease` (`Process.exit`) unless `NARRA_CHECK_LEAKS` is set,
and `Process.exit` skips the copy report, so `NARRA_CHECK_COPIES=1` alone always exits 0 for emit.

## Work

- D104 choice, gap 1: a `const` or `let` is live until its block ends (sharedwrites.nra:
  `scopeEnds`, `heldWrites`, `LiveReason.ScopeEnd`). First cut reported 173 sites in selfhost/src.
  Narrowed to writes that copy a heap object: the shared value (the origin's type, the held
  place's type) must hold an array, map, box, string builder or generator, an enum never counts
  (its payload is not a place), and the path from the shared place to the written place must
  pass a heap object (or the write is a `mutable` argument that may reach inside). 119 sites.
- Fixes: `mergeInto` takes `copy(...)` of the caller's state (1785 of the 1927 copies);
  `parse` moves the lexer's tokens into the parser in `readTokens`; `orderModules` builds its
  tables inside the walk; `query.nra` gives each node its own edge list; layouts and heap enums in
  `check.nra` are built in helpers; `pushBlockContext` in stmts.nra; literal pushed directly in
  declare/check/traits/instances; the rest take the compiler's `copy(...)` fix (about 75 sites,
  applied from `check --json`).
- `exitWithoutRelease` returns (instead of `Process.exit`) when `NARRA_CHECK_COPIES` is set, so
  the report runs on `emit`.
- Gap 2 (shared elements): kept documented, see D104.
- After: debug stage 2 from the branch: `NARRA_CHECK_COPIES=1 narrac-dbg emit selfhost` exit 0,
  also with `NARRA_CHECK_LEAKS=1`; `check testdata/arrays.nra` exit 0.

## Checks

- `sh scripts/test.sh` (from the seed): PASS, 7123 cases in 192 s.
- copies suite: PASS, 1 case, 7 s. Release `emit selfhost`: 2.43-2.46 G instructions before, 2.42 G after.

## Rebase onto feat/batch-10 (423966e4)

- One conflict (DECISIONS.md, D103 and D104 kept). No new E0978 site in code that landed since the base.
- E0978 dedupe by line. Stage 2 debug (with the per-holder copy check runtime): `check testdata/arrays.nra` exit 0, `emit selfhost` exit 0 with and without `NARRA_CHECK_LEAKS`.
- `sh scripts/test.sh` after the rebase: `test.sh: PASS, 7147 cases in 136s`.

## Review fixes

- Enums count by their payloads; `is`/`match` bindings of a non-place subject and `for` bindings over one hold until their block ends; a `copy(...)`-initialized local ends at its last use; E0978 dedupe by the written place range. 52 more sites fixed.
- Debug stage 2: `check testdata/arrays.nra` exit 0, `emit selfhost` exit 0 with and without `NARRA_CHECK_LEAKS`.
- `sh scripts/test.sh`: `test.sh: PASS, 7162 cases in 233s`.

## Second review

- Removed the `copy(...)`-initializer exception; samples rewritten; probes ci8/ci12 added as error tests; `findPlaceEnd` handles an integer index.
- Debug stage 2: `check testdata/arrays.nra` exit 0, `emit selfhost` exit 0 with and without `NARRA_CHECK_LEAKS`.
- `test.sh: PASS, 7172 cases in 175s`.
