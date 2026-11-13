# Outcome: inline

## Built

- `@inline` parses (`KNOWN_ATTRIBUTES` row in `selfhost/src/parser.nra`). On a member of an `extern { } from` block the parser reports E1031; in an implement block it stays E0990; before an extern block it stays E1000.
- `selfhost/src/inlining.nra`: `checkInlineFunctions` (run from `checkModules` in `body.nra`, after `checkContractCalls`) reports E1030 for the first call chain through `@inline` functions back to the start, once, at the first call, naming the chain (`isEven -> isOdd -> isEven`), and E1031 for a generator. It walks the plain body and every instance body. `isInlineFunction` and `collectBlockUses` (calls and names of a block, marked when inside an arrow) are exported for the emitter.
- `selfhost/src/emitinline.nra`: `emitInlineCall` (dispatched from `emitPlainOrInstanceCall` in `emitexprs.nra`) evaluates the arguments in the caller, left to right, then writes the callee's body (the instance body for a generic call) in a C block with a result variable and a `narra_inline_N` label, reading the callee's side tables. `return` becomes a `BlockInfo` return. `emitInlineValueCall` (dispatched from the `Ref.Value` arm of `emitCall`) writes a bound arrow's body where the copy calls its parameter, back in the caller's state; any other value call goes to `emitValueCall` as before.
- Emitter state in `emitstate.nra`: `Emitter.inlineSites` (`InlineSite`, `InlineCaller`, `InlineArrow`), `inlineCount`, `labelSuffix`, and `LoopInfo.suffix` so loop labels of two copies in one C function differ.
- An arrow argument is copied only when every read of its parameter is a call outside any arrow function in the body. Otherwise the closure is built at the call (stack environment as before). A named function or a closure variable is passed as a value. An `@inline` function used as a value keeps one ordinary C function; a program that only calls it has none (reachability follows `noteReference`, so `emitReachableBodies` did not change).
- Samples: `testdata/inline.nra` (the brief's program), `inline_returns`, `inline_mutable`, `inline_generic`, `inline_nested` (two copies in one expression, a copy inside another's arrow, the same function inside its own arrow, an arrow assigning captured `let`s, nested loops of two copies), `inline_values` (parameter passed on, named function, closure variable, value use, recursion through a plain function), `inline_generator` (a generator caller), `inline_attributes` (`@target`, `@trusted`), `inline_cost`, `testdata/packages/inline` (across modules, generic, in a test block), `testdata/panics/inline_bounds.nra` (a panic in a copy).
- Errors: `testdata/errors/inline_{self_call,cycle,tail_call,generator,extern,implement}`.
- Emit check `inline/no-closure` (`selfhost/tests/src/emitchecks.nra`): the C of `testdata/inline.nra` has no `_applyTwice` and no `narra_env`.
- Goldens: `testdata/astdump/inline_attribute`, `testdata/fmt/comment_inline_attribute` (stable under `narrac fmt`).
- E1030 and E1031 rows and explanations in `diagnostics.nra`. SPEC section 2 states the rules; section 11 no longer says `static inline`. D95 has a "Built" paragraph for `@inline`. `compiler/README.md` drops `@inline` from "Not supported yet" and has an `@inline` section naming both files.

Callback cost (`testdata/inline_cost.nra`, a capturing arrow through a helper shaped like `Options.map`, 300M calls, 3 rounds each, macOS arm64):

| build | without `@inline` | with `@inline` |
| --- | --- | --- |
| `-O2 -DNARRA_RELEASE` (release) | 1.71 to 1.72 ns (one round 2.57) | 1.66 to 1.73 ns |
| `-O2 -DNARRA_RELEASE -fno-inline` | 6.4 to 7.3 ns | 3.4 to 4.2 ns |
| `-O0` (debug, 100M calls) | 28.9 to 31.8 ns | 24.9 to 25.2 ns |

In a release build clang already inlines a static helper and its stack closure when both are in one C file, so the gain shows only where the C compiler does not inline (a large helper, `-fno-inline`, debug builds). The C of the inline loop has no environment struct, no closure function and no indirect call (std-option-result recorded one stack environment and one indirect call per call).

## Proof

Not a fix node.

## Deviations

- Per-site C names: the brief asks for a suffix per call site on the copied locals. Local C names are already program-wide (`v_<name>_<number>`, the number from `checker.localCount`), so caller and callee never clash; a second copy of the same body is a sibling C block, or an inner block where shadowing is correct. Only loop labels (function-wide in C) get a per-copy suffix. A generator caller writes the copy's locals as C locals instead of frame fields, since no `yield` runs inside a copy.
- E1031 for an extern block member is reported by the parser (members accept only `@extern` there, E1000 before), not by the checker.
- Files outside owns and shared: `selfhost/src/body.nra` (one call and one import; `check.nra` is the base module that `inlining.nra` imports, so the call cannot live there), `testdata/errors/target_attribute_unknown.{nra,err}` (used `@inline` as its unknown attribute; now `@cold`, and the list of known attributes grew), `testdata/panics/inline_bounds.nra` and one row in `selfhost/tests/src/panics.nra` (the panic case the worker instructions ask for), `testdata/packages/inline/` (a call across modules needs a package).
- No `async` case: top-level `async function` does not parse yet (plan review).
- `emitReachableBodies` is unchanged: an expansion never calls `noteReference` for the callee, so the rule holds without a change there. lazy-values can change that function freely.

## Stubs left

- None. Marking `Options.map`, `andThen`, `orElse` and other std helpers `@inline` waits for the end-of-batch seed update (orchestrator, 2026-11-12).

## Gotchas

- Code that runs while a copy is written sees `emitter.body` set to the callee's body. Anything that indexes the caller's tables by local index (`refCaptures`, `hoistedArrays`, `blocks`) must be saved in `InlineCaller` too; a new per-body field of `Emitter` needs the same.
- A call through a parameter is copied only when the parameter is the top site's; a closure function written inside a copy keeps `inlineSites`, which is safe because a parameter read inside an arrow is never bound.
- `@inline` with a named function argument (`call(4, negate)`) still makes an indirect call through a thunk. Calling the named function directly would be a later improvement.

## Proposed learnings

- A new per-body field of `Emitter` must be saved in `InlineCaller` (emitinline.nra) | copies switch bodies while one C function is written | selfhost/src | no

## Verification

- Targeted suites with the branch compiler: errors `PASS, 621 cases`, panics `PASS, 5 cases`, emit `PASS, 19 cases`, programs `PASS, 399 cases`, packages `PASS, 36 cases`, dumps `PASS, 164 cases`, codes `PASS, 1 cases`, sources `PASS, 1 cases`, spec `PASS, 62 cases`, habits `PASS, 687 cases`, guide `PASS, 294 cases`.
- Every sample ran with and without `@inline` (the attribute removed) under `-sanitize`, `NARRA_CHECK_LEAKS=1` and `NARRA_CHECK_COPIES=1`, and in a release build: same output, exit 0.
- `sh scripts/test.sh`: `test.sh: PASS, 7500 cases in 523s`

## Size

- Estimated L, actual L.
