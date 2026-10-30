# Outcome: selfhost-check-imports

## Built

The cross-module half of the declaration checker in Narra, the std builtin registry as Narra data, and the Go dump lines that compare them. About 930 net Narra lines: three new modules (700 lines, 81 of them the registry table) and 230 lines in existing ones. Go: about 300 lines in `decldump.go`, one new harness.

### Modules (selfhost/src)

| module | lines | ports |
| --- | --- | --- |
| `narrac/scopes` (scopes.nra) | 316 | the import loop of `declareModule` (`importModules`), `useModule`, `importPackageModule`, `exported`, `importTrait`, `importConst`, the imported-constant check of `declareConsts` (`checkImportedConsts`), `typeNamed`, `declareType` and `checkTypeName` (moved from declare.nra), `entryFunc`, `checkEntry` (`checkEntries`), `sourceScope`, `exportsOf`, `declName` |
| `narrac/builtins` (builtins.nra) | 196 | `Builtin` (as `StdFunc`), `Module` (as `StdModule`), the registries (`Registry`), `registerModule`/`registerByName` (`register`, panics on a duplicate name), `moduleList`, `moduleOrder`, the registry dump |
| `narrac/stdlib` (stdlib.nra) | 188 | the data half of every `std_*.go`: `STD_FUNCS`, a top-level `const` table of 81 entries; `stdRegistry()`; `stdType(mutable c, text)` and `typeProblems` |

Changed: `check.nra` (ModScope and Checker fields, `EntryInfo`), `declare.nra` (calls into scopes.nra, registry set up in `newPackageChecker`, `checkEntries` in `declarePackage`), `resolve.nra` (`qualifiedType` ported), `declsdump.nra` (module blocks, entries, qualified type names, loud impl check), `main.nra` (`narrac builtins`).

Import order: types <- builtins <- check <- generics <- scopes <- resolve <- stdlib, traits <- declare. `qualifiedType` stays in resolve.nra because it calls `instantiate` (the resolve/instantiate cycle); resolve.nra imports `exported` and `sourceScope` from scopes.nra. `stdType` needs `instantiate`, so it is in stdlib.nra, and check.nra imports only builtins.nra (it starts with an empty registry; `newPackageChecker` sets `c.registry = stdRegistry()`).

### Module scopes

```
struct ModScope {
  path, file
  isSource: bool                // a package module (Go's src != nil); false for the prelude
  imports: Array<ImportDecl>    // source order, for the dump
  decls: Array<Decl>
  exports: Map<string, int>     // exported name -> index in decls (last one wins, as Go's Exports map)
  types: Map<string, Type>      // own and imported, plus primitives, prelude types, StringBuilder after std/string
  generics: Map<string, int>    // Checker.generics, own and imported
  funcs: Map<string, int>       // Checker.funcs, own and imported
  consts: Map<string, int>      // Checker.consts, own and imported
  traits: Map<string, int>      // Checker.traits, own and imported (prelude traits via lookupTrait)
  modules: Map<string, string>  // namespace alias -> module path (package or std)
  imported: Map<string, int>    // std function imported by name -> Registry.funcs index
}
```

- An imported name maps to the same table index (or the same `Type` value) as in the declaring module, so cross-module identity is index equality. Two modules that both declare `Token` have two `StructInfo` entries.
- New `Checker` fields: `registry: Registry`, `scopeIndex: Map<string, int>` (module path to scope), `pkgName`, `single` (for "unknown module ...; package p has no file"), `used: Array<string>` (std modules in first-import order, Go's `Program.Modules`, for the emitter's runtime C), `entries: Array<EntryInfo>` (`{ binary, func }`, `func` a `Checker.funcs` index or -1, one per `[[binary]]`).
- Package modules are declared in load order, imports first, so an import always finds the imported module's scope and entries. `importModules` runs first in `declareModule`, before type names, as in Go.

### Registry layout

```
struct StdFunc { name ("Str.slice", unique), module ("std/string", "" for print/assert/panic),
                 key ("slice", the import name; `export` is a keyword),
                 params: Array<string>, mutables: Array<bool>, result: string,   // type text
                 effects: Array<string>, callbacks: Array<int>, check: string }
struct StdModule { path, funcs: Map<string, int> }
struct Registry { funcs: Array<StdFunc>, prelude: Map<string, int>, modules: Map<string, StdModule>,
                  order: Array<string> (registration order), byName: Map<string, int> }
```

- `check` is "" for a builtin with a fixed `Params`/`Result` in Go; `"fixed"` for one whose Go Check is an inline closure calling `checkFixed` (the Narra table holds the signature in `params`/`result`); otherwise the Go Check function name, and `params`/`result` are empty: `checkArrayLength` (length, values), `checkArrayPush`, `checkArrayFilled`, `checkBoxNew`, `checkMapSize`, `checkMapRemove`, `checkMapIter` (keys, entries), `checkRangeUpTo`, `checkRangeBetween`, `checkOptionMap`, `checkOptionAndThen`, `checkOptionWithDefault`, `checkOptionOrElse`, `checkOptionIs`, `checkOptionOkOr`, `checkResultMap` (map, mapErr), `checkResultAndThen`, `checkResultWithDefault`, `checkResultToOption` (ok, err), `checkTestCompare`, `checkTestPanics`, `checkTestGolden`.
- Types are text because Option, Result and tuples are instances only a checker can make. `stdType(mutable c, text)` resolves `int`, `f64`, `bool`, `char`, `string`, `void`, `StringBuilder`, `Array<..>`, `Box<..>`, `Map<.., ..>`, `#[..]` and the prelude generics (`Option<..>`, `Result<.., ..>`) through `instantiate`. Call it only where Go's Check would run (at a call site), never in the declaration passes, or new instances change the dump's instance and order lines.
- Registered: the prelude (print, assert, panic) and std/int, std/float, std/string, std/array, std/box, std/char, std/env, std/fs, std/map, std/option (7), std/process, std/range, std/result (6), std/test. `moduleList(r)` gives Go's text for "unknown module" errors.
- Choice: hand-written table, not a generator (81 lines, under the 300-line threshold). `TestSelfhostBuiltins` fails when it differs from Go's registry.

### How bodies resolve names and std calls (selfhost-check-bodies)

- A plain name `x`: local, then `scopes[cur].consts[x]` (Checker.consts; `ConstInfo.t` is the declared type or None), then `scopes[cur].funcs[x]` (Checker.funcs, own or imported), then `scopes[cur].imported[x]` (a std function imported by name), then `registry.prelude[x]`. This is Go's order (local, constant, function).
- `A.f(args)`: `scopes[cur].modules[A]` gives the path. If `sourceScope(c, path) >= 0`, it is a package module: `exported(mutable c, pos, s, f)` reports a private or missing name and returns the `Decl`; a `Func` is `c.scopes[s].funcs[f]`, a `Const` is a value call (Go's `moduleCall`). Otherwise `c.registry.modules[path]` and its `funcs[f]` give the `StdFunc`; when `f` is missing, report Go's error for `A.f`.
- Dispatch on `StdFunc.check`: "" or "fixed" means check the arguments against `stdType` of each `params` entry and return `stdType(result)` (Go's `checkFixed` or the plain `Params` path), using `mutables`; any other value names the Go function to port (`checkArrayPush` and so on). `callbacks` feeds the effect pass, `effects` too.
- `Array.push` without the import (E0908 in Go): the namespace is the part of `StdFunc.name` before the dot, and its module is `StdFunc.module`; Go's `namespaceModules` map is not kept separately.
- `T.f` where T is a qualified type `A.T`: `qualifiedType` in resolve.nra handles types; there is no qualified trait (the contract says so).
- Mark used std modules: `importModules` already calls `useModule`; `c.used` is complete after the declaration passes.

### Dump

- Go `decldump.go` and Narra `declsdump.nra`: a `module "<m>"` block after each package module's declarations, with `import "<path>" package|std [alias=A] [names=a,b]` and `scope <kind> <name> <target>` lines (type, generic, func, const, trait, builtin, alias; sorted). Then one `entry "<binary>" <module>.<fn>|none` line per binary. The format is in the doc comment of `decldump.go`.
- Every type in the decls dump is now written with its module when it is not a prelude type: `q/lib.Token`, `Pair<q/lib.Token, q/main.Token>`, also in single-file dumps (`hello.P`). Prelude types and tuples keep their names.
- `DeclDump` now runs `entryFunc` and `checkEntry` after `flushMapKeys`, so a wrong entry signature is a dump error.
- `builtin` lines: `BuiltinDump()` in Go and `narrac builtins` in Narra (registry dump). They are not part of `narra decls`, so goldens do not change when a std function is added.
- `dumpImpl` in declsdump.nra panics on a position or order mismatch, and `dump` panics if a recorded implement block was not printed (the #32 review note).

### Tests

- `TestSelfhostImports` (selfhost_imports_test.go): the whole dump, every line kind, on `declInputs` plus the packages of `testdata/declsdump/imports/`. 271 inputs compared, 30 skipped, all "does not load" (parse and narra.toml errors, the import cycle, `std` as a package name; TestSelfhostModules covers them). Error formats accepted: decls + traits + `portedImportErrors`.
- `TestDeclDumpImports`: goldens in `testdata/declsdump/imports/`: `qualified`, `alias_clash`, `generic_imports`, `impl_results.nra`, and `err_unknown_module`, `err_unknown_export`, `err_private`, `err_dup_names`, `err_const_type`, `err_qualified`, `err_entry_signature`.
- `TestSelfhostBuiltins` (+ `TestSelfhostBuiltinsProbeFails`), `TestStdModuleOrder`.
- The "imports a module of its package" skip is removed from `TestSelfhostDecls` and `TestSelfhostTraits`; they now compare selfhost/ and every multi-module package too (225 and 246 compared). Their remaining skips are load failures and error formats outside their own lists (import errors are compared by TestSelfhostImports).

### Coordinator item: instances in impl method and function-type results

Already fixed on main by #33 (parser). `testdata/trait_method_generic_return.nra` (an input of the harnesses) and the new `declsdump/imports/impl_results.nra` (`Option<Array<int>>`, `(int) => Option<int>`, `Option<Self>`, `#[Self, int]`, `(int) => #[int, bool]`) match Go. No checker change was needed.

## Deviations

- Files outside the owned and shared paths: `decldump_test.go` (three expected strings now carry the module), `selfhost_decls_test.go` and `selfhost_traits_test.go` (the import skip removed, as the task asked), `selfhost/src/resolve.nra` (`qualifiedType`, its HOOK from selfhost-check-decls), the goldens `testdata/declsdump/*.dump` and `testdata/declsdump/traits/*.dump` (regenerated for module-qualified types; only type text changed).
- `builtin` lines live in a separate dump, not in `narra decls`, and the 12 inline-closure signatures are checked by compiling probes instead of being printed by Go (Go keeps them inside closures).
- The brief listed module, import and builtin; the dump also has `scope` (the scope's names) and `entry` lines.
- Go-side change: `decldump.go` only (plus tests). `narra emit`, `narra check` and `narra check --json` are byte-identical to a main build on 280 inputs.
- No change requests, no decision records, no diagnostic codes.

## Stubs left

- `stdType` is used only by `narrac builtins` (`typeProblems`). Bodies call it (selfhost-check-bodies).
- The named check kinds are data only; their Narra ports are selfhost-check-bodies.
- `c.used` is recorded but nothing reads it yet (selfhost-emit).
- Module C names and the xxHash32 collision check: selfhost-emit.

## Gotchas

- `export` is a keyword: the registry field is `key`.
- `stdType` creates instances. Calling it during the declaration passes changes the dump.
- A Go `init` order change (a new `std_*.go` file) changes `moduleList`; the Narra table must list modules in the same order. `TestSelfhostBuiltins` catches it, and the table order is the order of `BuiltinDump`.
- A struct with field defaults works in a top-level `const` table, so a new registry entry is one `StdFunc { ... }` line with only the fields it needs.
- An imported function or constant whose exporting module has a duplicate declaration of the same name can resolve to a different entry than Go (Go uses the last exported decl, Narra the scope's entry). It only happens in programs that already have a "declared twice" error.
- Go's `checkEntry` errors are reported in the entry function's module; enter `c.funcs[k].scope` first (done in `checkEntries`).

## Proposed learnings

- Keep fixed signature data in the Narra registry as type text and resolve it at the call site | Option, Result and tuple types are instances that only a checker can make, and making them early changes instance order | selfhost/ | yes (TestSelfhostBuiltins, typeProblems)
- Register a std function in `stdlib.nra` in the same change that adds it in Go | `TestSelfhostBuiltins` fails otherwise | selfhost/, compiler/internal/narra | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (internal/narra 96 s), on 9a4b2b6 + this branch.
- `go test ./internal/narra -run 'TestDeclDump|TestSelfhostDecls|TestSelfhostImports|TestSelfhostTraits|TestSelfhostBuiltins|TestStdModuleOrder'`: pass.
- emit/check comparison: 280 inputs (every `.nra` under `compiler/testdata` outside `packages/` and `pkgdump/`, every directory with `narra.toml`, `selfhost/`) x `emit`, `check`, `check --json`: 0 differences against a main build.
- Timing, `decls selfhost/`: narrac debug 0.75 s wall (0.32 s user), release 0.47 s (0.15 s user), Go `narra decls` 0.04 s. A generated package scales linearly: 20 modules x 100 imported names 0.06 s, 40 x 100 0.14 s, 40 x 200 0.29 s (Go 0.10 s).

## Size

- Estimated L, actual L (about 930 net Narra lines, 300 Go dump lines). The registry stayed small enough to hand-write.
