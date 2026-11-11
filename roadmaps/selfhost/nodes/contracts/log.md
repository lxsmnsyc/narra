# Log: contracts

- 2026-11-11: branch feat/contracts reset to feat/batch-7. Nothing of `requires` existed (SPEC example failed to parse).
- Design: `FuncDecl.requires: Array<Expr>`; parser reads `requires` (contextual, start of a line after the signature). Conditions are checked in the callee's body (contracts.nra checkRequires, called from body.nra after declareParams), normalized to `Requirement { parameter, comparison, bound }` in `checker.requirements[funcIndex]`, and seeded as facts with operands.refineFact. Calls record argument facts (`checker.contractCalls`, from calls.nra checkFunctionCall) and are proven after every body (body.nra checkModules, after checkPending), so call order and module order do not matter.
- idcheck.nra collectFuncIds must include the conditions, or the body's id range misses them (outside the shared list; minimal edit).
- Samples pass: testdata/contracts_requires.nra runs; errors samples give the expected E0984/E0985/E0986/E0982.
- E0986 no longer returns void (the follow-up "expected (int, int) => int, found void" was noise).
- First full run: SPEC sample had top-level statements (now inside a function), stale spec-skip line removed, guide sample used int template parts (now Int.toString). Rerun of spec, guide, programs suites: pass.
- Review round 1: fixed the `>`-ended result type parse, moved requirement reading to the declaration phase (LSP), sentinel bounds, u64 literals from 2^63 up (parseLiteral), E0985 wording, fmt comments on requires lines. New cases: errors/contracts_extremes, fmt/comment_requires, lsp/incremental_requires.
- Review round 2: parameters shadow module constants and aliases in conditions; a bound that wraps at the parameter's width is E0984.
