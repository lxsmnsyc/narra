---
id: parameter-defaults
title: "Parameter defaults"
status: todo
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
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/parameter-defaults/**
  - selfhost/src/paramdefaults.nra
  - testdata/param_defaults*.nra
  - testdata/param_defaults*.out
  - testdata/errors/param_defaults*.nra
  - testdata/errors/param_defaults*.err
  - testdata/astdump/param_defaults*
  - testdata/fmt/param_defaults*
  - testdata/declsdump/param_defaults*
shared:
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/body.nra
  - selfhost/src/calls.nra
  - selfhost/src/exprs.nra
  - selfhost/src/closures.nra
  - selfhost/src/effects.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/query.nra
  - selfhost/src/lsp*.nra
  - SPEC.md
  - compiler/README.md
  - docs/guide/03-functions-and-closures.md
  - docs/guide/11-narra-for-typescript-programmers.md
reserves: [diag E1210-E1219]
provides: []
depends_on:
  - on: constructor-field-defaults
    kind: hard
    why: parameter defaults use isConstantValue, which this fix changes
  - on: parameter-defaults-design
    kind: hard
    why: the design gate fixes the syntax, codes and SPEC text this node builds (D115)
---
# Parameter defaults

## Goal

Build D115: constant defaults on the trailing parameters of declared functions. A call may leave out a parameter with a default, by passing fewer positional arguments or by naming later ones; the checker fills the default into the call, so each function stays one C function with a fixed parameter list.

## Context

- Source: D115 in `DECISIONS.md`; SPEC section 2, Functions, the bullet "Parameter defaults (D115, not built yet)". They override this brief where they differ.
- Proof first: `testdata/param_defaults.nra` fails to compile on the batch base (the parser rejects `=` after a parameter type).
- Model to copy: struct field defaults. `checkFieldDefaults` in `selfhost/src/body.nra` checks each default as a body of its own (`startBody`, `keyBody` with a `default|...` key, `isConstantValue`, error E0106), `checkGenericDefaults` handles defaults whose type uses a type parameter, and the struct literal emitter in `selfhost/src/emitexprs.nra` (`defaultBodyKey`, `switchBody`) emits an omitted field from that body. Parameter defaults use the same steps with a key such as `paramdefault|<function>|<parameter>`.
- Fill step: `matchNamedArguments` in `selfhost/src/calls.nra` reports E0331 for a missing parameter, and `checkArguments` in `selfhost/src/exprs.nra` reports E0316 for a wrong positional count. Both accept a missing trailing parameter that has a default. Record the filled parameters in a side table by call id (as `argumentOrders` does) so the effect pass and the emitter see them.
- Parsing: `parseFuncHeader` in `selfhost/src/parser.nra` serves functions, trait functions and extern functions; `parseArrow` and `parseFuncType` parse the other parameter lists. Add `default: Option<Expr> = None` to `Param` in `ast.nra` (a field default, so the other `Param` literals need no change).

Codes (range E1210-E1219):

- E1210: `a parameter without a default cannot follow one with a default; give "<name>" a default or move it before "<first>"`.
- E1211: `a mutable parameter cannot have a default`.
- E1212: `a parameter default must be a constant` (the rule of `isConstantValue`).
- E1213: `an arrow function parameter cannot have a default` and `a function type cannot have parameter defaults` (two rows, one code).
- E1214: `a trait function cannot have parameter defaults`, in the trait and in an `implement` block.
- E1215: `a function of an extern block cannot have parameter defaults`.
- E1216: `optional parameters are written with a default; write "<name>: Option<T> = None"` for `name?: T`, with the fix in the JSON output.
- E1217-E1219: free.
- E0316 gets a second format row for a function with defaults: `this call passes %d arguments, the function takes %d to %d`.

## Steps

- [ ] Parser, AST, astdump (a `.default` slot line), fmt (`name: T = value`), declsdump.
- [ ] Checker: E1210, E1211, E1213 to E1216 at the declaration; each default checked as a body against the parameter type, E1212 for a non-constant; generic defaults checked with the type parameters open.
- [ ] Calls: fill left-out defaults in `matchNamedArguments` and `checkArguments`; inference ignores them; `requires` contracts see the filled value (`recordContractCall`).
- [ ] Function values: the type of a function named as a value lists every parameter and no defaults; a call through a value passes every argument (no change expected, add a test).
- [ ] Effect pass: a filled array default charges `alloc` at the call.
- [ ] Emitter: emit each filled default from its body at the call, in parameter order, after reordered named arguments; `@inline` and `@const` calls see the filled argument.
- [ ] Language service and `narrac query`: hover and signature text show `= value`.
- [ ] SPEC: drop "not built yet" from the bullet. Guide: a short section in chapter 3 and the sentence in chapter 11 ("There are no optional parameters, default parameters, ...") changed to say defaults exist and are constants. `compiler/README.md`: the feature list.

## Acceptance

- [ ] `testdata/param_defaults.nra` with checked output: positional calls that leave out one and two defaults, a named call that skips a middle default, a reordered named call, string, bool, `None`, `Some(1)`, enum constructor and `[]` defaults, a generic function with `fallback: Option<T> = None` called with and without it, an exported function with defaults called from another module, an `@inline` function and a generator with defaults, and a function with defaults passed as a value and called with every argument. Clean under AddressSanitizer, the leak check and the copy check.
- [ ] `testdata/param_defaults_effects.nra`: an exported function that calls one with an array default lists `uses alloc`; leaving it out compiles only with `alloc` declared (an error test for the missing effect).
- [ ] Error tests, one file each: `param_defaults_order` (E1210), `param_defaults_mutable` (E1211), `param_defaults_not_constant` (E1212, a call and a parameter read), `param_defaults_arrow` and `param_defaults_func_type` (E1213), `param_defaults_trait` (E1214), `param_defaults_extern` (E1215), `param_defaults_optional` (E1216 with its fix), `param_defaults_generic` (`count: T = 0`, a type error), `param_defaults_missing` (a required parameter left out, E0331 and E0316 with the new format), `param_defaults_value_call` (a call through a function value that leaves out an argument, E0316).
- [ ] astdump and fmt goldens for the new slot; the `sources`, `codes` and `spec` suites pass.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: one syntax slot and one fill step that reuse the field default machinery; about 500 lines with tests.

## Notes for the worker

- Do not use defaults in `selfhost/src` or seed-compiled std: the seed cannot compile them until the batch's seed update. An adoption sweep (for example `checkExpression(..., None)`, `argumentText(..., 0)`) is a later node.
- `parseFuncHeader` is shared by trait and extern functions: parse the default there and report E1214 and E1215 in the checker or at the callers, so the message names the place.
- An omitted struct field default with a non-empty array literal is not charged `alloc` by the effect pass today (`Expr.StructLit` in `effects.nra` walks only the written fields). Do not copy that gap; if it is a real bug, report it in your outcome as a separate fix.
- Shared files carry only dispatch rows; put the checks in `selfhost/src/paramdefaults.nra`.
