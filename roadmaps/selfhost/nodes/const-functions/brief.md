---
id: const-functions
title: "@const functions: top-level constants computed by the compiler"
status: todo
priority: should
type: feature
story: batch-11
early: false
outline: false
milestone:
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
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/const-functions/**
  - selfhost/src/consteval.nra
  - selfhost/src/constfunctions.nra
  - selfhost/src/emitconstvalues.nra
  - testdata/const_functions*
  - testdata/errors/const_eval_*
shared:
  - selfhost/src/parser.nra
  - selfhost/src/exprs.nra
  - selfhost/src/check.nra
  - selfhost/src/emit.nra
  - selfhost/src/emitconsts.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/emitchecks.nra
  - testdata/astdump/**
  - testdata/declsdump/**
  - testdata/fmt/**
  - SPEC.md
  - compiler/README.md
reserves: [diag E1050, diag E1051, diag E1052, diag E1053, diag E1054, diag E1055, diag E1056, diag E1057]
provides: []
depends_on:
  - on: const-eval-design
    kind: hard
    why: D105 fixes the rules, limits and C output this node builds; it must be approved first
---

# @const functions: top-level constants computed by the compiler

## Goal

A top-level `const` may call a function marked `@const`, and the compiler computes the result while it compiles (D105). Today the attribute does not parse and the call is rejected:

```
attr.nra:1:2: expected a name, found "const"
plain.nra:5:20: constant "area" cannot call function "square"; a top-level constant holds literals, constructors, structs, arrays, function names and other constants
```

(E0012 with `@const`, E0192 without it, `narrac check` on `feat/batch-11` 20f9a487.) The node adds the attribute, the declaration checks, a tree-walking evaluator over the checked AST, and a writer that turns the result into C through the D61 getter.

```ts
import * as Arrays from "std/array"
import * as Ranges from "std/range"

@const
function crcTable(): Array<int> uses alloc {
  let table: Array<int> = []
  for (const entry of Ranges.upTo(256)) {
    let value = entry
    for (const _bit of Ranges.upTo(8)) {
      if ((value & 1) != 0) {
        value = 3988292384 ^ (value >> 1)
      } else {
        value = value >> 1
      }
    }
    Arrays.push(mutable table, value)
  }
  return table
}

const CRC_TABLE = crcTable()
```

## Context

- Source: D105 in `DECISIONS.md` (all ten rules; read it first). SPEC section 10 (around line 1596, "a call to an `@const` function") and section 12 (around line 1633). D61 for the getter, D58 for effects, D92 for the panics that remain.
- Measurements behind the limits and the C writer: `roadmaps/selfhost/nodes/const-eval-design/outcome.md`. The CRC-32 table above is 26,370 steps.
- Attribute: `KNOWN_ATTRIBUTES` and `parseAttribute` in `selfhost/src/parser.nra` (around line 1357). `const` is a keyword, so `parseAttribute` must accept it after `@` (`expectIdent` fails today). `addAttributes` already keeps attributes on `FuncDecl`; `findAttribute` (used in `effects.nra` and `audit.nra`) reads them. fmt and astdump print attributes by name already; add goldens.
- Constant check: `checkConstCall` in `selfhost/src/exprs.nra` (around line 2895) accepts only `Ref.Constructor`. Add `Ref.Function(index)` when that function has `@const`; its arguments go through `checkConstValue` as constructor arguments do. Keep the E0192 text for other calls.
- Body rules (`selfhost/src/constfunctions.nra`, new): run after effects are known (`effects.nra` fixed point), once per `@const` function, even when no constant calls it.
- Evaluator (`selfhost/src/consteval.nra`, new): walks the checked AST using the body's side tables (`findBody`, `BodyInfo.refs`, `first`). Values are a `ConstValue` enum over Narra values (`int`, sized numbers, `float`, `bool`, `char`, `string`, `bytes`, `Array<ConstValue>`, an ordered map, struct fields, enum tag and payload, tuples). Supported std rows call the same std function inside the compiler. Before each call the evaluator checks the row's panic preconditions (`Str.sliceFrom`, `charAt`, `nextChar` call `narra_panic`, which would end the compiler) and charges the memory budget from the arguments (`Arrays.filled`, `withCapacity`, `Str.repeat`, `join`, `replace`, concatenation, templates), so a failing or oversized call never runs (D105 items 5 and 6). Std source functions of the supported modules (`std/src/ext__*` files) run as ordinary bodies; `Arrays.map` takes a function and is E1052.
- C writer (`selfhost/src/emitconstvalues.nra`, new): a scalar result is static and written at each use (extend `isStaticConstant` and `constRef` in `emitexprs.nra`, around lines 1263 and 1319). Any other result goes behind the getter of `emitconsts.nra`, whose body the writer fills from the value instead of the AST. An array of scalars is a `static const` C array plus one `memcpy` into the counted array. A float that is not finite is written as `NAN`, `INFINITY` or `-INFINITY`, never as the printed `nan` or `inf`.
- Determinism: `emit.nra` (around line 87) writes `#pragma STDC FP_CONTRACT OFF` before `RUNTIME_H`. This changes every emitted C file, so the seed and the C goldens of the emit checks change at the batch's seed update.
- No threads exist, so the getter's `_ready` flag stays the once guard, as for D61.

## Error codes

Reserved: E1050 to E1057. E1058 and E1059 stay spare.

- E1050: an `@const` function's signature uses a type D105 item 3 does not allow (function type, `Cell`, `Ptr`, `Shared`, a generator, `Async`, `Dynamic`, a channel), has a trait-constrained type parameter, or is a generator or async function.
- E1051: an `@const` function's effects go beyond `alloc`; the message names the effect and the call or expression that brings it.
- E1052: an `@const` body calls a function that is neither `@const` nor an allowed std function (this includes `Math.sin`, `cos`, `tan`, `atan2`, `exp`, `log` and `pow`, D105 item 9, the functions that take a function, such as `Arrays.map`, `Options.map` and `Results.map`, and any std module outside this node's list), reaches a user `implement` block, uses an arrow function, or calls through a function value. Built-in and derived implementations (`==`, map keys, templates, `Ordered` on numbers and strings) run.
- E1053: evaluating a constant reached `panic`, a failed `assert`, an index out of range or a std panic. The message has the panic text; the call chain follows as notes, one per call from the constant down.
- E1054: the step budget (10 million) ran out.
- E1055: the memory budget (64 MiB, counted as D105 item 5 says) ran out.
- E1056: more than 1000 nested calls. Self-calls in tail position run as a loop and do not count.
- E1057: a constant reaches itself through an `@const` function that reads it.

## Steps

- [ ] `@const` parses on a function declaration; fmt, astdump and declsdump goldens.
- [ ] `constfunctions.nra`: E1050, E1051, E1052 at the declaration.
- [ ] `checkConstCall` accepts `@const` calls; E1057 for cycles, next to E0191.
- [ ] `consteval.nra`: statements (`const`, `let`, assignment and compound assignment, `if`, `while`, C-style `for`, `for ... of` over `Range`, `Array`, `Map` and `FixedArray`, `match`, `return`, `break`, `continue`), every expression kind in item 3, patterns, `mutable` arguments as places, tail self-calls as loops, the budgets.
- [ ] Std rows of `std/string`, `std/array`, `std/map`, `std/int`, `std/float` and `std/range` (45 registry rows, less `Arrays.map`). List the supported modules in SPEC; a call into any other std module is E1052 with a message that names the function. The rest of D105's list (`std/char`, `std/option`, `std/result`, `std/math`, `std/bytes`, `std/fixedarray`, the sized number modules) is a follow-up node for the orchestrator to add (suggested id `const-functions-std`).
- [ ] `emitconstvalues.nra` and the hooks in `emitexprs.nra` and `emitconsts.nra`; the pragma in `emit.nra`.
- [ ] Samples, error tests, SPEC and README.

## Acceptance

- [ ] Samples (`testdata/const_functions*.nra` with `.out`): the CRC-32 table; `square(4)` used as a scalar; a keyword `Map<string, int>` built in a loop; a struct and an enum result; a string built with templates and `std/string`; an `@const` function that calls another and one that reads a top-level constant; a float table (`Float.sqrt` of 0 to 15, and `a * b + c` on floats); a constant holding `Float.sqrt(-1.0)`, `1.0 / 0.0` and `-1.0 / 0.0` (the C has `NAN`, `INFINITY` and `-INFINITY` and compiles); an `@const` function from another module; an `@const` function also called at run time.
- [ ] Run-twice test: each sample prints its constant and the same call made at run time (pass the arguments through a local so the call is not a constant), and the two lines are equal in the `.out`.
- [ ] Every sample is clean under the leak check and sanitizers; `narra_consts_release` frees each built value.
- [ ] Emit checks in `selfhost/tests/src/emitchecks.nra`: a scalar result appears as a literal at its use and has no getter; the CRC table's C holds a `static const` array and a `memcpy`, with no store per element; the emitted C starts with `#pragma STDC FP_CONTRACT OFF`; the same program emits byte-identical C twice.
- [ ] Error tests (`testdata/errors/const_eval_*`): a call without `@const` (still E0192); E1050 for a function-typed parameter and for a `Cell` result; E1051 for a body that prints; E1052 for a call to a plain function, for `Math.sin`, for `Arrays.map`, for a user `implement` block reached through `==` and for an arrow; E1053 for `assert(false)`, for `items[10]` on a short array and for `Str.charAt` on a bad position (the compiler reports it and does not panic), each with its call chain; E1054 for a `while (true)` loop; E1055 for `Arrays.filled(0, 100000000)` (the compiler's peak memory stays small); E1056 for a recursion that is not a tail call; E1057 for a constant read by its own `@const` function.
- [ ] `narrac check` reports evaluation errors (E1053 sample run through `check`, not only `build`).
- [ ] SPEC.md section 10 states the rules of D105 items 2, 3, 5, 6 and 8; section 12 notes that evaluation happens in the compiler with the budgets; section 9 or 11 notes the FMA pragma. `compiler/README.md` drops `@const` functions from "Not supported yet" and lists what stays out (arrow functions and function values in `@const` bodies, any std row not supported).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

L. Rubric scores that are not low: new code (about 1500 to 2000 lines: an evaluator for most of the language, 45 std rows, a C writer), cross-cutting (parser, checker, emitter, dumps), and a second semantics to keep equal to the emitted C. Integration with existing code is low: D61's getter, the effect fixed point and the attribute list already exist.

## Notes for the worker

- Do not use `@const` in `selfhost/src` or `std/src` in this node (seed rule).
- `consteval.nra` and the other new modules must compile under the D104 rule with no E0978, and the `copies` suite must stay at exit 0. Thread evaluator state through one `mutable` struct and keep its arrays unique (selfhost/CLAUDE.md).
- The inline node also adds a row to `KNOWN_ATTRIBUTES` and lazy-values also changes `exprs.nra` and `emitconsts.nra`. The second to land rebases and keeps both sides.
- The evaluator runs on the compiler's C stack, which overflows at about 20000 Narra frames. Measure the frames one interpreted call takes; if 1000 nested calls do not fit, lower the depth limit and say so in the outcome and D105.
- Do not build closures or function values in the evaluator. D105 leaves them out of the first build.
- Do not emit static C data for counted values. The getter still builds the counted array at first use; only the scalar elements come from a `static const` array.
- The run-twice test is the check that the two semantics agree. When it fails, fix the evaluator, not the sample.
- Open question for the worker: strings in a `static const` table. Recommended: leave arrays of strings to element stores for now; string literals are already static `lstr` values, so a later change can point at them.
