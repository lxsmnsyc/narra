# Outcome: selfhost-check-traits

## Built

The trait half of the declaration checker in Narra, one Go dump extension and one harness. About 1050 Narra lines (`traits.nra` 831; the rest are additions to `types.nra`, `check.nra`, `declsdump.nra`), 110 Go lines in `decldump.go`.

### Modules and functions (selfhost/src/traits.nra, module `narrac/traits`)

Ports of `check_traits.go` and the Map key half of `std_map.go`. Import order: types <- check <- generics <- resolve <- traits <- declare; declsdump imports traits.

| Narra | Go | notes |
| --- | --- | --- |
| `declareTraits(mutable c, f)` | `declareTraits` | trait names, `implements` and `derives` lists; called in `declareModule` after type names |
| `funcBounds(mutable c, fd, tps)` | `funcBounds` | called in `declareFunc` before the parameters are resolved |
| `resolveTraits(mutable c, f)` | `resolveTraits`, `resolveTraitDecl`, `resolveImpl` | called last in `declareModule`; also fills `paramTypes` for trait and impl function parameters |
| `typeImplements(mutable c, t, tr): bool` | `implements` | `implements` is a keyword |
| `implFunc(c, t, tr, name): Option<ImplRef>` | `implFunc` | None means the compiler writes it |
| `checkBounds(mutable c, pos, fd, tps, targs)` | `checkBounds` | tps are `Checker.params` indexes |
| `checkEquatable(mutable c, pos, t)` | `checkEquatable` | `containsFunc(c, t, mutable seen)` too |
| `fieldsOf(c, t)`, `typeExprString(te)`, `findTraitFunc(c, tr, name): int`, `typeTraitsOf(mutable c, t): int` | `fieldsOf`, `typeExprString`, loop over `tr.Funcs`, `c.traits.types[typeKeyOf(t)]` | |
| `mapKeyProblem(mutable c, t, mutable seen)`, `flushMapKeys(mutable c)` | `mapKeyProblem`, `flushMapKeys` | `declarePackage` flushes after the constant types, before `finish` |
| `lookupTrait(c, name): int`, `stdTrait(c, name): int`, `checkMapKey(mutable c, pos, key)`, `setParamType` (moved here) | same | in check.nra; `-1` means none |

`sameSignature` and `signatureText` are private. `substitute` runs in Go's order (result first, then each parameter), because it can create instances.

### Table layout

In `types.nra`:

```
struct TraitInfo   { name, module, std, decl: TraitDecl, selfParam: int (Checker.params), funcs: Array<TraitFunc> }
struct TraitFunc   { name, decl: FuncDecl, params: Array<Type>, result: Type }   // Self is TParam(selfParam)
struct TypeTraits  { name, module, pos, generic, implExprs, derivExprs: Array<TypeExpr>,
                     listed: Map<int,bool>,   // Go's implements, keyed by trait index
                     derived: Map<int,bool>,  // Go's derives
                     blocks: Map<int,int> }   // Go's impls: trait index -> Checker.impls index
struct ImplInfo    { decl: ImplDecl, scope: int, traitIdx: int, t: Type, funcs: Array<ImplFunc> }
struct ImplFunc    { name, decl: FuncDecl, result: Type }     // name as written; Go renames to implName
struct ImplRef     { block: int, index: int }                 // Checker.impls[block].funcs[index]
struct MapKeyCheck { pos, key, file }
```

In `Checker` (check.nra): `traits: Array<TraitInfo>`, `typeTraits: Array<TypeTraits>`, `typeTraitIndex: Map<string,int>` keyed `"<module>|<name>"` (Go's `typeKey`; an instance shares its template's entry; tuples use `"|Tuple"`, made on first use), `bounds: Map<int, Array<int>>` (param index to trait indexes), `busy: Map<string,bool>`, `impls: Array<ImplInfo>`, `implFuncs: Map<string, ImplRef>` keyed `"<typeKey(t)>|<Trait>.<function>"`, `mapKeys: Array<MapKeyCheck>`, `errKeys`. `ModScope.traits: Map<string,int>` maps names visible in a module to trait indexes. Scope 0 (prelude) traits are found by `lookupTrait` from every module.

Parameter types of trait and impl functions are in `paramTypes` by `Param.id`. Result types are in `TraitFunc.result` and `ImplFunc.result`.

### What selfhost-check-bodies calls

- `findTraitFunc(c, tr, name)` and `c.traits[tr].funcs[k]` for `Trait.function(args)` (Go's `traitCall`), then `typeImplements`, `implFunc` (Some: call `c.impls[ref.block].funcs[ref.index].decl`; None: the compiler provides it) and `stdTrait`.
- `c.impls` for the bodies of impl functions (Go's `implBodies`): check each `ImplInfo.funcs[k].decl` with env `Self` = `ImplInfo.t`. Only blocks that passed their checks are in `c.impls`.
- `checkBounds` at generic call sites, `checkEquatable` for `==` and `!=`, `checkMapKey` for Map types and literals in bodies. `flushMapKeys` already runs in `declarePackage`; the full checker calls it after all bodies, as Go does.
- Bodies of generic functions: `c.bounds[param]` gives the constraint of a type parameter.
- `holds=` lines exist only for the types of the `order` lines (concrete structs and enums, and instances without type parameters). Bodies must call `typeImplements` on any type, instances included, and never read the dump for it.

### What selfhost-check-imports does

- `importTrait`: add the exported trait of another module to `c.scopes[s].traits[name]`, reporting `"<name>" is declared twice` when the name is taken by a trait, type or generic. Imports must run before `declareTraits` of the importing module, as in Go.
- `ownsType` already compares `TypeTraits.module` with the current module path, so cross-module orphan impls work once imports exist.
- Remove the skip "imports a module of its package" in `traitSkip` (`selfhost_traits_test.go`).

### Dump

New line kinds `trait`, `impl` and `bound`, plus the `function` sub-line, defined in the doc comment of `decldump.go` and printed by `declsdump.nra`. `impl` has four modes: `listed` and `derived` (the `implements` and `derives` lists, right after the type), `block` (an `implement` block, with its functions) and `holds=<bool>` (the answer of `implements` for Equatable, Hashable, Ordered and Show for each type of the `order` lines). `bound <T> <Trait>` sits under a generic `func`. The Go dump now calls `flushMapKeys`, so Map key errors in signatures appear in it.

Goldens: all `testdata/declsdump/*.dump` gained the prelude trait lines. New cases in `testdata/declsdump/traits/` (`TestDeclDumpTraits`): `basic`, `holds`, `map_keys`, and error cases `err_derive`, `err_duplicates`, `err_impl_target`, `err_impl_twice_fn`, `err_map_float`, `err_missing_fn`, `err_partial_hash`, `err_signature`, `err_trait_decl`, `err_unknown_trait`.

### Harness

`TestSelfhostTraits` (selfhost_traits_test.go) compares the whole `narrac decls` output with `narra decls` (stdout, or stderr and exit 1 for errors) on every input of `declInputs`. It accepts the declaration error formats plus the trait formats (`portedTraitErrors`). 188 inputs compared, 61 skipped. `dump` is linear in the number of `implement` blocks and sorts the instance blocks with a merge sort (N=4000 blocks: 3.2 s before, 0.9 s after; Go 0.07 s). `TestSelfhostDecls` is unchanged and still passes (bound and trait lines are not in its `declKinds`; the `param` and `result` lines of trait functions are).

### Also done

`errorIn` keeps a `Map<string,bool>` (`Checker.errKeys`) keyed by file, line, column and message, instead of rescanning `errs` (the note from the selfhost-check-decls review).

## Deviations

- Files edited outside the owned paths and the brief's `shared` list: `selfhost/src/resolve.nra` (one call `checkMapKey(mutable c, te.pos, k)` in the Map case, the import, two comments). The decls outcome named this hook; the file is not in the shared list.
- `selfhost_test.go` is not edited by this branch (it changed on main through literal-type-args). `selfhost_decls_test.go` is unchanged; the trait error formats are in the new harness only.
- `declare.nra`: `setParamType` moved to check.nra (import cycle: traits.nra needs it, declare.nra imports traits.nra).
- `check_traits.go` and `std_map.go` are not edited. Go-side change: `decldump.go` only. `narra emit` and `narra check` do not call it, so their output is unchanged (git diff shows no other Go file touched).
- Tuples: Go records the `Tuple` trait facts in `tupleOf`; here `typeTraitsOf` makes them on first use. Same answers, so `resolve.nra` needs no hook.
- No change requests, no new diagnostic codes.

### Skipped inputs in TestSelfhostTraits

188 inputs are compared and 61 skipped. The test logs each reason.

- Does not load (30): the same inputs as in the selfhost-check-decls outcome.
- Imports a module of its own package (28): left for selfhost-check-imports, with `importTrait`. Includes `selfhost/`, `packages/app`, `packages/traits`, `packages/errors/trait_orphan_impl` and the others listed in the selfhost-check-decls outcome. imports must remove this skip.
- Error from an unported stage (3): `unknown module "std/nope"` and `unknown module "other/mod"` (the std registry) and `module "std/int" has no export "x"` (imports). imports must remove this skip by adding those formats to the ported list.

## Stubs left

- `traitCall`, `implBodies` and the call-site `checkBounds` use: selfhost-check-bodies.
- `importTrait`: selfhost-check-imports.
- Implementation function names are not mangled (`implName` is a C name, selfhost-emit); `ImplFunc.name` is the source name.
- `qualifiedType` is still the stub; trait names are never qualified (the contract says qualified trait names do not work yet).

## Gotchas

- `implements`, `derives`, `trait` and `implement` are keywords in Narra: they cannot be field or function names. The port uses `typeImplements`, `listed`, `derived`, `blocks`, `traitIdx`.
- `errorf(mutable c, pos, "...${c.traits[tr].name}...")` is rejected (the message uses `c`). Build the message into a local first.
- `Map` has no `??`; use `m[k] is Some(x)` with a small helper (`typeTraitsAt`).
- `typeImplements` is `mutable c` because of the `busy` set and the lazy tuple entry. `dump` takes `mutable c` for the `holds` lines.
- Substitution order matters: `sameSignature` substitutes the result before the parameters and `signatureText` the parameters before the result, as Go does. `substitute` can create instances, which changes the `order` lines.
- A duplicate type declaration plus a failing `derives` can differ: Go uses the declaration's own type, Narra looks the name up in the scope (the later one). Only in programs that already have a duplicate-type error.
- `Equatable` and `Hashable` on a function type are only reported when a derive, a bound or a Map key needs them; `==` on functions is a body error (`checkEquatable`, ported, not called yet).

## Proposed learnings

- Name Narra functions and fields so they avoid the keywords `implements`, `derives`, `trait`, `implement` | the parser rejects them as names | selfhost/ | no
- Build an error message that reads the checker into a local before `errorf(mutable c, ...)` | the mutable alias rule rejects `c` in another argument | selfhost/ | no (the compiler enforces it)

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 259 s).
- `go test ./internal/narra -run 'TestDeclDump|TestSelfhostDecls|TestSelfhostTraits|TestSelfhostPrelude|TestSelfhostNodeIds'`: pass after the `errorIn` change.
- Emit and check output: only `decldump.go` changed among Go files, and only the hidden `decls` command uses it.

## Size

- Estimated M (about 850 Narra lines), actual M-L: about 1050 Narra lines including the dump, tables and `errorIn` change; no rework of the Go checker was needed.
