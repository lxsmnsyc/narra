# Outcome: lexer-punctuation

## Built

- `selfhost/src/lexer.nra`: the `PUNCTS` array is replaced by the const `PUNCTS_BY_FIRST_BYTE` (`Map<int, Array<string>>`, candidates longest first). `lexPunctuation` looks up the byte at `peek(lexer, 0)` and tries only those candidates with `hasPrefix`. An unknown first byte gives the same E0004 error as before.
- Token text, kinds and order are unchanged. `TokPunct` still carries the punctuation string and the parser still compares token text. Comparing by kind is left to parser-token-kinds.

## Proof

- base: 48823a59 (feat/batch-5), `/usr/bin/time -l narrac emit selfhost`, instructions retired, 5 runs, median 2628747558 (range 2626583582 to 2674455446; 0.1% spread without one outlier). Load average about 5.
- branch: same command, median 2548668842 (range 2544306695 to 2557173765). Gain 3.0%, well above the spread.

## Result

Not an experiment node.

## Deviations

- The lookup is a `Map<int, Array<string>>` const, not an array indexed by byte, because a 128-row literal is hard to read. parser-token-kinds may switch to a flat array if it reshapes this table.

## Stubs left

- None.

## Gotchas

- `PUNCTS_BY_FIRST_BYTE` keys are byte values (61 is `=`). Within a key, keep the order longest first when adding punctuation.
- parser-token-kinds can turn the strings in this table into kinds without changing the lookup.
- The seed was not regenerated. Measurements use a narrac built from selfhost by `sh selfhost/bootstrap.sh`.

## Proposed learnings

- Dispatch on the first byte before testing multi-byte candidates | avoids scanning the whole table per token | selfhost/src | no

## Verification

- `sh selfhost/bootstrap.sh`: fixed point holds.
- `sh scripts/test.sh`: result in the worker report.

## Size

- Estimated S, actual S.
