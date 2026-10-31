# Outcome: import-alias-rules

## Built

- D72 in the Go compiler. An alias equal to a type name is E0421, with a fix that renames the alias in the import and in every `Alias.` use. Two imports with one alias is E0422. A named import that clashes with a declaration is E0423. Two named imports with one local name from different modules is E0424. `Array.f` without an import is E0908, with a fix that adds the plural import and renames the call. `Array.f` when the module imports std/array under another alias is E0914. Code: `compiler/internal/narra/import_names.go`, `parser.go` (`importDecl`), `check.go`, `modules_check.go`, `check_call.go` (`missingImport`), `diagnostic_fixes.go`, `diagnostic_codes.go`.
- `ImportDecl.Names` is `[]ImportName{Name, Local}`. `import { a as b }` works for std and package modules, for functions, constants, traits and types.
- Narra port (`selfhost/src`): `ImportName` in `ast.nra`, the parser, `astdump.nra`, `declsdump.nra`, `scopes.nra` (renamed imports and the duplicate-alias error E0422).
- Migration: `roadmaps/selfhost/nodes/import-alias-rules/migrate.py` renamed 70 files in `compiler/testdata`, `selfhost` and `bench` to `Arrays`, `Maps`, `Boxes`, `Ranges`, `Options`, `Results`. It is idempotent.
- Tests: `testdata/errors/import_*.nra`, `testdata/packages/errors/import_alias_imported_type`, `testdata/diagnostics/import_alias_array.nra`, and the sample `testdata/import_rename.nra` with `.out`.

## Deviations

- The brief's sample uses `map` from `std/array`, which does not exist. The sample renames `map` from `std/option` and `std/result`, and `length` from `std/array`.
- Edited the goldens `compiler/testdata/declsdump/**` and `astdump/exprs.ast` (aliases changed). `declsdump/imports/alias_clash.dump` now holds an E0422 error line, because that case used the same alias twice. Added the E0422 format to `portedImportErrors` in `selfhost_imports_test.go`.
- Builtin names stay `Array.push` and so on (C names and `stdlib.nra` unchanged). Only the alias changed.
- The user-type hiding of prelude `Box` and `Map` stays (`userBox`, `userMap`). Only the "alias wins" case is gone, because an alias can no longer equal the type.
- `compiler/README.md` updated.

## Stubs left

- The Narra checker does not report E0421, E0423 or E0424. `TestSelfhostImports` skips inputs with those errors. A later selfhost-check node can port them.

## Gotchas

- Builtin names (`Array.push`) are not aliases. Do not rename them.
- The E0421 fix is not offered when a template string mentions `Alias.`, because the parser keeps no positions inside `${}`.
- A module with several files may bind the same alias to the same path in each file.

## Proposed learnings

- A rename script must match the exact import line, not a mention in a comment | the first run damaged error tests | roadmaps | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (every sample under the sanitizer and leak runs, TestSelfhostParser, TestSelfhostImports).

## Size

- Estimated M, actual M.
