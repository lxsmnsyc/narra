---
goal: Decide what Narra does when code writes a value that another live name shares (const synonym = prices; Arrays.push(mutable prices, 300))
date: 2026-11-16
rounds: 2 of 2
status: approved
roadmap: roadmaps/selfhost
supersedes: roadmaps/selfhost/debate/value-semantics/debate.md
decision: DECISIONS.md D124
---

# Debate: what `let b = a` means

## Settled structure

A second name for a value that changes in place is an error at the line that makes it (Q3). `copy(a)` is the way to get a second value. One value that can change has one name.

```narra
let prices = [100, 200]
const synonym = prices
// error E0971: synonym would be a second name for the array in prices.
//   Use prices, or copy(prices) for an array of its own.
```

- **What counts as a rename.** A `const` or `let` initializer, or the right side of `=`, that is a bare name: a local, a parameter, a pattern or loop binding, or a top-level constant. The rule applies when the name's type is one that D94 tracks: a struct (all-scalar ones too), `Array`, `Map`, `FixedArray`, `Box`, `StringBuilder`, a generator or a coroutine.
- **Still allowed.** Renames of numbers, `bool`, `char`, `string`, `bytes`, enums (`Option` and `Result` included), tuples and function values. These never change in place, so a second name cannot see a change.
- **Not renames.** These keep today's rules (E0978, E0976): reading a field or element into a name (`const items = bag.items`), passing an argument, returning, struct, array and map literals, captures, and `copy(a)`.
- **The fix.** `narrac apply-fix` offers `copy(a)`. Where the compiler's pattern of a default reassigned later applies (`let ordered = arguments; if (...) { ordered = matched }`), it is rewritten as an `if` or `match` expression, or as a helper, by hand.

Order:

1. **The rule (node `restrict-renames`).** The check, code E0971, the `copy(a)` fix, error samples, and the compiler sites rewritten. The seed is regenerated if the compiler's own sites change.
2. **Docs (same node).** SPEC section 6 gets a new "Renames" paragraph, with D94 and D104 amended by D124. Guide chapter 2, the 3 renames in chapters 9 and 13 (J15), the E0978 text in chapter 9, and home section 02 (`const saved = copy(hero)` already fits).

## Budget

- device: none at run time; one check per binding in the shared-write walk.
- time: one node.
- tokens: about 100k to 200k (J15: about 13 typed compiler sites, plus samples and docs).
- compiler size: about +50 to 150 lines.

## Rejected

- **Silent copy-on-write, as in Swift (P1 to P3, O7):** `copy` already exists, so a hidden copy makes no sense (Q1). It also hides quadratic copies (O1, measured at 20x).
- **Rewording E0978 only (P6, P13):** it leaves `let b = a` as counter-intuitive as before (Q1).
- **References for every array and map (P7):** it breaks J5 (no cycles without a collector) and J6 (no data races).
- **Implicit local alias (P9):**
  - Proof T1 showed it is buildable, with a non-retaining alias (J13).
  - It needs: the store-into-`mutable` gap closed first; E0978 kept for a container `mutable` call (a use-after-free, J14); pointer guards after branches; an inline-struct design.
  - Its limits: it still stops at plain parameters, and it would cost 1.05M to 1.7M tokens and add 900 to 1,900 lines (Q3).
- **Move, as in Rust (P12):** the restriction errors at the binding without making the old name unusable (Q3).

## Parked

- **Explicit `const b = alias a` (P14):** to be added if code turns out to need two names for one value. It uses T1's design and carries the same J14 constraints.

## Proofs

- **T1 for P9:** can a non-retaining alias write in place with no hidden flag and clean checks?
  - Done in about 5 minutes (cap 30).
  - Yes for the snippet, for reassignment and for a one-branch merge with a pointer guard.
  - The judge's follow-up found a use-after-free through a container `mutable` call (J14).
  - Files are in `evidence/T1/`.

## Prior art

- **Go and Rust (J12, measured):** Go slices and maps alias, and `append` splits them depending on spare capacity. Rust rejects the snippet with E0382 or E0502.
- **Explicit copies (J9, J10):** Swift copies silently. Mojo and Hylo moved to explicit copies.

## Open risks

- **Bindings whose type is unknown** (30 sites in std, J15): the node measures them with the check itself. Accepted.
- **Field reads stay under E0978** (`const items = bag.items; Arrays.push(mutable bag.items, 3)`). Accepted: a field read is how code reads, and the error names the later use.

## Inferred choices

- **No migration path,** because the language is unreleased (J11).

## Answers from the user

- **Q1:** silent copy-on-write is out. `copy` already exists, and `let b = a` is what needs resolving (2026-11-16).
- **Q2:** proof first, then decide (2026-11-16).
- **Q3:** restrict renames, since a second name for one value is a code smell (2026-11-16).

## Overruled

- **P1, P2, P3, P6, O7:** rejected by the maintainer (Q1).
- **P9, P12:** rejected by the maintainer (Q3). P14 parked.

## Cost of this debate

- 2 of 2 rounds, plus proof T1 and the rename count: about 328k tokens.
