# Outcome: export-key-diagnostics

## Built

- `ImportDecl.writtenPath` (`syntax/ast.nra`, `parser.nra`, `driver/exports.nra`): the path the importer wrote. E0101 and E0402 (`check/scopes.nra`, new `exportedWritten`) name the written key and add a note with the real module. E0421 suggests the written path. `narrac pkg` imports lines (`driver/modules.nra`) show the written key.
- E0474 lists the exposed std modules (`moduleList`).
- `removeImport` (`check/unused.nra`) skips quotes inside `//` comments.
- Cases: `testdata/deps/build/error-export-key` (deps suite, `selfhost/tests/src/depsbuild.nra`), `testdata/errors/unused_import_multiline` (errors and habits fix suites), six rewritten E0474 goldens.

## Proof

- base: `NARRAC=<seed build> sh scripts/test.sh --suite deps --filter export-key` fails with `module "lib/streaming/reader" has no export "missing"`. `apply-fix testdata/errors/unused_import_multiline.nra` leaves `  Shape,` at the top of the output. E0474 prints no list.
- branch: the same commands pass; the fix output starts with a blank line and the file checks clean.

## Deviations

- `testdata/errors/*.err` (six std-internal cases) and `testdata/stage2/errors/std_internal_import.err` changed with the E0474 text. The `declsdump` import line keeps the resolved path.
- The E0474 message is long (about 80 module names).

## Stubs left

- None.

## Gotchas

- Anything that reads `ImportDecl.path` sees the real module; use `writtenPath` for text shown to the user.

## Proposed learnings

- None.

## Verification

- See the final report for the full `sh scripts/test.sh` line.

## Size

- Estimated M, actual M.
