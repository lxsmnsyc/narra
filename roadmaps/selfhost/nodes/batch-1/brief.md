---
id: batch-1
title: "Batch 1: every detailed should node, one pull request"
status: active
priority: should
type: story
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-09
spent:
owns:
  - roadmaps/selfhost/nodes/batch-1/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 1: every detailed should node, one pull request

## Goal

Run the one-shot strategy the user chose on 2026-11-09: every node that has a full brief lands on `feat/batch-1` as direct commits, with targeted tests per node. At the end: one full check, one regenerated seed, one fresh review and one pull request into main.

Children: func-type-effects, std-tables-split, unused-imports, mutable-bindings, mutable-markers, identity-cycles, sized-numbers, sized-conversions, sized-f32, bytes-type, named-args-syntax, named-args-check, named-args-query, shape-trait, shape-derive, std-json, language-service, lsp-navigation, lsp-incremental. The named-arguments story was folded in; its acceptance is below.

## Gates

The user decided on 2026-11-09 that the orchestrator makes the gate decisions of gated children (shape-trait, sized-numbers, std-json, identity-cycles, mutable-markers), writes each to DECISIONS.md, and lists all of them at the top of the batch pull request. The user may reject any of them before merging. This story is gated, so the pull request waits for the user.

## Acceptance

- [ ] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] Named arguments: SPEC.md section 2 (Functions) describes them; the D12 line "no labeled arguments" is amended; D87 loses "Not built yet".
- [ ] `sh scripts/test.sh` passes on the story branch, and `sh scripts/linux-check.sh` passes on Linux aarch64.
- [ ] The seed is regenerated once at the end and `sh selfhost/bootstrap.sh` reports it current.
- [ ] The pull request lists every gate decision the orchestrator made.
