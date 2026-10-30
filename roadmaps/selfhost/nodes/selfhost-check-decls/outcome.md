# Outcome: selfhost-check-decls

## Built

The whole brief, including `generics.nra` and `orderTypes` (no scope left over). 1788 Narra lines in seven new modules, a 251-line Go dump, and the `CheckPackage` split.

### Modules (selfhost/src)

| module | lines | ports |
| --- | --- | --- |
| `narrac/types` (types.nra) | 159 | `Type`, `Field`, `Variant`, the table entry structs, `typeKey`, `isBoxedEnum` |
| `narrac/check` (check.nra) | 330 | `Checker`, `ModScope`, `FuncInfo`, `ConstInfo`, `CheckError`, `errorf` (Go's `report` dedup), `newScope`, `enter`, `typeString`, `typeModule`, `containsTypeParam`, `typeErrorf`, `finish` + `orderTypes` (`visitOrder`) |
| `narrac/generics` (generics.nra) | 224 | `makeParams`, `envOf`, `envWith` (Go's `envOf` with args), `unify`, `templateEdges`, `growsArgs`, `genericEdges`, `checkPolyRecursion` |
| `narrac/resolve` (resolve.nra) | 337 | `resolveType`, `resolveOpt`, `funcResult`, `qualifiedType` (stub), `userType` (Go's `userBox`/`userMap`), `tupleType`, `tupleOf`, `resolveFields`, `resolveVariants`, `buildGeneric`, `instantiate`, `substitute`, `substituteAll` |
| `narrac/declare` (declare.nra) | 317 | `PRELUDE`, `newPackageChecker`, `declareSourceModule`, `declareModule`, `declareType`, `declareGeneric`, `resolveDistinct`, `declareFunc`, `setParamType`, `declareConsts`, `declareConstTypes`, `declarePackage` |
| `narrac/declsdump` (declsdump.nra) | 168 | `dump(c)`: the `narra decls` format |
| `narrac/idcheck` (idcheck.nra) | 253 | `collectIds(file)`, `checkIdList(ids, first, end)` |

`instantiate`/`substitute` live in resolve.nra, not generics.nra: `resolveType -> instantiate -> buildGeneric -> resolveFields -> resolveType` is one cycle, and Narra modules cannot import in a cycle. generics.nra holds the parts outside the cycle. Import order: types <- check <- generics <- resolve <- declare; declsdump imports check; main imports all.

### The `Type` model (D63)

```
enum Type derives Equatable {
  TInt TF64 TBool TChar TStr TVoid TStringBuilder
  TStruct(index)   // Checker.structs
  TEnum(index)     // Checker.enums
  TDistinct(index) // Checker.distincts
  TParam(index)    // Checker.params
  TArray(elem) TBox(elem) TMap(key, val) TFunc(params: Array<Type>, result)
  TRange TModule(path)   // body checker only; declared now so later matches stay stable
}
```

- Two types are the same exactly when they are `==`, which is Go's `sameType`. Every struct, enum, distinct type and type parameter is an index into a table, and every generic instance and tuple gets its own entry. So `==` on indexes is identity by module and name, never by name alone. Composite types compare structurally.
- `typeString(c, t)` prints Go's `Type.String()` text. `typeKey(t)` is an injective text key (`S3`, `A(int)`, `F(int,S2;void)`) for maps keyed by type. Go keys instances by C name (`typeID`, xxHash32); the Narra key has no hash because C names belong to selfhost-emit. The dump never prints the key.
- Go's `Variant.Enum` back-pointer is not kept, since the enum index is known wherever a variant is read. `Field.default` is the AST `Option<Expr>`. `Field.pos` is the field declaration, used by later error messages.

### Declaration tables (in `Checker`)

| field | entry | notes |
| --- | --- | --- |
| `structs: Array<StructInfo>` | `{ name, module, scope, pos, opaque, tuple, args, generic, fields }` | declared struct (`generic` -1, `args` []), instance (`generic` = template index), or tuple (`tuple`, name `Tuple`, module "", fields `0`, `1`, ...) |
| `enums: Array<EnumInfo>` | `{ name, module, scope, pos, args, generic, variants }` | same scheme |
| `distincts: Array<DistinctInfo>` | `{ name, module, opaque, base: Option<Type> }` | `base` is None until `resolveDistinct`; that is how "defined in terms of itself" is found, including forward references (Go behavior) |
| `params: Array<ParamInfo>` | `{ name }` | one entry per type parameter per declaration (Go's `typeParamID`); a substitution is `Map<int, Type>` keyed by this index |
| `generics: Array<GenericInfo>` | `{ name, module, scope, pos, params, isStruct, opaque, fieldDecls, variantDecls, built, fields, variants }` | Go's `genericType`; `fields`/`variants` are the template, empty until `buildGeneric` and emptied again by a failed `checkPolyRecursion` |
| `funcs: Array<FuncInfo>` | `{ name, module, scope, decl, typeParams, result }` | parameter types are in the side table, not here |
| `consts: Array<ConstInfo>` | `{ name, module, scope, decl, t: Option<Type> }` | `t` is the declared type (set by `declareConstTypes`); None means unannotated, for the body checker to infer |
| `instances: Map<string, Type>` | key `"<generic index>|<typeKeys(args)>"` or `"T|<typeKeys(elems)>"` | instance and tuple cache |
| `instTypes: Array<Type>` | creation order | Go's `instTypes`; an instance joins after its fields are substituted, as in Go |
| `concrete: Array<Type>` | non-generic structs and enums, declaration order, prelude first | what Go passes to `finish` |
| `order: Array<Type>` | set by `finish` | Go's `Program.Types`, the C definition order |

`pos` in StructInfo/EnumInfo is the zero position in the prelude, and the generic's declaration for an instance. This is what Go's `typeErrorf` finds. `scope` is the declaring module's scope index, so `typeErrorf` reports in that file.

### Checker context

```
struct Checker {
  structs enums distincts params generics funcs consts   // tables above
  instances instTypes concrete
  scopes: Array<ModScope>   cur: int                    // Go's mods / c.mod
  env: Map<string, Type>                                // type parameters in scope
  paramTypes: Array<Type>                               // side table by Param.id
  localN: int                                           // kept for the body checker
  errs: Array<CheckError>                               // { file, pos, message }
  order: Array<Type>
}
struct ModScope { path, file, decls: Array<Decl>, types: Map<string, Type>,
                  generics: Map<string, int>, funcs: Map<string, int>, consts: Map<string, int> }
```

- Every pass takes `mutable c: Checker`. Tables are written in place (`c.structs[i].fields = fields`). Build a table entry into a local before `Array.push(mutable c.structs, info)`, because `Array.push(mutable c.x, Info { ... modPath(c) ... })` breaks the mutable alias rule.
- `scopes[0]` is the prelude (path "", file "prelude"). `newScope` seeds every module with the primitive names (`int`, `i64`, `f64`, `bool`, `char`, `string`) and the prelude's types and generics. `enter(c, s)` returns the previous index. Code that resolves another module's types must enter that module first; `buildGeneric` does.
- Fields later nodes extend: selfhost-check-traits adds trait tables (trait list, type trait facts keyed by `"<module>|<name>"`, bounds by param index, impls) and the Map key queue to `Checker`. selfhost-check-imports adds imported names, aliases and the registry to `ModScope`. selfhost-check-bodies adds locals and scopes, and uses `localN`.

### Side tables

- `paramTypes: Array<Type>`, indexed by `Param.id`, TVoid until set. It is filled for top-level function parameters. Trait and impl function parameters are left to selfhost-check-traits; arrow parameters are left to bodies.
- Size: `Array.filled(TVoid, preludeNextId)`. The package's files use ids `[0, Package.nextId)`, and the prelude is parsed after them with `firstId = Package.nextId`, so its ids never clash. The tables therefore hold `[0, prelude nextId)`. `setParamType` ignores ids outside the range (dummy id -1 cannot reach the checker, but the guard is there).

### Driver

`declarePackage(pkg)` = `newPackageChecker` (prelude) + `declareSourceModule` per module in load order + `declareConstTypes` + `finish`. The full checker must instead interleave as Go does: `declareSourceModule(m)`, then m's bodies, for each module, then `finish`. `declareModule` follows Go's `declareModule` step by step. Imports only record `std/string`, which adds `StringBuilder` to the scope.

### Dump

- Go: `DeclDump(path)` in `compiler/internal/narra/decldump.go`, hidden command `narra decls <path>`. The format is defined in the doc comment at the top of that file. Narra: `narrac decls <path>` (`declsdump.nra`). Errors go to stderr, exit 1.
- Line kinds of this node: `type` (struct, enum, distinct, tuple) with `field`, `variant`, `base` sub-lines; `generic`; `func` with `param` and `result`; `const`; `order`; `error`. Attributes come first and the type text comes last, because types contain spaces. Module declarations come in load order (prelude first) and in source order within a module. Then the instance blocks, sorted by their first line. Then the `order` lines. A failed run prints only the errors.
- Annotated constants: the dump resolves the declared type after all modules are declared. Go does this in `constType` (bodies); `DeclDump` does it explicitly, and the Narra side stores it in `ConstInfo.t`.
- Goldens: `compiler/testdata/declsdump/<case>.nra` or `<case>/` (a package), with `<case>.dump`. `TestDeclDump` compares them (`NARRA_UPDATE=1` rewrites). Subdirectories without `narra.toml` (`traits/`, `imports/`) are ignored, so later nodes can add their own tests there.

### Extension points

- Comments marked `HOOK <node>` in declare.nra and resolve.nra mark where each later node plugs in: `declareTraits(f)` after the name pass, `funcBounds` in `declareFunc`, `resolveTraits(f)` after `buildGeneric`, a trait name counting as a type name in `declareConsts`, `checkMapKey` in `resolveType`'s Map case, tuple trait facts in `tupleOf` (traits); the import loop in `declareModule`, the imported-constant check in `declareConsts`, `qualifiedType` (imports); tests, `checkDefaults` (bodies).
- Dump: add a `case` in `dumpModule` (both Go and Narra) for trait and impl declarations, and new kinds to the doc comment's Owners list. The harness compares only `declKinds` (in `selfhost_decls_test.go`); a new node adds its own kinds set and its own harness.
- Tests: extend `portedDeclErrors` in `selfhost_decls_test.go` when a node ports more error formats, or keep a separate list in its own harness.

### Go refactor

`modules_check.go`: `CheckPackage` now calls `newPackageChecker(pkg)` (parses the prelude, makes the Checker, declares the prelude) and, for each module, `c.declareSourceModule(m)` (new scope, `prog.Files`, `declareModule` per file), followed by that module's bodies as before. The per-module interleaving of declarations and bodies is unchanged, so no output changes. `check.go`: `Checker.prelude *File` keeps the parsed prelude for the dump. `cmd/narra/main.go`: hidden `decls` command.

Proof: a script runs `narra emit`, `narra check` and `narra check --json` on every `.nra` under `compiler/testdata` outside `packages/` and `pkgdump/`, every directory with `narra.toml`, and `selfhost/`. It saves stdout, stderr and the exit code. A binary built from `origin/main` (2f52283, after #24) and one built from this branch give byte-identical results on all 219 inputs (`diff -r` is empty). The same check on c8e17a7 before the rebase also matched (187 inputs).

### Id walker

`idcheck.nra` visits every `Expr`, `Pattern`, `Param`, `VarStmt` and `ForStmt` of every declaration (function and trait/impl function parameters and bodies, constant values, field defaults, tests). `narrac pkg --ids <path>` runs it on each module after `checkIds` and still prints `ids ok <nextId>`. `narrac pkg --ids-dup <path>` copies the first collected id over the second before checking, so the check must fail. `TestSelfhostNodeIds` runs `--ids` on every loadable input of `TestSelfhostModuleIds` (294 inputs), and its `dup` subtest requires `--ids-dup` to fail with "is used twice".

## Deviations

- Files edited outside the owned paths: none. Shared files edited: `check.go` (one field), `modules_check.go` (split), `cmd/narra/main.go` (`decls`), `selfhost/src/main.nra` (`decls`, `--ids`, `--ids-dup`), `selfhost/src/modules.nra` (`quoteDump` is now exported). `selfhost_test.go` and `compiler/README.md` are unchanged.
- `selfhost/` as one package is skipped in `TestSelfhostDecls`, against the brief's input list. Its modules import each other's types, which needs module scopes from selfhost-check-imports. Every multi-module package whose modules import each other is skipped for the same reason; `testdata/declsdump/two_modules` covers two modules without imports.
- Instance cache keys are the Narra `typeKey` text, not Go's C-name text (see above).
- Error codes, fixes and JSON are not ported. `CheckError` has file, position and message, and dedup compares those three (Go also compares the code, which follows from the format). Codes go with the driver in selfhost-bootstrap.
- `std/test` (merged in #24) is a new registered std module. Registry work is selfhost-check-imports'; nothing here depends on it.
- No change requests.

### Skipped inputs in TestSelfhostDecls

144 inputs are compared and 69 skipped. The test logs each skip with its reason:

- Does not load (30). TestSelfhostModules covers these. Parse errors: `astdump/error.nra`, `diagnostics/ts_array_type.nra`, `ts_import_default.nra`, `ts_null.nra`, `errors/patterns_shorthand.nra`, `patterns_spread_order.nra`, `semicolon.nra`, `strict_eq.nra`. Load errors: `packages/errors/cycle` and the pkgdump cases `array_bad`, `array_trailing`, `bad_escape`, `bad_name`, `bad_toml`, `binary_no_entry`, `binary_not_string`, `dependencies`, `dup_binary`, `import_cycle`, `link_bad`, `missing_bracket`, `missing_bracket2`, `name_not_string`, `no_name`, `no_sources`, `parse_errors`, `self_import`, `std_name`, `string_trailing`, `unterminated`.
- Imports a module of its own package (28), left for selfhost-check-imports: `selfhost/`, `packages/app`, `closures`, `cnames_modules`, `const`, `opaque_range`, `patterns`, `traits`, `pkgdump/nested_order`, and `packages/errors/` `closure_escape`, `contains_itself_imported`, `ctor_generic_mismatch`, `ctor_imported_ambiguous`, `ctor_wrong_enum`, `distinct_arith`, `distinct_int`, `distinct_mix`, `opaque_assign`, `opaque_distinct_from`, `opaque_literal`, `opaque_pattern`, `opaque_read`, `opaque_spread`, `private_call`, `private_import`, `struct_generic_mismatch`, `trait_orphan_impl`, `unknown_module`.
- An error from a stage this node does not port (11). Trait errors: `errors/traits_derive_field`, `traits_dup_bound`, `traits_foreign_impl`, `traits_foreign_option`, `traits_missing_impl`, `traits_not_listed`, `traits_signature`, `closures_trait_escaping`. Import errors: `errors/unknown_module`, `astdump/decls.nra`, `pkgdump/export_kinds`.

## Stubs left

- `qualifiedType` always reports `unknown module "<alias>" in type ...`. That is Go's text for an alias that is not a package module. selfhost-check-imports replaces it.
- No trait work: trait declarations, `derives`, `implements`, bounds, impl signatures and Map key checks (HOOKs, selfhost-check-traits). A constant named like a prelude trait is a Go error that the Narra side does not report yet.
- No `checkDefaults` (field default expressions) and no tests list (selfhost-check-bodies).
- Go's `finish` also lists arrays, boxes, maps and function types for the emitter. That is not ported; Narra composite types are values and the emitter can collect them (selfhost-emit). No C names (selfhost-emit).

## Gotchas

- Found while writing cases, not fixed (not this node's file): the parser drops the line break after a type that ends in `>`. So `distinct type Ids = Array<int>` followed by any declaration fails with "expected the end of the statement". Go and Narra agree. `testdata/declsdump/distinct.nra` puts such a distinct type last. A parser fix belongs to its own PR.
- Go's `checkDefaults` runs `c.expr` on field defaults during the declaration pass. On valid programs that creates no instances in the samples. A default that names a generic instance only in its expression would put an instance in Go's dump that Narra lacks; the body port must run defaults at the same point (HOOK).
- An instance made while its own template is being built (a self-reference) gets empty fields, because `built` is set before the template is resolved. This matches Go and is why `List<T>` in `Array<List<T>>` terminates. Keep this order.
- `instTypes` order must match Go exactly for the `order` lines: substitute the args first, then instantiate, and push an instance after its fields.
- Narra syntax met: `match` arms cannot be `=> return x` (use `return match (t) { ... }` or a block); `for (const #[a, b] of ...)` is not supported (destructure inside the loop); `Array.filled(value, n)` takes the value first.
- The prelude text is duplicated in `declare.nra` (`PRELUDE`). `TestSelfhostPrelude` fails when it differs from `preludeSrc`.
- `TestSelfhostDecls` judges skips by error format (`portedDeclErrors`). A ported format coming from an unported stage (for example "unknown type" inside an impl signature) would show as a mismatch, not a skip. None occurs today.

## Proposed learnings

- Put mutually recursive passes (resolve, instantiate, build) in one Narra module | modules cannot import in a cycle | selfhost/ | no
- Build a table entry into a local before pushing it onto a field of a mutable context | `Array.push(mutable c.x, f(c))` breaks the mutable alias rule | selfhost/ | no (the compiler enforces it)

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 76 s), on main 2f52283 after the rebase.
- `go test ./internal/narra -run 'TestDeclDump|TestSelfhostDecls|TestSelfhostNodeIds|TestSelfhostParser|TestSelfhostModules'` (the brief's verify): pass.
- Byte-identical emit/check comparison: 219 inputs, no difference.

## Size

- Estimated L, actual L. 1788 Narra lines (1535 without the id walker) for about 1050 Go lines (1.46x), plus a 251-line Go dump and the harnesses.
