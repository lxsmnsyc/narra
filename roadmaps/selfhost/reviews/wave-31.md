# Wave 31 report: batch-23, the compiler uses the new std

Batch-23 starts from `main` after PR #129 (C in the benchmarks). Local branch `feat/batch-23`, worktree `../narra-batch-23`. Maintainer request, 2026-11-20: "begin run".

## Landed

- compiler-std-sorts: 24 hand-written sorts in `selfhost/src` now call `Arrays.sort` or `Arrays.sortBy`; their helpers and three recursive `mergeSort` copies are gone. `insertUse` in `tools/query.nra` stays: it removes duplicates while it inserts. The fresh review found 9 sorts the first pass left while its outcome said none remained; a second worker replaced them and corrected the outcome. Character checks stay: the lexer counts `_` as a letter, the trivia space check excludes line feeds, and the manifest check reads bytes as `int`.
- compiler-std-sets: 150 membership-only `Map<K, bool>` are `Set<K>`; four maps whose `false` means something stay. Fresh review: clean. Its nit (call `Sets.has` then `Sets.add`) is fixed in 11 places.
- compiler-std-files: `Process.cwd` replaces `pwd` and `pwd -P`; `Fs.removeDirAll` and `Fs.rename` replace `rm -rf` and the cache install and remove scripts. The git call, `chmod -R a-w` and `mktemp -d` stay (no environment in `Process.run`, no permission or temporary-folder function).
- std-header-comments: 14 std headers no longer cite SPEC sections or decision numbers.

Emitted C of `narrac emit selfhost` is byte-identical before and after each compiler node.

## Compiler cost

Interleaved runs of `narrac emit` on the batch-23 source, `main` compiler against the batch-23 compiler, each round the median of 5, load 4.2 to 3.7:

| round | main | batch-23 |
| ----- | ----- | -------- |
| 1 | 0.402 | 0.418 |
| 2 | 0.414 | 0.418 |
| 3 | 0.423 | 0.419 |
| 4 | 0.406 | 0.528 |

Peak memory 209 MB for both. The binary grows from 5,645,000 to 5,646,200 bytes. No change beyond noise.

## Checks

- `sh scripts/test.sh`: PASS, 10925 cases.
- `sh selfhost/bootstrap.sh`: fixed point, seed current (regenerated in `bd7de089`).
- `pnpm --dir website build`: passes.
- `sh scripts/linux-check.sh`: PASS, 10925 cases.

## Outside the batch

- PR #129, merged: C in the benchmarks, Odin removed, C on the site chart.
- PR #130, open: the Vercel ignore step moved to `website/scripts/vercel-ignore.sh` (Vercel rejects an `ignoreCommand` over 256 characters).

## Still to do

- [ ] Push `feat/batch-23` and open its pull request.
- [ ] After it merges: first-release (`v0.1.0`), pushed by the maintainer.
