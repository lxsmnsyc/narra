# Narra

## Install

On macOS and Linux:

```sh
curl -fsSL https://github.com/lxsmnsyc/narra/releases/latest/download/install.sh | sh
```

On Windows, other systems and building from source, see [docs/install.md](docs/install.md). The language guide is in `docs/guide/`.

## What Narra is

Narra is a compiled language with Rust-style semantics. It is memory safe without a garbage collector, and builds native binaries through C. Its syntax looks familiar to TypeScript programmers, but Narra is a separate language: TypeScript code does not compile as Narra.

The name comes from the narra, the national tree of the Philippines, a hardwood known for lasting a long time.

Status: version 0.1.0 is the first release. Syntax and the standard library can still change. The compiler, `narrac`, is written in Narra (`selfhost/`).

To have a coding agent write Narra with `narrac`, add the skill in [skills/narra](skills/narra/SKILL.md) to Claude Code: `/plugin marketplace add lxsmnsyc/narra`, then `/plugin install narra@narra`. In a trial of 10 tasks, an agent passed 19 of 20 runs with the skill and 15 of 20 without. See the [results](skills/narra/evals/results.md).

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
- `selfhost/` holds the compiler, `seed/` the C seed that builds it, `testdata/` the test samples and `std/src/` the standard modules written in Narra.
- `compiler/README.md` lists what the compiler supports.

## Working on the compiler

The checked-in C file `seed/narrac.c` builds the compiler with `cc` alone: `sh selfhost/bootstrap.sh`. `sh scripts/test.sh` runs the test suites. [seed/README.md](seed/README.md) explains the seed. There is no Go compiler.

## License

MIT, see [LICENSE](LICENSE).

## Prior art

- [Roc](https://www.roc-lang.org/) is the closest existing language: ML semantics, reference counting, native output.
- [Koka](https://koka-lang.github.io/) for Perceus reference counting and effect inference.
- [Lean 4](https://lean-lang.org/) for reference counting with in-place reuse, compiled to C.
- [TypeScript](https://www.typescriptlang.org/) for syntax.
- [Rust](https://www.rust-lang.org/) for enums, pattern matching, `Option` and `Result`.
- [Hylo](https://www.hylo-lang.org/) for mutable value semantics, the model behind `mutable` parameters.
- [Vale](https://vale.dev/) for generational handles.
- [Go](https://go.dev/) and [Zig](https://ziglang.org/) for build speed and C interop.
