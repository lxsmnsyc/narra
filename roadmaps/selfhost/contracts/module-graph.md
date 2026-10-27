---
name: module-graph
version: 1
status: draft
defined_by: modules
---

# Contract: module-graph

## Purpose

How a package's files become modules, how names stay unique in the generated C, and what `lang.toml` holds. Used by c-interop (`[link]`), opaque-distinct (visibility across modules), build-conditions, query-cli and the self-hosted compiler.

## Interface

Layout and names:

- A package is a directory with `lang.toml` and `src/`. A file `src/<path>.lang` is module `"<package>/<path>"`. Example: package `app`, file `src/text.lang`, module `"app/text"`. There is no directory named after the package inside `src/`.
- A declaration's qualified name is `<module>.<name>`, such as `app/text.helper`.
- C names add the module, with `/` and `.` replaced by `_`: function `f_app_text__helper`, struct `S_app_text__Token`. Prelude and `std/*` names keep their current C names.

`lang.toml` keys the compiler reads:

```toml
[package]
name = "app"

[[binary]]
name = "app"
entry = "cli.run"          # module path inside the package, then the function

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

type Module struct {
	Path    string // "app/text"
	Files   []*File
	Exports map[string]Decl // filled from declarations with Exported set
}

type Package struct {
	Name     string
	Dir      string
	Modules  []*Module // dependency order, imports first
	Binaries []Binary
	Link     LinkConfig
}

func LoadPackage(dir string) (*Package, error)
func CheckPackage(pkg *Package) (*Program, error)
func Emit(prog *Program, entry *FuncDecl, tests bool) string
```

- `File` gains a `Path string` field with the source file path, used for diagnostics.
- The entry function keeps the signature `export function run(args: Array<string>): int`, under the name given in `entry`.
- `langc run file.lang` and `langc build file.lang` still accept a single file. It is treated as a package with one module and one binary whose entry is `run`.

## Behavior

- Import cycles are errors that list the cycle.
- Names are private unless exported. Importing a private name is an error.
- Each module is checked against the exported signatures of the modules it imports.

## Changes

- v1: initial.
