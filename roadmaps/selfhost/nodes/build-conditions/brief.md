---
id: build-conditions
title: "Build conditions: std/build constants and one emitted branch"
status: active
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
added:
spent:
owns:
  - roadmaps/selfhost/nodes/build-conditions/**
shared:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/check.nra
  - selfhost/src/stmts.nra
  - selfhost/src/exprs.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitter.nra
  - selfhost/src/builtins.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/runtimetext.nra
  - selfhost/src/stdbuild.nra
  - selfhost/src/fixes.nra
  - selfhost/src/driver.nra
  - selfhost/src/lspcache.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/diagnostics.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
  - docs/guide/**
reserves: [diag E0989]
provides: []
depends_on: []
---

# Build conditions: std/build constants and one emitted branch

## Goal

D95 and SPEC section 12 "Build conditions": `import * as Build from "std/build"` gives `Build.os`, `Build.arch` and `Build.debug` as compile-time constants (`Os` and `Arch` enums, `Build.debug` false under `-release`). An `if` whose condition tests only these constants keeps one branch in the output, with no runtime check; every branch is still type-checked. The host decides the values (macOS and Linux, arm64 and x86_64) through one target table that cross-targets will extend.

## Context

- Effects of every branch count toward the function's effects, also the branch not emitted (orchestrator, 2026-11-11). A function then has the same signature on every target, and SPEC already checks every branch on every build. Record it in SPEC section 12.
- `else if` chains and `&&`/`||` of build constants fold too. A condition mixing build constants with runtime values is an ordinary `if`.
- `std/build` is a new std module: register it like `std/time` (`stdtext.nra` path switch, a registry file such as `stdbuild.nra` modelled on `stdtime.nra`, `runtimetext.nra` if it needs runtime text, and the module-name map in `fixes.nra`).
- Plumbing: today `options.release` in `driver.nra` only adds `-DNARRA_RELEASE` to the `cc` arguments (`cCompilerArguments`). The checker must know the target to fold and to type `Build.debug`: store a target (os, arch, debug) on the Checker when it is created, set from the driver options for `build`, `run`, `test`, `emit` and `check`, and from the host in debug mode for the language server (`lspcache.nra`). The emitter reads it from the Checker.
- The node replaces the outline brief for build-conditions; `@target` and `--all-targets` are build-targets.

## Acceptance

- [ ] Samples: a `Build.os` branch that prints per OS (golden output for the host), a branch with a type error in the non-host branch (error), and `Build.debug` in debug and release (two runs).
- [ ] The emitted C for the sample contains only the host branch.
- [ ] Fast check passes: `sh scripts/test.sh`
