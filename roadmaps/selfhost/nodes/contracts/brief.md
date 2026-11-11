---
id: contracts
title: "Contracts: requires, proven at compile time"
status: todo
priority: should
type: feature
story: batch-7
early: false
outline: false
milestone:
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
added: 2026-11-04
spent:
owns:
  - roadmaps/selfhost/nodes/contracts/**
shared:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/check.nra
  - selfhost/src/body.nra
  - selfhost/src/calls.nra
  - selfhost/src/funcvalues.nra
  - selfhost/src/resolve.nra
  - selfhost/src/operands.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/diagnostics.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
  - docs/guide/**
reserves: [diag E0984, diag E0985, diag E0986]
provides: []
depends_on: []
---

# Contracts: requires, proven at compile time

## Goal

D95, contracts part 1: `requires <condition>` lines between a function's signature and its body. The condition names parameters (`requires count >= 0`). Each caller must prove it from constants and narrowing, with the prover proven-operands built (`selfhost/src/operands.nra`, D92). Anything unproven is a compile error. Inside the body the condition is a known fact (a `requires divisor != 0` lets the body divide by `divisor`).

## Context

- D95 replaces SPEC.md section 13's "checked in debug builds" text; rewrite that paragraph and the example (SPEC around line 1603) to D95, and note in D95 what was built.
- The prover (`operands.nra`) keeps a `Fact { low, high, nonzero }` per integer local, stored in `checker.localNames` under `factKey(local)`; `findFactLocal` accepts only a plain integer-typed name. So in this node a condition ranges over integer parameters only, with the forms proven-operands narrows: comparisons with integer constants, `!= 0`, and `&&` of those. Any other condition (a float parameter, a field path, a call, `||`) is an error at the declaration (reserved code), not a silent pass. SPEC's example `requires x >= 0.0` changes to an integer one; floats and fields are a later extension, said in SPEC.
- Flow: at the start of the body, seed each parameter's fact from the conditions (body checking in `body.nra`). At a call (`calls.nra`), map arguments to parameters and prove each condition from the argument: an integer constant, or a local whose fact implies it (the caller's own `requires` facts count, since they were seeded the same way).
- A function with `requires` cannot be used as a function value, and `typeof` of it is E0982 (maintainer, 2026-11-11). A function type carries no condition, so a call through the value could pass an unproven argument, and the body relies on the condition with no runtime check left (a `requires d != 0` body divides by `d`; proven-operands removed the zero test from `narra_div`). The function-value case gets its own reserved code (`funcvalues.nra` is where function values are formed); `typeof` keeps E0982, which already means "no function type to write" (D96 amendments). Record both in SPEC and D95. Carrying the condition in function types is a later design. `ensures` does not have this problem: dropping a known fact about a result is safe.
- The node replaces the outline brief for contracts; ensures is contracts-ensures.

## Acceptance

- [ ] Samples: a proven call (constant, narrowed name, caller's own `requires`), an unproven call (error), an unsupported condition (float, field, `||`: errors), the body dividing by a `requires d != 0` parameter, and a function with `requires` used as a value (error).
- [ ] `fmt` and `astdump` handle the new lines; `narrac fmt` round-trips the samples.
- [ ] Fast check passes: `sh scripts/test.sh`
