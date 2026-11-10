# Outcome: option-inline

## Built

- An enum with fields is a tagged union held by value, `struct E { int32_t tag; union { struct {...} v_V; ... } u; }`, unless it reaches itself or its payload is over 32 bytes. Those keep the counted heap payload of D55. Fieldless enums stay `{ tag }`.
- Layout decision: `decideHeapEnum` in `selfhost/src/check.nra`. `finish` stores it per enum instance in `Checker.heapEnums`; `isHeapEnum` reads it (an enum made after `finish` falls back to the heap layout). Recursion test `reachesEnum` follows structs, enums and distinct types held by value. Size estimate `measurePayload` (largest variant, fields rounded up to 8 bytes) against `INLINE_PAYLOAD_LIMIT = 32`. Decisions and sizes are memoized per enum and struct (`LayoutMemo`), so the walk is linear and a struct that holds itself (reported later as E0109) ends it.
- `finish` now runs the layout decision before `visitOrder`; an inline enum is ordered after its field types, like a struct.
- Counted-type rule moved to `isCountedType` in `check.nra`, with caches `Checker.countedStructs` and `countedEnums` filled by `finish` when the package has no errors. `needsRc` (`emitstate.nra`) calls it. Without the cache, the field walk for inline enums on every `needsRc` cost more than the layout saved.
- Emitter: `buildVariantFieldPath` (`emitstate.nra`) writes `.u->v_V.f_x` or `.u.v_V.f_x`; used by field reads (`emitexprs.nra`), patterns (`emitmatch.nra`), Option and Result helpers (`emitoptions.nra`) and derived traits (`emittraits.nra`). `emitdecls.nra`: the inline type definition, `inlineEnumRcFunctions` (retain and release switch on the tag and count that variant's fields), and `enumAllocation` builds an inline value as a compound literal with no allocation.
- `consume` no longer retains a fieldless constructor (`isTagOnly`), for heap and inline enums.
- Docs: D99 written, D55 amended, SPEC section 6 updated, `compiler/README.md` (layout line moved from "Differences from the spec" to "Supported"), ownership comment in `emitstate.nra`.
- Samples: `testdata/enum_inline_options` (Option of int, string and a struct with counted fields; copies outliving a replaced original; nested options and nested matches; equality; map lookups), `enum_inline_results` (Result with counted fields on both sides, an inline enum with a guard, copies in arrays and a closure, a binding whose subject is replaced, destructuring from an inline payload), `enum_inline_recursive` (a recursive enum and an `Option` that reaches it, a struct-enum cycle, two enums that reach each other, an inline enum holding a heap enum, a 10-string payload over the limit, a 2000-deep chain freed), `enum_inline_generators` (inline enums held across yields, generators dropped while suspended).

## Proof

Instructions retired, median of 5, base and branch compilers built by `selfhost/bootstrap.sh`, macOS arm64, load 5 to 10.

| program | base (feat/batch-4 8de99b4) | branch |
| --- | --- | --- |
| compiler (`narrac emit selfhost`) | 3.281 G, 130 MB | 3.084 G (-6.0%), 125 MB |
| fib | 1.213 G | 1.213 G |
| sieve | 2.305 G | 2.291 G |
| matmul | 2.447 G | 2.446 G |
| strings | 1.096 G | 1.095 G |
| interp | 1.936 G | 1.964 G (+1.4%) |
| particles | 2.612 G | 2.612 G |

- fib, sieve, matmul, strings and particles compile to byte-identical C on both sides; their differences are noise.
- User time, interleaved: compiler 0.26 s to 0.24 s (9 runs), interp 0.15 s to 0.18 s (15 runs).
- interp regresses. Its `Op` payload moved into the value, so clang can load it before the tag test and turns the `Add`/`Sub` arms into `csel`, which lengthens the dependency chain on the accumulator. A hand-written `switch` and a pointer read gave the same or worse time. Recorded as the cost in D99.
- Option/Result-heavy code (scratch loop, 20 M `Option<int>` and `Result<int, string>` returns and matches): 1.767 G to 0.440 G instructions, 0.09 s to 0.02 s.
- Size limit sweep on e64faa2 (compiler instructions, with the cache): all heap 3.53 G, 8 bytes 3.48, 16 3.46, 32 3.45, 48 3.51, 64 3.56, no limit 3.67 (3.87 without the cache). The selfhost source has inline enums up to 808 bytes (`Option<Emitter>`). 32 bytes is the minimum.
- Soundness, stage 1 mutants (base compiler on mutated source): inline retain emits nothing, all four samples read freed memory under ASan; inline release emits nothing, all four fail the leak check (9 to 60 live allocations).

## Deviations

- `selfhost/src/check.nra` (not shared): the layout decision, the counted-type rule and caches, and a fix in `visitOrder`. Base bug: `struct Node { next: Link }` declared before `enum Link { Next(node: Node) }` was E0109 "contains itself", because a heap enum visited its fields while the struct holding it was still open. A heap enum now skips fields whose type is still being visited. `enum_inline_recursive` covers it.
- `selfhost/src/emittraits.nra` (not shared): derived traits read payloads through `variantFieldPath`.
- `testdata/declsdump/nested_generics.dump` and `shadow_prelude.dump` (not owned): an inline `Option` is now ordered after its field type, and its impl lines move with it.
- The seed is not regenerated (worker rules); the orchestrator does it.
- `bench/run.sh` wall times not recorded: the machine never reached load 2. Instructions retired and interleaved user time are reported instead.
- `sh scripts/linux-check.sh` not run; the batch runs it once.

## Stubs left

- None.

## Gotchas

- `isTagOnly` recognises a fieldless constructor by its C text, `((Name){.tag = N})` from `constructorValue`. Changing that text shape turns the retain back on (safe, slower).
- `isHeapEnum` and the counted caches are filled by `finish`. Code that builds enum or struct instances after `finish` gets the heap layout and the uncached walk.
- perceus-reuse works only on heap payloads now; most `Option` and `Result` instances are inline.
- An inline enum value is copied by value like a struct, so const-params-pointer applies to it too.

## Proposed learnings

- Cache per-type facts the emitter asks on every retain or release | an uncached field walk in `needsRc` cost more than inline payloads saved | selfhost/src | no
- Measure a layout change on the hot loop of each bench, not only instructions | interp gained no allocations and lost 20% to clang if-conversion with 1.4% more instructions | bench | no

## Verification

- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C (162761 lines).
- `NARRAC=<branch narrac> sh scripts/test.sh --suite programs`: PASS.
- Fresh review (one round): no soundness bug found; the reviewer's own sanitized programs (arrays, maps, generators, closures, constants, std Option/Result helpers, derived traits) ran clean. Fixed from it: the size walk hung on `struct S { a: S, b: S }` (invalid program) and was exponential in `Option` nesting (1.6 s at depth 28); now memoized (0.08 s at depth 28, the invalid program reports E0109 at once). Renamed `variantFieldPath` and `variantMembers` to verbs. Kept: `isTagOnly` matching C text (see Gotchas).
- After the review fixes: compiler 3.285 G base, 3.081 G branch (median of 5).
- `sh scripts/test.sh` on the final tree: PASS in 180 s, 8 shards with 5828 passed and 0 failed, and the sanitized bootstrap job passed; stage 2 and stage 3 emit identical C (162826 lines).

## Size

- Estimated L, actual L. The size limit and the counted cache took most of the measuring.
