# Log: named-args-query

- 2026-11-09: branch `feat/named-args-query` from `origin/feat/batch-1`.
- `callers` and `type-of` already handle labelled calls: the callee name site and the effect pass's call graph do not depend on argument order.
- Added `SiteKind.LabelSite` (querysites.nra) for each `name: value` argument, and `refs FUNCTION.PARAMETER` (query.nra), which lists the labels of calls that resolve to the function.
- The label's call ref is the call node's ref, as for `CalleeSite`, so the match needs no parameter index: a label equals the parameter name.
