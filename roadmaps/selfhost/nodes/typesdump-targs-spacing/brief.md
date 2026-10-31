---
id: typesdump-targs-spacing
title: Keep the spaces inside type arguments in the types dump header
status: active
priority: must
type: fix
outline: false
milestone: m8-selfhost
target:
from:
size: S
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestTypesDump|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-31
owns:
  - roadmaps/selfhost/nodes/typesdump-targs-spacing/**
shared:
  - compiler/internal/narra/typesdump.go
  - selfhost/src/typesdump.nra
  - compiler/testdata/typesdump/**
reserves: []
provides: []
depends_on:
  - on: selfhost-check-builtins
    kind: hard
    why: PR 43 changes the Narra targs printer to copy the Go bug; this fix edits the same line
---

# Keep the spaces inside type arguments in the types dump header

## Goal

The `body ... targs=` header of `narra types` drops every space after a comma, including the ones inside a type. `typesdump.go` builds it with `strings.ReplaceAll(dumpTypeList(b.targs), ", ", ",")`, so `(int, str), bool` prints as `targs=(int,str),bool` instead of `targs=(int, str),bool`. The doc comment (`targs=<type>,<type>`) means only the separator. selfhost-check-builtins (#43) made `typesdump.nra` copy the bug for parity.

## Acceptance

- [ ] Go joins the dumped type arguments with `,` and keeps each type's own text unchanged.
- [ ] `typesdump.nra` does the same (back to `Str.join(parts, ",")` over the per-type strings).
- [ ] At least one golden has a tuple or function type argument and shows the inner spaces; goldens regenerated with `NARRA_UPDATE=1`, and the diff touches only `targs=` header lines.
- [ ] Every selfhost parity test keeps its compared/skipped counts.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

S: two one-line changes plus goldens.
