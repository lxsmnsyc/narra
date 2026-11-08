# Log: test-runner-core

- Reset to origin/feat/c-seed (bc239e2), branch feat/test-runner-core.
- Built Go `narra`, then narrac (release) from `selfhost/`. The runner is built by narrac and removes temp dirs with `rm -rf` through `sh -c`.
- Goldens: `.err` from Go `narra check` for 286 `errors/*.nra` and 26 `packages/errors/*/` (written as `expected.err` inside the folder, because a file beside the folder would break Go's `filepath.Glob("errors/*")` in TestPackageErrors). narrac `check` equals all 312. The 2 `stage2/errors` goldens come from narrac, Go cannot load their std modules.
- Struct fields of function type: the field list is newline separated, and `narrac fmt` removes the commas; a trailing `uses io, alloc,` in the last field is a parse error, so none is written.
- A `+` between strings is an error in Narra (template strings only); Result narrowing needs one `is Err` early return per value, not `a is Err(_) || b is Err(_)`.
- Full run with 8 shards plus a bootstrap job: 513 cases, all pass.
- Break test: one `.out`, one `.err`, one panic sample each fail with a diff or message and pass again after restoring (see outcome).
