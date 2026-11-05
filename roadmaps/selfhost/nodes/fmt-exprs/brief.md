---
id: fmt-exprs
title: "fmt: expressions, types and patterns"
status: done
priority:
type: feature
story: fmt
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostFmt|TestSelfhostBootstrap'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - selfhost/src/fmtexprs.nra
  - compiler/testdata/fmt/expr_*
  - compiler/testdata/fmt/type_*
  - compiler/testdata/fmt/pattern_*
  - roadmaps/selfhost/nodes/fmt-exprs/**
shared:
  - selfhost/src/fmtprinter.nra
  - selfhost/src/fmt.nra
reserves: []
provides: []
depends_on:
  - on: fmt-doc
    kind: contract
    via: fmt-doc
  - on: fmt-trivia
    kind: contract
    via: fmt-trivia
  - on: fmt-cli
    kind: behavior
    via: fmt-printer
---

# fmt: expressions, types and patterns

## Goal

`selfhost/src/fmtexprs.nra`: `printExpression`, `printPattern` and `printType` of `contracts/fmt-printer.md`, for every `Expr`, `Pattern` and `TypeExpr` variant in `selfhost/src/ast.nra`, reflowing with groups as Prettier does for TypeScript. Then fmt-cli's driver uses them in place of its stub. The rules are in the story brief `nodes/fmt/brief.md`.

## Context

- Variants to cover (ast.nra): literals, templates, identifiers, field and index reads, `is`, map, struct, array and tuple literals with spread, calls, unary and binary operators, `match`, block expressions (and the ternary, which parses to a `BlockExpr` the parser marks in `ternaryIds`; print it back as `c ? a : b`), arrow functions; every pattern form; every type form.
- Precedence: the parser's precedence table in `selfhost/src/parser.nra`. Print parentheses where precedence needs them, and keep the ones `Trivia.parenthesized` records.
- Prettier rules to follow: a broken binary chain breaks after each operator and indents the rest; call arguments hug a last arrow function or object literal; member chains are not a Narra construct (no method calls); `match` arms each on their own line once the match breaks; template literals print as written.
- Copy literal text (numbers, strings, characters, templates) from the source by position, so escapes and number spellings stay as written.

## Acceptance

- [ ] Golden cases `compiler/testdata/fmt/{expr,type,pattern}_<case>.nra` and `.fmt` cover every variant above, each in a fitting and a broken form, with `trailingComma` all and none, comments inside expressions, and written and needed parentheses. `TestSelfhostFmt` runs them.
- [ ] fmt-cli's stability checks (idempotence, tokens, AST) pass on every `.nra` file under `compiler/testdata/` and `selfhost/src/` with the real expression printer.
- [ ] TestSelfhostBootstrap passes. Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M. Uncertainty medium (Prettier's expression rules, adapted), coupling medium (two contract dependencies and the frozen fmt-printer), blast radius low (one new module; the stub wiring line in fmt.nra), verifiability low (golden and stability tests), size medium (about 4 files, 700 to 800 lines), context medium (ast.nra, the parser's precedence table, the contracts).

## Notes for the worker

- `generators` may land on main while this story is open. When the story branch takes in main, print `function*`, `yield` and `yield*` too, and add them to the golden tests.

- Rebase on the story branch after fmt-cli lands and replace its stub fields in `fmt.nra` with the real functions; that is the only edit outside your own files.
- Recurse only on nesting (selfhost/CLAUDE.md); loop over argument, element and arm lists.
- An arrow body that is a block calls `printer.printBlock` from fmt-cli.
