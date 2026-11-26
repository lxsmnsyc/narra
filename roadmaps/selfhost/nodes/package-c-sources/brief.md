---
id: package-c-sources
title: "Packages compile their own C sources"
status: active
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
spent:
owns:
  - roadmaps/selfhost/nodes/package-c-sources/**
  - testdata/packages/c-sources/**
shared:
  - selfhost/src/driver/*.nra
  - selfhost/src/check/*.nra
  - selfhost/src/syntax/*.nra
  - selfhost/src/diagnostics*.nra
  - selfhost/src/driver/diagnostics.nra
  - selfhost/tests/src/*.nra
  - testdata/errors/**
  - docs/guide/**
  - SPEC.md
  - compiler/README.md
reserves: [diag E1660-E1662]
provides: []
depends_on: []
---

# Packages compile their own C sources

## Goal

A package ships C files, `narrac build` compiles them with the same `cc`, target and flags as the emitted C, and links them into the program. An `extern` block names the C file it binds with `from "./path"`. This lets a package write C wrappers (variadic functions such as `snprintf`) and is the base that node `c-packages` builds on for C libraries fetched from upstream.

## Design (maintainer approved, 2026-11-26)

`extern {} from` points to the file:

| `from` | Meaning |
|---|---|
| `"./path"` or `"../path"` | a C file of this package, compiled and linked; must stay inside the package folder |
| `"<package>/<path>"` | a file of a C package dependency (node `c-packages`, not this one) |
| a bare name such as `"z"` | a system library, `-lz`, as today |

A `from "./wrap.c"` block needs no manifest entry: the file it names is compiled. The new manifest table `[c]` holds the rest, all paths relative to the package folder and inside it:

```toml
[c]
sources = ["c/lib/*.c", "!c/lib/main.c"]   # more files to compile; `*` matches files in one folder, `!` removes
include = ["c/lib"]                         # -I folders
defines = ["MAX=64"]                        # -D values

[c.os.Linux]                                # added on top when building for that Os variant
defines = ["USE_LINUX"]
```

- `[c.os.<name>]` uses the `Os` variants of `std/build`: `MacOS`, `Linux`, `Windows`, `Wasi`.
- `[link] search` stays as it is. `[c]` is not `[link]`; keys other than these are E1000 as elsewhere.
- Each file compiles with the same `cc`, `--target` flags (the wasm row for `Wasi-Wasm32`) and sanitizer flags as the emitted C, with warnings off, and links after the Narra C file. A dependency's sources compile as part of the program; one package's `include` and `defines` never apply to another package's files.
- Objects are cached under the existing build cache, keyed by a hash of the file, the flags and the headers it includes (`cc -MD`), so an unchanged rebuild compiles no C file and a header edit recompiles only the files that include it.
- Errors: E1660 a listed or named source file is missing; E1661 a source path outside the package folder; E1662 a C compile error, naming the package, the file and `cc`'s first error line.
- Not in this node: fetched C sources (`[c] git`, `from "<package>/<path>"`: node `c-packages`), C calling Narra (node `c-export-functions`), make or cmake, C++.

## Context

- Source: `inbox.md` 2026-11-20 (guide-c-interop); SPEC section 9 names the variadic gap. `GOAL.md` Decisions 2026-11-21: C-backed packages (`tls`, `crypto`, `sqlite`, `compress`, `intl`) wait for this node and `c-packages`.
- SPEC section 9 (`extern { } from`, `@target` on extern blocks, linking) and section 15 (manifest, `[link]`) change; describe the new forms there and in the C interop guide chapter.
- The parser and checker accept a path in `from`; a path form is checked for the package-folder rule at load time.

## Acceptance

- [ ] A test package ships `wrap.c` with a `snprintf` wrapper behind a fixed signature, bound with `extern { ... } from "./wrap.c"`, and a Narra sample in another package depends on it and calls it, clean under AddressSanitizer and the leak check.
- [ ] A package with `[c] sources`, `include`, `defines` and `[c.os.*]` builds, and an OS-specific define applies only on its OS (checked through a function that reports it).
- [ ] E1660, E1661 and E1662 each have an error test naming the package and file.
- [ ] A second build with no change compiles no C file, and editing an included header recompiles only its includers (checked by the build log or the cache in the test harness).
- [ ] SPEC sections 9 and 15 and the C interop guide describe `from "./path"` and `[c]`.
- [ ] `sh scripts/test.sh` passes and `sh selfhost/bootstrap.sh` reaches the fixed point.

## Sizing

M: one `from` form, one manifest table, a compile step in the driver, an object cache.
