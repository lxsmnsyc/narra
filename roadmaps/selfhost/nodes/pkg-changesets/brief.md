---
id: pkg-changesets
title: "changesets package: version bumps, changelogs and member tags"
status: todo
priority: could
type: feature
story:
early: false
outline: false
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: M
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/pkg-changesets/**
  - packages/changesets/**
shared:
  - narra.toml
  - scripts/test.sh
reserves: []
provides: []
depends_on:
  - on: std-semver
    kind: hard
    why: bumps and compares versions with std/semver
  - on: process-run-options
    kind: hard
    why: runs git in the repo folder with a set environment
---

# changesets package: version bumps, changelogs and member tags

## Goal

A library and CLI that release workspace members the way changesets does for npm. A contributor adds a file under `.changeset/` naming members and a bump level with a summary. `version` applies the pending files: it bumps each member's `version` in its `narra.toml`, bumps members that depend on a bumped member inside the workspace, writes `CHANGELOG.md` entries, and deletes the applied files. `tag` creates the `<name>/v<version>` git tags for members whose version has no tag yet.

## Context

- Source: maintainer request 2026-11-25. First-party packages are tagged `<name>/v<version>` per member (SPEC section 14, `GOAL.md` Decisions 2026-11-21); nothing computes those versions today.
- It is ecosystem tooling and stays out of narrac and `narra.toml` (no new manifest tables). It reads the workspace from the root `narra.toml` `[workspace] members` and each member's manifest.
- File format: frontmatter lines `"<member>": patch|minor|major` between `---` lines, then the summary. Parse only that; it is not YAML.
- Bump rule for unreleased versions: follow the changesets behavior for `0.x` (a `minor` stays `minor`). Write the rule in the README.
- `status` lists pending changes and exits 1 when a changed member has no changeset, for use in checks.

## Acceptance

- [ ] On a fixture workspace of three members (one depending on another), `version` with two changeset files gives the expected manifests, changelogs and deleted files, compared to goldens.
- [ ] `tag` on a temporary git repo creates exactly the missing tags; a second run creates none.
- [ ] Bad files (unknown member, unknown level, no frontmatter) are reported with file and line, and change nothing.
- [ ] `sh scripts/test.sh` runs the package tests and passes.

## Sizing

M: a small parser, manifest edits that keep the rest of the file, changelog writing, git calls.

## Notes for the worker

- Edit `narra.toml` files in place: change only the `version` line, keep comments and order. Do not round-trip through `std/toml`.
