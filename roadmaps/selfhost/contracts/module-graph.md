---
name: module-graph
version: 3
status: frozen
defined_by: modules
---

# Contract: module-graph

## Purpose

How a package's files become modules, how names stay unique in the generated C, and what `narra.toml` holds. Used by c-interop (`[link]`), opaque-distinct (visibility across modules), build-conditions, query-cli and the self-hosted compiler.

## Interface

Layout and names:

- A package is a directory with `narra.toml` and `src/`. A file `src/<path>.nra` is module `"<package>/<path>"`. Example: package `app`, file `src/text.nra`, module `"app/text"`. There is no directory named after the package inside `src/`.
- A declaration's qualified name is `<module>.<name>`, such as `app/text.helper`.
- A declaration's C name is `<kind>_<module hash>_<name>`, with kind `f`, `S` or `E`. The module hash is xxHash32, seed 0, as 8 lowercase hex digits, over the module path relative to the package (`app/text` hashes `text`; a single-file module hashes its whole path). Example: `app/text.helper` is `f_<hash of "text">_helper`. Prelude names have no hash segment (`E_Ordering`).
- Composite types (generic instances, tuples, function types, arrays) are interned and named `<kind>_<head>_<hash>`. The head is `<module hash>_<name>` for a generic instance (just `<name>` in the prelude), `Tuple<n>`, `Fn<n>`, or `Array`. The hash is xxHash32 over `head|arity|arg1|arg2...`, where each arg is the C name of the argument without a pointer star (a primitive is its Narra name). Function types list the parameters, then the result, as arguments.
- Generic function instances are named `<name>_<hash>` (`FuncDecl.Name`), with the hash over `<module hash>_<name>|arity|arg names`. Implementation functions are named `<Type>_<Trait>_<fn>_<hash>`, the hash over type, trait module, trait and function.
- Generated helper functions put the operation first: `retain_<type>`, `release_<type>`, `eq_<type>`, `hash_<type>`, `cmp_<type>`, `show_<type>` (arrays keep `<array>_new`, `_push` and so on). Arrow functions are `k_<hash>_<owner name>_<n>`, with `K` for the environment and `D` for the drop function. A thunk for a named function is `w_<hash>_<name>`.
- The compiler keeps a table from C name to what it names, and stops with an internal error if two different things get one name. Two modules whose relative paths hash alike are an error.
- The generated C has a comment table right after the `// Types` header, sorted by name: `/* S_Tuple2_<h> = #[int, #[int, int]] */`.
- Top-level names, module path segments, package names, trait names and trait function names have no `__` or edge `_` restriction. Modules `a-b` and `a_b` do not collide.
- A package may not be named `std`. `[[binary]]` names are unique.

`narra.toml` keys the compiler reads:

```toml
[package]
name = "app"

[[binary]]
name = "app"
entry = "cli.run"          # "cli.run" or "app/cli.run"; see Behavior

[link]                     # used by c-interop; ignored until then
libraries = ["sqlite3"]
search = ["/opt/homebrew/lib"]
```

Go types and entry points:

```go
type Binary struct {
	Name  string
	Entry string // "cli.run"
}

type LinkConfig struct {
	Libraries []string
	Search    []string
}

type SourceModule struct { // Module is the std module type of contract builtins
	Path    string // "app/text"
	Files   []*File
	Exports map[string]Decl // filled from declarations with Exported set
}

type Package struct {
	Name     string
	Dir      string
	Modules  []*SourceModule // dependency order, imports first
	Binaries []Binary
	Link     LinkConfig
}

func Load(path string) (*Package, error)       // a directory goes to LoadPackage; a .nra file is a one-module package
func LoadPackage(dir string) (*Package, error)
func CheckPackage(pkg *Package) (*Program, error)
func Emit(prog *Program, entry *FuncDecl, tests bool) string
```

- `File` gains a `Path string` field with the source file path, used for diagnostics.
- `Program.Entries map[string]*FuncDecl` holds each binary's entry function, by binary name. The caller passes one to `Emit` (nil in test mode). `Program.Files map[string]string` maps a module path to its file, for panic messages.
- `FuncDecl.Module`, `TestDecl.Module`, `StructType.Module` and `EnumType.Module` hold the declaring module path; "" is the prelude. Types compare by module and name, never by name alone.
- `Program.File` and `Program.Entry` are removed. `Check(f *File)` stays as a wrapper that checks one file as package `main`.
- The entry function keeps the signature `export function run(args: Array<string>): int`, under the name given in `entry`.
- `narra run file.nra` and `narra build file.nra` still accept a single file. `hello.nra` is package `hello` with module `"hello"` and binary `hello` whose entry is `"hello.run"`, so it emits `f_<hash of "hello">_run`.

## Behavior

- Import cycles are errors that list the cycle.
- Names are private unless exported. Importing a private name is an error.
- Each module is checked against the exported signatures of the modules it imports.
- Entry form: `"cli.run"` resolves to `"<package>/cli"` when that module exists; otherwise the module part is a full path, so `"app/cli.run"` (SPEC.md section 15) also works.
- Every line of a compile error starts with its file. `Diagnostic.File` is filled by the checker and the loader.

## Changes

- v1: initial.
- v2 (frozen 2026-10-28, from the modules change request): `SourceModule`, `Load`, `Program.Entries`, `Program.Files`, declaring-module fields, the injective C name rule, the removal of `Program.File` and `Program.Entry`, both entry forms, the file on every error line.
- v3 (frozen 2026-10-30, from the hashed-cnames change request): hashed declaration names, interned composite type names, operation-first helper names, the collision table, the comment table, removal of the `__` and edge `_` rules.
