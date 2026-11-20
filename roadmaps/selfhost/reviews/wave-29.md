# Wave 29 report: batch-21, docs that teach every concept

Batch-20 (PR #125) merged and is set `done`. The maintainer picked `v0.1.0` as the first tag, a full release, after the docs teach the concepts the spec describes (2026-11-20). Batch-21 runs on the local branch `feat/batch-21`, in the worktree `../narra-batch-21`.

## Landed

32 children, each one commit after a worker check; compiler changes also got a fresh review.

- Audit and plan: docs-concept-audit (185 concepts from SPEC sections 1 to 15 and the other sources; approved, all 22 writing nodes before the tag, the 20-chapter order), guide-renumber.
- Guide chapters: corrections (tuple index, `const` as `mutable`, `copy`, all 55 std modules, the spec title), basics gaps and numbers (chapter 2), patterns (5), types (4, 7), serialization (7), `Cell` and state (9), files and the system (11, new), generators (12, new), async (14, new), threads (15, new), unsafe code and C (16, new), proofs and compile-time constants (17, new), tool gaps (10, 18), building for other systems (19, new), and a final read-through (index, cross-links, contradictions, SKILL.md).
- Reference: the `/reference` section, `narra.toml`, the commands page with a complete `narrac --help`, the error index (`narrac explain --list --json`, generated), the std reference (55 generated pages from `narrac query interface std/<module>`, with readable builtin signatures checked against the compiler).
- Spec: spec-corrections (loops read arrays, ranges and generators; constant parameters in struct literals; the section 8 samples; `Channel.values` effects; `return` in `unsafe`; `narrac move` moved to Deferred).
- Highlighters: grammar-fixes and grammar-fixes-two (loop labels, named `mutable` arguments, function types in type arguments, capitalized names, `implements` in a type, capitals after `typeof`), for the site and the TextMate grammar.
- Compiler bugs found by the writers, each with a case that fails on the base: sized `parse` without `std/int` failed in the C compiler; a constant holding a function lost its effects; `Thread.scope` with a body that returns nothing failed inside std.
- Also on the branch: the page ToC fix on `/spec` (the indicator jumped back a section).

## Orchestrator rulings

- The C export chain (c-export-functions, c-abi-export, node-addon, wasm-js-glue) stays `later`.
- Capitalized names: two or more capitals or underscores and no lowercase is a constant, any other capitalized name a class, on both highlighters.
- Builtin signature bounds such as `from<T: i8 | i16 | i32>` are display text only; Narra has no union types.
- Generated reference pages are committed, with a staleness case in the `generated` suite, and need only `sh` and `narrac`.
- std-doc-comments and spec-decision-links stay after the release.

## Checks

- `sh scripts/test.sh`: PASS, 10820 cases.
- `sh selfhost/bootstrap.sh`: fixed point, seed current (regenerated in `d9295d57`).
- `pnpm --dir website build`: passes; 0 highlighter problems over 782 fences.

## Still to do

- [ ] `sh scripts/linux-check.sh`.
- [ ] Push `feat/batch-21` and open the batch-21 pull request for the maintainer.
- [ ] After it merges: first-release (`v0.1.0`), which the maintainer pushes; the prerelease paragraph in `docs/install.md` becomes one line for a full release.
