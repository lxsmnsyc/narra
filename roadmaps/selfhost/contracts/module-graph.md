---
name: module-graph
version: 2
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
- C names add the module: every byte of the module path outside `[A-Za-z0-9_]` becomes `_`, then `__` follows. Function `f_app_text__helper`, struct `S_app_text__Token`, generic instance `f_app_util_list__bagOf__int`. Prelude and `std/*` names keep their current C names.
- The mapping is injective: a module path segment (after mapping) and the package name must not contain `__` or start or end with `_`, and top-level names must not contain `__`. The first `__` after `f_`, `S_` or `E_` separates the module from the name. Two modules whose mapped paths are equal (`a-b` and `a_b`) are an error.
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
- `narra run file.nra` and `narra build file.nra` still accept a single file. `hello.nra` is package `hello` with module `"hello"` and binary `hello` whose entry is `"hello.run"`, so it emits `f_hello__run`.

## Behavior

- Import cycles are errors that list the cycle.
- Names are private unless exported. Importing a private name is an error.
- Each module is checked against the exported signatures of the modules it imports.
- Entry form: `"cli.run"` resolves to `"<package>/cli"` when that module exists; otherwise the module part is a full path, so `"app/cli.run"` (SPEC.md section 15) also works.
- Every line of a compile error starts with its file. `Diagnostic.File` is filled by the checker and the loader.

## Changes

- v1: initial.
- v2 (frozen 2026-10-28, from the modules change request): `SourceModule`, `Load`, `Program.Entries`, `Program.Files`, declaring-module fields, the injective C name rule, the removal of `Program.File` and `Program.Entry`, both entry forms, the file on every error line.
