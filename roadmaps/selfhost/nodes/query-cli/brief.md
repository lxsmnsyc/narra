---
id: query-cli
title: narra query
status: active
priority: should
type: feature
story:
early: false
outline: false
milestone: m9-adoption
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostQuery|TestSelfhostBootstrap'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - selfhost/src/query*.nra
  - compiler/internal/narra/selfhost_query_test.go
  - compiler/testdata/query/**
  - roadmaps/selfhost/nodes/query-cli/**
shared:
  - selfhost/src/main.nra
  - selfhost/src/check.nra
  - selfhost/src/effects.nra
  - compiler/README.md
  - SPEC.md
reserves: []
provides: [query-output]
depends_on:
  - on: modules
    kind: behavior
    via: module-graph
  - on: selfhost-bootstrap
    kind: hard
    why: written in Narra on the self-hosted checker; the language service, lint and skills build on it
---

# narra query

## Goal

`narrac query` answers the six questions of `SPEC.md` section 14 about a package, from the self-hosted loader and checker: `type-of FILE:LINE:COL`, `def NAME`, `refs NAME`, `callers FUNCTION`, `interface MODULE` and `recursion MODULE`, each with a text form and a `--json` form. It freezes the `query-output` contract, which the language service, lint and the agent skills read.

## Context

- `Modules.load(path)` (`selfhost/src/modules.nra`) loads a package or a single file; `Body.checkPackage` then `Effects.checkEffects` run every check pass (see `typesDump` in `selfhost/src/main.nra` for the sequence and error handling).
- Checker results are side tables by node id, kept per body: `Checker.bodies[i].types` and `.refs` (`Ref` in `selfhost/src/check.nra`), with `findBodySlot`. A generic function can have several bodies (one per instance, `BodyInfo.typeArguments`).
- `Ref` tells what an identifier names: `Local`, `Const`, `Function`, `Builtin`, `Constructor`, `Implementation`, `TraitFunction`, `Distinct`, `Value` (a call through a function value).
- The effect pass builds a call graph (`selfhost/src/effects.nra`, top comment and `funcEffects`); `callers` and `recursion` should read it rather than build a second one. Export what is needed from `effects.nra`.
- Expressions carry only a start `Pos` (`StringLit` also has `end`). Identifier ranges end at start plus the name's length; columns count characters, as in the lexer.
- Qualified names follow `contracts/module-graph.md`: `<module>.<name>`, such as `app/text.helper`.
- Tests: `compiler/internal/narra/selfhost_test.go` (`selfhostBuild`, `runNarrac`).
- Source: `SPEC.md` section 14 "Queries", D11 in `DECISIONS.md`, `contracts/query-output.md` (draft, frozen by this node).

## Steps

- [ ] `selfhost/src/query.nra`: `runQuery(arguments)` loads and checks the package (path from `--package DIR`, default `.`), then dispatches. Load or check errors print as `narrac types` prints them and exit 1.
- [ ] A position index: from a file and line:column, find the innermost expression, pattern or declaration name that starts at or covers that token. One loop per body or module, recursing only on nesting.
- [ ] `type-of`: the type text (`Check.typeString`) and, for a function or call, its effect set.
- [ ] `def`: the declaration position of a qualified or bare name; a bare name lists every match.
- [ ] `refs`: every resolved use of an exported declaration, across modules, in file and position order.
- [ ] `callers`: every function whose body calls the function directly (from the call graph); calls through function values are not counted and the text says so once.
- [ ] `interface`: the exported signatures of a module in source order (functions with effects, structs, enums, traits, constants, type of each).
- [ ] `recursion`: the strongly connected components of the module's call graph with a cycle, plus self calls, iteratively (no recursion per function).
- [ ] `--json` for each command as `contracts/query-output.md` says, with a small JSON string escaper in `query.nra` (std/json does not exist yet).
- [ ] Freeze the contract: write the final shapes into `contracts/query-output.md` through the orchestrator (report them in `outcome.md`), and add `narrac query` to `compiler/README.md`.

## Acceptance

- [ ] `compiler/testdata/query/<case>/` holds small packages (`narra.toml`, `src/`) and a `queries.txt`: lines `> <command and arguments>` each followed by the expected stdout up to the next `>` line. `TestSelfhostQuery` runs each command against its package and compares byte for byte, text and `--json` forms both.
- [ ] Cases cover: type-of on a local, a call with effects, a field read, a generic function's parameter, a position between tokens (error); def of a function, struct, enum variant and constant, qualified and bare, an unknown name (exit 1); refs across two modules; callers including a call inside an arrow function; interface of a module with every declaration kind; recursion with a self call, a two-function cycle and a non-recursive module.
- [ ] A query on a package with check errors prints the errors and exits 1; an unknown subcommand prints usage and exits 2.
- [ ] TestSelfhostBootstrap passes with the new modules in the package.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L. Uncertainty medium (index from positions to ids without end positions, generic bodies), coupling medium (two dependencies, three consumers through the contract, shares main.nra with fmt), blast radius medium (exports from check.nra and effects.nra, append-only), verifiability low (golden tests), size medium (about 5 files, 800 to 1000 lines), context high (checker side tables and the effect pass, over 20k tokens). Split candidate if it grows: `callers` and `recursion` as a second node on the frozen contract.

## Notes for the worker

- Do not change what the checker or effect pass computes; only export readers. TestSelfhostEmit and TestSelfhostBootstrap compare the emitted C with Go, so any change in check order shows up there.
- For a generic function, report the declared types (with type parameter names), not each instance; pick the body whose `typeArguments` is empty, or the declaration's signature when there is none.
- Names: `findNodeAt`, `collectCallers`, `findCycles`, not `nodeAt` or `cyclesOf` (roadmap CLAUDE.md).
- fmt runs in the same wave and also adds a line to `run` in `main.nra`. Rebase before review.
- No new diagnostic codes; query errors are plain `error ...` lines.

## Open questions

1. Command name `narrac query` until the self-hosted compiler takes the name `narra`. Confirm.
2. `def` and `refs` with a bare name that several modules declare: list all (proposed), or require a qualified name?
3. `type-of` on a token that names no expression (a keyword, a type annotation): error (proposed), or the type the annotation denotes?
