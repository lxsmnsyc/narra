# Outcome: type-of

## Built

- `typeof name` in a type position (D96). `typeof` is a contextual word: `parseTypeBody` (`selfhost/src/parser.nra`) reads it when a name follows. `TypeExpr.typeofName` holds the name (`name` is empty); `SourceFacts.hasTypeof` marks files that use it.
- `resolveTypeof` in `selfhost/src/resolve.nra`: a parameter written earlier in the same signature (`Checker.signatureNames`), a local or parameter of the body, a top-level constant (own or imported) and a function (own, imported, or declared further down the file). A function gives the full function type with its declared effects and `from`. `checkModuleBodies` (`body.nra`) checks the constants first in a file that uses `typeof`, so bodies see inferred constant types.
- E0981 (a name resolved by its own type, `Checker.busy` key `typeof|name`; `checkVariable` in `stmts.nra` marks a local's own annotation) and E0982 (a type, a module, a generic function, a function with effect variables or a `mutable` parameter, and a same-module constant without a written type in a declaration). Rows in `diagnostics.nra`.
- `fmt` (`printTypeBody`), `astdump` (`typeof="name"` attribute), `narrac query` signatures (`typeof name` as written; hover shows the resolved type).
- Samples: `testdata/typeof_basic`, `typeof_forward`, `typeof_unused` (programs), `errors/typeof_self`, `typeof_not_value`, `typeof_path`, `typeof_inferred_const`, `fmt/typeof_names`, `astdump/typeof_names`, `query/typeof_names/`. SPEC.md has a new "typeof" subsection with two checked samples; D96 and `compiler/README.md` updated.

## Deviations

- Same-module constants without a written type cannot be the target of `typeof` in a signature, field or alias (E0982 asks for the written type). The declaration passes run before any value is checked and `resolve.nra` cannot call the expression checker (no import cycles). Bodies and other modules work. Lifting it needs the declaration passes to check constants on demand. Recorded in D96, SPEC.md and `compiler/README.md`.
- The brief says query signatures show the resolved type; they are written from syntax, so signatures print `typeof name` and hover shows the resolved type.
- A parameter read only by a `typeof` is still reported as never read (E0921/E0922); name it with `_`.
- The parse error for a field path after `typeof` is the generic `expected ")", found "."` (E0011); no new code.
- Files outside the owned paths (shared): `selfhost/src/{ast,astdump,body,check,diagnostics,fmtexprs,parser,querysignature,resolve,stmts}.nra`, `selfhost/tests/src/fmtsuite.nra` (golden prefix `typeof_`), SPEC.md, DECISIONS.md, `compiler/README.md`.

## Stubs left

- None.

## Gotchas

- `typeof` of a function resolves the signature in the current module with `functionIndex` -1, so a function with effect variables is rejected.
- Adding a field to `TypeExpr` or `SourceFacts` is cheap when it has a default.
- The seed does not read `typeof`; `selfhost/src` must not use it until the next seed update.

## Proposed learnings

- None.

## Verification

- `NARRAC=<built> sh scripts/test.sh --filter typeof`, `--suite spec`: pass.
- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C.
- `NARRAC=<built> sh scripts/test.sh` (full): PASS, 6059 cases, 0 failed.

## Size

- Estimated M, actual M.
