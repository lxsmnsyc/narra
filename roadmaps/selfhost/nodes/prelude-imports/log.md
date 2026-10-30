# Log: prelude-imports

- Used D67 and diagnostic E0908 (next free in the E09xx standard library range).
- Registration: new `registerNamespaceModule(ns, path, funcs)` in builtins.go sets each Builtin.Name to `ns.name` (the emitter, effects pass and for-of checks key on those names) and calls `registerModule`. `namespaceFuncs` and `namespaces` are gone; `namespaceModules` maps `Array` to `std/array` for the error.
- Call resolution (check_call.go): an import alias in `c.modules` is tried first, so an alias equal to a type name works with no other change (types resolve through the type path, never `c.modules`). Only when there is no alias does the old namespace branch run, now an error with a fix. A user type `Box` or `Map` still hides the namespace.
- Fix: `Checker.importFix` inserts the import line before the module's first import (`modScope.importAt`, set in declareModule) or at 1:1. The `xs.push(v)` method-call fix adds the same edit when `Array` is not imported, so it still compiles after the fix.
- Migration: `migrate.py` (this folder, idempotent) is the only editor. 53 .nra files changed, 84 insertions, 0 deletions. Checked with `git diff -U0 -- '*.nra'` filtered for added lines that are not an import or blank (empty).
- Golden: only `astdump/exprs.ast` changed. Checked by stripping every `L:C` and diffing old and new: only the new `.imports` entry and `end=L:C` string attributes differ (positions).
- No selfhost Narra port change: declare.nra records only `std/string`; the decl dump has no body checking. Dump tests pass.
- Dead end: the bash tool refuses long heredoc commands in this worktree path; files were written with the Write tool.
