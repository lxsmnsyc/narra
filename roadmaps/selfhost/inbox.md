# Inbox: selfhost

Ideas and gaps found during work, not yet nodes. Triage with the user.

- 2026-10-30, from the prelude-imports review: `SPEC.md` documents `Range.stepped`, but no std module registers it. Either implement it in `std/range` or drop it from the spec. This was already true on main.
- 2026-10-30, from the selfhost-check-imports review (#34): `narrac decls` is quadratic in the number of declarations in one module (2000: 0.3 s, 4000: 1.15 s, 8000: 4.9 s); main shows the same, so it predates #34. Not located; likely the declaration passes or the dump. Fix before the selfhost compiler checks large generated files.
- 2026-10-30, same review: `TestSelfhostBuiltins` checks fixed-signature builtins with a one-way compile probe, so a Go parameter widened beyond the Narra table, or a result that is a subtype, passes. A Go-side dump of `checkFixed` signatures would make it exact.
- 2026-10-31, from the selfhost-check-builtins review (#43): the Go types dump `targs=` header (`typesdump.go:180`) replaces every ", " with ",", which also strips spaces inside tuple, function and generic types; its doc only means the separator. #43 made the Narra dump match for parity. Fix: join the parts with "," in Go, go back to `Str.join(parts, ",")` in typesdump.nra, regenerate the goldens. S.
- 2026-10-31, from the selfhost-check-calls outcome (#41): E0425 (importing a constructor by name) is still `unsupported:` in the Narra checker; it needs `ImportName` positions in the parser. 1 input.
- 2026-10-31, from the readable-names outcomes (#38, #39): exported short names remain (AST fields `id`, `op`, `col`, `params`, `args`, `decls`, `stmts`, `funcs`; idcheck `funcIds`/`blockIds`/`exprIds`; `Diag`/`diagText`; `Checker.cur`, `Checker.errs`, `BodyInfo.targs`). One rename across all modules, best between selfhost-check-effects and selfhost-emit.
