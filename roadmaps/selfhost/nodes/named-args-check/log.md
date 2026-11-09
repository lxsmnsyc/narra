# Log: named-args-check

- 2026-11-09: branch `feat/named-args-check` from `origin/feat/batch-1`. Go is retired (PR #101), so only `selfhost/src` changes.
- Canonical view decided first: the AST keeps the written order (evaluation order); the checker records a permutation in `Checker.argumentOrders` only when it differs from the written order; passes that pair an argument with a parameter call `orderArguments`.
- Matching sits in `checkFunctionCall` (plain, generic and module functions) and `checkTraitCall`. Every other callee kind is rejected after the check by its `Ref` (Value, Builtin, Constructor, Distinct), which replaces the E0213 placeholder from named-args-syntax.
- Consumers that pair arguments with parameters: `checkArguments`/`checkGenericCall`/`checkTraitCall` (get the ordered list), effects `walkCall` and `findCallResultType`, `emitCall` (plain, instance and impl calls), `emitTraitCall`, `isSelfCall` and `emitTailJump`. Order-insensitive passes (use collector, closures' mutable roots, id check, typesdump, const checks) keep the written order.
- Emitter: `holdReorderedArguments` holds an argument when a later-written one is passed before it, or a later one may write (the arg-eval-order rule). A call in parameter order has no entry and emits the same C as the positional call (checked by emitting `move(1, 2, 3)` and `move(1, y: 2, z: 3)`: only the file-name hash differs).
- First full build failed the fixed point only because sources changed during the build; rebuilt clean.
- `this call passes N arguments` (E0316) after a duplicate was noise; it is skipped once a duplicate is reported.
- The bootstrap script and some compound shell commands were refused by the worktree guard; ran steps one by one.
