# copy-builtin log

- 2026-11-10: branch feat/copy-builtin reset to feat/batch-5 (48823a59). No `copy` builtin or E0975 existed.
- Runtime: credit table and NARRA_COPY_* hook macros in runtime.h (empty under NARRA_RELEASE). Hooks at retain, before each decrement, and around each copy on write in emitdecls.nra (array, box), runtime_map.h, std_string.c (builder), runtime.h (generators). main and the test child call NARRA_COPY_START/REPORT. Exit code 103.
- Checker: row in stdprelude.nra, check in new stdcopy.nra, routed from checkNamedCall in calls.nra (it needs the node id for the fix's callEnds). Emitter: emitPreludeCall -> emitCopyCall (grant on, owned temp + retain, grant off).
- Prelude names: a local `const panic = 3` is accepted and hides the builtin (findLocal runs first); a top-level `function panic` is never called. Same answer for `copy`, recorded in SPEC.
- First run of copy_nested under NARRA_CHECK_COPIES: no report. A sample without copy reports; the credit given by a dropped copy(...) ends when the count is back to 1 (copy_report).
- O18 seen in practice: after a credited copy on write, a child granted a credit but never written keeps it, so a later unrelated hidden copy of that child goes unreported.
- Hints: 13 `const copy = %s` hints (not 10) rewritten to `copy(%s)`, report sites in exprs.nra and funcvalues.nra, formats in diagnostics.nra, guide error lines.
- Full check: sh scripts/test.sh PASS (234 s, 0 failed). Outcome written.
- Rebased on feat/batch-5 e13504a5 (lexer-punctuation, scope-map, array-write-cold-paths). Conflicts: emitdecls.nra array _unique (hooks moved into _copy), runtimecore.nra (regenerated). Seed taken from feat/batch-5.
- Review fixes: no E0975 for untracked types inside generic instances; the fix keeps parentheses for non-primary arguments; E0290 suggests copy(X) only for a part of a tracked constant, else a let local.
