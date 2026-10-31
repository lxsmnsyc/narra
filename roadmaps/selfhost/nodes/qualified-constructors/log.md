# Log: qualified-constructors

- Qualified forms already worked; the work was the new errors and the migration.
- `ctor` now takes the enum from its name only. Bare names are E0323 with a fix, except Some/None/Ok/Err. The ambiguity error (E0300) is gone from the source; the table row stays so the code is never reused.
- Three-part form `Geo.Shape.Circle`: expression through `moduleEnumName`, pattern through a dotted `Pattern.Enum` (parser loop in Go and in selfhost/src/parser.nra).
- Dead end: the first migration used the first rank for the fix and wrote `Rank.Less` where the expected type was `Ordering`. `bareOwner` now checks the expected type first.
- Dead end: a pattern on an enum the module cannot name (`Decl` in resolve.nra) got `Decl.Struct` with no import. The fix now adds `import { Decl } from "narrac/ast"`.
- `export type A = B` is not implemented in the Go compiler, so the alias route in the brief cannot be tested. Left as a gap in compiler/README.md.
- Removed tests for the ambiguity error: errors/traits_ctor_ambiguous.nra and packages/errors/ctor_imported_ambiguous.
