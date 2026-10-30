# Change request: module-graph v1 to v1 (draft fix before freeze)

- From node: modules
- Status: accepted

## Problem

1. The contract names the package module type `Module`. `Module` already names a standard library module in `compiler/internal/narra/builtins.go` (contract `builtins` v1, frozen). Both cannot exist in one Go package.
2. The contract leaves out how a binary's entry function reaches `Emit`, and how a single file gets a module path. Both affect every C name.
3. SPEC.md section 15 writes the entry as `"app/cli.run"`; the contract writes `"cli.run"`.

## Proposed change

- Rename the type to `SourceModule` (fields unchanged: `Path`, `Files`, `Exports`).
- Add to the interface:
  - `func Load(path string) (*Package, error)`: a directory goes to `LoadPackage`; a `.nra` file becomes package `<stem>` with module `"<stem>"` and binary `<stem>` with entry `"<stem>.run"`. So `hello.nra` emits `f_hello__run`.
  - `Program.Entries map[string]*FuncDecl`, by binary name, filled by `CheckPackage` for every binary whose exported entry function exists. The caller passes one to `Emit`.
  - `Program.Files map[string]string`: module path to file, used in panic messages.
  - `FuncDecl.Module`, `TestDecl.Module`, `StructType.Module`, `EnumType.Module`: the declaring module path, "" for the prelude.
- Make the C name mapping injective. `/` and `.` become `_`, so `a/b_c` and `a/b/c` collide, and module `a` with function `b__c` collides with module `a__b` with function `c` (both `f_p_a__b__c`). Rule: a module path segment, after the `_` mapping, must not contain `__` and must not start or end with `_`; a package name follows the same rule; top-level names must not contain `__`. Then the first `__` after the `f_`, `S_` or `E_` prefix always separates the module from the name. Two modules whose mapped paths are still equal (`a-b` and `a_b`) are an error.
- `Program.File` and `Program.Entry` are removed. `Check(f *File)` stays as a wrapper that checks one file as package `main`.
- Every line of a multi-line compile error now starts with its file (`src/text.nra:3:4: ...`). Before, only the first line had the file. `Diagnostic.File` is filled by the checker and by the loader.
- A package may not be named `std`, and `[[binary]]` names must be unique.
- Entry form: accept both. `"cli.run"` resolves to `"<package>/cli"` when that module exists; otherwise the module part is used as a full path, so `"app/cli.run"` also works.

## Affected nodes

json-diagnostics, c-interop, opaque-distinct, build-conditions, query-cli, the selfhost nodes, traits (rebases on modules).

## Workaround in place

Implemented as proposed in `compiler/internal/narra/modules.go` and `modules_check.go`.

## Decision

Accepted 2026-10-28 with PR #4. The contract is now v2, frozen.

---

# Change request: module-graph v2 to v3

- From node: hashed-cnames
- Status: open

## Problem

The v2 C name rule joins module paths, names and type arguments with `__`. It needed rules against `__` and leading or trailing `_` in names, a `0` marker for implementations, and a separate tuple prefix. Three collision bugs came from it.

## Proposed change

Replace the "C names" and "injective mapping" bullets of Layout and names with:

- A declaration's C name is `<kind>_<module hash>_<name>`, with kind `f`, `S` or `E`. The module hash is xxHash32, seed 0, as 8 lowercase hex digits, over the module path relative to the package (`app/text` hashes `text`; a single-file module hashes its whole path). Example: `app/text.show` is `f_<hash of "text">_show`. Prelude names have no hash segment (`E_Ordering`).
- Composite types (generic instances, tuples, function types, arrays) are interned and named `<kind>_<head>_<hash>`. The head is `<module hash>_<name>` for a generic instance (just `<name>` in the prelude), `Tuple<n>`, `Fn<n>`, or `Array`. The hash is xxHash32 over `head|arity|arg1|arg2...`, where each arg is the C name of the argument without a pointer star (a primitive is its Narra name). So `#[int, #[int, int]]` is `S_Tuple2_<h>` built from `int` and the name of `#[int, int]`. Function types list the parameters, then the result, as arguments.
- Generic function instances are named `<name>_<hash>` (FuncDecl.Name), with the hash over `<module hash>_<name>|arity|arg names`. Implementation functions are named `<Type>_<Trait>_<fn>_<hash>`, the hash over type, trait module, trait and function.
- Generated helper functions come first: `retain_<type>`, `release_<type>`, `eq_<type>`, `hash_<type>`, `cmp_<type>`, `show_<type>` (arrays keep `<array>_new`, `_push` and so on). Arrow functions are `k_<hash>_<owner name>_<n>`, with `K` for the environment and `D` for the drop function, and a thunk for a named function is `w_<hash>_<name>`.
- The compiler keeps a table from C name to what it names, and stops with an internal error if two different things get one name.
- The generated C has a comment table right after the `// Types` header, sorted by name: `/* S_Tuple2_<h> = #[int, #[int, int]] */`.
- Removed rules: no `__` or edge `_` restriction on top-level names, module path segments, package names, trait names or trait function names. Modules `a-b` and `a_b` no longer collide; two modules whose relative paths hash alike are an error.

## Affected nodes

modules (done), traits (done), closures, patterns, opaque-distinct, c-interop, every selfhost node that ports the naming.

## Workaround in place

None needed; the node implements this in `compiler/internal/narra/cnames.go`.

## Decision

