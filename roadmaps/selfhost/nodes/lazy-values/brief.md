---
id: lazy-values
title: "lazy values: a top-level constant built on first force"
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
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/lazy-values/**
  - selfhost/src/stdlazy.nra
  - selfhost/src/lazyvalues.nra
  - testdata/lazy_values*
  - testdata/errors/lazy_*
shared:
  - selfhost/tests/src/emitchecks.nra
  - selfhost/src/emit.nra
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/trivia.nra
  - selfhost/src/exprs.nra
  - selfhost/src/body.nra
  - selfhost/src/effects.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/stdcalls.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/emitconsts.nra
  - selfhost/src/emit.nra
  - selfhost/src/diagnostics.nra
  - testdata/astdump/**
  - testdata/fmt/**
  - testdata/declsdump/**
  - SPEC.md
  - compiler/README.md
reserves: [diag E1032, diag E1033, diag E1034, diag E1035]
provides: []
depends_on: []
---

# lazy values: a top-level constant built on first force

## Goal

`const table = lazy buildTable()` declares a top-level `Lazy<T>` whose value is computed by the first `Lazies.force(table)` and returned by every later one. Unlike a plain top-level constant, the expression may call functions. Today it fails to parse:

```
main.nra:5:20: expected the end of the statement, found "buildTable"
```

(E0013, `narrac build` of the program below.) The plan: a `lazy` flag on `ConstDecl`, a checker that skips the constant-expression rule for it, a `std/lazy` module with `force`, and the getter that top-level constants already have (D61) as the once guard.

```ts
import * as Lazies from "std/lazy"

function buildTable(): Array<int> {
  return [1, 2, 3]
}

const table = lazy buildTable()

export function main(): int {
  return Lazies.force(table)[0] - 1
}
```

## Context

- Source: SPEC section 10 (`lazy` bullet, around line 1597), SPEC section 3 (no import side effects, around line 349), README "Not supported yet" (`@const` functions and `lazy` values). Split from const-lazy on 2026-11-12; `@const` functions go to const-eval-design.
- Most of the machinery exists. D61 and the README's "Differences from the spec" (top-level constants that are not scalars) build a constant on first use behind a getter: `emitGetter` in `selfhost/src/emitconsts.nra` writes `static T G_name; static int G_name_ready;` and `g_name()`, and `emitConstRelease` frees it in `narra_consts_release` before the leak report. A lazy value is that getter with the constant-expression check skipped. `lazyConstants` there already collects the non-static constants; a lazy value is never static (`isStaticConstant` in `exprs.nra`).
- The rule to skip is `checkConstValue` in `selfhost/src/exprs.nra` (E0192 "constant cannot call function", around line 2914, and E0193). For a lazy value the expression is checked as an ordinary expression in the constant's body (`const|N`, see `findBody` in `emitGetter`).
- AST: a field on `ConstDecl` in `selfhost/src/ast.nra`, `lazyPos: Option<Pos>`, set when `lazy` is written. No new `Expr` kind, because `lazy` is valid only on a top-level constant. The field still needs a case in `astdump` (a `lazy` flag), `declsdump`, `fmt` (`printConst` prints `lazy `) and `trivia` (the keyword can carry a comment on either side, so it needs a trivia node: selfhost/CLAUDE.md). `buildconditions.nra` rebuilds a `ConstDecl` with a spread and keeps the field.
- `lazy` stays a contextual word, read in `parseConstDecl` (`parser.nra`, around line 1280) after `=`. Making it a keyword would change the `lex` goldens and forbid `lazy` as a name.
- `std/lazy` is a compiler module, not Narra source, because `force` reads the getter: `selfhost/src/stdlazy.nra` with its `LAZY_FUNCS` rows and checks, one `registerAll` line in `stdlib.nra`, one `if` in `checkStdCall` (`stdcalls.nra`) and one in `emitNamedCall` (`emitstd.nra`). `Lazies.force(table)` emits `g_table()`, which returns the value borrowed.
- Rules (`selfhost/src/lazyvalues.nra`, new):
  - E1032: `lazy` anywhere but the value of a top-level `const`. A local `const x = lazy build()` gets this error in place of E0013.
  - E1033: a lazy name used other than as the argument of `Lazies.force`: passed to a function, stored, compared, used in another constant's value or in a template. So `Lazy<T>` needs no runtime representation in this node.
  - E1034: the lazy expression has an effect other than `alloc`. A first force can happen anywhere, so `io` or `unsafe` there would hide the effect in `Lazies.force`. `Lazies.force(table)` has the effects of the expression (`alloc` or none), added in `effects.nra`.
  - E1035: a lazy value whose expression forces itself, directly or through functions it calls or other lazy values it forces. Record per body the lazy values it forces (beside the effects fixed point in `effects.nra`) and report a cycle like E0191 does for constants. A named function used as a value counts as called.
- Reachability: `emitReachableBodies` in `emit.nra` reaches a lazy value's getter only through `Lazies.force`, so an unforced lazy value emits no C.
- No threads exist yet (node threads is todo). The `_ready` flag is the once guard for one thread, as for D61 constants. The README line that says the threads node must add a once guard covers lazy values too.

## Acceptance

- [ ] Samples (`testdata/lazy_values*.nra` with `.out`): the program above; a lazy string built with a template; a lazy struct and a lazy `Map`; a lazy value forced inside a loop that prints once from its builder (a counter shows the build ran once); a lazy value that forces another; an exported lazy value forced from another module; a lazy value never forced (no getter in the C); `Lazies.force` inside a closure.
- [ ] Every sample is clean under the leak check and sanitizers: `narra_consts_release` frees each forced value and skips the unforced ones.
- [ ] Error tests: a local `lazy` (E1032), a lazy name passed to a function, compared and used in a plain constant (E1033), a lazy expression calling a function that prints (E1034), a lazy value forcing itself and a cycle of two (E1035), `const x = buildTable()` without `lazy` (still E0192).
- [ ] Emit check in `selfhost/tests/src/emitchecks.nra`, or a line in the existing `tree-shake/unreached` case: an unforced lazy value has no getter in the C.
- [ ] astdump, declsdump and fmt goldens with `lazy`, one with comments around the keyword. `narrac fmt` is stable on them.
- [ ] SPEC.md section 10 states the rules, with `Lazies.force` and `import * as Lazies from "std/lazy"`; section 3's list of standard modules (around line 312) gets `std/lazy`. `compiler/README.md` drops `lazy` values from "Not supported yet" and gets a line for `std/lazy`.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M. The getter, the release and the reachability exist; the work is the flag through parser and dumps, one std module and four checks.

## Notes for the worker

- Do not change how plain top-level constants are emitted. A lazy value only reuses their getter.
- Seed rule: do not use `lazy` in `selfhost/src` in this node.
- Open question: `Lazy.force(table)` as SPEC writes it, or `Lazies.force(table)`? Recommended: `Lazies.force` from `std/lazy`, with `import { Lazy } from "std/lazy"` to write the type, as `std/cell` does (`Cell`, `Cells`). A prelude type does not bring its functions (D67), and `Lazy` is not in the prelude. Update the SPEC text.
- Open question: may a `Lazy<T>` be passed as a value? Recommended: not in this node (E1033). A value needs a runtime handle (a pointer to the getter) and a type rule for generic code. Nothing asks for it yet.
- Open question: does `Lazies.force` hold the build's `alloc`? Recommended: yes, the force call carries the expression's effects. Saying "pure" while the first call allocates would make D58's effect list wrong.
- Open question: a panic in the expression (`assert`, an index out of range)? Recommended: it panics at the first force, like a D61 constant (D61's cost line). D92 allows these panics.

## Plan review (2026-11-12)

- Type of a lazy name: decided by the orchestrator, the name has type `T` with a lazy flag on the declaration, not a new `Type` variant. Drop the `Lazy` import and `Lazy<T>` type from the plan and samples; `Lazies.force(table)` takes the lazy name and returns `T`, and any other use of the name is E1033.
- Reachability: inline also changes `emitReachableBodies` (emit.nra:177). Whichever lands second keeps both rules and reruns the `tree-shake` emit checks.
