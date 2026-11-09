# Outcome: mutable-markers

## Built

- `DECISIONS.md` D85: the `mutable` marker goes only on `for ... of` over an array: `for (mutable item of items)`, `for (mutable item of Arrays.values(items))` and `for (const #[index, mutable item] of Arrays.entries(items))`. Six rules (place source, contents element, binding acts as a `mutable` parameter, whole-body borrow of the source, uniqueness before the loop, not in `function*` or `async` yet), every other position with its answer and reason, prior art and cost.
- `SPEC.md` section 2: "Bindings" points to the loop form and says the other markers do not exist; "Control flow" describes the loop form with a sample. Section 6: "In-place updates", "Mutable parameters" (argument bullet) and "Enums and tuples are replaced" no longer say "open".
- `compiler/README.md` "Differences from the spec": the loop form is not built yet.
- `roadmaps/selfhost/nodes/mutable-markers/count.py`: the counts below. `python3 roadmaps/selfhost/nodes/mutable-markers/count.py -v` prints the sites.

## Gate decisions

Chosen by the worker under the orchestrator's delegation (2026-11-09), recorded in D85.

1. **Loop marker: chosen.** `for (mutable item of items)` and `#[index, mutable item]` in `Arrays.entries`. Alternatives: defer it too (keeps D81's index-loop hint as the only fix), or allow it with reads of the source in the body. Why: it is the TypeScript loop D81 turned into an error, the element pointer cannot dangle when the body cannot touch the source, and it reuses the pointer path `mutable items[index]` already has.
2. **Borrow scope: whole source, whole body.** Alternative: allow reads of other elements and forbid only writes. Why: one rule, same as Rust `iter_mut`; the 2 corpus loops never touch the array except through the index.
3. **Pattern markers (`x is Some(mutable y)`, `match` arms, nested, `while ... is`): deferred.** Alternative: allow them with a copy-on-write step on enum payloads. Why: D55 (payloads are shared and never written) and D81 rule 7 win for now; 0 write-backs in selfhost/src; the design (place subject, borrow over the arm, payload uniqueness) is in D85 for when a program needs it. This was the maintainer's example, so it is the item most likely to be overridden.
4. **Destructuring markers (`const { mutable position } = player`, `const #[mutable a, b] = pair`): rejected.** They are local reference bindings with another spelling; tuple elements are not places.
5. **Local reference binding `mutable frame = emitter.frames[index]` (the inbox item): deferred.** D81 rejected a third binding kind. 11 functions (49 writes) could use it, but most interleave `f(mutable checker, ...)` calls the reference would block. A helper with a `mutable` parameter already passes the path once.
6. **`x = f(mutable x)` (the inbox item): rejected,** stays E0929 (D81 rule 10).
7. **Not in `function*` or `async` bodies in the first build.** Alternative: support the pointer as a frame field. Why: sound, but its own emitter path with no user.
8. **Maps (`Maps.entries` values in place) and `mutable` arrow parameters: out of scope.** A map value pointer can be invalidated by a rehash; arrow parameters belong to section 6 (E0028).

## What this means for mutable-bindings

mutable-bindings builds D81 and, on top of it, the D85 loop form. Everything else D81 left open is now closed: pattern, `for (const ...)` and destructured bindings are read-only copies, with no marker. Its E0315-style error for these bindings should get the hint `for (mutable p of players)` when the binding comes from a `for ... of` over an array of a contents type, and the index-loop hint otherwise. If the maintainer wants D85 built in a separate node, mutable-bindings can ship D81 alone first: D85 adds syntax that is a parse error today, so nothing breaks either way.

Plan for the loop form (selfhost only; Go is retired):

- **Parser** `selfhost/src/parser.nra` about line 1820 to 1881 (the `for` head): accept `mutable` where `const`/`let` go, with one name; and accept `#[name, mutable name]` as a `for` pattern when the iterable is `Arrays.entries`. Today the head with a pattern desugars to a hidden name plus a `VarStmt`; the marker form must not desugar that way (a `VarStmt` copies). Add a flag to `ForStmt` (`selfhost/src/ast.nra:278`) or a `ForMutable` statement, with cases in `astdump.nra`, `fmt.nra`, `query.nra` and the walkers.
- **Checker** `selfhost/src/stmts.nra` `checkFor` (line 736): the iterable is `Array.values`/`Array.entries` of a place, or a place of array type; check the place as `checkMutableArgument` does (`exprs.nra:2204`); the element is a contents type (the D81 rule 1 check mutable-bindings writes for parameters); mark the binding local as a reference (`LocalInfo` gains a flag, or reuse `param && isMutable` semantics); walk the body and report any use of a path overlapping the source (reuse the E0319/E0929 overlap test, `exprs.nra:1864`, `stmts.nra:1610`); reject whole assignment (`stmts.nra:371`, the D81 code for assigning a `mutable` parameter) and escape (`exprs.nra:309` E0915 style).
- **Emitter** `selfhost/src/emitstmts.nra` `emitArrayFor` (line 495): take `emitLvaluePointer` of the source once, make it unique (the array `_at` helper `indexPointer` uses, `emitexprs.nra:1090`), loop with `T* v = &arr->items[i]` and no retain of the array or the element, and no release at scope end. `emitstate.nra` `isReference` (line 276) must return true for the binding so reads and writes go through the pointer.
- **Tests:** `testdata/` run samples (struct elements, nested `for (mutable q of p.items)`, `Arrays.entries` form, a shared array copied once, a closure capturing the binding), all under ASan and the leak check; `testdata/errors/` for each new code; a `typesdump` or `astdump` golden for the new syntax; a `fmt` golden.
- **New diagnostic codes needed** (none assigned here; take free numbers from `selfhost/src/diagnostics.nra`): loop source is not a place; element type has no contents; source used inside the body; `mutable` loop binding assigned whole (or the D81 whole-assignment code); `mutable` loop in a `function*` or `async` body. Five, or four with the reuse.
- **Seed rule:** selfhost/src cannot use the form until the seed is regenerated. After that, `check.nra:621` (`for (mutable error of checker.errors)`) and `resolve.nra:532` (`for (const #[index, mutable field] of Arrays.entries(fields))`) are the two candidates.
- **Guide:** `docs/guide/09-memory-and-mutable.md` and `11-narra-for-typescript-programmers.md`, where mutable-bindings already rewrites the `for (const p of players)` hint.

## Counts

`python3 roadmaps/selfhost/nodes/mutable-markers/count.py` (text scans; each count is a lower or upper bound as described in the script).

| Count | selfhost/src | std/src | testdata |
| --- | --- | --- | --- |
| Element-path writes (`a[i].f = v`) | 255 | 0 | 27 |
| `mutable path[...]` arguments | 23 | 0 | 0 |
| Functions writing one element path 3+ times (writes) | 11 (49) | 0 | 0 |
| Index loops over an array | 119 | 0 | 6 |
| Index loops writing their own source by index | 2 | 0 | 0 |
| `for ... of` loops writing their source by index | 0 | 0 | 0 |
| `is` bindings whose subject is stored back in the body | 8 (0 payload write-backs) | 0 | 2 (list walks) |
| Destructuring declarations | 39 | 0 | 56 |
| `for ... of` with destructuring | 0 | 0 | 17 |

The 23 `mutable path[...]` arguments are mostly `mutable emitter.frames[top]` (10) and `mutable checker.bodies[...]...` (6), each passed once per function. The 8 `is` matches are list walks (`current = right`) and one `Map` entry (`builtins.nra:72`), none a struct changed inside an `Option` or enum.

## Deviations

- The brief's acceptance asked for "Go and selfhost sites"; Go is retired, so the plan names selfhost sites only.
- The brief's candidate `for (mutable p of players)` with reads of `players` in the body: forbidden, not weighed further (decision 2).

## Stubs left

- None. Nothing is built.

## Gotchas

- `emitArrayFor` today `consume`s the array and pushes it on `emitter.frames[top]`, so the loop holds a second reference. With a pointer local that would make every write copy (count 2); the marker path must not hold the array.
- The `for` head with a pattern desugars to a `VarStmt`, which copies. The marker form cannot reuse that desugaring.
- `count.py` regexes miss loops whose bound is a local (`for (const i of Ranges.upTo(count))`), so "index loops" is a lower bound.

## Proposed learnings

- None.

## Verification

- `out/narrac check` on a sample with `for (mutable p of players)`: parse error "write `for (const item of items)`" (so the README line holds).
- `python3 roadmaps/selfhost/nodes/mutable-markers/count.py`: the counts above.
- `NARRAC=out/narrac sh scripts/test.sh`: PASS in 135 s (8 shards, 4762 passed, 0 failed, bootstrap passed). Only documents changed.

## Size

- Estimated S, actual S.
