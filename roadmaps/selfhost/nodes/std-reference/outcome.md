# Outcome: std-reference

## Built

- `scripts/std/generate.nra`: a Narra program run as `narrac run -release scripts/std/generate.nra <narrac> <repository> <folder>`. It needs only the compiler under test. It lists the builtin modules from `narrac builtins` (no `std/internal/*`) and the source modules from `std/src/*.nra` (no `ext__` or `internal__` files), runs `narrac query --json interface <module>` for each, decodes it with `std/json`, and writes one page per module plus `README.md` into `<folder>`. A Markdown file in `<folder>` that no module wrote is removed. `scripts/std-reference.sh` is the wrapper (`NARRAC=... sh scripts/std-reference.sh` writes `docs/reference/std`).
- `listModuleInfo()` in the generator holds, per module, the import alias, a one-line description and the guide chapter. The generator stops when a public module has no row or a row names no module. The chapter title comes from the chapter's first heading; when a chapter heading names the module (`## Iterators: \`std/iterator\``), the link goes to that section. A chapter value `NN-name#anchor` fixes the section and the generator stops when the chapter has no such heading: the sized integer modules link `02-values-and-types#sized-integers`, `std/char` links `08-modules-and-packages#importing-from-the-standard-library`.
- `docs/reference/std/`: 55 module pages (`std/async/time` is `async/time.md`) and `README.md`, the index with one line per module. Served at `/reference/std` and `/reference/std/<module>`.
- Page layout (`docs/reference/std/string.md`):
  - `# std/string`, the description, an `import * as Str from "std/string"` fence, the guide link, the "generated, do not edit" line.
  - `## Types`, one `### Name` per type: its signature (a struct, enum or trait body gets one member per line), then `Functions:` with links to the functions that take or return it (a builtin type uses the registry's `functions` list; a source type lists the functions whose parameters or result name it). Types with the same list sit one after the other, with one `Functions of A, B:` line after the last.
  - `## Constants`, one `### name` per constant.
  - `## Functions`, one `### name` per function: the signature, `Effects:` (the list after the last top-level `uses`, or `none`), and `Details:` with the registry's display note when there is one. A note is split at `; `; a clause that every function of a page repeats (three or more functions) is printed once under `## Functions` as `Details for every function below:`.
  - Signatures sit in `narra fragment` fences, so they are highlighted, never parsed as Markdown, and the guide and site checks do not compile them. A note is prose with `\ \` < > & | * _ [ ]` escaped.
- `docs/reference/README.md` links `[Std](std/README.md)` in place of the "Coming" line.
- Website: the README of a reference subfolder is that folder's index. `std/README.md` is served at `/reference/std` (`website/src/pages.ts`), links to it are rewritten to `/reference/std` (`website/plugins/markdown.ts`), and search uses that URL (`website/plugins/search.ts`).
- Case `generated/stdreference` in `selfhost/tests/src/generated.nra`: runs the generator into a temporary folder and compares every page with the committed one. It fails when a page differs or is missing, and when a committed page belongs to no module. `sh scripts/test.sh --suite generated --update` runs the generator on `docs/reference/std` itself.

## Deviations

- The generator lives at `scripts/std/generate.nra`, next to `scripts/errors/generate.nra`. The brief's `owns` list names only `scripts/std-reference.sh`.
- Three small website edits outside `website/src` (`plugins/markdown.ts`, `plugins/search.ts`) so the index is `/reference/std` and not `/reference/std/README`.
- Descriptions, aliases and chapters are a hand table in the generator. Descriptions of each export stay out of scope (node std-doc-comments).
- `std/char` has no chapter that teaches it; its page links the std module list of chapter 8.
- `containsPage` in `selfhost/tests/src/generated.nra` stays a second copy of `containsText` in the generator: the generator is a standalone script run by `narrac run`, and the test program cannot import it.

## Stubs left

- None.

## Gotchas

- A change to a registry row, a std source export or the guide headings that name a module makes `generated/stdreference` fail until `--update` runs. Rebuild the compiler first, since the pages come from the compiler under test.
- A new public std module stops the generator until it gets a row in `listModuleInfo`.
- Pages use `###` for each type, constant and function, so the page ToC and search list each export. `std/simd` has the longest ToC (20 types, 22 functions); its 20 types fall in 3 groups with 3 `Functions of` lines.
- The effects line reads the last ` uses ` outside brackets. A result type that is a function type with its own `uses` would be misread; no std export has one today.

## Proposed learnings

- A new public std module needs a row in `listModuleInfo` of `scripts/std/generate.nra` | the std reference describes and links it | scripts/std | yes (generated/stdreference fails)

## Verification

- `NARRAC=<built> sh scripts/test.sh --suite generated`: `PASS, 26 cases`. With a line appended to `docs/reference/std/async/time.md` it fails (`docs/reference/std/async/time.md is stale`); with an extra `docs/reference/std/gone.md` it fails (`belongs to no std module`); `--update` restores both.
- `NARRAC=<built> sh scripts/test.sh --suite sources`: `PASS, 1 cases`.
- `pnpm --dir website build`: passes, 85 pages prerendered, 56 of them under `/reference/std`. Checked in the built HTML of `array` and `string` (builtin), `iterator`, `args` and `async/time` (source) and `ptr` (builtin with a note): generic signatures such as `entries<T>(items: Array<T>): Array<#[int, T]>` render as code with no stray tags, every `#` link on the page has its anchor, the guide section links (`/guide/13-standard-library#iterators-stditerator` and 8 more) resolve, `/reference` links `/reference/std`, and the search index holds the std sections. `pnpm --dir website typecheck`: passes.
- After review fixes: `generated` PASS, 26 cases; `sources` PASS, 1 cases; staleness case still fails on a changed page; `pnpm --dir website build` passes; the simd page went from 45 headings and 416 in-page links to 45 and 61. Full `sh scripts/test.sh`: `PASS, 10820 cases in 190s`.

## Size

- Estimated M, actual M.
