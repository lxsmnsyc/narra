# Outcome: guide-typescript

## Built

- `docs/guide/11-narra-for-typescript-programmers.md`: chapter 11, optional reading for TypeScript programmers. 18 topic sections: what carries over; statements and operators; numbers; strings and printing; conditions; arrays and maps; copies instead of shared references; missing values; errors without exceptions; functions; closures; classes and objects; interfaces and traits; unions and enums; modules; changing a value in a function; effects; tests; and a short list for Rust and Go readers. 55 tested samples (programs and `narra error` samples) and one `narra fragment` (a `test` block, checked by hand with `narra test`). Each habit links to the chapter and section that teaches the Narra form.
- Covers every item of both "For chapter 11" lists (guide-basics, guide-advanced) and every `ts_*` diagnostic in `compiler/testdata/diagnostics/` and `compiler/testdata/errors/`: `number`, single quotes, `+` on strings, `.length`, `T[]`, method calls (`items.push`), `Arrays` without an import, `null`, `throw`, `try`/`catch`, `for ... in`, `class`, `interface`, `new`, `this` (use and declaration), default imports. Also habits found by probing: `;`, `===`, `+=`, `undefined`, overloads, nested functions, field shorthand, `export default`, relative import paths, top-level `let`, structural typing, method calls on trait values, the counter closure.
- `docs/guide/README.md` already had the index line ("optional reading"); unchanged.

## Deviations

- Added one "Next:" line at the end of `docs/guide/10-effects-and-testing.md`, linking chapter 11. guide-advanced's outcome left it for when chapter 11 exists. Outside this node's owned paths.
- `??`, `?.`, `switch`, `++`, optional and default parameters, explicit type arguments and C-style `for` are taught in prose with the Narra form as a sample, not as `narra error` samples: their compiler messages are parse errors that do not name the habit (for example `switch` gives `expected the end of the statement, found "{"`, and `f<int>(x)` gives four unrelated type errors).

## Stubs left

- None.

## Gotchas

- The `.length` message says `Array.length(x)`, but the module alias is `Arrays` (an alias cannot be `Array`). The chapter quotes the message as is. A compiler fix would change the sample.
- `unknown module "Arrays"; import it first, for example ...` ends with `"std/..."`, not `"std/array"`; the sample matches only the part before the example.
- Habits with poor messages (could get TypeScript-habit diagnostics later): `switch`, `x++`, `a ?? b` (reports a multi-line conditional), `a?.b`, `name?: T`, `x: T = default`, `...rest`, `f<int>(x)`, `export default`/`export *`/`export { } from` (all `expected a declaration`), `console.log` (unknown name plus a method-call error), `type X = A | B`.
- A `mutable` parameter cannot be returned, so a "counter" helper must return nothing or take a plain parameter.

## Proposed learnings

- TypeScript comparisons belong only in `docs/guide/11-*.md`; chapters 1 to 10 stay standalone | maintainer rule 2026-11-05 | docs/guide | no

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestGuide`: pass (all chapters; 55 samples in chapter 11, every program in debug with ASan, UBSan and the leak check, and in release).
- The test fragment was saved to a file and run with `narra test`: `1 passed, 0 failed`.
- Every link in chapter 11 resolves to an existing file and heading (local script).
- `check_prose.py` (plain-writing): clean on chapter 11.
- Full check `cd compiler && go vet ./... && go test -timeout 45m ./...`: pass (vet clean, `ok narra/internal/narra 349.3s`).
- Self-review, one round: fixed an unclear enum example, split the Rust and Go note, dropped a `?` operator mention from the TypeScript section, and fixed prose-checker findings ("JavaScript", long sentences). Second check clean.

## Size

- Estimated S, actual S.
