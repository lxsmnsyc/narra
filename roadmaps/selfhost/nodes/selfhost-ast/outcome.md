# Outcome: selfhost-ast

## Built

- `compiler/internal/narra/astdump.go`: `ASTDump(src) (string, bool)`. The doc comment at the top of the file defines the dump format completely (lines, attributes, slots, every node kind, errors). Read it first.
- `compiler/cmd/narra/main.go`: hidden command `narra parse <file>`; prints the dump, or `error <line>:<col>: <message>` and exits 1.
- `compiler/internal/narra/astdump_test.go`: `TestASTDump` dumps every `.nra` under `compiler/testdata/` twice (no panic, equal output), compares `compiler/testdata/astdump/{decls,exprs,patterns,error}.nra` with their `.ast` goldens, and checks that the goldens contain every node kind. `NARRA_UPDATE_GOLDEN=1` rewrites the goldens.
- `selfhost/src/ast.nra` (module `narrac/ast`): Narra types, one per Go type in `ast.go` (structs for declarations and helpers; enums `Decl`, `Stmt`, `Expr`, `Pattern`). Its header comment lists the differences from Go.
- `selfhost/src/astdump.nra` (module `narrac/astdump`): `dump(f: File): string uses alloc`, the same printer. It cannot be diffed against the Go dump until a Narra parser exists; it compiles and printed the right text for a hand-built tree.

### Dump format in short

One node per line, two spaces per depth: `<kind> <line>:<col> <attrs>`. Slot lines `.name` (no position) group children. Attrs: `name="escaped"`, bare flags (`export opaque mutable escaping let func tuple bare`), `list=a,b`. Single-list nodes (block, template, array, pat-ctor, pat-struct, pat-tuple, variant, exprstmt, return, unary, field-expr, pat-lit) print children directly. Ids, checker fields and `File.Path` are not printed. Example:

```
file
  .decls
    func 1:2 name="main" export uses=io
      .result
        type 1:2 name="int"
      .body
        block 1:2
          return 1:2
```

## Deviations

- Only `compiler/cmd/narra/main.go` (shared) was touched outside owned paths. `compiler/README.md` is unchanged: it documents no hidden commands.

## Stubs left

- `selfhost/src/main.nra` still prints tokens; selfhost-front adds the parser and wires `dump` in.

## Gotchas

- Narra AST differences from Go (also in the header of `ast.nra`): ids on expressions, patterns, `Param`, `VarStmt`, `ForStmt`; field names avoid keywords (`exported`, `isMutable`, `effects`, `impls`, `derived`, `orElse`, `traitType`); `PatCtor.bare` replaces nil `Args`; `CharLit.value` is a one-char string; checker-only fields are not mirrored; `IfStmt.orElse` is `Option<Stmt>` (BlockStmt or IfStmt). Every Go AST type has a Narra type.
- Positions the parser must reproduce: a declaration's pos is its keyword (`const`, `function`), not `export`; a call's pos is the `(`; a field-expr's pos is the member name; binary pos is the operator; index pos is the `[`; a struct literal with a type name uses the name's pos; template parts all take the template's pos (inner tokens are re-posed) and its string pieces carry the template's `end`; `x is P` has the pos of `is`; a shorthand pattern field `{ a }` makes a `pat-name` at the field's pos.
- `Name()` and `Name` differ by `bare`. A lone `Empty` is a `pat-name`; the checker decides later.
- Go parser quirk: `uses E, mutable x: T` inside a parameter list fails to parse (`paramAfterComma` ignores `mutable`). The samples avoid it.
- `bool` prints `value=true|false`; `string` has `end=<line>:<col>`; `char` prints the character escaped like text.
- Enum variant names are global constructors in Narra (`Func`, `Struct`, `Test`); astdump imports them by name from `narrac/ast`.
- `narra build selfhost` type-checks modules nobody imports, so a broken `astdump.nra` fails `TestSelfhostLexer`.

## Proposed learnings

- Print optional and repeated children under slot lines so absent and empty differ in a dump | a byte diff then points at the wrong field | selfhost/, compiler/internal/narra/astdump.go | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.
- `go test ./internal/narra -run 'TestASTDump|TestSelfhostLexer'`: pass.
- Scratch program with a hand-built tree through `astdump.dump`: output matches the format.

## Size

- Estimated M, actual M.
