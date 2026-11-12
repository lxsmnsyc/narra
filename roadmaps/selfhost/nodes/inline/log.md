# Log: inline

- 2026-11-12: started on feat/inline from feat/batch-11 (a01c3540). Not built yet: `@inline` is E0990.
- Design notes from reading the emitter:
  - Local C names are `v_<name>_<number>` with a program-wide number, so caller and callee locals never clash. Only a second expansion of the same body in one C function repeats names. Each expansion is a C block `{ }`, so repeats are siblings or shadow correctly; a generator caller writes the callee's locals as C locals of the resume function (no yield can happen inside an expansion), so frame fields are not shared.
  - C labels are function-wide: loop labels (`narra_break_<line>_<column>`) get a per-expansion suffix.
  - Reachability is driven by `noteReference`; an expansion does not note the callee, so a program that only calls an `@inline` function has no C function for it without a change to `emitReachableBodies`.
  - The check pass runs from `body.nra` (`checkModules`), because `check.nra` is the base module that `inlining.nra` imports.
- Built check pass, emitter, samples, errors, goldens, docs. Rebased once onto feat/batch-11 (c440978d); diagnostics.nra conflict with E1020/E1021 rows, kept both.
- Measured callback cost (inline_cost.nra): release -O2 equal (~1.7 ns, clang already inlines the static helper); -fno-inline 6.4-7.3 vs 3.4-4.2 ns; -O0 29 vs 25 ns.
