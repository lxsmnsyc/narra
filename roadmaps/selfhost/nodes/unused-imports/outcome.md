# Outcome: unused-imports

## Built

- E0946: an import name (`import { a, b as c }`) or module alias (`import * as A`) the module never uses. The message is `"Str" is imported but never used; remove it, or import it as "_Str" to mark it as unused on purpose`, at the name or alias. The fix removes the name with its comma. When no name of the import is used, each fix removes the whole import (its lines, up to the line with the path).
- E0947: a private function that nothing other than itself refers to. The message is `private function "triple" is never used; remove it, or name it "_triple" to mark it as unused on purpose`, at the name. The fix adds `_` to the name and to its own calls. These are found by a text scan from the declaration to the first line that starts with `}`, skipping comments, strings and chars.
- A name starting with `_` is unused on purpose. This covers the names the compiler writes: the shape-derive hidden imports `__Shape`, `__Value`, `__ShapeText`, the functions `__shapeToValue_<T>`, `__shapeFromValue_<T>`, `__shapePath`, `__shapeKind`, and `__for_L_C`. Closures are not top-level functions.
- When an import counts as used: its local name is written in the module outside the imports and not after a `.` (`SourceFacts.referenced`, filled in `parser.nra` `finishFacts`). The tokens include the `${}` parts of templates, which are lexed again. Types, traits, `derives` lists and patterns all count.
- When a function counts as used: some body other than its own (another function, a test block, a constant, a field default, an implement function) has a `Ref.Function` to it, as a call or a function value. A test block that is the only user counts. A function a binary names as entry counts as used, so E0419 and the entry errors still report. So do the std/shape `*ToValue`/`*FromValue` functions, which `setShapeRef` calls by name.
- The check runs once, at the end of `body.nra` `checkPackage` after `finish`, and only when the package has no error so far. Errors from later passes (effects, emit, entry checks in the driver) come after it, so a sample whose only error is in a later pass must have no unused names.
- Code: new `selfhost/src/unused.nra` (`reportUnusedTopLevel`). Also `ast.nra` (`SourceFacts.referenced`), `parser.nra` (`collectReferencedNames`, `collectTemplateNames`), `body.nra` (the call) and `diagnostics.nra` (two rows).
- Tests: `testdata/errors/unused_import.nra` (one unused name of two), `unused_import_alias.nra`, `unused_import_all.nra` (every name unused, so the import goes), `unused_function.nra`, `unused_function_recursive.nra`. All five are also habits `fix/` cases, whose fix must leave a clean file. `testdata/unused_top_level.nra` (program) has `_` imports, a `_` function, a struct that derives Shape and never calls Shape, a for-destructuring, and a function used only as a value. `testdata/tests.nra` already had a private function (`double`) that only a test block uses.
- Docs: `compiler/README.md` replaces "Unused imports and private functions are not errors yet" (two places) with a bullet that describes the rule. `docs/guide/03-*.md` gets one sentence on imports and private functions in "A name you do not use".

## Cleanup counts

- `selfhost/src`: 59 unused import names removed in 22 files. One private function removed (`fmt.nra` `getPatternPosition`). `selfhost/tests`: none.
- `std/src`: none (the shape providers are exempt). stdtext.nra did not change.
- testdata programs and packages: 9 imports removed (json 2, json_floats 1, shape 1, sized_int 1, packages/const 1, packages/opaque_range 3).
- testdata: 64 private functions renamed with `_`: 31 in errors/ (effect and generator error samples, whose error comes after the check), 29 in typesdump/, 1 in query/basic (`hidden`), and 1 each in cnames/edited, runner-cases/no_tests and emit-errors/noentry. The typesdump and query goldens were regenerated; the diffs are renames and column shifts.
- Habit samples: habits/throw_call, habits/throw_result, diagnostics/ts_throw and diagnostics/ts_number now call their private function from `run`. packages/errors/ctor_import got an exported `describe(_person: SchoolPerson)`, and `expected.err` gained the line `src/main.nra:4:35: unknown type "SchoolPerson"`.
- Guide samples, bench: none were reported (guide suite passes).
- Error samples whose expected error the checker reports were not combed: the check is silent there.

## New codes

- E0946, E0947 (the reserved ones).

## Gate decisions

- Import usage is checked on tokens, not by resolution. Alternative: mark use at every name lookup in the checker (about 70 sites over types, functions, constants, traits and module aliases, in hot files). The token rule is short and lives in the parser. It can only miss an error, never report a wrong one: a field, a field key or a named argument with the same name counts as a use. That gap is in the README.
- Private function usage uses the checker's refs (`Ref.Function`), because it must exclude the function's own body; tokens cannot.
- Report only when the package has no other error, after `finish`. Alternative: per-body gating like E0921. Top-level names have no single body, and an earlier error can hide a use.
- The function fix renames to `_name` rather than deleting the function. A deletion needs the function's end and its doc comment, which the AST does not record. The cleanup in `selfhost/src` deleted the function by hand.
- When every name of an import is unused, each diagnostic's fix removes the whole import. Alternative: remove one name per fix, which leaves `import {} from` after the last one.
- Exemptions: the entry function a binary names, and the std/shape provider functions (called by name). Alternative for std/shape: export them, which would add public API.

## Deviations

- Edited outside the owned paths: `selfhost/src/{ast,parser,body,diagnostics,unused}.nra` plus the cleanup files, `compiler/README.md`, `docs/guide/03-*.md`, the testdata files listed above. All of these are shared paths in the brief.

## Stubs left

- None.

## Gotchas

- A new sample with an error that a later pass reports (effects, emit, a missing entry) must have no unused imports or private functions, or E0946/E0947 come first.
- The fix scan for a function's own calls stops at the first line that starts with `}`. An unformatted one-line function scans on until such a line.
- Typesdump samples are lists of private functions. A new one needs `export` or a `_` name.

## Proposed learnings

- A sample for a later-pass error (effects, emit) must not leave an import or private function unused | E0946/E0947 report first and hide the expected error | testdata/errors | yes

## Proposed follow-ups

- None needed. A stricter import rule (resolution-based) only if the field-name gap matters.

## Verification

- Suites run one at a time with the built `out/narrac`, all passing at the end: programs, errors, packages, habits, codes, panics, dumps, query, lsp, std-tests, guide, sources, fmt, fmt-stable, fmt-doc, trivia, node-ids, edge, runner, runner-options, manifest, emit, generated, bootstrap. No full `scripts/test.sh` run, per the batch rule.
- `sh selfhost/bootstrap.sh out` reaches the fixed point (a stage 3 line, older seed runtime). The committed seed builds the cleaned `selfhost/src`.

## Size

- Estimated M, actual M.
