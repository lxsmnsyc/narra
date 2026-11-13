---
id: std-args
title: "std/args"
status: active
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
  - std/src/args.nra
  - selfhost/src/stdargs.nra
  - testdata/args*.nra
  - testdata/args*.out
  - testdata/errors/args_*
  - roadmaps/selfhost/nodes/std-args/**
shared:
  - selfhost/src/stdlib.nra
  - selfhost/src/stdtext.nra
  - testdata/errors/unknown_module.err
  - testdata/declsdump/imports/err_unknown_module.dump
  - SPEC.md
  - compiler/README.md
  - docs/guide/
reserves: [diag E1330-E1339]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 4 fixes the spec types, the syntax, the errors and the help text

---

# std/args

## Goal

Build `std/args` (alias `Args`) as D118 item 4 says: a `Command` value describes flags, positionals and subcommands, `Args.parse(command, arguments)` returns `Result<Parsed, ArgsError>`, and `Args.help` writes the help. Pure Narra source that the seed compiles, since narrac will use it.

## Context

- Source: D118 item 4; Node's `util.parseArgs` for names, clap for `Help` as an error.
- What narrac does today: `selfhost/src/driver.nra` `parseOptions` (Go `flag` syntax, one or two dashes), the hand loops in `fmt.nra` `runFmt` and `query.nra` `runQuery`, and the usage line in `main.nra`. Moving narrac is not this node (see Notes).

## Steps

- [ ] Types: `Command`, `Flag` (defaults as in D118), `enum Rest { None, Positionals, Raw }`, `Parsed`, `enum ArgsError`.
- [ ] `parse`: `--name`, `--name=value`, `--name value`, `-x`, `-x value`, `-xvalue`, bundled `-abc`, `--`, a lone `-` as a positional, subcommands by name, `-h` and `--help` on every command, `BadSpec` for duplicate names or a short name that is not one character.
- [ ] `has`, `value` (last one wins), `values`, `positional`, `rest`, `commandPath`.
- [ ] `help(command, path)`: usage line, summary, aligned flag and command columns; `formatError(error)` with a closest-name hint for unknown flags and commands (edit distance 2 at most).
- [ ] Registry row, `stdlib.nra`, regenerated `stdtext.nra`, module-list goldens; SPEC, README and guide text.

## Acceptance

- [ ] `testdata/args.nra` parses a narrac-like spec (`build` with `--release`, `-o <path>`, a repeated `--filter`, a positional file; `run` with `Rest.Raw`) over 15 or more argument lists and prints each `Parsed` or error text.
- [ ] `testdata/args_help.nra` prints the help of the root and of one subcommand (golden output).
- [ ] `testdata/args_errors.nra`: every `ArgsError` case, and the hint text for `--relase`.
- [ ] Error test: `testdata/errors/args_flag_missing_name.nra` (a `Flag {}` without `name`, the missing-field error).
- [ ] The module uses only features of the committed seed: `narrac0` built from `seed/narrac.c` compiles a program importing it (the `stage 2` tests cover std sources).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: about 700 lines of Narra and three samples, no compiler change.

## Notes for the worker

- Do not change narrac in this node. narrac can import `std/args` only after the seed update at the end of the batch; the move (and `-release` becoming `--release` in `scripts/test.sh` and `bench/run.sh`, with no shim) is a later node the orchestrator adds (proposed name narrac-args).
- No wrapping in the help text; no terminal width.
- Codes E1330-E1339 are reserved; none is expected.
