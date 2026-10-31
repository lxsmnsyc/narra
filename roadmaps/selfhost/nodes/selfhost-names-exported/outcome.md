# Outcome: selfhost-names-exported

## Built

One scripted rename pass over all of `selfhost/src` (token-aware: string, char and template text untouched, code inside `${}` renamed). The same word is renamed everywhere it is an identifier: fields, variant payloads, locals, parameters. No behaviour, dump text or message changed.

### Renamed names, old -> new (for selfhost-emit)

Fields, variant payload names and the locals that carried them:

| old | new |
| --- | --- |
| `args` | `arguments` (TypeExpr, CallExpr, PatCtor, PatTuple, EnumInfo, StructInfo) |
| `params` | `parameters` (FuncDecl, ArrowExpr, Checker, FuncInstance, GenericInfo, StdFunc) |
| `decls` | `declarations` (File, ModScope) |
| `stmts` | `statements` (Block) |
| `funcs` | `functions` (TraitDecl, ImplDecl, ImplInfo, Checker, ModScope, Registry) |
| `typeParams` | `typeParameters` |
| `effectParams` | `effectParameters` |
| `impls` | `implementations` (StructDecl, EnumDecl, Checker) |
| `vals` | `values` (MapLit) |
| `elems` | `elements` (ArrayLit) |
| `cond` | `condition` (IfStmt, WhileStmt) |
| `iter` | `iterable` (ForStmt) |
| `lit` | `literal` (PatLit) |
| `op` | `operator` (UnaryExpr, BinaryExpr) |
| `elem` | `element` (Type.TArray, Type.TBox) |
| `val` | `valueType` (Type.TMap) |
| `col`, `endCol` | `column`, `endColumn` (Pos, Token, Lexer, LoadError) |
| `errs` | `errors` (Checker, Ordered, Walk) |
| `cur` | `current` (Checker) |
| `env` | `environment` (BodyState, Checker, FuncInstance) |
| `targs` | `typeArguments` (BodyInfo, FuncInstance; the dump keyword `targs=` is unchanged) |
| `typeArgs` | `typeArguments` (MapLit, StructLit) |
| `decl` | `declaration` (Callee, ConstInfo, FuncInfo, ImplFunc, ImplInfo, and others) |
| `dir` | `directory` (Package) |
| `pkgName`, `pkg` | `packageName`, `package` |
| `localN` | `localCount` |
| `traitIdx` | `traitIndex` |
| `derivExprs`, `implExprs` | `derivedExpressions`, `implementationExpressions` |
| `instTypes`, `errKeys` | `instanceTypes`, `errorKeys` |
| `mods` | `modules` (Ordered, Walk) |
| `expr` | `expression` (ExprStmt, RootExpr) |
| `ctorType` | `constructorType` (ConstructorResult) |
| `func` | `funcIndex` (EntryInfo, FuncInstance) |
| `hintParams`, `selfParam` | `hintParameters`, `selfParameter` |
| `implFuncs` | `implFunctions` |
| `fieldDecls`, `variantDecls` | `fieldDeclarations`, `variantDeclarations` |

Functions and types:

| old | new | file |
| --- | --- | --- |
| `Diag` | `Diagnostic` | parser.nra |
| `diagText` | `formatDiagnostic` | parser.nra |
| `kindName` | `getKindName` | lexer.nra |
| `funcIds`, `blockIds`, `exprIds` | `collectFuncIds`, `collectBlockIds`, `collectExpressionIds` | idcheck.nra |
| `errorf`, `typeErrorf`, `errorIn` | `reportError`, `reportTypeError`, `reportErrorIn` | check.nra |
| `irrefutable` | `isIrrefutable` | patterns.nra |
| `declName` | `getDeclarationName` | scopes.nra |
| `areArgsIrrefutable` | `areArgumentsIrrefutable` | patterns.nra |
| `anyTypeParam`, `containsTypeParam` | `anyTypeParameter`, `containsTypeParameter` | |
| `setParamType`, `makeParams` | `setParameterType`, `makeParameters` | |

In `containsOpen` (exprs.nra) the `Type.TFunc` binding is `functionParameters` to avoid shadowing.

## Deviations

- Kept as they are: `id` (domain word, per the brief), `pos` / `Pos` (used in about 60 structs and in Go's AST under the same name), `ast`, `std`, `isFunc`, `binary` and `exported` (a function and a field share each name), type names that mirror Go (`FuncDecl`, `ImplInfo`, `PatCtor`, `Arg`, `Param`, `ModScope`), `Tok` prefixes, and exported functions with noun names that are not abbreviations (`typeKey`, `typeList`, `funcBounds`, `implFunc`, `enter`, `finish`, `lex`, `hex`, `run`, `load`). Renaming those to action names is a separate change if wanted.
- Comments: the words `func`, `lit`, `val`, `iter`, `cond`, `op`, `cur`, `mods`, `expr`, `decl`, `env`, `dir`, `col`, `pkg` were not rewritten in comments (ordinary prose there); other renamed identifiers were.
- No file outside `selfhost/src` and this node folder changed. `selfhost/CLAUDE.md` needed no edit.

## Stubs left

- None.

## Gotchas

- Narra reports shadowing when a rename makes a binding clash with an existing name; run `narra check selfhost` after a rename pass.
- Go dump keywords (`targs=`, `ctor=`) stay short; only the Narra field names changed.

## Proposed learnings

- None.

## Verification

- Literal check: every string, char and template text in `selfhost/src` is identical to origin/main (0 files differ).
- `go test -count=1 ./internal/narra -run 'TestSelfhost|TestASTDump|TestTypesDump' -v`: per-test PASS and SKIP counts identical before and after (5247 subtests; TestSelfhostBodies 353 pass / 30 skip, TestSelfhostDecls 283/87, TestSelfhostImports 348/32, TestSelfhostTraits 304/66, TestASTDump 416, TestTypesDump 19).
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass.

## Size

- Estimated M, actual M.
