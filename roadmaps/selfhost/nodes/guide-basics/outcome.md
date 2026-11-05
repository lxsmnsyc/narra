# Outcome: guide-basics

## Built

- `compiler/internal/narra/guide_test.go`: `TestGuide`. Reads `../../../docs/guide/*.md`, extracts fenced blocks (`narra`, `narra error`, `narra fragment`, `text`) by `contracts/guide-samples.md`. One subtest per block, named `<file>:<line>` (line of the opening fence). A `narra` block runs through `runProgram` in every entry of `modes` (ASan/UBSan debug with leak check, and release), needs exit 0 and stdout equal to the `text` block that directly follows it (only blank lines between), and fails with "needs a `text` block" otherwise. A `narra error` block must start with `// error: <text>`; the test compiles it with `CompileToC` and requires the message to contain the text after stripping `sample.nra:` (so positions are `line:col` counted inside the sample, comment line included). `narra fragment` is skipped. Other fence types (`sh`, `toml`, `text` alone) are ignored.
- `docs/guide/README.md`: all ten chapters with links (6 to 10 are dead until guide-advanced lands) and the "Planned" list.
- `docs/guide/01-getting-started.md` to `05-pattern-matching.md`: 34 subtests in total, all pass. Revised 2026-11-05 to the standalone style: no SPEC.md or DECISIONS.md references and no TypeScript or Rust asides in chapters 1 to 5 or the README. The README lists chapter 11 as optional reading (dead link until guide-typescript lands).

## For chapter 11

Comparisons removed from the chapters, for guide-typescript:

- Ch1 intro: Narra is TypeScript syntax with Rust semantics.
- Ch1 syntax: no semicolons, and a `;` is an error with a fix.
- Ch1 functions: parameter and result types always written.
- Ch1 errors: TypeScript habits get a message that says what to write.
- Ch2 bindings: `const` is deep (a TypeScript `const` array can still change).
- Ch2 bindings: no `+=` or `++` (not in the compiler yet, so also a Planned item).
- Ch2 numbers: one `number` in TypeScript versus `int`, sized integers and `f64`, with no implicit conversion.
- Ch2 strings: both quote styles are strings in TypeScript; here `'a'` is a `char`.
- Ch2 strings: no `+` for strings, use templates, and template parts must be strings.
- Ch2 conditions: no truthiness, and `===` is an error.
- Ch2 arrays: `T[]` is not a type, write `Array<T>`; no methods (`items.push`).
- Ch2 arrays: assignment copies (TypeScript shares the reference).
- Ch2 loops: no `for ... in`; C-style `for` is not there yet.
- Ch2 ternary: same form as TypeScript, `bool` condition only.
- Ch3 functions: no optional parameters, varargs, overloads or currying; no method calls; no nested function declarations.
- Ch3 closures: a closure cannot capture `let` (a TypeScript closure sees later assignments).
- Ch4 structs: no `class`, `this`, `new` or `interface`; `{ name }` shorthand is not an object literal.
- Ch4 enums: replace TypeScript `enum`, tagged unions with `kind`, and class hierarchies.
- Ch5 match: replaces `switch`; no fallthrough.
- Ch5 block expressions: `{ ... }` in value position is a block, not an object literal.
- Ch5 ternary: same as TypeScript.

References removed: 38 SPEC.md/DECISIONS.md links, section references and D-numbers (D12, D15, D20, D21, D22, D23, D24, D26, D30, D37, D38, D39, D40, D43, D46, D52, D53, D55, D60, D62, D68, D74, D75, D76 and SPEC sections), including the SPEC links in the README Planned list. 20 comparisons removed, listed above.

## Deviations

- None from the brief. The contract says "a first line `// error: 3:9: ...`" without the file name; the test strips the sample file name, as described above.

## Stubs left

- None.

## Gotchas

The first item is for the roadmap only (it names SPEC and D-numbers); do not repeat it in the chapters.

- The compiler lacks things the spec and DECISIONS describe: compound assignment and C-style `for` (D76), `Arrays.map` and other array helpers (std/array has only `length`, `values`, `entries`, `push`, `filled`), `Float.PI`, and explicit type arguments in calls. The chapters avoid them. Where SPEC and compiler disagree, the guide follows the compiler.
- `print` and `${...}` take only strings, so every number goes through `Int.toString` or `Float.toString`.
- An `is` pattern that binds names must be the whole `if` condition (`x is Some(v) && v > 0` fails). Samples use `while (true)` with `if ... else { return }` to walk a linked list.
- Error samples count the `// error:` comment as line 1, so reported lines are one more than in the visible code.
- `TestGuide` also fails on an info string that starts with `narra` (any case) but is not `narra`, `narra error` or `narra fragment`, on a fence that is never closed, and when no guide file matches. `TestGuideExtract` covers the first two.
- Chapters 6 to 10 samples should use `Options`/`Results` aliases (plural) as chapter samples 2 to 5 do.
- `.nra` samples in the guide are not formatted by `narrac fmt` and `TestSourcesFormatted` does not cover `docs/`.

## Proposed learnings

- Guide samples are checked by `TestGuide`; run `go test -count=1 ./internal/narra -run TestGuide` after editing any `docs/guide/*.md` | the test is the only link between the prose and the compiler | docs/guide | yes

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestGuide`: pass (34 subtests).
- Broke a sample by hand (output text in chapter 1): TestGuide failed at `TestGuide/01-getting-started.md:16/debug` with want and got. Reverted.
- `cd compiler && go vet ./... && go test -timeout 20m ./...` passed on 5278a62, about 13 min (machine loaded by other agents).
- After review fixes: `cd compiler && go vet ./... && go test -timeout 30m ./...` passed on the review-fix commit, about 24 min (loaded machine). A temporary guide file with ```` ```narra errror ```` failed at `99-tmp.md:3`, and one with an unclosed fence failed with `line 3: the fence is never closed`; both removed.

## Size

- Estimated M, actual M.
