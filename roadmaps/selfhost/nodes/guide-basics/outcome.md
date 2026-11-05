# Outcome: guide-basics

## Built

- `compiler/internal/narra/guide_test.go`: `TestGuide`. Reads `../../../docs/guide/*.md`, extracts fenced blocks (`narra`, `narra error`, `narra fragment`, `text`) by `contracts/guide-samples.md`. One subtest per block, named `<file>:<line>` (line of the opening fence). A `narra` block runs through `runProgram` in every entry of `modes` (ASan/UBSan debug with leak check, and release), needs exit 0 and stdout equal to the `text` block that directly follows it (only blank lines between), and fails with "needs a `text` block" otherwise. A `narra error` block must start with `// error: <text>`; the test compiles it with `CompileToC` and requires the message to contain the text after stripping `sample.nra:` (so positions are `line:col` counted inside the sample, comment line included). `narra fragment` is skipped. Other fence types (`sh`, `toml`, `text` alone) are ignored.
- `docs/guide/README.md`: all ten chapters with links (6 to 10 are dead until guide-advanced lands) and the "Planned" list.
- `docs/guide/01-getting-started.md` to `05-pattern-matching.md`: 34 subtests in total, all pass. Each chapter links `SPEC.md` sections and D-numbers.

## Deviations

- None from the brief. The contract says "a first line `// error: 3:9: ...`" without the file name; the test strips the sample file name, as described above.

## Stubs left

- None.

## Gotchas

- The compiler lacks things the spec and DECISIONS describe: compound assignment and C-style `for` (D76), `Arrays.map` and other array helpers (std/array has only `length`, `values`, `entries`, `push`, `filled`), `Float.PI`, and explicit type arguments in calls. The chapters avoid them. Where SPEC and compiler disagree, the guide follows the compiler.
- `print` and `${...}` take only strings, so every number goes through `Int.toString` or `Float.toString`.
- An `is` pattern that binds names must be the whole `if` condition (`x is Some(v) && v > 0` fails). Samples use `while (true)` with `if ... else { return }` to walk a linked list.
- Error samples count the `// error:` comment as line 1, so reported lines are one more than in the visible code.
- Chapters 6 to 10 samples should use `Options`/`Results` aliases (plural) as chapter samples 2 to 5 do.
- `.nra` samples in the guide are not formatted by `narrac fmt` and `TestSourcesFormatted` does not cover `docs/`.

## Proposed learnings

- Guide samples are checked by `TestGuide`; run `go test -count=1 ./internal/narra -run TestGuide` after editing any `docs/guide/*.md` | the test is the only link between the prose and the compiler | docs/guide | yes

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestGuide`: pass (34 subtests).
- Broke a sample by hand (output text in chapter 1): TestGuide failed at `TestGuide/01-getting-started.md:16/debug` with want and got. Reverted.
- `cd compiler && go vet ./... && go test -timeout 20m ./...`: see the pull request.

## Size

- Estimated M, actual M.
