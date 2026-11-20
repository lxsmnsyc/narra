# Outcome: spec-corrections

## Built

`SPEC.md` corrections, each checked with the prebuilt narrac:

- Section 2, Control flow: the `for ... of` line and the "An iterator is a `Generator<T>`" bullet. They now say a loop reads an array, a `Range` or a generator. `Maps.keys`, `Maps.entries` and `Str.chars` return arrays, a `Range` is its own type, and `Arrays.values` and `Arrays.entries` stand only at the head of a loop (E0920). Evidence: `check` accepted `Array<string> = Maps.keys(m)`, `Array<char> = Str.chars(..)`, `Array<#[string, int]> = Maps.entries(m)`. It rejected `const a = Arrays.values(items)` (E0920) and passing an array or a `Range` as `Generator<int>`.
- Section 4, Constant type parameters: a struct literal does not work out `N` from its fields. `Schedule { hours: [1, 2, 3] }` fails ("cannot tell constant parameter N"); `Schedule<3> { ... }` and a literal with an expected type compile.
- Section 8, Threads sample: it called undefined `sum`, `firstHalf` and `secondHalf`. It is now a module with `sumRange` and `sumInTwoThreads`. It is checked by the spec suite.
- Section 8, Channels sample: it used an undefined `scope`. It is now a module inside `Thread.scope`. It ran (printed the sum 4950 in an equivalent `run`).
- Section 8, `Channel.values`: a `for` loop over it also needs `alloc`. Evidence: `Function _noAlloc uses alloc (through values) but its signature does not declare it`.
- Section 9, `unsafe`: a `return` inside the block returns from the enclosing function (ran, exit 0); `const x = unsafe { ... }` is a syntax error.
- Section 3 and section 16: `narrac move old new` moved to Deferred as not built. `narrac move a b` prints the usage line and exits 2.
- `selfhost/tests/spec-skip.txt`: removed the Threads and Channels lines, since both samples now check clean (the suite fails on a skip line whose sample passes). This file is outside the node's owned paths; it is required by the suite.

## Proof

- base: `SPEC.md` Threads and Channels samples were listed in `spec-skip.txt` as not checking.
- branch: `NARRAC=<prebuilt> JOBS=2 sh scripts/test.sh --suite spec`: PASS, 65 cases.

## Deviations

- The brief's `%` check for the thread sample does not apply: the spec sample had no `%`; its failure was the undefined names. The new sample uses a range sum.
- The guide-compile-time note (SPEC says E1053 and E1360 while messages show no code) is not a spec error. `narrac explain E1053` and `E1360` match the text, and codes show in `check --json`. No change.
- Not changed, for the orchestrator: `narrac --help` lists only some commands (guide-tools-gaps); the manifest notes (unknown keys, `entry = "app/cli.run"`) are in the inbox.

## Stubs left

- None.

## Gotchas

- Compiler bug: `Thread.scope` with a body that returns no value fails inside std: `std/src/thread.nra:42:3: this expression has no value` (`const result = body(opened)` with `T = void`). Repro: `Thread.scope((scope) => { Thread.start(scope, () => 1) })` as a statement. SPEC says `Thread.scope(f)` returns the result of `f` and states no restriction, so I left the spec text and made the samples return a value. Needs a fix in std/thread or a rule in SPEC.
- Section 9 text says `I32`, `U8` and `CString` names; the guide-c-interop note says the site and TextMate highlighters disagree on them. Not a SPEC issue.

## Proposed learnings

- Give every SPEC.md sample that names helper values its own definitions | a hidden context comment does not show in the rendered spec | SPEC.md | yes, the spec suite

## Verification

- `NARRAC=<prebuilt> JOBS=2 sh scripts/test.sh --suite spec`: PASS, 65 cases.
- `pnpm --dir website build`: passes.

## Size

- Estimated S, actual S.
