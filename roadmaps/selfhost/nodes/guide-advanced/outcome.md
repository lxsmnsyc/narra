# Outcome: guide-advanced

## Built

- `docs/guide/06-option-and-result.md`, `07-traits.md`, `08-modules-and-packages.md`, `09-memory-and-mutable.md`, `10-effects-and-testing.md`. 36 `narra`-family blocks: tested programs with `text` output, `narra error` samples, and `narra fragment` blocks (chapter 8 package layout, `narra.toml`, two-module example, shell commands; chapter 10 test file and `narra test` output).
- Chapter 9 explains why a closure cannot capture `let` or a `mutable` parameter (error samples for both) and compares `mutable` with a TypeScript object reference. It describes today's rules and does not promise the m11 closure work.
- Chapter 10 shows a `std/test` file and `narra test` output as fragments (TestGuide runs `run`), plus `-filter`, `-time`, and the rule that `Test.*` works only in test blocks (error sample).
- Chapters carry no `SPEC.md`, `DECISIONS.md` or D-number references and no TypeScript, Rust or JavaScript comparisons (maintainer's change of 2026-11-05). Removed: about 40 references (SPEC section mentions and D-numbers across the 5 chapters, plus the closing pointer to the spec) and 11 comparisons.

## For chapter 11

- ch6 Option/Result: `Option<T>` is like `T | undefined` that the compiler makes you check; `Result<T, E>` replaces exceptions.
- ch6 Option/Result: for Rust readers, same types, but no `?` operator and no `unwrap`.
- ch7 Traits: a trait is like a TypeScript `interface` or Rust `trait`, with a narrower design (functions only; all traits listed with the type).
- ch7 Traits: the implementation rule is a stricter form of Rust's orphan rule (only the declaring module implements a trait).
- ch8 Modules: the import syntax is the same subset as ES modules; for Rust readers a file is a `mod` with `pub` replaced by `export`.
- ch8 Modules: export rules that differ from TypeScript (no `export { x } from`, no `export *`, no `export default`, no top-level `let`).
- ch9 Memory: in TypeScript `const b = a` on an object gives two names for one object; in Narra assignment copies the value.
- ch9 Memory: table comparing a TypeScript object parameter with a `mutable` parameter (callee can change it silently vs only if declared; visible at the call; cannot be kept; exclusive access; argument must be a `let`, `mutable` parameter or part of one); closer to Rust `&mut`.
- ch9 Memory: JavaScript closures capture the variable itself; Narra closures cannot capture `let` or `mutable`.
- ch9 Memory: no garbage collector, no borrow checker, no lifetimes or `&` (for Go/Rust readers).
- ch10 Effects: closer to a checked capability list than Rust `async` or `unsafe` coloring; TypeScript has nothing like it (a printing and a pure function have the same type).

## Deviations

- Chapter 8 keeps tested programs single-file; the multi-module example is fragments. The harness has no multi-file sample form; guide-basics or a follow-up could add one.
- Not taught because the compiler lacks them: `Dynamic<Trait>`, `derives Shape` (compiler says "not supported yet"), contracts (`requires`/`ensures`). Chapter 7 says so in one line. Not in the brief's scope but the index "Planned" list (guide-basics) may want them.
- Each chapter ends with a "Next:" link to the following chapter; chapter 10 ends without one. Chapter 5 should link to chapter 6 (not mine).

## Stubs left

- None. `TestGuide` was not on my branch.

## Gotchas

- Samples use the error-text-only form `// error: <text>` (substring match), as in `compiler/testdata/errors/`, not `// error: 3:9: ...`.
- Several error messages in the samples are copied from `testdata/errors/`; if messages change, the samples need updating.
- Int/Float values do not interpolate in templates: use `Int.toString`. `Float.toString(9.0)` prints `9.0`. `Show` of an enum prints `Dot(Point { ... })` without the enum name.
- Ch 9 error sample for `mutable` reuse ("so no other argument can use it") and ch 10 ("but its function type does not declare it") are copied from testdata and match by substring.

## Proposed learnings

- Sample programs in the guide must use only std functions listed in `compiler/internal/narra/builtins*.go`; guessing names (`Str.toUpper`, `Arrays.get`) fails | the std surface is small | docs/guide | yes (TestGuide)

## Verification

- A local script extracted each `narra`, `narra error` block of chapters 6 to 10 and ran it (`narra run`, `NARRA_CHECK_LEAKS=1`; `narra check` for errors) against the Go compiler built from the story branch: 0 failures. Not run under ASan or other modes, since `TestGuide` is not on this branch; the programs follow patterns already in `compiler/testdata/`.
- `narra test` on the chapter 10 test file produced exactly the output shown.
- Fast check `cd compiler && go vet ./... && go test -timeout 20m ./...`: see the final report.
- Acceptance item "TestGuide passes on the story branch" is open until guide-basics lands.

## Size

- Estimated M, actual M.
