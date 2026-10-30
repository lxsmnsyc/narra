# Outcome: hashed-cnames

## Built

All C naming is in `compiler/internal/narra/cnames.go`. `mangle`, `typeArgsMangle`, `tupleCName`, `cModule`, `cSafeSegment`, `checkDeclNames` and `implFuncName` are gone.

- `xxh32(data, seed)`: xxHash32, little-endian words, no dependencies. Tested against reference values (`""`, `"abc"`, the spammish sentence, 15/16/17 bytes, nonzero seed) in `cnames_test.go`.
- `moduleHash(path)`: 8 hex digits over the module path without the package name (`relModule`). A single-file module hashes its whole path.
- `declCName(kind, module, name)`: `f_<hash>_<name>`, `S_...`, `E_...`. Prelude (module "") has no hash. `funcCName(fd)` uses it.
- Interned types, `internedName(kind, head, args)`: `<kind>_<head>_<hash>`, hash over `head|arity|arg ids`. `structCName`, `enumCName`, `arrayCName` (`A_Array_<h>`), `funcTypeCName` (`F_Fn<n>_<h>`), tuples (`S_Tuple<n>_<h>`). `typeID(t)` is a type's C name without a star; `typeArgsKey` is the checker's map key (used for generic instances, arrays, function types, tuples).
- `instanceName(module, name, targs)`: FuncDecl.Name of a generic function instance, `<name>_<hash>`.
- `implName(trait, fn, type)`: FuncDecl.Name for `implement` functions, `<Type>_<Trait>_<fn>_<hash>`.
- `opFuncName(op, t)`: `eq_`, `hash_`, `cmp_`, `show_`, `retain_`, `release_` + type id. Arrays keep `<array>_new/_push/...`.
- `closureCName(kind, owner, n)`: `k_<hash>_<name>_<n>`, `K_...` environment, `D_...` drop. `thunkCName`: `w_<hash>_<name>`.
- `registerCName(name, key)`: process-wide table; panics with "internal error: C name X is used for both ..." on a clash. Every builder above registers.
- `writeCNameTable`: called from `Emit` after the `// Types` header; lines `/* S_Tuple2_<h> = #[int, #[int, int]] */` sorted by name, for generic instances, tuples, arrays and function types.
- Loader: removed `__` and edge-`_` rules for declaration names, module path segments, package names, trait and trait function names. Two modules whose relative paths hash alike give an error.
- Tests: `cnames_test.go` (hash vectors, same C from two directories, type names stable when an unrelated function is added, collision panic). Sample `testdata/packages/cnames_modules/` (module `Option`, `a__b`/`b__c`, `_b`, trait `T_` with `_f`, type `Point_retain`, nested tuples of different shape). Existing `Tuple.nra` stays.
- Change requests: `contracts/module-graph.change-request.md` (v3) and `contracts/trait-model.change-request.md` (v3).

## Deviations

- Removed four error tests that checked the dropped rules: `testdata/errors/traits_func_underscore.nra`, `traits_name_underscore.nra`, `testdata/packages/errors/dunder_name`, `dunder_module` (outside owned paths, required by the brief). Removed the `a__b` case in `modules_test.go`.
- Edited outside owned paths: `compiler/README.md` (naming text), `compiler/internal/narra/CLAUDE.md` (naming rules), `modules_toml.go` (package name rule), `testdata/Tuple.nra` (comment), `check_traits.go`, `check_types.go`, `check_call.go`, `check_closure.go`, `check_patterns.go`, `emit*.go`, `types.go`, `modules.go` (shared).
- Generated C "unchanged apart from names": checked by output and exit code of every sample under debug and release (all pass). No normalized textual diff of C was run.

## Stubs left

- None.

## Gotchas

- Names are computed on every `CName()` call; no cache.
- The table is process-wide. Keys use the module path without the package name, so the same relative module path in two packages in one process is the same name by design.
- Generated helpers put the operation first (`retain_S_...`). Do not build `<type>_retain`.
- A user type named `Box_<8 hex digits>` could equal an instance name; the table stops the compiler with an internal error instead of miscompiling.
- The self-hosted port needs: xxh32 with `*%` and rotates, the fold string format `head|arity|arg|arg`, and the registry.

## Proposed learnings

- Build every C name through `cnames.go` | one place keeps names unique | compiler/internal/narra | yes (the collision table panics)
- Put the operation before the type in generated helper names | suffixes can equal user type names | compiler/internal/narra | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (all samples under ASan/UBSan and the leak check, release builds too).

## Size

- Estimated M, actual M.
