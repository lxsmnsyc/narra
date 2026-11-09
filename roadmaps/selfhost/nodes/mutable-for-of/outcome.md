# Outcome: mutable-for-of

## Built

D85 in `selfhost/src`: `for (mutable item of items)`, the same over `Arrays.values(items)`, and `for (const #[index, mutable item] of Arrays.entries(items))`.

- **AST.** `Stmt.ForStmt` has a 7th field, `reference: Option<LoopReference>` (`ast.nra`). `LoopReference { indexId, indexName }`; `indexName` is "" without the index. Every `ForStmt` match site got the field. `astdump` prints ` mutable` and ` index="..."` on the `for` line; `fmt` prints the head back; `idcheck`, `typesdump` and `querysites` cover the index local.
- **Parser.** `parseReferenceFor` and `isMarkedEntries` (`parser.nra`). The marker head is not desugared into a `VarStmt`, so nothing copies the element.
- **Checker.** `checkReferenceFor` (`stmts.nra`). The binding is `BindingKind.Reference` (`check.nra`), declared with `isMutable` true and `param` false. That gives the `let` capture rules (by reference, stack closures only). `canChangeContents` accepts it. The source is checked as a `mutable` argument would be: E0290 for a constant, E0954 when it is not a place, `checkContentsChange` (E0315 plain parameter, E0950 copying binding). The element must pass `isContentsType` (E0955). Not allowed in a `function*` body (E0957).
- **Borrow (rule 4).** `Checker.loopBorrows: Array<LoopBorrow>` and `borrowChecked`. `checkLoopBorrows` (`exprs.nra`, called at the top of `checkExpression`) compares each whole place with the borrowed source. A place is its root name plus a field path, and `[]` matches any index. One place overlaps another when one path is a prefix of the other. Inside a loop over `game.players`, `game`, `game.players` and `game.players[0].hp` are E0956 and `game.round` is fine. This covers reads, writes, `mutable` arguments and closure captures, because all of them go through `checkExpression`.
- **Rule 3.** Whole assignment is E0958. Return or store is E0915 (new format). Capture by an escaping closure is E0613 (new format). `const copy = item` copies.
- **Emitter.** `emitReferenceFor` (`emitstmts.nra`) emits `A** p = <emitLvaluePointer(source)>; *p = A_unique(*p); A* a = *p;`, then `T* v_item = &a->items[i]`. The array gets no frame entry, and the element is never retained or released. `isReference` (`emitstate.nra`) is true for the kind, so reads, writes, `mutable` arguments and closure environments use the pointer. `containerName` in `emitstate.nra` is now exported.
- **Tests.** `testdata/mutable_for.nra` covers field writes, a nested array, a `mutable` argument, `Arrays.values`, entries with `continue` and `break`, nested reference loops, a source under a `mutable` parameter field, a stack closure, a copy of the element, `return` from inside the loop, and array elements. `testdata/mutable_for_shared.nra` covers a source shared with a local, a field and an element of a shared array: each is copied once and the other holder is unchanged. 16 error tests `testdata/errors/mutable_for_*`: call, constant, map, param, int, use, use_path, use_root, use_closure, generator, whole, escape, heap_closure, entries_index, entries_missing, hint (`for (const ...)` over `game.players` suggests the marker form). Goldens: `testdata/fmt/stmt_for_mutable`, `testdata/astdump/for_mutable`, `testdata/typesdump/bodies/mutable_for`.
- **Docs.** `DECISIONS.md` D85: dropped "Not built yet". `compiler/README.md`: the "not built" line now lists what differs: async is not built, so that rule has nothing to check; the entries head takes two names only; the borrow check is conservative about indexes. `SPEC.md` already matched; no change. Guide chapter 9 has a new section, "Heal the whole party": the E0950 error sample with its hint, the marker and entries forms with output, and the borrow rule. Chapter 11 has a sentence linking to it.

## New codes

- E0954: source is not an array place (4 formats: `#[i, mutable x]` without `Arrays.entries`, `mutable x` over `Arrays.entries`, not an Array, not a place).
- E0955: element type has no contents.
- E0956: source, or a path that overlaps it, used in the body.
- E0957: the form in a `function*` body.
- E0958: loop element assigned whole.
- New formats of existing codes: E0613 (escaping closure captures the element), E0915 (element returned or stored).

## Gate decisions

1. **AST shape: a field on `ForStmt`, not a new statement kind.** A new kind would be silently skipped by every walker with a `_ =>` arm. The field makes each site visible. Cost: 20 one-token pattern edits.
2. **Binding: `BindingKind.Reference` with `isMutable` true, not `param` true.** With `param` true, messages would say "mutable parameter" and an unused element would be reported as an unused parameter (E0922). `isMutable` gives the pointer-capture and stack-closure rules unchanged.
3. **Whole assignment gets its own code, E0958, not E0949.** The E0949 text says "mutable parameter ... return the new value instead", which is wrong for a loop. All five reserved codes are used.
4. **Borrow check by comparing whole places in `checkExpression`, with any index matching any index.** The alternative was a separate walk of the body after checking. That walk would duplicate the expression walker and miss captures resolved by the checker. `teams[0]` vs `teams[1]` overlapping is conservative, as D85 rule 4 asks (the checker cannot compare indexes).
5. **A plain parameter as the source reports E0315 (from `checkContentsChange`), not E0954.** The fix for it is the same as for a `mutable` argument: declare the parameter `mutable`.
6. **Async: no check.** Async functions do not parse or check yet (README "Not supported yet"), so rule 6 for `async` has nothing to attach to. Whoever builds async adds the check next to E0957.

## What this means for later nodes

- The seed does not know the form yet. After the orchestrator regenerates the seed, `check.nra:621` (`for (mutable error of checker.errors)`) and `resolve.nra:532` (`#[index, mutable field]`) can use it.
- Any new `Stmt.ForStmt` match needs 7 fields.
- An async node must reject the form in async bodies (E0957 is the model; it needs another format or code).
- Pattern markers (D85 deferred items) can reuse `LoopBorrow` and `checkLoopBorrows` for the borrow of the subject.

## Deviations

- The brief's list of error tests puts "the form in a generator or async body" together; only the generator case exists (see gate decision 6).
- No full `sh scripts/test.sh` run, as the orchestrator asked mid-task. Suites run are listed below.

## Stubs left

- None.

## Gotchas

- `checker.loopBorrows` is restored by assigning the saved array back, which makes a second owner for one push. This is fine here, because borrows nest only as deep as the loops.
- An unresolved source (an unknown module) still declares the element as `void`, so its body can show follow-on errors, as in the plain `for` path.

## Verification

- `sh selfhost/bootstrap.sh out`: stage 2 and stage 3 emit identical C.
- Suites (`NARRAC=out/narrac JOBS=2 sh scripts/test.sh --suite X`), all PASS: errors, codes, sources, fmt, fmt-stable, dumps, programs (ASan and the leak check), query, node-ids, trivia, emit, guide, std-tests, bootstrap.

## Size

- Estimated M, actual M (about 620 lines added in `selfhost/src` and tests).
