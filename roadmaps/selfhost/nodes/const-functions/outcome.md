# Outcome: const-functions

## Built

- `@const` parses on a top-level function (`KNOWN_ATTRIBUTES` row; `parseAttribute` takes the `const` keyword as the name). Goldens: `testdata/astdump/const_attribute`, `testdata/declsdump/const_functions`, `testdata/fmt/comment_const_attribute`.
- `checkConstCall` (`exprs.nra`) accepts a call of an `@const` function; its arguments go through `checkConstValue`. Other calls keep E0192.
- `selfhost/src/constfunctions.nra`: `checkConstFunctions` after the effect pass, once per `@const` function: E1050 (signature types, generator, trait constraint, effect parameter), E1051 (declared or inferred effect beyond `alloc`, naming the effect and the call that brings it), E1052 (calls of functions that are not `@const` or an allowed std function, `Math.sin` and the other C library functions, functions that take a function such as `Arrays.map` and `Map.update`, std modules outside the six, user `implement` blocks through `==` or a map key or a call, arrows, function values, explicit trait calls).
- `selfhost/src/consteval.nra`: `checkConstants` (the entry, runs the checks then the evaluator) and the tree-walking evaluator. Statements: `const`, `let` with patterns, assignment and compound assignment into nested places, `if` (build-folded conditions honored), `while`, C-style `for`, `for ... of` over ranges (`upTo`, `between`, stepped values), arrays, `Array.values`/`entries`, `Map.keys`/`entries`, fixed arrays, `for (mutable ...)`, labels, `match` in statement and value position (guards over each alternative of a list), `return`, block expressions. Every expression kind but arrows, all patterns, `mutable` arguments as places, generic `@const` functions (instance bodies), std source functions of the six modules (`Arrays.slice`, `Int.floorDiv`, `Int.compare`), struct field defaults, tail self-calls as loops. Values are `ConstValue` (`check.nra`), kept in `Checker.constValues`. Limits: 10 M steps (E1054), 64 MiB counted as D105 item 5 (E1055, charged before `Arrays.filled`, `withCapacity`, `Str.repeat`, `join`, `replace`, `concat`, templates), 1000 calls (E1056), plus a nesting depth of 10000 and a value depth of 1000 (both E1056, see Deviations). A std string, array or map call costs a step per 8 bytes or elements it goes through. `copy(...)` and a value read from a place and stored elsewhere (a `const`/`let`, an assignment, a push, a literal field or element, a return) charge the outer level again (`16 + 8 * length` for an array), as the copy on write it can cause. Equality, map keys and the value-depth measure walk values with their own stack, not the C stack. Panics are E1053 with the std preconditions checked first (`Str.sliceFrom`, `charAt`, `nextChar`, `byteAt`); E1057 for a constant read while it is computed. Errors are one diagnostic at the constant with the call chain as `file:line:column: note:` lines; a recursion's repeated note shows its count.
- Std rows supported: all 51 registry rows of `std/string`, `std/array`, `std/map`, `std/int`, `std/float`, `std/range` except `Map.update` (takes a function); `Array.values`/`entries` only as loop iterables.
- `selfhost/src/emitconstvalues.nra`: scalar results written at each use (`isStaticConstant`, `constRef` hooks in `emitexprs.nra`), other results built in the D61 getter from the value (`emitGetter` hook in `emitconsts.nra`). Arrays and fixed arrays of scalars: one `static const` C array and one `memcpy`. Non-finite floats: `NAN`, `INFINITY`, `(-INFINITY)`. Floats otherwise use the shortest round-trip text. Strings stay element stores (the brief's recommendation).
- `emit.nra` writes `#pragma STDC FP_CONTRACT OFF` first in every C file.
- E1050 to E1057 rows and explanations in `diagnostics.nra`.
- Samples (each prints the constant and the same call at run time, equal lines in the `.out`): `testdata/const_functions` (CRC-32), `_scalar` (square, a bool, a 5000-deep tail recursion, i8 wrapping, a static constant built on a computed one), `_map` (keyword map in a loop, overwrite, remove), `_struct` (struct, enum, tuple, Option, fixed array, defaults, spread), `_text` (templates, builder, std/string, a top-level constant read, a function calling another), `_float` (`Float.sqrt` 0 to 15, `a * b + c`), `_nonfinite`, `_language` (generics, `mutable` parameters, nested places, labels, patterns, u64/u8, std source functions), and `testdata/packages/const_functions_modules` (an exported `@const` function from another module).
- Error tests `testdata/errors/const_eval_*` (23): E0192, E1050 x2, E1051 x2, E1052 x6 (with `f32_literal`), E1053 x3 with chains, E1054 x2 (with `scan_steps`), E1055 x3 (with `copy_memory`, `shared_memory`), E1056 x3 (calls, nesting, `value_depth`), E1057. Samples `const_functions_deep_equal` (equality of two 60000-deep lists at compile time) and `const_functions_deep_calls` (999 calls with six nested additions each). Every repro of the fresh review is one of these.
- Emit checks: `const-functions/scalar-inline`, `static-table`, `fp-contract`, `deterministic`.
- SPEC section 10 (rules of items 2, 3, 5, 6, 8), section 11 (pragma), section 12 (evaluation in the compiler). `compiler/README.md`: `@const` section, "Not supported yet" now lists only what stays out.

## Proof

Not a fix node. Measurements (Apple M2, macOS 27):

- `narrac emit selfhost` instructions, release build, median of 5: before 2,597,720,610 (base at 7941611a); after 2,795,980,739 (this branch before the rebase onto 47cd410a). The base compiler on the new source takes 2,790,176,341, so the rise is the larger source (about 4,500 new lines), not a slower compiler. Final tree (rebased onto 47cd410a, review fixes): 2,919,372,562 with the branch compiler against 2,881,622,027 with the base compiler on the same source (+1.3%, which includes the three nodes rebased in that the base compiler does not have).
- The pragma matters: without it, clang -O2 on arm64 prints 0.7400000000000001 for one run-time entry of `const_functions_float`; with it, run time and compile time agree.
- 10 M steps take 0.24 s (about 440 instructions per step). The E1055 test peaks at 4 MB.
- Stack per interpreted call, five nested blocks per call: about 4.5 KB at -O2 (999 calls fit in 4 MiB); about 16 KB at -O0 (600 calls overflow 8 MiB).

## Deviations

- Files outside owns and shared: `selfhost/src/driver.nra`, `selfhost/src/emitter.nra` and `selfhost/src/lspcache.nra` (one `checkConstants` call and import each; the effect pass runs in three pipelines), `testdata/errors/target_attribute_unknown.err` (the known attribute list gains `@const`), `testdata/packages/const_functions_modules/**` (package samples live there).
- `ConstValue` and `Checker.constValues` are in `check.nra`: the evaluator and the emitter both import it, and neither can import the other.
- E1057 is found by the evaluator, not in `checkConstCall`: the check of a constant's value does not walk the `@const` function's body.
- Limits beyond D105 (for the D105 update). The rule: a computation may make at most 1000 nested calls (tail self-calls excepted), and the calls, blocks, loops and expressions under way across all open calls may nest at most 10000 deep, each call, block, loop statement and expression counting one level; the computed value may nest at most 1000 levels (arrays, structs, enums, maps). All three are E1056 at the constant. Six nested additions per call still allow 999 calls (`const_functions_deep_calls`). Measured on a release build: the nesting limit fires before a crash with a 6 MiB stack for if-heavy and loop-heavy bodies (a 5 MiB stack overflows on the if-heavy one), so an 8 MiB stack keeps a margin of about 1.3. The value depth exists because the C writer recurses once per level (a List 8000 deep crashed `emit` before).
- Memory counting adds copies D105 item 5 does not list: `copy(...)` and stores of values read from places (see Built). Without it, a shared 1,000,000-element row written 400 times reached 1.86 GB inside the 64 MiB budget.
- Steps: a std call also costs a step per 8 bytes or elements it scans (D105 item 5 counts one step per expression). Without it, 2000 scans of an 8 MB string took 70 s inside the step budget.
- `narrac check`, `build` and `emit` now compute constants after the shared-write check (E0978), so a program it rejects allocates nothing. The language service still computes them in its effect phase (`lspcache.nra` keeps effect errors apart from shared-write errors).
- Float literals in `@const` bodies are read exactly only with at most 18 significant digits and a power of ten that reduces to at most 22 (one exact product or quotient); others are E1052 at the constant. The compiler has no decimal-to-double parser it may call (`std/internal/float` is internal to std). An f32 literal is read as f64 then rounded; when the f64 read was inexact and lands exactly halfway between two f32 values (the only case where the double rounding can differ from C), it is refused with E1052.

## Stubs left

- `std/char`, `std/option`, `std/result`, `std/math` (less the C library functions), `std/bytes`, `std/fixedarray` and the sized number modules: follow-up node `const-functions-std` (D105).
- A full decimal float reader for long literals.
- Arrays of strings in a computed value are one store per element; pointing at the static string literals is a follow-up.

## Gotchas

- Narra rejects `mutable` parameters of enum type and whole assignment of a `mutable` parameter, so the evaluator mutates through arrays: `takeAt(mutable items, index)` and `storeIntoArray`, with the slot array as root. Reading a value next to a write must happen in a helper, or the write copies.
- A block arm of a match in value position gets its value from `return`; a trailing bare expression is not the value.
- `std/<name>` builtins have their Narra part in module `std/ext/<name>`; `isConstStdModule` maps it back.
- A compiler built at -O0 still overflows its stack before the limits on deeply nested recursions; only release builds are protected.
- DECISIONS.md D105 still ends with "Not built yet." (not in this node's paths).

## Proposed learnings

- Evaluator state that is mutated in place lives in arrays, not in enum-typed `mutable` parameters | Narra forbids both forms | selfhost/src | no
- Any new pipeline that runs `Effects.checkEffects` must call `checkConstants` after it | three pipelines exist (driver, emitter, lspcache) | selfhost/src | no

## Verification

- `sh scripts/test.sh` on the final tree (on 47cd410a, review fixes included): `test.sh: PASS, 8103 cases in 130s`
- Targeted with the branch compiler: programs `--filter const_functions` PASS, 18 cases (sanitizers, leak check, copy check, release); errors `--filter const_eval` PASS, 18 cases; packages PASS, 3 cases; emit PASS, 24 cases; dumps `--filter const_` PASS, 6 cases; fmt PASS, 1 case.
- `sh selfhost/bootstrap.sh`: fixed point (stage 2 and stage 3 identical), seed not regenerated.

## Size

- Estimated L, actual L. The evaluator came out larger than estimated (about 3,200 lines in `consteval.nra` after formatting).
