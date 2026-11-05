# Outcome: guide-advanced

## Built

- `docs/guide/06-option-and-result.md` to `10-effects-and-testing.md`, rewritten on 2026-11-05 as a guided tutorial after the maintainer's review. Each chapter follows one thread, and each section builds a program toward a goal, shows its output, then explains it. Restrictions appear as `narra error` samples with the real message and the fix.
  - 6: a reader for `key = value` settings, starting from the `Option` and `Result` that chapters 2 and 5 met. It covers returning an `Option` (find a key; later samples read a `Map` so the lookup is shown once), the `std/option` fallbacks, keeping the reason from `Int.parse`, an error enum with early returns, the unchecked-Result error, a `std/result` version of the same function, and when to panic instead.
  - 7: a shop catalog with books and songs. It covers a `Label` trait, a generic `printCatalog<T: Label>`, `derives Show`, a hand-written `Ordered` for "cheapest", map keys and a generic `countOf<T: Equatable>` (building on chapter 4's `==`), and a wrapper struct for a type you did not declare.
  - 8: a word counter. It starts with std imports (namespace, named, `as`, plural aliases, prelude), then splits into a package, building on chapter 1's package. It covers `export` and `uses alloc` (pointing back to chapter 4), the private-import error, top-level constants and the top-level `let` error, a third module used through aliases, and the import-cycle error.
  - 9: a game player. It recaps copies from chapters 2 and 4, defines the heap and reference counting, then changes a player in a function (constant-parameter error, then spread or `mutable`). Game tasks hit each `mutable` check: heal (missing `mutable`), add to the starting kit (constant), swap two players in an array (two parts of one variable), log hp before and after (read in the same expression), keep a snapshot (return). Closures build on chapter 3: why only constants are captured, a parameter-based alternative, and the `mutable` capture error. It ends with freeing memory.
  - 10: a temperature module. It covers the missing-`uses` errors for `alloc` and `io`, inference for private functions, the callback error and the effect-variable fix, a `std/test` file and its `narra test` output as fragments, `-filter`, the missing-`Show` error, and the `Test.*` outside a test block error.
- Multi-file code (chapter 8) and the test file (chapter 10) are `narra fragment` blocks. The commands and output are `sh` and `text` blocks. I built each package and test file in a scratch folder, and the error text and output shown are what the compiler printed.

## Deviations

- Chapter 8 teaches `narra run .` and `narra build .` from inside the package folder. The form in chapter 1 (`narra build app` from the parent folder) fails when the binary has the same name as the folder: the linker tries to write the binary over the directory (`ld: ... Is a directory`). Chapter 1 belongs to guide-basics.
- Not taught: `Dynamic<Trait>`, `derives Shape`, `export type` (none implemented), `Arrays.map` and sorting (std does not have them), `unsafe` (no source yet; chapter 10 lists only `alloc` and `io`). The earlier draft's "Not yet" and "planned" notes are gone. The index's "Planned" list is guide-basics' to keep.
- The earlier draft taught `export type Token = Text.Token`, which the compiler does not support. That text is gone.

## For chapter 11

The earlier worker's list, kept:

- ch6 Option/Result: `Option<T>` is like `T | undefined`, except that the compiler makes you check it. `Result<T, E>` takes the place of exceptions.
- ch6 Option/Result: for Rust readers, the types are the same, but there is no `?` operator and no `unwrap`.
- ch7 Traits: a trait is like a TypeScript `interface` or a Rust `trait`, but narrower: functions only, and all of a type's traits are listed with the type.
- ch7 Traits: the implementation rule is a stricter form of Rust's orphan rule: only the module that declares the type implements a trait for it.
- ch8 Modules: the import syntax is a subset of ES modules. For Rust readers, a file is a `mod`, and `export` takes the place of `pub`.
- ch8 Modules: export rules that differ from TypeScript: no `export { x } from`, no `export *`, no `export default`, no top-level `let`.
- ch9 Memory: in TypeScript, `const b = a` on an object gives two names for one object. In Narra, assignment copies the value.
- ch9 Memory: a TypeScript object parameter compared with a `mutable` parameter. The callee changes it silently versus only when declared. The change is visible at the call. The parameter cannot be kept. Access is exclusive. The argument must be a `let`, a `mutable` parameter or part of one. This is closer to Rust's `&mut`.
- ch9 Memory: JavaScript closures capture the variable itself. Narra closures cannot capture `let` or `mutable`.
- ch9 Memory: no garbage collector, no borrow checker, no lifetimes and no `&` (for Go and Rust readers).
- ch10 Effects: closer to a checked capability list than to Rust's `async` or `unsafe` coloring. TypeScript has nothing like it: a function that prints and a pure function have the same type.

Added in the rewrite:

- ch6: `Options.withDefault` against `??`: `withDefault` always evaluates the fallback, so `Options.orElse` is the lazy form (`a ?? b()` in TypeScript is lazy).
- ch6: a `Result` that is dropped is a compile error. A TypeScript promise or return value can be dropped silently. `const _ = f()` stands in for `void f()`.
- ch6: `Options.map` and `Results.andThen` against optional chaining (`a?.b`) and `.then` chains.
- ch7: no method calls on trait values: `Label.label(book)` where TypeScript writes `book.label()`. No structural typing: a struct has a trait only when it lists it with `implements` or `derives`, even if it has a function with the right shape.
- ch7: `Show.show(value)` against `JSON.stringify` or `String(value)`. `derives Show` against Rust's `#[derive(Debug)]`.
- ch7: `==` compares contents, also as map keys. In TypeScript, `===` on objects and `Map` keys compare identity, so two equal objects are two keys.
- ch8: package-rooted import paths (`"words/count"`) against TypeScript's relative paths (`"./count"`). The plural std aliases exist because an alias cannot share a name with a type, where TypeScript allows `import * as Array`.
- ch8: an exported function's `uses` list is part of its API, like a TypeScript return type.
- ch9: the classic JavaScript counter closure (`let n = 0; return () => n++`) does not compile. The Narra forms are a parameter, a `mutable` argument or a returned value.
- ch9: copy-on-write arrays against JavaScript's shared arrays: `const before = scores; scores.push(30)` changes `before` in JavaScript and not in Narra.
- ch10: `test` blocks in the source file against separate `*.test.ts` files with Jest or Vitest. `Test.eq` against `expect(a).toEqual(b)`, and `Test.panicsWith` against `expect(f).toThrow(message)`.

## Stubs left

- None. `guide_test.go` is part of the story branch since guide-basics merged (3143d2e).

## Gotchas

- Error samples now carry `line:col:` and the full message, which the harness matches as a substring. Any change to a message's wording or position breaks the sample. That is intended, but the positions count the `// error:` comment line.
- `"player" is constant; declare it with `let` to change it` is the message for assigning to a field of a parameter. For a parameter, the fix is `mutable` or returning a new value, not `let`. Chapter 9 says so in a sentence. A better message would remove that sentence (compiler work, not this node).
- `Book implements Label but has no ... block or `derives Label`` suggests `derives` for a user trait, which cannot be derived. Chapter 7's sample matches only the part before "or".
- An exported function that allocates needs `uses alloc`. Chapter 8 meets it before effects are taught, and shows it as the error the reader hits, with a pointer to chapter 10.
- `Int.parse("80x", 10)` gives the reason `invalid digit for the base`, which chapter 6's output shows.

## Proposed learnings

- A guide sample that quotes a multi-file error or a `narra test` run must come from building that exact package or file, because positions and wording are part of the text | the harness cannot run fragments | docs/guide | no
- Exported functions in guide samples need `uses alloc` as soon as they build a string or an array, so chapters before effects keep exported helpers pure or show the error on purpose | effect checking runs on every exported function, also in a single file | docs/guide | yes (TestGuide)

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestGuide'` with `guide_test.go` from `origin/docs/guide-basics`: `ok narra/internal/narra 22.7s`, every program in every mode, and every error sample.
- A local script ran every `narra` and `narra error` block with the built `narra` (`NARRA_CHECK_LEAKS=1`; `check` for errors): 0 failures.
- The chapter 8 package (three modules) was built and run from the fragments, and each error was produced by editing it: export without `uses alloc`, importing `addWord`, an import cycle. The chapter 10 test file is byte-identical to the file `narra test` ran, including `-filter`.
- `check_prose.py` from the plain-writing skill: `prose: ok` on all five chapters.
- Fast check `cd compiler && go vet ./... && go test -count=1 -timeout 20m ./...` with the harness copied in, on 994f6b5: vet clean, `ok narra/internal/narra 801.2s`.
- After the review fixes: TestGuide `ok narra/internal/narra 24.9s` with the harness from `origin/docs/guide-basics` at b73a6e7. The chapter 8 package was rebuilt from the new fragments (two modules, private import, missing `uses alloc`, three modules, cycle): every quoted output matches. The chapter 10 test file is unchanged.
- Rebased on `origin/feat/language-guide` at 3143d2e (chapters 1 to 5, README and `guide_test.go`): TestGuide over chapters 1 to 10 gives `ok narra/internal/narra 61.4s`. Every link between chapters resolves except README's link to `11-narra-for-typescript-programmers.md`, which guide-typescript writes. Chapter 10 has no "Next:" line until chapter 11 exists.

## Size

- Estimated M, actual M.
