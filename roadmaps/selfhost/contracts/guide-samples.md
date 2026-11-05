---
name: guide-samples
version: 1
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

A fragment that is not a whole program (a signature, a snippet inside prose) uses the info string `narra fragment` and is not run. Use fragments sparingly; prefer a short whole program.

## Behavior

- Programs run like `compiler/testdata/*.nra`: every mode of `modes` in `run_test.go`, clean under AddressSanitizer and the leak check, exit code 0.
- An error sample passes when compiling it fails with a message that contains the expected text, as in `TestCompileErrors`.
- Any other `narra` block without a following `text` block fails the test, so a sample cannot go untested by accident.
- Samples use only std modules and features the compiler implements.

## Stub

Until `TestGuide` lands on the story branch, check a sample by saving it to a file and running `narra run <file>` (`cd compiler && go run ./cmd/narra run <file>`).

## Contract tests

`TestGuide` in `compiler/internal/narra/guide_test.go`, added by guide-basics.

## Changes

- v1: initial, written while refining language-guide. Fences are `narra`, not `ts`, because a website with Narra highlighting is planned (maintainer, 2026-11-05).
