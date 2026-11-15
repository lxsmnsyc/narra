---
name: dependencies
version: 0
status: draft
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
- No code path reaches the network in tests. Fixture repositories are made by the test runner with `git init` and tags in its scratch folder.
