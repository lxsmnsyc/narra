# Outcome: emit-string-helpers

## Built

- `cString` copies runs of unchanged bytes with one `Str.slice` each and appends escapes directly. `findByteChar` and the `ASCII` table are removed.
- `writeLine` appends the indent, the text and the newline to `emitter.out` in turn, with no `padding + text + "\n"` string.
- Only `selfhost/src/emitstate.nra` changed, in those two functions.

## Result

- Instructions retired for `narrac emit selfhost` (base d92bf72, 3 runs each): base 3.615, 3.592, 3.596 billion; branch 3.425, 3.411, 3.412 billion. About 5.2% fewer, far outside the 0.6% spread. Peak memory unchanged (137 MB).
- Emitted C is byte-identical (same md5 from the base and branch compilers); bootstrap stages 1 and 2 agree.
- `sh scripts/test.sh` passes.

## Deviations

- None. The seed is not regenerated.
