# Outcome: qualified-constructors

## Built

- D71 in the Go compiler. A bare user constructor is E0323 with a fix that writes `Enum.Name` (and adds the enum import in a pattern when the module cannot name it). `Some`, `None`, `Ok`, `Err` stay bare. `import { Teacher }` of a constructor is E0425, with a fix that imports the enum. A duplicate constructor in one enum is E0194. The ambiguity error E0300 and the expected-type lookup for bare names are gone. Code: `check_call.go` (`ctor`, `ctorOf`, `enumByName`, `bareOwner`, `moduleEnumName`), `check_match.go` (`requireQualified`, `enumNameFor`), `modules_check.go` (`ctorImport`, dotted `namesEnum`), `check.go`, `parser.go`, `diagnostic_codes.go`.
- `Geo.Shape.Circle` works in expressions and in patterns (`Pattern.Enum` holds `Geo.Shape`), also for generic enums. `ImportName` has a `Pos`.
- Narra port: the pattern loop in `selfhost/src/parser.nra` accepts the three-part form.
- Migration: `roadmaps/selfhost/nodes/qualified-constructors/migrate.py` applies the compiler's own fixes to `compiler/testdata`, `selfhost` and `bench` (45 files). Idempotent.
- Tests: `errors/ctor_bare_{expr,match,is}.nra`, `errors/ctor_dup.nra`, `diagnostics/ctor_bare.nra`, `packages/errors/ctor_import`, `internal/narra/ctor_import_test.go` (applies the E0425 fix), samples `ctor_qualified.nra` and `packages/ctor_qualified`.

## Deviations

- `export type A = B` is not implemented in the Go compiler, so constructors through an alias are untested. Noted in `compiler/README.md`.
- Edited files outside the brief's paths: `compiler/internal/narra/parser.go`, `ast.go` (three-part patterns, `ImportName.Pos`), `compiler/internal/narra/ctor_import_test.go` (new), `selfhost/CLAUDE.md` (the "global constructors" rule).
- The constructor-import diagnostics case is a Go test (`ctor_import_test.go`), not a `testdata/diagnostics` case, because that loader takes single files.
- Removed `errors/traits_ctor_ambiguous.nra` and `packages/errors/ctor_imported_ambiguous`: the program is now valid.
- The migration added enum imports to selfhost files that matched on enums they did not import (`resolve.nra` now lists `Decl` in its existing ast import).
- Bare constructor shared by two enums of one rank is E0324 with no fix; the expected type picks the enum otherwise, also through a namespace (`Geo.Shape.Circle`).

## Stubs left

- The Narra checker does not exist yet for this rule. selfhost-check-patterns and selfhost-check-calls port E0323, E0425 and E0194.
- The E0425 fix does not dedupe two constructors of one enum in one import line.

## Gotchas

- Names such as `Type.TStruct` stutter; not renamed (separate optional change).
- In a multi-file module the import added by the E0323 fix goes to the first file's import position.
- Wrap `Int.toString` in template parts in samples; bare ints are rejected.

## Proposed learnings

- A fix that names a type must also import it when the module does not | the first migration produced `Decl.Struct` in a module without `Decl` | compiler/internal/narra | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (samples under sanitizer and leak runs, every selfhost harness).
- Brief verify command: pass.

## Size

- Estimated M, actual M.
