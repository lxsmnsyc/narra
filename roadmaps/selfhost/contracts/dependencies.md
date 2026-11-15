---
name: dependencies
version: 1
status: frozen
defined_by: deps-manifest-lock
---

# Contract: dependencies

## Purpose

The data that flows between the four package dependency nodes: what `narra.toml` declares, what `narra.lock` records, and where a fetched dependency sits on disk. deps-fetch, deps-resolve and package-dependencies build against these shapes while deps-manifest-lock writes the parser and the lock file.

## Interface

```ts
// selfhost/src/manifest.nra
export struct Dependency {
  name: string     // the key in [dependencies], also the import path prefix
  git: string      // a URL or a local path; tests use file:// URLs
  version: string  // "1.2.0": the minimum version, tag "v1.2.0"
  line: int        // narra.toml line, for diagnostics
}
// Manifest gains: version: string, dependencies: Array<Dependency>

// selfhost/src/lockfile.nra
export struct LockEntry {
  name: string
  git: string
  version: string  // the selected version
  commit: string   // the full commit hash of the tag
  content: string  // the git tree hash of that commit
}
export function parseLock(text: string): Result<Array<LockEntry>, LoadError>
export function formatLock(entries: Array<LockEntry>): string  // sorted by name

// selfhost/src/depcache.nra (deps-fetch)
// The checkout of one dependency version, read-only once written.
export function checkoutPath(git: string, version: string): string
// <cache>/git/<sanitized url>/<version>, <cache> = $NARRA_CACHE or $HOME/.cache/narra
```

## Rules

- A version is `MAJOR.MINOR.PATCH`; its tag is `v` plus the version.
- The content hash is the git tree hash (`git rev-parse <commit>^{tree}`), so no new hash function is needed.
- `narra.lock` is a TOML file with one `[[package]]` table per entry, keys in the order of `LockEntry`.
- `Manifest.version` is checked as `MAJOR.MINOR.PATCH` (E0431). A `[dependencies]` value is an inline table on one line, with no trailing comma.
- `narra.lock` starts with a comment header line, then `[[package]]` tables separated by blank lines, sorted by name. A repeated, missing or unknown key in an entry is E0438.
- `manifest.nra` also exports `isVersion`, `stripComment` and `parseString`.
- Codes E0430 to E0439 belong to the manifest and lock readers. E0407 is removed.
- Tests reach the reader through the test-only command `narrac pkg --deps <dir>`.
- Planned version 2 (D101, workspaces): `Dependency` gains `workspace: bool`, and member tags are `<name>/v<version>`. Nodes after workspace-design build against version 1 until that change request is accepted.
- No code path reaches the network in tests. Fixture repositories are made by the test runner with `git init` and tags in its scratch folder.
