# Outcome: prelude-imports

## Built

- `Array.*`, `Map.*`, `Range.*` and `Box.*` are std modules: `std/array` (`std_array.go`: length, values, push, filled), `std/range` (`std_range.go`: upTo, between), `std/box` (`std_box.go`: new), `std/map` (`std_map.go`: size, remove, keys, entries). Builtin names, effects and emitters are unchanged. `print`, `assert`, `panic` stay in the prelude (`std_prelude.go`); types and constructors stay global.
- Registration: `registerNamespaceModule(ns, path, funcs)` (builtins.go) sets `Builtin.Name = ns + "." + name` and calls `registerModule`, so the modules appear in `stdModules` like `std/string`. `namespaceModules` (ns to path) exists only for the error. The old `registerNamespace`, `namespaces` and `namespaceFuncs` are removed. selfhost-check-imports should port `stdModules` with these four as ordinary modules and no namespace special case.
- Alias equal to a type name: `import * as Array from "std/array"` puts `Array` in `Checker.modules`. A call `Array.f(...)` checks `c.modules` before anything else, so it resolves to the module. `Array<int>` goes through type resolution, which never reads `c.modules`, so the type is untouched. A user type named `Box` or `Map` still hides the prelude one: without an import, `Box.new` is the user's; with `import * as Box`, the module wins.
- Without the import: error E0908 ``%s.%s needs an import; add `import * as %s from %q` `` with a fix that inserts the import line before the module's first import, or at 1:1 (`Checker.importFix`, `modScope.importAt`). The method-call fix `xs.push(v)` adds the same edit.
- Tests: `testdata/errors/prelude_{array,map,range,box}.nra`, `testdata/diagnostics/prelude_import_fix.nra` (code E0908, the fix must compile).
- Migration: `roadmaps/selfhost/nodes/prelude-imports/migrate.py`. 53 `.nra` files (compiler/testdata, selfhost/src, bench) gained 84 lines, all `import * as NS from "std/ns"` plus a blank line in files that had no imports. 0 deletions. One golden changed (`testdata/astdump/exprs.ast`).
- Docs: SPEC.md section 3 (prelude bullet, new bullet, import example), D67 in DECISIONS.md, compiler/README.md Supported.

Proposed SPEC.md text (gate):

> **Prelude.** The prelude holds the types and constructors `Option`, `Some`, `None`, `Result`, `Ok`, `Err`, `Array`, `Map`, `Range`, `Box` and the primitive types, and the functions `print`, `assert` and `panic`. Nothing else is in scope without an import. These three are the only global functions.
>
> **A type in the prelude does not bring its functions.** The functions of `Array`, `Map`, `Range` and `Box` live in `std/array`, `std/map`, `std/range` and `std/box` (D67). `Array.push(mutable xs, v)` needs `import * as Array from "std/array"`. The alias may equal the type name: `Array<int>` is the type and `Array.push` is the module. Without the import the call is a compile error (E0908) whose fix adds the import line. A type the program declares named `Box` or `Map` hides the prelude one. Code samples in this document leave these imports out.

## Deviations

- SPEC.md code samples elsewhere (`Range.upTo`, `Array.push` ...) were not given import lines; a sentence says samples omit them.
- Files edited outside the owned paths, all in `shared`: builtins.go, check.go (`importAt` set), check_call.go, diagnostic_codes.go, diagnostic_fixes.go, modules_check.go (one field), std_prelude.go, std_map.go. check_types.go and decldump.go needed no change.
- No change requests.

## Stubs left

- None.

## Gotchas

- Imports are per module scope; `importAt` records the first file's first import.
- A file that declares its own `Box` or `Map` type and also imports the alias: the alias wins for `Box.new`.
- Goldens with positions shift by the added import lines (two lines when the file had no imports).

## Proposed learnings

- Register a std namespace through `registerNamespaceModule`, not as a prelude entry | the prelude has no functions except print, assert, panic | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (sample programs under ASan, UBSan and leak check, release builds, selfhost dumps).
- Migration is only additions: `git diff -U0 -- '*.nra'` shows no removed line and no added line other than an import or a blank.
- `exprs.ast`: with every `L:C` removed, old and new differ only by the new `.imports` entry and `end=` attributes (positions).

## Size

- Estimated M, actual M.
