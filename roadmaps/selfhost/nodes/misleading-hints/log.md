# Log: misleading-hints

## Proof on the base (6478acd, origin/feat/batch-2)

Built with `sh selfhost/bootstrap.sh <dir>`, then `narrac check <file>`.

- Hint 1 (field of a parameter): already right. D81 made `checkContentsChange` report "parameter ... is read-only; declare it as `mutable player` ...".
  `narrac check testdata/errors/mutable_contents_param.nra` prints
  `mutable_contents_param.nra:7:3: parameter "player" is read-only; declare it as `mutable player` to change the caller's value, or copy it first with `const copy = player``.
  The existing error test already covers it. No change.
- Hint 2 (user trait): reproduced. New test `testdata/errors/traits_missing_impl_user.nra`:
  `Money implements Label but has no `implement Label for Money` block or `derives Label``.
  Also in `testdata/declsdump/traits/err_trait_decl.dump`.

## Fix

`checkListed` in `selfhost/src/traits.nra` adds "or `derives X`" only when the trait is a std trait or Shape (same test as the E0503 check).

## Guide

Ch 7 (line 69) and ch 9 (line 25) already show the new messages. Guide suite passes, no guide edit needed.
