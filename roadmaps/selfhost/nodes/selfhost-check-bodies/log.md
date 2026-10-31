# Log: selfhost-check-bodies

- Branch feat/selfhost-check-bodies from origin/main 93b7058.
- Dispatch layout: a function type cannot take a `mutable` parameter (parse error "expected a name, found mutable"), so a struct-field closure cannot carry `mutable checker`. A trait can: `trait Walker { function checkExpression(walker: Self, mutable checker: Checker, ...) }` in check.nra, implemented by `BodyWalker` in body.nra, and generic functions `<W: Walker>` in exprs/stmts pass `walker` on. Tested in a scratch package before committing to it (trait call across modules with a mutable param works; the struct needs `implements Walker`).
- Go side: `checkPackageBodies(pkg, logBodies)` split out of CheckPackage (no output change: 346 inputs x emit/check/check --json byte-identical to a main build). Bodies logged when their check starts (`logBody`), in check.go, check_const.go, check_types.go (checkDefaults gets an owner name), check_traits.go (implBodies).
- Impl header: resolveImpl renames impl functions to their C name; the dump takes the source name back from `traits.implFuncs`.
- Format choice: type last (decls dump rule), so attributes come before it. Kinds are the AST dump names. Untyped node prints `-`.
- Scope stack: std/array has no pop or slice, so `localScopes` keeps slots and a `scopeDepth` counter.
- Generic struct defaults: Go checks them inside buildGeneric. resolve.nra's HOOK now queues the generic (`pendingDefaults`) when checkingBodies; body.nra drains the queue after each module's declarations, before the non-generic defaults. Error order can differ from Go only when a trait error of the same module falls between.
- Ported E0421/E0423/E0424 (scopes.nra checkImportNames, importModules) and E0194 (declare.nra). E0425 (constructor import) needs ImportName positions; in types mode it is `unsupported`.
- Maintainer rule (2026-10-31, from the coordinator): new Narra code uses full descriptive names (`checker`, `expression`, `leftType`), action-named function declarations, for-of loops, step comments inside functions. Applied to all new code; existing code untouched except where edited.
- Sweep: 345 inputs, 143 compared, 202 skipped (30 do not load, the rest unsupported). No mismatch.
