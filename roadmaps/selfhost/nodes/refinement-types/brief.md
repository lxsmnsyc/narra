---
id: refinement-types
title: "Refinement types on distinct types"
status: active
priority: must
type: feature
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/refinement-types/**
  - selfhost/src/refinements.nra
  - testdata/refinement_*
  - testdata/errors/refinement_*
shared:
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/trivia.nra
  - selfhost/src/querysignature.nra
  - selfhost/src/types.nra
  - selfhost/src/declare.nra
  - selfhost/src/calls.nra
  - selfhost/src/contracts.nra
  - selfhost/src/operands.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/consteval.nra
  - selfhost/src/deriveshape.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E1220-E1223]
provides: []
depends_on:
  - on: refinement-types-design
    kind: hard
    why: builds D116; starts only after the maintainer approves it
---

# Refinement types on distinct types

## Goal

Build D116: a distinct type over an integer type or `string` may carry `requires` and a list of `ensures`-style patterns joined by `&&`. `T.from` is proven at compile time, `T.check(value): Option<T>` tests at run time, unwrapping a refined integer gives the prover the type's bounds, and `derives Shape` turns a failed pattern into `Err`.

```ts
export distinct type Port = int derives Show, Shape requires >= 1 && <= 65535
export distinct type Method = string requires "GET" | "POST"

const DEFAULT_PORT = Port.from(8080)        // proven from the literal

function share(total: int, port: Port): int {
  return total / Port.toInt(port)            // 1 <= n <= 65535, so the divisor is proven
}

function parsePort(raw: int): Option<Port> {
  return Port.check(raw)                     // run-time test
}
```

## Context

- D116 in DECISIONS.md (decision, codes, rejected forms). SPEC section 4, "Refinements" under "Distinct types" (written by the spike, marked not built yet). D95 and D102 for the `ensures` patterns and facts; D108 for distinct types.
- Parser: `parseOpaqueOrDistinct` (`selfhost/src/parser.nra:1514`), `parseEnsuresLines` and `parser.contractPattern` (the flag that allows `PatCompare`), `isContractLine`. A comparison bound is `parseUnary`, so it stops at `&&`.
- AST: `DistinctDecl` (`ast.nra:267`) gets `refinement: Array<Pattern> = []`. Types: `DistinctInfo` (`types.nra:135`) gets the resolved list.
- Checker: `checkDistinctCall` (`calls.nra:1271`) handles `from` and the unwrap; `findUnwrapName`. The `ensures` prover: `isValueProven`, `collectBounds`, `checkComparisonBounds`, `isFittingBound` (`contracts.nra`). Facts: `findOperandFact` (`operands.nra:762`), `seedResultLocal` (`contracts.nra:1535`), `findPatternFact` and `findAlternativesFact` for the hull.
- Emitter: `emitexprs.nra:624` emits every `Ref.Distinct` call as its argument; `check` needs its own case. `consteval.nra:2847` does the same in the evaluator.
- `derives Shape`: `writeDistinctFromValue` (`deriveshape.nra:392`) returns `Ok(T.from(base))` today, which an unproven `from` would reject.
- Evidence the feature is missing: with the batch-13 base, `distinct type Port = int requires >= 1` is a parse error, and `slot(Port.toInt(port))` against `requires index >= 1` is E0985.

## Steps

- [ ] **Syntax.** After the trait lists, `requires` (a word, not a keyword, as in functions) on the same line or the next starts the list. Parse items with `contractPattern` set, joined by `&&`. Store them in `DistinctDecl.refinement`. A missing item, `&&` with nothing after it, a second `requires`, or `requires` before `implements` or `derives` is E1220.
- [ ] **Tools.** `astdump` prints a `.refinement` slot; `fmt` prints ` requires a && b` after the trait lists and keeps a line break before `requires`; `trivia` gives `requires` a node of its own so a comment beside it stays (selfhost/CLAUDE.md); `querysignature.nra:210` and hover print the list.
- [ ] **Declaration check** (new module `selfhost/src/refinements.nra`, with only dispatch lines in shared files). Resolve the base; anything but an integer type or `string` is E1221 (another distinct type included). Check each item against the base with the `ensures` rules (`checkComparisonBounds`, `isFittingBound`): on `string` only literals and or-patterns of literals; everything D116 part 1 lists is E1222. Store the items and, for an integer base, the hull fact in `DistinctInfo`.
- [ ] **`from`.** In `checkDistinctCall`, for a refined type, prove the argument against each item with `isValueProven`. Each failed item is listed in one E1223 at the argument: `Port.from needs raw to satisfy <= 65535; check it first or call Port.check`, or with a constant `70000 does not satisfy Port (<= 65535)`. Items print as written (`describePattern`).
- [ ] **`check`.** A refined type accepts `check` beside `from` and the unwrap; it returns `Option<T>`. E0322 stays for `check` on a type without `requires` (update its format to list `check` only when it exists). E0321 covers `check` on an opaque type outside its module. The emitter writes the items as one C condition (comparisons, `||` between alternatives, the runtime's string equality for string literals) and builds `Some` or `None`. The `@const` evaluator evaluates the same test.
- [ ] **Facts.** `findOperandFact` returns the type's hull for a call of the unwrap function of a refined integer type, whatever the argument. `seedResultLocal` seeds `const n = Port.toInt(port)` (and `let` until assigned) with it. `!= 0` gives non-zero; a string refinement gives nothing.
- [ ] **`derives Shape`.** `writeDistinctFromValue` calls `check` for a refined type and returns `Err` with `<value> does not satisfy <Type> (<items>)`, the value written as `Show` writes it (a string quoted). The path prefix comes from the caller, as for every Shape error.
- [ ] **Docs.** SPEC: drop the "Not built yet" line under "Refinements" and turn the inline samples into checked `ts` blocks (with `<!-- error: E1223 -->` for the failing one). DECISIONS.md: a "Built" paragraph under D116 with any amendment. `compiler/README.md`: the feature list.

## Acceptance

- [ ] Proof first: `testdata/refinement_types.nra` fails to compile on the batch base (`git stash` is not allowed; check it out in a scratch worktree).
- [ ] `testdata/refinement_types.nra` with `.out`: `Port` and `Method` declared; `Port.from` of a literal, of a top-level `const`, of a name inside `if (raw >= 1 && raw <= 65535)` and inside a function with `requires raw >= 1 && raw <= 65535`; `Port.check` of 0, 1, 65535 and 65536 printed; `Method.check("GET")` and `Method.check("PUT")`; a divisor `total / Port.toInt(port)` and a shift with a `u8` refinement `requires < 8` proven through a struct field; `const n = Port.toInt(port)` passed to `requires n >= 1`; an `ensures >= 1` function returning `Port.toInt(port)`. Clean under AddressSanitizer, the leak check and the copy check.
- [ ] `testdata/refinement_shape.nra` with `.out`: a struct with a `Port` field and a `Method` field decoded with `Json.decode`; a good value, `.port: 70000 does not satisfy Port (>= 1 && <= 65535)`, and `.method: "PUT" does not satisfy Method ("GET" | "POST")`; `Json.encode` round trip unchanged.
- [ ] A `@const` function that calls `Port.check` gives the same results as at run time.
- [ ] Error tests, one file each under `testdata/errors/`: `refinement_syntax` (E1220: `requires` with no item, a trailing `&&`, `requires` before `derives`), `refinement_base` (E1221: `f64`, `char`, a struct, another distinct type), `refinement_pattern` (E1222: `>= 1` on `string`, a non-constant bound, `<= 300` on `u8`, a u64 bound of 2^63, `Some(_)`, `_`, a name), `refinement_from` (E1223: an unchecked parameter, `70000`, a name checked only at the low end, `Method.from(text)` of a variable, the message naming the failed items), `refinement_opaque` (E0321 for `check` outside the module), `refinement_check_plain` (E0322 for `check` on a type without `requires`).
- [ ] `astdump`, `fmt` (with a stable-format case and a comment beside `requires`), `narrac query` signature and hover cover the new syntax; the `sources`, `spec` and `codes` suites pass.
- [ ] `narrac emit selfhost` instructions do not rise beyond the noise (no refined type in the compiler).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

L: one new syntax form across parser, AST, fmt, astdump, trivia and query, plus checker, emitter, evaluator and derive changes; each reuses an existing prover or emitter part. Not split: the Shape step is about 20 lines and needs `check` first.

## Notes for the worker

- Codes: E1220 parse, E1221 base, E1222 item, E1223 unproven `from`. E1224-E1239 stay free for later D116 work. Add the rows at the end of the E1200 range in `diagnostics.nra`, in number order, beside other batch-13 rows; the header range line is already there.
- No `Pattern` variant is added. `&&` joins items in the list only; `Some(>= 1 && <= 9)` stays a parse error.
- No distinct function is a value today (`const unwrap = UserId.toInt` is `unknown name`); keep it so. Do not add `Dynamic`, generic or `match` exhaustiveness work for refined types (D116 rejects them).
- `Port.toInt(port) + 1` has no fact; do not add arithmetic facts.
- The compiler and std do not use the feature (seed rule).
