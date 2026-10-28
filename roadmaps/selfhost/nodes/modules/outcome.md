# Outcome: modules

## Built

All code in `compiler/internal/narra/`.

- `modules.go`: `Binary`, `LinkConfig`, `SourceModule` (the contract's `Module`; see Deviations), `Package`, `Load(path)`, `LoadPackage(dir)`, module ordering and cycle detection (`orderModules`), `cModule`, `funcCName`, `FormatError`.
- `modules_toml.go`: `parseManifest`, a TOML subset reader (tables, `[[binary]]`, basic and literal strings, string arrays over several lines, comments). It reads `[package] name`, `[[binary]] name, entry`, `[link] libraries, search`. Other tables are skipped; `[dependencies]` is an error.
- `modules_check.go`: `modScope`, `(*Checker).enter`, `CheckPackage`, imports of package modules, `Alias.f(...)` calls (`moduleCall`), `Alias.T` types (`qualifiedType`), entry lookup and signature check.
- `driver.go`: `CompileToC(path, tests)` and `CompileBinary(path, bin, tests)` take a `.nra` file or a package directory. `BuildOptions.Binary` picks the binary.
- `cmd/narra`: `-bin name` on build, run and emit. `narra build <dir>` builds every `[[binary]]` into `./<name>`.
- Tests: `modules_test.go` (`TestPackages`, `TestPackageErrors`, `TestManifest`, `TestEntryForms`). Sample package `compiler/testdata/packages/app` (modules `app/cli`, `app/text`, `app/util/list`, one binary, a test block). Error packages in `compiler/testdata/packages/errors/`: `private_import`, `private_call`, `cycle`, `unknown_module`, `missing_entry`, `ctor_generic_mismatch`, `ctor_wrong_enum`, `struct_generic_mismatch`, `dunder_name`, `dunder_module`, `contains_itself_imported`.

Module graph API (contract module-graph, with the change request):

```go
func Load(path string) (*Package, error)       // dir -> LoadPackage; file -> one-module package
func LoadPackage(dir string) (*Package, error) // reads narra.toml and src/**.nra, orders modules, reports cycles
func CheckPackage(pkg *Package) (*Program, error)
func Emit(prog *Program, entry *FuncDecl, tests bool) string // entry is nil in test mode
```

- `Package.Modules` is in dependency order, imports first. `File.Path` is relative to `Package.Dir` (`src/text.nra`, or the base name for a single file).
- `Program.Entries[binaryName]` is the entry `*FuncDecl` of each binary whose exported entry function exists. `Program.Files[modulePath]` is the file used in panic messages.
- `FuncDecl.Module`, `TestDecl.Module`, `StructType.Module`, `EnumType.Module` hold the declaring module path; "" is the prelude.

C naming scheme:

- `cModule(path)` turns a module path into a prefix: every byte that is not `[A-Za-z0-9_]` becomes `_`, then `__` is appended. The prelude ("") has no prefix.
- Functions: `f_<prefix><name>`, for example `f_app_text__show`. Generic instances: `f_app_util_list__bagOf__int`.
- Structs and enums: `S_<prefix><name>`, `E_<prefix><name>`, plus type arguments: `S_app_util_list__Bag__int`. Arrays follow: `A_E_app_text__Token`.
- Prelude and std names are unchanged: `E_Option__int`, `lstr_*`, `A_int`.
- A single file `hello.nra` is package `hello`, module `"hello"`, binary `hello` with entry `"hello.run"`. So every existing program's C names changed, for example `f_run` became `f_hello__run`. Apart from that, the C for every file in `compiler/testdata/` is byte-identical to main (checked by dumping C before and after and undoing the prefix).
- The mapping is injective: module path segments (after the `_` mapping) and package names cannot contain `__` or start or end with `_`, and top-level names cannot contain `__` (`cSafeSegment`, `checkDeclNames`). So the first `__` after `f_`, `S_` or `E_` separates the module from the name. `LoadPackage` still reports two modules whose mapped paths are equal (`a-b` and `a_b`).

Checker design: one `Checker`. Each module has a `modScope` with its own `types`, `generics`, `funcs`, `modules` (aliases) and `imported` tables; `enter` swaps them into the checker fields, so the rest of the checker code reads `c.types` etc. as before. Modules are checked in dependency order, declarations then bodies. The prelude is checked once as module "" and copied into every scope. A generic function instance or a generic type template is resolved and checked in its declaring module's scope (`pendingFunc.mod`, `genericType.mod`, `Checker.funcMod`). `substitute` finds generics by `(Module, Name)` through `genericOf`.

How a feature node adds a cross-module case:

- A working program: add `compiler/testdata/packages/<name>/` with `narra.toml` (one `[[binary]]`), `src/*.nra`, and `run.out` with the expected stdout. `TestPackages` builds it in debug (ASan, UBSan, leak check) and release mode, and also runs its `test` blocks.
- A compile error: add `compiler/testdata/packages/errors/<name>/` whose `narra.toml` starts with `# error: <text>`. `TestPackageErrors` checks that compiling fails with that text.

## Deviations

- Contract type `Module` is `SourceModule`: `Module` is the std module type of the frozen builtins contract. Change request: `roadmaps/selfhost/contracts/module-graph.change-request.md`, which also asks to add `Load`, `Program.Entries`, `Program.Files` and the `Module` fields.
- `entry` accepts both `"cli.run"` (contract) and `"app/cli.run"` (SPEC.md section 15). The contract form wins when both name a module.
- Error text: every line of a multi-error message now starts with its file (`infer_none.nra:3:3: ...`); before, only the first line did. Single-error messages are unchanged. `Diagnostic.File` is now filled, by the checker (`errorf`) and by `Load` for parse errors.
- `Program.File` and `Program.Entry` are gone; `Check(f *File)` remains as a wrapper that checks one file as package `main`.
- Entry signature error now names the function (`export function start(...)`) and is checked only for binary entries, not for every exported `run`.
- Edits outside owned and shared paths: `compiler/README.md` (usage, supported features, "Differences from the spec"), `compiler/internal/narra/check_expr.go` (one line in `structLit`, at review request), and the first line of `compiler/testdata/errors/contains_itself.nra`, whose pin is now `contains_itself.nra:2:1:` because the error points at the declaration. The change request is the other file outside the node folder. - Names with `__` (top-level declarations, module path segments, package names) are rejected, a package cannot be named `std`, and `[[binary]]` names must be unique. These keep C names injective and imports unambiguous.

## Stubs left

- `Package.Link` is parsed but not passed to the C compiler. c-interop uses it.
- No `.nrai` interface files and no incremental builds (brief: out of scope). One C file per program.
- No `[dependencies]`, `export const`, `export type`, or `X.Token.Word` constructors through a namespace.
- No struct literals through a namespace: `B.Pub { ... }` is a parse error (`expected the end of the statement, found "{"`). Import the struct by name instead.
- An exported function can return a private type, and callers can read its fields (`A.mkPriv().v`). Nothing warns about it; SPEC.md does not say whether that is allowed.

## Gotchas

- Never match a `StructType` or `EnumType` by `Name` alone: two modules can declare the same name. Compare pointers, or `Name` and `Module` (`unify`, `ctor`, `structLit` do). For a name written in the current module, `namesEnum` resolves it and compares. `expectType` prints `p/b.Kind` when both types print the same (`typeName`).
- Anything that resolves a `TypeExpr` or checks a body for a declaration of another module must `enter` that module's scope first, as `funcInstance` and `buildGeneric` do. Otherwise names are looked up in the caller's module.
- Diagnostics take their file from the current scope (`c.mod.file`). An error about a type found after all modules are checked must use `typeErrorf`, which reports at the declaration in the declaring module's file (as `orderTypes` does).
- `Emit` needs the entry from `Program.Entries`; `CompileBinary` reports a missing entry.

## Proposed learnings

- Build C names only through `CName()` and `funcCName`, never from `Name` | names must carry the module | compiler/internal/narra | yes, a package test with the same private name in two modules
- Enter the declaring module's scope before resolving another module's types or bodies | lookups use the current module's tables | compiler/internal/narra | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (all samples under ASan, UBSan and the leak check, plus the new package tests).
- C for every `.nra` in `compiler/testdata/` (programs, panics, errors) dumped before and after: identical after removing the `<stem>__` prefix. Only error texts differ: multi-line errors gain the file prefix, and "contains itself" now has the declaration position.
- CLI by hand: `narra build testdata/packages/app`, `narra run testdata/hello.nra`, `narra run -bin nope <dir>`, `narra emit -bin app <dir>`, `narra test testdata/packages/app`.

## Size

- Estimated L, actual L.
