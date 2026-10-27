# Lang

Lang is a draft design for a compiled language with TypeScript-style syntax and Rust-style semantics. It is memory safe without a garbage collector, and builds native binaries through C. "Lang" is a working name.

Status: specification draft. There is no compiler yet.

## Goals

- **Simple.** One way to write each thing. A small grammar with no ambiguity.
- **Type safe.** Static types, type inference, no null, exhaustive pattern matching.
- **Memory safe.** No dangling pointers, no use after free, no data races in safe code.
- **Cheap artifacts.** Small binaries and a runtime of a few kilobytes.
- **Fast builds.** Modules compile in parallel, and a change that keeps a module's interface does not rebuild its dependents.
- **Agent oriented.** Code an agent can read, search and change without guessing, and a compiler an agent can query.

## Files

- `SPEC.md` is the language specification.
- `DECISIONS.md` lists the design decisions, with the alternatives that were rejected and why.

## Prior art

- [Roc](https://www.roc-lang.org/) is the closest existing language: ML semantics, reference counting, native output.
- [Koka](https://koka-lang.github.io/) for Perceus reference counting and effect inference.
- [Lean 4](https://lean-lang.org/) for reference counting with in-place reuse, compiled to C.
- [TypeScript](https://www.typescriptlang.org/) for syntax.
- [Rust](https://www.rust-lang.org/) for enums, pattern matching, `Option` and `Result`.
- [Hylo](https://www.hylo-lang.org/) for mutable value semantics, the model behind `mutable` parameters.
- [Vale](https://vale.dev/) for generational handles.
- [Go](https://go.dev/) and [Zig](https://ziglang.org/) for build speed and C interop.
