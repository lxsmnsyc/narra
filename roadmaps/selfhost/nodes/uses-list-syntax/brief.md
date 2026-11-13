---
id: uses-list-syntax
title: "Effect lists inside type lists"
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
  - roadmaps/selfhost/nodes/uses-list-syntax/**
  - testdata/uses_list_*.nra
  - testdata/uses_list_*.out
  - testdata/errors/uses_list_*
  - testdata/fmt/type_uses_*
  - testdata/parse/uses_list_*
shared:
  - selfhost/src/parser.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/fmt.nra
  - selfhost/src/check.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/query.nra
  - selfhost/src/querysignature.nra
  - selfhost/src/resolve.nra
  - selfhost/src/traits.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/registry.nra
  - testdata/**/*.nra
  - testdata/**/*.err
  - testdata/**/*.fmt
  - testdata/**/*.dump
  - testdata/query/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
  - docs/guide/10-effects-and-testing.md
reserves: [diag E1240-E1249]
provides: []
depends_on:
  - on: uses-list-design
    kind: hard
    why: the design gate fixes the syntax, codes and SPEC text this node builds (D117)
---

# Effect lists inside type lists

## Goal

Build D117. In a type, `uses` takes one effect name or a comma list in parentheses (`() => int uses (io, alloc)`); a function declaration's own list stays a bare comma list. This fixes `Map<Generator<int> uses io, int>` (parsed today as one argument with effects `io, int`), `#[Generator<int> uses io, Array<int>]` (a parse error today) and the trailing comma after `g: Generator<int> uses io,` in a parameter list (`expected a name, found ")"` today).

## Context

- Source: D117 in `DECISIONS.md` and the outcome of `uses-list-design`. They override this brief where they differ.
- Evidence (at `feat/batch-13`, `narrac parse`): `function f(m: Map<Generator<int> uses io, int>): int` dumps `uses=io,int`; `function f(t: #[Generator<int> uses io, Array<int>]): int` fails with `expected "]", found "<"`; a parameter list ending `g: Generator<int> uses io,\n)` fails with `expected a name, found ")"`; `function f(): () => int uses io, alloc` gives both effects to the result type.
- Parser: `parseUsesList` in `selfhost/src/parser.nra` has three callers, the generator branch of `parseType`, `parseFuncType` and `parseFuncDecl`. Give the two type callers a type form (one name, or `(` names `)` with an optional trailing comma inside) and keep the bare loop for `parseFuncDecl`. Delete `isParamAfterComma` once nothing calls it.
- The AST does not change: `TypeExpr.effects` holds the names either way, so `astdump` goldens keep their text (positions may move only where sources change).
- Type display: `usesSuffix` in `check.nra` is used for types (generator names, function types in `check.nra`, `declsdump.nra`, `query.nra` type-of) and must print `uses (a, b)` for two or more effects there. Declaration lists (`querysignature.nra` line for a `FuncDecl`, the hover `uses a, b` line in `lspnavigation.nra`, the `effects:` line of `query`) stay bare. Add a second helper for the type form rather than a flag.
- fmt: `printType` and `printFunctionType` in `fmtexprs.nra` print the type form; `uses (io)` prints as `uses io`. Remove the `endsWithUses` trailing-comma exception (`fmt.nra` and `fmtexprs.nra`, five call sites), so a broken parameter list always ends with a comma, as oxfmt does.
- Measured by the spike: 27 type lists with more than one effect, none in `selfhost/src` or `std/src`, so no seed step is needed. The 2 in `selfhost/tests/src/registry.nra` change in the same commit; the test runner is built by the compiler under test.

## Steps

- [ ] Proof first: add `testdata/uses_list_types.nra` with `Map<Generator<int> uses (io, alloc), int>`, a tuple `#[Generator<int> uses io, Array<int>]`, an unnamed function-type parameter list `(() => int uses io, int) => int`, a trailing comma after a `uses` parameter, and `Generator<string> uses (io, alloc) uses alloc` on a non-`function*` function. Show it fails on the batch base.
- [ ] Parser: type form of the list, declaration form unchanged. E1240 where a `,` follows a type's one effect and the list cannot go on: after a declaration's result type, and in a declaration's parameter list before a name with no `:`. No automatic fix (the rule against fixes for old forms).
- [ ] Checker: E1241 in the unknown-type path (`resolve.nra` near the E0118 report, and `traits.nra` if it has its own): a type name that is `io`, `alloc`, `unsafe` or an effect variable in scope. When the type before it in the same list ends with a `uses` list, the message adds ``write `uses (io, alloc)` ``.
- [ ] Display and fmt changes above.
- [ ] Rewrite the 27 sources (`count` with `narrac parse ... | grep uses=` per D117) and the goldens (`sh scripts/test.sh --update`), then read the golden diff: only `uses a, b` to `uses (a, b)` in types and the fmt trailing comma may change.
- [ ] SPEC: drop the "not built yet" notes D117 added in section 7 and the generator section. `compiler/README.md`: the Effects and Generators lines name the form. `docs/guide/10-effects-and-testing.md`: one paragraph and one example of `uses (io, alloc)` in a function type, without SPEC or D-number citations.
- [ ] Add the two code rows and their `narrac explain` texts to `selfhost/src/diagnostics.nra`.

## Acceptance

- [ ] `testdata/uses_list_types.nra` runs with checked output, clean under AddressSanitizer, the leak check and the copy check; it fails to compile on the batch base.
- [ ] `testdata/parse/uses_list_*.nra` dumps show `uses=io,alloc` on the type and the declaration list each where written, for `function f(): () => int uses (io, alloc) uses io`.
- [ ] Error tests: `testdata/errors/uses_list_result_comma.nra` (E1240 after `Generator<string> uses io, alloc {`), `uses_list_param_comma.nra` (E1240 for `f: () => int uses io, alloc,` in a declaration parameter list), `uses_list_effect_type.nra` (E1241 for `Map<Generator<int> uses io, alloc>`, message with the parentheses hint), `uses_list_effect_variable.nra` (E1241 for an effect variable `E` used as a type), `uses_list_decl_parens.nra` (`function f(): int uses (io) {` is a parse error, plain `expected a name`).
- [ ] fmt: `testdata/fmt/type_uses_*` cover `uses (io)` printing as `uses io`, a two-effect type, and a broken parameter list whose last parameter ends in `uses io` keeping its trailing comma.
- [ ] A message for a two-effect type prints `uses (E, alloc)` (the `generator_bound_*.err` goldens).
- [ ] `narrac query` type-of and the hover show the type form; `query signature` of a declaration keeps the bare list.
- [ ] `sources` and `codes` suites pass.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: one parser rule and two codes, but the display change touches about 19 goldens and 27 test sources.

## Notes for the worker

- No migration path: do not add an error that names the old form. E1240 and E1241 are grammar and type errors a new user writes too; keep their text about the rule, not about a change.
- Do not use parentheses in `selfhost/src` or `std/src`: the seed does not accept them until the end-of-batch seed update. Nothing there needs them.
- Existing fmt bug, not part of this node: a comment after `uses` moves (`Generator<int> uses /* c */ io` formats as `Generator<int /* c */> uses io`). A comment inside the new parentheses may move the same way; test one, and if it moves, add a line to the batch inbox instead of fixing the trivia here.
- Other batch-13 builds edit `parser.nra`, `fmt.nra` and `diagnostics.nra`: keep the change in the three `parseUsesList` call sites and the display helpers; the second to land rebases.
