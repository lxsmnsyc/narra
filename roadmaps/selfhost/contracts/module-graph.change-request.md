# Change request: module-graph v1 to v1 (draft fix before freeze)

- From node: modules
- Status: open

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

