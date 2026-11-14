# Outcome: inline-benchmark

## Built

- `bench/inline/inline.nra`: a micro-benchmark that calls `Options.map`, `andThen`, `orElse` and `withDefault` 100 million times in one loop, with a capturing arrow, a named function and a closure fallback. Not added to `bench/run.sh`.
- No compiler change is committed. The measurement used a local patch (`isInlineFunction` returns `false`, parameters renamed `_checker`, `_funcIndex`), reverted before the commit.
- `@inline` sites: 3 (`map`, `andThen`, `orElse` in `std/src/option.nra`), 0 in `selfhost/src`. No file in `selfhost/src` or `bench/` calls those three functions. Machinery: `selfhost/src/inlining.nra` (371 lines) and `emitinline.nra` (458 lines), 829 lines in all.

## Result

- result: missed. The hypothesis was a gain outside the noise; `@inline` buys nothing measurable.
- numbers: instructions and cycles, median of 7 runs (5 for `emit selfhost`), runs of the two variants interleaved, 1-minute load 2.3 to 2.5 throughout. Gain is `without / with - 1` (positive means `@inline` helps). Spread is (max - min) / median.

| program       | instr with | instr without | instr gain | instr spread | cycles with | cycles without | cycles gain | cycles spread |
| ------------- | ---------- | ------------- | ---------- | ------------ | ----------- | -------------- | ----------- | ------------- |
| fib           | 1212.9M    | 1212.9M       | 0.00%      | 0.04%        | 373.6M      | 373.5M         | -0.02%      | 1.7%          |
| sieve         | 495.7M     | 497.4M        | +0.34%     | 6.3%         | 440.1M      | 436.8M         | -0.73%      | 10.5%         |
| matmul        | 2405M      | 2404.7M       | -0.01%     | 0.04%        | 358.1M      | 357.0M         | -0.29%      | 4.0%          |
| strings       | 1095.3M    | 1095.4M       | +0.01%     | 0.03%        | 732.5M      | 734.0M         | +0.21%      | 35%           |
| interp        | 1963.4M    | 1963.4M       | 0.00%      | 0.05%        | 625.1M      | 625.1M         | -0.01%      | 1.1%          |
| particles     | 1512.2M    | 1512.2M       | 0.00%      | 0.1%         | 224.2M      | 224.2M         | 0.00%       | 8.4%          |
| dot           | 329.4M     | 329.6M        | +0.04%     | 0.4%         | 110.6M      | 110.7M         | +0.05%      | 4.1%          |
| inline        | 1911.5M    | 1911.5M       | 0.00%      | 0.03%        | 1126.1M     | 1126.2M        | +0.01%      | 0.3%          |
| emit selfhost | 3501.3M    | 3497.0M       | -0.12%     | 1.6%         | 1242.4M     | 1230.4M        | -0.96%      | 4.6%          |

- Every gain is inside its spread. Programs with a spread under 0.5% on instructions show 0.00% to 0.04%.
- The emitted C of fib, sieve, matmul, strings, interp, particles and dot is byte-identical with and without the attribute: none of them calls an `@inline` function. Their rows measure only the noise between two separately built binaries.
- `inline` is the one program where the attribute changes the emitted C. With `@inline` the three calls are copied into `run`. Without it, the C has three ordinary generic functions and three closures, and clang -O2 inlines them. The instruction counts agree to the digit (1911.5M), so the machine code is the same and `@inline` gains 0% here too.
- `emit selfhost`: the compiler source has no `@inline` call, so both compilers run the same code. The rows measure layout noise (instructions 1.6% spread, cycles 4.6%).
- narrac size: 5,134,056 bytes with, 5,116,712 bytes without (-0.3%). Both come from the same source apart from the patch, so part of the difference is layout; the inliner is about 830 lines of Narra.
- reproduce: branch `spike/inline-benchmark` on `feat/batch-15` at aa6720f2. Build each narrac with `sh selfhost/bootstrap.sh <out>` (without: the patch above applied). Per program, `narrac build -release`, then `/usr/bin/time -l <exe>` and read `instructions retired` and `cycles elapsed`; `/usr/bin/time -l <narrac> emit selfhost` for the compiler row. Scratch in `/private/tmp/claude-501/inline-benchmark/` (`measure.sh`, `st.pl`, `raw-*.txt`).

## Recommendation

Remove `@inline`. Reasons:

- The only users are three std functions that nothing in the repository calls, and clang -O2 inlines their copies anyway (identical instruction count on a loop that calls all three).
- No bench program and not the compiler gains anything measurable.
- It costs about 830 lines of compiler, two error codes, and a second code path through the emitter (labels, arrow copies, value uses) that every emitter change must keep correct.
- Limit of the evidence: the micro-benchmark uses small bodies that clang inlines. A callee clang would refuse to inline (large, or called from many sites) could differ; no such `@inline` function exists and none was measured.

A follow-up outline brief is at `nodes/remove-inline/brief.md`. The maintainer decides before any removal (gate).

## Deviations

- The brief asked for the compiler's hot `@inline` helpers: there are none, so the compiler row measures noise. A micro-benchmark (`bench/inline`) stands in for the std functions.
- `bench/CLAUDE.md` asks for a load under 2 to record a run. The task set the limit at 4; the load was 2.3 to 2.5. The numbers are not for `BASELINE.md`.
- Each program was built once per variant, so the C-identical programs differ only in link layout.

## Stubs left

- None.

## Gotchas

- `@inline` does more than speed; removal must decide each of these:
  - Arrow arguments: an arrow passed to an `@inline` function is copied to each place the body calls the parameter, reads the caller's variables in place, and needs no closure environment. Without `@inline` the same arrow is an ordinary non-escaping closure with a stack environment. The non-escaping rule for function-type parameters does not change.
  - Effects and `from`: they are checked on the declaration, so `uses` rows and `from` markers do not change.
  - Labels: copies use `narra_inline_N` jump labels and the `_iN` / `_aN` label suffix (`labelSuffix`) so two copies in one C function do not clash; `return` inside the body becomes a jump. All of it goes with the emitter path.
  - Errors: E1030 (recursion chain), E1031 (generator, async or `extern` member). E0990 (attribute in an `implement` block) stays for other attributes.
  - A program that writes `@inline` after removal gets the parser's unknown-attribute error, with no migration message (the project rule for unreleased syntax).
- An `@inline` function used as a value already gets an ordinary C function.

## Proposed learnings

- Measure a language-level inlining feature against the C compiler at -O2 before building it | clang inlines small static functions anyway; `@inline` gained 0.00% on a loop of its own std helpers | selfhost/src | no

## Verification

- Both compilers: `sh selfhost/bootstrap.sh` reaches the fixed point (stage 2 and stage 3 C identical; 249,695 lines with, 249,689 without).
- `cmp` of the emitted C per program with and without: 7 identical, `inline` differs as described.
- The patch is reverted; `git status` shows only `bench/inline/` and the node folders.

## Size

- Estimated S, actual S.
