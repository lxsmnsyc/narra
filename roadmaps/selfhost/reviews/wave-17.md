# Wave 17 report

## Merged

- #95 fix: five names from closure-borrow that the unused-names rule rejects. Main was red after #92 and #94 merged without a text conflict. A commit subject with "fix #95" pushed to main closed the pull request by mistake; the maintainer reopened and merged it.
- #93 coroutines, after two merges of main into the branch (std module list, `resolve.nra`, the borrow helpers #92 moved to `exprs.nra`, and a `_idle` rename for the unused-names rule).
- #96 mutable-reference (D81): `mutable` changes the contents of a struct, array, map or other contents type, never the binding; `const` fixes the name only. The first draft misread the request and was rewritten from the maintainer's choice. The fresh review found that D81 recorded the spike's extra rules as decided (fixed) and a use-after-free through nested closures in rule 9 (fixed). The maintainer chose read-only copying bindings (option b) and accepted rules 6 to 10.
- #97 call-type-args: `f<T>(x)` and `Trait.f<T>(x)`. Review round: SCREAMING_CASE constants in argument lists parsed as type arguments (fixed: names of two or more characters with no lowercase letter are constants), lowercase type parameters, stale E0512 text, missing tests.
- #98 effects-indirect-calls: calls through fields, indexes and call results charge their effects. The fresh review found a callee's effect variable leaking into the caller (a false positive and a new hole), fixed by substituting the bound effects. Updated onto main after #97 (one more `CallExpr` field).
- #99 linux-check: `scripts/linux-check.sh` runs the full check in Docker. Linux found a missing `_POSIX_C_SOURCE` and a strict-aliasing bug in the emitted `Str.split` code (undefined behaviour on macOS too).

Main after #96, #97 and #99: full check passes (800 s). #98 updated onto that main: passes (684 s).

## Decisions by the maintainer

- Wave 16's unused-names and coroutines decisions: confirmed.
- D81: the model above, the widened type list, read-only plain parameters, read-only copying bindings (option b), rules 6 to 10.
- Unused imports and private functions: approved as node unused-imports.

## Plan changes

- New nodes: unused-imports (M), linux-check (S, done), mutable-markers (S spike, D85: where a `mutable` marker can make a binding a reference, as in `x is Some(mutable y)`), func-type-effects (must, M, D86: function types carry their effects; the review of #98 showed that following written types cannot be sound). effects-indirect-calls raised to must.
- mutable-bindings refined into a full brief (L) that builds D81.
- Inbox: SPEC still says there is no `+=` (D76); arguments are not evaluated left to right when a later block argument reassigns a `let`.

## Process lessons

- A merge-order note in a report did not prevent a red main. When two ready pull requests interact without a text conflict, update the second onto the first before the hand-off and say the order in both pull requests. Done for #97 and #98.
- Never write "fix #N" or "closes #N" in a commit pushed to main.
- Auto mode blocks force-pushes; merging main into the branch and pushing fast-forward works instead.
- Every fresh review again found a real bug (D81 nested closures, #97 parser regression, #98 effect variable leak). Keep fresh reviews for M nodes and for S fixes in the checker.
