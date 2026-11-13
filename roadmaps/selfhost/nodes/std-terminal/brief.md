---
id: std-terminal
title: "std/terminal"
status: todo
priority: should
type: feature
story: batch-14
early: false
outline: false
milestone: p2-std
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
  - std/src/terminal.nra
  - selfhost/src/stdterminal.nra
  - selfhost/runtime/std_terminal.c
  - selfhost/src/runtimestdterminal.nra
  - testdata/terminal*.nra
  - testdata/terminal*.out
  - testdata/errors/terminal_*
  - roadmaps/selfhost/nodes/std-terminal/**
shared:
  - selfhost/src/stdinternal.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/emitstd.nra
  - testdata/errors/unknown_module.err
  - testdata/declsdump/imports/err_unknown_module.dump
  - SPEC.md
  - compiler/README.md
  - docs/guide/
reserves: [diag E1370-E1379]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 7 fixes detection, NO_COLOR handling and the style functions

---

# std/terminal

## Goal

Build `std/terminal` (alias `Terminal`) as D118 item 7 says: `isTerminal`, `columns` and `useColor` over two new C primitives in `std/internal/terminal` (`isatty`, `ioctl` with `TIOCGWINSZ`), and pure style functions (`bold`, `red`, `bgRed` and the rest, `style(text, Style)`, `stripStyles`) that nest.

## Context

- Source: D118 items 7 and 8; picocolors for names and nesting (reset codes 22, 23, 24, 27, 29, 39, 49 and reopening the outer style).
- Model for a C primitive: `std/internal/clock` (`selfhost/runtime/std_clock.c`, `selfhost/src/runtimestdclock.nra`, its rows in `stdinternal.nra`, its emitter in `emitstd.nra`).
- `NO_COLOR`, `FORCE_COLOR`, `TERM` come from `Env.get`.

## Steps

- [ ] `std/internal/terminal`: `isTerminal(fd: int): bool uses io` and `columns(fd: int): int uses io` (0 when unknown) in `std_terminal.c`, linked only when used.
- [ ] `std/src/terminal.nra`: `enum Stream { Stdin, Stdout, Stderr }`, `isTerminal`, `columns` (`Option<int>`), `useColor` with the rule order of D118, the style functions, `Style`, `Color` (16 named, `Ansi256`, `Rgb`), `style`, `stripStyles`.
- [ ] Registry rows, `stdlib.nra`, regenerated `stdtext.nra` and runtime text, module-list goldens; SPEC, README and guide text.

## Acceptance

- [ ] `testdata/terminal_styles.nra` prints the escaped form of each style, of nested styles (`bold` around text with a `red` part: the bold is reopened after the red ends), of `style` with `Rgb` and `Ansi256`, and `stripStyles` of each.
- [ ] `testdata/terminal_detect.nra` run with standard output to a file: `isTerminal(Stream.Stdout)` is false, `columns` is `None`, `useColor` is false; with `FORCE_COLOR=1` true; with `NO_COLOR=1` and `FORCE_COLOR=1` false. (Use the sample runner's environment support, or `Process.run` of the sample itself with set variables.)
- [ ] Error test: `testdata/errors/terminal_internal_import.nra` (a program imports `std/internal/terminal`: E0426).
- [ ] Linux: `sh scripts/linux-check.sh` passes (the `ioctl` header differs).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: about 300 lines of Narra and 60 of C, one primitive module.

## Notes for the worker

- Runtime text changes need two seed updates (GOAL constraints); the orchestrator does them at the end of the batch. narrac does not use this module in this batch.
- No tagged-template tag (D114 is open); the style functions are what a later tag would call.
- Codes E1370-E1379 are reserved; none is expected.
