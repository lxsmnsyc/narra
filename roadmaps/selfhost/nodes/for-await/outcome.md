# Outcome: for-await

## Built

- `for await (const x of stream) { ... }` over an `AsyncGenerator<T>` (D110 item 2) in an async function, an async arrow, an `async { ... }` block and an async generator. Destructuring heads, labels, `break`, `continue`, `return` and labeled exits work as in `for ... of`.
- AST: `Stmt.ForStmt` has a last field `awaitPos: Option<Pos>` (the `await` keyword; None for a plain `for ... of`). Every `Stmt.ForStmt(` pattern has 8 fields now.
- Parser (`parser.nra`): `parseForAwait` reads `await` after `for` and records a trivia node `k<pos>` for it. The head takes `const` (and patterns after it) or `mutable`; `for await (let ...)` and the C-style form are E0030 ("write `for await (const item of stream)`").
- Checker, `selfhost/src/forawait.nra`: `checkForAwaitPlace` (E1122, names the place as `reportAwaitPlace` does), `findStreamElement` (E1123 with the Generator and Array<Async<T>> hints; makes the `Option<T>` instance the emitter reads), `reportPlainStreamLoop` (E1124), `reportMutableStreamLoop` (E1105, then the loop is checked as `const`). Dispatch rows in `checkFor` (`stmts.nra`). `seedLoopFacts` needed no change: `declaration.generator` is set for `async function*` too.
- Effects (`effects.nra`): the loop charges the stream type's effects (existing `walkGeneratorRun`) plus `alloc` at the `await`.
- Emitter, `selfhost/src/emitforawait.nra`: `emitForAwait`, dispatched first in `emitFor` (`emitstmts.nra`). The loop takes its own reference to the stream (`consume`, no clone) in a frame temp of the statement; each turn makes a step with `narra_stream_next`, awaits it as `emitAwait` does, releases it and breaks when `narra_step_filled` is false, else binds x to a retained copy of the `Some` payload and releases the step before the body. `findStreamOption` in `emitasyncgenerators.nra` is now exported.
- fmt (`fmt.nra` `printFor`) prints `for await (`, with comments around the keyword through key `k`. astdump prints ` await` on the `for` node.
- `Net.chunks(connection, maxBytes): AsyncGenerator<Result<bytes, AsyncError>> uses io, alloc` in `std/src/async__net.nra` (D110 item 7): `Ok(chunk)` per non-empty read, ends at an empty read, one `Err` then end on an error. `stdtext.nra` regenerated.
- Codes: E1122, E1123 (3 rows), E1124, one more E1105 row, one more E0030 row, with explanations (`diagnostics.nra`).
- SPEC: no "not built yet" left for `for await` / `Net.chunks` (section 2 bullet, the TypeScript table row, section 8 "Async generators"); the sample is a checked `ts` block. `compiler/README.md`: feature line added, the "Not supported yet" line removed.
- Tests: samples `testdata/for_await_{basic,break,destructure,cancel,chunks}`; errors `testdata/errors/for_await_{place,source,plain_loop,mutable,let}`; goldens `testdata/fmt/stmt_for_await`, `testdata/astdump/for_await`, `testdata/query/for_await/`.

## Proof

- base: feat/batch-13 21efc615, compiler built from the seed then `build -release selfhost`. `base-narrac run testdata/for_await_basic.nra` exits 1: `for_await_basic.nra:18:7: expected "(", found "await"`.
- branch: all five samples give their checked output in debug (AddressSanitizer, leak check, copy check) and release.

## Deviations

- The field is `awaitPos: Option<Pos>`, not `awaits: bool`: fmt needs the keyword position for its trivia key (the brief said "for example").
- Goldens outside the owned paths: `testdata/fmt/stmt_for_await.{nra,fmt}`, `testdata/astdump/for_await.{nra,ast}`, `testdata/query/for_await/` (the acceptance asks for fmt, astdump and query coverage; these folders are where they live).
- `for await (mutable x of s)` gives a message of its own under E1105 and is then checked as `const`, so no E0954 follows; the brief said "goes through the existing path".
- `for await (let x of s)` is E0030 (parse error). D110 does not name `let`; SPEC now says the head takes `const`.
- `Net.chunks` with `maxBytes <= 0` ends at once (read gives empty bytes); SPEC says so.

## Stubs left

- None.

## Gotchas

- A comment between `await` and `(` (`for await /* c */ (const x of s)`) moves before the iterable, as a comment after `(` already does for every `for ... of`: the loop head has no trivia node. A comment between `for` and `await` stays.
- `for_await_chunks` needs one `.out` (no platform split): the busy message is the std's own text, not `strerror`.
- The emitter assumes an async frame (`f->h`); E1122 keeps `for await` out of other bodies, so a new body kind that allows it must be a task frame.

## Proposed learnings

- A new keyword field on an AST variant should be an `Option<Pos>`, not a bool | fmt keys the keyword's comments by its position | selfhost/src | no

## Verification

- Rebased on feat/batch-13 11d255b0 (exit 0).
- `NARRAC=<branch -release narrac> sh scripts/test.sh`: PASS, 8862 cases in 356s.
- `narrac emit selfhost` from the base compiler (11d255b0, seed-built then -release) and the branch compiler, on the branch source: byte-identical (229302 lines).
- `sh scripts/test.sh --suite generated --update` after the `std/src/async__net.nra` edit.

## Size

- Estimated M, actual M: about 400 lines in `selfhost/` and `std/` (generated text excluded), plus tests.
