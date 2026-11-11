---
id: c-interop
title: "C interop: extern blocks, CString and linking"
status: done
priority: should
type: feature
story: batch-8
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
added: 2026-11-11
spent: 461803
owns:
  - roadmaps/selfhost/nodes/c-interop/**
  - selfhost/src/externs.nra
  - selfhost/src/emitexterns.nra
shared:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/check.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/idcheck.nra
  - selfhost/src/trivia.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
  - selfhost/tests/spec-skip.txt
  - selfhost/src/manifest.nra
  - selfhost/src/driver.nra
  - selfhost/src/declare.nra
  - selfhost/src/resolve.nra
  - selfhost/src/calls.nra
  - selfhost/src/cnames.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emit.nra
  - selfhost/src/builtins.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/runtimetext.nra
  - std/src/**
  - testdata/manifest/**
  - selfhost/src/exprs.nra
  - selfhost/src/effects.nra
  - selfhost/src/modules.nra
  - selfhost/src/main.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/declsdump.nra
reserves: [diag E1000, diag E1001, diag E1002]
provides: []
depends_on:
  - on: unsafe-ptr
    kind: hard
    why: extern functions take and return Ptr<T> and are uses unsafe
  - on: attributes
    kind: hard
    why: @extern needs the general attribute list
---

# C interop: extern blocks, CString and linking

## Goal

D95 C interop, step 1 part 2: C functions are imported in a block that names their library, like an import: `extern { function crc32(crc: u64, data: Ptr<u8>, length: u32): u64 } from "z"`. `from "c"` needs no extra link. Every function in the block is `uses unsafe`. `@extern("deflateInit_")` gives a member its C name. `CString.from(text): Option<CString>` (None when the text holds a zero byte) and a way back to a string. `[link] libraries` in `narra.toml` is replaced by `from` (no deprecation path: remove it and update the manifest samples).

## Context

- Design: D95 "C interop"; SPEC sections 9 and 15 (`[link]`).
- Linking today: `[link] libraries` and `search` are parsed (`manifest.nra`) and carried into `Package.libraries`/`search` (`modules.nra`, which also dumps them; `main.nra` builds a package with `libraries: []`), but nothing passes them to `cc`: `cCompilerArguments` (`driver.nra`) takes `(options, cPath, output)` and always adds `-lm`, and `compileToC` returns only the C text. So `from "m"` already links. Make `compileToC` return the libraries named by extern blocks with the text, and pass them to `cCompilerArguments` as `-l<name>`. Remove `[link] libraries` (and its `Package` field and dump line); keep `[link] search` and pass it as `-L<dir>`, which is not passed today either. With tree shaking, decide whether an unreached block still links its library, and say so in SPEC.
- Rewrite the single-function `extern function ... = "name"` form in SPEC section 9 (around line 1472) and the section 12 `@target` sample (`GetTickCount64`, around line 1599) to the block form. `@target` goes before an `extern { } from` block and applies to all its members (one library per block); c-audit-target builds it.
- C names: generated names are hashed; extern names are the only unhashed ones besides step 2's exports, so check them against the generated name table (see the c-interop outline notes from hashed-cnames).
- Attributes (attributes outcome): add `extern` to `KNOWN_ATTRIBUTES` in `parser.nra`. The argument parser accepts only `name = Enum.Variant` today; `@extern("deflateInit_")` needs a string argument, so extend `AttributeArgument` for it.
- Tests must use only libc and libm (`from "c"`, `from "m"`): the Linux check image has `libc6-dev` and nothing else.

## Acceptance

- [ ] Package dump goldens updated for the removed `libraries` field; `[link] search` reaches `cc`.
- [ ] Samples: `strlen` on a `CString` from `from "c"`, `sqrt` from `from "m"`, `@extern` renaming, a call outside `unsafe` without `@trusted` (error), a name that clashes with a generated C name (error), `[link] libraries` now an unknown key, and a program linking a library only through a reached block.
- [ ] Fast check passes: `sh scripts/test.sh`
