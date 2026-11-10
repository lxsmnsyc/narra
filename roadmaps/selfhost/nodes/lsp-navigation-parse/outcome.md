# Outcome: lsp-navigation-parse

## Built

- `selfhost/src/lspnavigation.nra`: `loadNavigation` loads the package with `LspCache.loadPackage(mutable server.cache, ...)` instead of `Modules.loadWith`, so unchanged sources keep their parse. `answerNavigation` and `loadNavigation` take `mutable server`.
- `selfhost/src/lsp.nra`: passes `mutable server` to `answerNavigation`.
- The check, effect pass and views still run per request. A cached whole checker is out of scope.

## Proof

- base (6924411, `narrac lsp` on package `selfhost/`, document `src/hash.nra`, method in `log.md`): hover, definition and documentSymbol take 500 to 670 ms each on a quiet round (min 495).
- branch: 371 to 417 ms in the paired quiet round (base 628 to 670 in the same round); 485 to 561 against 737 to 854 in a loaded pair. About 35 to 40% less, matching the 37% parse share. Spread: one round in three was swamped by load from other agents and is discarded in the log. Answers are byte-identical on base and branch.

## Deviations

- None.

## Stubs left

- None. The check phase is still about 60% of a request.

## Gotchas

- Navigation shares `server.cache.parses` with the diagnostics checks, so a navigation request can start the parse cache over (the id high-water rule in `loadPackage`) and make the next diagnostics check a whole one.

## Proposed learnings

- None.

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh --suite lsp`: pass.
- `sh scripts/test.sh`: see report.

## Size

- Estimated M, actual S.
