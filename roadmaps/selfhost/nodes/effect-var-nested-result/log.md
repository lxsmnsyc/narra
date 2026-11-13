# Log: effect-var-nested-result

## Proof on the base

- Base: feat/batch-13 at 0bab063f, compiler built with `sh selfhost/bootstrap.sh /private/tmp/claude-501/effect-var-nested-result/base` (release flags).
- Command: `narrac check main.nra` on the repro copied from `/private/tmp/claude-501/std-iterator/p4/main.nra` (`wrap<effect E>` returns `Option<Generator<int> uses E>`, the caller `pure` is `uses alloc` and runs an `io` source through it).
- Result: exit 1 with `main.nra:19:25: Function \`pure\` uses \`E\` (through \`Generator<int> uses E\`) but its signature does not declare it; write \`uses E\``. The callee's variable `E` leaks into the caller, and `io` is never charged.
- The same leak shows for a function type in the result: `keepSome<effect E>(escaping action: (value: int) => int uses E): Option<(value: int) => int uses E>` makes the caller report `uses E`.

## Work

- `bindCallResult` (exprs.nra) now walks the result type. Each `Generator` whose binder is the callee gets its own bound slot, inside Option, Result and other enums, Array, Map, Box, Async, FixedArray, tuples and generic structs. Struct and enum instances are made again with the bound arguments. A result with no such generator is returned unchanged.
- The slot keeps the effect names the callee declared in a new `Checker.boundNames` (check.nra), so the effect pass (`bindResultEffects` in effects.nra) can fill every slot of the call, not only the top one.
- A function type inside the result (`Option<(value: int) => int uses E>`) takes a slot the same way: its binder becomes `-2 - slot`, and `findTypeEffects` (effects.nra) reads the slot. A function type that is the whole result keeps its names, since the effect pass already finds it through the call that made it (`findSourceCall`).

## Private functions with a written `uses` list (coordinator, 2026-11-13)

- Repros from the std review: a private `hidden(): int uses alloc { return countAll(noisy()) }` with `countAll<T, effect E>` and an `io` generator, and a private function that declares `uses alloc` and passes an `io` closure to `apply<effect E>`. Both passed on the base (`narrac check` exit 0), and so did a private `uses alloc` function that calls `print` directly: `analyseFunction` returned before the check for every private function.
- SPEC section 7 says the compiler infers effects for private functions. It says nothing about a private function that writes a list. Decision: a written list holds. A private function with a `uses` list (or a private `function*` with `uses` on its result type) is checked against it, as an exported one is (E0800). A private function without a list stays inferred. Callers of a private function still use its inferred set.
- The compiler had five private functions whose lists missed `alloc` (`printError` and `finishBuild` in driver.nra, `writeStdout` in fmt.nra and main.nra, `writeTrace` in lspcache.nra). Fixed by adding `alloc`.
