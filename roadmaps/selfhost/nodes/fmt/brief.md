---
id: fmt
title: narra fmt
status: todo
priority: should
type: story
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostFmt|TestSelfhostBootstrap'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: written in Narra and built by the self-hosted compiler; the Go compiler gets no formatter
---

# narra fmt

## Goal

The `fmt` subcommand of the self-hosted compiler formats Narra source into the one layout that `SPEC.md` section 2 promises. It works like oxfmt and Prettier: it prints the AST through a document IR (group, indent, line, softline, hardline) with Wadler/Prettier printing, reflowing to a line width. It is opinionated, with few options. Maintainer decision, 2026-11-05.

## Rules

- Config: a `[fmt]` table in `narra.toml` with `printWidth` (default 100), `tabWidth` (2), `useTabs` (false) and `trailingComma` (`"all"` or `"none"`, default `"all"`). No other options. There is no semicolon option (Narra has none) and no quote option: `SPEC.md` section 2 fixes double quotes for strings and single quotes for characters.
- CLI, as oxfmt: `fmt [paths]` rewrites files in place (a directory means every `.nra` file under it; no path means the current package); `--check` prints each file that would change and exits 1, or exits 0; `--stdin` (or `-`) reads one file from stdin and writes the result to stdout, for editors. The command is `narrac fmt` until the self-hosted compiler takes the name `narra`.
- Blank lines between statements and declarations are kept, at most one, as in Prettier.
- Comments are kept and attached to nodes. Parentheses the author wrote are kept in a side table; the printer adds parentheses only where precedence needs them, and keeps written ones (see `contracts/fmt-trivia.md`).
- Trivia lives in side tables, not in the AST (decision 2026-10-30). The parser change must not change the emitted C or any dump the selfhost tests compare.

## Children

- `fmt-doc`: the document IR and printer, a pure library with unit tests. Freezes `fmt-doc`.
- `fmt-trivia`: the lexer and parser keep comments, blank lines and parentheses in side tables. Freezes `fmt-trivia`.
- `fmt-cli`: config, CLI, driver, declarations and statements, golden harness, idempotence and token checks. Implements `fmt-printer`.
- `fmt-exprs`: expressions, types and patterns, on `fmt-printer`.

Waves: `fmt-doc` and `fmt-trivia` run together, then `fmt-cli` and `fmt-exprs` together. `fmt-cli` merges into the story branch first.

## Acceptance

- [ ] Every child's acceptance holds on the story branch.
- [ ] For every `.nra` file in `compiler/testdata/` (recursively) and `selfhost/src/`: formatting is idempotent, keeps the token sequence (kind and text) and every comment's text, and keeps the AST (the `narrac parse` dump with positions removed).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L as a whole, four M children.

## Notes

- Reformatting the repo with the new formatter is a later change (inbox).
- The language service reuses `fmt-trivia` (doc comments for hover) and the `--stdin` mode (document formatting).
