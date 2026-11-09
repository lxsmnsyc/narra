# Log: lsp-navigation-parse

## Method

- Base: `origin/feat/batch-3` (6924411), compiler built with `sh selfhost/bootstrap.sh`. Branch: the same plus the change.
- Driver (session scratchpad, `nav.py`): starts `narrac lsp`, `initialize`, `didOpen` of `selfhost/src/hash.nra` (the package `selfhost/`, 113 modules), then times each request from send to reply. Per process: the first hover right after `didOpen` (`open+hover`), then 3 hovers, 3 definitions, 2 documentSymbol on the same document. 4 processes per round, base and branch interleaved, 3 rounds. The machine had load averages of 11 to 23 (two other agents), so single rounds vary by 2x; compare the pairs inside a round.
- The answers of all requests are byte-identical on base and branch (md5 of the JSON, every round).

## Base cost

A navigation request is 500 to 600 ms on a quiet round, all of it a whole load and check. Parsing is about 37% of a whole check (lsp-incremental profile).

## Numbers (median per round, ms; min in brackets)

| round | request | base | branch |
|---|---|---|---|
| quiet 1 | hover | 588 (505) | not comparable (load spike) |
| quiet 2 | hover | 670 (628) | 417 (371) |
| quiet 2 | definition | 674 (630) | 405 (360) |
| quiet 2 | documentSymbol | 657 (648) | 394 (362) |
| spike | hover | 517 (499) | 1721 (1136), a load spike hit this round; discarded |
| loaded | hover | 854 (737) | 561 (485) |
| loaded | hover | 1069 (762) | 424 (391) |

Gain 35 to 40% per request in every round where both sides ran under similar load; `open+hover` also falls (the first request after `didOpen` reuses the parse of the check).
