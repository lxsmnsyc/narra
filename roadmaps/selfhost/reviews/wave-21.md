# Wave 21 report: batch-8

Batch-8 ran on `feat/batch-8` as one story: the D95 fixed arrays and C interop step 1. Children landed as direct commits after a worker check and a fresh review.

## Landed

- attributes: declarations carry a general attribute list (`KNOWN_ATTRIBUTES`, `findAttribute`); E1000-E1099 range row. Reviewed inline.
- const-generics: `const N: int` on functions and structs; `N` is a constant per instance for the prover. Review found a parse regression (`both(a < 3, 4 > (c))`); the maintainer chose TypeScript's syntactic rule, so that line now parses as a generic call and E0606 suggests parentheses. Also: E0979/E0980 name the instance and its call, shadowing a constant parameter is E0103.
- fixed-arrays: `FixedArray<T, N>` stored inline, indexes proven at both ends with no runtime check, 1 MiB per value. Review found a hole that was also on main: the rest of a `&&`/`||` condition could reassign a narrowed name inside an arrow (a hidden division by zero on main, an out-of-bounds read with fixed arrays). Fixed in its own commit, c455b865.
- unsafe-ptr: `unsafe { }`, `Ptr<T>` over plain values, `@trusted`, E0996-E0998. Review found the unsafe effect escaping through generators and trait implementations; fixed for `unsafe` (an implementation may use it only when its trait function declares it). `alloc` and `io` through implementations stay a documented gap (inbox, needs a maintainer decision).
- c-interop: `extern { } from "lib"`, `@extern`, `std/cstring`, libraries linked from reached members, `[link] search` passed as `-L`, `[link] libraries` removed, E1000-E1002. Members are declared with assembler labels so system headers never clash. Review found `@extern` names could hit the compiler's prefix helpers (`release_<type>`); closed.
- c-audit-target: `narrac audit`, `@target` on extern blocks applied per member.
- End of batch: the whole-branch review allowed pointers to fixed arrays of plain values (the two features had landed in parallel), fixed fmt for comments around an extern block's `from`, named implementation members in `narrac audit`, and fixed four stale SPEC and guide texts. Two samples used `cbrt`, which glibc does not round exactly; they use `sqrt`. Seed regenerated twice (after the children and after the review fixes).

## Checks

- `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch; the bootstrap reports the seed current.
- `narrac emit` on the batch-7 source: +0.9% instructions and +4 MB peak memory; the acceptance limit was 2% instructions.

## Process lessons

- Two features built in parallel each added a Type variant and neither handled the other (`Ptr<FixedArray<...>>`). When two nodes add types in one wave, the second to land should check the cross product.
- Reviews keep finding soundness holes in older code once a new feature makes them reachable (the narrowing hole, the effect gaps through generators and implementations). Asking reviewers to probe the base for the same pattern paid off each time.
- A one-line glibc/macOS libm difference failed the Linux check: samples must use functions whose results are exact on both (IEEE-required ones such as `sqrt`).
