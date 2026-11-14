# Log: std-toml

## Proof on the base

- Base: feat/batch-14 at 7569f176. Compiler: `sh selfhost/bootstrap.sh /private/tmp/claude-501/std-toml/base2` from a `git archive` of the base (release build).
- Command: `/private/tmp/claude-501/std-toml/base2/narrac run testdata/toml.nra`.
- Result: `toml.nra:3:1: unknown module "std/toml" (this compiler supports std/int, ...)`.

## Work

- Branch compiler: `base2/narrac emit selfhost` then `cc -O2 -DNARRA_RELEASE` (release), after `--suite generated --update`.
- Seed: `cc` of `seed/narrac.c` into narrac0; the three samples run with `std/src/toml.nra` copied into a scratch package as `t/toml` (its `std/internal/float` import swapped for a local stub, since only std modules may import `std/internal/`). All exit 0.
- Emit: `narrac emit selfhost` of the base tree (`git archive feat/batch-14`) with the base and the branch compiler: byte-identical, 223,942 lines.
- `NARRAC=<branch> sh scripts/test.sh --filter toml`: PASS, 28 cases.

## Full check

- `sh scripts/test.sh` on the squashed commit (base feat/batch-14 7569f176): `test.sh: PASS, 8636 cases in 376s`.

## Review fixes

- Fixed: `stringify` writes negative zero as `-0.0`; `stringify` returns `Err` past 512 levels of nesting, the limit of `parse`. Tests in `testdata/toml.nra`. Noted in the outcome: `addNode` memory per input byte.
- Rebased on feat/batch-14 6c8c5998 (std-arena landed; kept both sides of stdlib.nra, compiler/README.md and the guide chapter; module-list goldens rewritten). `--filter toml`: PASS, 28 cases. `--suite errors`: PASS, 767 cases. guide 307, sources 1, dumps 172, generated 20: PASS.
