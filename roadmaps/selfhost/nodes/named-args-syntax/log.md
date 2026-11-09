# Log: named-args-syntax

- Added `pos` and `name` to `Arg`; parser, astdump, fmtexprs updated. Source uses no named arguments (seed rule).
- Checker placeholder first used E0000; the `codes` suite rejected it (not in table). Switched to E0213.
- fmt goldens must be named `expr_*` to be picked up; created an empty .fmt and ran `--update` with `JOBS=1` (a filter that matches nothing in a shard fails that shard).
- Ternary-looking cases were dropped from the parse sample: Narra has no `?:`.
