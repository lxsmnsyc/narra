---
id: param-destructuring
title: "Destructuring in parameters"
status: dropped
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/param-destructuring/**
  - testdata/param_destructuring*
  - testdata/errors/param_destructuring*
shared:
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/idcheck.nra
  - selfhost/src/trivia.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/body.nra
  - selfhost/src/check.nra
  - selfhost/src/scopes.nra
  - selfhost/src/match.nra
  - selfhost/src/calls.nra
  - selfhost/src/contracts.nra
  - selfhost/src/sharedresults.nra
  - selfhost/src/closures.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitclosures.nra
  - selfhost/src/emitgenerators.nra
  - selfhost/src/emittail.nra
  - selfhost/src/emitmatch.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/querysignature.nra
  - selfhost/src/query.nra
  - selfhost/src/diagnostics.nra
  - std/src/shape.nra
  - selfhost/src/stdtext.nra
  - testdata/astdump/**
  - testdata/fmt/**
  - testdata/declsdump/**
  - testdata/query/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: destructuring-design
    kind: hard
    why: builds D103 part 2, which the maintainer approves at the gate
  - on: array-patterns
    kind: soft
    why: FixedArray patterns as parameters come with whichever of the two lands second
---

# Destructuring in parameters

## Goal

Dropped by D103 (2026-11-12): conflicts with named arguments (D87).

D103 part 2. A struct or tuple pattern in place of a parameter name, in function declarations, arrow functions and functions of `implement` blocks: `function area({ width, height }: Rect): int`, `(#[x, y]: #[int, int]): int => x + y`, `Arrays.map(pairs, (#[key, value]) => ...)` with the type from context.

- Meaning: `const <pattern> = <argument>` at the start of the body. The pattern must be irrefutable (E0715). Bound names are read-only copies (D81).
- `mutable` or `escaping` on a destructured parameter is E0721, with a hint to name the parameter and read its fields.
- Trait function declarations and `extern` functions keep names only; a pattern there is a parse error. An `implement` function may destructure; conformance compares types.
- Diagnostics, `narrac query`, `declsdump` and the language server print the pattern as written. The C parameter gets a generated name.
- Named arguments (D87): a destructured parameter is passed by position. When a call leaves it out, the existing missing-argument error adds that the parameter is destructured and has no name.
- Each bound name follows the unused rule. Shadowing between bound names and other parameters is the usual error.
- `requires` naming a bound name is E0984 (a field of the argument, D95). `from` naming a bound name means the whole argument.
- Self tail calls (`emittail.nra`) jump back above the destructuring, so the names rebind on each pass. Generator and async frames bind them at the first resume, like other locals.
- After array-patterns lands, FixedArray patterns work as parameters too (whichever node lands second adds the case and a sample).

## Context

- D103 in DECISIONS.md; SPEC section 2 "Functions" and "Bindings" get the rule.
- `Param` in `ast.nra` gains an optional pattern; `name` holds a generated name when it is set. `match.nra` already binds a destructuring (`collectPatternLocals`, `bindPatternLocals`); reuse it for the prologue instead of writing a second binder.
- The checker and emitter switches over parameters are spread over the shared files listed; `querysignature.nra` prints signatures.
- `std/src/shape.nra` tuple2ToValue, tuple3ToValue and tuple4ToValue are the three corpus sites. Rewriting them waits for the seed update after this node (seed rule); `std/src` is std text, so the batch that rewrites them needs the std text step (`stdtext.nra`).

## Acceptance

- [ ] Samples: struct, tuple and nested patterns in a function declaration, an arrow function with a written type and one with a type from context, an `implement` function, a generator, a self tail call, a closure capturing a bound name, a named-argument call that passes the destructured parameter by position; all clean under the leak check and sanitizers.
- [ ] Error tests: refutable pattern (E0715), `mutable` and `escaping` (E0721), an unused bound name, a missing destructured argument in a named call, `requires` on a bound name (E0984), a pattern in a trait declaration.
- [ ] astdump, fmt, declsdump and query goldens show the pattern.
- [ ] SPEC.md section 2 describes the form; D103 gets a "Built" line.
- [ ] Fast check passes: `sh scripts/test.sh`
