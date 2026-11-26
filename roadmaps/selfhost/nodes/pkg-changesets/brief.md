---
id: pkg-changesets
title: "changesets package: version bumps, changelogs and member tags"
status: done
priority: could
type: feature
story: batch-26
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
spent: 278000
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

## Design (maintainer, 2026-11-26)

The CLI is what matters; the library modules behind it are the worker's call (record them in the outcome).

CLI, binary `changesets`, run from the workspace root:

- `changesets add <member>:<level> ... -m "<summary>"` writes `.changeset/<random-words>.md`. Arguments only, no prompts.
- `changesets status [--since <ref>]` lists pending releases; exits 1 when a member changed since `<ref>` (default `main`) and no pending changeset names it.
- `changesets version` applies every pending file: bumps `narra.toml` versions, writes `CHANGELOG.md` entries, deletes the applied files.
- `changesets tag` creates the `<name>/v<version>` tags that do not exist yet. It never pushes.

Rules:

- Several changesets for one member: the highest level wins. `0.x`: `minor` 0.3.1 to 0.4.0, `patch` to 0.3.2, `major` to 1.0.0.
- A member that depends on a bumped member through `{ workspace = true }` gets a `patch` bump, cascading through chains, with the changelog line "Updated dependency <name> to <version>".
- Only the `version = "..."` line under `[package]` changes; comments, order and spacing stay.
- `CHANGELOG.md`: a new section at the top, `## <version>`, then `### Major changes`, `### Minor changes`, `### Patch changes` as present, one bullet per summary.
- An unknown member, unknown level, missing frontmatter or a member named twice in one file is reported with file and line, and nothing on disk changes.
- Out of scope: a config file (fixed or linked groups), prereleases.

## Acceptance

- [ ] On a fixture workspace of three members (one depending on another), `version` with two changeset files gives the expected manifests, changelogs and deleted files, compared to goldens.
- [ ] `tag` on a temporary git repo creates exactly the missing tags; a second run creates none.
- [ ] Bad files (unknown member, unknown level, no frontmatter, a member named twice) are reported with file and line, and change nothing; a fuzz check feeds random text to the file parser.
- [ ] `status --since` exits 1 for a changed member with no changeset and 0 otherwise, on a temporary git repo.
- [ ] `packages/changesets/README.md` documents the CLI and the bump rules.
- [ ] `sh scripts/test.sh` runs the package tests and passes.

## Sizing

M: a small parser, manifest edits that keep the rest of the file, changelog writing, git calls.

## Notes for the worker

- Edit `narra.toml` files in place: change only the `version` line, keep comments and order. Do not round-trip through `std/toml`.
