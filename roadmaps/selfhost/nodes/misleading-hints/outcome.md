# Outcome: misleading-hints

## Built

- Hint 2 fixed in `selfhost/src/traits.nra` (`checkListed`): the E0507 message names `derives X` only for built-in derivable traits (std traits and Shape). A user trait gets "... block".
- New error test `testdata/errors/traits_missing_impl_user.nra` and `.err`; `testdata/declsdump/traits/err_trait_decl.dump` updated.
- Hint 1 (field of a parameter) was already fixed by D81; nothing changed.
- Guide chapters 7 and 9 already show the new messages; no guide edit.

## Proof

- base: 6478acd, `narrac check testdata/errors/traits_missing_impl_user.nra` prints `... block or `derives Label``. `narrac check testdata/errors/mutable_contents_param.nra` already prints the parameter hint (hint 1 not reproduced).
- branch: same command prints `Money implements Label but has no `implement Label for Money` block`; built-in `Show` still suggests `derives Show` (`traits_missing_impl.err`).

## Deviations

- Hint 1 needed no change. Guide needed no edit.

## Stubs left

- None.

## Gotchas

- The E0507 row in the code table in `diagnostics.nra` still has the long format text; the message is built inline in `traits.nra`.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh` with the built narrac: all suites pass except the `err_trait_decl.dump` golden, which changed as expected; it was rewritten with `--suite dumps --update` after the full run. Suites errors, dumps, guide and sources pass.

## Size

- Estimated S, actual S.
