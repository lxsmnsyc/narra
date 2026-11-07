# Log: mutable-reference

## 2026-11-07

- Branch `spike/mutable-reference` from origin/main (66ea319).
- Read GOAL, roadmap CLAUDE.md, the brief, `mutable-read-conflict/outcome.md`, SPEC sections 2 and 6, D27, D50, D77-D79.
- SPEC.md code blocks are not compiled by any test (`TestGuide` reads `docs/guide/*.md` only; SPEC uses `ts` fences). No marking needed for unbuilt samples. The guide samples are compiled, so the guide is left as it is for mutable-bindings.
- Wrote `count.py` (this folder). It blanks comments and strings, finds `mutable name: Type` parameters, classifies the type against the struct, enum, trait and distinct declarations of the tree (selfhost and std share one table; testdata per directory; guide per chapter), and looks for whole assignments (`name =`, compound forms, `++`, `--`) in the function body.
- A first version also matched `(mutable x)` arrow parameters; the regex hit call arguments followed by `=>` later on the line, so it was dropped. E0028 already rejects mutable arrow parameters.
- Types that change in place, found while counting: besides structs, `Array` and `Map`, the corpus passes `StringBuilder` (53 selfhost parameters), `Generator` and `Coroutine` (`std/src/coroutine.nra`) as `mutable`. SPEC also has `Box<T>` (`b.value = v`), `FixedArray` (element writes) and `Net.read(mutable conn)` on std connection handles. Rule 1's list must include them or the std breaks; D81 names them "contents types".
- SPEC section 8 `Shared.update(hits, (mutable count) => { count = count + 1 })` is an `int` and a whole assignment, and E0028 already rejects it. Rewritten in D81 to a callback that returns the new value.
- Two trait functions in testdata take `mutable shape: Self`; their implementations are for structs.
- No bare generic `mutable x: T` anywhere in the four trees.
- Guide: chapters 2, 3, 9 and 11 teach the old rule; 8 only uses `mutable` on arrays (still valid).
- Go sites the build node changes: `check_stmt.go` AssignStmt (`!root.Mutable`, E0227), `check_call.go` `checkMutableArg` (E0315), `check_const.go` (E0290, top-level, stays), `emit_closure.go` `captureValue`/`envStruct` (`Local.Mutable` doubles as "captured by pointer"). Selfhost: `exprs.nra:1714`, `exprs.nra:1769`, `stmts.nra:351`.

## 2026-11-07, review round

- Maintainer: approved the widened type list and read-only plain parameters; read-only pattern bindings need more evaluation.
- D81 now separates rules decided by the maintainer (1 to 5) from rules proposed by the spike (6 to 10), and adds an open question with options (a) and (b) for bindings that copy part of another value. My first reason for read-only pattern bindings was wrong: they are retained copies (`emit_match.go:11`), not D50 pointers.
- Rule 9 (closures) now covers nested closures: every closure between the change and the declaration captures the `const` by reference. A stack environment's by-value capture is an unretained bitwise copy, so a push through an inner closure could reallocate it.
- Dropped distinct types from the allowed list and E0915 from the exclusivity rule. `Shared.update` is back to its old sample with an open note; `Shared` is not built.
- SPEC sections 2 and 14 name value copies as the one TypeScript difference with no diagnostic.
- `count.py` prints the never-reassigned `let` scan.
