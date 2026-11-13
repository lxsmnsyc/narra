# Log: for-await

## Proof (base)

- Base: feat/batch-13 at 21efc615 (compiler built from the seed, then `build -release selfhost` of that tree).
- `base-narrac run testdata/for_await_basic.nra` exits 1: `for_await_basic.nra:18:7: expected "(", found "await"`.

## Branch

- AST: `ForStmt` gained `awaitPos: Option<Pos>` (not `awaits: bool`): fmt needs the keyword's position for its trivia key `k`. 32 patterns got one more `_` by script.
- Parser: `parseForAwait` after `for`; `for await` takes `const` or `mutable` heads; `let` and the C-style form are E0030.
- `for await (mutable x of s)` reports E1105 alone (its own message), then checks as a `const` loop, so no E0954 follows.
- A comment between `await` and `(` moves before the iterable, as any comment after `(` does today (no trivia node for the loop head). Not in the fmt golden.
- All 5 samples clean in debug (ASan, NARRA_CHECK_LEAKS, NARRA_CHECK_COPIES) and release.
- Rebased on feat/batch-13 11d255b0 (exit 0). `NARRAC=<branch -release> sh scripts/test.sh`: PASS, 8862 cases.
- `emit selfhost`: base (11d255b0, seed-built then -release) and branch compilers on the branch source give identical C (229302 lines).
