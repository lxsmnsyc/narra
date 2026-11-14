# Outcome: remove-lazy

## Built

- Lazy values are gone (D119): `const x = lazy expr`, `std/lazy` and `Lazies.force`. `lazy` was a contextual word, so `const x = lazy 1` is now an ordinary parse error with no special message.
- Deleted `selfhost/src/lazyvalues.nra` and `selfhost/src/stdlazy.nra`; removed `ConstDecl.lazyPos` (ast), the `lazy` flag from the AST and declarations dumps, the `lazy` printing in `fmt`, the force and cycle logic in `effects.nra` (lazy effect table, call-edge graph), the use checks in `exprs.nra`, and the dispatch rows in `stdlib`, `stdcalls`, `emitstd`.
- E1032-E1035 stay in `diagnostics.nra` with the text `retired: this error is no longer reported` (as E0022 does). Never reuse them.
- The emitter's "lazy constants" (the D61 getter constants) are renamed getter constants (`Consts.getterConstants`); the C is unchanged.
- `stdregex.nra`: `isConstantBody` no longer asks whether the constant is lazy. Regexes.checked in a top-level const is unchanged.
- Removed the lazy samples, goldens (fmt, astdump, declsdump, packages), the SPEC section and mentions, the `compiler/README.md` entry and the guide's "Planned" mention.
- `grep -rn -i lazy selfhost/src std/src SPEC.md docs/guide` finds only unrelated words (lazy generator operators and regex quantifiers).

## Proof

- base: d64eee66 (feat/batch-15): `narrac run testdata/lazy_values.nra` prints `1`, `3`, ... (accepted); the acceptance grep finds 163 lines.
- branch: the sample and its goldens are removed; the grep finds only unrelated words (see Built).

## Deviations

- Deleted `testdata/errors/regex_checked_lazy.*` (outside the owned `lazy*` prefix): it tested E1360 for `lazy Regexes.checked(...)`, which is no longer syntax. The other `regex_checked_*` samples keep covering E1360.
- `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump` changed: the module list no longer has `std/lazy`.
- Renamed emitter helpers in `emit.nra` and `emitconsts.nra` (`lazyConstants` to `getterConstants`, variable `lazy` to `getterConsts`) so the word does not remain in the compiler.
- SPEC.md line "No import side effects" and `docs/guide/README.md` "Planned" list edited to drop `lazy`.

## Stubs left

- None.

## Gotchas

- The seed still knows lazy values; `seed/narrac.c` was not regenerated. The seed update at the end of the batch drops them.
- E1360's explanation text no longer says "not lazy".

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C.
- Emit comparison: base and branch compilers emit byte-identical C for all 350 `testdata/*.nra` programs, and for `selfhost` as of the base commit (11,851,618 bytes, `cmp` identical).
- Size: narrac binary 5,113,160 bytes to 5,095,320 (-17,840). `narrac emit selfhost` instructions retired: 3,525,614,757 (base compiler on base source) to 3,483,282,982 (branch compiler on branch source), about -1.2 percent; C lines 247,390 to 246,410. Measured on a loaded machine.
- Full check: see the final line in the report.

## Size

- Estimated S, actual S.
