---
goal: Replace copy-on-write value semantics for arrays and structs, which hide where copies happen and let TypeScript-valid code mean something different, with a model whose copies and shared mutation are visible in the source
date: 2026-11-09
rounds: 3 of 3
status: approved
roadmap: roadmaps/selfhost
supersedes: none
decision: DECISIONS.md D94
---

# Debate: visible copies and shared writes

## Settled structure

Keep value semantics and TypeScript-style in-place writes (`a[0] = 9`, `u.age = v`, `mutable` parameters, `for (mutable ...)`, `Cell` as the shared reference). Add one compile rule, the shared write, so that TypeScript-valid code either means the same thing in Narra or does not compile, and no copy happens that the source does not show.

1. **The shared-write rule** (P1, P17, O17). A write to a place is a compile error while another live name shares that place's root or a containing path. The check runs before the write clears the facts on the place.
   - Tracked: every struct (including all-scalar ones like `Pos`), arrays, maps, `FixedArray`, `Box`, `StringBuilder`, generators and coroutines. Not tracked: strings, scalars and enums.
   - Sharing works in both directions and through chains: after `const items = a.items`, a write through `items` conflicts with a later use of `a.items`, and after `const b = a; const c = b`, all three share.
   - Passing `mutable R` writes `R` and everything below it; passing `mutable R.f` writes only `R.f` and below.
   - A non-escaping closure that writes a captured `const` counts as a write to it where the closure is passed.
   - Always shared with their source: plain parameters, pattern and `for (const ...)` bindings, captured values, and values read from identity types (`Cells.get`, `await`, a received message) (P9 as kept by P17).
   - Assigning a whole place ends its sharing: `const saved = checker.labels; checker.labels = []` is fine, and `saved` keeps the old array, as in TypeScript.
   - Liveness is checked inside one function body only. Errors name the write, the shared place, the sharing name and its later use (P17):
     ```ts
     const info = emitter.checker.consts[constIndex]
     startFunction(mutable emitter, ...)
     // error: startFunction(mutable emitter, ...) may change
     //   emitter.checker.consts[constIndex], which info still uses at line 158
     // fix:   const info = copy(emitter.checker.consts[constIndex])
     ```
2. **`copy(x)`** (P11, P12). The one way to duplicate a value, at any depth. At the call it is a single retain; the checker treats the result as unshared, and copy-on-write then copies only the nodes on each written path, once each. `Arrays.copy`, `Maps.copy` and spread stop being copy tools.
3. **The `from` result marker** (P8, P10, O13). Nothing is inferred across functions. A function whose result shares a parameter says so: `function mergeSort(items: Array<SortItem>): Array<SortItem> from items`. Returning a value that shares a parameter without `from` is an error, checked the way D46 checks `escaping`. About 30 to 50 functions in the compiler need it.
   - Refinement after the debate (maintainer review): `from` is part of the function type, like effects (D86): `(grid: Grid) => Array<int> from grid`. A function with fresh results fits a slot with `from`; a function with `from` does not fit a slot without it (error with a fix that adds the marker). Open for the node: how a closure that returns a captured value states it, and generic results (`from items` covers the whole parameter).
4. **The debug copy check** (P4 as kept by P11, P18, O18). Debug builds report every copy-on-write that no `copy(...)` explains, at exit like the leak check (D80). `copy(x)` grants a credit that lasts only while the sharing it created lasts, with debug-only hooks at every reference-count decrement. It covers heap objects; all-scalar structs rely on the compile-time rule, backed by error samples for each D2 form. Release builds are unchanged. It replaces the unbuilt `--report-copies`, `NARRA_TRACE_COPIES` and `@unique` (J3).
5. **`Cell`** (P13). `Cells.modify` is not added. A value from `Cells.get` or `Cells.update`'s callback shares with the cell, so changing it needs a visible `copy`.

Order (P20), one worker at a time:

| # | Node | Tokens |
|---|------|--------|
| 1 | `copy(x)`, the credit set and the decrement hooks | 150k to 225k |
| 2 | The `from` marker, with a fix that adds `from p` | 150k to 225k |
| 3 | The shared-write analysis, as a warning | 450k to 600k |
| 4 | Fix the compiler's rejected sites: `narrac apply-fix` for `copy` and `from`, then the save/restore sites by hand | 200k to 350k |
| 5 | The warning becomes an error; the debug copy check and D2 error samples | 150k to 200k |
| 6 | SPEC section 6, D81 rule 5, D27, the guide | 80k to 100k |

## Budget

- device: one worker at a time; no runtime or memory cost in release builds (J9, P12, P20)
- time: six nodes in sequence, after the D92 nodes (P20)
- tokens: about 1.45M, range 1.18M to 1.70M (P20, O14, J8)
- language-design: D2 restored, every copy visible as `copy(...)`, about 120 to 220 edits in the compiler, 30 to 50 `from` marks (P17, P19, O10, O13)

## Rejected

- Immutable arrays and structs with `ref` and a mutable array type (A1): the maintainer kept in-place writes (Q1); rebinding a shared value still copies silently, so it needs the same analysis on top of rewriting about 6,000 `mutable` uses and 700 writes (O8, J13).
- Inferring across functions whether a result shares a parameter (P2): the marker-less inference D46 rejected; edits in one function would raise errors in unedited callers (O3).
- Shallow copy tools (`Arrays.copy`, `Maps.copy`, spread) as the explicit copy (P4): a copied struct still shares its nested arrays, so the next nested write is shared again (O4).
- `Cells.modify` (P5): reaching the same cell twice needs a panic (D92) or a hidden copy (O5).
- The first cost estimates, 650k and 1.2M (P7, P14): below calibration (O2, O14).

## Prior art

- Swift kept copy-on-write values but added `consume`, `borrowing` and `copy` to control copies (SE-0366, SE-0377) (J10).
- Koka (Perceus) and Lean 4 reuse unique values in place but still copy shared values silently (J11).
- OCaml and Reason have `ref` and mutable arrays with reference semantics, made safe by a tracing GC; Hylo makes copies explicit with `.copy()`; Rust moves on assignment and needs `.clone()` (J12).
- Rust's borrow checker before non-lexical lifetimes (RFC 2094) shows liveness-based errors moving with unrelated edits; Hylo has no production users yet (O6).

## Open risks

- Errors depend on liveness, so adding a later use can turn an accepted write into an error (O6): accepted, because each error names the later use and the fix is a free `copy`.
- The analysis may reject more correct code than the scans estimate; the counts come from text matching, not a parse (O1, O10): measured in node 3, which ships as a warning first, so the real count is known before any code must change.
- The debug credit can be used by the wrong holder, so at most one hidden copy per `copy(...)` call can go unreported (O18): accepted.
- Save/restore sites with a `mutable` call between the save and the reset: 4 to 31 need a hand edit (O9): handled in node 4.

## Inferred choices

- Liveness inside one function body only, with signature markers for what crosses functions (P8, P10): from D46, which requires costs to be written in signatures.

## Answers from the user

- Q1: keep TypeScript-style in-place writes and add the shared-write compile error, rather than immutable arrays and structs with `ref` (2026-11-09).

## Overruled

None.

Debate spend: about 690k tokens over three rounds.
