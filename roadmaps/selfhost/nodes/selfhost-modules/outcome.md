# Outcome: selfhost-modules

## Built

- `selfhost/src/manifest.nra` (module `narrac/manifest`, 314 lines): port of `modules_toml.go`.
  - `parseManifest(text): Result<Manifest, LoadError> uses alloc`
  - `struct Manifest { name, binaries: Array<Binary>, libraries, search }`, `struct Binary { name, entry }`
  - `struct LoadError { file, line, col, message }`. `file == ""` and `line == 0` means a bare message (the Go error was not a diagnostic). Manifest errors have file `narra.toml`, col 1, and line 0 for the checks after the loop.
- `selfhost/src/modules.nra` (module `narrac/modules`, 431 lines): port of `modules.go`.
  - `load(path): Result<Package, Array<LoadError>> uses io, alloc`. A directory is a package; a file is a one-module package (`Load` in Go). Errors are an array because Go reports several (parse errors in several files, several cycles).
  - `binary(p, name): Result<Binary, string>`, `entryModule(p, entry): #[string, string]` (module, function), `dump(p): string`, `errorText(errs): string`.
- `Package`:
  ```
  struct Package { name, dir: string, modules: Array<SourceModule>, binaries: Array<Binary>,
                   libraries: Array<string>, search: Array<string>, nextId: int, single: bool }
  struct SourceModule { path: string, file: string, ast: File, exports: Array<string> }
  ```
  `path` is `"<package>/<rel>"`, `file` is `src/<rel>.nra` (the base name for a single file), `ast` is the parsed `narrac/ast.File`, `exports` are the exported function, const, struct, enum, trait and distinct names, sorted by bytes, without duplicates. There is one file per module (no `Files` array). `Package.dir` is the given path (the parent directory for a file).
- Id check: `SourceModule` also has `firstId` and `endId` (the file's range `[firstId, endId)`). `narrac pkg --ids <path>` checks that the ranges are disjoint, cover `[0, nextId)` without a gap, and prints `ids ok <nextId>`; `TestSelfhostModuleIds` runs it on every loadable input. It checks the chaining, not that each node id lies in its file's range (the parser owns that).
- Id sequence: files are parsed in sorted path order of `src/` and each `Parsed.nextId` becomes the next file's `firstId`, starting at 0. `Package.nextId` is the final value. Ids follow the sorted order, not the load order. A single file starts at 0.
- Load order: depth-first over the modules in sorted order; a module is appended after the modules it imports (imports outside the package are skipped, as in Go). A cycle is reported at the import that closes it: `import cycle: a -> b -> a`, in the importing file at the import position.
- `narrac pkg <path>` in `main.nra`. Go side: `compiler/internal/narra/pkgdump.go` (`PkgDump`) and hidden `narra pkg <path>`. The dump format is defined in the doc comment at the top of `pkgdump.go`. Goldens: `compiler/testdata/pkgdump/<case>.dump` for the 28 cases in `compiler/testdata/pkgdump/<case>/`; `TestPkgDump` compares them (`NARRA_UPDATE=1` rewrites).
- `TestSelfhostModules` (293 subtests): `narrac pkg` equals `narra pkg` on every package in `testdata/packages` (and `errors/`), the pkgdump cases, `selfhost/`, and every single `.nra` file under `testdata/` and `selfhost/src`.
- Nits folded in: `narrac` returns 1 when writing stdout fails (`writeStdout`); the `selfhost/narra.toml` comment is updated. One selfhost build is shared by all selfhost tests (`buildSelfhost`, `sync.Once`, removed in `TestMain`).

## Deviations

- `parser.nra` (selfhost-front's file): `quoteGo` is now `export ... uses alloc`. Nothing else changed there.
- `selfhost_test.go`: `compareSelfhost` now uses `buildSelfhost` and `runNarrac`; behavior is the same.
- `Result<Package, Array<LoadError>>` instead of `Result<Package, Diag>`, see above.
- No change requests. No Go loader change.
- A symlinked directory in `src/` behaves differently: Go's WalkDir does not follow it, Narra's `listDir` does, so `src/loop -> .` fails in Narra.
- Invalid UTF-8 in `narra.toml` gives a different error text than in Go.
- `popLast` copies the stack on each visit and `sortStrings` is an insertion sort. Both are fine for tens of files and quadratic for hundreds.

## Stubs left

- The check for two modules with the same C name hash (`LoadPackage`) is not ported; it needs xxHash32. The check is deferred to selfhost-emit, which needs xxh32 for C names anyway.
- `Str.trim` is ASCII only; Go's `TrimSpace` also trims Unicode spaces. Only an exotic narra.toml differs.
- OS error texts (missing `narra.toml`, unreadable file, missing `src/`) differ from Go's. The comparison avoids those inputs.

## Gotchas

- `Fs.listDir` returns Err for a file, so `collect` treats Err as "a file". A directory named `x.nra` is walked, not loaded.
- Go sorts the full list of paths, so `a.nra` comes before `a/b.nra` and `B.nra` before `a-b.nra`. `sortStrings` (insertion sort on byte order) does the same on relative paths.
- Narra has no ternary and `match` arms are statements; use `if` and a trailing `return`.
- `exports` holds names only. The checker finds the declaration by name in `ast.decls`.
- `dump` quoting copies `astdump.quote`; keep both in step with `quoteDump` in Go.
- A single file `hello.nra` gives module `hello` and binary entry `hello.run`.

## Proposed learnings

- Sort the full path list, not the directory walk, when output order must match Go | `a.nra` vs `a/b.nra` differ between walk order and sort order | selfhost/ | yes (nested_order)

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (119 s).
- `go test ./internal/narra -run 'TestPkgDump|TestSelfhostModules|TestSelfhostParser'`: pass.

## Size

- Estimated M, actual M. 745 Narra lines (314 + 431) for 569 Go lines (1.3x); the Go dump is 107 lines.
