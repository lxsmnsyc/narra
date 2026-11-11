---
name: guide-samples
version: 2
status: frozen
defined_by: guide-basics
---

# Contract: guide-samples

## Purpose

How a code sample in `docs/guide/*.md` is marked and checked, so the chapters of guide-basics and guide-advanced can be written at the same time and one test checks them all. spec-samples-test may adopt the same format for `SPEC.md`.

## Interface

A tested program is a fenced block with the info string `narra`, holding a whole file that exports `run`, directly followed by a fenced `text` block with its exact stdout:

````md
```narra
export function run(args: Array<string>): int {
  print("hello")
  return 0
}
```

```text
hello
```
````

A sample that must not compile has the info string `narra error`, and its first line is a comment with the expected message, in the format of `compiler/testdata/errors/` without the file name:

````md
```narra error
// error: 3:9: a `;` ...
export function run(args: Array<string>): int {
  ...
}
```
````

A file of a package sample is a fenced block with the info string `narra file PACKAGE PATH`. `PACKAGE` is one word: the name of the package folder and, when there is no `narra.toml` block, also the package name that imports start with. `PATH` is relative to the package folder and has no `..`, for example `src/count.nra` or `narra.toml`. The body is the file's text.

````md
```narra file shop src/main.nra
export function run(args: Array<string>): int {
  print("hello")
  return 0
}
```

```text
hello
```
````

A fragment that is not a whole program (a signature, a snippet inside prose) uses the info string `narra fragment` and is not run. Use fragments sparingly; prefer a short whole program.

## Behavior

- Programs run like `testdata/*.nra`: in debug mode (`-sanitize`, clean under AddressSanitizer and the leak check) and in release mode, exit code 0.
- An error sample passes when compiling it fails with a message that contains the expected text, in the format of `testdata/errors/`.
- Any other `narra` block without a following `text` block fails the test, so a sample cannot go untested by accident.
- `narra file` blocks with the same `PACKAGE` in one chapter add up in order. A later block with the same `PATH` replaces the earlier one, so a chapter can grow a package step by step.
- A `narra file` block directly followed by a `text` block is a run point: the suite writes the files so far into a temporary folder named `PACKAGE`, runs `narrac run` on it in debug and release mode, and compares stdout with the `text` block.
- A package with no `narra.toml` block gets one whose package and binary are named `PACKAGE` and whose entry is `main.run`.
- A `narra file` block without a following `text` block is fine when a later block of the same package is a run point. If the last block of a package has none, the chapter fails.
- A `narra file` info string without exactly a package and a path fails the chapter.
- `narra fragment` stays for snippets that are not files of a package.
- Samples use only std modules and features the compiler implements.

## Stub

Check a sample by saving it to a file and running `narrac run <file>`, or `narrac run <folder>` for a package.

## Contract tests

The `guide` suite in `selfhost/tests/src/guide.nra`, and its `extractor` case, which checks the fence parser, including `narra file` blocks.

## Changes

- v1: initial, written while refining language-guide. Fences are `narra`, not `ts`, because a website with Narra highlighting is planned (maintainer, 2026-11-05).
- v2: `narra file PACKAGE PATH` blocks for multi-file package samples; chapters 1, 4 and 8 use them. Stub, contract tests and behavior now name the self-hosted test suite instead of the retired Go harness (guide-multifile-samples, 2026-11-12).
