# Outcome: selfhost-check-bodies

## Built

The core of the body checker in Narra, the `narra types` / `narrac types` dump, its harness and goldens, and the import name errors E0421, E0423 and E0424. About 1,990 new Narra lines (exprs 1237, stmts 237, body 234, typesdump 287) plus about 450 lines in check.nra, scopes.nra and declare.nra. The Go side adds `typesdump.go` (409 lines) and two tests.

### Modules (selfhost/src)

| module | ports |
| --- | --- |
| `narrac/exprs` (exprs.nra) | `checkExpression` (Go's `expr`), `checkExpressionKind` (`exprInner`, one `match`), names (`checkName`), fields (`checkField`, with `moduleConstValue`), index, unary, `checkBinary`, `checkStructLiteral` (with `typeArgs`, E0237, and generic struct inference), array literals, `checkCall` (plain `call`), `checkArguments` (`args`), `checkBuiltinCall` (check kind "" and "fixed"), `checkMutableArgument`, `findAssignedLocal` (`lvalueRoot`), `checkConstWrite`, `checkConstType` (`constType`), `checkConstValue`, `describeCallee` (`calleeName`), `isConstantValue` (`isConstant`), `expectType`, `describeType` (`typeName`), the integer literal check |
| `narrac/stmts` (stmts.nra) | `checkBlock`, `checkStatement` (`stmt`), plain `ifStmt`, `isResult`, `diverges` |
| `narrac/body` (body.nra) | `checkPackage` (Go's `checkPackageBodies`), `BodyWalker`, field defaults (`checkDefaults`, also of generic structs), function bodies (`funcBody`, `funcParam`), constants, tests (`inTest`), the impl stub |
| `narrac/typesdump` (typesdump.nra) | `dump(checker)`: the `narra types` format |

Import order: check <- exprs <- stmts <- body; typesdump imports check and declsdump; main imports body and typesdump. In new Narra code, names follow the maintainer's rule (see Gotchas).

### Dispatch layout: the Walker trait (the cycle decision)

A Narra function type cannot take a `mutable` parameter, so a closure in a struct field cannot carry `mutable checker`. A trait can. check.nra declares:

```
export trait Walker {
  function checkExpression(walker: Self, mutable checker: Checker, expression: Expr, want: Option<Type>): Type uses alloc
  function checkBlock(walker: Self, mutable checker: Checker, block: Block) uses alloc
}
```

body.nra implements it with `struct BodyWalker implements Walker {}`. Every function in exprs.nra and stmts.nra is generic, `<W: Walker>(mutable checker: Checker, walker: W, ...)`, and passes `walker` on. A later node's module sits below exprs.nra (exprs imports it), takes `walker: W` with `<W: Walker>`, and calls back with `Walker.checkExpression(walker, mutable checker, x, want)` or `Walker.checkBlock(walker, mutable checker, block)`. Its arm in `checkExpressionKind` becomes one line calling into its module. Add a function to the trait (and to `BodyWalker`) if a module needs another entry point, such as `checkStatement`.

### Per-body side tables (D70)

In check.nra:

```
enum Ref { Unresolved, Local(index) /* BodyInfo.locals */, Const(index) /* Checker.consts */,
           Function(index) /* Checker.funcs */, Builtin(index) /* Registry.funcs */ }
struct LocalInfo { name, localType: Type, isMutable, param, pos, noEscape, used }   // narrowing goes beside it
enum BodyRoot { RootFunc(decl: FuncDecl), RootBlock(block: Block), RootExpr(expr: Expr) }
struct BodyInfo {
  module, kind ("func" | "test" | "impl" | "const" | "default"), name (header text), targs: Array<Type>,
  root: BodyRoot,
  first: int,                    // lowest node id of the body
  types: Array<Option<Type>>,    // by id - first; None = not checked (Go's nil)
  refs: Array<Ref>,              // by id - first
  declared: Array<int>,          // by id - first: the local a Param/VarStmt/ForStmt declares, or -1
  locals: Array<LocalInfo>,
}
```

New `Checker` fields: `checkingBodies` (set by `narrac types`; `narrac decls` keeps skipping defaults), `pendingDefaults` (generic structs built since the last default check), `bodies`, `body` (current, -1), `localScopes` + `scopeDepth` (the block scope stack; std/array has no pop, so slots are reused), `resultType`, `loops`, `inTest`. `ConstInfo` gained `state` (`CONST_NEW`, `CONST_BUSY`, `CONST_DONE`) and `valueType`.

Helpers in check.nra: `startBody(checker, kind, name, targs, root, resultType, env): BodyState` (walks the root with idcheck's exported `funcIds`, `blockIds` and `exprIds` to find the id range, saves the outer body's state, opens one scope), `finishBody(checker, saved)`, `findBodySlot(body, id)`, `setNodeType`, `getNodeType`, `setRef`, `getRef`, `findLocal` (Go's `lookup`), `getLocal`, `declareLocal` (Go's `declare`, with the shadowing error), `pushScope`, `popScope`, `unsupported`, `getExpressionId`, `getExpressionPos`, `isNumeric`. A constant checked at its first use opens its own body in the middle of another one; `BodyState` restores the outer one. `env` is saved too.

Where things are recorded: the type of every checked expression is set in `checkExpression`. Refs go on the Ident (local, const), on the FieldExpr (`Alias.CONST`: const) and on the CallExpr (function, builtin). The Ident of `Alias` in `Alias.CONST` gets type `TModule(path)`. A callee Ident gets no type, as in Go.

### Driver

`checkPackage(pkg)`: `newPackageChecker`, `checkingBodies = true`, then for each module: `declareSourceModule`, `checkDefaults` (queued generic struct defaults first, then the non-generic structs' defaults in declaration order), `checkModuleBodies` (functions, constants, impl blocks, tests in declaration order). Then `flushMapKeys`, `checkEntries`, `finish`. The pending generic instance queue does not exist yet: selfhost-check-calls adds it after the module loop, before `flushMapKeys`, as Go's `checkPending` does. `declarePackage` (`narrac decls`) is unchanged and still calls `declareConstTypes`; the body checker does not, because `checkConstType` resolves the annotation where Go does.

### The types dump (frozen)

The format is in the doc comment of `compiler/internal/narra/typesdump.go`. It has one block per checked body, in the order the checks start, with headers `body module="<m>" func|test|impl|const|default ...` and ` targs=` for an instance. Each node is a line `<line>:<col> <kind> [attr...] <type>`, and each local is a line `<line>:<col> local <name> [mutable] <type>`. Errors print as `error <file>:<line>:<col>: <message>` on stderr with exit 1. The Owners list there says which node prints which attributes. The Go dump already prints `ctor=`, `narrow=`, `trait=`, `value`, `distinct=` and `func=` on instances and impl bodies, so later nodes only make the Narra side match.

Go changes (no output change for emit or check; 346 inputs x `emit`, `check`, `check --json` byte-identical to a main build):
- `modules_check.go`: `CheckPackage` = `checkPackageBodies(pkg, false)` + the effect pass.
- `check.go`: `Checker.logBodies`, `Checker.bodies`. `checkBodies` and `checkPending` call `logBody`.
- `check_types.go`: `checkDefaults(owner string, fields)` logs each default (owner `m.Point`, or `m.Pair` for a generic).
- `check_const.go`: `constType` logs the constant. `check_traits.go`: `implBodies` logs each function with its source name (`implSourceName` in typesdump.go).
- `cmd/narra/main.go`: hidden `narra types`.

### Tests

- `TestTypesDump` (typesdump_test.go): goldens in `compiler/testdata/typesdump/bodies/`. These are `basics`, `generic_structs`, `module_const/` (package: `Alias.CONST`, module type, imported struct), `instances` (Go only: generic instances with targs, impl body, for loop local), and the error cases `err_names`, `err_mismatch`, `err_arg_count`, `err_assign`, `err_return`, `err_loops`, `err_const_rules`, `err_int_literal`, `err_template`, `err_struct_literal` (missing, unknown and repeated fields, E0237, uninferable parameter), `err_fields`, `err_statements`, `err_defaults` and `err_opaque_field/`. An `err_` case must fail and no other may. `NARRA_UPDATE=1` rewrites them. `TestTypesDumpCoverage` requires every header kind, line kind and attribute of this node to appear in some golden.
- `TestSelfhostBodies` (selfhost_bodies_test.go): `narrac types` against `narra types` on `importInputs` plus the package cases of typesdump/bodies. It skips only "does not load" or a Narra output with `unsupported:`, logs each skip with its construct, and prints the counts.
- `TestSelfhostImports`: `portedImportErrors` now lists E0421, E0423, E0424 and E0194. The E042x inputs are compared, not skipped.

### Compared and skipped (TestSelfhostBodies)

The test compares 143 inputs and skips 202:

| count | reason | node |
| --- | --- | --- |
| 36 | an enum constructor | calls |
| 30 | does not load | (TestSelfhostModules) |
| 19 | implement block bodies | calls |
| 14 | a match expression | patterns |
| 10 each | a Map literal; a tuple literal | patterns |
| 10 | a distinct type function | calls |
| 10 | an arrow function | closures |
| 8 | a call of a function value | closures |
| 7, 4, 3, 2, 2, 2, 1 x 7 | builtins with a named check: Array.length, Test.eq, Array.push, Array.values, Map.size, Range.upTo, Box.new, Option.isSome, Option.map, Result.mapErr, Test.golden, Test.panicsWith | builtins |
| 5 | a constant declared twice | (see Stubs) |
| 5 | a namespace without its import (E0908) | calls |
| 3 each | a destructuring declaration (patterns); a mutable argument beside other arguments; a trait function call (calls) | |
| 2 each | `is` (patterns); a call of a generic function (calls); a non-escaping function parameter (closures) | |
| 1 each | a Map index, a spread in a struct literal (patterns); a call of a package module function, an import of a constructor (calls) | |

No compared input differs.

## Deviations

- Files edited outside the owned and shared paths. Go: `check_types.go` (checkDefaults signature and log line), `check_const.go`, `check_traits.go` (one log line each). Narra: `idcheck.nra` (`funcIds`, `blockIds`, `exprIds` exported), `declsdump.nra` (`dumpTypeString` exported), `resolve.nra` (its HOOK queues generic struct defaults). Also `DECISIONS.md` (D70, reserved) and the goldens.
- Format details against the brief's plan. The type comes last and attributes come before it, following the decls dump rule, because types contain spaces. Kinds are the AST dump's names (`ident`, `call`, `field-expr`), not Go type names. The local line kind is `local`. A resolved namespace alias shows as the type `module <path>` instead of a `module=` attribute.
- Generic struct defaults are checked after the module's declaration passes, not inside `buildGeneric`. The bodies come out in the same order. The error order can differ from Go only when a trait error of the same module falls in between.
- Ported beyond the brief: E0194 (a constructor declared twice, `declare.nra`) and E0424 at its Go place (the std import loop in `scopes.nra`).
- The dedup quadratic-sort inbox item (`sortStrings` in modules.nra) is not fixed. It stays in the inbox.
- No change requests.

## Stubs left

Every stub reports `unsupported: <construct> (<node>)` through `unsupported()` in check.nra.
- selfhost-check-calls: enum constructors (bare names, `Enum.V`, `Alias.Enum.V`, and the conservative `hasConstructor` check, which is yes for any variant name of any enum), generic calls and the pending queue, `checkExclusive` (the stub in exprs.nra makes any call with a mutable argument plus another argument unsupported), trait calls, package-module calls, distinct calls, E0908 (`isMissingImport`), implement block bodies (`checkImplBodies` in body.nra), constructor imports E0425 (scopes.nra: they need `ImportName` positions). `checkConstCall` must accept constructor calls, and `isConstantValue` must accept constructors.
- selfhost-check-patterns: `is`, match (expression and statement), destructuring `VarStmt`, Map literals, Map index (and `assigning`/MapSet), tuple literals, spreads, narrowing (`checkField`'s enum case reports Go's "may hold any of its variants" error because nothing is ever narrowed yet; `checkIf` has no `is` handling).
- selfhost-check-closures: arrow functions, calls of function values and function fields, functions used as values (`Name`, `Alias.f`), non-escaping parameters (locals with `noEscape`, arguments for such parameters).
- selfhost-check-builtins: every `StdFunc.check` other than "" and "fixed" (`checkBuiltinCall`). `for` loops need Range/Array/Map builtins, so `checkFor` is ported but only reachable once they are.
- A constant declared twice is unsupported: Go checks the duplicate's value too, and the Narra tables have no entry for it.

## Gotchas

- Maintainer rule for all new Narra code (2026-10-31): full descriptive names (`checker`, not `c`; `expression`, `leftType`, `index`, `local`, `parameters`), function declarations with action names (`checkBinary`, `findLocal`, never `<noun>Of`), for-of loops where Narra allows, and short comments on the steps inside a function. Do not rename existing code you do not touch.
- `function` is a keyword: it cannot be a binding name (`is Some(function)` fails to parse). `from` and `to` are keywords too.
- Write into the body tables only through `setNodeType` and `setRef`. They ignore ids outside the body's range.
- `getNodeType` reads the current body. The struct literal uses it to skip a field already checked during generic inference, as Go reads `Type()`.
- A later module that needs Checker data from body tables imports check.nra only. Never import exprs.nra from below.
- In Go, `TypesDump` runs with `logBodies`. Any new place in Go that checks a body (a new kind) must call `logBody` and get a header kind in the doc comment.
- `typesdump.go` panics on an unknown expression or pattern kind. A new AST node needs a case there and in typesdump.nra.

## Proposed learnings

- Reach a recursive checker from a lower Narra module through a trait implemented by a marker struct, not through a closure | function types cannot take `mutable` parameters, and modules cannot import in a cycle | selfhost/ | no
- Mark a construct a later node ports with `unsupported:` and skip on it in the harness | partial ports stay comparable with no false mismatch | selfhost/, compiler/internal/narra | yes (TestSelfhostBodies)

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 99 s).
- The brief's verify command (`go test ./internal/narra -run 'TestTypesDump|TestSelfhostBodies|TestSelfhostDecls|TestSelfhostTraits|TestSelfhostImports'`): pass.
- `narra emit`, `narra check` and `narra check --json` on 346 inputs (every `.nra` under `compiler/testdata` outside packages/ and pkgdump/, every narra.toml directory, selfhost/) are byte-identical to a build of origin/main.
- `narrac types selfhost`: 0.5 s wall.

## Size

- Estimated L, actual L: about 1,990 new Narra lines plus about 450 in existing modules, against the estimated 1,350. The Walker generics and the unsupported paths added length.
