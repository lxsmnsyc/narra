# Outcome: const-params-pointer

## Built

- A constant (not `mutable`) parameter is passed as `const T*` when its C value is over 16 bytes. This covers structs, inline enums and ranges, and distinct types over them.
- Size: `isPassedByPointer` in `selfhost/src/check.nra` uses the C layout with field alignment. `finish` computes it once per struct and enum into `Checker.structLayouts` and `enumLayouts` (`CLayout`). A type made after `finish` is measured when it is needed.
- The choice depends only on the type, so every site agrees:
  - signatures: `parameterCType` and `functionSignature` in `cnames.nra`
  - direct calls and generic instances: `emitFunctionCall`
  - function types, thunks, arrow functions, `emitValueCall` and `callFunctionValue`: `emitclosures.nra`
  - user trait implementations called from the derived helpers: `emittraits.nra`
  - `retain_*` and `release_*` of large types: `emitdecls.nra`, `structOrEnumCode`
- Arguments go through `passArgument` (`emitstate.nra`). A place (`isPlace`: names, `.`, `->`, `(*p)`) is passed by address. Any other value is stored in a temporary first.
- Inside the callee, `isPointerParameter` makes `localRef` read `(*v_x)`.
- Docs: a D50 amendment in DECISIONS.md, a "Supported" line in `compiler/README.md`, and the ownership comment in `emitstate.nra`.
- Samples:
  - `testdata/params_pointer_calls`: direct calls, a parameter passed on, call results, literals, a constant, a generic instance, an inline enum, an Option, `==` through an implementation, a named function as a value, arrow functions, a nested closure and spread copies.
  - `params_pointer_tail`: two large arguments swapped on every jump, three rotated, one passed on unchanged and one rebuilt from itself.
  - `params_pointer_escape`: a generator and a coroutine that outlive the caller's value, and closures that are returned or stored in a struct.

### Aliasing cases and how each is handled

- **Same variable passed as `mutable` and in another argument.** The checker already rejects it, also for a field (E0319). Covered by `testdata/errors/mutable_twice` and `mutable_read_conflict_*`.
- **A closure captured by reference writes the variable.** A closure that assigns a variable, or passes it as `mutable`, rejects any other argument that reads that variable or a field of it (E0614). Only non-escaping closures passed straight to the call capture by reference, so no other path exists. Covered by `closure_borrow_field_arg` and `closure_borrow_mutable_arg`.
- **Self tail call.** Every new argument is evaluated into a temporary first. Then each large parameter's value is stored in `<name>_held`, and the parameter points at it. The caller's value is never written.
- **Generator or coroutine.** The frame copies `*p` and retains it when it is created.
- **Escaping closure.** The environment copies the parameter, as before. In the closure body the parameter is `const T* = &k_env->c_x`. Stack environments copy as well.
- **A later argument that changes an earlier one.** `holdEarlierArguments` already holds the earlier one in a retained temporary, so its address is the temporary's.

## Proof

Instructions retired, `/usr/bin/time -l`, 3 runs each, macOS arm64. Load average was 4 to 10 because other worktrees were running.

- Compiler, `narrac emit selfhost`:
  - On top of `feat/batch-4` 5ba04ec (after the rebase): base 3.177 G, 3.159 G, 3.135 G; branch 2.690 G, 2.636 G, 2.633 G. The median goes from 3.159 G to 2.636 G (-16.6%). Peak memory is 128 MB on both.
  - Before the rebase, on 9501a3c: 3.106 G to 2.621 G (-15.6%).
- Bench programs, medians, 9501a3c base and branch:

  | program | base | branch |
  | --- | --- | --- |
  | fib | 1.213 G | 1.213 G |
  | sieve | 2.266 G | 2.289 G |
  | matmul | 2.446 G | 2.446 G |
  | strings | 1.096 G | 1.096 G |
  | interp | 1.964 G | 1.963 G |
  | particles | 2.612 G | 2.612 G |

  - Only particles changes its C: `move` takes a `const Particle*`, and clang inlines it.
  - The other five compile to byte-identical C, so their differences are noise.
- `bench/run.sh -n 5`, wall clock, base then branch, under load:

  | program | base | branch |
  | --- | --- | --- |
  | fib | 0.122 | 0.113 |
  | sieve | 0.260 | 0.283 |
  | matmul | 0.130 | 0.114 |
  | strings | 0.218 | 0.216 |
  | interp | 0.200 | 0.187 |
  | particles | 0.129 | 0.107 |
  | compiler | 0.353 | 0.247 |

  Peak memory of the compiler was 123 MB on base and 125 MB on the branch.
- Mutant compilers, each with one copy removed, run on the three samples with `-sanitize`:
  - tail (the parameter points at the jump's temporary): `params_pointer_tail` prints the same name on every step, then heap-use-after-free.
  - generator (the frame keeps the caller's pointer): `params_pointer_escape` prints `bb b-tag 2` instead of `aa a-tag 1`, then heap-use-after-free.
  - closure (the environment keeps the caller's pointer): `params_pointer_escape` prints `ff/f-tag/4` instead of `ee/e-tag/3`, then an UBSan error.
  - The other samples pass under each mutant.

## Deviations

- Files outside the shared list:
  - `selfhost/src/check.nra`: the layout cache and `isPassedByPointer`.
  - `selfhost/src/cnames.nra`: signatures and `parameterCType`.
  - `selfhost/src/emitgenerators.nra`: the frame copy.
  - `selfhost/src/emittraits.nra`: `&a` when a derived helper calls a user implementation.
- Seed not regenerated (worker rules). Bootstrap says the seed is older than the source; the orchestrator regenerates it.
- `sh scripts/linux-check.sh` not run; the batch runs it once.
- Wall-clock bench numbers were taken under load 5 to 10. Instructions retired are the numbers to trust.

## Stubs left

- None. Reads of fields and elements still copy; D50 also covers those reads, and that part is not built.

## Gotchas

- `isPlace` judges the shape of the emitted C text. A non-place text made only of names, `.`, `->`, `)` and `(*` would be passed by address, and that would be unsound. No emitter makes one today. A place with another shape is only copied, which is slower but safe.
- The derived helpers (`eq_`, `hash_`, `cmp_`, `show_`) still take large values by value. They are a separate copy cost.
- A tail-loop function now declares `<param>_held` for each large parameter. C local names end in `_<number>`, so the names cannot collide.
- `isPointerParameter` is false inside a generator body, because the frame holds a value. Code that names a parameter's C variable directly must check it, as `emitGenerator` and `emittail.nra` do.

## Proposed learnings

- Choose a calling convention from the type alone | function values, thunks and direct calls must agree without knowing each other | selfhost/src | yes (fixed point and the samples)

## Verification

- `sh selfhost/bootstrap.sh out/final` after the rebase: stage 2 and stage 3 emit identical C (164332 lines).
- `NARRAC=<branch narrac> sh scripts/test.sh` before the rebase: PASS (5843 cases and the bootstrap job).
- Fresh review, one round: no soundness bug and no missed call site. The reviewer probed argument order, loop elements, E0614 cases, escapes, tail swaps, temporaries in conditions, maps, constants, spreads and perceus reuse. Fixed from it: renamed `parameter` to `parameterType` and `rcFunctionValue` to `formatRcValue` in `emitdecls.nra`. Kept: byte numbers in `isPlace` (the repo compares `Str.byteAt` to ints), and `parameterCType` (named like `cType`).
- `sh scripts/test.sh` on the final tree, after the rebase and the review fixes: PASS in 97 s, 8 shards with 5853 passed and 0 failed, and the bootstrap job passed.

## Size

- Estimated M, actual M.
