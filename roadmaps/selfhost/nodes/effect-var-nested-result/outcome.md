# Outcome: effect-var-nested-result

## Built

- `bindCallResult` in `selfhost/src/exprs.nra` walks a call's result type (`holdsBoundEffects`, `bindResultType`). Every `Generator` and function type that names the callee's effect variables gets its own bound slot, wherever it sits: Option, Result and other enum instances, generic struct instances, tuples, Array, Map, Box, Async, FixedArray, a generator's element type, and async generators (`TAsyncGenerator`, from async-generators): an `Option<AsyncGenerator<T> uses E>` result charges the source's effects at `Async.next`. Struct and enum instances are made again with the bound arguments. A result with no such type is returned unchanged.
- A bound generator keeps its slot in `origin` (`-2 - slot`), as before. A bound function type nested in the result keeps its slot in `binder` (`-2 - slot`, a new use of a negative binder). A function type that is the whole result is left as it was: the effect pass already finds it through the call that made it (`findSourceCall`).
- New `Checker.boundNames` (`selfhost/src/check.nra`): the effect names the callee declared for each slot. `bindResultEffects` in `selfhost/src/effects.nra` fills every slot of the call from them (`collectBoundSlots`), and `findTypeEffects` reads the slot of a bound function type.
- Private functions with a written `uses` list (coordinator request): `analyseFunction` checks the body of a private function against its written list (E0800), as for an exported one. A private `function*` is checked against `uses` on its result type when it writes one. A private function without a list stays inferred, and callers of a private function still use its inferred set. SPEC section 7 says so in one sentence.
- Tests: `testdata/async_generators_bound_nested.nra` and `testdata/errors/async_generators_bound_option.nra` (a stream through an Option). `testdata/generator_bound_nested.nra` (an `io` source through Option, Array, tuple, generic struct and an Option-wrapped function type is charged `io`; a pure source keeps the caller pure; a private function without a list is inferred). Error tests `testdata/errors/generator_bound_option.nra` (the std-iterator repro, caller `uses alloc`), `func_effects_var_result_option.nra` (function type in an Option), `effects_private_declared.nra` and `effects_private_declared_closure.nra` (the two std-review repros: private functions declared `uses alloc` that run `io` through `countAll<T, effect E>` and `apply<effect E>`).

## Proof

- base: feat/batch-13 at 0bab063f, release compiler from `sh selfhost/bootstrap.sh`. `narrac check` on `/private/tmp/claude-501/std-iterator/p4/main.nra` gives `main.nra:19:25: Function \`pure\` uses \`E\` (through \`Generator<int> uses E\`) but its signature does not declare it` (E leaks, io missed). The function-type variant reports `uses E` the same way. The two private repros and a private `uses alloc` function calling `print` check with exit 0.
- branch: the same repro gives `Function \`pure\` uses \`io\` (through \`Generator<int> uses io\`)`. The function-type variant reports `io`. The private repros report E0800 for `io`.

## Deviations

- `selfhost/src/check.nra`: one new field, `boundNames`, beside `boundGenerators`. Not in the brief's shared paths; the effect pass needs the declared names of every slot, not only the top one.
- Checking written lists on private functions was asked by the coordinator mid-task. It needed: one sentence in `SPEC.md` section 7; `alloc` added to five private functions of the compiler whose lists missed it (`printError` and `finishBuild` in `driver.nra`, `writeStdout` in `fmt.nra` and `main.nra`, `writeTrace` in `lspcache.nra`); and the missing effect added to private functions in 18 samples (`testdata/array_patterns.nra`, `async_api_void`, `async_detach`, `async_map_await`, `async_nested_await`, `async_plain_entry`, `bytes`, `exits/async_await_cycle`, `exits/async_exit_code`, `proven_operands`, `sized_int`, `std_math`, `std_no_panic_numbers`, `std_no_panic_strings`, `std_option`, `std_result`, `std_spec_int`, `wrapping`). Mostly `alloc` for template strings and `await`, and `io` for `print`.

## Stubs left

- None.

## Gotchas

- batch-14 and any branch that merges later: a private function whose written `uses` list misses an effect its body has is now E0800. Expect the same kind of one-word fixes in std sources and samples on feat/batch-14 (std review found such functions there).
- A bound function type nested in a result prints with no `uses` (its effects live in the slot), so a message that prints that type shows `() => int` rather than `() => int uses io`. Bound generators already printed their slot's effects through `generatorEffectList`; function types have no such hook in `typeString`.
- A function-type binder below -1 is now a slot index. Code that reads a `TFunc` binder must test `binder >= 0` before indexing `checker.functions`, as the existing sites do.
- `uses E` in a type takes every following `, name` as an effect until D117 is built, so a test cannot write `#[Generator<int> uses E, int]` or `Result<Generator<int> uses E, string>`; put the generator last.

## Proposed learnings

- A private function that writes `uses` is held to it | callers and readers trust a written list; silently ignoring it hid real `io` | selfhost/src, std/src, testdata | yes (E0800 tests in testdata/errors)

## Verification

- `sh scripts/test.sh` after the rebase on feat/batch-13 (ef0a8023), with the compiler `sh selfhost/bootstrap.sh` built from the seed: `PASS, 8718 cases in 694s` (sanitizers, leak check and the bootstrap fixed point included).
- `narrac emit selfhost` from the base compiler (0bab063f) and from the branch compiler, on the branch source: byte-identical C.

## Size

- Estimated S, actual M. The private `uses` rule added by the coordinator touched the compiler and 18 samples.
